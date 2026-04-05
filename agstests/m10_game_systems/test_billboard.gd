## UT-M10-90..97 — AGSBillboardController direction selection (T-GS26)
##
## Tests cover the static velocity_to_row() math: 1-way, 4-way, and 8-way
## direction quantization with various velocity/camera combinations.
## The scene-tree-dependent physics_process behaviour is covered by the
## manual test checklist in TODO_TESTS.md.
extends "res://utils/test_base.gd"

const CTRL_SCRIPT := "res://m10_game_systems/runtime/ags_billboard_controller.gd"

func suite_name() -> String:
	return "M10: BillboardController"


# Convenience: load the script and call the static method.
func _row(vel: Vector2, cam: Vector2, angles: int) -> int:
	var script := load(CTRL_SCRIPT) as GDScript
	return script.velocity_to_row(vel, cam, angles)


# ── 1-way ─────────────────────────────────────────────────────────────────────

# UT-M10-90: 1-way always returns row 0 regardless of direction.
func test_90_one_way_always_row_0() -> void:
	var cam := Vector2(0.0, -1.0)
	assert_eq(_row(Vector2(0, -1), cam, 1), 0, "1-way forward: row 0")
	assert_eq(_row(Vector2(1, 0),  cam, 1), 0, "1-way right:   row 0")
	assert_eq(_row(Vector2(0, 1),  cam, 1), 0, "1-way back:    row 0")
	assert_eq(_row(Vector2(-1, 0), cam, 1), 0, "1-way left:    row 0")


# ── 4-way ─────────────────────────────────────────────────────────────────────

# UT-M10-91: 4-way — camera at +Z looking at origin (cam_fwd = (0,-1)).
# Moving in same direction as camera (away) = N = row 0.
func test_91_four_way_cardinal_directions() -> void:
	# Camera is at +Z side, looking toward origin: cam_fwd = (0, -1).
	var cam := Vector2(0.0, -1.0)
	assert_eq(_row(Vector2(0, -1), cam, 4), 0, "N (away from camera) = row 0")
	assert_eq(_row(Vector2(1, 0),  cam, 4), 1, "E = row 1")
	assert_eq(_row(Vector2(0, 1),  cam, 4), 2, "S (toward camera) = row 2")
	assert_eq(_row(Vector2(-1, 0), cam, 4), 3, "W = row 3")


# UT-M10-92: 4-way — camera rotated 90° (cam_fwd = (-1, 0), camera at +X).
func test_92_four_way_rotated_camera() -> void:
	# Camera at +X looking toward origin: cam_fwd = (-1, 0).
	# Camera's right = world -Z (Vector2(0, -1)).
	var cam := Vector2(-1.0, 0.0)
	# Moving in world -X = away from camera = N = row 0.
	assert_eq(_row(Vector2(-1, 0), cam, 4), 0, "world -X = cam-N = row 0")
	# Moving in world -Z = camera-right = E = row 1.
	assert_eq(_row(Vector2(0, -1), cam, 4), 1, "world -Z = cam-E = row 1")


# UT-M10-93: 4-way — diagonal velocity snaps to nearest cardinal.
func test_93_four_way_diagonal_snaps() -> void:
	var cam := Vector2(0.0, -1.0)
	# NE diagonal — closer to N than E? At exactly 45° it should snap to E (boundary).
	# 44° from N → still N.
	var near_n := Vector2(sin(deg_to_rad(44)), -cos(deg_to_rad(44))).normalized()
	assert_eq(_row(near_n, cam, 4), 0, "44° from N snaps to N (row 0)")
	# 46° from N → E.
	var near_e := Vector2(sin(deg_to_rad(46)), -cos(deg_to_rad(46))).normalized()
	assert_eq(_row(near_e, cam, 4), 1, "46° from N snaps to E (row 1)")


# ── 8-way ─────────────────────────────────────────────────────────────────────

# UT-M10-94: 8-way — camera at +Z, all 8 cardinal/diagonal directions.
func test_94_eight_way_all_directions() -> void:
	var cam := Vector2(0.0, -1.0)
	var cases := [
		[Vector2(0, -1),                              0],  # N
		[Vector2(1, -1).normalized(),                 1],  # NE
		[Vector2(1, 0),                               2],  # E
		[Vector2(1, 1).normalized(),                  3],  # SE
		[Vector2(0, 1),                               4],  # S
		[Vector2(-1, 1).normalized(),                 5],  # SW
		[Vector2(-1, 0),                              6],  # W
		[Vector2(-1, -1).normalized(),                7],  # NW
	]
	for c in cases:
		var row: int = _row(c[0], cam, 8)
		assert_eq(row, c[1], "8-way vel=%s expected row %d" % [str(c[0]), c[1]])


# UT-M10-95: 8-way — boundary snapping: velocity at exact 22.5° boundary.
func test_95_eight_way_boundary_snapping() -> void:
	var cam := Vector2(0.0, -1.0)
	# 22° from N → still N (row 0).
	var near_n := Vector2(sin(deg_to_rad(22)), -cos(deg_to_rad(22))).normalized()
	assert_eq(_row(near_n, cam, 8), 0, "22° from N → N (row 0)")
	# 23° from N → NE (row 1).
	var near_ne := Vector2(sin(deg_to_rad(23)), -cos(deg_to_rad(23))).normalized()
	assert_eq(_row(near_ne, cam, 8), 1, "23° from N → NE (row 1)")


# ── Wrap-around ───────────────────────────────────────────────────────────────

# UT-M10-96: Angle wraps correctly at the 0°/360° boundary.
# 8-way NW bucket = [292.5°, 337.5°).  [337.5°, 360°) wraps back to N.
func test_96_wrap_around_boundary() -> void:
	var cam := Vector2(0.0, -1.0)
	# -22° = 338° — inside [337.5°, 360°) → wraps to N (row 0).
	var n_wrap := Vector2(sin(deg_to_rad(-22)), -cos(deg_to_rad(-22))).normalized()
	assert_eq(_row(n_wrap, cam, 8), 0, "338° wraps to N (row 0)")
	# -23° = 337° — inside [292.5°, 337.5°) → NW (row 7).
	var nw := Vector2(sin(deg_to_rad(-23)), -cos(deg_to_rad(-23))).normalized()
	assert_eq(_row(nw, cam, 8), 7, "337° = NW (row 7)")


# ── Instantiation ─────────────────────────────────────────────────────────────

# UT-M10-97: AGSBillboardController instantiates without crash.
func test_97_controller_instantiates() -> void:
	var script := load(CTRL_SCRIPT) as GDScript
	assert_not_null(script, "Failed to load ags_billboard_controller.gd")
	var node: Node = script.new()
	assert_not_null(node, "script.new() returned null")
	assert_eq(node.sprite_angles, 1, "sprite_angles defaults to 1")
	assert_eq(node.frames_per_angle, 1, "frames_per_angle defaults to 1")
	assert_eq(node.fps, 8.0, "fps defaults to 8.0")
	assert_eq(node.sprite_locked, false, "sprite_locked defaults to false")
	node.free()
