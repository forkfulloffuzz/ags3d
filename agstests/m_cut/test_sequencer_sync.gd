## T-CUT13 — AGSSequencer sync point tests.
extends "res://utils/test_base.gd"

const SeqScript = preload("res://../game_prototype/.engine/runtime/ags_sequencer.gd")

func suite_name() -> String:
	return "M-CUT: SequencerSync"

func _make_seq() -> Node:
	var s: Node = SeqScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup(s: Node) -> void:
	s.queue_free()
	await _tree.process_frame

# UT-CUT13-01: <<sync>> with no ids waits for all active backgrounds.
func test_01_sync_all_waits_all_bg() -> void:
	var seq := _make_seq()
	var done_ids: Array[String] = []
	seq.step_complete.connect(func(id: String) -> void: done_ids.append(id))
	await seq.run([
		{"type": "wait", "duration": 0.05, "bg": "bg_a"},
		{"type": "sync", "ids": []},  # wait for all
	])
	assert_true(done_ids.has("bg_a"), "sync all should wait for bg_a to complete")
	await _cleanup(seq)

# UT-CUT13-02: <<sync id>> waits for the named id only.
func test_02_sync_named_id() -> void:
	var seq := _make_seq()
	var done_ids: Array[String] = []
	seq.step_complete.connect(func(id: String) -> void: done_ids.append(id))
	await seq.run([
		{"type": "wait", "duration": 0.05, "bg": "bg_a"},
		{"type": "wait", "duration": 0.2, "bg": "bg_b"},  # longer, not synced explicitly
		{"type": "sync", "ids": ["bg_a"]},
	])
	assert_true(done_ids.has("bg_a"), "sync named should wait for bg_a")
	await _cleanup(seq)

# UT-CUT13-03: sync over an already-complete step passes immediately.
func test_03_sync_over_complete_passes() -> void:
	var seq := _make_seq()
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true)
	await seq.run([
		{"type": "wait", "duration": 0.0, "bg": "bg_a"},
		{"type": "wait", "duration": 0.05},  # give bg_a time to finish
		{"type": "sync", "ids": ["bg_a"]},   # should pass immediately
	])
	assert_true(completed[0], "sync over already-complete step should not hang")
	await _cleanup(seq)

# UT-CUT13-04: <<sync>> with multiple named ids waits for all of them.
func test_04_sync_multiple_ids() -> void:
	var seq := _make_seq()
	var done_ids: Array[String] = []
	seq.step_complete.connect(func(id: String) -> void: done_ids.append(id))
	await seq.run([
		{"type": "wait", "duration": 0.03, "bg": "bg_x"},
		{"type": "wait", "duration": 0.05, "bg": "bg_y"},
		{"type": "sync", "ids": ["bg_x", "bg_y"]},
	])
	assert_true(done_ids.has("bg_x"), "bg_x should complete before sync finishes")
	assert_true(done_ids.has("bg_y"), "bg_y should complete before sync finishes")
	await _cleanup(seq)
