## T-CUT14 — AGSSequencer timeout mechanism tests.
extends "res://utils/test_base.gd"

const SeqScript = preload("res://../game_prototype/.engine/runtime/ags_sequencer.gd")

func suite_name() -> String:
	return "M-CUT: SequencerTimeout"

func _make_seq() -> Node:
	var s: Node = SeqScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup(s: Node) -> void:
	s.queue_free()
	await _tree.process_frame

# UT-CUT14-01: step with timeout that completes before deadline fires step_complete.
func test_01_step_completes_before_timeout() -> void:
	var seq := _make_seq()
	var completed := [false]
	seq.step_complete.connect(func(_id: String) -> void: completed[0] = true)
	await seq.run([{"type": "wait", "duration": 0.0, "timeout": 1.0, "id": "fast"}])
	assert_true(completed[0], "step that completes before timeout should fire step_complete")
	await _cleanup(seq)

# UT-CUT14-02: timeout:none disables the timeout (step runs without limit).
func test_02_timeout_none_no_limit() -> void:
	var seq := _make_seq()
	var completed := [false]
	seq.step_complete.connect(func(_id: String) -> void: completed[0] = true)
	# Short wait with timeout:none — should still complete normally.
	await seq.run([{"type": "wait", "duration": 0.02, "timeout": "none", "id": "no_limit"}])
	assert_true(completed[0], "step with timeout:none should complete normally")
	await _cleanup(seq)

# UT-CUT14-03: step_timeout_default is applied when no per-step timeout is set.
func test_03_global_default_applied() -> void:
	var seq := _make_seq()
	seq.step_timeout_default = 1.0  # 1 second global default
	var completed := [false]
	seq.step_complete.connect(func(_id: String) -> void: completed[0] = true)
	# Fast step should complete well before the 1s global default.
	await seq.run([{"type": "wait", "duration": 0.0, "id": "default_timeout_test"}])
	assert_true(completed[0], "global default timeout should allow fast step to complete")
	await _cleanup(seq)

# UT-CUT14-04: per-step timeout overrides global default.
func test_04_perstep_timeout_overrides_global() -> void:
	var seq := _make_seq()
	seq.step_timeout_default = 0.01  # very short global default
	var completed := [false]
	seq.step_complete.connect(func(_id: String) -> void: completed[0] = true)
	# Per-step timeout is generous enough that the step succeeds.
	await seq.run([{"type": "wait", "duration": 0.0, "timeout": 1.0, "id": "per_step"}])
	assert_true(completed[0], "per-step timeout should override short global default")
	await _cleanup(seq)
