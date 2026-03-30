@tool
extends EditorNode3DGizmoPlugin

## Gizmo plugin for AGSPoint.
## Draws a named diamond at the point position + a vertical line to the floor.

const DIAMOND_RADIUS := 0.15
const FLOOR_LINE_LEN := 2.0


func _init() -> void:
	create_material("diamond", Color(1.0, 0.85, 0.0, 1.0))
	create_material("line",    Color(1.0, 0.85, 0.0, 0.6))


func _get_gizmo_name() -> String:
	return "AGSPoint"


func _has_gizmo(node: Node3D) -> bool:
	return node.get_class() == "AGSPoint"


func _redraw(gizmo: EditorNode3DGizmo) -> void:
	gizmo.clear()

	var dmat := get_material("diamond", gizmo)
	var lmat := get_material("line",    gizmo)
	var r    := DIAMOND_RADIUS

	# Diamond: 6 vertices of an octahedron, drawn as 12 edges.
	var top    := Vector3( 0,  r,  0)
	var bottom := Vector3( 0, -r,  0)
	var px     := Vector3( r,  0,  0)
	var mx     := Vector3(-r,  0,  0)
	var pz     := Vector3( 0,  0,  r)
	var mz     := Vector3( 0,  0, -r)
	var mid    := [px, pz, mx, mz]
	var lines  := PackedVector3Array()
	for i in 4:
		var a: Vector3 = mid[i]
		var b: Vector3 = mid[(i + 1) % 4]
		lines.append(top);  lines.append(a)
		lines.append(bottom); lines.append(a)
		lines.append(a);    lines.append(b)
	gizmo.add_lines(lines, dmat)

	# Vertical line downward to approximate floor.
	var drop := PackedVector3Array([Vector3.ZERO, Vector3(0, -FLOOR_LINE_LEN, 0)])
	gizmo.add_lines(drop, lmat)
