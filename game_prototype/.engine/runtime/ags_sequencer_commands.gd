## ags_sequencer_commands.gd — Command executor extension for AGSSequencer (T-CUT16–T-CUT21)
##
## Extends the base AGSSequencer with game-command dispatch. Use this script
## as the AutoLoad (instead of ags_sequencer.gd) for full cutscene support.
##
## Implemented command types:
##   character  — T-CUT17: walk_to, run_to, animation, face_to, spawn_at,
##                          hide, show, expression, move_speed
##   camera     — T-CUT16: set, move_to, look_at, follow, shake, fov, return (stub)
##   audio      — T-CUT18: music, sound, ambient, voice (stub)
##   visual     — T-CUT19: fade_in, fade_out, flash, vignette, letterbox,
##                          overlay, video (stub)
##   flow/state — T-CUT20: parallel, if/else, on event (base wait/action/set are in base class)
##   dialogue   — T-CUT21: line, narrator, title_card, subtitle, choice, dialogue (stub)
##
## Step dictionary format — character example:
##   {"type": "character", "character": "player", "command": "walk_to", "point": "door"}
##   {"type": "character", "character": "player", "command": "animation", "clip": "Wave", "loop": false}
##   {"type": "character", "character": "player", "command": "face_to", "target": "window"}
##   {"type": "character", "character": "player", "command": "spawn_at", "point": "spawn_main"}
##   {"type": "character", "character": "player", "command": "hide"}
##   {"type": "character", "character": "player", "command": "show"}
##   {"type": "character", "character": "player", "command": "expression", "name": "happy"}
##   {"type": "character", "character": "player", "command": "move_speed", "value": 6.0}
##   {"type": "character", "character": "player", "command": "run_to", "point": "exit", "speed": 8.0}

extends "ags_sequencer.gd"

# ---------------------------------------------------------------------------
# Step dispatch — override base class to add command types
# ---------------------------------------------------------------------------

func _dispatch_step(step: Dictionary) -> bool:
	var stype: String = step.get("type", "")
	match stype:
		"character":
			return await _exec_character(step)
		"camera":
			return await _exec_camera(step)
		"audio":
			return await _exec_audio(step)
		"visual":
			return await _exec_visual(step)
		"parallel":
			return await _exec_parallel(step)
		"if":
			return await _exec_if(step)
		"on_event":
			return await _exec_on_event(step)
		"dialogue_line", "narrator_line", "title_card", "subtitle", "choice", "dialogue":
			return await _exec_dialogue(step)
		_:
			# Delegate all other types (wait, action, set, fail, unknown) to base class.
			return await super._dispatch_step(step)


# ---------------------------------------------------------------------------
# T-CUT17 — Character commands
# ---------------------------------------------------------------------------

## Resolve a character node by name via AGSRuntime.
## Returns null and logs an error if not found.
func _get_character(char_name: String) -> Node:
	if not Engine.has_singleton("AGSRuntime"):
		push_error("AGSSequencerCommands: AGSRuntime singleton not found")
		return null
	var runtime: Object = Engine.get_singleton("AGSRuntime")
	var ch: Object = runtime.call("get_character", char_name)
	if ch == null:
		push_error("AGSSequencerCommands: character '%s' not found" % char_name)
		return null
	return ch as Node


func _exec_character(step: Dictionary) -> bool:
	var char_name: String = step.get("character", "")
	var ch: Node = _get_character(char_name)
	if ch == null:
		return false

	var cmd: String = step.get("command", "")
	match cmd:
		"walk_to":
			await ch.walk_to(step.get("point", ""))
			return true

		"run_to":
			var original_speed: float = ch.move_speed
			var run_speed: float = step.get("speed", original_speed * 2.0) as float
			ch.move_speed = run_speed
			await ch.walk_to(step.get("point", ""))
			ch.move_speed = original_speed
			return true

		"animation":
			var clip: String = step.get("clip", "")
			var loop: bool = step.get("loop", false)
			if clip.is_empty():
				push_warning("AGSSequencerCommands: animation command missing 'clip'")
				return false
			return await ch.play_clip(clip, loop)

		"face_to":
			await ch.face_to(step.get("target", ""))
			return true

		"spawn_at":
			var point_name: String = step.get("point", "")
			var pos: Vector3 = _resolve_point(point_name)
			ch.global_position = pos
			return true

		"hide":
			ch.visible = false
			return true

		"show":
			ch.visible = true
			return true

		"expression":
			ch.set_expression(step.get("name", ""))
			return true

		"move_speed":
			ch.move_speed = step.get("value", ch.move_speed) as float
			return true

		_:
			push_warning("AGSSequencerCommands: unknown character command '%s'" % cmd)
			return true  # Unknown command is a no-op, not a failure.


## Resolve a named point from the current room.
## Returns Vector3.ZERO if the room or point is not found.
func _resolve_point(point_name: String) -> Vector3:
	if not Engine.has_singleton("AGSRuntime"):
		return Vector3.ZERO
	var runtime: Object = Engine.get_singleton("AGSRuntime")
	var room_name: String = runtime.call("get_current_room") as String
	if room_name.is_empty():
		return Vector3.ZERO
	var room: Object = runtime.call("get_room", room_name)
	if room == null:
		return Vector3.ZERO
	return room.call("get_point", point_name) as Vector3


# ---------------------------------------------------------------------------
# T-CUT16 — Camera commands (stub; implemented in T-CUT16)
# ---------------------------------------------------------------------------

func _exec_camera(step: Dictionary) -> bool:
	push_warning("AGSSequencerCommands: camera commands not yet implemented (T-CUT16)")
	return true


# ---------------------------------------------------------------------------
# T-CUT18 — Audio commands (stub; implemented in T-CUT18)
# ---------------------------------------------------------------------------

func _exec_audio(step: Dictionary) -> bool:
	push_warning("AGSSequencerCommands: audio commands not yet implemented (T-CUT18)")
	return true


# ---------------------------------------------------------------------------
# T-CUT19 — Visual commands (stub; implemented in T-CUT19)
# ---------------------------------------------------------------------------

func _exec_visual(step: Dictionary) -> bool:
	push_warning("AGSSequencerCommands: visual commands not yet implemented (T-CUT19)")
	return true


# ---------------------------------------------------------------------------
# T-CUT20 — Flow/state commands (stub for complex types; T-CUT20)
# ---------------------------------------------------------------------------

## Execute a <<parallel>> block: all sub-steps fire simultaneously.
## Completes when the longest sub-step completes.
func _exec_parallel(step: Dictionary) -> bool:
	push_warning("AGSSequencerCommands: parallel blocks not yet implemented (T-CUT20)")
	return true


## Evaluate <<if>> / <<else>> / <<end_if>>.
func _exec_if(step: Dictionary) -> bool:
	push_warning("AGSSequencerCommands: if blocks not yet implemented (T-CUT20)")
	return true


## Register <<on event:>> handler.
func _exec_on_event(step: Dictionary) -> bool:
	push_warning("AGSSequencerCommands: on_event not yet implemented (T-CUT20)")
	return true


# ---------------------------------------------------------------------------
# T-CUT21 — Dialogue commands (stub; implemented in T-CUT21)
# ---------------------------------------------------------------------------

func _exec_dialogue(step: Dictionary) -> bool:
	push_warning("AGSSequencerCommands: dialogue commands not yet implemented (T-CUT21)")
	return true
