@tool
extends ConfirmationDialog

## AG Studio Project Settings dialog.
##
## Reads/writes game.agp (INI-style) in the project root.
## Exposes: start_room (dropdown of discovered .agroom files),
## and global variables editor.

var _plugin: EditorPlugin
var _start_room_option: OptionButton
var _status_label: Label

var _globals_scroll: ScrollContainer
var _globals_container: VBoxContainer
var _globals_rows: Array[Dictionary] = []

# Parsed .agp content: section → key → value (all strings, unquoted)
var _agp_data: Dictionary = {}
var _agp_path: String = ""
# Maps OptionButton index → relative room path (e.g. "rooms/start/start.agroom")
var _room_options: Array[String] = []


func setup(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	title = "AG Studio — Project Settings"
	min_size = Vector2i(480, 200)
	confirmed.connect(_on_confirmed)
	_build_ui()


func _build_ui() -> void:
	var vbox := VBoxContainer.new()
	vbox.add_theme_constant_override("separation", 12)
	add_child(vbox)

	var heading := Label.new()
	heading.text = "Game"
	heading.add_theme_font_size_override("font_size", 13)
	vbox.add_child(heading)
	vbox.add_child(HSeparator.new())

	var row := HBoxContainer.new()
	row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	vbox.add_child(row)

	var lbl := Label.new()
	lbl.text = "Start room"
	lbl.custom_minimum_size = Vector2(120, 0)
	row.add_child(lbl)

	_start_room_option = OptionButton.new()
	_start_room_option.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(_start_room_option)

	vbox.add_child(HSeparator.new())

	# ---- Globals section ----
	var globals_header := Button.new()
	globals_header.text = "Global Variables  (click to expand)"
	globals_header.alignment = Alignment.ALIGN_LEFT
	vbox.add_child(globals_header)

	_globals_scroll = ScrollContainer.new()
	_globals_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_globals_scroll.custom_minimum_size.y = 160
	_globals_scroll.visible = false

	_globals_container = VBoxContainer.new()
	_globals_container.add_theme_constant_override("separation", 4)
	_globals_scroll.add_child(_globals_container)
	vbox.add_child(_globals_scroll)

	globals_header.pressed.connect(func() -> void:
		_globals_scroll.visible = not _globals_scroll.visible
		globals_header.text = "Global Variables  " + ("(click to collapse)" if _globals_scroll.visible else "(click to expand)")
	)

	var globals_btn_row := HBoxContainer.new()
	vbox.add_child(globals_btn_row)
	var add_var_btn := Button.new()
	add_var_btn.text = "+ Add Variable"
	add_var_btn.pressed.connect(_add_global_row)
	globals_btn_row.add_child(add_var_btn)

	_status_label = Label.new()
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD
	_status_label.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7))
	vbox.add_child(_status_label)


# ---------------------------------------------------------------------------
# Public — call before popup_centered()
# ---------------------------------------------------------------------------

func load_settings() -> void:
	_agp_path = _find_agp()
	if _agp_path.is_empty():
		_status_label.text = "game.agp not found in project root."
		return

	_agp_data = _parse_agp(_agp_path)
	_populate_rooms()
	_populate_globals()


# ---------------------------------------------------------------------------
# Internal
# ---------------------------------------------------------------------------

func _find_agp() -> String:
	var root := ProjectSettings.globalize_path("res://")
	var path := root.path_join("game.agp")
	if FileAccess.file_exists(path):
		return path
	return ""


func _populate_rooms() -> void:
	_start_room_option.clear()
	_room_options.clear()

	var root := ProjectSettings.globalize_path("res://")
	var rooms := _find_agrooms(root)

	var current: String = _agp_data.get("project", {}).get("start_room", "")

	for abs_path: String in rooms:
		var rel: String = abs_path.replace(root, "").trim_prefix("/")
		_room_options.append(rel)
		_start_room_option.add_item(rel)
		if rel == current or abs_path == current:
			_start_room_option.selected = _start_room_option.item_count - 1

	if _start_room_option.item_count == 0:
		_status_label.text = "No .agroom files found."


func _find_agrooms(dir: String) -> Array[String]:
	var result: Array[String] = []
	_scan_dir(dir, result)
	return result


func _scan_dir(dir: String, result: Array[String]) -> void:
	var da := DirAccess.open(dir)
	if not da:
		return
	da.list_dir_begin()
	var entry := da.get_next()
	while entry != "":
		if not entry.begins_with("."):
			var full := dir.path_join(entry)
			if da.current_is_dir():
				if entry != "addons" and entry != ".engine":
					_scan_dir(full, result)
			elif entry.ends_with(".agroom"):
				result.append(full)
		entry = da.get_next()
	da.list_dir_end()


