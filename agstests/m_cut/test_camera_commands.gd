## T-CUT16 — AGSSequencerCommands camera command tests.
##
## NOTE: get_viewport() returns null in headless mode, so _get_active_camera()
## returns null for most commands. Tests verify graceful handling (no crash,
## step succeeds as no-op) and test the parts that DO work headless:
## fov/set on a directly-provided camera name via AGSRuntime.
extends "res://utils/test_base.gd"

const SeqCmds := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")

func suite_name() -> String:
	return "M-CUT: CameraCommands (T-CUT16)"


# ── Helpers ────────────────────────────────────────────────────────────────────

func _make_seq() -> Node:
	var s: Node = SeqCmds.new()
	_tree.root.add_child(s)
	return s

func _make_camera(cam_name: String) -> Node:
	var cam: Node = AGSCamera.new()
	cam.camera_name = cam_name
	_tree.root.add_child(cam)  # _ready fires → registers with AGSRuntime
	return cam

func _make_room(room_name: String) -> Node:
	var room: Node = AGSRoom.new()
	room.room_name = room_name
	_tree.root.add_child(room)
	return room

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


# ── UT-CUT16-01: set — no crash when no camera found (headless) ───────────────

func test_01_set_no_crash_headless() -> void:
	note("get_viewport() is null headless — camera 'set' is a no-op but must not crash")
	var seq := _make_seq()

	var step := {"type": "camera", "command": "set", "point": "nonexistent_camera"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete even when camera not found")
	await _cleanup_nodes([seq])


# ── UT-CUT16-02: fov instant — no crash when no active camera (headless) ─────

func test_02_fov_instant_no_crash_headless() -> void:
	note("get_viewport() is null headless — fov command is a no-op but must not crash")
	var seq := _make_seq()

	var step := {"type": "camera", "command": "fov", "value": 90.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete even when camera not found")
	await _cleanup_nodes([seq])


# ── UT-CUT16-03: unknown command is a no-op ───────────────────────────────────

func test_03_unknown_command_noop() -> void:
	var seq := _make_seq()

	var step := {"type": "camera", "command": "do_a_barrel_roll"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "unknown camera command should complete as no-op")
	await _cleanup_nodes([seq])


# ── UT-CUT16-04: set with camera registered in AGSRuntime ─────────────────────

func test_04_set_activates_named_camera() -> void:
	var seq := _make_seq()
	var cam := _make_camera("cam_16_04")

	var step := {"type": "camera", "command": "set", "point": "cam_16_04"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "set should succeed with registered camera")
	# Camera was activated via AGSRuntime.set_camera — it should now be current.
	assert_true((cam as Camera3D).current, "camera should be current after set")
	await _cleanup_nodes([seq, cam])


# ── UT-CUT16-05: fov instant with named camera changes its FOV ────────────────

func test_05_fov_instant_with_named_camera() -> void:
	var seq := _make_seq()
	var cam := _make_camera("cam_16_05")
	(cam as Camera3D).fov = 60.0

	var step := {"type": "camera", "camera": "cam_16_05", "command": "fov", "value": 95.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "fov step should complete")
	assert_eq((cam as Camera3D).fov, 95.0, "camera FOV should be updated to 95.0")
	await _cleanup_nodes([seq, cam])


# ── UT-CUT16-06: fov with duration tweens and completes ──────────────────────

func test_06_fov_tween_completes() -> void:
	var seq := _make_seq()
	var cam := _make_camera("cam_16_06")
	(cam as Camera3D).fov = 60.0

	var step := {"type": "camera", "camera": "cam_16_06", "command": "fov", "value": 80.0, "duration": 0.05}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "fov tween step should complete")
	assert_eq((cam as Camera3D).fov, 80.0, "camera FOV should reach 80.0 after tween")
	await _cleanup_nodes([seq, cam])


# ── UT-CUT16-07: move_to with no target camera — no crash ────────────────────

func test_07_move_to_no_target_no_crash() -> void:
	note("move_to with missing target camera — returns immediately without error")
	var seq := _make_seq()

	var step := {"type": "camera", "command": "move_to", "point": "ghost_cam", "duration": 0.05}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "move_to with no target should complete gracefully")
	await _cleanup_nodes([seq])


