@tool
extends VBoxContainer

## AG Studio Project Panel dock
##
## Displays a tree of all AG source files grouped by type:
##   rooms/     — .agroom + companion .agscript pairs
##   characters/ — .agchar files
##   scripts/    — .agscript files not paired with a room
##
## Emits file_activated(path: String) when a tree item is double-clicked.

signal file_activated(path: String)

# Icons fetched from the editor theme.
var _icon_room:   Texture2D
var _icon_char:   Texture2D
var _icon_script: Texture2D
var _icon_folder: Texture2D

var _tree: Tree
var _refresh_btn: Button
var _plugin: EditorPlugin  # set by ag_studio.gd after instantiation

# File-watcher so we pick up external changes.
var _dir_watcher: EditorFileSystemDirectory  # unused for now; refresh is manual + on fs_changed

func _ready() -> void:
	name = "Project"
	_build_ui()
	_fetch_icons()
	# Connect to the editor filesystem so we refresh when files change.
	var efs := EditorInterface.get_singleton().get_resource_filesystem()
	efs.filesystem_changed.connect(_on_filesystem_changed)
	call_deferred("refresh")


func _build_ui() -> void:
	# Toolbar row.
	var toolbar := HBoxContainer.new()
	toolbar.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	add_child(toolbar)

	var lbl := Label.new()
	lbl.text = "Project"
	lbl.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	lbl.add_theme_font_size_override("font_size", 11)
	toolbar.add_child(lbl)

	_refresh_btn = Button.new()
	_refresh_btn.text = "↺"
	_refresh_btn.tooltip_text = "Refresh file tree"
	_refresh_btn.flat = true
	_refresh_btn.pressed.connect(refresh)
	toolbar.add_child(_refresh_btn)

	# Separator.
	var sep := HSeparator.new()
	add_child(sep)

	# Tree.
	_tree = Tree.new()
	_tree.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_tree.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_tree.hide_root = true
	_tree.select_mode = Tree.SELECT_ROW
	_tree.item_activated.connect(_on_item_activated)
	_tree.item_mouse_selected.connect(_on_item_right_clicked)
	add_child(_tree)


func _fetch_icons() -> void:
	var base := EditorInterface.get_singleton().get_base_control()
	_icon_room   = base.get_theme_icon("Node3D",   "EditorIcons")
	_icon_char   = base.get_theme_icon("CharacterBody3D", "EditorIcons")
	_icon_script = base.get_theme_icon("Script",   "EditorIcons")
	_icon_folder = base.get_theme_icon("Folder",   "EditorIcons")


## Rebuild the tree from the project filesystem.
func refresh() -> void:
	if not _tree:
		return
	_tree.clear()
	var root := _tree.create_item()

	var res_base := "res://"
	var abs_base := ProjectSettings.globalize_path(res_base)

	var rooms      := _find_files(abs_base, ".agroom")
	var chars      := _find_files(abs_base, ".agchar")
	var all_scripts := _find_files(abs_base, ".agscript")

	# Scripts that share a stem with a .agroom are "room scripts" — exclude
	# them from the standalone scripts section.
	var room_script_stems: Dictionary = {}
	for p in rooms:
		room_script_stems[p.get_basename()] = true

	var standalone_scripts: Array[String] = []
	for p in all_scripts:
		if not room_script_stems.has(p.get_basename()):
			standalone_scripts.append(p)

	_populate_section(root, "Rooms",      rooms,              _icon_room,   abs_base)
	_populate_section(root, "Characters", chars,              _icon_char,   abs_base)
	_populate_section(root, "Scripts",    standalone_scripts, _icon_script, abs_base)


# ---------------------------------------------------------------------------
# Tree population
# ---------------------------------------------------------------------------

func _populate_section(root: TreeItem, label: String, files: Array, icon: Texture2D, base: String) -> void:
	if files.is_empty():
		return
	var section := _tree.create_item(root)
	section.set_text(0, label)
	section.set_icon(0, _icon_folder)
	section.set_selectable(0, false)
	section.collapsed = false

	# Group rooms by their containing directory (room name = folder).
	if label == "Rooms":
		_populate_rooms(section, files, base)
	else:
		for path in files:
			_add_leaf(section, path, icon, base)


func _populate_rooms(section: TreeItem, room_files: Array, base: String) -> void:
	# Each .agroom file may have a companion .agscript with the same stem.
	for room_path in room_files:
		var room_dir   := room_path.get_base_dir()
		var stem       := room_path.get_file().get_basename()
		var rel        := room_path.replace(base, "")
		var display    := stem

		var item := _tree.create_item(section)
		item.set_text(0, display)
		item.set_icon(0, _icon_room)
		item.set_metadata(0, room_path)
		item.set_tooltip_text(0, rel)

		# Companion script sub-item.
		var script_path := room_dir.path_join(stem + ".agscript")
		if FileAccess.file_exists(script_path):
			var sub := _tree.create_item(item)
			sub.set_text(0, stem + ".agscript")
			sub.set_icon(0, _icon_script)
			sub.set_metadata(0, script_path)
			sub.set_tooltip_text(0, script_path.replace(base, ""))


func _add_leaf(parent: TreeItem, path: String, icon: Texture2D, base: String) -> void:
	var item := _tree.create_item(parent)
	item.set_text(0, path.get_file())
	item.set_icon(0, icon)
	item.set_metadata(0, path)
	item.set_tooltip_text(0, path.replace(base, ""))


# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

func _on_item_activated() -> void:
	var item := _tree.get_selected()
	if not item:
		return
	var path = item.get_metadata(0)
	if path:
		file_activated.emit(str(path))
		# Open in external editor for now; T-E09 will route rooms to the Room editor.
		if str(path).ends_with(".agscript"):
			EditorInterface.get_singleton().edit_script(
				load(ProjectSettings.localize_path(str(path))))
		else:
			OS.shell_open(str(path))


func _on_item_right_clicked(pos: Vector2, mouse_btn: int) -> void:
	if mouse_btn != MOUSE_BUTTON_RIGHT:
		return
	var item := _tree.get_selected()
	if not item:
		return
	# Placeholder — context menu will be expanded in future tasks.
	var menu := PopupMenu.new()
	menu.add_item("Open externally")
	menu.id_pressed.connect(func(_id): OS.shell_open(str(item.get_metadata(0))))
	add_child(menu)
	menu.popup(Rect2i(DisplayServer.mouse_get_position(), Vector2i.ZERO))
	menu.popup_hide.connect(menu.queue_free)


func _on_filesystem_changed() -> void:
	refresh()


# ---------------------------------------------------------------------------
# File scanner
# ---------------------------------------------------------------------------

## Returns absolute paths of all files with [param ext] under [param dir],
## excluding .godot and addons directories.
func _find_files(dir: String, ext: String) -> Array[String]:
	var result: Array[String] = []
	_scan_dir(dir, ext, result)
	return result


func _scan_dir(dir: String, ext: String, result: Array[String]) -> void:
	var da := DirAccess.open(dir)
	if not da:
		return
	da.list_dir_begin()
	var entry := da.get_next()
	while entry != "":
		if entry.begins_with("."):
			entry = da.get_next()
			continue
		var full := dir.path_join(entry)
		if da.current_is_dir():
			if entry != "addons":
				_scan_dir(full, ext, result)
		elif entry.ends_with(ext):
			result.append(full)
		entry = da.get_next()
	da.list_dir_end()
