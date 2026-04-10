@tool
extends Control

## AG Studio Room Editor sidebar — AGS node tree + Add buttons.
##
## Works alongside the native Godot 3D viewport. When an AGS room scene is
## open in the editor, this panel shows the AGS node hierarchy and lets
## authors add new AGS nodes to the scene.
##
## Connect scene_changed to detect when a room scene becomes active.

signal node_added(node_path: String, node_type: String)

var _plugin: EditorPlugin
var _room_root: Node = null

var _node_tree: Tree
var _add_floor_btn: Button
var _add_blocker_btn: Button
var _add_point_btn: Button
var _add_camera_btn: Button
var _add_spawn_btn: Button
var _add_hotspot_btn: Button
var _add_region_btn: Button
var _add_item_btn: Button
var _import_blender_btn: Button


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "RoomEditor"
	_build_ui()
	get_editor_interface().resource_saved.connect(_on_resource_saved)


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var vbox := VBoxContainer.new()
	vbox.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	vbox.add_theme_constant_override("separation", 4)
	add_child(vbox)

	vbox.add_child(_build_toolbar())
	vbox.add_child(HSeparator.new())
	vbox.add_child(_build_node_tree())
	vbox.add_child(HSeparator.new())
	vbox.add_child(_build_import_row())


func _build_toolbar() -> HBoxContainer:
	var tb := HBoxContainer.new()
	tb.add_theme_constant_override("separation", 2)

	_add_floor_btn = _make_add_btn("+ Floor", Color(0.3, 0.8, 0.3))
	_add_floor_btn.pressed.connect(_on_add_floor)
	tb.add_child(_add_floor_btn)

	_add_blocker_btn = _make_add_btn("+ Blocker", Color(0.8, 0.3, 0.3))
	_add_blocker_btn.pressed.connect(_on_add_blocker)
	tb.add_child(_add_blocker_btn)

	_add_point_btn = _make_add_btn("+ Point", Color(1.0, 0.8, 0.2))
	_add_point_btn.pressed.connect(_on_add_point)
	tb.add_child(_add_point_btn)

	_add_camera_btn = _make_add_btn("+ Camera", Color(0.4, 0.6, 1.0))
	_add_camera_btn.pressed.connect(_on_add_camera)
	tb.add_child(_add_camera_btn)

	_add_spawn_btn = _make_add_btn("+ Spawn", Color(0.5, 0.3, 0.9))
	_add_spawn_btn.pressed.connect(_on_add_spawn)
	tb.add_child(_add_spawn_btn)

	_add_hotspot_btn = _make_add_btn("+ Hotspot", Color(0.3, 0.5, 0.9))
	_add_hotspot_btn.pressed.connect(_on_add_hotspot)
	tb.add_child(_add_hotspot_btn)

	_add_region_btn = _make_add_btn("+ Region", Color(0.7, 0.3, 0.7))
	_add_region_btn.pressed.connect(_on_add_region)
	tb.add_child(_add_region_btn)

	_add_item_btn = _make_add_btn("+ Item", Color(1.0, 0.6, 0.2))
	_add_item_btn.pressed.connect(_on_add_item)
	tb.add_child(_add_item_btn)

	return tb


func _build_node_tree() -> Control:
	var scroll := ScrollContainer.new()
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_node_tree = Tree.new()
	_node_tree.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_node_tree.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_node_tree.item_activated.connect(_on_node_item_activated)
	_node_tree.item_selected.connect(_on_node_selected)
	scroll.add_child(_node_tree)
	return scroll


func _build_import_row() -> HBoxContainer:
	var row := HBoxContainer.new()
	_import_blender_btn = Button.new()
	_import_blender_btn.text = "Import from Blender"
	_import_blender_btn.pressed.connect(_on_import_blender)
	row.add_child(_import_blender_btn)
	return row


func _make_add_btn(label: String, col: Color) -> Button:
	var btn := Button.new()
	btn.text = label
	btn.flat = true
	btn.tooltip_text = "Add %s to scene" % label.trim_prefix("+ ")
	return btn


func _set_enabled(enabled: bool) -> void:
	for btn: Button in [_add_floor_btn, _add_blocker_btn, _add_point_btn,
			_add_camera_btn, _add_spawn_btn, _add_hotspot_btn,
			_add_region_btn, _add_item_btn]:
		btn.disabled = not enabled


func _refresh_tree() -> void:
	_node_tree.clear()
	if not is_instance_valid(_room_root):
		return
	var root := _node_tree.create_item()
	root.text = _room_root.name
	root.set_icon(0, _get_class_icon(_room_root.get_class()))
	_populate_tree(root, _room_root)


