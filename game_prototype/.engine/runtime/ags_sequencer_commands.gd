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
# T-CUT16 — Camera commands
# ---------------------------------------------------------------------------

## Resolve the active Camera3D from the viewport (or null in headless).
func _get_active_camera() -> Camera3D:
	if get_viewport() == null:
		return null
	return get_viewport().get_camera_3d()

## Resolve a camera node by name via AGSRuntime, or return the active camera.
## [param cam_name]: if non-empty, looks up that name; otherwise returns active camera.
func _get_camera(cam_name: String) -> Camera3D:
	if not cam_name.is_empty() and Engine.has_singleton("AGSRuntime"):
		var runtime: Object = Engine.get_singleton("AGSRuntime")
		var cam: Object = runtime.call("get_camera", cam_name)
		if cam != null:
			return cam as Camera3D
	return _get_active_camera()

## Parse an ease type string to a Tween.EaseType constant.
func _parse_ease(ease_str: String) -> Tween.EaseType:
	match ease_str:
		"ease_in":    return Tween.EASE_IN
		"ease_out":   return Tween.EASE_OUT
		"linear":     return Tween.EASE_IN_OUT  # closest linear approx
		_:            return Tween.EASE_IN_OUT

## Parse a trans type string to a Tween.TransitionType constant.
func _parse_trans(trans_str: String) -> Tween.TransitionType:
	match trans_str:
		"linear":  return Tween.TRANS_LINEAR
		"sine":    return Tween.TRANS_SINE
		"cubic":   return Tween.TRANS_CUBIC
		"expo":    return Tween.TRANS_EXPO
		_:         return Tween.TRANS_SINE


