@tool
extends Control

## AG Studio GUI Layout Editor — canvas for .agui files.
##
## Shows when the author double-clicks a .agui file in the Project panel.
## Features:
##   - Screen-sized canvas with anchor position indicators
##   - Palette of widget types to add (InventoryBar, VerbBar, StatusLine)
##   - Click canvas to place widget at that anchor position
##   - Widget property form: anchor, offset, type-specific fields
##   - Save writes .agui file and triggers ag build

signal file_saved(path: String)

var _plugin: EditorPlugin

var _abs_path: String = ""
var _gui_name: String = ""
var _layer: int = 1
var _widgets: Array[Dictionary] = []

var _canvas: Control
var _canvas_painter: Control
var _widget_list: ItemList
var _form: VBoxContainer
var _layer_sb: SpinBox
var _anchor_selector: OptionButton
var _offset_x_sb: SpinBox
var _offset_y_sb: SpinBox
var _columns_sb: SpinBox
var _verbs_edit: LineEdit
var _font_edit: LineEdit
var _save_btn: Button
var _status_label: Label
var _current_widget_index: int = -1
var _drag_widget_type: String = ""
var _drag_preview: Control = null

const _ANCHOR_LABELS := ["top_left", "top_right", "bottom_left", "bottom_right", "top", "bottom", "center"]
const _ANCHOR_COLORS := {
	"top_left":     Color(0.2, 0.8, 0.2, 0.7),
	"top_right":    Color(0.8, 0.8, 0.2, 0.7),
	"bottom_left":  Color(0.2, 0.4, 0.9, 0.7),
	"bottom_right": Color(0.9, 0.3, 0.3, 0.7),
	"top":          Color(0.6, 0.3, 0.9, 0.7),
	"bottom":       Color(0.3, 0.7, 0.7, 0.7),
	"center":       Color(0.7, 0.7, 0.7, 0.7),
}
const _WIDGET_COLORS := {
	"InventoryBar": Color(0.2, 0.6, 0.9, 0.8),
	"VerbBar":      Color(0.9, 0.6, 0.1, 0.8),
	"StatusLine":   Color(0.5, 0.9, 0.4, 0.8),
}

