@tool
extends Control

## 2D Animation Viewer — bottom panel showing sprite sheet grid and preview.
## Opens automatically when a billboard (type=2d) .agchar is selected.

var _plugin: EditorPlugin
var _sprite: Sprite2D
var _sprite_sheet: Texture2D
var _hframes: int = 1
var _vframes: int = 1
var _frames_per_angle: int = 4
var _current_frame: int = 0
var _current_angle: int = 0
var _anim_timer: Timer
var _fps: float = 8.0

var _angle_spinbox: SpinBox
var _frame_slider: HSlider
var _fps_sb: SpinBox
var _anim_preview: Sprite2D
var _anim_viewport: SubViewport
var _grid_container: GridContainer


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "AnimViewer2D"
	_build_ui()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var hbox := HSplitContainer.new()
	hbox.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.size_flags_vertical = Control.SIZE_EXPAND_FILL
	hbox.split_offset = 250
	add_child(hbox)

	# Left: sprite sheet grid
	var left := VBoxContainer.new()
	left.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	left.size_flags_vertical = Control.SIZE_EXPAND_FILL
	hbox.add_child(left)

	var grid_header := Label.new()
	grid_header.text = "Sprite Sheet"
	grid_header.add_theme_font_size_override("font_size", 12)
	left.add_child(grid_header)

	_grid_container = GridContainer.new()
	_grid_container.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_grid_container.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_grid_container.columns = _frames_per_angle
	left.add_child(_grid_container)

	var controls := HBoxContainer.new()
	left.add_child(controls)

	var angle_lbl := Label.new()
	angle_lbl.text = "Angle:"
	controls.add_child(angle_lbl)
	_angle_spinbox = SpinBox.new()
	_angle_spinbox.min_value = 0
	_angle_spinbox.max_value = 7
	_angle_spinbox.value_changed.connect(func(v: float) -> void: _current_angle = int(v))
	controls.add_child(_angle_spinbox)

	var fps_lbl := Label.new()
	fps_lbl.text = "FPS:"
	controls.add_child(fps_lbl)
	_fps_sb = SpinBox.new()
	_fps_sb.min_value = 1
	_fps_sb.max_value = 30
	_fps_sb.value = 8
	_fps_sb.value_changed.connect(func(v: float) -> void: _fps = v)
	controls.add_child(_fps_sb)

	# Right: animated preview
	var right := VBoxContainer.new()
	right.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	right.size_flags_vertical = Control.SIZE_EXPAND_FILL
	hbox.add_child(right)

	var preview_header := Label.new()
	preview_header.text = "Preview"
	preview_header.add_theme_font_size_override("font_size", 12)
	right.add_child(preview_header)

	var vp_container := SubViewportContainer.new()
	vp_container.stretch = true
	vp_container.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	vp_container.size_flags_vertical = Control.SIZE_EXPAND_FILL
	right.add_child(vp_container)

	_anim_viewport = SubViewport.new()
	_anim_viewport.transparent_bg = false
	vp_container.add_child(_anim_viewport)

	var env := WorldEnvironment.new()
	var env_data := Environment.new()
	env_data.background_mode = Environment.BG_COLOR
	env_data.background_color = Color(0.12, 0.12, 0.12)
	env.environment = env_data
	_anim_viewport.add_child(env)

	_anim_preview = Sprite2D.new()
	_anim_preview.centered = false
	_anim_viewport.add_child(_anim_preview)

	var scrubber := HBoxContainer.new()
	right.add_child(scrubber)
	_frame_slider = HSlider.new()
	_frame_slider.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_frame_slider.steps = 1
	_frame_slider.value_changed.connect(_on_frame_scrub)
	scrubber.add_child(_frame_slider)

	_anim_timer = Timer.new()
	_anim_timer.timeout.connect(_advance_frame)
	add_child(_anim_timer)


func load_character(agchar_path: String) -> void:
	var parsed := _parse_agchar(FileAccess.get_file_as_string(agchar_path))
	var sheet_path: String = parsed.get("sprite_sheet", "")
	_hframes = int(parsed.get("frames_per_angle", "4"))
	_vframes = int(parsed.get("sprite_angles", "8"))
	_frames_per_angle = _hframes

	_grid_container.columns = _hframes

	for ch in _grid_container.get_children():
		ch.queue_free()

	if sheet_path.is_empty() or not ResourceLoader.exists(sheet_path):
		return

	_sprite_sheet = load(sheet_path)
	_build_grid()


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


func _build_grid() -> void:
	if not _sprite_sheet:
		return
	var tex_size := _sprite_sheet.get_size()
	var cell_w := tex_size.x / _hframes
	var cell_h := tex_size.y / _vframes

	for row in range(_vframes):
		for col in range(_hframes):
			var cell := TextureRect.new()
			cell.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
			cell.custom_minimum_size = Vector2(cell_w, cell_h)
			var atlas := AtlasTexture.new()
			atlas.atlas = _sprite_sheet
			atlas.region = Rect2(col * cell_w, row * cell_h, cell_w, cell_h)
			cell.texture = atlas
			cell.gui_input.connect(func(ev: InputEvent) -> void:
				if ev is InputEventMouseButton and (ev as InputEventMouseButton).pressed:
					_current_angle = row
					_current_frame = col
					_angle_spinbox.value = row
					_update_preview()
			)
			_grid_container.add_child(cell)


func _update_preview() -> void:
	if not _sprite_sheet:
		return
	var tex_size := _sprite_sheet.get_size()
	var cell_w := tex_size.x / _hframes
	var cell_h := tex_size.y / _vframes

	var atlas := AtlasTexture.new()
	atlas.atlas = _sprite_sheet
	atlas.region = Rect2(_current_frame * cell_w, _current_angle * cell_h, cell_w, cell_h)
	_anim_preview.texture = atlas
	_anim_preview.frame = _current_angle * _hframes + _current_frame

	_frame_slider.max_value = _hframes * _vframes - 1
	_frame_slider.value = _current_angle * _hframes + _current_frame


func _on_frame_scrub(value: float) -> void:
	var idx := int(value)
	_current_angle = idx / _hframes
	_current_frame = idx % _hframes
	_angle_spinbox.value = _current_angle
	_update_preview()


func _advance_frame() -> void:
	_current_frame += 1
	if _current_frame >= _hframes:
		_current_frame = 0
	_update_preview()
