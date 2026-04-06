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

		var bg_id: String = step.get("bg", "")
		if bg_id != "":
			# Background step: fire without awaiting.
			_fire_bg_step(step, bg_id)
		else:
			# Foreground step: execute and wait for completion.
			await _execute_step(step)
		i += 1

	# Wait for all remaining background steps before signalling complete.
	await _sync_all_bg()

	_active = false
	_bg_steps.clear()
	sequence_complete.emit()

# ---------------------------------------------------------------------------
# Step execution
# ---------------------------------------------------------------------------

## Execute a foreground step, waiting for its completion signal.
func _execute_step(step: Dictionary) -> void:
	var sid: String = step.get("id", step.get("type", "step"))
	step_started.emit(sid)

	var done := await _dispatch_step(step)
	if done:
		step_complete.emit(sid)
	else:
		step_failed.emit(sid)

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
		_:
			# Unknown step type — no-op (executor plugins handle known types).
			return true

## Override in subclasses or connect a handler to process action/set commands.
func _on_command(raw: String) -> void:
	pass  # Game integrators connect to command_fired signal via emit_command.

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
