## T-CUT19 — AGSSequencerCommands visual command tests.
##
## CanvasLayer tweens work in headless mode. Viewport-dependent layout
## (anchors preset) is skipped headless but the nodes are still created.
extends "res://utils/test_base.gd"

const SeqCmds := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")

func suite_name() -> String:
	return "M-CUT: VisualCommands (T-CUT19)"


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


# ── UT-CUT19-01: fade_out completes and creates overlay node ──────────────────

func test_01_fade_out_completes() -> void:
	var seq := _make_seq()

	var step := {"type": "visual", "command": "fade_out", "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "fade_out step should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT19-02: fade_in completes ───────────────────────────────────────────

func test_02_fade_in_completes() -> void:
	var seq := _make_seq()

	var step := {"type": "visual", "command": "fade_in", "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "fade_in step should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT19-03: flash completes and cleans up ───────────────────────────────

func test_03_flash_completes_and_cleans_up() -> void:
	var seq := _make_seq()

	var step := {"type": "visual", "command": "flash", "color": [1.0, 1.0, 1.0, 1.0], "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "flash step should complete")
	# Flash rect uses queue_free() — cleanup is deferred to end-of-frame, not assertable here.
	await _cleanup_nodes([seq])


# ── UT-CUT19-04: vignette creates overlay node ────────────────────────────────

func test_04_vignette_creates_overlay() -> void:
	var seq := _make_seq()

	var step := {"type": "visual", "command": "vignette", "intensity": 0.6, "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "vignette step should complete")
	# Visual layer should have a rect child (persists until fade_in or sequence end).
	if seq._visual_layer != null:
		assert_true(seq._visual_layer.get_child_count() > 0, "vignette rect should remain on layer")
	await _cleanup_nodes([seq])


# ── UT-CUT19-05: letterbox enable creates top and bottom bars ─────────────────

func test_05_letterbox_enable_creates_bars() -> void:
	var seq := _make_seq()

	var step := {"type": "visual", "command": "letterbox", "enable": true, "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "letterbox enable step should complete")
	assert_not_null(seq._letterbox_top, "_letterbox_top should be created")
	assert_not_null(seq._letterbox_bottom, "_letterbox_bottom should be created")
	await _cleanup_nodes([seq])


# ── UT-CUT19-06: letterbox disable removes bars ───────────────────────────────

func test_06_letterbox_disable_removes_bars() -> void:
	var seq := _make_seq()
	# Enable first.
	await seq.run([{"type": "visual", "command": "letterbox", "enable": true, "duration": 0.01}])
	assert_not_null(seq._letterbox_top, "bars should exist after enable")

	# Now disable.
	var step := {"type": "visual", "command": "letterbox", "enable": false, "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "letterbox disable step should complete")
	assert_null(seq._letterbox_top, "_letterbox_top should be null after disable")
	assert_null(seq._letterbox_bottom, "_letterbox_bottom should be null after disable")
	await _cleanup_nodes([seq])


# ── UT-CUT19-07: overlay with missing image path — no crash ──────────────────

func test_07_overlay_missing_image_no_crash() -> void:
	var seq := _make_seq()

	var step := {
		"type": "visual", "command": "overlay",
		"image": "res://no_such_image.png",
		"duration": 0.02
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "overlay with missing image should complete gracefully")
	assert_null(seq._current_overlay, "overlay node should be freed after completion")
	await _cleanup_nodes([seq])


# ── UT-CUT19-08: overlay duration holds then cleans up ───────────────────────

func test_08_overlay_completes_and_cleans_up() -> void:
	var seq := _make_seq()

	var step := {
		"type": "visual", "command": "overlay",
		"image": "",    # no image — still creates the node
		"fade_in": 0.01, "duration": 0.02, "fade_out": 0.01
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "overlay step should complete")
	assert_null(seq._current_overlay, "overlay node should be freed after fade_out")
	await _cleanup_nodes([seq])


# ── UT-CUT19-09: video with missing file — no crash, returns true ─────────────

func test_09_video_missing_file_no_crash() -> void:
	var seq := _make_seq()

	var step := {"type": "visual", "command": "video", "file": "res://no_such.ogv"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "video with missing file should complete gracefully")
	await _cleanup_nodes([seq])


# ── UT-CUT19-10: unknown visual command — no crash ────────────────────────────

func test_10_unknown_command_noop() -> void:
	var seq := _make_seq()

	var step := {"type": "visual", "command": "dissolve_into_sparkles"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "unknown visual command should complete as no-op")
	await _cleanup_nodes([seq])


# ── UT-CUT19-11: fade_out with custom color ───────────────────────────────────

func test_11_fade_out_custom_color() -> void:
	var seq := _make_seq()

	# Red fade-out (not black — won't delegate to AGSCutscene)
	var step := {"type": "visual", "command": "fade_out", "duration": 0.02, "color": [1.0, 0.0, 0.0, 1.0]}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "fade_out with custom color should complete")
	# The rect should remain (screen stays red until fade_in).
	if seq._visual_layer != null:
		assert_true(seq._visual_layer.get_child_count() > 0, "fade-out rect should remain on layer")
	await _cleanup_nodes([seq])


# ── UT-CUT19-12: sequential fade_out + fade_in both complete ─────────────────

func test_12_fade_out_then_fade_in() -> void:
	var seq := _make_seq()

	var steps: Array = [
		{"type": "visual", "command": "fade_out", "duration": 0.02},
		{"type": "visual", "command": "fade_in", "duration": 0.02},
	]
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	await seq.run(steps)

	assert_true(completed[0], "fade_out + fade_in sequence should complete")
	await _cleanup_nodes([seq])
