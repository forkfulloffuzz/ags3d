@tool
extends ConfirmationDialog

## AG Studio Project Settings dialog.
##
## Reads/writes game.agp (INI-style) in the project root.
## Currently exposes: start_room (dropdown of discovered .agroom files).

var _plugin: EditorPlugin
var _start_room_option: OptionButton
var _status_label: Label

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


func _on_confirmed() -> void:
	if _agp_path.is_empty():
		return

	var idx := _start_room_option.selected
	if idx < 0 or idx >= _room_options.size():
		return

	var chosen := _room_options[idx]
	if not _agp_data.has("project"):
		_agp_data["project"] = {}
	_agp_data["project"]["start_room"] = chosen

	var err := _write_agp(_agp_path, _agp_data)
	if err != OK:
		_status_label.text = "Error saving game.agp (code %d)." % err
		show()


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
