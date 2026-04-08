## ags_sequencer.gd — AGS3D Core Cutscene Sequencer (T-CUT12)
##
## Add as an AutoLoad named AGSSequencer.
##
## Executes a list of sequencer steps (produced from a compiled .agcut JSON).
## Supports foreground steps (block until complete), background steps (fire and
## continue, tracked by id), and sync points (wait for named bg ids).
##
## Step states:
##   pending   → not yet started
##   running   → currently executing
##   complete  → finished successfully
##   failed    → timed out or executor returned an error
##   skipped   → bypassed by skip system
##
## Signals:
##   step_started(id: String)
##   step_complete(id: String)
##   step_failed(id: String)
##   sequence_complete
##   sequence_failed(reason: String)
##
## Usage:
##   var steps = AGSSequencer.load_cutscene("chapter1_opening")
##   await AGSSequencer.run(steps)

extends Node

# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

signal step_started(id: String)
signal step_complete(id: String)
signal step_failed(id: String)
signal sequence_complete
signal sequence_failed(reason: String)

# ---------------------------------------------------------------------------
# Step state constants
# ---------------------------------------------------------------------------

const STATE_PENDING  := "pending"
const STATE_RUNNING  := "running"
const STATE_COMPLETE := "complete"
const STATE_FAILED   := "failed"
const STATE_SKIPPED  := "skipped"

# ---------------------------------------------------------------------------
# Internal state
# ---------------------------------------------------------------------------

## True while a sequence is executing.
var _active: bool = false

## Background steps indexed by their id.
## Each entry: { "id": String, "state": String, "signal": Signal }
var _bg_steps: Dictionary = {}  # id → { state, done_signal }

## Loaded cutscene data, indexed by title.
var _cutscenes: Dictionary = {}  # title → Array[Dictionary]

# ---------------------------------------------------------------------------
# Loading
# ---------------------------------------------------------------------------

## Load all cutscene JSON files from the generated directory.
func load_all(generated_dir: String = "res://.engine/generated/cutscenes/") -> void:
	_cutscenes.clear()
	var dir := DirAccess.open(generated_dir)
	if dir == null:
		push_warning("AGSSequencer: generated cutscene directory not found: " + generated_dir)
		return
	dir.list_dir_begin()
	var fname := dir.get_next()
	while fname != "":
		if fname.ends_with(".json"):
			_load_file(generated_dir.path_join(fname))
		fname = dir.get_next()
	dir.list_dir_end()

func _load_file(path: String) -> void:
	var text := FileAccess.get_file_as_string(path)
	if text == "":
		return
	var parsed: Variant = JSON.parse_string(text)
	if parsed == null or not parsed is Dictionary:
		return
	var title: String = (parsed as Dictionary).get("title", "")
	if title != "":
		_cutscenes[title] = (parsed as Dictionary).get("steps", [])

## Return the step list for a cutscene title (for testing / direct injection).
func get_steps(title: String) -> Array:
	return _cutscenes.get(title, [])

# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

## Execute a cutscene by title (blocking coroutine).
## Resolves steps from the compiled cutscene data.
func play(title: String) -> void:
	var steps: Array = _cutscenes.get(title, [])
	if steps.is_empty():
		push_warning("AGSSequencer.play: cutscene not found: " + title)
		return
	await run(steps)

## Execute a list of step Dictionaries directly (blocking coroutine).
## This is the core execution engine.
func run(steps: Array) -> void:
	if _active:
		push_warning("AGSSequencer.run: a sequence is already active, ignoring")
		return
	_active = true
	_bg_steps.clear()
	_retry_tracker.clear()

	var i := 0
	while i < steps.size():
		var step: Dictionary = steps[i]
		var stype: String = step.get("type", "")

		if stype == "sync":
			await _execute_sync(step)
			i += 1
			continue

		if stype == "end":
			break

		if stype == "label":
			i += 1
			continue

		if stype == "skip_to":
			var target_label: String = step.get("label", step.get("name", ""))
			var target := _find_label(steps, target_label)
			if target >= 0:
				i = target
			else:
				push_warning("AGSSequencer: skip_to label '%s' not found" % target_label)
				i += 1
			continue

		var bg_id: String = step.get("bg", "")
		if bg_id != "":
			# Background step: fire without awaiting.
			_fire_bg_step(step, bg_id)
			i += 1
			continue

		# Foreground step: execute and wait for completion.
		var ok: bool = await _execute_step(step)
		if not ok:
			var keep_going := await _apply_fallback(step, steps)
			if not keep_going:
				return
			# If _apply_fallback did a jump_to, _jump_target holds the new index.
			if _jump_target >= 0:
				i = _jump_target
				_jump_target = -1
				continue
			elif _retry_pending:
				_retry_pending = false
				continue  # retry same i (do NOT increment)
		i += 1

	# Wait for all remaining background steps before signalling complete.
	await _sync_all_bg()

	_active = false
	_bg_steps.clear()
	_retry_tracker.clear()
	sequence_complete.emit()

