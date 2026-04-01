@tool
extends ConfirmationDialog

## AG Studio Project Wizard (T-E17)
##
## Accessible from the AG Studio menu. Collects:
##   - Project folder  (native folder picker)
##   - Project name
##   - Initial room name
## Then writes the scaffold and calls ag build.

signal project_created(project_dir: String)

var _plugin: EditorPlugin

var _folder_edit: LineEdit
var _project_name_edit: LineEdit
var _room_name_edit: LineEdit
var _status_label: Label


func setup(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	title = "New AG Studio Project"
	min_size = Vector2i(520, 280)
	confirmed.connect(_on_confirmed)
	_build_ui()


func _build_ui() -> void:
	var vbox := VBoxContainer.new()
	vbox.add_theme_constant_override("separation", 10)
	add_child(vbox)

	# Project folder
	vbox.add_child(_label("Project folder"))
	var folder_row := HBoxContainer.new()
	folder_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	vbox.add_child(folder_row)

	_folder_edit = LineEdit.new()
	_folder_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_folder_edit.placeholder_text = "/home/user/Projects/mygame"
	folder_row.add_child(_folder_edit)

	var browse_btn := Button.new()
	browse_btn.text = "…"
	browse_btn.pressed.connect(_browse_folder)
	folder_row.add_child(browse_btn)

	# Project name
	vbox.add_child(_label("Project name"))
	_project_name_edit = LineEdit.new()
	_project_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_project_name_edit.placeholder_text = "My Adventure"
	vbox.add_child(_project_name_edit)

	# Initial room name
	vbox.add_child(_label("First room name"))
	_room_name_edit = LineEdit.new()
	_room_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_room_name_edit.placeholder_text = "start"
	_room_name_edit.text = "start"
	vbox.add_child(_room_name_edit)

	# Status
	_status_label = Label.new()
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD
	_status_label.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7))
	vbox.add_child(_status_label)


# ---------------------------------------------------------------------------
# Actions
# ---------------------------------------------------------------------------

func _browse_folder() -> void:
	var dialog := FileDialog.new()
	dialog.file_mode = FileDialog.FILE_MODE_OPEN_DIR
	dialog.access = FileDialog.ACCESS_FILESYSTEM
	dialog.dir_selected.connect(func(path: String) -> void:
		_folder_edit.text = path
		dialog.queue_free()
	)
	add_child(dialog)
	dialog.popup_centered_ratio(0.6)


func _on_confirmed() -> void:
	var folder: String = _folder_edit.text.strip_edges()
	var proj_name: String = _project_name_edit.text.strip_edges()
	var room_name: String = _room_name_edit.text.strip_edges().to_lower().replace(" ", "_")

	if folder.is_empty() or proj_name.is_empty() or room_name.is_empty():
		_status_label.text = "All fields are required."
		show()
		return

	var err := _write_scaffold(folder, proj_name, room_name)
	if err != OK:
		_status_label.text = "Error writing scaffold (code %d)." % err
		show()
		return

	project_created.emit(folder)


func _write_scaffold(folder: String, proj_name: String, room_name: String) -> Error:
	# Ensure directory exists
	DirAccess.make_dir_recursive_absolute(folder)

	# project.godot — minimal Godot project file
	var godot_proj := (
		"config_version=5\n\n" +
		"[application]\n\n" +
		"config/name=%s\n" +
		"config/features=PackedStringArray(\"4.4\", \"Forward Plus\")\n" +
		"config/icon=\"res://icon.svg\"\n"
	) % proj_name.json_escape().replace('"', '"')
	# Use a simple ini-style name (no quotes needed for plain names)
	godot_proj = (
		"config_version=5\n\n" +
		"[application]\n\n" +
		"config/name=\"%s\"\n" +
		"config/features=PackedStringArray(\"4.4\", \"Forward Plus\")\n"
	) % proj_name.replace('"', "'")
	var err := _write(folder.path_join("project.godot"), godot_proj)
	if err != OK: return err

	# game.agp
	var agp := 'Project "%s" {\n    start_room = "%s"\n}\n' % [proj_name, room_name]
	err = _write(folder.path_join("game.agp"), agp)
	if err != OK: return err

	# rooms/<room>/<room>.agroom
	var room_dir := folder.path_join("rooms").path_join(room_name)
	DirAccess.make_dir_recursive_absolute(room_dir)
	var agroom := (
		'Room "%s" {\n' +
		'    initial_camera = "main"\n\n' +
		'    Camera "main" {\n' +
		'    }\n\n' +
		'    WalkableSurface "floor" {\n' +
		'        size   = (10.0, 10.0)\n' +
		'        offset = (0.0, -0.05, 0.0)\n' +
		'    }\n\n' +
		'    SpawnPoint "player_start" {\n' +
		'        character = "player"\n' +
		'        position  = (0.0, 0.0, 0.0)\n' +
		'    }\n' +
		'}\n'
	) % room_name
	err = _write(room_dir.path_join(room_name + ".agroom"), agroom)
	if err != OK: return err

	# rooms/<room>/<room>.agscript
	var agscript := '// %s.agscript — room script for \'%s\'\n\nfunction room_Enter() {\n}\n' % [room_name, room_name]
	err = _write(room_dir.path_join(room_name + ".agscript"), agscript)
	if err != OK: return err

	# characters/player.agchar
	var char_dir := folder.path_join("characters")
	DirAccess.make_dir_recursive_absolute(char_dir)
	var agchar := 'Character "player" {\n    display_name = "Player"\n}\n'
	err = _write(char_dir.path_join("player.agchar"), agchar)
	if err != OK: return err

	# Run ag build inside the new project directory via a shell so the cwd is correct.
	_status_label.text = "Running ag build…"
	var ag_bin := _find_ag_binary()
	if ag_bin.is_empty():
		push_warning("[AGS] project_wizard: ag binary not found, skipping build")
	else:
		var output: Array = []
		var exit_code := OS.execute("bash", ["-c", "cd '%s' && '%s' build --force" % [folder, ag_bin]], output, true)
		if exit_code != 0:
			push_warning("[AGS] project_wizard: ag build exited %d\n%s" % [exit_code, "\n".join(output)])

	# Open the new project in a fresh Godot editor instance.
	OS.create_process(OS.get_executable_path(), ["--editor", "--path", folder])

	return OK


func _write(path: String, content: String) -> Error:
	var fa := FileAccess.open(path, FileAccess.WRITE)
	if not fa:
		return ERR_FILE_CANT_WRITE
	fa.store_string(content)
	fa.close()
	return OK


func _find_ag_binary() -> String:
	var res_path: String = ProjectSettings.globalize_path("res://")
	var repo_root: String = res_path.get_base_dir()
	for path: String in [repo_root.path_join("bin/ag"), repo_root.path_join("tools/ag/ag")]:
		if FileAccess.file_exists(path):
			return path
	return ""


func _label(text: String) -> Label:
	var lbl := Label.new()
	lbl.text = text
	lbl.add_theme_font_size_override("font_size", 11)
	return lbl
