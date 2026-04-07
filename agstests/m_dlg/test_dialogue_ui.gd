## T-DLG16 — AGSDialogueUI presenter tests.
extends "res://utils/test_base.gd"

const DialogueScript = preload("res://../game_prototype/.engine/runtime/ags_dialogue.gd")
const UIScript       = preload("res://../game_prototype/.engine/runtime/ags_dialogue_ui.gd")

func suite_name() -> String:
	return "M-DLG: DialogueUI"

func _make_pair() -> Array:
	var eng: Node = DialogueScript.new()
	_tree.root.add_child(eng)
	var ui: Node = UIScript.new()
	_tree.root.add_child(ui)
	# Wire UI to engine directly (bypasses AutoLoad lookup).
	ui._dlg = eng
	eng.dialogue_started.connect(ui._on_dialogue_started)
	eng.dialogue_ended.connect(ui._on_dialogue_ended)
	eng.line_ready.connect(ui._on_line_ready)
	eng.waiting_for_advance.connect(ui._on_waiting_for_advance)
	eng.choices_ready.connect(ui._on_choices_ready)
	return [eng, ui]

func _inject_nodes(eng: Node, nodes: Array) -> void:
	for n in nodes:
		eng._nodes[n["title"]] = n

func _cleanup(eng: Node, ui: Node) -> void:
	eng.queue_free()
	ui.queue_free()
	await _tree.process_frame

# UT-DLG16-01: Panel is hidden on init.
func test_01_panel_hidden_on_init() -> void:
	var pair := _make_pair()
	var ui: Node = pair[1]
	assert_false(ui._panel.visible, "Panel should be hidden on init")
	await _cleanup(pair[0], ui)

# UT-DLG16-02: Panel becomes visible when dialogue starts.
func test_02_panel_shows_on_started() -> void:
	var pair := _make_pair()
	var eng: Node = pair[0]
	var ui:  Node = pair[1]
	_inject_nodes(eng, [
		{"title": "n", "body": [{"type": "command", "raw": "<<end>>"}]}
	])
	await eng.start(null, "n")
	assert_false(ui._panel.visible, "Panel should be hidden after dialogue ends")
	await _cleanup(eng, ui)

# UT-DLG16-03: speaker_label and text_label are updated on line_ready.
func test_03_labels_updated_on_line_ready() -> void:
	var pair := _make_pair()
	var eng: Node = pair[0]
	var ui:  Node = pair[1]
	_inject_nodes(eng, [
		{"title": "greet", "body": [
			{"type": "speaker_line", "speaker": "guard", "text": "Halt!", "loc_key": "g:0:x"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "greet")
	assert_eq(ui._speaker_label.text, "guard", "speaker_label should be 'guard'")
	assert_eq(ui._text_label.text, "Halt!", "text_label should be 'Halt!'")
	await _cleanup(eng, ui)

# UT-DLG16-04: speaker_label is hidden for narration (empty speaker).
func test_04_speaker_hidden_for_narration() -> void:
	var pair := _make_pair()
	var eng: Node = pair[0]
	var ui:  Node = pair[1]
	_inject_nodes(eng, [
		{"title": "narr", "body": [
			{"type": "narration", "text": "Silence.", "loc_key": "n:0:x"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "narr")
	assert_false(ui._speaker_label.visible, "speaker_label should be hidden for narration")
	await _cleanup(eng, ui)

# UT-DLG16-05: portrait_requested is emitted when emotion tag is set.
func test_05_portrait_requested_emitted() -> void:
	var pair := _make_pair()
	var eng: Node = pair[0]
	var ui:  Node = pair[1]
	_inject_nodes(eng, [
		{"title": "emote", "body": [
			{"type": "speaker_line", "speaker": "Elara", "text": "Hi.", "loc_key": "e:0:x", "emotion": "happy"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	var got_char := [""]
	var got_emotion := [""]
	ui.portrait_requested.connect(func(c: String, e: String) -> void:
		got_char[0] = c
		got_emotion[0] = e
	)
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "emote")
	assert_eq(got_char[0], "Elara", "portrait_requested: char mismatch")
	assert_eq(got_emotion[0], "happy", "portrait_requested: emotion mismatch")
	await _cleanup(eng, ui)

# UT-DLG16-06: choices_container is populated with correct button count.
func test_06_choice_buttons_created() -> void:
	var pair := _make_pair()
	var eng: Node = pair[0]
	var ui:  Node = pair[1]
	_inject_nodes(eng, [
		{"title": "opts", "body": [
			{"type": "option", "text": "A", "loc_key": "o:0:x", "body": [{"type": "command", "raw": "<<end>>"}]},
			{"type": "option", "text": "B", "loc_key": "o:1:x", "body": [{"type": "command", "raw": "<<end>>"}]},
			{"type": "option", "text": "C", "loc_key": "o:2:x", "body": [{"type": "command", "raw": "<<end>>"}]},
		]}
	])
	eng.choices_ready.connect(func(_opts: Array) -> void:
		eng.choose(0)
	, CONNECT_ONE_SHOT)
	await eng.start(null, "opts")
	# After dialogue ends buttons should be cleared.
	assert_eq(ui._choice_buttons.size(), 0, "Choice buttons should be cleared after choice")
	await _cleanup(eng, ui)

# UT-DLG16-07: choosing an option hides choices_container and shows text_label.
func test_07_choice_hides_container() -> void:
	var pair := _make_pair()
	var eng: Node = pair[0]
	var ui:  Node = pair[1]
	_inject_nodes(eng, [
		{"title": "pick", "body": [
			{"type": "option", "text": "Go", "loc_key": "p:0:x", "body": [
				{"type": "narration", "text": "Moving.", "loc_key": "p:1:x"},
				{"type": "command", "raw": "<<end>>"},
			]},
		]}
	])
	var choices_hidden_after := [false]
	eng.choices_ready.connect(func(_opts: Array) -> void:
		eng.choose(0)
	, CONNECT_ONE_SHOT)
	eng.line_ready.connect(func(_sp: String, _tx: String, _lk: String, _em: String) -> void:
		choices_hidden_after[0] = not ui._choices_container.visible
	, CONNECT_ONE_SHOT)
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "pick")
	assert_true(choices_hidden_after[0], "choices_container should be hidden after choice")
	await _cleanup(eng, ui)

# UT-DLG16-08: _showing_line is reset to false after dialogue ends.
func test_08_showing_line_reset_on_end() -> void:
	var pair := _make_pair()
	var eng: Node = pair[0]
	var ui:  Node = pair[1]
	_inject_nodes(eng, [
		{"title": "simple", "body": [
			{"type": "narration", "text": "Done.", "loc_key": "s:0:x"},
			{"type": "command", "raw": "<<end>>"},
		]}
	])
	eng.waiting_for_advance.connect(func() -> void: eng.advance(), CONNECT_ONE_SHOT)
	await eng.start(null, "simple")
	assert_false(ui._showing_line, "_showing_line should be false after dialogue ends")
	await _cleanup(eng, ui)
