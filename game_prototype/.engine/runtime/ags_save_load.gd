## ags_save_load.gd — Save/load wrapper with cutscene blocking (T-CUT25)
##
## Add as an AutoLoad named AGSSaveLoad.
##
## Wraps AGSRuntime.save_game() / load_game() / game_saved() and blocks saves
## while a cutscene is running. When save_game() is called mid-cutscene, it
## queues the slot and automatically saves when the cutscene finishes.
##
## Signals:
##   save_blocked(slot: int)      — emitted when save was attempted but blocked
##   save_queued(slot: int)       — emitted when a blocked save has been queued
##   queued_save_completed(slot)  — emitted after a queued save executes
##
## Usage:
##   AGSSaveLoad.save_game(1)     — save to slot 1 (blocked if cutscene running)
##   AGSSaveLoad.load_game(1)     — load from slot 1
##   AGSSaveLoad.game_saved(1)    — returns true if slot 1 has a save file
##   AGSSaveLoad.save_blocked     — true when saves are currently blocked

extends Node

# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

## Emitted when save_game() was called while a cutscene is active.
signal save_blocked(slot: int)

## Emitted when a blocked save has been queued for deferred execution.
signal save_queued(slot: int)

## Emitted after a previously-queued save is executed on cutscene end.
signal queued_save_completed(slot: int)

# ---------------------------------------------------------------------------
# Internal state
# ---------------------------------------------------------------------------

## Slot queued for saving after the cutscene ends. -1 = none queued.
## Only one slot is queued at a time; a newer call replaces the previous queue.
var _queued_slot: int = -1

# ---------------------------------------------------------------------------
# Public API — mirrors AGSRuntime save methods
# ---------------------------------------------------------------------------

## Returns true while a cutscene sequence is active (saves are blocked).
var save_blocked: bool:
	get:
		return _is_cutscene_active()

## Save to [param slot].
## If a cutscene is currently running, queues the save and returns false.
## Emits save_blocked and save_queued if blocked.
## Returns true if the save executed immediately.
func save_game(slot: int) -> bool:
	if _is_cutscene_active():
		save_blocked.emit(slot)
		_queued_slot = slot
		save_queued.emit(slot)
		return false
	_do_save(slot)
	return true

## Load from [param slot]. Delegates to AGSRuntime.load_game().
func load_game(slot: int) -> void:
	var runtime := _get_runtime()
	if runtime != null:
		runtime.call("load_game", slot)

## Returns true if [param slot] has a save file. Delegates to AGSRuntime.
func game_saved(slot: int) -> bool:
	var runtime := _get_runtime()
	if runtime == null:
		return false
	return runtime.call("game_saved", slot) as bool

# ---------------------------------------------------------------------------
# Queued save processing
# ---------------------------------------------------------------------------

func _ready() -> void:
	# Connect to the sequencer's sequence_complete / sequence_failed signals
	# so we can flush the queued save when the cutscene ends.
	_connect_sequencer()

## Try to connect to AGSSequencer / AGSSequencerCommands once both AutoLoads
## are ready. Called from _ready(); the sequencer may not yet be registered
## as a singleton if AutoLoad order puts us first.
func _connect_sequencer() -> void:
	var seq := _get_sequencer()
	if seq == null:
		# Retry after the first process frame — AutoLoads initialise in order.
		await get_tree().process_frame
		seq = _get_sequencer()
	if seq == null:
		return
	if not seq.is_connected("sequence_complete", _on_sequence_ended):
		seq.sequence_complete.connect(_on_sequence_ended)
	if not seq.is_connected("sequence_failed", _on_sequence_ended_reason):
		seq.sequence_failed.connect(_on_sequence_ended_reason)


func _on_sequence_ended() -> void:
	_flush_queued_save()

func _on_sequence_ended_reason(_reason: String) -> void:
	_flush_queued_save()

func _flush_queued_save() -> void:
	if _queued_slot < 0:
		return
	var slot := _queued_slot
	_queued_slot = -1
	_do_save(slot)
	queued_save_completed.emit(slot)

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

func _do_save(slot: int) -> void:
	var runtime := _get_runtime()
	if runtime != null:
		runtime.call("save_game", slot)

func _is_cutscene_active() -> bool:
	var seq := _get_sequencer()
	if seq == null:
		return false
	return seq.call("is_playing") as bool

func _get_runtime() -> Object:
	if Engine.has_singleton("AGSRuntime"):
		return Engine.get_singleton("AGSRuntime")
	return null

func _get_sequencer() -> Node:
	# Test injection: set _seq_override meta to bypass singleton lookup.
	if has_meta("_seq_override"):
		return get_meta("_seq_override") as Node
	if Engine.has_singleton("AGSSequencer"):
		return Engine.get_singleton("AGSSequencer") as Node
	if get_tree() != null:
		var n := get_tree().root.get_node_or_null("/root/AGSSequencer")
		if n != null:
			return n
	return null
