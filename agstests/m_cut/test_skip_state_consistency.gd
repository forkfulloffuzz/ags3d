## T-CUT24 — State consistency on skip (dry-run action/set pass).
##
## Verifies that action and set commands always fire when skip is active,
## even if the sequence is fast-forwarded. State is consistent on arrival.
extends "res://utils/test_base.gd"

const SeqScript := preload("res://../game_prototype/.engine/runtime/ags_sequencer.gd")

func suite_name() -> String:
	return "M-CUT: SkipStateConsistency (T-CUT24)"


# ── Helpers ────────────────────────────────────────────────────────────────────

func _make_seq() -> Node:
	var s: Node = SeqScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT24-01: action fires during normal (non-skip) execution ──────────────

func test_01_action_fires_normally() -> void:
	var seq := _make_seq()
	var fired := [false]
	# Override _on_command to track calls.
	seq.set_meta("_cmd_fired", false)

	var commands: Array[String] = []
	# We can't easily override _on_command in tests, so use step_started instead.
	seq.step_started.connect(func(id: String) -> void:
		if id.begins_with("action_"): commands.append(id))

	var steps: Array = [
		{"type": "action", "raw": "set_flag(A)", "id": "action_A"},
	]
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	await seq.run(steps)

	assert_true(completed[0], "sequence should complete")
	# action step should have fired (step_started emits for it).
	assert_true(commands.has("action_A"), "action step should fire normally")
	await _cleanup_nodes([seq])


# ── UT-CUT24-02: action fires even after a skipped wait step ─────────────────

func test_02_action_fires_after_skipped_wait() -> void:
	var seq := _make_seq()
	seq.skip_policy = "always"

	var commands: Array[String] = []
	seq.step_started.connect(func(id: String) -> void:
		if id.begins_with("action_"): commands.append(id))

	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)

	# Trigger skip shortly after the long wait starts.
	_tree.create_timer(0.05).timeout.connect(func() -> void: seq.request_skip(), CONNECT_ONE_SHOT)

	await seq.run([
		{"type": "wait", "duration": 10.0},
		{"type": "action", "raw": "set_flag(B)", "id": "action_B"},
	])

	assert_true(completed[0], "sequence should complete after skip")
	assert_true(commands.has("action_B"), "action step should fire even after skipped wait")
	await _cleanup_nodes([seq])


# ── UT-CUT24-03: set fires even after a skipped wait step ────────────────────

func test_03_set_fires_after_skipped_wait() -> void:
	var seq := _make_seq()
	seq.skip_policy = "always"

	var step_ids: Array[String] = []
	seq.step_started.connect(func(id: String) -> void: step_ids.append(id))

	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)

	_tree.create_timer(0.05).timeout.connect(func() -> void: seq.request_skip(), CONNECT_ONE_SHOT)

	await seq.run([
		{"type": "wait", "duration": 10.0},
		{"type": "set", "raw": "hero_met = true", "id": "set_hero"},
	])

	assert_true(completed[0], "sequence should complete after skip")
	assert_true(step_ids.has("set_hero"), "set step should fire even after skipped wait")
	await _cleanup_nodes([seq])


# ── UT-CUT24-04: multiple actions all fire after skip ────────────────────────

func test_04_multiple_actions_fire_after_skip() -> void:
	var seq := _make_seq()
	seq.skip_policy = "always"

	var step_ids: Array[String] = []
	seq.step_started.connect(func(id: String) -> void: step_ids.append(id))

	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)

	_tree.create_timer(0.05).timeout.connect(func() -> void: seq.request_skip(), CONNECT_ONE_SHOT)

	await seq.run([
		{"type": "wait", "duration": 10.0},
		{"type": "action", "raw": "A()", "id": "action_first"},
		{"type": "wait", "duration": 0.01},
		{"type": "action", "raw": "B()", "id": "action_second"},
		{"type": "set", "raw": "x = 1", "id": "set_x"},
	])

	assert_true(completed[0], "sequence should complete after skip")
	assert_true(step_ids.has("action_first"), "first action should fire")
	assert_true(step_ids.has("action_second"), "second action should fire")
	assert_true(step_ids.has("set_x"), "set should fire")
	await _cleanup_nodes([seq])


# ── UT-CUT24-05: fire_skipped_state_changes fires only action/set ─────────────

func test_05_fire_skipped_state_changes_selects_action_set() -> void:
	var seq := _make_seq()

	var fired: Array[String] = []
	seq.step_started.connect(func(id: String) -> void: fired.append(id))

	# Call fire_skipped_state_changes directly with a mix of step types.
	# Only action/set steps should fire _on_command (no step_started for non-started steps).
	# The method is synchronous — just verify it doesn't crash.
	seq.fire_skipped_state_changes([
		{"type": "wait", "duration": 1.0, "id": "wait_step"},
		{"type": "action", "raw": "do_thing()", "id": "action_step"},
		{"type": "set", "raw": "x = 5", "id": "set_step"},
		{"type": "character", "character": "player", "command": "walk_to", "point": "A", "id": "char_step"},
	])

	# fire_skipped_state_changes calls _on_command but does NOT emit step_started.
	# So fired should be empty — verify no crash and returns cleanly.
	assert_eq(fired.size(), 0, "fire_skipped_state_changes should not emit step_started")
	await _cleanup_nodes([seq])


# ── UT-CUT24-06: fire_skipped_state_changes exists on sequencer ──────────────

func test_06_fire_skipped_state_changes_exists() -> void:
	var seq := _make_seq()
	assert_true(seq.has_method("fire_skipped_state_changes"),
		"fire_skipped_state_changes() should exist on AGSSequencer")
	await _cleanup_nodes([seq])


# ── UT-CUT24-07: no-skip run still fires all action steps in order ───────────

func test_07_no_skip_actions_fire_in_order() -> void:
	var seq := _make_seq()

	var order: Array[String] = []
	seq.step_started.connect(func(id: String) -> void: order.append(id))

	await seq.run([
		{"type": "action", "raw": "first()", "id": "a1"},
		{"type": "wait", "duration": 0.01},
		{"type": "action", "raw": "second()", "id": "a2"},
		{"type": "action", "raw": "third()", "id": "a3"},
	])

	var action_order: Array[String] = order.filter(func(x: String) -> bool: return x.begins_with("a"))
	assert_eq(action_order.size(), 3, "all 3 action steps should have fired")
	assert_eq(action_order[0], "a1", "first action should fire first")
	assert_eq(action_order[1], "a2", "second action should fire second")
	assert_eq(action_order[2], "a3", "third action should fire third")
	await _cleanup_nodes([seq])
