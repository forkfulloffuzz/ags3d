@tool
extends Control

## AG Studio Room Editor — main screen panel (T-E09).
##
## Embeds a SubViewport showing the selected room's .tscn scene.
## Camera orbit: left-drag to orbit, scroll to zoom, right-drag to pan.

const ORBIT_SPEED  := 0.005
const PAN_SPEED    := 0.01
const ZOOM_SPEED   := 0.1
const ZOOM_MIN     := 1.0
const ZOOM_MAX     := 100.0

var _plugin: EditorPlugin

var _viewport: SubViewport
var _camera: Camera3D
var _room_root: Node3D

# Camera state
var _cam_target   := Vector3.ZERO
var _cam_distance := 10.0
var _cam_yaw      := 0.4   # radians
var _cam_pitch    := 0.7   # radians

# Drag state
var _orbiting := false
var _panning  := false

# Status bar
var _status_label: Label


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "RoomEditor"
	_build_ui()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var vbox := VBoxContainer.new()
	vbox.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(vbox)

	vbox.add_child(_build_toolbar())

	var split := HSplitContainer.new()
	split.size_flags_vertical = Control.SIZE_EXPAND_FILL
	split.split_offset = -220
	vbox.add_child(split)

	# ---- Left: viewport stack ----
	var vp_holder := Control.new()
	vp_holder.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	vp_holder.size_flags_vertical = Control.SIZE_EXPAND_FILL
	split.add_child(vp_holder)

	var vp_container := SubViewportContainer.new()
	vp_container.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	vp_container.stretch = true
	# Prevent SubViewportContainer from forwarding input to the game scene.
	vp_container.mouse_filter = Control.MOUSE_FILTER_IGNORE
	vp_holder.add_child(vp_container)

	_viewport = SubViewport.new()
	_viewport.transparent_bg = false
	_viewport.handle_input_locally = false
	vp_container.add_child(_viewport)

	_setup_viewport_scene()

	# Invisible overlay captures all mouse for camera control.
	var overlay := Control.new()
	overlay.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	overlay.mouse_filter = Control.MOUSE_FILTER_STOP
	overlay.gui_input.connect(_on_viewport_input)
	vp_holder.add_child(overlay)

	# ---- Right: property sidebar placeholder ----
	var sidebar := _build_sidebar()
	split.add_child(sidebar)

	# ---- Status bar ----
	_status_label = Label.new()
	_status_label.text = "No room loaded. Double-click a room in the Project panel."
	_status_label.add_theme_constant_override("margin_left", 8)
	vbox.add_child(_status_label)


func _build_toolbar() -> HBoxContainer:
	var tb := HBoxContainer.new()
	tb.add_theme_constant_override("separation", 4)

	var node_types := [
		"+ Floor", "+ Blocker", "+ Point", "+ Camera",
		"+ Spawn", "+ Hotspot", "+ Region",
	]
	for label in node_types:
		var btn := Button.new()
		btn.text = label
		btn.flat = true
		btn.disabled = true
		btn.tooltip_text = "Available in T-E10"
		tb.add_child(btn)

	var sep := VSeparator.new()
	tb.add_child(sep)

	var grid_btn := MenuButton.new()
	grid_btn.text = "Grid 0.5m"
	grid_btn.disabled = true
	tb.add_child(grid_btn)

	return tb


func _build_sidebar() -> VBoxContainer:
	var sidebar := VBoxContainer.new()
	sidebar.custom_minimum_size.x = 220

	var header := Label.new()
	header.text = "Properties"
	header.add_theme_font_size_override("font_size", 12)
	header.add_theme_constant_override("margin_left", 8)
	sidebar.add_child(header)

	sidebar.add_child(HSeparator.new())

	var hint := Label.new()
	hint.text = "Select an AGS node\nto edit its properties.\n\n(Inspector plugin\nimplemented in T-E11)"
	hint.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	hint.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	hint.size_flags_vertical = Control.SIZE_EXPAND_FILL
	hint.autowrap_mode = TextServer.AUTOWRAP_WORD
	sidebar.add_child(hint)

	return sidebar


