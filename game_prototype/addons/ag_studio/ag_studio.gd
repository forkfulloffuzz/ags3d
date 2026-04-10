@tool
extends EditorPlugin

## AG Studio EditorPlugin — additive approach.
##
## AG Studio augments the Godot editor with AGS3D-specific panels and gizmos.
## Native Godot UI (3D viewport, Scene tree, FileSystem, inspector, etc.) is
## preserved and works alongside AG Studio panels.
##
## Use --godot-editor to launch the standard Godot editor without AG Studio
## panels (useful for debugging).

const PLUGIN_NAME := "AG Studio"

var _godot_editor_mode: bool = OS.get_cmdline_args().has("--godot-editor")

var _project_panel: Control
var _build_log: Control
var _room_editor: Control
var _char_editor: Control
var _script_editor: Control
var _item_editor: Control
var _anim_viewer_3d: Control
var _anim_viewer_2d: Control
var _menu_btn: MenuButton
var _play_btn: Button
var _settings_dialog: ConfirmationDialog
var _wizard: ConfirmationDialog
var _gizmo_plugins: Array[EditorNode3DGizmoPlugin] = []
var _inspector_plugin: EditorInspectorPlugin

const _GIZMO_SCRIPTS := [
	"res://addons/ag_studio/gizmos/ags_walkable_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_blocker_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_hotspot_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_trigger_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_point_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_spawn_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_camera_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_item_gizmo.gd",
]


func _enter_tree() -> void:
	var ur := get_undo_redo()
	for path: String in _GIZMO_SCRIPTS:
		var plugin: EditorNode3DGizmoPlugin = load(path).new()
		if plugin.has_method("setup"):
			plugin.setup(ur)
		add_node_3d_gizmo_plugin(plugin)
		_gizmo_plugins.append(plugin)

	var ip: EditorInspectorPlugin = preload("res://addons/ag_studio/ags_inspector_plugin.gd").new()
	ip.setup(self)
	_inspector_plugin = ip
	add_inspector_plugin(_inspector_plugin)

	if _godot_editor_mode:
		return

	var bl: VBoxContainer = preload("res://addons/ag_studio/build_log.gd").new()
	bl.set_plugin(self)
	_build_log = bl
	add_control_to_bottom_panel(_build_log, "Build Log")

	var pp: VBoxContainer = preload("res://addons/ag_studio/project_panel.gd").new()
	pp.size_flags_vertical = Control.SIZE_EXPAND_FILL
	pp.set_plugin(self)
	_project_panel = pp
	_project_panel.file_activated.connect(_on_file_activated)
	add_control_to_dock(DOCK_SLOT_LEFT_UL, _project_panel)

	var re: Control = preload("res://addons/ag_studio/room_editor.gd").new()
	re.set_plugin(self)
	_room_editor = re
	add_control_to_container(CONTAINER_SPATIAL_EDITOR_SIDE_LEFT, _room_editor)

	var ie: Control = preload("res://addons/ag_studio/item_editor.gd").new()
	ie.set_plugin(self)
	_item_editor = ie
	add_control_to_dock(DOCK_SLOT_RIGHT_UL, _item_editor)

	var av3d: Control = preload("res://addons/ag_studio/anim_viewer_3d.gd").new()
	av3d.set_plugin(self)
	_anim_viewer_3d = av3d

	var av2d: Control = preload("res://addons/ag_studio/anim_viewer_2d.gd").new()
	av2d.set_plugin(self)
	_anim_viewer_2d = av2d

	var ce: Control = preload("res://addons/ag_studio/char_editor.gd").new()
	ce.set_plugin(self)
	ce.set_anim_viewer(av3d)
	_char_editor = ce
	add_control_to_bottom_panel(_char_editor, "Character")

	_menu_btn = MenuButton.new()
	_menu_btn.text = "AG Studio"
	var menu := _menu_btn.get_popup()
	menu.add_item("New Project…", 0)
	menu.add_separator()
	menu.add_item("Build", 1)
	menu.add_separator()
	menu.add_item("Settings…", 2)
	menu.id_pressed.connect(_on_menu_item)
	add_control_to_container(CONTAINER_TOOLBAR, _menu_btn)

	_play_btn = Button.new()
	_play_btn.text = "▶ Play"
	_play_btn.tooltip_text = "Build project then play main scene (F5)"
	_play_btn.pressed.connect(_on_play_pressed)
	add_control_to_container(CONTAINER_TOOLBAR, _play_btn)

	get_editor_interface().scene_changed.connect(_on_scene_changed)


