## ags_dialogue_state.gd — AGS3D Dialogue State Tracking (T-DLG15)
##
## Add as an AutoLoad named AGSDialogueState.
##
## Tracks per-node and per-option visit state for the dialogue engine.
## Persists in the save graph under the key "dialogue_state".
##
## Connect to AGSDialogue signals in _ready():
##   AGSDialogue.dialogue_started.connect(_on_dialogue_started)
##   AGSDialogue.choice_made.connect(_on_choice_made)
##
## Query API (mirrors AGS-spirit surface):
##   AGSDialogueState.node_visited("guard_greeting")     → bool
##   AGSDialogueState.option_seen("guard_greeting", 0)   → bool
##   AGSDialogueState.visit_count("guard_greeting")      → int

extends Node

# ---------------------------------------------------------------------------
# Schema versioning for save compatibility
# ---------------------------------------------------------------------------

const STATE_SCHEMA_VERSION := 1

# ---------------------------------------------------------------------------
# State tables
# ---------------------------------------------------------------------------

## Set of node titles that have been entered (visited_nodes).
var _visited_nodes: Dictionary = {}   # title → visit count (int)

## Set of (title, index) option pairs seen by the player.
## Key: "title:index" string.
var _seen_options: Dictionary = {}    # "title:index" → true

## Tracks the current active node title during a conversation.
var _current_node: String = ""

## Tracks current option index being offered (set by choices_ready).
var _current_choices: Array = []

# ---------------------------------------------------------------------------
# Lifecycle
# ---------------------------------------------------------------------------

func _ready() -> void:
	if Engine.has_singleton("AGSDialogue") or _has_autoload("AGSDialogue"):
		var dlg: Node = _get_autoload("AGSDialogue")
		if dlg != null:
			dlg.dialogue_started.connect(_on_dialogue_started)
			dlg.choices_ready.connect(_on_choices_ready)
			dlg.choice_made.connect(_on_choice_made)

func _has_autoload(name: String) -> bool:
	return Engine.get_main_loop() != null and Engine.get_main_loop().root.has_node("/root/" + name)

func _get_autoload(name: String) -> Node:
	var root := Engine.get_main_loop().root if Engine.get_main_loop() != null else null
	if root == null:
		return null
	return root.get_node_or_null("/root/" + name)

# ---------------------------------------------------------------------------
# Signal handlers
# ---------------------------------------------------------------------------

func _on_dialogue_started(node_title: String) -> void:
	_current_node = node_title
	_visited_nodes[node_title] = (_visited_nodes.get(node_title, 0) as int) + 1

func _on_choices_ready(options: Array) -> void:
	_current_choices = options

func _on_choice_made(index: int) -> void:
	if _current_node != "" and index >= 0:
		_seen_options[_current_node + ":" + str(index)] = true

# ---------------------------------------------------------------------------
# Query API
# ---------------------------------------------------------------------------

## Returns true if the player has entered node_title at least once.
func node_visited(node_title: String) -> bool:
	return _visited_nodes.has(node_title) and _visited_nodes[node_title] > 0

## Returns true if the player has selected option at index in node_title.
func option_seen(node_title: String, index: int) -> bool:
	return _seen_options.has(node_title + ":" + str(index))

## Returns the number of times node_title has been entered.
func visit_count(node_title: String) -> int:
	return _visited_nodes.get(node_title, 0) as int

# ---------------------------------------------------------------------------
# Save / load integration
# ---------------------------------------------------------------------------

## Serialise state to a Dictionary for the save graph.
## Call from AGSRuntime's save hook.
func serialise() -> Dictionary:
	return {
		"schema_version": STATE_SCHEMA_VERSION,
		"visited_nodes": _visited_nodes.duplicate(),
		"seen_options": _seen_options.duplicate(),
	}

## Restore state from a previously serialised Dictionary.
## Call from AGSRuntime's load hook.
func deserialise(data: Dictionary) -> void:
	var version: int = data.get("schema_version", 0)
	if version < 1:
		push_warning("AGSDialogueState: unknown schema version %d, skipping load" % version)
		return
	_visited_nodes = data.get("visited_nodes", {})
	_seen_options = data.get("seen_options", {})

## Reset all state (called on new game).
func reset() -> void:
	_visited_nodes.clear()
	_seen_options.clear()
	_current_node = ""
	_current_choices.clear()
