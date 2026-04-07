## T-CUT15 — AGSSequencer fallback policy tests.
extends "res://utils/test_base.gd"

const SeqScript = preload("res://../game_prototype/.engine/runtime/ags_sequencer.gd")

func suite_name() -> String:
	return "M-CUT: SequencerFallback"

func _make_seq() -> Node:
	var s: Node = SeqScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup(s: Node) -> void:
	s.queue_free()
	await _tree.process_frame

# UT-CUT15-01: skip_and_continue — step fails, sequence reaches completion.
func test_01_skip_and_continue() -> void:
	var seq := _make_seq()
	var completed := [false]
	var failed_ids: Array[String] = []
	seq.sequence_complete.connect(func() -> void: completed[0] = true)
	seq.step_failed.connect(func(id: String) -> void: failed_ids.append(id))
	await seq.run([
		{"type": "fail", "id": "bad_step", "on_fail": "skip_and_continue"},
		{"type": "wait", "duration": 0.0},
	])
	assert_true(completed[0], "skip_and_continue should allow sequence to reach completion")
	assert_true(failed_ids.has("bad_step"), "failed step should still emit step_failed")
	await _cleanup(seq)

# UT-CUT15-02: halt — step fails, sequence_failed emitted, no sequence_complete.
func test_02_halt() -> void:
	var seq := _make_seq()
	var completed := [false]
	var halt_reasons: Array[String] = []
	seq.sequence_complete.connect(func() -> void: completed[0] = true)
	seq.sequence_failed.connect(func(reason: String) -> void: halt_reasons.append(reason))
	await seq.run([
		{"type": "fail", "id": "bad_step", "on_fail": "halt"},
		{"type": "wait", "duration": 0.0},
	])
	assert_false(completed[0], "halt should not emit sequence_complete")
	assert_true(halt_reasons.size() > 0, "halt should emit sequence_failed with a reason")
	await _cleanup(seq)

# UT-CUT15-03: log_and_continue — step fails, sequence continues to completion.
func test_03_log_and_continue() -> void:
	note("WARNING below is intentional: log_and_continue policy emits a push_warning — this is the expected behaviour")
	var seq := _make_seq()
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true)
	await seq.run([
		{"type": "fail", "id": "noisy_step", "on_fail": "log_and_continue"},
		{"type": "wait", "duration": 0.0},
	])
	assert_true(completed[0], "log_and_continue should allow sequence to reach completion")
	await _cleanup(seq)

# UT-CUT15-04: retry_once — step fails on first attempt, succeeds on retry.
func test_04_retry_once_succeeds() -> void:
	var seq := _make_seq()
	var completed := [false]
	var fail_count := [0]
	seq.sequence_complete.connect(func() -> void: completed[0] = true)
	seq.step_failed.connect(func(_id: String) -> void: fail_count[0] += 1)
	# The step starts as "fail"; the step_failed signal changes it to "wait"
	# so the retry attempt succeeds.
	var step := {"type": "fail", "id": "retry_step", "on_fail": "retry_once"}
	seq.step_failed.connect(func(_id: String) -> void:
		step["type"] = "wait"
		step["duration"] = 0.0
	)
	await seq.run([step])
	assert_true(completed[0], "retry_once should complete after retry succeeds")
	assert_eq(fail_count[0], 1, "step should have failed exactly once before retry")
	await _cleanup(seq)

# UT-CUT15-05: retry_once exhausted — fails twice, sequence_failed emitted.
func test_05_retry_once_exhausted() -> void:
	note("ERROR below is intentional: retry_once exhausted calls push_error before halting — this is the expected behaviour")
	var seq := _make_seq()
	var halted := [false]
	seq.sequence_failed.connect(func(_r: String) -> void: halted[0] = true)
	await seq.run([
		{"type": "fail", "id": "bad_step", "on_fail": "retry_once"},
	])
	assert_true(halted[0], "retry_once exhausted should halt with sequence_failed")
	await _cleanup(seq)

# UT-CUT15-06: jump_to label — step fails, sequence jumps to named label.
func test_06_jump_to_label() -> void:
	var seq := _make_seq()
	var visited: Array[String] = []
	seq.step_complete.connect(func(id: String) -> void: visited.append(id))
	await seq.run([
		{"type": "fail", "id": "bad_step", "on_fail": "jump_to safe_point"},
		{"type": "wait", "id": "skipped", "duration": 0.0},
		{"type": "label", "name": "safe_point"},
		{"type": "wait", "id": "after_jump", "duration": 0.0},
	])
	assert_false(visited.has("skipped"), "step between fail and label should be skipped")
	assert_true(visited.has("after_jump"), "step after label should execute")
	await _cleanup(seq)

# UT-CUT15-07: per-step on_fail overrides cutscene_fallback.
func test_07_perstep_overrides_cutscene_fallback() -> void:
	var seq := _make_seq()
	seq.cutscene_fallback = "halt"  # would halt without per-step override
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true)
	await seq.run([
		{"type": "fail", "id": "bad_step", "on_fail": "skip_and_continue"},
	])
	assert_true(completed[0], "per-step on_fail should override cutscene_fallback")
	await _cleanup(seq)

# UT-CUT15-08: global default is halt when no policy set.
func test_08_default_is_halt() -> void:
	var seq := _make_seq()
	seq.cutscene_fallback = ""  # clear fallback → default to halt
	var halted := [false]
	seq.sequence_failed.connect(func(_r: String) -> void: halted[0] = true)
	await seq.run([
		{"type": "fail", "id": "bad_step"},
	])
	assert_true(halted[0], "when no policy set, default should be halt")
	await _cleanup(seq)
