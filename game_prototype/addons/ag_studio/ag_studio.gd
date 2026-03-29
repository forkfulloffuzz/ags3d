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
var _room_editor: Control

func _enter_tree() -> void:
	# --godot-editor: skip all AG Studio setup, leave the standard editor intact.
	if _godot_editor_mode:
		return

	_project_panel = _make_placeholder("Project")
	_build_log     = _make_placeholder("Build Log")
	_room_editor   = _make_placeholder("Room Editor")

	# Register custom docks.
	add_control_to_dock(DOCK_SLOT_LEFT_UL, _project_panel)
	add_control_to_bottom_panel(_build_log, "Build Log")

	# Hide native Godot docks that AG Studio replaces.
	_hide_native_docks()


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
		remove_control_from_container(CustomControlContainer.CONTAINER_TOOLBAR, _room_editor)
		_room_editor.queue_free()
		_room_editor = null

	# Restore native docks when plugin is deactivated.
	_restore_native_docks()


# ---------------------------------------------------------------------------
# EditorPlugin overrides
# ---------------------------------------------------------------------------

func _has_main_screen() -> bool:
	# AG Studio owns its own main screens (Room editor, Script editor).
	# Returning true here reserves a slot; the actual screen control is
	# registered via add_editor_plugin_screen() in T-E09.
	return false  # placeholder until T-E09


func _get_plugin_name() -> String:
	return PLUGIN_NAME


func _get_plugin_icon() -> Texture2D:
	# Use a built-in icon until a custom one is provided.
	return get_editor_interface().get_base_control().get_theme_icon("Node", "EditorIcons")


# ---------------------------------------------------------------------------
# Dock management helpers
# ---------------------------------------------------------------------------

func _hide_native_docks() -> void:
	var ei := get_editor_interface()

	# FileSystem dock — direct API accessor.
	ei.get_file_system_dock().hide()

	# Scene and Import docks — hidden by searching the dock container by title.
	for title in ["Scene", "Import"]:
		_set_dock_visible_by_title(title, false)


func _restore_native_docks() -> void:
	var ei := get_editor_interface()
	ei.get_file_system_dock().show()
	for title in ["Scene", "Import"]:
		_set_dock_visible_by_title(title, true)


## Walk the editor tree looking for a dock tab whose text matches [param title]
## and set its visibility. Godot exposes no direct API for this, so we iterate
## the dock containers (TabContainer nodes inside the main editor VBoxContainer).
func _set_dock_visible_by_title(title: String, visible: bool) -> void:
	var base := get_editor_interface().get_base_control()
	_walk_for_dock_tab(base, title, visible)


func _walk_for_dock_tab(node: Node, title: String, visible: bool) -> bool:
	if node is TabContainer:
		for i in node.get_tab_count():
			if node.get_tab_title(i) == title:
				var child: Control = node.get_tab_control(i)
				if child:
					child.visible = visible
				return true
	for child in node.get_children():
		if _walk_for_dock_tab(child, title, visible):
			return true
	return false


# ---------------------------------------------------------------------------
# Placeholder factory — returns a minimal labeled panel for scaffolding.
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
