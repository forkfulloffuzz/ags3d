## ags_dialogue_ui.gd — AGS3D Dialogue Presenter (T-DLG16)
##
## Add as an AutoLoad named AGSDialogueUI (or instantiate as a child of your
## game's UI layer).
##
## Renders dialogue lines and choice lists using a CanvasLayer so they draw
## on top of the 3D scene. Connects to AGSDialogue signals automatically.
##
## Signals emitted by the presenter (for portrait / voice systems):
##   portrait_requested(char_name: String, emotion: String)
##
## Layout (created programmatically so no .tscn is required):
##
##   CanvasLayer (layer 50)
##   └─ Panel (dialogue_panel)
##      ├─ VBoxContainer
##      │  ├─ Label  (speaker_label)
##      │  └─ Label  (text_label)
##      └─ VBoxContainer (choices_container)
##         └─ [Button × N]  (one per option)

extends Node

# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

## Emitted when a line has an emotion tag — portrait system should react.
signal portrait_requested(char_name: String, emotion: String)

# ---------------------------------------------------------------------------
# Settings
# ---------------------------------------------------------------------------

## Time in seconds before the line auto-advances (0 = wait indefinitely).
@export var auto_advance_delay: float = 0.0

## Input action name for advancing dialogue.
@export var advance_action: String = "ui_accept"

# ---------------------------------------------------------------------------
# UI nodes (created in _ready)
# ---------------------------------------------------------------------------

var _layer: CanvasLayer
var _panel: Panel
var _speaker_label: Label
var _text_label: Label
var _choices_container: VBoxContainer
var _choice_buttons: Array[Button] = []

## Reference to the dialogue engine (set in _ready).
var _dlg: Node = null

# ---------------------------------------------------------------------------
# State
# ---------------------------------------------------------------------------

var _showing_line: bool = false
var _showing_choices: bool = false
var _auto_advance_timer: SceneTreeTimer = null

# ---------------------------------------------------------------------------
# Lifecycle
# ---------------------------------------------------------------------------

func _ready() -> void:
	_build_ui()
	_panel.visible = false

	# Connect to AGSDialogue if present.
	var root := get_tree().root if get_tree() != null else null
	if root != null:
		_dlg = root.get_node_or_null("/root/AGSDialogue")
	if _dlg != null:
		_dlg.dialogue_started.connect(_on_dialogue_started)
		_dlg.dialogue_ended.connect(_on_dialogue_ended)
		_dlg.line_ready.connect(_on_line_ready)
		_dlg.waiting_for_advance.connect(_on_waiting_for_advance)
		_dlg.choices_ready.connect(_on_choices_ready)

func _build_ui() -> void:
	_layer = CanvasLayer.new()
	_layer.layer = 50
	add_child(_layer)

	_panel = Panel.new()
	_panel.custom_minimum_size = Vector2(800, 160)
	_layer.add_child(_panel)

	var vbox := VBoxContainer.new()
	vbox.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	vbox.add_theme_constant_override("separation", 8)
	_panel.add_child(vbox)

	_speaker_label = Label.new()
	_speaker_label.add_theme_font_size_override("font_size", 16)
	vbox.add_child(_speaker_label)

	_text_label = Label.new()
	_text_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_text_label.add_theme_font_size_override("font_size", 14)
	vbox.add_child(_text_label)

	_choices_container = VBoxContainer.new()
	_choices_container.add_theme_constant_override("separation", 4)
	_panel.add_child(_choices_container)

	# Anchor panel to bottom of viewport.
	_panel.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
	_panel.offset_top = -160

# ---------------------------------------------------------------------------
# Input
# ---------------------------------------------------------------------------

func _unhandled_input(event: InputEvent) -> void:
	if not _showing_line:
		return
	if event.is_action_pressed(advance_action):
		get_viewport().set_input_as_handled()
		_try_advance()

func _try_advance() -> void:
	if _auto_advance_timer != null:
		_auto_advance_timer = null
	if _dlg != null:
		_dlg.advance()

# ---------------------------------------------------------------------------
# Signal handlers
# ---------------------------------------------------------------------------

func _on_dialogue_started(_node_title: String) -> void:
	_panel.visible = true

func _on_dialogue_ended(_node_title: String) -> void:
	_panel.visible = false
	_showing_line = false
	_showing_choices = false
	_clear_choices()

func _on_line_ready(char_name: String, text: String, _loc_key: String, emotion: String) -> void:
	_showing_choices = false
	_clear_choices()
	_choices_container.visible = false

	_speaker_label.text = char_name
	_speaker_label.visible = char_name != ""
	_text_label.text = text
	_showing_line = true

	if emotion != "":
		portrait_requested.emit(char_name, emotion)

func _on_waiting_for_advance() -> void:
	if auto_advance_delay > 0.0 and get_tree() != null:
		_auto_advance_timer = get_tree().create_timer(auto_advance_delay)
		_auto_advance_timer.timeout.connect(_try_advance, CONNECT_ONE_SHOT)

func _on_choices_ready(options: Array) -> void:
	_showing_line = false
	_showing_choices = true
	_clear_choices()

	_speaker_label.visible = false
	_text_label.visible = false
	_choices_container.visible = true

	for i in range(options.size()):
		var opt: Dictionary = options[i]
		var btn := Button.new()
		btn.text = opt.get("text", "")
		btn.disabled = not opt.get("available", true)
		var idx := i  # capture for closure
		btn.pressed.connect(func() -> void: _on_choice_pressed(idx))
		_choices_container.add_child(btn)
		_choice_buttons.append(btn)

func _on_choice_pressed(index: int) -> void:
	if _dlg != null and _showing_choices:
		_showing_choices = false
		_text_label.visible = true
		_choices_container.visible = false
		_dlg.choose(index)

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

func _clear_choices() -> void:
	for btn in _choice_buttons:
		btn.queue_free()
	_choice_buttons.clear()
