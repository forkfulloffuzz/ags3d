@tool
extends EditorPlugin

## AG Studio EditorPlugin
##
## Entry point for the AG Studio editor layer. Uses a whitelist approach:
## everything NOT in the keep-list is hidden, so new Godot panels are
## automatically suppressed without having to name them explicitly.
##
## Whitelist sections:
##   KEEP_DOCK_TABS    — tab titles to keep visible in side-dock containers
##   KEEP_BOTTOM_TABS  — tab titles to keep in the bottom panel bar
##   KEEP_MAIN_SCREENS — top-bar main-screen button labels to keep
##   KEEP_MENUS        — top menu bar entries to keep
##
## Respects --godot-editor flag: skips all customisation when set.

const PLUGIN_NAME := "AG Studio"

# Dock tabs to keep (everything else is hidden).
# "Project" is our own panel added via add_control_to_dock().
const KEEP_DOCK_TABS: Array[String] = ["Project"]

# Bottom-panel tabs to keep.
# "Build Log" is our placeholder added via add_control_to_bottom_panel().
const KEEP_BOTTOM_TABS: Array[String] = ["Build Log", "Output"]

# Top main-screen buttons to keep (label text on the buttons).
# "Room" will be added in T-E09. Until then only keep nothing AG-specific;
# hide 2D, 3D, Script, Game, AssetLib.
const KEEP_MAIN_SCREENS: Array[String] = []

# Top menu bar entries to keep.
const KEEP_MENUS: Array[String] = []

# True when launched with --godot-editor.
var _godot_editor_mode: bool = OS.get_cmdline_args().has("--godot-editor")

var _project_panel: Control
var _build_log: Control

# Nodes hidden by us so we can restore them on exit.
var _hidden_nodes: Array[Node] = []


func _enter_tree() -> void:
	print("[AGS] _enter_tree: godot_editor_mode=%s" % _godot_editor_mode)
	if _godot_editor_mode:
		return

	var pp: VBoxContainer = preload("res://addons/ag_studio/project_panel.gd").new()
	pp.size_flags_vertical = Control.SIZE_EXPAND_FILL
	pp.set_plugin(self)
	_project_panel = pp

	_build_log = _make_placeholder("Build Log")

	add_control_to_dock(DOCK_SLOT_LEFT_UL, _project_panel)
	add_control_to_bottom_panel(_build_log, "Build Log")

	call_deferred("_apply_whitelist")


func _exit_tree() -> void:
	print("[AGS] _exit_tree: godot_editor_mode=%s" % _godot_editor_mode)
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

	_restore_all()


# ---------------------------------------------------------------------------
# EditorPlugin overrides
# ---------------------------------------------------------------------------

func _has_main_screen() -> bool:
	return false  # placeholder until T-E09


func _get_plugin_name() -> String:
	return PLUGIN_NAME


func _get_plugin_icon() -> Texture2D:
	return get_editor_interface().get_base_control().get_theme_icon("Node", "EditorIcons")


# ---------------------------------------------------------------------------
# Whitelist application
# ---------------------------------------------------------------------------

func _apply_whitelist() -> void:
	print("[AGS] _apply_whitelist")
	var base: Control = get_editor_interface().get_base_control()

	# FileSystem dock (no tab — direct child, hide explicitly).
	var fs_dock: FileSystemDock = get_editor_interface().get_file_system_dock()
	_hide_node(fs_dock)
	_hide_node(fs_dock.get_parent())

	# Walk the full editor tree and apply whitelists.
	_walk(base)

	# Top main-screen buttons.
	_apply_main_screen_whitelist(base)

	# Top menu bar.
	_apply_menu_whitelist(base)

	# Play/pause/stop toolbar + renderer selector.
	_apply_play_toolbar_whitelist(base)


func _walk(node: Node) -> void:
	if node is TabContainer:
		var tc := node as TabContainer
		_apply_tab_whitelist(tc)
		# Don't recurse into tab containers whose children we just processed.
		return
	for child in node.get_children():
		_walk(child)


func _apply_tab_whitelist(tc: TabContainer) -> void:
	# Determine which whitelist applies based on the container's role.
	# Bottom panel containers sit inside a node named "BottomPanel" or similar.
	# We distinguish by checking whether any existing tab is a known bottom tab.
	var is_bottom := _is_bottom_panel(tc)
	var keep: Array[String] = KEEP_BOTTOM_TABS if is_bottom else KEEP_DOCK_TABS

	var any_visible := false
	for i in tc.get_tab_count():
		var title := tc.get_tab_title(i)
		var should_hide: bool = not (title in keep)
		if should_hide and not tc.is_tab_hidden(i):
			tc.set_tab_hidden(i, true)
			print("[AGS]   hide tab '%s' in %s" % [title, tc.name])
		elif not should_hide:
			any_visible = true

	if not any_visible:
		_hide_node(tc)


func _is_bottom_panel(tc: TabContainer) -> bool:
	# Walk up looking for a node whose name contains "Bottom".
	var n: Node = tc
	for _i in 6:
		if n == null:
			break
		if "Bottom" in n.name or "bottom" in n.name:
			return true
		n = n.get_parent()
	return false


