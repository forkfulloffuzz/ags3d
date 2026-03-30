@tool
extends EditorInspectorPlugin

## AGS Inspector plugin (T-E11)
##
## Intercepts all AGS node types and replaces the generic Godot Inspector with
## clean AGS-specific property forms. Godot's default property list is
## suppressed for these types.
##
## Registration: call add_inspector_plugin(plugin) from ag_studio.gd.

const AGS_CLASSES := [
	"AGSRoom",
	"AGSWalkableSurface",
	"AGSBlockerVolume",
	"AGSPoint",
	"AGSCamera",
	"AGSSpawnPoint",
	"AGSHotspot",
	"AGSTriggerRegion",
]

var _plugin: EditorPlugin


func setup(p: EditorPlugin) -> void:
	_plugin = p


func _can_handle(object: Object) -> bool:
	return object.get_class() in AGS_CLASSES


func _parse_begin(object: Object) -> void:
	var cls: String = object.get_class()
	var node := object as Node3D

	var container := VBoxContainer.new()
	container.add_theme_constant_override("separation", 6)

	_add_header(container, cls)

	match cls:
		"AGSRoom":       _build_room(container, node)
		"AGSWalkableSurface": _build_walkable(container, node)
		"AGSBlockerVolume":   _build_blocker(container, node)
		"AGSPoint":      _build_point(container, node)
		"AGSCamera":     _build_camera(container, node)
		"AGSSpawnPoint": _build_spawn(container, node)
		"AGSHotspot":    _build_hotspot(container, node)
		"AGSTriggerRegion":   _build_trigger(container, node)

	add_custom_control(container)


# ---------------------------------------------------------------------------
# Per-type forms
# ---------------------------------------------------------------------------

func _build_room(c: VBoxContainer, node: Node3D) -> void:
	_add_text_field(c, node, "Room name", "room_name")

	# initial_camera dropdown — lists names of AGSCamera children
	var cam_names: Array[String] = []
	for child in node.get_children():
		if child.get_class() == "AGSCamera":
			cam_names.append(child.get("camera_name") if child.get("camera_name") else child.name)

	var lbl := _label("Initial camera")
	c.add_child(lbl)
	var opt := OptionButton.new()
	opt.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var current: String = node.get("initial_camera")
	var sel_idx := 0
	for i in cam_names.size():
		opt.add_item(cam_names[i])
		if cam_names[i] == current:
			sel_idx = i
	if cam_names.is_empty():
		opt.add_item("(no cameras in room)")
		opt.disabled = true
	else:
		opt.select(sel_idx)
		opt.item_selected.connect(func(idx: int) -> void:
			node.set("initial_camera", cam_names[idx])
			_mark_scene_changed()
		)
	c.add_child(opt)


func _build_walkable(c: VBoxContainer, node: Node3D) -> void:
	var shape := _get_box_shape(node)
	if shape:
		_add_vec2_field(c, "Size (X / Z)", shape, "size",
			func(v: Vector2) -> void:
				shape.size = Vector3(v.x, shape.size.y, v.y)
				_mark_scene_changed()
		)
		_add_float_field(c, "Offset Y", node, "position",
			func() -> float: return node.position.y,
			func(v: float) -> void:
				node.position.y = v
				_mark_scene_changed()
		)
	else:
		_add_no_shape_hint(c)

	_add_position_field(c, node)


func _build_blocker(c: VBoxContainer, node: Node3D) -> void:
	var shape := _get_box_shape(node)
	if shape:
		_add_vec3_field(c, "Size", shape, "size",
			func(v: Vector3) -> void:
				shape.size = v
				_mark_scene_changed()
		)
	else:
		_add_no_shape_hint(c)

	_add_position_field(c, node)


func _build_point(c: VBoxContainer, node: Node3D) -> void:
	_add_text_field(c, node, "Point name", "point_name")
	_add_position_field(c, node)


func _build_camera(c: VBoxContainer, node: Node3D) -> void:
	_add_text_field(c, node, "Camera name", "camera_name")
	_add_position_field(c, node)
	# look_at is encoded in the transform; display as read-only for now.
	var lbl := _label("Look-at: edit via gizmo in 3D viewport")
	lbl.add_theme_color_override("font_color", Color(0.6, 0.6, 0.6))
	c.add_child(lbl)


func _build_spawn(c: VBoxContainer, node: Node3D) -> void:
	# Dropdown of known .agchar names from the project filesystem
	var char_names: Array[String] = _find_char_names()
	var lbl := _label("Character")
	c.add_child(lbl)
	var opt := OptionButton.new()
	opt.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var current: String = node.get("spawn_character")
	var sel_idx := 0
	for i in char_names.size():
		opt.add_item(char_names[i])
		if char_names[i] == current:
			sel_idx = i
	if char_names.is_empty():
		opt.add_item("(no .agchar files found)")
		opt.disabled = true
	else:
		opt.select(sel_idx)
		opt.item_selected.connect(func(idx: int) -> void:
			node.set("spawn_character", char_names[idx])
			_mark_scene_changed()
		)
	c.add_child(opt)
	_add_position_field(c, node)


func _build_hotspot(c: VBoxContainer, node: Node3D) -> void:
	_add_text_field(c, node, "Hotspot name", "hotspot_name")
	var shape := _get_box_shape(node)
	if shape:
		_add_vec3_field(c, "Size", shape, "size",
			func(v: Vector3) -> void:
				shape.size = v
				_mark_scene_changed()
		)
	else:
		_add_no_shape_hint(c)
	_add_position_field(c, node)


