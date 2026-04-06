## ags_dialogue.gd — AGS3D Runtime Dialogue Engine (T-DLG14)
##
## Add as an AutoLoad named AGSDialogue.
##
## Usage from AGS-spirit scripts (compiled to GDScript):
##   await AGSDialogue.start(guard, "guard_greeting")
##   await AGSDialogue.start_default(guard)
##   await AGSDialogue.start_item(item_node, "on_examine")
##
## The presenter (T-DLG16, ags_dialogue_ui.gd) connects to signals and
## renders lines/choices. State tracking (T-DLG15) connects to signals
## and persists visit counts.
##
## JSON format (produced by ag build from .agdlg sources):
##   { "source": "...", "nodes": [ { "title", "character", "body": [...] } ] }
## Statement types in body: "speaker_line", "narration", "option", "command".

extends Node

# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

## Emitted when a dialogue conversation begins.
signal dialogue_started(node_title: String)

## Emitted when a dialogue conversation ends (any exit path).
signal dialogue_ended(node_title: String)

## Emitted when a speaker line or narration is ready to display.
## For narration, char_name is "".
signal line_ready(char_name: String, text: String, loc_key: String, emotion: String)

## Emitted when the engine is waiting for the player to advance past a line.
## The presenter calls advance() when the player presses the input.
signal waiting_for_advance()

## Emitted when a set of choices is ready to display.
## Each entry: { "text": String, "loc_key": String, "available": bool }
signal choices_ready(options: Array)

## Emitted when a choice is made (by the player or programmatically).
signal choice_made(index: int)

## Emitted when an <<action>> or <<set>> command fires.
## The game can connect this to evaluate expressions.
signal command_fired(raw: String)

# ---------------------------------------------------------------------------
# State
# ---------------------------------------------------------------------------

## All loaded dialogue nodes, indexed by title.
var _nodes: Dictionary = {}

## Whether a dialogue is currently playing.
var _active: bool = false

## Whether the engine is waiting for advance() to be called.
var _waiting: bool = false

## Signal used internally to resume after advance() or choose().
var _resume_signal: Signal

## Loaded flag — set to true after load_all() is called.
var _loaded: bool = false

# ---------------------------------------------------------------------------
# Loading
# ---------------------------------------------------------------------------

## Load all dialogue JSON files from the generated directory.
## Call once from your game's autoload _ready().
func load_all(generated_dir: String = "res://.engine/generated/dialogue/") -> void:
	_nodes.clear()
	var dir := DirAccess.open(generated_dir)
	if dir == null:
		push_warning("AGSDialogue: generated dialogue directory not found: " + generated_dir)
		_loaded = true
		return
	dir.list_dir_begin()
	var fname := dir.get_next()
	while fname != "":
		if fname.ends_with(".json"):
			_load_file(generated_dir.path_join(fname))
		fname = dir.get_next()
	dir.list_dir_end()
	_loaded = true

func _load_file(path: String) -> void:
	var text := FileAccess.get_file_as_string(path)
	if text == "":
		push_warning("AGSDialogue: could not read " + path)
		return
	var parsed: Variant = JSON.parse_string(text)
	if parsed == null or not parsed is Dictionary:
		push_warning("AGSDialogue: invalid JSON in " + path)
		return
	var nodes: Array = parsed.get("nodes", [])
	for node_data in nodes:
		if node_data is Dictionary and node_data.has("title"):
			_nodes[node_data["title"]] = node_data

# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

## Start a dialogue conversation at the given node title (blocking coroutine).
## char_node may be null for narrator-only conversations.
func start(char_node: Node, node_title: String) -> void:
	if _active:
		push_warning("AGSDialogue.start: dialogue already active, ignoring request for " + node_title)
		return
	_active = true
	dialogue_started.emit(node_title)
	await _execute_node(node_title)
	_active = false
	dialogue_ended.emit(node_title)

## Start the default entry-point node for a character (first root in .agchar dialogue block).
## Requires the character node to expose a get_default_dialogue_node() method.
func start_default(char_node: Node) -> void:
	var node_title: String = ""
	if char_node != null and char_node.has_method("get_default_dialogue_node"):
		node_title = char_node.get_default_dialogue_node()
	if node_title == "":
		push_warning("AGSDialogue.start_default: no default node for character")
		return
	await start(char_node, node_title)

## Start a dialogue triggered by item interaction.
## trigger is the field name: "on_examine" or "on_use_failed".
func start_item(item_node: Node, trigger: String) -> void:
	var node_title: String = ""
	if item_node != null and item_node.has_method("get_dialogue_node"):
		node_title = item_node.get_dialogue_node(trigger)
	if node_title == "":
		return
	await start(null, node_title)

## Called by the presenter or input handler when the player advances past a line.
func advance() -> void:
	if _waiting:
		_waiting = false
		_advance_signal.emit()

## Called by the presenter when the player selects a choice.
func choose(index: int) -> void:
	if _waiting:
		_waiting = false
		choice_made.emit(index)
		_advance_signal.emit(index)

