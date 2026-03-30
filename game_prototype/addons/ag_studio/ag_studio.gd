@tool
extends EditorPlugin

## AG Studio EditorPlugin — whitelist approach.
##
## KEEP_DOCK_TABS    — side-dock tab titles to keep
## KEEP_BOTTOM_TABS  — bottom-panel tab titles to keep
## KEEP_MAIN_SCREENS — top-bar main-screen buttons to keep (empty = hide all)
## KEEP_MENUS        — top MenuBar entries to keep (empty = hide all)
##
## Everything else is hidden. The walk skips Window subclasses (dialogs)
## so it never touches popup internals.

const PLUGIN_NAME := "AG Studio"

const KEEP_DOCK_TABS:    Array[String] = ["Project"]
const KEEP_BOTTOM_TABS:  Array[String] = ["Build Log", "Output"]
const KEEP_MAIN_SCREENS: Array[String] = []
const KEEP_MENUS:        Array[String] = []

var _godot_editor_mode: bool = OS.get_cmdline_args().has("--godot-editor")

var _project_panel: Control
var _build_log: Control
var _room_editor: Control
var _hidden_nodes: Array[Node] = []
var _gizmo_plugins: Array[EditorNode3DGizmoPlugin] = []

const _GIZMO_SCRIPTS := [
	"res://addons/ag_studio/gizmos/ags_walkable_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_blocker_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_hotspot_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_trigger_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_point_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_spawn_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_camera_gizmo.gd",
]


func _enter_tree() -> void:
	if _godot_editor_mode:
		return

	var pp: VBoxContainer = preload("res://addons/ag_studio/project_panel.gd").new()
	pp.size_flags_vertical = Control.SIZE_EXPAND_FILL
	pp.set_plugin(self)
	_project_panel = pp
	_project_panel.file_activated.connect(_on_file_activated)

	_build_log = _make_placeholder("Build Log")

	# Room editor main screen
	var re: Control = preload("res://addons/ag_studio/room_editor.gd").new()
	re.set_plugin(self)
	re.visible = false
	_room_editor = re
	get_editor_interface().get_editor_main_screen().add_child(_room_editor)
	_room_editor.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	# Gizmo plugins
	var ur := get_undo_redo()
	for path: String in _GIZMO_SCRIPTS:
		var plugin: EditorNode3DGizmoPlugin = load(path).new()
		if plugin.has_method("setup"):
			plugin.setup(ur)
		add_node_3d_gizmo_plugin(plugin)
		_gizmo_plugins.append(plugin)

	add_control_to_dock(DOCK_SLOT_LEFT_UL, _project_panel)
	add_control_to_bottom_panel(_build_log, "Build Log")

	call_deferred("_apply_whitelist")


func _exit_tree() -> void:
	if _godot_editor_mode:
		return

	if _project_panel:
		remove_control_from_docks(_project_panel)
		_project_panel.queue_free()
		_project_panel = null

	if _build_log:
		remove_control_from_bottom_panel(_build_log)
		_build_log.queue_free()
		_build_log = null

	if _room_editor:
		_room_editor.queue_free()
		_room_editor = null

	for plugin: EditorNode3DGizmoPlugin in _gizmo_plugins:
		remove_node_3d_gizmo_plugin(plugin)
	_gizmo_plugins.clear()

	_restore_all()


# ---------------------------------------------------------------------------
# EditorPlugin overrides
# ---------------------------------------------------------------------------

func _has_main_screen() -> bool:
	return true

func _make_visible(visible: bool) -> void:
	if _room_editor:
		_room_editor.visible = visible

func _get_plugin_name() -> String:
	return PLUGIN_NAME

func _get_plugin_icon() -> Texture2D:
	return get_editor_interface().get_base_control().get_theme_icon("Node", "EditorIcons")


# ---------------------------------------------------------------------------
# Whitelist application
# ---------------------------------------------------------------------------

func _apply_whitelist() -> void:
	var base: Control = get_editor_interface().get_base_control()

	# FileSystem dock — hide control and its parent slot.
	var fs_dock: FileSystemDock = get_editor_interface().get_file_system_dock()
	_hide_node(fs_dock)
	_hide_node(fs_dock.get_parent())

	# Scene dock (node tree) — no public accessor, find by class.
	var scene_dock := _find_by_class(base, "SceneTreeDock")
	if scene_dock:
		_hide_node(scene_dock)
		_hide_node(scene_dock.get_parent())

	# Walk the main editor tree — skip Window subclasses (dialogs/popups).
	_walk(base)

	# Main-screen buttons (2D / 3D / Script / Game / AssetLib).
	_apply_main_screen_whitelist(base)

	# Top MenuBar.
	_apply_menu_whitelist(base)

	# Play/stop toolbar and renderer selector.
	_apply_play_toolbar_whitelist(base)

	# Log any TabContainer tabs still visible (helps catch stragglers).
	_log_visible_tabs(base)


## Recursively walk [param node], skipping Window subclasses and their subtrees.
func _walk(node: Node) -> void:
	# Never touch dialog/popup internals — they trigger internal errors when
	# their tabs are hidden before they have been properly initialised.
	if node is Window:
		return
	if node is TabContainer:
		_apply_tab_whitelist(node as TabContainer)
		return  # don't recurse — tabs' children were handled above
	for child in node.get_children():
		_walk(child)