func _populate_tree(parent: TreeItem, node: Node) -> void:
	for child: Node in node.get_children():
		if child.name.begins_with("AG"):
			var item := _node_tree.create_item(parent)
			item.text = child.name
			item.set_icon(0, _get_class_icon(child.get_class()))
			item.set_metadata(0, _room_root.get_path_to(child))
			_populate_tree(item, child)


func _get_class_icon(cls: String) -> Texture2D:
	var icon_name := "Object"
	match cls:
		"AGSWalkableSurface": icon_name = "GridMap"
		"AGSBlockerVolume": icon_name = "CubeMesh"
		"AGSPoint": icon_name = "Marker3D"
		"AGSCamera": icon_name = "Camera3D"
		"AGSSpawnPoint": icon_name = "CharacterBody3D"
		"AGSHotspot": icon_name = "Hotspot3D"
		"AGSTriggerRegion": icon_name = "VisibleOnScreenNotifier3D"
		"AGSItem", "AGSRoomItem": icon_name = "InventoryItem"
		"AGSRoom": icon_name = "WorldEnvironment"
	return get_theme_icon(icon_name, "EditorIcons")


func _on_add_floor() -> void:
	_add_node("AGSWalkableSurface", "Floor_%d" % _count_children("AGSWalkableSurface"))


func _on_add_blocker() -> void:
	_add_node("AGSBlockerVolume", "Blocker_%d" % _count_children("AGSBlockerVolume"))


func _on_add_point() -> void:
	_add_node("AGSPoint", "Point_%d" % _count_children("AGSPoint"))


func _on_add_camera() -> void:
	_add_node("AGSCamera", "Camera_%d" % _count_children("AGSCamera"))


func _on_add_spawn() -> void:
	_add_node("AGSSpawnPoint", "Spawn_%d" % _count_children("AGSSpawnPoint"))


func _on_add_hotspot() -> void:
	_add_node("AGSHotspot", "Hotspot_%d" % _count_children("AGSHotspot"))


func _on_add_region() -> void:
	_add_node("AGSTriggerRegion", "Region_%d" % _count_children("AGSTriggerRegion"))


func _on_add_item() -> void:
	_add_node("AGSRoomItem", "Item_%d" % _count_children("AGSRoomItem"))


func _count_children(type: String) -> int:
	if not is_instance_valid(_room_root):
		return 0
	var count := 0
	for ch: Node in _room_root.get_children():
		if ch.get_class() == type:
			count += 1
	return count


func _add_node(type: String, base_name: String) -> void:
	if not is_instance_valid(_room_root):
		push_warning("[AGS RoomEditor] No room scene open")
		return

	var unique_name := _make_unique_name(base_name)
	var ur := _plugin.get_undo_redo()

	ur.create_action("Add %s" % type)
	ur.add_do_method(self, "_do_add_node", type, unique_name)
	ur.add_undo_method(self, "_undo_remove_node", unique_name)
	ur.commit_action()


func _do_add_node(type: String, node_name: String) -> void:
	if not is_instance_valid(_room_root):
		return
	var node: Node = _create_ags_node(type)
	node.name = node_name
	_room_root.add_child(node)
	node.owner = _room_root
	_refresh_tree()
	node_added.emit(_room_root.get_path_to(node), type)


func _undo_remove_node(node_name: String) -> void:
	if not is_instance_valid(_room_root):
		return
	var node := _room_root.find_child(node_name, false, false)
	if node:
		_room_root.remove_child(node)
		node.queue_free()
	_refresh_tree()


func _make_unique_name(base: String) -> String:
	if not is_instance_valid(_room_root) or not _room_root.has_node(base):
		return base
	var i := 1
	while _room_root.has_node(base + "_%d" % i):
		i += 1
	return base + "_%d" % i