func _build_trigger(c: VBoxContainer, node: Node3D) -> void:
	_add_text_field(c, node, "Region name", "region_name")
	var shape := _get_box_shape(node)
	if shape:
		_add_vec3_field(c, "Size", shape, "size",
			func(v: Vector3) -> void:
				shape.size = v
				_mark_scene_changed()
		)
	else:
		_add_no_shape_hint(c)
	_add_position_field(c, node)


# ---------------------------------------------------------------------------
# Field helpers
# ---------------------------------------------------------------------------

func _add_header(c: VBoxContainer, cls: String) -> void:
	var lbl := Label.new()
	lbl.text = cls
	lbl.add_theme_font_size_override("font_size", 13)
	c.add_child(lbl)
	c.add_child(HSeparator.new())


func _label(text: String) -> Label:
	var lbl := Label.new()
	lbl.text = text
	lbl.add_theme_font_size_override("font_size", 11)
	return lbl


func _add_text_field(c: VBoxContainer, node: Node3D, label: String, prop: String) -> void:
	c.add_child(_label(label))
	var edit := LineEdit.new()
	edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	edit.text = str(node.get(prop))
	edit.text_submitted.connect(func(v: String) -> void:
		node.set(prop, v)
		_mark_scene_changed()
	)
	edit.focus_exited.connect(func() -> void:
		node.set(prop, edit.text)
		_mark_scene_changed()
	)
	c.add_child(edit)


func _add_position_field(c: VBoxContainer, node: Node3D) -> void:
	c.add_child(_label("Position"))
	var row := HBoxContainer.new()
	row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	for axis in ["x", "y", "z"]:
		var spin := SpinBox.new()
		spin.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		spin.step = 0.01
		spin.min_value = -9999.0
		spin.max_value =  9999.0
		spin.value = node.position[axis]
		spin.suffix = axis
		spin.value_changed.connect(func(v: float) -> void:
			node.position[axis] = v
			_mark_scene_changed()
		)
		row.add_child(spin)
	c.add_child(row)


func _add_vec3_field(c: VBoxContainer, label: String, obj: Object, _prop: String,
		on_change: Callable) -> void:
	c.add_child(_label(label))
	var row := HBoxContainer.new()
	row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var current: Vector3 = obj.get(_prop)
	for i in 3:
		var axis: String = ["x", "y", "z"][i]
		var spin := SpinBox.new()
		spin.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		spin.step = 0.01
		spin.min_value = 0.01
		spin.max_value = 9999.0
		spin.value = current[i]
		spin.suffix = axis
		spin.value_changed.connect(func(_v: float) -> void:
			var val: Vector3 = obj.get(_prop)
			val[i] = spin.value
			on_change.call(val)
		)
		row.add_child(spin)
	c.add_child(row)


func _add_vec2_field(c: VBoxContainer, label: String, obj: Object, _prop: String,
		on_change: Callable) -> void:
	c.add_child(_label(label))
	var row := HBoxContainer.new()
	row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var current: Vector3 = obj.get(_prop)
	for i in [0, 2]:  # X and Z only
		var axis: String = ["x", "z"][i / 2]
		var spin := SpinBox.new()
		spin.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		spin.step = 0.01
		spin.min_value = 0.01
		spin.max_value = 9999.0
		spin.value = current[i]
		spin.suffix = axis
		spin.value_changed.connect(func(_v: float) -> void:
			on_change.call(Vector2(
				row.get_child(0).value,
				row.get_child(1).value
			))
		)
		row.add_child(spin)
	c.add_child(row)


func _add_float_field(c: VBoxContainer, label: String, _obj: Object, _prop: String,
		getter: Callable, setter: Callable) -> void:
	c.add_child(_label(label))
	var spin := SpinBox.new()
	spin.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	spin.step = 0.01
	spin.min_value = -9999.0
	spin.max_value =  9999.0
	spin.value = getter.call()
	spin.value_changed.connect(func(v: float) -> void: setter.call(v))
	c.add_child(spin)


func _add_no_shape_hint(c: VBoxContainer) -> void:
	var lbl := _label("(no CollisionShape3D child found)")
	lbl.add_theme_color_override("font_color", Color(0.8, 0.4, 0.4))
	c.add_child(lbl)


# ---------------------------------------------------------------------------
# Utilities
# ---------------------------------------------------------------------------

func _get_box_shape(node: Node3D) -> BoxShape3D:
	for child in node.get_children():
		if child is CollisionShape3D:
			var col := child as CollisionShape3D
			if col.shape is BoxShape3D:
				return col.shape as BoxShape3D
	return null


func _find_char_names() -> Array[String]:
	var result: Array[String] = []
	var base: String = ProjectSettings.globalize_path("res://")
	_scan_for_chars(base, result)
	return result


func _scan_for_chars(dir: String, result: Array[String]) -> void:
	var da := DirAccess.open(dir)
	if not da:
		return
	da.list_dir_begin()
	var entry: String = da.get_next()
	while entry != "":
		if not entry.begins_with("."):
			var full: String = dir.path_join(entry)
			if da.current_is_dir() and entry != "addons":
				_scan_for_chars(full, result)
			elif entry.ends_with(".agchar"):
				result.append(entry.get_basename())
		entry = da.get_next()
	da.list_dir_end()


func _mark_scene_changed() -> void:
	if _plugin:
		_plugin.get_editor_interface().get_edited_scene_root().set_meta("_ags_modified", true)
