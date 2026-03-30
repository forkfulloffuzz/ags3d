@tool
extends "res://addons/ag_studio/gizmos/ags_gizmo_box.gd"

## Gizmo plugin for AGSWalkableSurface.
## Draws a green wireframe box; handles resize on XZ edges only.

func _init() -> void:
	_color   = Color(0.0, 0.85, 0.25, 1.0)
	_xz_only = true
	create_material("line",   _color)
	create_handle_material("handle")


func _get_gizmo_name() -> String:
	return "AGSWalkableSurface"


func _has_gizmo(node: Node3D) -> bool:
	return node.get_class() == "AGSWalkableSurface"


func _redraw(gizmo: EditorNode3DGizmo) -> void:
	gizmo.clear()
	var node := gizmo.get_node_3d()
	var shape := _get_box_shape(node)
	_draw_box(gizmo, shape.size if shape else Vector3(1.0, 0.1, 1.0))