func _apply_main_screen_whitelist(base: Control) -> void:
	# Main-screen buttons are Button nodes inside a HBoxContainer that contains
	# the 2D/3D/Script/Game/AssetLib buttons. Find by looking for buttons whose
	# text matches known screen names.
	var known := ["2D", "3D", "Script", "Game", "AssetLib"]
	_walk_for_buttons(base, known, KEEP_MAIN_SCREENS)


func _walk_for_buttons(node: Node, known: Array, keep: Array[String]) -> void:
	if node is Button:
		var btn := node as Button
		if btn.text in known and not (btn.text in keep):
			_hide_node(btn)
			print("[AGS]   hide main-screen button '%s'" % btn.text)
	for child in node.get_children():
		_walk_for_buttons(child, known, keep)


func _apply_menu_whitelist(base: Control) -> void:
	# MenuBar or top-level MenuButton nodes.
	_walk_for_menus(base)


func _apply_play_toolbar_whitelist(base: Control) -> void:
	# The play/stop/renderer bar has no stable name, so we find it by locating
	# a Button whose tooltip contains a known play-bar keyword, then hide its
	# nearest HBoxContainer or VBoxContainer ancestor that sits directly inside
	# the title bar row (depth <= 6 from base).
	var keywords := ["Run Project", "Pause Scene", "Stop", "Remote Debug",
					  "Play Current Scene", "Play Custom Scene", "Movie Maker Mode",
					  "Forward+", "Mobile", "Compatibility"]
	var container := _find_play_toolbar(base, keywords)
	if container:
		_hide_node(container)
	else:
		print("[AGS]   play toolbar not found — hiding by tooltip scan")
		_hide_buttons_by_tooltips(base, keywords)


func _walk_for_menus(node: Node) -> void:
	if node is MenuBar:
		var mb := node as MenuBar
		for i in mb.get_menu_count():
			var title := mb.get_menu_title(i)
			if not (title in KEEP_MENUS):
				mb.set_menu_hidden(i, true)
				print("[AGS]   hide menu '%s'" % title)
		return
	for child in node.get_children():
		_walk_for_menus(child)


## Find the container that holds the play toolbar by locating a button whose
## tooltip contains one of [param keywords], then walking up to find the
## highest BoxContainer ancestor within [param max_depth] levels of [param base].
func _find_play_toolbar(base: Control, keywords: Array) -> Node:
	return _walk_find_play_container(base, keywords, base)


func _walk_find_play_container(node: Node, keywords: Array, base: Node) -> Node:
	if node is Button or node is MenuButton:
		var tooltip: String = (node as Control).tooltip_text
		for kw: String in keywords:
			if kw.to_lower() in tooltip.to_lower():
				# Walk up from this button to find the direct BoxContainer
				# child of the title-bar row.
				return _ancestor_box_container(node, base)
	for child in node.get_children():
		var result := _walk_find_play_container(child, keywords, base)
		if result:
			return result
	return null


func _ancestor_box_container(node: Node, stop: Node) -> Node:
	# Return the highest HBoxContainer/VBoxContainer ancestor before stop.
	var best: Node = null
	var cur: Node = node.get_parent()
	while cur and cur != stop:
		if cur is HBoxContainer or cur is VBoxContainer:
			best = cur
		cur = cur.get_parent()
	return best if best else node.get_parent()


func _hide_buttons_by_tooltips(node: Node, keywords: Array) -> void:
	if node is Button or node is MenuButton:
		var tooltip: String = (node as Control).tooltip_text
		for kw: String in keywords:
			if kw.to_lower() in tooltip.to_lower():
				_hide_node(node)
				print("[AGS]   hide button tooltip='%s'" % tooltip)
				break
	for child in node.get_children():
		_hide_buttons_by_tooltips(child, keywords)


# ---------------------------------------------------------------------------
# Restore
# ---------------------------------------------------------------------------

func _restore_all() -> void:
	print("[AGS] _restore_all: restoring %d nodes" % _hidden_nodes.size())
	for n in _hidden_nodes:
		if is_instance_valid(n):
			n.show()
	_hidden_nodes.clear()

	# Restore all tab containers: un-hide every tab.
	var base: Control = get_editor_interface().get_base_control()
	_restore_tabs(base)

	# Restore menu bar.
	_restore_menus(base)


func _restore_tabs(node: Node) -> void:
	if node is TabContainer:
		var tc := node as TabContainer
		for i in tc.get_tab_count():
			tc.set_tab_hidden(i, false)
		tc.show()
		return
	for child in node.get_children():
		_restore_tabs(child)


func _restore_menus(node: Node) -> void:
	if node is MenuBar:
		var mb := node as MenuBar
		for i in mb.get_menu_count():
			mb.set_menu_hidden(i, false)
		return
	for child in node.get_children():
		_restore_menus(child)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

func _hide_node(node: Node) -> void:
	if node == null or not is_instance_valid(node):
		return
	if node is CanvasItem and (node as CanvasItem).visible:
		(node as CanvasItem).hide()
		_hidden_nodes.append(node)
		print("[AGS]   hide node %s" % node.name)


func _make_placeholder(label: String) -> Control:
	var p := PanelContainer.new()
	p.name = label.replace(" ", "")
	var lbl := Label.new()
	lbl.text = label + "\n(not yet implemented)"
	lbl.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	lbl.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	p.add_child(lbl)
	return p
