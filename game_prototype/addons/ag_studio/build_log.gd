@tool
extends VBoxContainer

## AG Studio Build Log dock (T-E15)
##
## Shows output from the most recent `ag build` / `ag validate` run.
## Errors are shown in red with clickable file:line links that open the
## offending .agscript in the system editor (T-E14 will route to the
## custom script editor instead).
##
## Usage: call set_plugin(p) immediately after instantiation, then use
## append_line() / clear() / run_build() from ag_studio.gd.

signal build_finished(success: bool)

var _plugin: EditorPlugin
var _log: RichTextLabel
var _build_btn: Button
var _clear_btn: Button

# Regex: match "path/to/file.ext:LINE: message"
const _ERROR_PATTERN := "^(.+\\.ag\\w+):(\\d+): (.+)$"


func set_plugin(p: EditorPlugin) -> void:
	_plugin = p


func _ready() -> void:
	name = "Build Log"
	_build_ui()


func _build_ui() -> void:
	var toolbar := HBoxContainer.new()
	add_child(toolbar)

	_build_btn = Button.new()
	_build_btn.text = "▶ Build"
	_build_btn.flat = true
	_build_btn.pressed.connect(run_build)
	toolbar.add_child(_build_btn)

	_clear_btn = Button.new()
	_clear_btn.text = "✕ Clear"
	_clear_btn.flat = true
	_clear_btn.pressed.connect(clear)
	toolbar.add_child(_clear_btn)

	add_child(HSeparator.new())

	_log = RichTextLabel.new()
	_log.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_log.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_log.bbcode_enabled = true
	_log.scroll_following = true
	_log.selection_enabled = true
	_log.context_menu_enabled = true
	_log.meta_clicked.connect(_on_link_clicked)
	add_child(_log)


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

func clear() -> void:
	_log.clear()


func append_line(line: String) -> void:
	var rx := RegEx.new()
	rx.compile(_ERROR_PATTERN)
	var m := rx.search(line)
	if m:
		var path: String = m.get_string(1)
		var lineno: String = m.get_string(2)
		var msg: String = m.get_string(3)
		var is_error: bool = msg.to_lower().begins_with("error") or line.to_lower().begins_with("error")
		var color: String = "ff5555" if is_error else "ffaa00"
		var meta: String = "%s:%s" % [path, lineno]
		_log.append_text("[color=#%s][url=%s]%s:%s[/url]: %s[/color]\n" % [color, meta, path, lineno, msg])
	else:
		_log.append_text(line + "\n")


## Run `ag build` in the project directory and stream output to the log.
func run_build() -> void:
	clear()
	append_line("→ ag build…")
	_build_btn.disabled = true

	var project_dir: String = ProjectSettings.globalize_path("res://")
	var ag_bin: String = _find_ag_binary()

	if ag_bin.is_empty():
		append_line("[color=#ff5555]ERROR: ag binary not found. Run .dev/build-ag.sh ag[/color]")
		_build_btn.disabled = false
		return

	var output: Array = []
	var exit_code: int = OS.execute(ag_bin, ["build"], output, true)

	for line: String in output:
		for part: String in line.split("\n"):
			if not part.strip_edges().is_empty():
				append_line(part)

	var success: bool = exit_code == 0
	if success:
		append_line("[color=#55ff55]✓ Build succeeded.[/color]")
	else:
		append_line("[color=#ff5555]✗ Build failed (exit %d).[/color]" % exit_code)

	_build_btn.disabled = false
	build_finished.emit(success)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

func _find_ag_binary() -> String:
	# Look for ag binary relative to the repo root (two levels above res://)
	var res_path: String = ProjectSettings.globalize_path("res://")
	var repo_root: String = res_path.get_base_dir()
	var candidates := [
		repo_root.path_join("bin/ag"),
		repo_root.path_join("tools/ag/ag"),
	]
	for path: String in candidates:
		if FileAccess.file_exists(path):
			return path
	return ""


func _on_link_clicked(meta: Variant) -> void:
	var parts: PackedStringArray = str(meta).split(":")
	if parts.size() < 2:
		return
	var file_path: String = parts[0]
	# Resolve relative path against res://
	if not file_path.is_absolute_path():
		file_path = ProjectSettings.globalize_path("res://").path_join(file_path)
	OS.shell_open(file_path)
