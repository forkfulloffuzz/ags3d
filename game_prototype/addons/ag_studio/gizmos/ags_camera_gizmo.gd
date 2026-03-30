@tool
extends EditorNode3DGizmoPlugin

## Gizmo plugin for AGSCamera.
## Adds a look-at arrow from the camera toward its -Z direction.
## The built-in Camera3D gizmo already draws the frustum.

const ARROW_LEN := 1.2
const ARROW_TIP := 0.2


func _init() -> void:
	create_material("arrow", Color(1.0, 0.6, 0.0, 1.0))


func _get_gizmo_name() -> String:
	return "AGSCamera"


func _has_gizmo(node: Node3D) -> bool:
	return node.get_class() == "AGSCamera"


func _redraw(gizmo: EditorNode3DGizmo) -> void:
	gizmo.clear()
	var mat   := get_material("arrow", gizmo)
	var lines := PackedVector3Array()

	# Shaft along -Z (camera look direction in local space)
	var tip := Vector3(0, 0, -ARROW_LEN)
	lines.append(Vector3.ZERO);         lines.append(tip)

	# Arrow head: two short diagonal lines in local XZ and YZ
	var t := ARROW_TIP
	lines.append(tip); lines.append(tip + Vector3( t, 0, t))
	lines.append(tip); lines.append(tip + Vector3(-t, 0, t))
	lines.append(tip); lines.append(tip + Vector3(0,  t, t))
	lines.append(tip); lines.append(tip + Vector3(0, -t, t))

	gizmo.add_lines(lines, mat)

	# Small sphere at origin so the camera node is click-selectable.
	var sphere := SphereMesh.new()
	sphere.radius = 0.15
	sphere.height = 0.3
	gizmo.add_collision_triangles(sphere.generate_triangle_mesh())