func _exit_tree() -> void:
	for plugin: EditorNode3DGizmoPlugin in _gizmo_plugins:
		remove_node_3d_gizmo_plugin(plugin)
	_gizmo_plugins.clear()

	if _inspector_plugin:
		remove_inspector_plugin(_inspector_plugin)
		_inspector_plugin = null

	if _godot_editor_mode:
		return

	if _project_panel:
		remove_control_from_docks(_project_panel)
		_project_panel.queue_free()
		_project_panel = null

	if _room_editor:
		remove_control_from_container(CONTAINER_SPATIAL_EDITOR_SIDE_LEFT, _room_editor)
		_room_editor.queue_free()
		_room_editor = null

	if _item_editor:
		remove_control_from_docks(_item_editor)
		_item_editor.queue_free()
		_item_editor = null

	if _anim_viewer_3d:
		_anim_viewer_3d.queue_free()
		_anim_viewer_3d = null

	if _anim_viewer_2d:
		_anim_viewer_2d.queue_free()
		_anim_viewer_2d = null

	if _char_editor:
		remove_control_from_bottom_panel(_char_editor)
		_char_editor.queue_free()
		_char_editor = null

	if _menu_btn:
		remove_control_from_container(CONTAINER_TOOLBAR, _menu_btn)
		_menu_btn.queue_free()
		_menu_btn = null

	if _wizard:
		_wizard.queue_free()
		_wizard = null

	if _settings_dialog:
		_settings_dialog.queue_free()
		_settings_dialog = null

	if _play_btn:
		remove_control_from_container(CONTAINER_TOOLBAR, _play_btn)
		_play_btn.queue_free()
		_play_btn = null

	if _build_log:
		remove_control_from_bottom_panel(_build_log)
		_build_log.queue_free()
		_build_log = null


# ---------------------------------------------------------------------------
# EditorPlugin overrides
# ---------------------------------------------------------------------------

func _has_main_screen() -> bool:
	return false

func _make_visible(_visible: bool) -> void:
	pass

func _get_plugin_name() -> String:
	return PLUGIN_NAME

func _get_plugin_icon() -> Texture2D:
	return get_editor_interface().get_base_control().get_theme_icon("Node", "EditorIcons")


func _scene_saved(filepath: String) -> void:
	if not filepath.ends_with(".tscn"):
		return
	var scene_root: Node = get_editor_interface().get_edited_scene_root()
	if not scene_root or scene_root.get_class() != "AGSRoom":
		return
	var sync_script := load("res://addons/ag_studio/room_sync.gd")
	if sync_script:
		var err: int = sync_script.write_agroom(scene_root)
		if err != OK:
			push_warning("[AGS] RoomSync: write_agroom returned error %d for %s" % [err, filepath])


func _forward_3d_draw_over_viewport(viewport_control: Control) -> void:
	_draw_billboard_warnings(viewport_control)


func _forward_3d_force_draw_over_viewport(viewport_control: Control) -> void:
	_draw_billboard_warnings(viewport_control)


