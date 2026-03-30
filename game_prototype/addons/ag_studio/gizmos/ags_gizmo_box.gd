@tool
extends EditorNode3DGizmoPlugin

## Base class for AGS box-volume gizmo plugins.
##
## Subclasses set _color and _xz_only to customise appearance.
## Handles map to the 6 (or 4 XZ-only) face half-extents of the BoxShape3D
## found in the node's first CollisionShape3D child.

var _undo_redo: EditorUndoRedoManager

## Override in subclass: wire colour for the box outline.
var _color := Color(1, 1, 1, 1)
## Override in subclass: restrict handles/drawing to XZ plane.
var _xz_only := false

# Handle index → (axis_index, sign) where axis_index: 0=X 1=Y 2=Z
const HANDLE_AXIS_SIGN := [
	[0,  1],  # 0: +X
	[0, -1],  # 1: -X
	[1,  1],  # 2: +Y
	[1, -1],  # 3: -Y
	[2,  1],  # 4: +Z
	[2, -1],  # 5: -Z
]
const HANDLE_AXIS_SIGN_XZ := [
	[0,  1],  # 0: +X
	[0, -1],  # 1: -X
	[2,  1],  # 2: +Z
	[2, -1],  # 3: -Z
]


func setup(ur: EditorUndoRedoManager) -> void:
	_undo_redo = ur


func _get_handle_count(gizmo: EditorNode3DGizmo) -> int:
	return 4 if _xz_only else 6


func _get_handle_name(gizmo: EditorNode3DGizmo, handle_id: int, secondary: bool) -> String:
	var names := ["+ X", "- X", "+ Y", "- Y", "+ Z", "- Z"] if not _xz_only \
		else ["+ X", "- X", "+ Z", "- Z"]
	return names[handle_id]


func _get_handle_value(gizmo: EditorNode3DGizmo, handle_id: int, secondary: bool) -> Variant:
	var shape := _get_box_shape(gizmo.get_node_3d())
	return shape.size if shape else Vector3.ONE


func _set_handle(gizmo: EditorNode3DGizmo, handle_id: int, secondary: bool,
		camera: Camera3D, screen_pos: Vector2) -> void:
	var node := gizmo.get_node_3d()
	var shape := _get_box_shape(node)
	if not shape:
		return

	var map := HANDLE_AXIS_SIGN_XZ if _xz_only else HANDLE_AXIS_SIGN
	var axis_idx: int = map[handle_id][0]

	var xform := node.global_transform
	var axis_local := Vector3.ZERO
	axis_local[axis_idx] = 1.0
	var axis_world := xform.basis * axis_local

	# Project onto a plane through the node origin, perpendicular to a third axis
	# that is most "away" from the camera — keeps the drag intuitive.
	var cam_dir := camera.global_transform.basis.z
	var perp := axis_world.cross(cam_dir).normalized()
	var plane_normal := axis_world.cross(perp).normalized()
	if plane_normal.length_squared() < 0.001:
		return

	var plane := Plane(plane_normal, xform.origin.dot(plane_normal))
	var ray_from := camera.project_ray_origin(screen_pos)
	var ray_dir  := camera.project_ray_normal(screen_pos)
	var hit: Variant = plane.intersects_ray(ray_from, ray_dir)
	if hit == null:
		return

	var local_hit := xform.affine_inverse() * (hit as Vector3)
	var half_size := abs(local_hit[axis_idx])
	var new_size := shape.size
	new_size[axis_idx] = max(0.01, half_size * 2.0)
	shape.size = new_size
	_redraw(gizmo)


func _commit_handle(gizmo: EditorNode3DGizmo, handle_id: int, secondary: bool,
		restore: Variant, cancel: bool) -> void:
	var node := gizmo.get_node_3d()
	var shape := _get_box_shape(node)
	if not shape:
		return
	if cancel:
		shape.size = restore as Vector3
		_redraw(gizmo)
		return
	if not _undo_redo:
		return
	var label: String = "Resize " + node.get_class()
	_undo_redo.create_action(label)
	_undo_redo.add_do_property(shape, "size", shape.size)
	_undo_redo.add_undo_property(shape, "size", restore as Vector3)
	_undo_redo.commit_action()


# ---------------------------------------------------------------------------
# Shared draw helpers
# ---------------------------------------------------------------------------

func _draw_box(gizmo: EditorNode3DGizmo, size: Vector3) -> void:
	var mat := get_material("line", gizmo)
	var h := size * 0.5
	var v := [
		Vector3(-h.x, -h.y, -h.z), Vector3( h.x, -h.y, -h.z),
		Vector3( h.x, -h.y,  h.z), Vector3(-h.x, -h.y,  h.z),
		Vector3(-h.x,  h.y, -h.z), Vector3( h.x,  h.y, -h.z),
		Vector3( h.x,  h.y,  h.z), Vector3(-h.x,  h.y,  h.z),
	]
	var lines := PackedVector3Array()
	# Bottom ring
	for i in 4:
		lines.append(v[i]);         lines.append(v[(i + 1) % 4])
	# Top ring
	for i in 4:
		lines.append(v[i + 4]);     lines.append(v[(i + 1) % 4 + 4])
	# Verticals
	for i in 4:
		lines.append(v[i]);         lines.append(v[i + 4])
	gizmo.add_lines(lines, mat)

	# Collision triangles — makes the box face-clickable in the editor viewport.
	# Without this only the tiny transform-origin arrows are click targets.
	var box_mesh := BoxMesh.new()
	box_mesh.size = size
	gizmo.add_collision_triangles(box_mesh.generate_triangle_mesh())

	# Handles
	var hmat := get_material("handle", gizmo)
	var map := HANDLE_AXIS_SIGN_XZ if _xz_only else HANDLE_AXIS_SIGN
	var handles := PackedVector3Array()
	var ids: Array[int] = []
	for i in map.size():
		var ax: int = map[i][0]
		var sg: int = map[i][1]
		var p := Vector3.ZERO
		p[ax] = h[ax] * sg
		handles.append(p)
		ids.append(i)
	gizmo.add_handles(handles, hmat, ids)


# ---------------------------------------------------------------------------
# Node helpers
# ---------------------------------------------------------------------------

func _get_box_shape(node: Node3D) -> BoxShape3D:
	for child in node.get_children():
		if child is CollisionShape3D:
			var col := child as CollisionShape3D
			if col.shape is BoxShape3D:
				return col.shape as BoxShape3D
	return null
