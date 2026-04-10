@tool
extends EditorNode3DGizmoPlugin

## Gizmo for AGSItem / AGSRoomItem nodes.
## Draws a small coloured box at the item's position.

var _undo_redo: EditorUndoRedoManager

func setup(ur: EditorUndoRedoManager) -> void:
	_undo_redo = ur


func _has_gizmo(for_node_3d: Node3D) -> bool:
	return for_node_3d.get_class() in ["AGSItem", "AGSRoomItem"]


func _get_gizmo_name() -> String:
	return "AGSItem"


func _redraw(gizmo: EditorNode3DGizmo) -> void:
	gizmo.clear()
	var node: Node3D = gizmo.get_node_3d()
	if not is_instance_valid(node):
		return

	var mat := create_material("item", Color(1.0, 0.6, 0.2, 0.8), true, false)
	var size := Vector3(0.3, 0.3, 0.3)

	var box_mesh := BoxMesh.new()
	box_mesh.size = size
	gizmo.add_mesh(box_mesh, mat)

	var lines := PackedVector3Array()
	var h := size * 0.5
	var v := [
		Vector3(-h.x, -h.y, -h.z), Vector3( h.x, -h.y, -h.z),
		Vector3( h.x, -h.y,  h.z), Vector3(-h.x, -h.y,  h.z),
		Vector3(-h.x,  h.y, -h.z), Vector3( h.x,  h.y, -h.z),
		Vector3( h.x,  h.y,  h.z), Vector3(-h.x,  h.y,  h.z),
	]
	for i in 4:
		lines.append(v[i]);         lines.append(v[(i + 1) % 4])
	for i in 4:
		lines.append(v[i + 4]);     lines.append(v[(i + 1) % 4 + 4])
	for i in 4:
		lines.append(v[i]);         lines.append(v[i + 4])
	gizmo.add_lines(lines, mat)
