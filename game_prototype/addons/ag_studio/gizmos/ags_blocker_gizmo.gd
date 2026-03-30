@tool
extends "res://addons/ag_studio/gizmos/ags_gizmo_box.gd"

## Gizmo plugin for AGSBlockerVolume.
## Draws a red wireframe box with full 6-face resize handles.

func _init() -> void:
	_color   = Color(0.9, 0.15, 0.15, 1.0)
	_xz_only = false
	create_material("line",   _color)
	create_handle_material("handle")


func _get_gizmo_name() -> String:
	return "AGSBlockerVolume"


func _has_gizmo(node: Node3D) -> bool:
	return node.get_class() == "AGSBlockerVolume"


func _redraw(gizmo: EditorNode3DGizmo) -> void:
	gizmo.clear()
	var node := gizmo.get_node_3d()
	var shape := _get_box_shape(node)
	_draw_box(gizmo, shape.size if shape else Vector3.ONE)
