## T-CUT12 — AGSSequencer core sequencer tests.
extends "res://utils/test_base.gd"

const SeqScript = preload("res://../../game_prototype/.engine/runtime/ags_sequencer.gd")

func suite_name() -> String:
	return "M-CUT: Sequencer"

func _make_seq() -> Node:
	var s: Node = SeqScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup(s: Node) -> void:
	s.queue_free()
	await _tree.process_frame

# UT-CUT12-01: is_active is false before run().
func test_01_inactive_initially() -> void:
	var seq := _make_seq()
	assert_false(seq.is_active(), "Sequencer should be inactive before run()")
	await _cleanup(seq)

# UT-CUT12-02: sequence_complete fires after an empty step list.
func test_02_empty_steps_completes() -> void:
	var seq := _make_seq()
	var completed: bool = false
	seq.sequence_complete.connect(func() -> void: completed = true)
	await seq.run([])
	assert_true(completed, "sequence_complete should fire after empty run()")
	await _cleanup(seq)

# UT-CUT12-03: is_active is false after run() returns.
func test_03_inactive_after_run() -> void:
	var seq := _make_seq()
	await seq.run([])
	assert_false(seq.is_active(), "Sequencer should be inactive after run() returns")
	await _cleanup(seq)

# UT-CUT12-04: wait step pauses execution for approximately the given duration.
func test_04_wait_step() -> void:
	var seq := _make_seq()
	var completed: bool = false
	seq.sequence_complete.connect(func() -> void: completed = true)
	await seq.run([{"type": "wait", "duration": 0.05}])
	assert_true(completed, "sequence_complete should fire after wait step")
	await _cleanup(seq)

# UT-CUT12-05: step_started and step_complete emit for foreground steps.
func test_05_step_signals_emitted() -> void:
	var seq := _make_seq()
	var started: Array[String] = []
	var completed_ids: Array[String] = []
	seq.step_started.connect(func(id: String) -> void: started.append(id))
	seq.step_complete.connect(func(id: String) -> void: completed_ids.append(id))
	await seq.run([{"type": "wait", "duration": 0.0, "id": "my_wait"}])
	assert_true(started.has("my_wait"), "step_started should emit with step id")
	assert_true(completed_ids.has("my_wait"), "step_complete should emit with step id")
	await _cleanup(seq)

# UT-CUT12-06: background step fires without blocking sequence.
func test_06_bg_step_nonblocking() -> void:
	var seq := _make_seq()
	var order: Array[String] = []
	seq.step_started.connect(func(id: String) -> void: order.append("start:" + id))
	seq.step_complete.connect(func(id: String) -> void: order.append("done:" + id))
	# bg step with longer wait, then a foreground step — foreground should complete before bg.
	var steps := [
		{"type": "wait", "duration": 0.1, "bg": "slow_bg"},
		{"type": "wait", "duration": 0.0, "id": "fast_fg"},
	]
	await seq.run(steps)
	# fast_fg should complete before slow_bg.
	var fg_idx := order.find("done:fast_fg")
	var bg_idx := order.find("done:slow_bg")
	assert_true(fg_idx >= 0, "fast_fg should complete")
	assert_true(bg_idx >= 0, "slow_bg should complete")
	assert_true(fg_idx < bg_idx, "fast_fg should complete before slow_bg")
	await _cleanup(seq)

# UT-CUT12-07: <<end>> step terminates the sequence early.
func test_07_end_step_terminates() -> void:
	var seq := _make_seq()
	var completed: bool = false
	seq.sequence_complete.connect(func() -> void: completed = true)
	var steps := [
		{"type": "end"},
		{"type": "wait", "duration": 10.0},  # should never run
	]
	await seq.run(steps)
	assert_true(completed, "sequence should complete after <<end>>")
	await _cleanup(seq)

# UT-CUT12-08: sync step with no ids waits for all background steps.
func test_08_sync_all_waits_for_bg() -> void:
	var seq := _make_seq()
	var bg_done: bool = false
	seq.step_complete.connect(func(id: String) -> void:
		if id == "bg1": bg_done = true
	)
	var steps := [
		{"type": "wait", "duration": 0.05, "bg": "bg1"},
		{"type": "sync", "ids": []},
	]
	await seq.run(steps)
	assert_true(bg_done, "sync all should wait for bg1 to complete")
	await _cleanup(seq)

# UT-CUT12-09: sync step with named id waits only for that id.
func test_09_sync_named_waits_for_id() -> void:
	var seq := _make_seq()
	var bg_done: bool = false
	seq.step_complete.connect(func(id: String) -> void:
		if id == "target_bg": bg_done = true
	)
	var steps := [
		{"type": "wait", "duration": 0.05, "bg": "target_bg"},
		{"type": "sync", "ids": ["target_bg"]},
	]
	await seq.run(steps)
	assert_true(bg_done, "named sync should wait for target_bg")
	await _cleanup(seq)

# UT-CUT12-10: run() while already active is ignored (no crash, no double-complete).
func test_10_reentrant_run_ignored() -> void:
	var seq := _make_seq()
	var complete_count: int = 0
	seq.sequence_complete.connect(func() -> void: complete_count += 1)
	var steps := [{"type": "wait", "duration": 0.05}]
	# Start first run in background (don't await yet).
	var _ = seq.run(steps)
	# Immediately try a second run — should be ignored.
	await seq.run([])  # this should be the ignored one
	await get_tree().create_timer(0.1).timeout
	assert_eq(complete_count, 1, "Only first run should complete (second ignored)")
	await _cleanup(seq)
