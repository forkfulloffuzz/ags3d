## ags_billboard_controller.gd — Billboard direction selection + frame cycling (T-GS26)
##
## Attach as a child of an AGSCharacter2D node. Each physics frame:
##   1. Reads the parent character's velocity vector.
##   2. Computes the angle between velocity and the current camera's forward
##      direction, projected onto the XZ plane.
##   3. Quantizes that angle to the nearest direction bucket (4-way or 8-way).
##   4. Advances the per-angle animation frame at [member fps].
##   5. Writes the combined (row * frames_per_angle + anim_frame) to Sprite3D.frame.
##
## Row layout in the sprite sheet (top row = index 0):
##   1-way:  row 0 only
##   4-way:  N(0), E(1), S(2), W(3)
##   8-way:  N(0), NE(1), E(2), SE(3), S(4), SW(5), W(6), NW(7)
##
## "N" = character moving away from the camera (back to camera).
## "S" = character moving toward the camera (front to camera).
##
## sprite_locked: when true, always use row 0 regardless of velocity/camera.
## Intended for rooms where the camera never orbits around the character and
## only one facing direction is needed.

extends Node

## Number of sprite angle variants. Must be 1, 4, or 8.
@export var sprite_angles: int = 1

## Number of animation frames per direction row.
@export var frames_per_angle: int = 1

## Animation playback speed in frames-per-second.
@export var fps: float = 8.0

## NodePath to the Sprite3D sibling/child (relative to this node's parent).
@export var sprite_path: NodePath = NodePath("../Sprite3D")

## When true, skip direction calculation and always use row 0.
## Use for single-angle sprites in rooms with a locked camera.
@export var sprite_locked: bool = false

## Column offset within each direction row.  Used by AGSAnimationPlayer2D to
## select which animation state (idle/walk/talk) is active.  The active frame
## cycles from frame_offset to frame_offset + frames_per_angle - 1.
var frame_offset: int = 0

var _sprite: Sprite3D = null
var _frame_timer: float = 0.0
var _current_anim_frame: int = 0
var _current_row: int = 0


func _ready() -> void:
	_sprite = get_node_or_null(sprite_path)
	if _sprite == null:
		push_error("AGSBillboardController: no Sprite3D found at path '%s'" % str(sprite_path))


func _physics_process(delta: float) -> void:
	if _sprite == null:
		return

	# --- animation frame cycling ---
	if frames_per_angle > 1:
		_frame_timer += delta
		if _frame_timer >= 1.0 / fps:
			_frame_timer = 0.0
			_current_anim_frame = (_current_anim_frame + 1) % frames_per_angle

	# --- direction selection ---
	if not sprite_locked and sprite_angles > 1:
		var parent := get_parent()
		if parent is CharacterBody3D:
			var vel: Vector3 = (parent as CharacterBody3D).velocity
			if vel.length_squared() > 0.01:
				var camera := get_viewport().get_camera_3d()
				var cam_fwd := Vector2(0.0, -1.0)  # fallback: camera at +Z
				if camera != null:
					var fwd := -camera.global_basis.z
					cam_fwd = Vector2(fwd.x, fwd.z)
					if cam_fwd.length_squared() > 0.0001:
						cam_fwd = cam_fwd.normalized()
				var vel_xz := Vector2(vel.x, vel.z)
				if vel_xz.length_squared() > 0.0001:
					vel_xz = vel_xz.normalized()
				_current_row = velocity_to_row(vel_xz, cam_fwd, sprite_angles)

	# hframes = total columns in sprite sheet.  Row width = hframes.
	var row_width: int = _sprite.hframes if _sprite.hframes > 0 else frames_per_angle
	_sprite.frame = _current_row * row_width + frame_offset + _current_anim_frame


## Compute the sprite sheet row for [param vel_xz] (normalised character
## velocity in XZ) given [param cam_fwd] (normalised camera forward in XZ)
## and [param angles] (1, 4, or 8).
##
## This is a static method so it can be tested directly without a scene tree.
static func velocity_to_row(vel_xz: Vector2, cam_fwd: Vector2, angles: int) -> int:
	if angles <= 1:
		return 0

	# Compute the signed clockwise angle from cam_fwd to vel_xz using
	# dot (cos) and cross product (sin).  This avoids the atan2(x,y) vs
	# atan2(y,x) ambiguity and naturally handles all quadrants.
	#
	# cross > 0 → vel is clockwise of cam_fwd (east / right)
	# cross < 0 → vel is counter-clockwise of cam_fwd (west / left)
	var dot  := cam_fwd.dot(vel_xz)
	var cross := cam_fwd.x * vel_xz.y - cam_fwd.y * vel_xz.x
	var rel  := atan2(cross, dot)         # range (-PI, PI]
	rel = fmod(rel + TAU, TAU)            # normalise to [0, TAU)

	match angles:
		4:
			# Bucket boundaries at 45°, 135°, 225°, 315°.
			return int(floor((rel + PI / 4.0) / (PI / 2.0))) % 4
		8:
			# Bucket boundaries at 22.5°, 67.5°, 112.5°, ...
			return int(floor((rel + PI / 8.0) / (PI / 4.0))) % 8
		_:
			return 0
