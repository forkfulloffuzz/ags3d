## T-CUT25 — Save blocking during cutscenes.
##
## Verifies that AGSSaveLoad queues saves when a cutscene is active, and
## executes the queued save once the cutscene finishes.
extends "res://utils/test_base.gd"

const SaveLoad := preload("res://../game_prototype/.engine/runtime/ags_save_load.gd")
const SeqCmds  := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")

func suite_name() -> String:
	return "M-CUT: SaveBlocking (T-CUT25)"


# ── Helpers ────────────────────────────────────────────────────────────────────

## Create a sequencer and a save-load node linked to it.
func _make_pair() -> Array:
	var seq: Node = SeqCmds.new()
	var sl: Node = SaveLoad.new()
	_tree.root.add_child(seq)
	_tree.root.add_child(sl)
	# Connect the save-load node to the sequencer manually (bypasses AutoLoad order).
	seq.sequence_complete.connect(sl._on_sequence_ended)
	seq.sequence_failed.connect(sl._on_sequence_ended_reason)
	# Inject a reference so _get_sequencer() finds it via the tree.
	sl.set_meta("_seq_override", seq)
	return [seq, sl]

func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT25-01: save_blocked property false when no sequence active ──────────

func test_01_save_not_blocked_when_idle() -> void:
	var pair: Array = _make_pair()
	var sl: Node = pair[1]
	assert_false(sl.save_blocked, "save_blocked should be false when no sequence is active")
	await _cleanup_nodes(pair)


# ── UT-CUT25-02: save_blocked property true while sequence active ─────────────

func test_02_save_blocked_while_sequence_active() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	var blocked_during := [false]
	# Start a long-running sequence in the background.
	seq.run([{"type": "wait", "duration": 10.0}])
	# Yield one frame so run() has had time to set _active.
	await _tree.process_frame

	blocked_during[0] = sl.save_blocked
	seq.request_skip()
	await _tree.process_frame

	assert_true(blocked_during[0], "save_blocked should be true while sequence is active")
	await _cleanup_nodes(pair)


# ── UT-CUT25-03: save_game returns false when blocked ────────────────────────

func test_03_save_game_returns_false_when_blocked() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	seq.run([{"type": "wait", "duration": 10.0}])
	await _tree.process_frame

	var result: bool = sl.save_game(1)
	assert_false(result, "save_game should return false while a cutscene is active")

	seq.request_skip()
	await _tree.process_frame
	await _cleanup_nodes(pair)


# ── UT-CUT25-04: save_blocked signal emitted on blocked save attempt ──────────

func test_04_save_blocked_signal_emitted() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	seq.run([{"type": "wait", "duration": 10.0}])
	await _tree.process_frame

	var blocked_slot := [-1]
	sl.save_blocked.connect(func(s: int) -> void: blocked_slot[0] = s, CONNECT_ONE_SHOT)
	sl.save_game(42)

	assert_eq(blocked_slot[0], 42, "save_blocked signal should carry the attempted slot number")

	seq.request_skip()
	await _tree.process_frame
	await _cleanup_nodes(pair)


# ── UT-CUT25-05: save_queued signal emitted on blocked save ──────────────────

func test_05_save_queued_signal_emitted() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	seq.run([{"type": "wait", "duration": 10.0}])
	await _tree.process_frame

	var queued_slot := [-1]
	sl.save_queued.connect(func(s: int) -> void: queued_slot[0] = s, CONNECT_ONE_SHOT)
	sl.save_game(7)

	assert_eq(queued_slot[0], 7, "save_queued signal should carry the queued slot number")

	seq.request_skip()
	await _tree.process_frame
	await _cleanup_nodes(pair)


# ── UT-CUT25-06: queued save executes after sequence_complete ────────────────

func test_06_queued_save_executes_on_sequence_complete() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	var completed_slot := [-1]
	sl.queued_save_completed.connect(func(s: int) -> void: completed_slot[0] = s, CONNECT_ONE_SHOT)

	# Start sequence and immediately try to save.
	seq.run([{"type": "wait", "duration": 10.0}])
	await _tree.process_frame
	sl.save_game(3)

	# Skip — sequence ends — queued save should fire.
	seq.request_skip()
	await seq.sequence_complete
	# One more frame for cleanup/signal propagation.
	await _tree.process_frame

	assert_eq(completed_slot[0], 3, "queued_save_completed should fire with slot 3 after sequence ends")
	await _cleanup_nodes(pair)


# ── UT-CUT25-07: newer queued slot replaces older one ────────────────────────

func test_07_newer_queue_replaces_older() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	seq.run([{"type": "wait", "duration": 10.0}])
	await _tree.process_frame

	sl.save_game(1)  # queued
	sl.save_game(2)  # should replace slot 1

	assert_eq(sl._queued_slot, 2, "newer save slot should replace the older queued one")

	seq.request_skip()
	await _tree.process_frame
	await _cleanup_nodes(pair)


# ── UT-CUT25-08: queued slot cleared after execution ─────────────────────────

func test_08_queued_slot_cleared_after_execution() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	seq.run([{"type": "wait", "duration": 10.0}])
	await _tree.process_frame
	sl.save_game(5)

	seq.request_skip()
	await seq.sequence_complete
	await _tree.process_frame

	assert_eq(sl._queued_slot, -1, "_queued_slot should be -1 after the queued save executes")
	await _cleanup_nodes(pair)


# ── UT-CUT25-09: save_game returns true and saves immediately when idle ────────

func test_09_save_game_returns_true_when_idle() -> void:
	var pair: Array = _make_pair()
	var sl: Node = pair[1]
	var saved := [false]

	# Intercept _do_save by checking _queued_slot stays -1.
	var result: bool = sl.save_game(99)
	# We can't check the actual file (AGSRuntime not available headless),
	# but the return value and queue state confirm the path taken.
	assert_true(result, "save_game should return true when no sequence is active")
	assert_eq(sl._queued_slot, -1, "_queued_slot should remain -1 on immediate save")
	await _cleanup_nodes(pair)


# ── UT-CUT25-10: queued save fires after sequence_failed ─────────────────────

func test_10_queued_save_fires_after_sequence_failed() -> void:
	var pair: Array = _make_pair()
	var seq: Node = pair[0]
	var sl: Node = pair[1]

	var completed_slot := [-1]
	sl.queued_save_completed.connect(func(s: int) -> void: completed_slot[0] = s, CONNECT_ONE_SHOT)

	# Use a fail step that triggers sequence_failed.
	seq.run([{"type": "wait", "duration": 10.0}, {"type": "fail"}])
	await _tree.process_frame
	sl.save_game(11)

	# Stop the sequence — triggers sequence_failed path.
	seq.stop()
	await _tree.process_frame

	assert_eq(completed_slot[0], 11,
		"queued_save_completed should fire with slot 11 after sequence_failed")
	await _cleanup_nodes(pair)
