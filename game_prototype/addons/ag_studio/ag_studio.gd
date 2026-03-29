@tool
extends EditorPlugin

## AG Studio EditorPlugin
##
## Entry point for the AG Studio editor layer. Responsible for:
##   - Hiding Godot's native docks (FileSystem, Scene, Import)
##   - Registering custom main screens (Room editor, Script editor)
##   - Registering custom docks (Project panel, Build Log)
##   - Respecting --godot-editor flag: does nothing when set

const PLUGIN_NAME := "AG Studio"

# True when launched with --godot-editor (skip all AG Studio setup).
# Uses OS.get_cmdline_args() so it works before the C++ flag is recompiled.
var _godot_editor_mode: bool = OS.get_cmdline_args().has("--godot-editor")

# Placeholder controls — replaced by real panels in T-E08 / T-E09.
var _project_panel: Control
var _build_log: Control
# _room_editor is not added until T-E09; not created here.

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
	print("[AGS] added Project dock")

	add_control_to_bottom_panel(_build_log, "Build Log")
	print("[AGS] added Build Log bottom panel")

	# Defer: editor layout is not fully built during _enter_tree().
	call_deferred("_hide_native_docks")
	print("[AGS] _enter_tree done — hide scheduled")


func _exit_tree() -> void:
	print("[AGS] _exit_tree: godot_editor_mode=%s" % _godot_editor_mode)
	if _godot_editor_mode:
		return

	if _project_panel:
		remove_control_from_docks(_project_panel)
		_project_panel.queue_free()
		_project_panel = null
		print("[AGS] removed Project dock")

	if _build_log:
		remove_control_from_bottom_panel(_build_log)
		_build_log.queue_free()
		_build_log = null
		print("[AGS] removed Build Log")

	_restore_native_docks()


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
# Dock management helpers
# ---------------------------------------------------------------------------

func _hide_native_docks() -> void:
	print("[AGS] _hide_native_docks called")
	var ei := get_editor_interface()

	# FileSystem dock — hide it and its parent container.
	var fs_dock := ei.get_file_system_dock()
	print("[AGS] FileSystem dock visible=%s, hiding" % fs_dock.visible)
	fs_dock.hide()
	# Hide the parent TabContainer too so the empty slot collapses.
	var fs_parent := fs_dock.get_parent()
	if fs_parent:
		fs_parent.hide()

	for title in ["Scene", "Import", "Inspector"]:
		var found := _set_dock_tab_hidden(title, true)
		print("[AGS] hide dock '%s': found=%s" % [title, found])


func _restore_native_docks() -> void:
	print("[AGS] _restore_native_docks called")
	var ei := get_editor_interface()

	var fs_dock := ei.get_file_system_dock()
	var fs_parent := fs_dock.get_parent()
	if fs_parent:
		fs_parent.show()
	fs_dock.show()

	for title in ["Scene", "Import", "Inspector"]:
		_set_dock_tab_hidden(title, false)


## Hide or show a single named tab and collapse/restore its container.
func _set_dock_tab_hidden(title: String, hidden: bool) -> bool:
	var base := get_editor_interface().get_base_control()
	return _walk_for_dock_tab(base, title, hidden)


func _walk_for_dock_tab(node: Node, title: String, hidden: bool) -> bool:
	if node is TabContainer:
		var tc := node as TabContainer
		for i in tc.get_tab_count():
			if tc.get_tab_title(i) == title:
				tc.set_tab_hidden(i, hidden)
				print("[AGS]   tab '%s' set_tab_hidden=%s in %s" % [title, hidden, tc.name])
				# Hide the container itself when all its tabs are hidden.
				if hidden:
					var any_visible := false
					for j in tc.get_tab_count():
						if not tc.is_tab_hidden(j):
							any_visible = true
							break
					if not any_visible:
						tc.hide()
						print("[AGS]   container %s hidden (all tabs hidden)" % tc.name)
				else:
					tc.show()
				return true
	for child in node.get_children():
		if _walk_for_dock_tab(child, title, hidden):
			return true
	return false


# ---------------------------------------------------------------------------
# Placeholder factory
# ---------------------------------------------------------------------------

func _make_placeholder(label: String) -> Control:
	var p := PanelContainer.new()
	p.name = label.replace(" ", "")
	var lbl := Label.new()
	lbl.text = label + "\n(not yet implemented)"
	lbl.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	lbl.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	p.add_child(lbl)
	return p