func _draw_billboard_warnings(vp: Control) -> void:
	var room_root := get_editor_interface().get_edited_scene_root()
	if not room_root or room_root.get_class() != "AGSRoom":
		return

	var camera: Camera3D = vp.get_viewport().get_camera_3d()
	if not is_instance_valid(camera):
		return

	var cam_pos := camera.global_position
	var look_at := Vector3.ZERO
	var cam_node: Node3D = null
	for ch: Node in room_root.get_children():
		if ch.get_class() == "AGSCamera":
			cam_node = ch as Node3D
			if cam_node.has("look_at"):
				look_at = cam_node.get("look_at")
			elif cam_node.has("position"):
				look_at = cam_node.global_position + Vector3(0, 0, 5)

	if look_at == Vector3.ZERO:
		return

	# W1: elevation angle > 30°
	var dx := cam_pos.x - look_at.x
	var dy := cam_pos.y - look_at.y
	var dz := cam_pos.z - look_at.z
	var horizontal := sqrt(dx * dx + dz * dz)
	if horizontal > 0:
		var elev_deg := rad_to_deg(atan2(abs(dy), horizontal))
		if elev_deg > 30.0:
			var screen_pos := camera.unproject_position(cam_pos)
			vp.draw_string(
				get_theme_default_font(),
				screen_pos + Vector2(10, -10),
				"W1: elevation %.0f° (max 30°)" % elev_deg,
				HORIZONTAL_ALIGNMENT_LEFT,
				-1,
				12,
				Color(1.0, 0.6, 0.0, 1.0)
			)

	# W3: horizontal arc > 45° for 4-angle billboard sprites
	# Check if any spawn point has a 2D billboard character
	var has_4angle := false
	for ch: Node in room_root.get_children():
		if ch.get_class() == "AGSSpawnPoint" and ch.has("character"):
			var char_name: String = ch.get("character")
			if not char_name.is_empty():
				var char_path := "res://characters/%s/%s.agchar" % [char_name, char_name]
				if FileAccess.file_exists(char_path):
					var src := FileAccess.get_file_as_string(char_path)
					var rx := RegEx.new()
					rx.compile('sprite_angles\\s*=\\s*(\\d+)')
					var m := rx.search(src)
					if m and m.get_string(1) == "4":
						has_4angle = true
						break

	if has_4angle:
		var arc_deg := rad_to_deg(atan2(abs(dx), abs(dz)))
		if arc_deg > 45.0:
			var screen_pos := camera.unproject_position(cam_pos)
			vp.draw_string(
				get_theme_default_font(),
				screen_pos + Vector2(10, 10),
				"W3: arc %.0f° (max 45° for 4-angle)" % arc_deg,
				HORIZONTAL_ALIGNMENT_LEFT,
				-1,
				12,
				Color(1.0, 0.3, 0.0, 1.0)
			)


# ---------------------------------------------------------------------------
# File activation routing
# ---------------------------------------------------------------------------

func _on_menu_item(id: int) -> void:
	match id:
		0:  # New Project
			if not _wizard or not is_instance_valid(_wizard):
				var wz: ConfirmationDialog = preload("res://addons/ag_studio/project_wizard.gd").new()
				wz.setup(self)
				get_editor_interface().get_base_control().add_child(wz)
				_wizard = wz
			_wizard.popup_centered()
		1:  # Build
			if _build_log:
				make_bottom_panel_item_visible(_build_log)
				(_build_log as Node).call("run_build")
		2:  # Settings
			if not _settings_dialog or not is_instance_valid(_settings_dialog):
				var sd: ConfirmationDialog = preload("res://addons/ag_studio/project_settings_dialog.gd").new()
				sd.setup(self)
				get_editor_interface().get_base_control().add_child(sd)
				_settings_dialog = sd
			(_settings_dialog as Node).call("load_settings")
			_settings_dialog.popup_centered()


func _on_play_pressed() -> void:
	if _build_log:
		make_bottom_panel_item_visible(_build_log)
		_play_btn.disabled = true
		var bl := _build_log as Node
		bl.connect("build_finished", _on_build_finished_play, CONNECT_ONE_SHOT)
		bl.call("run_build")


func _on_build_finished_play(success: bool) -> void:
	_play_btn.disabled = false
	if success:
		get_editor_interface().play_main_scene()


func _on_file_activated(abs_path: String) -> void:
	if abs_path.ends_with(".agchar"):
		if _char_editor:
			make_bottom_panel_item_visible(_char_editor)
			(_char_editor as Node).call("load_char", abs_path)
		if _anim_viewer_3d:
			make_bottom_panel_item_visible(_anim_viewer_3d)
	elif abs_path.ends_with(".agscript"):
		if _script_editor:
			(_script_editor as Node).call("load_script", abs_path)
	elif abs_path.ends_with(".agitem"):
		if _item_editor:
			(_item_editor as Node).call("load_item", abs_path)


func _on_scene_changed(scene_root: Node) -> void:
	if _room_editor and is_instance_valid(_room_editor):
		(_room_editor as Node).call("set_room_scene", scene_root)