func _create_ags_node(type: String) -> Node:
	var node: Node
	match type:
		"AGSWalkableSurface":
			node = Node3D.new()
			node.set("class_name", type)
			var col := CollisionShape3D.new()
			var shape := BoxShape3D.new()
			shape.size = Vector3(4.0, 0.1, 4.0)
			col.shape = shape
			node.add_child(col)
			col.owner = node
		"AGSBlockerVolume":
			node = Node3D.new()
			node.set("class_name", type)
			var col := CollisionShape3D.new()
			var shape := BoxShape3D.new()
			shape.size = Vector3(1.0, 2.0, 1.0)
			col.shape = shape
			node.add_child(col)
			col.owner = node
		"AGSPoint":
			node = Node3D.new()
			node.set("class_name", type)
			node.set("point_name", "new_point")
		"AGSCamera":
			node = Camera3D.new()
			node.set("camera_name", "new_camera")
			node.fov = 70.0
		"AGSSpawnPoint":
			node = Node3D.new()
			node.set("class_name", type)
			node.set("character", "")
		"AGSHotspot":
			node = Node3D.new()
			node.set("class_name", type)
			var col := CollisionShape3D.new()
			var shape := BoxShape3D.new()
			shape.size = Vector3(0.5, 1.5, 0.5)
			col.shape = shape
			node.add_child(col)
			col.owner = node
		"AGSTriggerRegion":
			node = Node3D.new()
			node.set("class_name", type)
			node.set("region_name", "new_region")
			var col := CollisionShape3D.new()
			var shape := BoxShape3D.new()
			shape.size = Vector3(2.0, 2.0, 2.0)
			col.shape = shape
			node.add_child(col)
			col.owner = node
		"AGSRoomItem":
			node = Node3D.new()
			node.set("class_name", type)
			node.set("item_name", "new_item")
			var col := CollisionShape3D.new()
			var shape := BoxShape3D.new()
			shape.size = Vector3(0.3, 0.3, 0.3)
			col.shape = shape
			node.add_child(col)
			col.owner = node
		_:
			node = Node3D.new()
	return node


func _on_node_item_activated() -> void:
	var sel := _node_tree.get_selected()
	if not sel:
		return
	var path: String = sel.get_metadata(0)
	if path.is_empty():
		return
	var node: Node = _room_root.get_node_or_null(path)
	if node:
		get_editor_interface().get_selection().clear()
		get_editor_interface().get_selection().add_node(node)


func _on_node_selected() -> void:
	_on_node_item_activated()


func _on_import_blender() -> void:
	if not is_instance_valid(_room_root):
		push_warning("[AGS] No room open to re-import")
		return
	var room_name := _room_root.name
	var blend_path := ProjectSettings.globalize_path("res://blender/%s/%s.blend" % [room_name, room_name])
	if not DirAccess.file_exists_absolute(blend_path):
		push_warning("[AGS] No Blender file found at %s" % blend_path)
		return
	var out_path := ProjectSettings.globalize_path("res://blender/%s/%s.glb" % [room_name, room_name])
	var sidecar := ProjectSettings.globalize_path("res://blender/%s/.ags3d_last_export" % room_name)

	var needs_confirm := false
	if DirAccess.file_exists_absolute(sidecar):
		var agroom_mtime := FileAccess.get_modified_time("res://rooms/%s/%s.agroom" % [room_name, room_name])
		var export_mtime_str := FileAccess.get_file_as_string(sidecar).strip_edges()
		if not export_mtime_str.is_empty():
			var export_mtime := float(export_mtime_str)
			if agroom_mtime > export_mtime:
				needs_confirm = true

	if needs_confirm:
		var confirm := ConfirmationDialog.new()
		confirm.dialog_text = "The .agroom file has been edited since the last Blender export. Overwrite with Blender data?"
		confirm.ok_button_text = "Import"
		confirm.cancel_button_text = "Cancel"
		add_child(confirm)
		confirm.confirmed.connect(func() -> void:
			confirm.queue_free()
			_run_blender_export_and_build(room_name, blend_path, out_path, sidecar)
		)
		confirm.canceled.connect(func() -> void:
			confirm.queue_free()
		)
		confirm.popup_centered()
	else:
		_run_blender_export_and_build(room_name, blend_path, out_path, sidecar)


func _run_blender_export_and_build(room_name: String, blend_path: String, out_path: String, sidecar: String) -> void:
	var blender_script := ProjectSettings.globalize_path("res://../tools/blender_addon/operators.py")
	var args := [
		"--background",
		"--python",
		"%s" % blender_script,
		"--",
		"export_room",
		blend_path,
		out_path,
	]
	var output := []
	var err := OS.execute("blender", args, output, true)
	if err != 0:
		push_error("[AGS] Blender export failed: %s" % str(output))
		return
	var fa := FileAccess.open(sidecar, FileAccess.WRITE)
	if fa:
		fa.store_string(str(Time.get_unix_time_from_system()))
		fa.close()
	var build_log := _plugin.get_node_or_null("BuildLog")
	if build_log and build_log.has_method("run_build"):
		build_log.call("run_build")


func _on_resource_saved(res: Resource) -> void:
	_refresh_tree()


func set_room_scene(scene_root: Node) -> void:
	_room_root = scene_root
	_set_enabled(is_instance_valid(_room_root))
	_refresh_tree()


func get_room_scene() -> Node:
	return _room_root