# ---------------------------------------------------------------------------
# Internal advance signal (one-shot per wait point)
# ---------------------------------------------------------------------------

signal _advance_signal(choice_index: int)

# ---------------------------------------------------------------------------
# Execution
# ---------------------------------------------------------------------------

func _execute_node(title: String) -> void:
	var node_data: Variant = _nodes.get(title)
	if node_data == null:
		push_warning("AGSDialogue: node not found: " + title)
		return
	var body: Array = node_data.get("body", [])
	await _execute_body(body)

func _execute_body(stmts: Array) -> void:
	var i := 0
	while i < stmts.size():
		var stmt: Dictionary = stmts[i]
		var stype: String = stmt.get("type", "")
		match stype:
			"speaker_line", "narration":
				await _execute_line(stmt)
			"option":
				# Collect contiguous option block starting at i.
				var opts: Array = []
				while i < stmts.size() and stmts[i].get("type", "") == "option":
					opts.append(stmts[i])
					i += 1
				await _execute_options(opts)
				continue  # i already advanced past all options
			"command":
				_execute_command(stmt.get("raw", ""))
		i += 1

func _execute_line(stmt: Dictionary) -> void:
	var speaker: String = stmt.get("speaker", "")
	var text: String = stmt.get("text", "")
	var loc_key: String = stmt.get("loc_key", "")
	var emotion: String = ""  # future: parse from commands list if present

	# Localise text if a locale is active.
	var display_text := _localise(loc_key, text)

	# Fire inline commands (e.g. <<action>>, <<emotion:>>) before displaying.
	var cmds: Array = stmt.get("commands", [])
	for cmd in cmds:
		_execute_command(cmd)

	line_ready.emit(speaker, display_text, loc_key, emotion)
	await _wait_for_advance()

func _execute_options(opts: Array) -> void:
	# Build option list, evaluating visibility/availability.
	var display_opts: Array = []
	var bodies: Array = []
	for opt in opts:
		var text: String = opt.get("text", "")
		var loc_key: String = opt.get("loc_key", "")
		var display_text := _localise(loc_key, text)
		var available: bool = _eval_condition(opt.get("condition", ""))
		display_opts.append({"text": display_text, "loc_key": loc_key, "available": available})
		bodies.append(opt.get("body", []))

	# If all options are unavailable, skip the choice block.
	var any_available := display_opts.any(func(o: Dictionary) -> bool: return o["available"])
	if not any_available:
		return

	choices_ready.emit(display_opts)
	var chosen_idx: int = await _wait_for_choice()
	choice_made.emit(chosen_idx)

	# Execute the chosen branch.
	if chosen_idx >= 0 and chosen_idx < bodies.size():
		await _execute_body(bodies[chosen_idx])

func _execute_command(raw: String) -> void:
	if raw == "":
		return
	# Handle built-in jump command.
	if raw.begins_with("<<jump ") and raw.ends_with(">>"):
		var target: String = raw.substr(7, raw.length() - 9).strip_edges()
		await _execute_node(target)
		return
	if raw == "<<end>>":
		return
	# All other commands emit a signal for the game to handle.
	command_fired.emit(raw)

# ---------------------------------------------------------------------------
# Wait helpers
# ---------------------------------------------------------------------------

func _wait_for_advance() -> void:
	_waiting = true
	waiting_for_advance.emit()
	await _advance_signal

func _wait_for_choice() -> int:
	_waiting = true
	var result: Array = [0]
	var conn := _advance_signal.connect(func(idx: int) -> void: result[0] = idx,
		CONNECT_ONE_SHOT)
	await _advance_signal
	# conn is auto-disconnected by CONNECT_ONE_SHOT; just return the captured index.
	return result[0]

# ---------------------------------------------------------------------------
# Localisation
# ---------------------------------------------------------------------------

func _localise(loc_key: String, fallback: String) -> String:
	# If a LocalisationRuntime is present, ask it for the translated string.
	if Engine.has_singleton("AGSLocalisation"):
		var tr: String = Engine.get_singleton("AGSLocalisation").get_string(loc_key)
		if tr != "":
			return tr
	return fallback

# ---------------------------------------------------------------------------
# Condition evaluation
# ---------------------------------------------------------------------------

## Evaluate a visible_if / available_if condition string.
## Returns true if the condition is empty (unconditional) or evaluates truthy.
func _eval_condition(cond: String) -> bool:
	if cond == "":
		return true
	# Simple flag check: "flag.name" → AGSRuntime.get_global("name")
	if cond.begins_with("flag.") or cond.begins_with("not flag."):
		var negate := cond.begins_with("not ")
		var expr := cond.substr(4) if negate else cond
		var flag_name := expr.substr(5)  # strip "flag."
		var val: Variant = AGSRuntime.get_global(flag_name)
		var result: bool = val != null and val != false and val != 0 and val != ""
		return result if not negate else not result
	# For complex conditions, emit a signal and default to true.
	# The game can override _eval_condition by extending this script.
	return true
