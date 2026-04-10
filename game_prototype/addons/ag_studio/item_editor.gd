@tool
extends Control

## Item Editor panel — lists and edits .agitem files.
## Shown as a dock panel alongside the native Godot inspector.

signal item_saved(path: String)

var _plugin: EditorPlugin
var _tree: Tree
var _form: VBoxContainer
var _name_edit: LineEdit
var _display_name_edit: LineEdit
var _desc_edit: TextEdit
var _sprite_edit: LineEdit
var _save_btn: Button
var _current_item: String = ""


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "ItemEditor"
	_build_ui()
	_populate_tree()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var hbox := HSplitContainer.new()
	hbox.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.size_flags_vertical = Control.SIZE_EXPAND_FILL
	hbox.split_offset = 180
	add_child(hbox)

	# Left: item tree
	var left := VBoxContainer.new()
	left.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.add_child(left)

	var header := Label.new()
	header.text = "Items"
	header.add_theme_font_size_override("font_size", 12)
	left.add_child(header)

	_tree = Tree.new()
	_tree.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_tree.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_tree.item_activated.connect(_on_item_activated)
	_tree.item_selected.connect(_on_item_selected)
	left.add_child(_tree)

	var btn_row := HBoxContainer.new()
	left.add_child(btn_row)
	var add_btn := Button.new()
	add_btn.text = "+ New"
	add_btn.pressed.connect(_on_new_item)
	btn_row.add_child(add_btn)
	var del_btn := Button.new()
	del_btn.text = "Delete"
	del_btn.pressed.connect(_on_delete_item)
	btn_row.add_child(del_btn)

	# Right: edit form
	_form = VBoxContainer.new()
	_form.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.add_child(_form)

	var form_header := Label.new()
	form_header.text = "Item Properties"
	form_header.add_theme_font_size_override("font_size", 12)
	_form.add_child(form_header)
	_form.add_child(HSeparator.new())

	_form.add_child(_field_label("Internal name"))
	_name_edit = LineEdit.new()
	_name_edit.editable = false
	_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_form.add_child(_name_edit)

	_form.add_child(_field_label("Display name"))
	_display_name_edit = LineEdit.new()
	_display_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_form.add_child(_display_name_edit)

	_form.add_child(_field_label("Description"))
	_desc_edit = TextEdit.new()
	_desc_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_desc_edit.custom_minimum_size.y = 80
	_form.add_child(_desc_edit)

	_form.add_child(_field_label("Sprite"))
	var sprite_row := HBoxContainer.new()
	sprite_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_form.add_child(sprite_row)
	_sprite_edit = LineEdit.new()
	_sprite_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_sprite_edit.placeholder_text = "inventory/items/sprite.png"
	sprite_row.add_child(_sprite_edit)
	var browse_btn := Button.new()
	browse_btn.text = "…"
	browse_btn.pressed.connect(_browse_sprite)
	sprite_row.add_child(browse_btn)

	_form.add_child(HSeparator.new())

	_save_btn = Button.new()
	_save_btn.text = "Save"
	_save_btn.disabled = true
	_save_btn.pressed.connect(_on_save)
	_form.add_child(_save_btn)

	_form.add_child(Strut.new(), true)


func _populate_tree() -> void:
	_tree.clear()
	var root := _tree.create_item()
	root.text = "Items"

	var items_dir := "res://inventory/"
	if not DirAccess.dir_exists_absolute(items_dir):
		return
	var da := DirAccess.open(items_dir)
	if not da:
		return
	da.list_dir_begin()
	var entry: String = da.get_next()
	while entry != "":
		if entry.ends_with(".agitem"):
			var child := _tree.create_item(root)
			child.text = entry.get_basename()
			child.set_metadata(0, items_dir.path_join(entry))
		entry = da.get_next()
	da.list_dir_end()


func _on_item_selected() -> void:
	var sel := _tree.get_selected()
	if not sel:
		return
	_load_item(sel.get_metadata(0))


