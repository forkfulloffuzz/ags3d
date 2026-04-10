@tool
extends Control

## AG Studio Character Editor — full-featured panel for .agchar files.
##
## Shows when the author double-clicks a .agchar file in the Project panel.
## Fields: type selector (3D/2D/puppet), mesh/sprite, speech, animation clips.

var _plugin: EditorPlugin

var _abs_path: String = ""
var _internal_name: String = ""
var _anim_clips: Array[String] = []
var _anim_viewer: Control

var _display_name_edit: LineEdit
var _type_selector: OptionButton
var _mesh_edit: LineEdit
var _speech_colour_btn: ColorPickerButton
var _speech_font_edit: LineEdit
var _sprite_sheet_edit: LineEdit
var _sprite_angles_sb: SpinBox
var _frames_per_angle_sb: SpinBox
var _anim_list: Tree
var _status_label: Label
var _save_btn: Button
var _anim_section: VBoxContainer


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "CharEditor"
	_build_ui()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var scroll := ScrollContainer.new()
	scroll.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	add_child(scroll)

	var vbox := VBoxContainer.new()
	vbox.custom_minimum_size.x = 520
	vbox.add_theme_constant_override("separation", 8)
	scroll.add_child(vbox)

	var header := Label.new()
	header.text = "Character"
	header.add_theme_font_size_override("font_size", 16)
	vbox.add_child(header)

	vbox.add_child(HSeparator.new())

	# Internal name (read-only)
	vbox.add_child(_field_label("Internal name"))
	var name_lbl := Label.new()
	name_lbl.name = "InternalNameLabel"
	name_lbl.add_theme_color_override("font_color", Color(0.6, 0.6, 0.6))
	vbox.add_child(name_lbl)

	vbox.add_child(HSeparator.new())

	# Type selector
	vbox.add_child(_field_label("Character type"))
	_type_selector = OptionButton.new()
	_type_selector.add_item("3D Mesh", 0)
	_type_selector.add_item("2D Billboard", 1)
	_type_selector.add_item("Puppet", 2)
	_type_selector.item_selected.connect(_on_type_changed)
	vbox.add_child(_type_selector)

	# 3D fields
	var mesh_section := VBoxContainer.new()
	mesh_section.name = "MeshSection"
	vbox.add_child(mesh_section)
	mesh_section.add_child(_field_label("Mesh file  (.glb — leave blank for default capsule)"))
	var mesh_row := HBoxContainer.new()
	mesh_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	mesh_section.add_child(mesh_row)
	_mesh_edit = LineEdit.new()
	_mesh_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	mesh_row.add_child(_mesh_edit)
	var browse_btn := Button.new()
	browse_btn.text = "…"
	browse_btn.pressed.connect(_browse_mesh)
	mesh_row.add_child(browse_btn)

	vbox.add_child(HSeparator.new())

	# 2D billboard fields
	var billboard_section := VBoxContainer.new()
	billboard_section.name = "BillboardSection"
	billboard_section.visible = false
	vbox.add_child(billboard_section)

	billboard_section.add_child(_field_label("Sprite sheet  (.png)"))
	var sheet_row := HBoxContainer.new()
	sheet_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	billboard_section.add_child(sheet_row)
	_sprite_sheet_edit = LineEdit.new()
	_sprite_sheet_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sheet_row.add_child(_sprite_sheet_edit)
	var sheet_browse := Button.new()
	sheet_browse.text = "…"
	sheet_browse.pressed.connect(_browse_sprite_sheet)
	sheet_row.add_child(sheet_browse)

	var angles_row := HBoxContainer.new()
	billboard_section.add_child(angles_row)
	angles_row.add_child(_field_label("Sprite angles:"))
	_sprite_angles_sb = SpinBox.new()
	_sprite_angles_sb.min_value = 2
	_sprite_angles_sb.max_value = 16
	_sprite_angles_sb.value = 8
	angles_row.add_child(_sprite_angles_sb)
	angles_row.add_child(_field_label("Frames per angle:"))
	_frames_per_angle_sb = SpinBox.new()
	_frames_per_angle_sb.min_value = 1
	_frames_per_angle_sb.max_value = 32
	_frames_per_angle_sb.value = 4
	angles_row.add_child(_frames_per_angle_sb)

	vbox.add_child(HSeparator.new())

	# Speech section
	var speech_header := Button.new()
	speech_header.text = "Speech  (click to expand)"
	speech_header.alignment = Alignment.ALIGN_LEFT
	speech_header.pressed.connect(func() -> void:
		var sec: VBoxContainer = speech_header.get_parent().get_node("SpeechSection")
		sec.visible = not sec.visible
		speech_header.text = "Speech  " + ("(click to collapse)" if sec.visible else "(click to expand)")
	)
	vbox.add_child(speech_header)

	var speech_section := VBoxContainer.new()
	speech_section.name = "SpeechSection"
	speech_section.visible = false
	vbox.add_child(speech_section)

	speech_section.add_child(_field_label("Speech colour"))
	var colour_row := HBoxContainer.new()
	speech_section.add_child(colour_row)
	_speech_colour_btn = ColorPickerButton.new()
	_speech_colour_btn.color = Color.WHITE
	_speech_colour_btn.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	colour_row.add_child(_speech_colour_btn)

	speech_section.add_child(_field_label("Speech font  (.ttf — optional)"))
	_speech_font_edit = LineEdit.new()
	_speech_font_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	speech_section.add_child(_speech_font_edit)

	vbox.add_child(HSeparator.new())

	# Display name
	vbox.add_child(_field_label("Display name"))
	_display_name_edit = LineEdit.new()
	_display_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_display_name_edit.placeholder_text = "Player"
	vbox.add_child(_display_name_edit)

	vbox.add_child(HSeparator.new())

	# Animation clips section
	var anim_header := Button.new()
	anim_header.text = "Animations  (click to expand)"
	anim_header.alignment = Alignment.ALIGN_LEFT
	anim_header.pressed.connect(func() -> void:
		var sec: VBoxContainer = anim_header.get_parent().get_node("AnimSection")
		sec.visible = not sec.visible
		anim_header.text = "Animations  " + ("(click to collapse)" if sec.visible else "(click to expand)")
	)
	vbox.add_child(anim_header)

	_anim_section = VBoxContainer.new()
	_anim_section.name = "AnimSection"
	_anim_section.visible = false
	vbox.add_child(_anim_section)

	var anim_hint := Label.new()
	anim_hint.text = "Animations are read from the .glb file. Select a clip and click Preview to play it."
	anim_hint.autowrap_mode = TextServer.AUTOWRAP_WORD
	_anim_section.add_child(anim_hint)

	_anim_list = Tree.new()
	_anim_list.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_anim_list.custom_minimum_size.y = 120
	_anim_section.add_child(_anim_list)

	vbox.add_child(HSeparator.new())

	# Save button
	_save_btn = Button.new()
	_save_btn.text = "Save"
	_save_btn.disabled = true
	_save_btn.pressed.connect(_save)
	vbox.add_child(_save_btn)

	# Status bar
	var status_row := HBoxContainer.new()
	status_row.add_child(Strut.new(), true)
	_status_label = Label.new()
	_status_label.text = "No character loaded."
	status_row.add_child(_status_label)
	vbox.add_child(status_row)

	set_meta("_name_lbl", name_lbl)


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

