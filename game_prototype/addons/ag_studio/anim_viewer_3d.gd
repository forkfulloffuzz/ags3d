@tool
extends Control

## 3D Animation Viewer — bottom panel showing character mesh + animations.
## Opens automatically when a .agchar file is selected.

signal clip_selected(clip_name: String)

var _plugin: EditorPlugin
var _preview: SubViewport
var _char_root: Node3D
var _anim_player: AnimationPlayer
var _clip_selector: OptionButton
var _play_btn: Button
var _stop_btn: Button
var _loop_toggle: CheckButton
var _frame_slider: HSlider
var _frame_label: Label
var _fps_sb: SpinBox
var _current_clip: String = ""
var _loop: bool = true
var _fps: float = 12.0
var _frame_timer: Timer


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "AnimViewer3D"
	_build_ui()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var vbox := VBoxContainer.new()
	vbox.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(vbox)

	# Header row
	var header := HBoxContainer.new()
	vbox.add_child(header)
	var lbl := Label.new()
	lbl.text = "3D Animation Preview"
	lbl.add_theme_font_size_override("font_size", 12)
	header.add_child(lbl)
	header.add_child(Strut.new(), true)

	var sep := HSeparator.new()
	vbox.add_child(sep)

	# SubViewport
	_preview = SubViewport.new()
	_preview.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_preview.size_flags_vertical = Control.SIZE_EXPAND_FILL
	var vp_container := SubViewportContainer.new()
	vp_container.stretch = true
	vp_container.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	vp_container.size_flags_vertical = Control.SIZE_EXPAND_FILL
	vp_container.add_child(_preview)
	vbox.add_child(vp_container)

	var env := WorldEnvironment.new()
	var env_data := Environment.new()
	env_data.background_mode = Environment.BG_COLOR
	env_data.background_color = Color(0.12, 0.12, 0.12)
	env.environment = env_data
	_preview.add_child(env)

	var light := DirectionalLight3D.new()
	light.rotation_degrees = Vector3(-45, 30, 0)
	light.light_energy = 1.5
	_preview.add_child(light)

	_anim_player = AnimationPlayer.new()
	_preview.add_child(_anim_player)

	# Controls
	var controls := HBoxContainer.new()
	vbox.add_child(controls)

	_clip_selector = OptionButton.new()
	_clip_selector.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_clip_selector.item_selected.connect(_on_clip_selected)
	controls.add_child(_clip_selector)

	_play_btn = Button.new()
	_play_btn.text = "▶"
	_play_btn.pressed.connect(_on_play)
	controls.add_child(_play_btn)

	_stop_btn = Button.new()
	_stop_btn.text = "■"
	_stop_btn.pressed.connect(_on_stop)
	controls.add_child(_stop_btn)

	_loop_toggle = CheckButton.new()
	_loop_toggle.button_pressed = true
	_loop_toggle.toggled.connect(func(v: bool) -> void: _loop = v)
	controls.add_child(_loop_toggle)
	var loop_lbl := Label.new()
	loop_lbl.text = "Loop"
	controls.add_child(loop_lbl)

	var fps_lbl := Label.new()
	fps_lbl.text = "FPS:"
	controls.add_child(fps_lbl)
	_fps_sb = SpinBox.new()
	_fps_sb.min_value = 1
	_fps_sb.max_value = 60
	_fps_sb.value = 12
	_fps_sb.step = 1
	_fps_sb.value_changed.connect(func(v: float) -> void: _fps = v)
	controls.add_child(_fps_sb)

	var sep2 := VSeparator.new()
	vbox.add_child(sep2)

	# Frame scrubber
	var scrubber := HBoxContainer.new()
	vbox.add_child(scrubber)
	_frame_slider = HSlider.new()
	_frame_slider.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_frame_slider.steps = 1
	_frame_slider.value_changed.connect(_on_scrub)
	scrubber.add_child(_frame_slider)
	_frame_label = Label.new()
	_frame_label.text = "0 / 0"
	_frame_label.custom_minimum_size.x = 60
	scrubber.add_child(_frame_label)

	_frame_timer = Timer.new()
	_frame_timer.timeout.connect(_advance_frame)
	add_child(_frame_timer)