func _on_item_activated() -> void:
	_on_save()


func _load_item(abs_path: String) -> void:
	_current_item = abs_path
	_name_edit.text = abs_path.get_file().get_basename()
	var src := FileAccess.get_file_as_string(abs_path)
	if src.is_empty():
		return
	var parsed := _parse_agitem(src)
	_display_name_edit.text = parsed.get("display_name", "")
	_desc_edit.text = parsed.get("description", "")
	_sprite_edit.text = parsed.get("sprite", "")
	_save_btn.disabled = false


func _parse_agitem(src: String) -> Dictionary:
	var result: Dictionary = {}
	var rx_kv := RegEx.new()
	rx_kv.compile('(\\w+)\\s*=\\s*"([^"]*)"')
	for m in rx_kv.search_all(src):
		result[m.get_string(1)] = m.get_string(2)
	return result


func _on_new_item() -> void:
	var dialog := ConfirmationDialog.new()
	dialog.dialog_text = "Item name:"
	dialog.ok_button_text = "Create"
	var input := LineEdit.new()
	input.name = "name_input"
	dialog.add_child(input)
	add_child(dialog)
	dialog.confirmed.connect(func() -> void:
		var name := (dialog.get_node("name_input") as LineEdit).text.strip_edges()
		if name.is_empty():
			return
		var path := "res://inventory/%s.agitem" % name
		var fa := FileAccess.open(path, FileAccess.WRITE)
		if fa:
			fa.store_string('Item "%s" {\\n    display_name = "%s"\\n    description = ""\\n    sprite = ""\\n}\\n' % [name, name])
			fa.close()
		_populate_tree()
		dialog.queue_free()
	)
	dialog.canceled.connect(func() -> void:
		dialog.queue_free()
	)
	dialog.popup_centered()


func _on_delete_item() -> void:
	var sel := _tree.get_selected()
	if not sel:
		return
	var path: String = sel.get_metadata(0)
	if path.is_empty() or not path.ends_with(".agitem"):
		return
	var confirm := ConfirmationDialog.new()
	confirm.dialog_text = "Delete %s?" % sel.text
	add_child(confirm)
	confirm.confirmed.connect(func() -> void:
		DirAccess.remove_absolute(path)
		_populate_tree()
		confirm.queue_free()
	)
	confirm.canceled.connect(func() -> void:
		confirm.queue_free()
	)
	confirm.popup_centered()


func _browse_sprite() -> void:
	var dialog := EditorFileDialog.new()
	dialog.file_mode = EditorFileDialog.FILE_MODE_OPEN_FILE
	dialog.filters = PackedStringArray(["*.png ; PNG Image", "*.svg ; SVG Image"])
	dialog.file_selected.connect(func(p: String) -> void:
		_sprite_edit.text = ProjectSettings.localize_path(p)
		dialog.queue_free()
	)
	add_child(dialog)
	dialog.popup_centered_ratio(0.6)


func _on_save() -> void:
	if _current_item.is_empty():
		return
	var name := _name_edit.text
	var dn := _display_name_edit.text.strip_edges()
	var desc := _desc_edit.text.strip_edges()
	var sprite := _sprite_edit.text.strip_edges()
	var lines: Array[String] = []
	lines.append('Item "%s" {' % name)
	if not dn.is_empty():
		lines.append('    display_name = "%s"' % dn)
	if not desc.is_empty():
		lines.append('    description = "%s"' % desc)
	if not sprite.is_empty():
		lines.append('    sprite = "%s"' % sprite)
	lines.append("}")
	lines.append("")
	var fa := FileAccess.open(_current_item, FileAccess.WRITE)
	if not fa:
		return
	fa.store_string("\n".join(lines))
	fa.close()
	item_saved.emit(_current_item)
	_populate_tree()


func _field_label(text: String) -> Label:
	var lbl := Label.new()
	lbl.text = text
	lbl.add_theme_font_size_override("font_size", 11)
	return lbl
