@tool
extends Control

## AG Studio Character Editor — main screen panel (T-E13)
##
## Opens when the author double-clicks a .agchar file in the Project panel.
## Reads the .agchar, presents editable fields, and writes changes back on save.

var _plugin: EditorPlugin

var _abs_path: String     # absolute path to the .agchar file
var _internal_name: String

var _display_name_edit: LineEdit
var _mesh_edit: LineEdit
var _status_label: Label
var _save_btn: Button


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "CharEditor"
	_build_ui()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var vbox := VBoxContainer.new()
	vbox.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	vbox.add_theme_constant_override("separation", 8)
	add_child(vbox)

	# ---- Header ----
	var header := Label.new()
	header.text = "Character"
	header.add_theme_font_size_override("font_size", 16)
	vbox.add_child(header)

	vbox.add_child(HSeparator.new())

	# ---- Form (centred, max 480px wide) ----
	var centre := CenterContainer.new()
	centre.size_flags_vertical = Control.SIZE_EXPAND_FILL
	vbox.add_child(centre)

	var form := VBoxContainer.new()
	form.custom_minimum_size.x = 480
	form.add_theme_constant_override("separation", 10)
	centre.add_child(form)

	# Internal name (read-only)
	form.add_child(_field_label("Internal name"))
	var name_lbl := Label.new()
	name_lbl.name = "InternalNameLabel"
	name_lbl.add_theme_color_override("font_color", Color(0.6, 0.6, 0.6))
	form.add_child(name_lbl)

	form.add_child(HSeparator.new())

	# Display name
	form.add_child(_field_label("Display name"))
	_display_name_edit = LineEdit.new()
	_display_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_display_name_edit.placeholder_text = "Player"
	form.add_child(_display_name_edit)

	# Mesh file
	form.add_child(_field_label("Mesh file  (.glb / .obj — leave blank for default capsule)"))
	var mesh_row := HBoxContainer.new()
	mesh_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	form.add_child(mesh_row)

	_mesh_edit = LineEdit.new()
	_mesh_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_mesh_edit.placeholder_text = "characters/player/player.glb"
	mesh_row.add_child(_mesh_edit)

	var browse_btn := Button.new()
	browse_btn.text = "…"
	browse_btn.pressed.connect(_browse_mesh)
	mesh_row.add_child(browse_btn)

	form.add_child(HSeparator.new())

	# Save button
	_save_btn = Button.new()
	_save_btn.text = "Save"
	_save_btn.disabled = true
	_save_btn.pressed.connect(_save)
	form.add_child(_save_btn)

	# ---- Status bar ----
	_status_label = Label.new()
	_status_label.text = "No character loaded."
	vbox.add_child(_status_label)

	# Store reference to name label for later
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
	_mesh_edit.text = parsed.get("mesh", "")
	_save_btn.disabled = false


# ---------------------------------------------------------------------------
# Serialisation
# ---------------------------------------------------------------------------

func _save() -> void:
	if _abs_path.is_empty():
		return

	var lines: Array[String] = []
	lines.append('Character "%s" {' % _internal_name)
	var dn: String = _display_name_edit.text.strip_edges()
	if not dn.is_empty():
		lines.append('    display_name = "%s"' % dn)
	var mesh: String = _mesh_edit.text.strip_edges()
	if not mesh.is_empty():
		lines.append('    mesh = "%s"' % mesh)
	lines.append("}")
	lines.append("")

	var fa := FileAccess.open(_abs_path, FileAccess.WRITE)
	if not fa:
		_status_label.text = "Error: could not write " + _abs_path
		return
	fa.store_string("\n".join(lines))
	fa.close()

	_status_label.text = "Saved. Running ag build…"

	# Trigger a rebuild so the .tscn is regenerated.
	if _plugin:
		var bl = _plugin.get_node_or_null("BuildLog")
		if bl and bl.has_method("run_build"):
			bl.call("run_build")
	_status_label.text = "Saved — " + _abs_path


# ---------------------------------------------------------------------------
# Parser (minimal — just extracts key = "value" pairs)
# ---------------------------------------------------------------------------

func _parse_agchar(src: String) -> Dictionary:
	var result: Dictionary = {}
	# Extract internal name from: Character "name" {
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
# File browser
# ---------------------------------------------------------------------------

func _browse_mesh() -> void:
	var dialog := EditorFileDialog.new()
	dialog.file_mode = EditorFileDialog.FILE_MODE_OPEN_FILE
	dialog.filters = PackedStringArray(["*.glb ; GL Transmission Format", "*.obj ; Wavefront OBJ"])
	dialog.file_selected.connect(func(path: String) -> void:
		# Store as project-relative path
		_mesh_edit.text = ProjectSettings.localize_path(path)
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