func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "GUIEditor"
	_build_ui()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var hbox := HSplitContainer.new()
	hbox.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.size_flags_vertical = Control.SIZE_EXPAND_FILL
	hbox.split_offset = 200
	add_child(hbox)

	var left := VBoxContainer.new()
	left.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.add_child(left)

	var header := Label.new()
	header.text = "GUI Layout"
	header.add_theme_font_size_override("font_size", 14)
	left.add_child(header)

	var layer_row := HBoxContainer.new()
	left.add_child(layer_row)
	layer_row.add_child(_field_label("Layer:"))
	_layer_sb = SpinBox.new()
	_layer_sb.min_value = 1
	_layer_sb.max_value = 100
	_layer_sb.value = 1
	_layer_sb.value_changed.connect(func(v: float) -> void: _layer = int(v))
	layer_row.add_child(_layer_sb)

	left.add_child(HSeparator.new())

	var palette := Label.new()
	palette.text = "Add Widget:"
	palette.add_theme_font_size_override("font_size", 11)
	left.add_child(palette)

	var btn_inv := Button.new()
	btn_inv.text = "+ InventoryBar"
	btn_inv.pressed.connect(func() -> void: _add_widget("InventoryBar"))
	left.add_child(btn_inv)

	var btn_verb := Button.new()
	btn_verb.text = "+ VerbBar"
	btn_verb.pressed.connect(func() -> void: _add_widget("VerbBar"))
	left.add_child(btn_verb)

	var btn_status := Button.new()
	btn_status.text = "+ StatusLine"
	btn_status.pressed.connect(func() -> void: _add_widget("StatusLine"))
	left.add_child(btn_status)

	left.add_child(HSeparator.new())

	var list_lbl := Label.new()
	list_lbl.text = "Widgets"
	list_lbl.add_theme_font_size_override("font_size", 11)
	left.add_child(list_lbl)

	_widget_list = ItemList.new()
	_widget_list.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_widget_list.item_selected.connect(_on_widget_selected)
	left.add_child(_widget_list)

	var right := VBoxContainer.new()
	right.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.add_child(right)

	var canvas_frame := Label.new()
	canvas_frame.text = "Canvas  (right-click widget in list to remove)"
	canvas_frame.add_theme_font_size_override("font_size", 11)
	right.add_child(canvas_frame)

	_canvas = Control.new()
	_canvas.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_canvas.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_canvas.custom_minimum_size.y = 300
	_canvas.draw.connect(_on_canvas_draw)
	_canvas.gui_input.connect(_on_canvas_input)
	right.add_child(_canvas)

	var form_lbl := Label.new()
	form_lbl.text = "Widget Properties"
	form_lbl.add_theme_font_size_override("font_size", 11)
	right.add_child(form_lbl)

	_form = VBoxContainer.new()
	_form.size_flags_vertical = Control.SIZE_EXPAND_FILL
	right.add_child(_form)

	var anchor_row := HBoxContainer.new()
	_form.add_child(anchor_row)
	anchor_row.add_child(_field_label("Anchor:"))
	_anchor_selector = OptionButton.new()
	for a: String in _ANCHOR_LABELS:
		_anchor_selector.add_item(a)
	_anchor_selector.item_selected.connect(_on_anchor_changed)
	anchor_row.add_child(_anchor_selector)

	var offset_row := HBoxContainer.new()
	_form.add_child(offset_row)
	offset_row.add_child(_field_label("Offset X:"))
	_offset_x_sb = SpinBox.new()
	_offset_x_sb.min_value = -2000
	_offset_x_sb.max_value = 2000
	_offset_x_sb.value_changed.connect(_on_widget_field_changed)
	offset_row.add_child(_offset_x_sb)
	offset_row.add_child(_field_label("  Y:"))
	_offset_y_sb = SpinBox.new()
	_offset_y_sb.min_value = -2000
	_offset_y_sb.max_value = 2000
	_offset_y_sb.value_changed.connect(_on_widget_field_changed)
	offset_row.add_child(_offset_y_sb)

	_columns_sb = SpinBox.new()
	_columns_sb.min_value = 1
	_columns_sb.max_value = 20
	_columns_sb.value = 8
	_columns_sb.value_changed.connect(_on_widget_field_changed)
	_verbs_edit = LineEdit.new()
	_verbs_edit.placeholder_text = 'verbs: "Look,Use,Pick up"'
	_verbs_edit.text_changed.connect(_on_widget_field_changed)
	_font_edit = LineEdit.new()
	_font_edit.placeholder_text = "font path (optional)"
	_font_edit.text_changed.connect(_on_widget_field_changed)

	right.add_child(HSeparator.new())

	var save_row := HBoxContainer.new()
	right.add_child(save_row)
	_save_btn = Button.new()
	_save_btn.text = "Save"
	_save_btn.disabled = true
	_save_btn.pressed.connect(_save)
	save_row.add_child(_save_btn)
	save_row.add_child(Strut.new(), true)
	_status_label = Label.new()
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD
	save_row.add_child(_status_label)


func load_gui(abs_path: String) -> void:
	_abs_path = abs_path
	_status_label.text = abs_path
	_current_widget_index = -1

	var src := FileAccess.get_file_as_string(abs_path)
	if src.is_empty():
		_status_label.text = "Error: could not read " + abs_path
		return

	var parsed := _parse_agui(src)
	_gui_name = parsed.get("name", abs_path.get_file().get_basename())
	_layer = parsed.get("layer", 1)
	_widgets = parsed.get("widgets", [])

	_layer_sb.value = _layer
	_populate_widget_list()
	_canvas.queue_sort()
	_save_btn.disabled = false
	_status_label.text = abs_path


func _populate_widget_list() -> void:
	_widget_list.clear()
	for w: Dictionary in _widgets:
		_widget_list.add_item("%s: %s" % [w.get("type", "?"), w.get("name", "?")])


func _on_widget_selected(index: int) -> void:
	_current_widget_index = index
	if index < 0 or index >= _widgets.size():
		_clear_form()
		return
	var w: Dictionary = _widgets[index]
	_show_widget_form(w)
	_canvas.queue_sort()


