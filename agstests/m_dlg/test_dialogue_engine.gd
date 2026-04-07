## T-DLG14 — AGSDialogue runtime engine tests.
extends "res://utils/test_base.gd"

const DialogueScript = preload("res://../game_prototype/.engine/runtime/ags_dialogue.gd")

func suite_name() -> String:
	return "M-DLG: DialogueEngine"

func _make_engine() -> Node:
	var eng: Node = DialogueScript.new()
	_tree.root.add_child(eng)
	return eng

func _inject_nodes(eng: Node, nodes: Array) -> void:
	for n in nodes:
		eng._nodes[n["title"]] = n

func _cleanup(eng: Node) -> void:
	eng.queue_free()
	await _tree.process_frame

# UT-DLG14-01: Engine initialises with empty node table and inactive state.
func test_01_initial_state() -> void:
	var eng := _make_engine()
	assert_false(eng._active, "Engine should not be active on init")
	assert_eq(eng._nodes.size(), 0, "Node table should be empty on init")
	await _cleanup(eng)

# UT-DLG14-02: load_all with missing directory sets _loaded without crash.
func test_02_load_all_missing_dir() -> void:
	note("WARNING below is intentional: verifying that load_all gracefully handles a missing directory")
	var eng := _make_engine()
	eng.load_all("res://nonexistent_dir/")
	assert_true(eng._loaded, "_loaded should be true even if dir is missing")
	await _cleanup(eng)

# UT-DLG14-03: start() emits dialogue_started and dialogue_ended signals.
func test_03_start_emits_started_ended() -> void:
	var eng := _make_engine()
	_inject_nodes(eng, [
		{"title": "test_node", "body": [{"type": "command", "raw": "<<end>>"}]}
	])
	var started_title := [""]
	var ended_title := [""]
	eng.dialogue_started.connect(func(t: String) -> void: started_title[0] = t)
	eng.dialogue_ended.connect(func(t: String) -> void: ended_title[0] = t)
	await eng.start(null, "test_node")
	assert_eq(started_title[0], "test_node", "dialogue_started should emit node title")
	assert_eq(ended_title[0], "test_node", "dialogue_ended should emit node title")
	await _cleanup(eng)

# UT-DLG14-04: A speaker_line emits line_ready and waits for advance().
func test_04_speaker_line_emits_and_waits() -> void:
	var eng := _make_engine()
	_inject_nodes(eng, [
		{"title": "greet", "body": [
			{"type": "speaker_line", "speaker": "guard", "text": "Halt!", "loc_key": "greet:0:abc"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	var received_speaker := [""]
	var received_text := [""]
	eng.line_ready.connect(func(sp: String, tx: String, _lk: String, _em: String) -> void:
		received_speaker[0] = sp
		received_text[0] = tx
	)
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "greet")
	assert_eq(received_speaker[0], "guard", "line_ready: speaker mismatch")
	assert_eq(received_text[0], "Halt!", "line_ready: text mismatch")
	await _cleanup(eng)

# UT-DLG14-05: Narration emits line_ready with empty speaker.
func test_05_narration_empty_speaker() -> void:
	var eng := _make_engine()
	_inject_nodes(eng, [
		{"title": "narr", "body": [
			{"type": "narration", "text": "Three years.", "loc_key": "narr:0:abc"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	var speaker_received := ["NOT_EMPTY"]
	eng.line_ready.connect(func(sp: String, _t: String, _lk: String, _em: String) -> void:
		speaker_received[0] = sp
	)
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "narr")
	assert_eq(speaker_received[0], "", "Narration should have empty speaker")
	await _cleanup(eng)

# UT-DLG14-06: A command fires command_fired signal.
func test_06_command_fires_signal() -> void:
	var eng := _make_engine()
	_inject_nodes(eng, [
		{"title": "cmd_node", "body": [
			{"type": "command", "raw": "<<action flag.done = true>>"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	var fired_raw := [""]
	eng.command_fired.connect(func(r: String) -> void: fired_raw[0] = r)
	await eng.start(null, "cmd_node")
	assert_eq(fired_raw[0], "<<action flag.done = true>>", "command_fired raw mismatch")
	await _cleanup(eng)

# UT-DLG14-07: <<jump>> navigates to another node.
func test_07_jump_navigates() -> void:
	var eng := _make_engine()
	_inject_nodes(eng, [
		{"title": "node_a", "body": [{"type": "command", "raw": "<<jump node_b>>"}]},
		{"title": "node_b", "body": [
			{"type": "narration", "text": "In B.", "loc_key": "b:0:abc"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	var texts: Array[String] = []
	eng.line_ready.connect(func(_sp: String, tx: String, _lk: String, _em: String) -> void:
		texts.append(tx)
	)
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "node_a")
	assert_eq(texts.size(), 1, "Expected 1 line after jump")
	assert_eq(texts[0], "In B.", "Jump should navigate to node_b")
	await _cleanup(eng)

# UT-DLG14-08: Options emit choices_ready; choose() selects the branch.
func test_08_options_choice_branch() -> void:
	var eng := _make_engine()
	_inject_nodes(eng, [
		{"title": "choice_node", "body": [
			{"type": "option", "text": "Yes", "loc_key": "c:0:abc", "body": [
				{"type": "narration", "text": "You chose yes.", "loc_key": "c:1:abc"},
				{"type": "command", "raw": "<<end>>"},
			]},
			{"type": "option", "text": "No", "loc_key": "c:2:abc", "body": [
				{"type": "command", "raw": "<<end>>"},
			]},
		]}
	])
	var choices_received: Array = []
	var branch_text := [""]
	eng.choices_ready.connect(func(opts: Array) -> void:
		choices_received.assign(opts)
		# Immediately choose option 0 (Yes).
		eng.choose(0)
	)
	eng.line_ready.connect(func(_sp: String, tx: String, _lk: String, _em: String) -> void:
		branch_text[0] = tx
	)
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "choice_node")
	assert_eq(choices_received.size(), 2, "Expected 2 choices")
	assert_eq(choices_received[0]["text"], "Yes", "First choice text mismatch")
	assert_eq(branch_text[0], "You chose yes.", "Wrong branch executed")
	await _cleanup(eng)

# UT-DLG14-09: start() ignores a second call while active.
func test_09_reentrant_start_ignored() -> void:
	var eng := _make_engine()
	_inject_nodes(eng, [
		{"title": "slow", "body": [
			{"type": "narration", "text": "Line.", "loc_key": "s:0:abc"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	var started_count := [0]
	eng.dialogue_started.connect(func(_t: String) -> void: started_count[0] += 1)
	eng.waiting_for_advance.connect(func() -> void:
		# While waiting, try to start again — should be ignored.
		eng.start(null, "slow")
		eng.advance()
	)
	await eng.start(null, "slow")
	assert_eq(started_count[0], 1, "Second start() while active should be ignored")
	await _cleanup(eng)