func _apply_tab_whitelist(tc: TabContainer) -> void:
	var keep: Array[String] = KEEP_BOTTOM_TABS if _is_bottom_panel(tc) else KEEP_DOCK_TABS
	var any_visible := false
	for i in tc.get_tab_count():
		var title := tc.get_tab_title(i)
		if title in keep:
			any_visible = true
		elif not tc.is_tab_hidden(i):
			tc.set_tab_hidden(i, true)
	if not any_visible:
		_hide_node(tc)


func _is_bottom_panel(tc: TabContainer) -> bool:
	var n: Node = tc
	for _i in 6:
		if n == null:
			break
		if "Bottom" in n.name or "bottom" in n.name:
			return true
		n = n.get_parent()
	return false


func _apply_main_screen_whitelist(base: Control) -> void:
	var known := ["2D", "3D", "Script", "Game", "AssetLib", PLUGIN_NAME]
	_walk_for_buttons(base, known, KEEP_MAIN_SCREENS)


func _walk_for_buttons(node: Node, known: Array, keep: Array[String]) -> void:
	if node is Window:
		return
	if node is Button:
		var btn := node as Button
		if btn.text in known and not (btn.text in keep):
			_hide_node(btn)
	for child in node.get_children():
		_walk_for_buttons(child, known, keep)


func _apply_menu_whitelist(base: Control) -> void:
	_walk_for_menus(base)


func _walk_for_menus(node: Node) -> void:
	if node is Window:
		return
	if node is MenuBar:
		var mb := node as MenuBar
		for i in mb.get_menu_count():
			if not (mb.get_menu_title(i) in KEEP_MENUS):
				mb.set_menu_hidden(i, true)
		return
	for child in node.get_children():
		_walk_for_menus(child)


func _apply_play_toolbar_whitelist(base: Control) -> void:
	# EditorRunBar is the play/stop/remote bar — find by class name.
	var run_bar := _find_by_class(base, "EditorRunBar")
	if run_bar:
		_hide_node(run_bar)

	# The renderer OptionButton lives in a sibling HBoxContainer of EditorRunBar
	# inside EditorTitleBar. Hide the whole sibling container.
	if run_bar:
		var title_bar := run_bar.get_parent()
		if title_bar:
			for child in title_bar.get_children():
				if child != run_bar and (child is HBoxContainer or child is VBoxContainer):
					_hide_node(child)


func _log_visible_tabs(node: Node) -> void:
	if node is Window:
		return
	if node is TabContainer:
		var tc := node as TabContainer
		if tc.visible:
			for i in tc.get_tab_count():
				if not tc.is_tab_hidden(i):
					print("[AGS] VISIBLE tab '%s' in %s" % [tc.get_tab_title(i), tc.name])
		return
	for child in node.get_children():
		_log_visible_tabs(child)


func _find_by_class(node: Node, cls: String) -> Node:
	if node.get_class() == cls:
		return node
	if node is Window:
		return null
	for child in node.get_children():
		var result := _find_by_class(child, cls)
		if result:
			return result
	return null


# ---------------------------------------------------------------------------
# Restore
# ---------------------------------------------------------------------------

func _restore_all() -> void:
	for n in _hidden_nodes:
		if is_instance_valid(n):
			n.show()
	_hidden_nodes.clear()

	var base: Control = get_editor_interface().get_base_control()
	_restore_tabs(base)
	_restore_menus(base)


func _restore_tabs(node: Node) -> void:
	if node is Window:
		return
	if node is TabContainer:
		var tc := node as TabContainer
		# Allow deselection on the underlying TabBar before un-hiding tabs.
		# set_tab_hidden(i, false) internally tries to adjust the current tab
		# selection, which triggers "Cannot deselect tabs" if deselect_enabled
		# is false on the TabBar.
		var tb := tc.get_tab_bar()
		tb.deselect_enabled = true
		for i in tc.get_tab_count():
			tc.set_tab_hidden(i, false)
		tc.show()
		return
	for child in node.get_children():
		_restore_tabs(child)


func _restore_menus(node: Node) -> void:
	if node is Window:
		return
	if node is MenuBar:
		var mb := node as MenuBar
		for i in mb.get_menu_count():
			mb.set_menu_hidden(i, false)
		return
	for child in node.get_children():
		_restore_menus(child)


# ---------------------------------------------------------------------------
# File activation routing
# ---------------------------------------------------------------------------

func _on_file_activated(abs_path: String) -> void:
	if not abs_path.ends_with(".agroom"):
		return
	# .agroom is not a Godot resource — load the generated .tscn instead.
	var tscn_abs: String = abs_path.get_basename() + ".tscn"
	var tscn_res: String = ProjectSettings.localize_path(tscn_abs)
	get_editor_interface().set_main_screen_editor(PLUGIN_NAME)
	if _room_editor:
		(_room_editor as Node).call("load_room", tscn_res)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

func _hide_node(node: Node) -> void:
	if node == null or not is_instance_valid(node):
		return
	if node is CanvasItem and (node as CanvasItem).visible:
		(node as CanvasItem).hide()
		_hidden_nodes.append(node)


func _make_placeholder(label: String) -> Control:
	var p := PanelContainer.new()
	p.name = label.replace(" ", "")
	var lbl := Label.new()
	lbl.text = label + "\n(not yet implemented)"
	lbl.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	lbl.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	p.add_child(lbl)
	return p