func _show_widget_form(w: Dictionary) -> void:
	for c: Node in _form.get_children():
		c.queue_free()

	var type_lbl := Label.new()
	type_lbl.text = "Type: %s" % w.get("type", "?")
	_form.add_child(type_lbl)

	var name_row := HBoxContainer.new()
	_form.add_child(name_row)
	name_row.add_child(_field_label("Name:"))
	var name_edit := LineEdit.new()
	name_edit.text = w.get("name", "")
	name_edit.text_changed.connect(func(t: String) -> void:
		w["name"] = t
		_populate_widget_list()
		_canvas.queue_sort()
	)
	name_row.add_child(name_edit)

	var anchor: String = w.get("anchor", "top_left")
	var anchor_idx := _ANCHOR_LABELS.find(anchor)
	if anchor_idx < 0:
		anchor_idx = 0
	_anchor_selector.selected = anchor_idx

	_offset_x_sb.value = float(w.get("offset_x", 0))
	_offset_y_sb.value = float(w.get("offset_y", 0))

	var wtype: String = w.get("type", "")
	match wtype:
		"InventoryBar":
			_form.add_child(_field_label("Columns:"))
			var row := HBoxContainer.new()
			_form.add_child(row)
			_columns_sb.value = float(w.get("columns", 8))
			row.add_child(_columns_sb)
		"VerbBar":
			_form.add_child(_field_label("Verbs (comma-separated):"))
			_verbs_edit.text = ", ".join(w.get("verbs", []))
			_form.add_child(_verbs_edit)
		"StatusLine":
			_form.add_child(_field_label("Font path:"))
			_font_edit.text = w.get("font", "")
			_form.add_child(_font_edit)


func _clear_form() -> void:
	for c: Node in _form.get_children():
		c.queue_free()


func _on_anchor_changed(index: int) -> void:
	if _current_widget_index < 0 or _current_widget_index >= _widgets.size():
		return
	_widgets[_current_widget_index]["anchor"] = _ANCHOR_LABELS[index]
	_canvas.queue_sort()


func _on_widget_field_changed(_value: float) -> void:
	if _current_widget_index < 0 or _current_widget_index >= _widgets.size():
		return
	var w: Dictionary = _widgets[_current_widget_index]
	w["offset_x"] = int(_offset_x_sb.value)
	w["offset_y"] = int(_offset_y_sb.value)

	match w.get("type", ""):
		"InventoryBar":
			w["columns"] = int(_columns_sb.value)
		"VerbBar":
			w["verbs"] = _verbs_edit.text.split(",", false)
			for i: int in range(w["verbs"].size()):
				w["verbs"][i] = w["verbs"][i].strip_edges()
		"StatusLine":
			w["font"] = _font_edit.text.strip_edges()

	_canvas.queue_sort()


func _add_widget(type_: String) -> void:
	var base_name := type_ if type_ != "StatusLine" else "status"
	var idx := 1
	while true:
		var name := "%s_%d" % [base_name.to_lower(), idx]
		var exists := false
		for w: Dictionary in _widgets:
			if w.get("name", "") == name:
				exists = true
				break
		if not exists:
			var w := {
				"type": type_,
				"name": name,
				"anchor": "top_left",
				"offset_x": 0,
				"offset_y": 0,
			}
			if type_ == "InventoryBar":
				w["columns"] = 8
				w["item_size_w"] = 48
				w["item_size_h"] = 48
			elif type_ == "VerbBar":
				w["verbs"] = ["Look", "Use"]
			elif type_ == "StatusLine":
				w["font"] = ""
			_widgets.append(w)
			_populate_widget_list()
			_widget_list.select(_widgets.size() - 1)
			_on_widget_selected(_widgets.size() - 1)
			_canvas.queue_sort()
			return
		idx += 1


