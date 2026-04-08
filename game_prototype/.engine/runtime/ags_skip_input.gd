## ags_skip_input.gd — T-CUT22: Cutscene skip and dialogue advance input routing.
##
## Add as an AutoLoad named AGSSkipInput to receive engine-level input.
## Input actions are read from game.agp [input] at startup and may be
## overridden via the exported properties.
##
## Routes:
##   dialogue_advance (single press)   → AGSDialogue.advance()
##   dialogue_advance (double press)   → AGSSequencer.request_skip()
##   dialogue_hold_advance (held)      → rapid advance (repeated advance() every hold_repeat_sec)
##   cutscene_skip (dedicated button)  → AGSSequencer.request_skip()
##
## All routing is engine-level; game scripts do not need to handle input.

extends Node

# ---------------------------------------------------------------------------
# Configuration (populated from game.agp [input] by AGSRuntime at startup)
# ---------------------------------------------------------------------------

## Input action name for single-press advance / double-press skip.
@export var action_advance: String = "dialogue_advance"

## Input action name for hold-to-rapid-advance.
@export var action_hold_advance: String = "dialogue_hold_advance"

## Input action name for dedicated cutscene skip button.
@export var action_cutscene_skip: String = "cutscene_skip"

## Maximum gap (seconds) between two presses to count as a double press.
@export var double_press_threshold: float = 0.3

## Interval (seconds) between advance() calls while hold_advance is held.
@export var hold_repeat_interval: float = 0.1

# ---------------------------------------------------------------------------
# Internal state
# ---------------------------------------------------------------------------

## Timestamp (seconds) of the last dialogue_advance press.
var _last_advance_time: float = 0.0

## Accumulated hold time for rapid-advance repeat.
var _hold_elapsed: float = 0.0

# ---------------------------------------------------------------------------
# Input handling
# ---------------------------------------------------------------------------

func _unhandled_input(event: InputEvent) -> void:
	# Dedicated skip button.
	if _action_exists(action_cutscene_skip) and event.is_action_pressed(action_cutscene_skip):
		_request_cutscene_skip()
		return

	# Advance / double-press skip.
	if _action_exists(action_advance) and event.is_action_pressed(action_advance):
		_on_advance_pressed()
		return


func _process(delta: float) -> void:
	# Rapid advance while hold_advance is held.
	if _action_exists(action_hold_advance) and Input.is_action_pressed(action_hold_advance):
		_hold_elapsed += delta
		if _hold_elapsed >= hold_repeat_interval:
			_hold_elapsed = 0.0
			_advance_dialogue()
	else:
		_hold_elapsed = 0.0


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

func _on_advance_pressed() -> void:
	var now: float = Time.get_unix_time_from_system()
	var elapsed: float = now - _last_advance_time
	_last_advance_time = now

	if elapsed < double_press_threshold:
		# Double press → cutscene skip.
		_request_cutscene_skip()
	else:
		# Single press → dialogue advance.
		_advance_dialogue()


## Call AGSDialogue.advance() if the dialogue system is waiting for input.
func _advance_dialogue() -> void:
	if Engine.has_singleton("AGSDialogue"):
		var dlg: Object = Engine.get_singleton("AGSDialogue")
		if dlg.get("_waiting_for_advance"):
			dlg.call("advance")
	elif get_tree() != null:
		var dlg: Node = get_tree().root.get_node_or_null("/root/AGSDialogue")
		if dlg != null and dlg.get("_waiting_for_advance"):
			dlg.call("advance")


## Ask the sequencer to skip the current cutscene.
func _request_cutscene_skip() -> void:
	if Engine.has_singleton("AGSSequencer"):
		var seq: Object = Engine.get_singleton("AGSSequencer")
		seq.call("request_skip")
	elif get_tree() != null:
		var seq: Node = get_tree().root.get_node_or_null("/root/AGSSequencer")
		if seq != null:
			seq.call("request_skip")


## Return true if the input action name is registered in the InputMap.
func _action_exists(action: String) -> bool:
	return InputMap.has_action(action)
