## T-CUT22 — Skip input routing tests.
##
## Tests verify:
##   - AGSSkipInput properties and defaults.
##   - AGSSequencer.request_skip() emits skip_requested.
##   - request_skip() is a no-op when no sequence is active.
##   - _advance_dialogue() and _request_cutscene_skip() helpers exist.
##   - Double press threshold is configurable.
extends "res://utils/test_base.gd"

const SkipInput := preload("res://../game_prototype/.engine/runtime/ags_skip_input.gd")
const SeqScript := preload("res://../game_prototype/.engine/runtime/ags_sequencer.gd")

func suite_name() -> String:
	return "M-CUT: SkipInput (T-CUT22)"


# ── Helpers ────────────────────────────────────────────────────────────────────

func _make_skip_input() -> Node:
	var s: Node = SkipInput.new()
	_tree.root.add_child(s)
	return s

func _make_seq() -> Node:
	var s: Node = SeqScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT22-01: AGSSkipInput instantiates ───────────────────────────────────

func test_01_skip_input_instantiates() -> void:
	var si := _make_skip_input()
	assert_not_null(si, "AGSSkipInput should instantiate")
	await _cleanup_nodes([si])


# ── UT-CUT22-02: default action names are set ────────────────────────────────

func test_02_default_action_names() -> void:
	var si := _make_skip_input()
	assert_eq(si.action_advance, "dialogue_advance", "default advance action should be 'dialogue_advance'")
	assert_eq(si.action_hold_advance, "dialogue_hold_advance", "default hold action should be 'dialogue_hold_advance'")
	assert_eq(si.action_cutscene_skip, "cutscene_skip", "default skip action should be 'cutscene_skip'")
	await _cleanup_nodes([si])


# ── UT-CUT22-03: double press threshold defaults to 0.3 ──────────────────────

func test_03_double_press_threshold_default() -> void:
	var si := _make_skip_input()
	assert_eq(si.double_press_threshold, 0.3, "double press threshold should default to 0.3")
	await _cleanup_nodes([si])


# ── UT-CUT22-04: hold repeat interval defaults to 0.1 ────────────────────────

func test_04_hold_repeat_interval_default() -> void:
	var si := _make_skip_input()
	assert_eq(si.hold_repeat_interval, 0.1, "hold repeat interval should default to 0.1")
	await _cleanup_nodes([si])


# ── UT-CUT22-05: request_skip() no-op when inactive ──────────────────────────

func test_05_request_skip_noop_when_inactive() -> void:
	var seq := _make_seq()
	var fired := [false]
	seq.skip_requested.connect(func() -> void: fired[0] = true)

	seq.request_skip()

	assert_false(fired[0], "request_skip() should be a no-op when no sequence is active")
	await _cleanup_nodes([seq])


# ── UT-CUT22-06: request_skip() emits skip_requested when active ──────────────

func test_06_request_skip_emits_signal_when_active() -> void:
	var seq := _make_seq()
	var fired := [false]
	seq.skip_requested.connect(func() -> void: fired[0] = true)

	# Manually set _active to simulate a running sequence.
	seq._active = true
	seq.request_skip()
	seq._active = false

	assert_true(fired[0], "request_skip() should emit skip_requested when sequence is active")
	await _cleanup_nodes([seq])


# ── UT-CUT22-07: skip_requested signal exists on sequencer ───────────────────

func test_07_skip_requested_signal_exists() -> void:
	var seq := _make_seq()
	assert_true(seq.has_signal("skip_requested"), "AGSSequencer should have skip_requested signal")
	await _cleanup_nodes([seq])


# ── UT-CUT22-08: _advance_dialogue helper exists ─────────────────────────────

func test_08_advance_dialogue_helper_exists() -> void:
	var si := _make_skip_input()
	assert_true(si.has_method("_advance_dialogue"), "_advance_dialogue() should exist")
	await _cleanup_nodes([si])


# ── UT-CUT22-09: _request_cutscene_skip helper exists ────────────────────────

func test_09_request_cutscene_skip_helper_exists() -> void:
	var si := _make_skip_input()
	assert_true(si.has_method("_request_cutscene_skip"), "_request_cutscene_skip() should exist")
	await _cleanup_nodes([si])


# ── UT-CUT22-10: _request_cutscene_skip no crash without sequencer ───────────

func test_10_request_skip_no_crash_no_sequencer() -> void:
	note("No AGSSequencer AutoLoad in agstests — _request_cutscene_skip() should not crash")
	var si := _make_skip_input()

	si._request_cutscene_skip()  # Should be a no-op, not a crash.

	assert_not_null(si, "_request_cutscene_skip() should not destroy the node")
	await _cleanup_nodes([si])


# ── UT-CUT22-11: _advance_dialogue no crash without AGSDialogue ──────────────

func test_11_advance_dialogue_no_crash_no_dlg() -> void:
	note("No AGSDialogue AutoLoad in agstests — _advance_dialogue() should not crash")
	var si := _make_skip_input()

	si._advance_dialogue()  # Should be a no-op.

	assert_not_null(si, "_advance_dialogue() should not destroy the node")
	await _cleanup_nodes([si])


# ── UT-CUT22-12: double_press_threshold is configurable ──────────────────────

func test_12_double_press_threshold_configurable() -> void:
	var si := _make_skip_input()
	si.double_press_threshold = 0.5
	assert_eq(si.double_press_threshold, 0.5, "double_press_threshold should be settable")
	await _cleanup_nodes([si])