func _on_canvas_draw() -> void:
	var cs: Vector2 = _canvas.size
	if cs.x < 2 or cs.y < 2:
		return

	_canvas.draw_rect(Rect2(Vector2.ZERO, cs), Color(0.08, 0.08, 0.1))

	var grid_step := 40.0
	var grid_color := Color(0.2, 0.2, 0.25, 0.5)
	var x := grid_step
	while x < cs.x:
		_canvas.draw_line(Vector2(x, 0), Vector2(x, cs.y), grid_color)
		x += grid_step
	var y := grid_step
	while y < cs.y:
		_canvas.draw_line(Vector2(0, y), Vector2(cs.x, y), grid_color)
		y += grid_step

	_draw_anchor_marker(Vector2(0, 0), "top_left")
	_draw_anchor_marker(Vector2(cs.x, 0), "top_right")
	_draw_anchor_marker(Vector2(0, cs.y), "bottom_left")
	_draw_anchor_marker(Vector2(cs.x, cs.y), "bottom_right")
	_draw_anchor_marker(Vector2(cs.x * 0.5, 0), "top")
	_draw_anchor_marker(Vector2(cs.x * 0.5, cs.y), "bottom")
	_draw_anchor_marker(Vector2(cs.x * 0.5, cs.y * 0.5), "center")

	for i: int in range(_widgets.size()):
		var w: Dictionary = _widgets[i]
		var anchor: String = w.get("anchor", "top_left")
		var pos := _anchor_to_canvas_pos(anchor, cs)
		var col: Color = _WIDGET_COLORS.get(w.get("type", ""), Color.WHITE)
		if i == _current_widget_index:
			col = Color.WHITE
		var of_x: float = w.get("offset_x", 0)
		var of_y: float = w.get("offset_y", 0)
		var box_rect := Rect2(pos + Vector2(of_x, of_y) - Vector2(40, 15), Vector2(80, 30))
		_canvas.draw_rect(box_rect, col, true)
		_canvas.draw_string(
			_canvas.get_theme_default_font(),
			box_rect.position + Vector2(4, 12),
			w.get("name", ""),
			HORIZONTAL_ALIGNMENT_LEFT, -1, 10, Color(1, 1, 1, 0.9)
		)
		if i == _current_widget_index:
			_canvas.draw_rect(box_rect, Color(1, 1, 1, 0.8), false, 2.0)


func _draw_anchor_marker(pos: Vector2, anchor: String) -> void:
	var col: Color = _ANCHOR_COLORS.get(anchor, Color.WHITE)
	_canvas.draw_circle(pos, 5.0, col)
	var lbl := anchor
	_canvas.draw_string(
		_canvas.get_theme_default_font(),
		pos + Vector2(8, -4),
		lbl, HORIZONTAL_ALIGNMENT_LEFT, -1, 9, col
	)


func _anchor_to_canvas_pos(anchor: String, cs: Vector2) -> Vector2:
	match anchor:
		"top_left":     return Vector2(0, 0)
		"top_right":    return Vector2(cs.x, 0)
		"bottom_left":  return Vector2(0, cs.y)
		"bottom_right": return Vector2(cs.x, cs.y)
		"top":          return Vector2(cs.x * 0.5, 0)
		"bottom":       return Vector2(cs.x * 0.5, cs.y)
		"center":       return Vector2(cs.x * 0.5, cs.y * 0.5)
	return Vector2.ZERO


func _on_canvas_input(event: InputEvent) -> void:
	if event is InputEventMouseMotion:
		pass
	elif event is InputEventMouseButton:
		var mb: InputEventMouseButton = event
		if mb.pressed and mb.button_index == MOUSE_BUTTON_LEFT:
			var cs: Vector2 = _canvas.size
			var rel_pos := mb.position
			var nearest_anchor: String = "top_left"
			var nearest_dist: float = 1e9
			for a: String in _ANCHOR_LABELS:
				var ap: Vector2 = _anchor_to_canvas_pos(a, cs)
				var d: float = ap.distance_to(rel_pos)
				if d < nearest_dist:
					nearest_dist = d
					nearest_anchor = a
			if nearest_dist < 40.0 and _current_widget_index >= 0:
				_widgets[_current_widget_index]["anchor"] = nearest_anchor
				var idx: int = _ANCHOR_LABELS.find(nearest_anchor)
				if idx >= 0:
					_anchor_selector.selected = idx
				_canvas.queue_sort()