# ---------------------------------------------------------------------------
# Step execution
# ---------------------------------------------------------------------------

## Default step timeout in seconds (0 = no global default).
## Set from game.agp [cutscenes] step_timeout_default at startup.
var step_timeout_default: float = 0.0

## Per-cutscene fallback policy (from compiled cutscene header).
## Overridden per-step via the "on_fail" field.
## Valid values: "skip_and_continue", "halt", "log_and_continue",
##               "retry_once", "jump_to <label>"
var cutscene_fallback: String = "halt"

## Tracks which step ids have already been retried (retry_once policy).
var _retry_tracker: Dictionary = {}

## Jump target index set by _apply_fallback when policy is "jump_to <label>".
## -1 means no pending jump.
var _jump_target: int = -1

## Set to true by _apply_fallback when the step should be retried.
var _retry_pending: bool = false

## Execute a foreground step, waiting for its completion signal.
## Respects per-step "timeout" field and step_timeout_default.
## "timeout:none" disables timeout for that step (e.g. dialogue, video).
## Returns true on success, false on failure.
func _execute_step(step: Dictionary) -> bool:
	var sid: String = step.get("id", step.get("type", "step"))
	step_started.emit(sid)

	var timeout_val: Variant = step.get("timeout", null)
	var use_timeout: float = 0.0
	if timeout_val is String and timeout_val == "none":
		use_timeout = 0.0  # explicitly disabled
	elif timeout_val is float or timeout_val is int:
		use_timeout = float(timeout_val)
	elif step_timeout_default > 0.0:
		use_timeout = step_timeout_default

	var done: bool
	if use_timeout > 0.0 and get_tree() != null:
		done = await _dispatch_with_timeout(step, use_timeout)
	else:
		done = await _dispatch_step(step)

	if done:
		step_complete.emit(sid)
	else:
		step_failed.emit(sid)
	return done

## Execute a step with a hard timeout. Returns false if the step times out.
func _dispatch_with_timeout(step: Dictionary, timeout_secs: float) -> bool:
	# Use single-element arrays so lambda closures can mutate these flags.
	var timer_done := [false]
	get_tree().create_timer(timeout_secs).timeout.connect(
		func() -> void: timer_done[0] = true,
		CONNECT_ONE_SHOT
	)

	var dispatch_result := [false]
	var dispatch_done := [false]
	_dispatch_step_async(step, dispatch_result, func() -> void: dispatch_done[0] = true)

	while not dispatch_done[0] and not timer_done[0]:
		await get_tree().process_frame

	if dispatch_done[0]:
		return dispatch_result[0]
	# Timer fired before step completed.
	return false

## Async wrapper for _dispatch_step that stores result and calls on_done.
func _dispatch_step_async(step: Dictionary, result_box: Array, on_done: Callable) -> void:
	result_box[0] = await _dispatch_step(step)
	on_done.call()

## Fire a background step without blocking. Tracks it by id.
func _fire_bg_step(step: Dictionary, bg_id: String) -> void:
	_bg_steps[bg_id] = {"state": STATE_RUNNING}
	step_started.emit(bg_id)
	# Run in background — capture result asynchronously.
	_run_bg_async(step, bg_id)

func _run_bg_async(step: Dictionary, bg_id: String) -> void:
	var ok := await _dispatch_step(step)
	if _bg_steps.has(bg_id):
		_bg_steps[bg_id]["state"] = STATE_COMPLETE if ok else STATE_FAILED
	if ok:
		step_complete.emit(bg_id)
	else:
		step_failed.emit(bg_id)

## Dispatch a step to its executor. Returns true on success, false on failure.
## Subclasses or command executor plugins override this to handle specific types.
func _dispatch_step(step: Dictionary) -> bool:
	var stype: String = step.get("type", "")
	match stype:
		"wait":
			var seconds: float = step.get("duration", 0.0) as float
			if get_tree() != null:
				await get_tree().create_timer(seconds).timeout
			return true
		"action", "set":
			# Fire command signal for the game to handle.
			var raw: String = step.get("raw", "")
			if raw != "":
				_on_command(raw)
			return true
		"fail":
			# Test-only step type that always fails.
			return false
		_:
			# Unknown step type — no-op (executor plugins handle known types).
			return true

