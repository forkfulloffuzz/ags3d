@tool
extends RefCounted

## .agroom ↔ .tscn sync (T-E12)
##
## Serialises the current state of an AGSRoom scene back to its .agroom file.
## Called after any undoable gizmo or inspector action on a room scene.
##
## Usage:
##   RoomSync.write_agroom(scene_root)


## Write the .agroom file that corresponds to [param root] (an AGSRoom node).
## The .agroom path is derived from the scene file path.
static func write_agroom(root: Node) -> Error:
	if root.get_class() != "AGSRoom":
		push_error("[AGS] RoomSync: root is not AGSRoom")
		return ERR_INVALID_PARAMETER

	var scene_path: String = root.scene_file_path
	if scene_path.is_empty():
		push_error("[AGS] RoomSync: scene has no file path")
		return ERR_FILE_NOT_FOUND

	var agroom_res: String = scene_path.get_basename() + ".agroom"
	var agroom_abs: String = ProjectSettings.globalize_path(agroom_res)

	var lines: Array[String] = _serialise(root)
	var fa := FileAccess.open(agroom_abs, FileAccess.WRITE)
	if not fa:
		push_error("[AGS] RoomSync: cannot write " + agroom_abs)
		return ERR_FILE_CANT_WRITE

	fa.store_string("\n".join(lines) + "\n")
	fa.close()
	return OK


# ---------------------------------------------------------------------------
# Serialiser
# ---------------------------------------------------------------------------

static func _serialise(root: Node) -> Array[String]:
	var lines: Array[String] = []
	var room_name: String = root.get("room_name")
	var initial_cam: String = root.get("initial_camera")

	lines.append('Room "%s" {' % room_name)
	lines.append('    initial_camera = "%s"' % initial_cam)

	for child in root.get_children():
		var cls: String = child.get_class()
		match cls:
			"AGSCamera":     lines.append_array(_camera(child))
			"AGSPoint":      lines.append_array(_point(child))
			"AGSWalkableSurface": lines.append_array(_walkable(child))
			"AGSBlockerVolume":   lines.append_array(_blocker(child))
			"AGSSpawnPoint": lines.append_array(_spawn(child))
			"AGSHotspot":    lines.append_array(_hotspot(child))
			"AGSTriggerRegion":   lines.append_array(_trigger(child))

	lines.append("}")
	return lines


static func _camera(node: Node) -> Array[String]:
	var pos := (node as Node3D).position
	# Reconstruct look_at from transform: camera looks along -Z in local space.
	var fwd: Vector3 = -(node as Node3D).global_transform.basis.z.normalized()
	var look_at: Vector3 = pos + fwd * 5.0
	return [
		"",
		'    Camera "%s" {' % node.get("camera_name"),
		"        position = %s" % _vec3(pos),
		"        look_at  = %s" % _vec3(look_at),
		"    }",
	]


static func _point(node: Node) -> Array[String]:
	var pos := (node as Node3D).position
	return [
		"",
		'    Point "%s" {' % node.get("point_name"),
		"        position = %s" % _vec3(pos),
		"    }",
	]


static func _walkable(node: Node) -> Array[String]:
	var shape := _box_shape(node)
	var size_x: float = shape.size.x if shape else 1.0
	var size_z: float = shape.size.z if shape else 1.0
	var offset := (node as Node3D).position
	return [
		"",
		'    WalkableSurface "%s" {' % _stem(node),
		"        size   = (%s, %s)" % [_f(size_x), _f(size_z)],
		"        offset = %s" % _vec3(offset),
		"    }",
	]


static func _blocker(node: Node) -> Array[String]:
	var shape := _box_shape(node)
	var size: Vector3 = shape.size if shape else Vector3.ONE
	var pos := (node as Node3D).position
	return [
		"",
		'    BlockerVolume "%s" {' % _stem(node),
		"        size     = %s" % _vec3(size),
		"        position = %s" % _vec3(pos),
		"    }",
	]


static func _spawn(node: Node) -> Array[String]:
	var pos := (node as Node3D).position
	return [
		"",
		'    SpawnPoint "%s" {' % _stem(node),
		'        character = "%s"' % node.get("spawn_character"),
		"        position  = %s" % _vec3(pos),
		"    }",
	]


static func _hotspot(node: Node) -> Array[String]:
	var shape := _box_shape(node)
	var size: Vector3 = shape.size if shape else Vector3.ONE
	var pos := (node as Node3D).position
	return [
		"",
		'    Hotspot "%s" {' % node.get("hotspot_name"),
		"        size     = %s" % _vec3(size),
		"        position = %s" % _vec3(pos),
		"    }",
	]


static func _trigger(node: Node) -> Array[String]:
	var shape := _box_shape(node)
	var size: Vector3 = shape.size if shape else Vector3.ONE
	var pos := (node as Node3D).position
	return [
		"",
		'    TriggerRegion "%s" {' % node.get("region_name"),
		"        size     = %s" % _vec3(size),
		"        position = %s" % _vec3(pos),
		"    }",
	]


# ---------------------------------------------------------------------------
# Formatting helpers
# ---------------------------------------------------------------------------

static func _vec3(v: Vector3) -> String:
	return "(%s, %s, %s)" % [_f(v.x), _f(v.y), _f(v.z)]


static func _f(v: float) -> String:
	# Keep up to 2 decimal places; strip trailing zeros.
	var s := "%.2f" % v
	if s.contains("."):
		s = s.rstrip("0").rstrip(".")
		# Ensure at least one decimal digit for readability
		if not s.contains("."):
			s += ".0"
	return s


static func _stem(node: Node) -> String:
	# Derive a snake_case name from the node's scene-tree name.
	return node.name.to_snake_case()


static func _box_shape(node: Node) -> BoxShape3D:
	for child in node.get_children():
		if child is CollisionShape3D:
			var col := child as CollisionShape3D
			if col.shape is BoxShape3D:
				return col.shape as BoxShape3D
	return null
