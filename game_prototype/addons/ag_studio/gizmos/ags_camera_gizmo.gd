@tool
extends EditorNode3DGizmoPlugin

## Gizmo plugin for AGSCamera.
##
## Draws a look-at arrow (orange) and adds billboard camera warning
## overlays (yellow spheres) when the camera configuration may cause
## visual artefacts for billboard-mode characters:
##   W1 — elevation > 30°  (may clip billboard sprites)
##   W2 — single-angle sprite without sprite_locked
##   W3 — horizontal arc > 45° with 4-angle sprites

const ARROW_LEN := 1.2
const ARROW_TIP := 0.2
const WARN_RADIUS := 0.15
const ELEV_THRESHOLD := 30.0
const ARC_THRESHOLD := 45.0


func _init() -> void:
	create_material("arrow", Color(1.0, 0.6, 0.0, 1.0))
	create_material("warn",  Color(1.0, 0.85, 0.0, 1.0))


func _get_gizmo_name() -> String:
	return "AGSCamera"


func _has_gizmo(node: Node3D) -> bool:
	return node.get_class() == "AGSCamera"


func _redraw(gizmo: EditorNode3DGizmo) -> void:
	gizmo.clear()

	var cam: Node3D = get_spatial_node(gizmo)
	if cam == null:
		return

	var arrow_mat := get_material("arrow", gizmo)
	var lines := PackedVector3Array()

	# Shaft along -Z (camera look direction in local space)
	var tip := Vector3(0, 0, -ARROW_LEN)
	lines.append(Vector3.ZERO); lines.append(tip)

	# Arrow head: two short diagonal lines in local XZ and YZ
	var t := ARROW_TIP
	lines.append(tip); lines.append(tip + Vector3( t, 0, t))
	lines.append(tip); lines.append(tip + Vector3(-t, 0, t))
	lines.append(tip); lines.append(tip + Vector3(0,  t, t))
	lines.append(tip); lines.append(tip + Vector3(0, -t, t))

	gizmo.add_lines(lines, arrow_mat)

	# Small sphere at origin so the camera node is click-selectable
	var sphere := SphereMesh.new()
	sphere.radius = 0.15
	sphere.height = 0.3
	gizmo.add_collision_triangles(sphere.generate_triangle_mesh())

	# Compute and draw billboard camera warnings
	var warnings := _compute_warnings(cam)
	if warnings.is_empty():
		return

	var warn_mat := get_material("warn", gizmo)
	for i in range(warnings.size()):
		var pos := _warning_sphere_offset(i, warnings.size())
		_draw_warning_sphere(gizmo, pos, warn_mat)


func _compute_warnings(cam: Node3D) -> Array[String]:
	var warnings: Array[String] = []
	if cam == null:
		return warnings

	# Find AGSRoom root by walking up from the camera
	var room := _find_ags_room(cam)

	# Collect billboard character sprite-angle info from the room
	var has_1angle := false
	var has_4angle := false
	var has_billboard := false
	if room != null:
		var chars := _collect_billboard_chars(room)
		has_1angle = chars["1angle"]
		has_4angle = chars["4angle"]
		has_billboard = chars["any"]

	if not has_billboard:
		return warnings  # No billboard chars → no warnings apply

	# Camera world position and look-at (project camera -Z forward)
	var cam_pos := cam.global_position
	var look_dir := -cam.global_transform.basis.z
	var look_at := cam_pos + look_dir * 3.0

	var dx := look_at.x - cam_pos.x
	var dy := look_at.y - cam_pos.y
	var dz := look_at.z - cam_pos.z
	var horizontal := sqrt(dx * dx + dz * dz)

	# W1 — elevation angle > ELEV_THRESHOLD
	if horizontal > 0.001:
		var elev_deg := rad_to_deg(atan2(abs(dy), horizontal))
		if elev_deg > ELEV_THRESHOLD:
			warnings.append("W1")

	# W3 — horizontal arc > ARC_THRESHOLD with 4-angle sprites
	if has_4angle and (abs(dx) > 0.001 or abs(dz) > 0.001):
		var arc_deg := rad_to_deg(atan2(abs(dx), abs(dz)))
		if arc_deg > ARC_THRESHOLD:
			warnings.append("W3")

	# W2 — single-angle sprites (sprite_locked is always false in generated scenes)
	if has_1angle:
		warnings.append("W2")

	return warnings


func _find_ags_room(from: Node3D) -> Node:
	var current: Node = from
	while current != null:
		if current.has_method("get_room_name"):
			return current
		current = current.get_parent()
	return null


func _collect_billboard_chars(room: Node) -> Dictionary:
	var result := {"any": false, "1angle": false, "4angle": false}
	if room == null:
		return result

	var stack: Array[Node] = [room]
	while not stack.is_empty():
		var node: Node = stack.pop_back()
		if node == null:
			continue
		# Detect AGSCharacter2D nodes that are billboard mode
		if node.get("character_name") != null and node.get("visual_mode") == "billboard":
			result["any"] = true
			var angles: int = node.get("sprite_angles", 1)
			if angles == 1:
				result["1angle"] = true
			elif angles == 4:
				result["4angle"] = true
		for child in node.get_children():
			stack.append(child)
	return result


func _warning_sphere_offset(idx: int, total: int) -> Vector3:
	# Arrange warning spheres in a small arc above the arrow tip
	var spread := PI / 4.0
	var start := -spread / 2.0
	var offset_angle := spread * float(idx) / max(1.0, float(total) - 1.0) if total > 1 else 0.0
	var angle := start + offset_angle
	var r := 0.4
	return Vector3(cos(angle) * r, 0.7 + sin(angle) * r * 0.3, -ARROW_LEN - 0.15)


func _draw_warning_sphere(gizmo: EditorNode3DGizmo, offset: Vector3, mat: Material) -> void:
	var sphere := SphereMesh.new()
	sphere.radius = WARN_RADIUS
	sphere.height = WARN_RADIUS * 2.0
	var mi := MeshInstance3D.new()
	mi.mesh = sphere
	mi.material_override = mat
	gizmo.add_mesh(mi)
