@tool
extends Control

## AG Studio Script Editor — main screen panel (T-E14)
##
## CodeEdit with AGS-spirit syntax highlighting for .agscript files.
## Ctrl+S saves and triggers ag build. Blocking calls are annotated with
## a clock gutter icon.

var _plugin: EditorPlugin

var _abs_path: String
var _code_edit: CodeEdit
var _status_label: Label
var _tab_bar: TabBar

# Open files: abs_path → { "text": String, "modified": bool }
var _open_files: Dictionary = {}
var _file_order: Array[String] = []


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "ScriptEditor"
	_build_ui()


func _build_ui() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var vbox := VBoxContainer.new()
	vbox.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(vbox)

	# ---- Tab bar ----
	_tab_bar = TabBar.new()
	_tab_bar.tab_close_display_policy = TabBar.CLOSE_BUTTON_SHOW_ACTIVE_ONLY
	_tab_bar.tab_changed.connect(_on_tab_changed)
	_tab_bar.tab_close_pressed.connect(_on_tab_close)
	vbox.add_child(_tab_bar)

	vbox.add_child(HSeparator.new())

	# ---- CodeEdit ----
	_code_edit = CodeEdit.new()
	_code_edit.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_code_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_code_edit.add_theme_color_override("font_color", Color(0.85, 0.85, 0.85))
	_code_edit.gutters_draw_line_numbers = true
	_code_edit.gutters_zero_pad_line_numbers = true
	_code_edit.gutters_draw_fold_gutter = false
	_code_edit.auto_brace_completion_enabled = true
	_code_edit.indent_automatic = true
	_code_edit.indent_size = 4
	_code_edit.syntax_highlighter = _make_highlighter()
	_code_edit.text_changed.connect(_on_text_changed)
	# Ctrl+S shortcut
	var save_action := InputEventKey.new()
	save_action.keycode = KEY_S
	save_action.ctrl_pressed = true
	var shortcut := Shortcut.new()
	shortcut.events = [save_action]
	_code_edit.set_meta("_save_shortcut", shortcut)
	vbox.add_child(_code_edit)

	# ---- Status bar ----
	_status_label = Label.new()
	_status_label.text = "No file open. Double-click a script in the Project panel."
	vbox.add_child(_status_label)

	# Add clock gutter for blocking calls
	_code_edit.add_gutter(0)
	_code_edit.set_gutter_name(0, "blocking")
	_code_edit.set_gutter_type(0, TextEdit.GUTTER_TYPE_ICON)
	_code_edit.set_gutter_width(0, 16)


func _input(event: InputEvent) -> void:
	if _code_edit.has_focus() and event is InputEventKey:
		var key := event as InputEventKey
		if key.pressed and key.keycode == KEY_S and key.ctrl_pressed:
			_save_current()
			get_viewport().set_input_as_handled()


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

func load_script(abs_path: String) -> void:
	if abs_path in _open_files:
		# Switch to existing tab
		var idx := _file_order.find(abs_path)
		if idx >= 0:
			_tab_bar.current_tab = idx
		return

	var text := _read_file(abs_path)
	_open_files[abs_path] = { "text": text, "modified": false }
	_file_order.append(abs_path)

	var tab_idx := _tab_bar.get_tab_count()
	_tab_bar.add_tab(abs_path.get_file())
	_tab_bar.current_tab = tab_idx

	_load_into_editor(abs_path)


# ---------------------------------------------------------------------------
# Tab management
# ---------------------------------------------------------------------------

func _on_tab_changed(idx: int) -> void:
	if idx < 0 or idx >= _file_order.size():
		return
	_save_current_to_cache()
	_load_into_editor(_file_order[idx])


func _on_tab_close(idx: int) -> void:
	if idx < 0 or idx >= _file_order.size():
		return
	var path: String = _file_order[idx]
	_open_files.erase(path)
	_file_order.remove_at(idx)
	_tab_bar.remove_tab(idx)
	if _file_order.is_empty():
		_code_edit.text = ""
		_status_label.text = "No file open."
	else:
		var new_idx: int = min(idx, _file_order.size() - 1)
		_tab_bar.current_tab = new_idx
		_load_into_editor(_file_order[new_idx])


func _load_into_editor(abs_path: String) -> void:
	_abs_path = abs_path
	var entry: Dictionary = _open_files.get(abs_path, {})
	_code_edit.text = entry.get("text", "")
	_code_edit.scroll_vertical = 0
	_status_label.text = abs_path
	_annotate_blocking_calls()


func _save_current_to_cache() -> void:
	if _abs_path.is_empty() or not _open_files.has(_abs_path):
		return
	_open_files[_abs_path]["text"] = _code_edit.text


func _on_text_changed() -> void:
	if _abs_path.is_empty() or not _open_files.has(_abs_path):
		return
	_open_files[_abs_path]["modified"] = true
	var idx := _file_order.find(_abs_path)
	if idx >= 0:
		var title: String = _abs_path.get_file()
		_tab_bar.set_tab_title(idx, "* " + title)