## Override in subclasses or connect a handler to process action/set commands.
func _on_command(raw: String) -> void:
	pass  # Game integrators connect to command_fired signal via emit_command.

# ---------------------------------------------------------------------------
# Fallback policy
# ---------------------------------------------------------------------------

## Resolve the effective fallback policy for a step.
## Priority: per-step "on_fail" → cutscene_fallback → "halt".
func _resolve_policy(step: Dictionary) -> String:
	var p: String = step.get("on_fail", "").strip_edges()
	if p != "":
		return p
	if cutscene_fallback.strip_edges() != "":
		return cutscene_fallback
	return "halt"

## Apply the fallback policy for a failed step.
## Sets _jump_target or _retry_pending as side-effects.
## Returns true if the sequence should continue, false if it should halt.
func _apply_fallback(step: Dictionary, steps: Array) -> bool:
	_jump_target = -1
	_retry_pending = false
	var sid: String = step.get("id", step.get("type", "step"))
	var policy: String = _resolve_policy(step)

	if policy == "skip_and_continue":
		_fire_state_changes(step)
		return true

	if policy == "log_and_continue":
		push_warning("AGSSequencer: step '%s' failed — continuing (log_and_continue)" % sid)
		_fire_state_changes(step)
		return true

	if policy == "retry_once":
		if not _retry_tracker.has(sid):
			_retry_tracker[sid] = true
			_retry_pending = true
			return true
		# Retry already exhausted — escalate to halt.
		push_error("AGSSequencer: step '%s' failed after retry — halting" % sid)
		_halt("step '%s' failed after retry" % sid)
		return false

	if policy.begins_with("jump_to "):
		var label: String = policy.substr(8).strip_edges()
		_fire_state_changes(step)
		var target: int = _find_label(steps, label)
		if target >= 0:
			_jump_target = target
			return true
		push_error("AGSSequencer: jump_to label '%s' not found — halting" % label)
		_halt("jump_to label '%s' not found" % label)
		return false

	# Default: halt.
	_halt("step '%s' failed" % sid)
	return false

## Fire embedded <<action>> / <<set>> commands from a failing step.
## State changes always execute regardless of fallback policy.
func _fire_state_changes(step: Dictionary) -> void:
	var raw: String = step.get("raw", "")
	if raw != "":
		_on_command(raw)

## Tear down active state and emit sequence_failed.
func _halt(reason: String) -> void:
	_active = false
	_bg_steps.clear()
	_retry_tracker.clear()
	sequence_failed.emit(reason)

## Find the index of a label step by name.  Returns -1 if not found.
func _find_label(steps: Array, label: String) -> int:
	for idx in steps.size():
		var s: Dictionary = steps[idx]
		if s.get("type", "") == "label" and s.get("name", "") == label:
			return idx
	return -1

# ---------------------------------------------------------------------------
# Sync points
# ---------------------------------------------------------------------------

## Wait for specific background step ids (or all if none specified).
func _execute_sync(step: Dictionary) -> void:
	var ids: Array = step.get("ids", [])
	if ids.is_empty():
		await _sync_all_bg()
	else:
		for id: String in ids:
			await _wait_for_bg(id)

## Wait until all active background steps are complete.
func _sync_all_bg() -> void:
	while true:
		var any_running := false
		for id: String in _bg_steps:
			if _bg_steps[id]["state"] == STATE_RUNNING:
				any_running = true
				break
		if not any_running:
			break
		await get_tree().process_frame

## Wait for a specific background step id to finish.
func _wait_for_bg(id: String) -> void:
	if not _bg_steps.has(id):
		return  # Already done or never started.
	while _bg_steps.get(id, {}).get("state", STATE_COMPLETE) == STATE_RUNNING:
		await get_tree().process_frame

# ---------------------------------------------------------------------------
# State queries
# ---------------------------------------------------------------------------

## Returns true while a sequence is executing.
func is_active() -> bool:
	return _active

## Returns the state of a background step by id.
func bg_step_state(id: String) -> String:
	return _bg_steps.get(id, {}).get("state", STATE_PENDING) as String

## Returns all currently tracked background step ids.
func bg_step_ids() -> Array:
	return _bg_steps.keys()