func _save() -> void:
	if _abs_path.is_empty():
		return

	var lines: Array[String] = []
	lines.append('GUI "%s" {' % _gui_name)
	lines.append("    layer = %d" % _layer)

	for w: Dictionary in _widgets:
		lines.append("")
		lines.append('    %s "%s" {' % [w.get("type", ""), w.get("name", "")])
		lines.append('        position = (%d, %d, %s)' % [
			w.get("offset_x", 0), w.get("offset_y", 0), w.get("anchor", "top_left")])
		match w.get("type", ""):
			"InventoryBar":
				lines.append("        columns = %d" % w.get("columns", 8))
				lines.append("        item_size = (%d, %d)" % [
					w.get("item_size_w", 48), w.get("item_size_h", 48)])
			"VerbBar":
				var verbs: Array = w.get("verbs", [])
				var verb_strs: Array[String] = []
				for v: String in verbs:
					verb_strs.append('"%s"' % v)
				lines.append("        verbs = [%s]" % ", ".join(verb_strs))
			"StatusLine":
				var font: String = w.get("font", "")
				if not font.is_empty():
					lines.append('        font = "%s"' % font)
		lines.append("    }")

	lines.append("}")
	lines.append("")

	var fa := FileAccess.open(_abs_path, FileAccess.WRITE)
	if not fa:
		_status_label.text = "Error: could not write " + _abs_path
		return
	fa.store_string("\n".join(lines))
	fa.close()

	_status_label.text = "Saved — " + _abs_path
	file_saved.emit(_abs_path)

	if _plugin:
		var bl := _plugin.get_node_or_null("BuildLog")
		if bl and bl.has_method("run_build"):
			bl.call("run_build")


func _parse_agui(src: String) -> Dictionary:
	var result := {
		"name": "",
		"layer": 1,
		"widgets": []
	}

	var lines := src.split("\n")
	var section := ""
	for i: int in range(lines.size()):
		var line: String = lines[i].strip_edges()
		if line.is_empty() or line.begins_with("#") or line.begins_with("//"):
			continue

		if line.begins_with('GUI "') or line.begins_with("GUI '"):
			var m := RegEx.new()
			m.compile('GUI\\s+"([^"]+)"')
			var nm := m.search(line)
			if nm:
				result["name"] = nm.get_string(1)
			var layer_m := RegEx.new()
			layer_m.compile("layer\\s*=\\s*(\\d+)")
			var lm := layer_m.search(line)
			if lm:
				result["layer"] = lm.get_string(1).to_int()
			section = "gui"
		elif line.begins_with("InventoryBar \"") or line.begins_with("VerbBar \"") or line.begins_with("StatusLine \""):
			var type_rx := RegEx.new()
			type_rx.compile('^(InventoryBar|VerbBar|StatusLine)\\s+"([^"]+)"')
			var tm := type_rx.search(line)
			if tm:
				var w := {
					"type": tm.get_string(1),
					"name": tm.get_string(2),
					"anchor": "top_left",
					"offset_x": 0,
					"offset_y": 0,
				}
				var block_lines: Array[String] = []
				var j: int = i + 1
				while j < lines.size() and not lines[j].strip_edges().begins_with("}"):
					block_lines.append(lines[j])
					j += 1
				for bl: String in block_lines:
					bl = bl.strip_edges()
					if bl.begins_with("position ="):
						var pos_rx := RegEx.new()
						pos_rx.compile("position\\s*=\\s*\\(([^,]+),\\s*([^,]+),\\s*([^)]+)\\)")
						var pm := pos_rx.search(bl)
						if pm:
							w["offset_x"] = pm.get_string(1).to_int()
							w["offset_y"] = pm.get_string(2).to_int()
							w["anchor"] = pm.get_string(3).strip_edges()
					elif bl.begins_with("columns ="):
						var c_rx := RegEx.new()
						c_rx.compile("columns\\s*=\\s*(\\d+)")
						var cm := c_rx.search(bl)
						if cm:
							w["columns"] = cm.get_string(1).to_int()
					elif bl.begins_with("item_size ="):
						var s_rx := RegEx.new()
						s_rx.compile("item_size\\s*=\\s*\\((\\d+),\\s*(\\d+)\\)")
						var sm := s_rx.search(bl)
						if sm:
							w["item_size_w"] = sm.get_string(1).to_int()
							w["item_size_h"] = sm.get_string(2).to_int()
					elif bl.begins_with("verbs ="):
						var v_rx := RegEx.new()
						v_rx.compile('"([^"]+)"')
						var verbs: Array[String] = []
						for vm in v_rx.search_all(bl):
							verbs.append(vm.get_string(1))
						w["verbs"] = verbs
					elif bl.begins_with("font ="):
						var f_rx := RegEx.new()
						f_rx.compile('font\\s*=\\s*"([^"]*)"')
						var fm := f_rx.search(bl)
						if fm:
							w["font"] = fm.get_string(1)
				result["widgets"].append(w)

	return result


func _field_label(text: String) -> Label:
	var lbl := Label.new()
	lbl.text = text
	lbl.add_theme_font_size_override("font_size", 10)
	return lbl
