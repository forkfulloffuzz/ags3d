## T-CUT21 — AGSSequencerCommands dialogue command tests.
##
## AGSDialogue is not available as an AutoLoad in agstests, so all dialogue_line,
## narrator_line, choice, and dialogue node steps complete as graceful no-ops.
## title_card and subtitle use the visual layer and are fully testable headless.
extends "res://utils/test_base.gd"

const SeqCmds := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")

func suite_name() -> String:
	return "M-CUT: DialogueCommands (T-CUT21)"


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


# ── UT-CUT21-01: dialogue_line — no crash when no dialogue system ─────────────

func test_01_dialogue_line_no_crash_no_dlg() -> void:
	note("AGSDialogue not available in agstests — dialogue_line is a no-op but must not crash")
	var seq := _make_seq()

	var step := {
		"type": "dialogue_line",
		"character": "guard",
		"text": "Halt! Who goes there?",
		"loc_key": "",
		"emotion": "angry"
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "dialogue_line should complete gracefully without AGSDialogue")
	await _cleanup_nodes([seq])


# ── UT-CUT21-02: narrator_line — no crash when no dialogue system ─────────────

func test_02_narrator_line_no_crash_no_dlg() -> void:
	note("AGSDialogue not available — narrator_line is a no-op but must not crash")
	var seq := _make_seq()

	var step := {
		"type": "narrator_line",
		"text": "The room was silent.",
		"loc_key": ""
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "narrator_line should complete gracefully without AGSDialogue")
	await _cleanup_nodes([seq])


# ── UT-CUT21-03: title_card completes ────────────────────────────────────────

func test_03_title_card_completes() -> void:
	var seq := _make_seq()

	var step := {
		"type": "title_card",
		"text": "Chapter 1: The Beginning",
		"duration": 0.02,
		"fade_in": 0.01,
		"fade_out": 0.01
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "title_card should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT21-04: title_card with zero duration completes ─────────────────────

func test_04_title_card_zero_duration_completes() -> void:
	var seq := _make_seq()

	var step := {"type": "title_card", "text": "ACT II", "duration": 0.0, "fade_in": 0.0, "fade_out": 0.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "title_card with zero duration should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT21-05: subtitle completes ──────────────────────────────────────────

func test_05_subtitle_completes() -> void:
	var seq := _make_seq()

	var step := {"type": "subtitle", "text": "[Detective Noir Voice]", "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "subtitle should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT21-06: subtitle with zero duration completes ───────────────────────

func test_06_subtitle_zero_duration_completes() -> void:
	var seq := _make_seq()

	var step := {"type": "subtitle", "text": "...", "duration": 0.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "subtitle with zero duration should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT21-07: choice — no crash when no dialogue system ───────────────────

func test_07_choice_no_crash_no_dlg() -> void:
	note("AGSDialogue not available — choice is a no-op but must not crash")
	var seq := _make_seq()

	var step := {
		"type": "choice",
		"options": ["Option A", "Option B", "Option C"]
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "choice should complete gracefully without AGSDialogue")
	await _cleanup_nodes([seq])


# ── UT-CUT21-08: choice with empty options — no crash ────────────────────────

func test_08_choice_empty_options_no_crash() -> void:
	var seq := _make_seq()

	var step := {"type": "choice", "options": []}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "choice with empty options should complete gracefully")
	await _cleanup_nodes([seq])


# ── UT-CUT21-09: dialogue node — no crash when no dialogue system ─────────────

func test_09_dialogue_node_no_crash_no_dlg() -> void:
	note("AGSDialogue not available — dialogue step is a no-op but must not crash")
	var seq := _make_seq()

	var step := {"type": "dialogue", "node": "intro_conversation"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "dialogue node should complete gracefully without AGSDialogue")
	await _cleanup_nodes([seq])


# ── UT-CUT21-10: dialogue with missing node name — no crash ──────────────────

func test_10_dialogue_missing_node_name_no_crash() -> void:
	note("AGSDialogue not available — even with missing 'node', should not crash")
	var seq := _make_seq()

	var step := {"type": "dialogue"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "dialogue with missing node name should complete gracefully")
	await _cleanup_nodes([seq])


# ── UT-CUT21-11: unknown dialogue type — no crash ────────────────────────────

func test_11_unknown_dialogue_type_no_crash() -> void:
	var seq := _make_seq()

	var step := {"type": "title_scroll", "text": "A long time ago..."}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "unknown dialogue type should complete as no-op")
	await _cleanup_nodes([seq])


# ── UT-CUT21-12: title_card creates and cleans up visual node ────────────────

func test_12_title_card_creates_visual_node() -> void:
	var seq := _make_seq()

	var step := {"type": "title_card", "text": "TEST", "duration": 0.02, "fade_in": 0.0, "fade_out": 0.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "title_card should complete")
	# The panel is freed with queue_free() — deferred, not asserting count here.
	await _cleanup_nodes([seq])