func _setup_viewport_scene() -> void:
	var env_node := WorldEnvironment.new()
	var env := Environment.new()
	env.background_mode = Environment.BG_COLOR
	env.background_color = Color(0.18, 0.18, 0.18)
	env.ambient_light_source = Environment.AMBIENT_SOURCE_COLOR
	env.ambient_light_color = Color(0.6, 0.6, 0.6)
	env_node.environment = env
	_viewport.add_child(env_node)

	var light := DirectionalLight3D.new()
	light.rotation_degrees = Vector3(-50, 45, 0)
	light.light_energy = 1.2
	_viewport.add_child(light)

	_camera = Camera3D.new()
	_viewport.add_child(_camera)
	_update_camera()


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

## Load a room .tscn (res:// path) into the viewport.
func load_room(res_path: String) -> void:
	if _room_root and is_instance_valid(_room_root):
		_room_root.queue_free()
		_room_root = null

	var packed: PackedScene = load(res_path)
	if not packed:
		push_error("[AGS] RoomEditor: failed to load '%s'" % res_path)
		_status_label.text = "Error: could not load %s" % res_path
		return

	_room_root = packed.instantiate()
	_viewport.add_child(_room_root)

	_status_label.text = res_path

	_focus_on_room()


# ---------------------------------------------------------------------------
# Camera
# ---------------------------------------------------------------------------

func _focus_on_room() -> void:
	# Try to find the initial_camera node for a default view.
	var cam: Node = null
	if _room_root:
		cam = _find_initial_camera(_room_root)

	if cam and cam is Camera3D:
		var c := cam as Camera3D
		_cam_target = c.global_position + c.global_basis.z * -5.0
		_cam_distance = 8.0
		var fwd := -c.global_basis.z
		_cam_yaw = atan2(fwd.x, fwd.z)
		_cam_pitch = clamp(asin(fwd.y), 0.1, PI / 2.0 - 0.05)
	else:
		_cam_target = Vector3.ZERO
		_cam_distance = 12.0
		_cam_yaw = 0.4
		_cam_pitch = 0.7

	_update_camera()


func _find_initial_camera(node: Node) -> Node:
	if node.get("camera_name") != null or node.get_class() == "AGSCamera":
		return node
	if node is Camera3D:
		return node
	for child in node.get_children():
		var result := _find_initial_camera(child)
		if result:
			return result
	return null


func _update_camera() -> void:
	if not is_instance_valid(_camera):
		return
	var offset := Vector3(
		sin(_cam_yaw) * cos(_cam_pitch),
		sin(_cam_pitch),
		cos(_cam_yaw) * cos(_cam_pitch)
	) * _cam_distance
	_camera.position = _cam_target + offset
	_camera.look_at(_cam_target, Vector3.UP)


# ---------------------------------------------------------------------------
# Input
# ---------------------------------------------------------------------------

func _on_viewport_input(event: InputEvent) -> void:
	if event is InputEventMouseButton:
		var mb := event as InputEventMouseButton
		match mb.button_index:
			MOUSE_BUTTON_LEFT:
				_orbiting = mb.pressed
			MOUSE_BUTTON_RIGHT:
				_panning = mb.pressed
			MOUSE_BUTTON_WHEEL_UP:
				_cam_distance = max(ZOOM_MIN, _cam_distance * (1.0 - ZOOM_SPEED))
				_update_camera()
			MOUSE_BUTTON_WHEEL_DOWN:
				_cam_distance = min(ZOOM_MAX, _cam_distance * (1.0 + ZOOM_SPEED))
				_update_camera()

	elif event is InputEventMouseMotion:
		var mm := event as InputEventMouseMotion
		if _orbiting:
			_cam_yaw -= mm.relative.x * ORBIT_SPEED
			_cam_pitch = clamp(_cam_pitch - mm.relative.y * ORBIT_SPEED, 0.05, PI * 0.45)
			_update_camera()
		elif _panning:
			# Pan in camera-local XY plane.
			var right := Vector3(cos(_cam_yaw), 0.0, -sin(_cam_yaw))
			var up    := Vector3(0.0, 1.0, 0.0)
			var speed := _cam_distance * PAN_SPEED
			_cam_target -= right * mm.relative.x * speed
			_cam_target += up    * mm.relative.y * speed
			_update_camera()