func _exec_camera(step: Dictionary) -> bool:
	var cam_name: String = step.get("camera", "")
	var cmd: String = step.get("command", "")

	match cmd:
		"set":
			# Instantly activate a named camera (switch to it) and optionally
			# override FOV and rotation.
			var target_name: String = step.get("point", cam_name)
			if not target_name.is_empty() and Engine.has_singleton("AGSRuntime"):
				var runtime: Object = Engine.get_singleton("AGSRuntime")
				runtime.call("set_camera", target_name)
			var cam: Camera3D = _get_camera(target_name if not target_name.is_empty() else cam_name)
			if cam == null:
				push_warning("AGSSequencerCommands: camera 'set' — no camera found")
				return true
			if step.has("fov"):
				cam.fov = step.get("fov") as float
			if step.has("rotation"):
				var r: Array = step.get("rotation") as Array
				if r.size() == 3:
					cam.rotation_degrees = Vector3(r[0], r[1], r[2])
			return true

		"move_to":
			# Tween the active camera to the named camera's position/transform.
			var cam: Camera3D = _get_camera(cam_name)
			if cam == null:
				push_warning("AGSSequencerCommands: camera 'move_to' — no camera found")
				return true
			var target_cam: Camera3D = _get_camera(step.get("point", ""))
			if target_cam == null or target_cam == cam:
				return true
			var duration: float = step.get("duration", 1.0) as float
			var tween := create_tween()
			tween.set_ease(_parse_ease(step.get("ease", "")))
			tween.set_trans(_parse_trans(step.get("ease", "")))
			tween.tween_property(cam, "global_position", target_cam.global_position, duration)
			tween.parallel().tween_property(cam, "global_rotation", target_cam.global_rotation, duration)
			if step.has("fov"):
				tween.parallel().tween_property(cam, "fov", step.get("fov") as float, duration)
			await tween.finished
			return true

		"look_at":
			# Tween camera to face a target (character or point name).
			var cam: Camera3D = _get_camera(cam_name)
			if cam == null:
				push_warning("AGSSequencerCommands: camera 'look_at' — no camera found")
				return true
			var target_name: String = step.get("target", "")
			var target_pos: Vector3 = _resolve_look_at_target(target_name)
			var duration: float = step.get("duration", 0.0) as float
			if duration <= 0.0:
				cam.look_at(target_pos)
			else:
				var target_basis := Basis.looking_at(target_pos - cam.global_position)
				var target_rot: Vector3 = target_basis.get_euler()
				var tween := create_tween()
				tween.set_ease(_parse_ease(step.get("ease", "")))
				tween.tween_property(cam, "global_rotation", target_rot, duration)
				await tween.finished
			return true

		"follow":
			# Follow a character for 'duration' seconds (0 = one frame, then return).
			var cam: Camera3D = _get_camera(cam_name)
			if cam == null:
				push_warning("AGSSequencerCommands: camera 'follow' — no camera found")
				return true
			var char_name: String = step.get("character", "")
			var ch: Node = _get_character(char_name)
			if ch == null:
				return false
			var offset_arr: Array = step.get("offset", [0.0, 2.0, -4.0]) as Array
			var offset := Vector3(
				offset_arr[0] if offset_arr.size() > 0 else 0.0,
				offset_arr[1] if offset_arr.size() > 1 else 2.0,
				offset_arr[2] if offset_arr.size() > 2 else -4.0
			)
			var duration: float = step.get("duration", 0.0) as float
			var elapsed: float = 0.0
			while elapsed < duration or duration <= 0.0:
				cam.global_position = (ch as Node3D).global_position + offset
				cam.look_at((ch as Node3D).global_position)
				if duration <= 0.0:
					break
				await get_tree().process_frame
				elapsed += get_process_delta_time()
			return true

		"shake":
			# Shake the active camera for 'duration' seconds.
			var cam: Camera3D = _get_camera(cam_name)
			if cam == null:
				push_warning("AGSSequencerCommands: camera 'shake' — no camera found")
				return true
			var intensity: float = step.get("intensity", 0.1) as float
			var duration: float = step.get("duration", 0.3) as float
			var falloff: bool = step.get("falloff", true)
			var origin: Vector3 = cam.global_position
			var elapsed: float = 0.0
			while elapsed < duration:
				await get_tree().process_frame
				elapsed += get_process_delta_time()
				var t: float = elapsed / duration
				var scale: float = (1.0 - t) if falloff else 1.0
				cam.global_position = origin + Vector3(
					randf_range(-intensity, intensity) * scale,
					randf_range(-intensity, intensity) * scale,
					0.0
				)
			cam.global_position = origin
			return true

		"fov":
			# Change camera FOV, optionally over a duration.
			var cam: Camera3D = _get_camera(cam_name)
			if cam == null:
				push_warning("AGSSequencerCommands: camera 'fov' — no camera found")
				return true
			var target_fov: float = step.get("value", 75.0) as float
			var duration: float = step.get("duration", 0.0) as float
			if duration <= 0.0:
				cam.fov = target_fov
			else:
				var tween := create_tween()
				tween.set_ease(_parse_ease(step.get("ease", "")))
				tween.tween_property(cam, "fov", target_fov, duration)
				await tween.finished
			return true

		"return":
			# Return to the room's initial camera, optionally tweened.
			var initial: String = _get_initial_camera_name()
			if initial.is_empty():
				push_warning("AGSSequencerCommands: camera 'return' — no initial camera in current room")
				return true
			if Engine.has_singleton("AGSRuntime"):
				var runtime: Object = Engine.get_singleton("AGSRuntime")
				runtime.call("set_camera", initial)
			var target_cam: Camera3D = _get_camera(initial)
			var cam: Camera3D = _get_active_camera()
			var duration: float = step.get("duration", 0.0) as float
			if cam == null or target_cam == null or duration <= 0.0:
				return true
			var tween := create_tween()
			tween.set_ease(_parse_ease(step.get("ease", "")))
			tween.tween_property(cam, "global_position", target_cam.global_position, duration)
			tween.parallel().tween_property(cam, "global_rotation", target_cam.global_rotation, duration)
			await tween.finished
			return true

		_:
			push_warning("AGSSequencerCommands: unknown camera command '%s'" % cmd)
			return true


## Resolve a look_at target to a world-space Vector3.
## Checks characters first, then falls back to named points.
func _resolve_look_at_target(name: String) -> Vector3:
	if Engine.has_singleton("AGSRuntime"):
		var runtime: Object = Engine.get_singleton("AGSRuntime")
		var ch: Object = runtime.call("get_character", name)
		if ch != null:
			return (ch as Node3D).global_position
	return _resolve_point(name)


## Return the initial camera name for the current room (empty string if unavailable).
func _get_initial_camera_name() -> String:
	if not Engine.has_singleton("AGSRuntime"):
		return ""
	var runtime: Object = Engine.get_singleton("AGSRuntime")
	var room_name: String = runtime.call("get_current_room") as String
	if room_name.is_empty():
		return ""
	var room: Object = runtime.call("get_room", room_name)
	if room == null:
		return ""
	return room.call("get_initial_camera") as String


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