func _clear_global_rows() -> void:
	for row_data: Dictionary in _globals_rows:
		row_data["hbox"].queue_free()
	_globals_rows.clear()


func _add_global_row(key := "", value := "") -> void:
	var hbox := HBoxContainer.new()
	hbox.size_flags_horizontal = Control.SIZE_EXPAND_FILL

	var key_edit := LineEdit.new()
	key_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	key_edit.placeholder_text = "variable_name"
	key_edit.text = key
	hbox.add_child(key_edit)

	var val_edit := LineEdit.new()
	val_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	val_edit.placeholder_text = "value"
	val_edit.text = value
	hbox.add_child(val_edit)

	var del_btn := Button.new()
	del_btn.text = "X"
	del_btn.custom_minimum_size.x = 28
	del_btn.pressed.connect(func() -> void:
		var idx := _globals_rows.find({"hbox": hbox, "key_edit": key_edit, "val_edit": val_edit})
		if idx >= 0:
			_globals_rows.remove_at(idx)
		hbox.queue_free()
	)
	hbox.add_child(del_btn)

	_globals_container.add_child(hbox)
	_globals_rows.append({"hbox": hbox, "key_edit": key_edit, "val_edit": val_edit})


func _populate_globals() -> void:
	_clear_global_rows()
	var globals: Dictionary = _agp_data.get("globals", {})
	for key: String in globals:
		_add_global_row(key, globals[key])


func _on_confirmed() -> void:
	if _agp_path.is_empty():
		return

	var idx := _start_room_option.selected
	if idx < 0 or idx >= _room_options.size():
		return

	var chosen := _room_options[idx]  # e.g. "rooms/park/park.agroom"
	if not _agp_data.has("project"):
		_agp_data["project"] = {}
	_agp_data["project"]["start_room"] = chosen

	# Collect globals from rows
	var globals_out: Dictionary = {}
	for row_data: Dictionary in _globals_rows:
		var k: String = row_data["key_edit"].text.strip_edges()
		var v: String = row_data["val_edit"].text.strip_edges()
		if k != "":
			globals_out[k] = v
	_agp_data["globals"] = globals_out

	var err := _write_agp(_agp_path, _agp_data)
	if err != OK:
		_status_label.text = "Error saving game.agp (code %d)." % err
		show()
		return

	# Derive the .tscn path and update project.godot run/main_scene.
	var tscn_rel: String = chosen.get_basename() + ".tscn"  # rooms/park/park.tscn
	var tscn_res: String = "res://" + tscn_rel
	_update_godot_main_scene(tscn_res)

	# Notify the editor so F5 picks up the change immediately.
	ProjectSettings.set_setting("application/run/main_scene", tscn_res)
	ProjectSettings.save()


func _update_godot_main_scene(tscn_res: String) -> void:
	var root := ProjectSettings.globalize_path("res://")
	var path := root.path_join("project.godot")
	var fa := FileAccess.open(path, FileAccess.READ)
	if not fa:
		push_warning("[AGS] Settings: cannot read project.godot")
		return
	var lines := fa.get_as_text().split("\n")
	fa.close()

	var out := PackedStringArray()
	for line: String in lines:
		if line.begins_with("run/main_scene="):
			out.append('run/main_scene="%s"' % tscn_res)
		else:
			out.append(line)

	var fw := FileAccess.open(path, FileAccess.WRITE)
	if not fw:
		push_warning("[AGS] Settings: cannot write project.godot")
		return
	fw.store_string("\n".join(out))
	fw.close()


# ---------------------------------------------------------------------------
# INI parser / writer (minimal — handles [section] and key = value lines)
# ---------------------------------------------------------------------------

func _parse_agp(path: String) -> Dictionary:
	var result: Dictionary = {}
	var fa := FileAccess.open(path, FileAccess.READ)
	if not fa:
		return result
	var section := ""
	while not fa.eof_reached():
		var line := fa.get_line().strip_edges()
		if line.begins_with("#") or line.begins_with(";") or line.is_empty():
			continue
		if line.begins_with("[") and line.ends_with("]"):
			section = line.substr(1, line.length() - 2).strip_edges()
			if not result.has(section):
				result[section] = {}
		elif "=" in line and section != "":
			var parts := line.split("=", false, 1)
			if parts.size() == 2:
				var key := parts[0].strip_edges()
				var val := parts[1].strip_edges().trim_prefix('"').trim_suffix('"')
				result[section][key] = val
	fa.close()
	return result


func _write_agp(path: String, data: Dictionary) -> Error:
	var fa := FileAccess.open(path, FileAccess.WRITE)
	if not fa:
		return ERR_FILE_CANT_WRITE
	for section: String in data:
		fa.store_line("[%s]" % section)
		for key: String in data[section]:
			fa.store_line('%s = "%s"' % [key, data[section][key]])
		fa.store_line("")
	fa.close()
	return OK