func load_char(abs_path: String) -> void:
	_abs_path = abs_path
	_status_label.text = abs_path

	var src := FileAccess.get_file_as_string(abs_path)
	if src.is_empty():
		_status_label.text = "Error: could not read " + abs_path
		return

	var parsed := _parse_agchar(src)
	_internal_name = parsed.get("internal_name", abs_path.get_file().get_basename())

	var name_lbl: Label = get_meta("_name_lbl")
	name_lbl.text = _internal_name

	_display_name_edit.text = parsed.get("display_name", "")

	var char_type: String = parsed.get("type", "3d")
	match char_type:
		"2d":  _type_selector.selected = 1
		"puppet": _type_selector.selected = 2
		_:     _type_selector.selected = 0

	_mesh_edit.text = parsed.get("mesh", "")
	_sprite_sheet_edit.text = parsed.get("sprite_sheet", "")
	_sprite_angles_sb.value = float(parsed.get("sprite_angles", "8"))
	_frames_per_angle_sb.value = float(parsed.get("frames_per_angle", "4"))

	var speech_col_str: String = parsed.get("speech_colour", "")
	if not speech_col_str.is_empty():
		_speech_colour_btn.color = Color(speech_col_str)
	_speech_font_edit.text = parsed.get("speech_font", "")

	_update_type_visibility()
	_save_btn.disabled = false
	_status_label.text = abs_path
	_refresh_animations_list()


func set_anim_viewer(viewer: Control) -> void:
	_anim_viewer = viewer


# ---------------------------------------------------------------------------
# Serialisation
# ---------------------------------------------------------------------------

