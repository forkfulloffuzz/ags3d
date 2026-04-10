@tool
extends VBoxContainer

## AG Studio Project Panel dock
##
## Displays a tree of all AG source files grouped by type:
##   Rooms      — .agroom + companion .agscript pairs
##   Characters — .agchar files
##   Scripts    — .agscript files not paired with a room
##
## Call set_plugin(plugin) immediately after instantiation.

signal file_activated(path: String)

var _plugin: EditorPlugin  # set by ag_studio.gd

var _tree: Tree
var _refresh_btn: Button

var _icon_room:   Texture2D
var _icon_char:   Texture2D
var _icon_script: Texture2D
var _icon_gui:    Texture2D
var _icon_folder: Texture2D


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "Project"
	_build_ui()
	# Icons require _plugin, which is set after _ready() via set_plugin().
	# Defer so they are available.
	call_deferred("_fetch_icons_and_refresh")
	# Auto-refresh when the filesystem changes.
	var efs: EditorFileSystem = _plugin.get_editor_interface().get_resource_filesystem()
	efs.filesystem_changed.connect(refresh)


func _build_ui() -> void:
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

	var sep := HSeparator.new()
	add_child(sep)

	_tree = Tree.new()
	_tree.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_tree.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_tree.hide_root = true
	_tree.select_mode = Tree.SELECT_ROW
	_tree.item_activated.connect(_on_item_activated)
	_tree.item_mouse_selected.connect(_on_item_right_clicked)
	add_child(_tree)


func _fetch_icons_and_refresh() -> void:
	var base: Control = _plugin.get_editor_interface().get_base_control()
	_icon_room   = base.get_theme_icon("Node3D",          "EditorIcons")
	_icon_char   = base.get_theme_icon("CharacterBody3D", "EditorIcons")
	_icon_script = base.get_theme_icon("Script",          "EditorIcons")
	_icon_gui    = base.get_theme_icon("Control",         "EditorIcons")
	_icon_folder = base.get_theme_icon("Folder",          "EditorIcons")
	refresh()


## Rebuild the tree from the project filesystem.
func refresh() -> void:
	if not _tree or not _plugin:
		return
	_tree.clear()
	var root: TreeItem = _tree.create_item()

	var abs_base: String = ProjectSettings.globalize_path("res://")

	var rooms:           Array[String] = _find_files(abs_base, ".agroom")
	var chars:           Array[String] = _find_files(abs_base, ".agchar")
	var all_scripts:     Array[String] = _find_files(abs_base, ".agscript")
	var guis:            Array[String] = _find_files(abs_base, ".agui")

	# Scripts that share a stem with a .agroom are room scripts — exclude them.
	var room_stems: Dictionary = {}
	for p: String in rooms:
		room_stems[p.get_basename()] = true

	var standalone_scripts: Array[String] = []
	for p: String in all_scripts:
		if not room_stems.has(p.get_basename()):
			standalone_scripts.append(p)

	_populate_rooms_section(root, rooms, abs_base)
	_populate_section(root, "Characters", chars,              _icon_char,   abs_base)
	_populate_section(root, "Scripts",    standalone_scripts, _icon_script, abs_base)
	_populate_section(root, "GUI",       guis,               _icon_gui,    abs_base)


# ---------------------------------------------------------------------------
# Tree population
# ---------------------------------------------------------------------------

func _populate_rooms_section(root: TreeItem, files: Array[String], abs_base: String) -> void:
	if files.is_empty():
		return
	var section: TreeItem = _tree.create_item(root)
	section.set_text(0, "Rooms")
	section.set_icon(0, _icon_folder)
	section.set_selectable(0, false)

	for room_path: String in files:
		var room_dir: String  = room_path.get_base_dir()
		var stem: String      = room_path.get_file().get_basename()

		var item: TreeItem = _tree.create_item(section)
		item.set_text(0, stem)
		item.set_icon(0, _icon_room)
		item.set_metadata(0, room_path)
		item.set_tooltip_text(0, room_path.replace(abs_base, ""))

		var script_path: String = room_dir.path_join(stem + ".agscript")
		if FileAccess.file_exists(script_path):
			var sub: TreeItem = _tree.create_item(item)
			sub.set_text(0, stem + ".agscript")
			sub.set_icon(0, _icon_script)
			sub.set_metadata(0, script_path)
			sub.set_tooltip_text(0, script_path.replace(abs_base, ""))


func _populate_section(root: TreeItem, label: String, files: Array[String],
		icon: Texture2D, abs_base: String) -> void:
	if files.is_empty():
		return
	var section: TreeItem = _tree.create_item(root)
	section.set_text(0, label)
	section.set_icon(0, _icon_folder)
	section.set_selectable(0, false)

	for path: String in files:
		var item: TreeItem = _tree.create_item(section)
		item.set_text(0, path.get_file())
		item.set_icon(0, icon)
		item.set_metadata(0, path)
		item.set_tooltip_text(0, path.replace(abs_base, ""))


# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

func _on_item_activated() -> void:
	var item: TreeItem = _tree.get_selected()
	if not item:
		return
	var path: String = str(item.get_metadata(0))
	if path.is_empty():
		return
	file_activated.emit(path)
	if path.ends_with(".agscript"):
		pass  # routed to Script editor by ag_studio._on_file_activated
	elif path.ends_with(".agroom"):
		pass  # routed to Room editor by ag_studio._on_file_activated
	elif path.ends_with(".agchar"):
		pass  # routed to Character editor by ag_studio._on_file_activated
	elif path.ends_with(".agui"):
		pass  # routed to GUI editor by ag_studio._on_file_activated
	else:
		OS.shell_open(path)


func _on_item_right_clicked(pos: Vector2, mouse_btn: int) -> void:
	if mouse_btn != MOUSE_BUTTON_RIGHT:
		return
	var item: TreeItem = _tree.get_selected()
	if not item:
		return
	var path: String = str(item.get_metadata(0))
	var menu := PopupMenu.new()
	menu.add_item("Open externally")
	menu.id_pressed.connect(func(_id: int) -> void: OS.shell_open(path))
	add_child(menu)
	menu.popup(Rect2i(DisplayServer.mouse_get_position(), Vector2i.ZERO))
	menu.popup_hide.connect(menu.queue_free)


# ---------------------------------------------------------------------------
# File scanner
# ---------------------------------------------------------------------------

func _find_files(dir: String, ext: String) -> Array[String]:
	var result: Array[String] = []
	_scan_dir(dir, ext, result)
	return result


func _scan_dir(dir: String, ext: String, result: Array[String]) -> void:
	var da: DirAccess = DirAccess.open(dir)
	if not da:
		return
	da.list_dir_begin()
	var entry: String = da.get_next()
	while entry != "":
		if entry.begins_with("."):
			entry = da.get_next()
			continue
		var full: String = dir.path_join(entry)
		if da.current_is_dir():
			if entry != "addons":
				_scan_dir(full, ext, result)
		elif entry.ends_with(ext):
			result.append(full)
		entry = da.get_next()
	da.list_dir_end()