# ---------------------------------------------------------------------------
# Save
# ---------------------------------------------------------------------------

func _read_file(abs_path: String) -> String:
	var fa := FileAccess.open(abs_path, FileAccess.READ)
	if fa == null:
		push_error("[AGS] ScriptEditor: cannot open '%s' (error %d)" % [abs_path, FileAccess.get_open_error()])
		return ""
	var text := fa.get_as_text()
	fa.close()
	return text


func _save_current() -> void:
	if _abs_path.is_empty():
		return
	_save_current_to_cache()
	var text: String = _open_files[_abs_path]["text"]
	var fa := FileAccess.open(_abs_path, FileAccess.WRITE)
	if not fa:
		_status_label.text = "ERROR: could not write " + _abs_path
		return
	fa.store_string(text)
	fa.close()
	_open_files[_abs_path]["modified"] = false
	var idx := _file_order.find(_abs_path)
	if idx >= 0:
		_tab_bar.set_tab_title(idx, _abs_path.get_file())
	_status_label.text = "Saved. Running ag build…"
	_run_build()
	_annotate_blocking_calls()


func _run_build() -> void:
	if not _plugin:
		return
	# Delegate to Build Log dock if present
	var root: Window = get_tree().get_root() if get_tree() else null
	if root:
		var bl := root.find_child("Build Log", true, false)
		if bl and bl.has_method("run_build"):
			bl.call("run_build")
			return
	_status_label.text = "Saved — " + _abs_path


# ---------------------------------------------------------------------------
# Blocking-call gutter annotation
# ---------------------------------------------------------------------------

# Known blocking names (mirrors blocking.go)
const BLOCKING_GLOBALS := ["Wait", "WaitKey", "WaitMouse", "WaitInput",
	"FadeIn", "FadeOut", "Display", "DisplayMessage"]
const BLOCKING_METHODS := ["WalkTo", "WalkStraight", "FaceTo", "Say", "Think",
	"PlayAnimation", "FaceDirection", "FaceCharacter", "FacePoint", "RunInteraction"]

func _annotate_blocking_calls() -> void:
	var line_count: int = _code_edit.get_line_count()
	var clock_icon: Texture2D = _get_clock_icon()

	# Clear existing gutter icons
	for i in line_count:
		_code_edit.set_line_gutter_icon(i, 0, null)

	var rx := RegEx.new()
	# Match: word.BlockingMethod( or BlockingGlobal(
	var all_blocking := BLOCKING_GLOBALS + BLOCKING_METHODS
	rx.compile("\\b(" + "|".join(all_blocking) + ")\\s*\\(")

	for i in line_count:
		var line: String = _code_edit.get_line(i)
		if rx.search(line):
			_code_edit.set_line_gutter_icon(i, 0, clock_icon)


func _get_clock_icon() -> Texture2D:
	if not _plugin:
		return null
	return _plugin.get_editor_interface().get_base_control().get_theme_icon("Time", "EditorIcons")


# ---------------------------------------------------------------------------
# Syntax highlighter
# ---------------------------------------------------------------------------

func _make_highlighter() -> CodeHighlighter:
	var h := CodeHighlighter.new()

	# Keywords
	var kw_color := Color(0.56, 0.80, 0.98)  # blue
	for kw in ["function", "if", "else", "while", "for", "do", "return",
			"switch", "case", "default", "break", "continue",
			"int", "float", "bool", "String", "void", "var",
			"true", "false", "null", "new", "this", "and", "or", "not"]:
		h.add_keyword_color(kw, kw_color)

	# Built-in functions (non-blocking)
	var builtin_color := Color(0.80, 0.63, 0.98)  # purple
	for fn in ["AddInventory", "LoseInventory", "HasInventory", "InventoryCount",
			"GoToRoom", "SetPlayerControl", "HideRoomItem", "ShowRoomItem",
			"PlayMusic", "StopMusic", "PlaySound", "SetStatusText",
			"SaveGame", "LoadGame"]:
		h.add_keyword_color(fn, builtin_color)

	# Blocking calls — amber
	var blocking_color := Color(1.0, 0.75, 0.2)
	for fn in BLOCKING_GLOBALS + BLOCKING_METHODS:
		h.add_keyword_color(fn, blocking_color)

	# Common objects
	var obj_color := Color(0.60, 0.95, 0.70)  # green
	for obj in ["player", "game", "global"]:
		h.add_keyword_color(obj, obj_color)

	# Strings
	h.add_color_region('"', '"', Color(0.87, 0.70, 0.45))

	# Line comments
	h.add_color_region("//", "", Color(0.50, 0.55, 0.50), true)

	# Block comments
	h.add_color_region("/*", "*/", Color(0.50, 0.55, 0.50))

	# Numbers
	h.number_color = Color(0.80, 0.90, 0.60)

	# Member access dot
	h.member_variable_color = Color(0.90, 0.90, 0.90)

	return h