func load_character(agchar_path: String) -> void:
	if _char_root and is_instance_valid(_char_root):
		_preview.remove_child(_char_root)
		_char_root.queue_free()
		_char_root = null

	var parsed := _parse_agchar(FileAccess.get_file_as_string(agchar_path))
	var mesh_path: String = parsed.get("mesh", "")
	if mesh_path.is_empty():
		_char_root = Node3D.new()
		_preview.add_child(_char_root)
		_update_clip_list()
		return

	var full_path := ProjectSettings.globalize_path(mesh_path)
	if not ResourceLoader.exists(mesh_path):
		_char_root = Node3D.new()
		_preview.add_child(_char_root)
		_update_clip_list()
		return

	var gltf := GLTFDocument.new()
	var state := GLTFState.new()
	var result := gltf.append_from_path(full_path, state)
	if result != OK:
		_char_root = Node3D.new()
		_preview.add_child(_char_root)
		_update_clip_list()
		return

	_char_root = gltf.generate_scene(state)
	_preview.add_child(_char_root)

	var ap: AnimationPlayer = _char_root.find_child("*", true, false)
	while ap != null and not (ap is AnimationPlayer):
		ap = ap.get_parent().find_child("*", true, false) if ap.get_parent() else null
	if ap is AnimationPlayer:
		var new_ap: AnimationPlayer = _anim_player
		# Copy animations from scene's AnimationPlayer to our viewer one
		for anim_name in (ap as AnimationPlayer).get_animation_list():
			var anim: Animation = (ap as AnimationPlayer).get_animation(anim_name).duplicate()
			new_ap.add_animation(anim_name, anim)
		_anim_player = new_ap

	_update_clip_list()


func _parse_agchar(src: String) -> Dictionary:
	var result: Dictionary = {}
	var rx_name := RegEx.new()
	rx_name.compile('Character\\s+"([^"]+)"')
	var m := rx_name.search(src)
	if m:
		result["internal_name"] = m.get_string(1)
	var rx_kv := RegEx.new()
	rx_kv.compile('(\\w+)\\s*=\\s*"([^"]*)"')
	for match in rx_kv.search_all(src):
		result[match.get_string(1)] = match.get_string(2)
	return result


func _update_clip_list() -> void:
	_clip_selector.clear()
	if not is_instance_valid(_anim_player):
		return
	for clip: String in _anim_player.get_animation_list():
		_clip_selector.add_item(clip)


func _on_clip_selected(index: int) -> void:
	_current_clip = _clip_selector.get_item_text(index)
	_play_btn.disabled = false


func _on_play() -> void:
	if _current_clip.is_empty() or not is_instance_valid(_anim_player):
		return
	var anim: Animation = _anim_player.get_animation(_current_clip)
	if not anim:
		return
	_anim_player.play(_current_clip)
	var fps := _fps if _fps > 0 else 12.0
	_frame_timer.start(1.0 / fps)


func _on_stop() -> void:
	_anim_player.stop()
	_frame_timer.stop()


func _advance_frame() -> void:
	if not is_instance_valid(_anim_player) or _current_clip.is_empty():
		_frame_timer.stop()
		return
	var anim: Animation = _anim_player.get_animation(_current_clip)
	if not anim:
		return
	var pos := _anim_player.current_animation_position
	var length := anim.length
	var step := 1.0 / _fps
	pos += step
	if pos >= length:
		if _loop:
			pos = pos - length * floor(pos / length)
		else:
			pos = length
			_frame_timer.stop()
	_anim_player.seek(pos, true)
	_update_scrubber(pos, length)


func _on_scrub(value: float) -> void:
	if not is_instance_valid(_anim_player) or _current_clip.is_empty():
		return
	var anim: Animation = _anim_player.get_animation(_current_clip)
	if not anim:
		return
	_anim_player.seek(value, true)
	_update_scrubber(value, anim.length)


func _update_scrubber(pos: float, length: float) -> void:
	_frame_slider.max_value = length
	_frame_slider.value = pos
	var total_frames := int(length * _fps)
	var current_frame := int(pos * _fps)
	_frame_label.text = "%d / %d" % [current_frame, total_frames]