# ── UT-CUT16-08: move_to tweens camera position ──────────────────────────────

func test_08_move_to_tweens_position() -> void:
	var seq := _make_seq()
	var source_cam := _make_camera("cam_16_08_src")
	var target_cam := _make_camera("cam_16_08_tgt")
	(source_cam as Camera3D).global_position = Vector3(0.0, 0.0, 0.0)
	(target_cam as Camera3D).global_position = Vector3(5.0, 2.0, -3.0)
	# Activate source camera so it's the "active" camera via AGSRuntime.
	Engine.get_singleton("AGSRuntime").call("set_camera", "cam_16_08_src")

	var step := {
		"type": "camera", "camera": "cam_16_08_src",
		"command": "move_to", "point": "cam_16_08_tgt", "duration": 0.05
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "move_to step should complete")
	assert_eq(
		(source_cam as Camera3D).global_position,
		Vector3(5.0, 2.0, -3.0),
		"source camera should reach target position"
	)
	await _cleanup_nodes([seq, source_cam, target_cam])


# ── UT-CUT16-09: shake — no crash when no active camera (headless) ───────────

func test_09_shake_no_crash_headless() -> void:
	note("get_viewport() is null headless — shake is a no-op but must not crash")
	var seq := _make_seq()

	var step := {"type": "camera", "command": "shake", "intensity": 0.1, "duration": 0.02}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "shake should complete gracefully in headless")
	await _cleanup_nodes([seq])


# ── UT-CUT16-10: shake with named camera moves and restores position ──────────

func test_10_shake_with_camera_restores_origin() -> void:
	var seq := _make_seq()
	var cam := _make_camera("cam_16_10")
	(cam as Camera3D).global_position = Vector3(1.0, 2.0, 3.0)
	Engine.get_singleton("AGSRuntime").call("set_camera", "cam_16_10")

	var step := {
		"type": "camera", "camera": "cam_16_10",
		"command": "shake", "intensity": 0.5, "duration": 0.04, "falloff": true
	}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "shake should complete")
	assert_eq(
		(cam as Camera3D).global_position,
		Vector3(1.0, 2.0, 3.0),
		"camera position should be restored to origin after shake"
	)
	await _cleanup_nodes([seq, cam])


# ── UT-CUT16-11: return — no crash when no room / initial camera ──────────────

func test_11_return_no_crash_no_room() -> void:
	note("No current room registered — camera 'return' is a no-op but must not crash")
	var seq := _make_seq()

	var step := {"type": "camera", "command": "return", "duration": 0.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "'return' without room should complete gracefully")
	await _cleanup_nodes([seq])


# ── UT-CUT16-12: look_at — no crash when camera not found (headless) ─────────

func test_12_look_at_no_crash_headless() -> void:
	note("get_viewport() is null headless — look_at is a no-op but must not crash")
	var seq := _make_seq()

	var step := {"type": "camera", "command": "look_at", "target": "nowhere", "duration": 0.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "look_at should complete gracefully in headless")
	await _cleanup_nodes([seq])


# ── UT-CUT16-13: set with fov override applies both ──────────────────────────

func test_13_set_with_fov_override() -> void:
	var seq := _make_seq()
	var cam := _make_camera("cam_16_13")
	(cam as Camera3D).fov = 60.0

	var step := {"type": "camera", "command": "set", "point": "cam_16_13", "fov": 110.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "set with fov override should complete")
	assert_eq((cam as Camera3D).fov, 110.0, "camera FOV should be overridden to 110.0")
	await _cleanup_nodes([seq, cam])
