## T-CUT20 — Flow and state command tests.
##
## Tests parallel blocks, if/else branching, on_event handlers, skip_to jumps,
## nested cutscene dispatch, and condition evaluation.
## Base class wait/action/set/label/end are covered by T-CUT12/T-CUT13/T-CUT14.
extends "res://utils/test_base.gd"

const SeqCmds := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")

func suite_name() -> String:
	return "M-CUT: FlowCommands (T-CUT20)"


# ── Helpers ────────────────────────────────────────────────────────────────────

func _make_seq() -> Node:
	var s: Node = SeqCmds.new()
	_tree.root.add_child(s)
	return s

func _run_step(seq: Node, step: Dictionary) -> Array:
	var completed := [false]
	var failed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	seq.sequence_failed.connect(func(_r: String) -> void: failed[0] = true, CONNECT_ONE_SHOT)
	await seq.run([step])
	return [completed[0], failed[0]]

func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT20-01: parallel block with two wait steps both complete ─────────────

func test_01_parallel_two_waits_complete() -> void:
	var seq := _make_seq()

	var step := {
		"type": "parallel",
		"steps": [
			{"type": "wait", "duration": 0.02},
			{"type": "wait", "duration": 0.01},
		]
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "parallel block should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT20-02: parallel block with empty steps completes ───────────────────

func test_02_parallel_empty_steps_completes() -> void:
	var seq := _make_seq()

	var step := {"type": "parallel", "steps": []}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "parallel with empty steps should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT20-03: if true — runs then branch ──────────────────────────────────

func test_03_if_true_runs_then_branch() -> void:
	var seq := _make_seq()

	var ran := [false]
	seq.step_started.connect(func(id: String) -> void:
		if id == "__then__": ran[0] = true)

	var step := {
		"type": "if",
		"condition": "true",
		"then": [{"type": "wait", "duration": 0.0, "id": "__then__"}],
		"else": []
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "if(true) should complete")
	assert_true(ran[0], "then branch step should have run")
	await _cleanup_nodes([seq])


# ── UT-CUT20-04: if false — runs else branch only ────────────────────────────

func test_04_if_false_runs_else_branch() -> void:
	var seq := _make_seq()

	var then_ran := [false]
	var else_ran := [false]
	seq.step_started.connect(func(id: String) -> void:
		if id == "__then__": then_ran[0] = true
		elif id == "__else__": else_ran[0] = true)

	var step := {
		"type": "if",
		"condition": "false",
		"then": [{"type": "wait", "duration": 0.0, "id": "__then__"}],
		"else": [{"type": "wait", "duration": 0.0, "id": "__else__"}]
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "if(false) should complete")
	assert_false(then_ran[0], "then branch should NOT run when condition is false")
	assert_true(else_ran[0], "else branch should run when condition is false")
	await _cleanup_nodes([seq])


# ── UT-CUT20-05: if with missing condition defaults to false ─────────────────

func test_05_if_empty_condition_is_false() -> void:
	var seq := _make_seq()

	var else_ran := [false]
	seq.step_started.connect(func(id: String) -> void:
		if id == "__else__": else_ran[0] = true)

	var step := {
		"type": "if",
		"condition": "",
		"then": [],
		"else": [{"type": "wait", "duration": 0.0, "id": "__else__"}]
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "if(empty) should complete")
	assert_true(else_ran[0], "empty condition defaults to false → else branch runs")
	await _cleanup_nodes([seq])


# ── UT-CUT20-06: on_event step completes immediately ────────────────────────

func test_06_on_event_completes_immediately() -> void:
	var seq := _make_seq()

	var step := {
		"type": "on_event",
		"event": "test:cut20:06",
		"steps": []
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "on_event step should complete immediately")
	await _cleanup_nodes([seq])


# ── UT-CUT20-07: on_event with missing event name — no crash ─────────────────

func test_07_on_event_missing_name_no_crash() -> void:
	note("on_event with missing 'event' field should complete gracefully")
	var seq := _make_seq()

	var step := {"type": "on_event", "steps": []}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "on_event with missing name should complete gracefully")
	await _cleanup_nodes([seq])


# ── UT-CUT20-08: skip_to jumps over intervening steps ───────────────────────

func test_08_skip_to_jumps_to_label() -> void:
	var seq := _make_seq()

	var skipped_ran := [false]
	var target_ran := [false]
	seq.step_started.connect(func(id: String) -> void:
		if id == "__skipped__": skipped_ran[0] = true
		elif id == "__target__": target_ran[0] = true)

	var steps: Array = [
		{"type": "skip_to", "label": "destination"},
		{"type": "wait", "duration": 0.0, "id": "__skipped__"},
		{"type": "label", "name": "destination"},
		{"type": "wait", "duration": 0.0, "id": "__target__"},
	]
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	await seq.run(steps)

	assert_true(completed[0], "sequence should complete after skip_to")
	assert_false(skipped_ran[0], "skipped step should NOT run")
	assert_true(target_ran[0], "step after label should run")
	await _cleanup_nodes([seq])


# ── UT-CUT20-09: skip_to with missing label — no crash ──────────────────────

func test_09_skip_to_missing_label_no_crash() -> void:
	note("skip_to with non-existent label should not crash — advances to next step")
	var seq := _make_seq()

	var steps: Array = [
		{"type": "skip_to", "label": "ghost_label"},
		{"type": "wait", "duration": 0.0},
	]
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	await seq.run(steps)

	assert_true(completed[0], "sequence should complete even with missing skip_to label")
	await _cleanup_nodes([seq])


# ── UT-CUT20-10: nested cutscene with unknown name — no crash ────────────────

func test_10_nested_cutscene_not_found_no_crash() -> void:
	var seq := _make_seq()

	var step := {"type": "cutscene", "name": "ghost_cutscene"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "nested cutscene with unknown name should complete as no-op")
	await _cleanup_nodes([seq])


# ── UT-CUT20-11: condition evaluation — true literals ───────────────────────

func test_11_condition_true_literals() -> void:
	var seq := _make_seq()
	# Verify _eval_condition is accessible and returns true for known values.
	assert_true(seq._eval_condition("true"), "'true' should eval to true")
	assert_true(seq._eval_condition("1"), "'1' should eval to true")
	assert_true(seq._eval_condition("yes"), "'yes' should eval to true")
	await _cleanup_nodes([seq])


# ── UT-CUT20-12: condition evaluation — false literals ──────────────────────

func test_12_condition_false_literals() -> void:
	var seq := _make_seq()
	assert_false(seq._eval_condition("false"), "'false' should eval to false")
	assert_false(seq._eval_condition("0"), "'0' should eval to false")
	assert_false(seq._eval_condition("no"), "'no' should eval to false")
	assert_false(seq._eval_condition(""), "empty string should eval to false")
	await _cleanup_nodes([seq])