func _save() -> void:
	if _abs_path.is_empty():
		return

	var char_type := "3d"
	match _type_selector.selected:
		1: char_type = "2d"
		2: char_type = "puppet"

	var lines: Array[String] = []
	lines.append('Character "%s" {' % _internal_name)
	var dn: String = _display_name_edit.text.strip_edges()
	if not dn.is_empty():
		lines.append('    display_name = "%s"' % dn)

	lines.append('    type = "%s"' % char_type)

	var mesh: String = _mesh_edit.text.strip_edges()
	if not mesh.is_empty():
		lines.append('    mesh = "%s"' % mesh)

	if char_type == "2d":
		var sheet: String = _sprite_sheet_edit.text.strip_edges()
		if not sheet.is_empty():
			lines.append('    sprite_sheet = "%s"' % sheet)
		lines.append('    sprite_angles = %d' % int(_sprite_angles_sb.value))
		lines.append('    frames_per_angle = %d' % int(_frames_per_angle_sb.value))

	var speech_col: String = _speech_colour_btn.color.to_html()
	if not speech_col.is_empty():
		lines.append('    speech_colour = "#%s"' % speech_col)

	var speech_font: String = _speech_font_edit.text.strip_edges()
	if not speech_font.is_empty():
		lines.append('    speech_font = "%s"' % speech_font)

	lines.append("}")
	lines.append("")

	var fa := FileAccess.open(_abs_path, FileAccess.WRITE)
	if not fa:
		_status_label.text = "Error: could not write " + _abs_path
		return
	fa.store_string("\n".join(lines))
	fa.close()

	_status_label.text = "Saved — " + _abs_path

	if _plugin:
		var bl := _plugin.get_node_or_null("BuildLog")
		if bl and bl.has_method("run_build"):
			bl.call("run_build")


# ---------------------------------------------------------------------------
# Parser
# ---------------------------------------------------------------------------

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


# ---------------------------------------------------------------------------
# Type selector
# ---------------------------------------------------------------------------

func _on_type_changed(index: int) -> void:
	_update_type_visibility()


func _update_type_visibility() -> void:
	var mesh_sec := get_node_or_null("MeshSection")
	var billboard_sec := get_node_or_null("BillboardSection")
	if mesh_sec and billboard_sec:
		var is_3d_or_puppet := _type_selector.selected != 1
		mesh_sec.visible = is_3d_or_puppet
		billboard_sec.visible = not is_3d_or_puppet


func _refresh_animations_list() -> void:
	for c: Node in _anim_list.get_children():
		c.queue_free()
	_anim_clips.clear()

	var mesh_path: String = _mesh_edit.text.strip_edges()
	if mesh_path.is_empty() or not ResourceLoader.exists(mesh_path):
		return

	var full_path := ProjectSettings.globalize_path(mesh_path)
	var gltf := GLTFDocument.new()
	var state := GLTFState.new()
	if gltf.append_from_path(full_path, state) != OK:
		return

	var clip_names: Array[String] = []
	for i in state.get_animation_count():
		clip_names.append(state.get_animation_name(i))
	clip_names.sort()

	for clip: String in clip_names:
		var row := HBoxContainer.new()
		row.size_flags_horizontal = Control.SIZE_EXPAND_FILL

		var lbl := Label.new()
		lbl.text = clip
		lbl.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		row.add_child(lbl)

		var preview_btn := Button.new()
		preview_btn.text = "Preview"
		preview_btn.pressed.connect(func() -> void:
			if _anim_viewer and _anim_viewer.has_method("load_character"):
				_anim_viewer.call("load_character", _abs_path)
		)
		row.add_child(preview_btn)

		_anim_list.add_child(row)
		_anim_clips.append(clip)


# ---------------------------------------------------------------------------
# File browsers
# ---------------------------------------------------------------------------

func _browse_mesh() -> void:
	var dialog := EditorFileDialog.new()
	dialog.file_mode = EditorFileDialog.FILE_MODE_OPEN_FILE
	dialog.filters = PackedStringArray(["*.glb ; GL Transmission Format", "*.obj ; Wavefront OBJ"])
	dialog.file_selected.connect(func(path: String) -> void:
		_mesh_edit.text = ProjectSettings.localize_path(path)
		_refresh_animations_list()
		dialog.queue_free()
	)
	add_child(dialog)
	dialog.popup_centered_ratio(0.6)


func _browse_sprite_sheet() -> void:
	var dialog := EditorFileDialog.new()
	dialog.file_mode = EditorFileDialog.FILE_MODE_OPEN_FILE
	dialog.filters = PackedStringArray(["*.png ; PNG Image"])
	dialog.file_selected.connect(func(path: String) -> void:
		_sprite_sheet_edit.text = ProjectSettings.localize_path(path)
		dialog.queue_free()
	)
	add_child(dialog)
	dialog.popup_centered_ratio(0.6)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

func _field_label(text: String) -> Label:
	var lbl := Label.new()
	lbl.text = text
	lbl.add_theme_font_size_override("font_size", 11)
	return lbl
