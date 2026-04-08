## T-CUT17 — AGSSequencerCommands character command tests.
##
## Tests the "character" step type dispatching in ags_sequencer_commands.gd.
##
## NOTE: walk_to and run_to require a navmesh to complete; in headless mode
## the step starts but navigation never finishes, so those tests use a short
## timeout and assert that step_started was emitted (dispatch is confirmed).
## face_to uses a tween and works headless when a room is in the scene tree.
extends "res://utils/test_base.gd"

const SeqCmds := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")
const CharScript := preload("res://../game_prototype/.engine/runtime/ags_character.gd")

func suite_name() -> String:
	return "M-CUT: CharCommands (T-CUT17)"


# ── Helpers ────────────────────────────────────────────────────────────────────

## Create an AGSSequencerCommands node and add it to the tree.
func _make_seq() -> Node:
	var s: Node = SeqCmds.new()
	_tree.root.add_child(s)
	return s

## Create an AGSCharacter3D with the ags_character.gd runtime script.
## The character registers with AGSRuntime on _ready() (triggered by add_child).
## [param char_name] must be unique per test to avoid AGSRuntime collisions.
func _make_char(char_name: String) -> Node:
	var ch: Node = AGSCharacter3D.new()
	ch.set_script(CharScript)
	ch.character_name = char_name
	_tree.root.add_child(ch)
	return ch

## Create a minimal AGSRoom and add it to the scene tree so that face_to
## and spawn_at can resolve a room. The room will register with AGSRuntime.
func _make_room(room_name: String) -> Node:
	var room: Node = AGSRoom.new()
	room.room_name = room_name
	_tree.root.add_child(room)
	return room

## Run a single step through the sequencer and return [completed: bool, failed: bool].
func _run_step(seq: Node, step: Dictionary) -> Array:
	var completed := [false]
	var failed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	seq.sequence_failed.connect(func(_r: String) -> void: failed[0] = true, CONNECT_ONE_SHOT)
	await seq.run([step])
	return [completed[0], failed[0]]

## Queue-free and wait a frame.
func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT17-01: hide command sets visible = false ────────────────────────────

func test_01_hide_sets_invisible() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_01")
	ch.visible = true

	var step := {"type": "character", "character": "c17_01", "command": "hide"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete")
	assert_false(ch.visible, "character should be invisible after hide")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-02: show command sets visible = true ─────────────────────────────

func test_02_show_sets_visible() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_02")
	ch.visible = false

	var step := {"type": "character", "character": "c17_02", "command": "show"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete")
	assert_true(ch.visible, "character should be visible after show")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-03: move_speed command sets the speed property ──────────────────

func test_03_move_speed_sets_value() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_03")

	var step := {"type": "character", "character": "c17_03", "command": "move_speed", "value": 9.5}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete")
	assert_eq(ch.move_speed, 9.5, "move_speed should be updated to 9.5")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-04: expression command emits expression_changed signal ───────────

func test_04_expression_emits_signal() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_04")

	var received := [""]
	ch.expression_changed.connect(func(name: String) -> void: received[0] = name, CONNECT_ONE_SHOT)

	var step := {"type": "character", "character": "c17_04", "command": "expression", "name": "surprised"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete")
	assert_eq(received[0], "surprised", "expression_changed should fire with correct name")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-05: unknown character name causes step failure ──────────────────

func test_05_unknown_character_fails() -> void:
	var seq := _make_seq()
	seq.cutscene_fallback = "log_and_continue"

	var started := [false]
	var failed_step := [false]
	seq.step_started.connect(func(_id: String) -> void: started[0] = true, CONNECT_ONE_SHOT)
	seq.step_failed.connect(func(_id: String) -> void: failed_step[0] = true, CONNECT_ONE_SHOT)

	var step := {"type": "character", "character": "no_such_char_17_05", "command": "hide"}
	await seq.run([step])

	assert_true(started[0], "step_started should fire")
	assert_true(failed_step[0], "step_failed should fire for unknown character")
	await _cleanup_nodes([seq])


# ── UT-CUT17-06: unknown character command is a no-op (not a failure) ────────

func test_06_unknown_command_noop() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_06")

	var step := {"type": "character", "character": "c17_06", "command": "do_the_funky_chicken"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete (unknown command is no-op)")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-07: animation command plays clip and awaits completion ───────────

func test_07_animation_plays_clip() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_07")

	# Build a minimal AnimationPlayer with a very short clip.
	var ap := AnimationPlayer.new()
	var lib := AnimationLibrary.new()
	var anim := Animation.new()
	anim.length = 0.001
	lib.add_animation("Cheer", anim)
	ap.add_animation_library("", lib)
	ch.add_child(ap)

	var step := {"type": "character", "character": "c17_07", "command": "animation", "clip": "Cheer", "loop": false}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sequence should complete after animation finishes")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-08: animation with missing clip returns false ────────────────────

func test_08_animation_missing_clip_fails() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_08")
	# No AnimationPlayer attached → play_clip returns false.

	note("WARNING 'no AnimationPlayer found' below: intentional — testing missing clip failure")
	var step := {"type": "character", "character": "c17_08", "command": "animation", "clip": "Ghost", "loop": false}
	seq.cutscene_fallback = "log_and_continue"

	var failed_step := [false]
	seq.step_failed.connect(func(_id: String) -> void: failed_step[0] = true, CONNECT_ONE_SHOT)
	await seq.run([step])

	assert_true(failed_step[0], "step_failed should fire when AnimationPlayer is missing")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-09: animation loop:true returns without waiting ──────────────────

func test_09_animation_loop_returns_immediately() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_09")

	var ap := AnimationPlayer.new()
	var lib := AnimationLibrary.new()
	var anim := Animation.new()
	anim.length = 100.0  # Long animation — would hang if loop=false awaited it.
	anim.loop_mode = Animation.LOOP_LINEAR
	lib.add_animation("IdleLoop", anim)
	ap.add_animation_library("", lib)
	ch.add_child(ap)

	# If loop:true awaited animation_finished, this test would hang forever.
	var step := {"type": "character", "character": "c17_09", "command": "animation", "clip": "IdleLoop", "loop": true}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "loop animation should complete the step immediately")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-10: face_to completes when room is in scene tree ────────────────

func test_10_face_to_completes_with_room() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_10")
	var room := _make_room("test_room_17_10")
	# Character is not a child of the room but _find_parent_room() falls back
	# to scanning root children, so it will find the AGSRoom.

	var face_fired := [false]
	ch.face_completed.connect(func() -> void: face_fired[0] = true, CONNECT_ONE_SHOT)

	var step := {"type": "character", "character": "c17_10", "command": "face_to", "target": "nowhere_special"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "face_to step should complete")
	assert_true(face_fired[0], "face_completed should have been emitted")
	await _cleanup_nodes([seq, ch, room])


# ── UT-CUT17-11: spawn_at moves character to point position ──────────────────

func test_11_spawn_at_sets_position() -> void:
	var seq := _make_seq()
	var ch := _make_char("c17_11")
	# No room → _resolve_point returns Vector3.ZERO → character moves to origin.
	ch.global_position = Vector3(5.0, 0.0, 5.0)

	var step := {"type": "character", "character": "c17_11", "command": "spawn_at", "point": "start"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "spawn_at step should complete")
	assert_eq(ch.global_position, Vector3.ZERO, "spawn_at should move character to point (Vector3.ZERO when no room)")
	await _cleanup_nodes([seq, ch])


# ── UT-CUT17-12: walk_to dispatches without hanging in headless ───────────────

func test_12_walk_to_dispatches_headless() -> void:
	note("walk_to in headless: navigation never completes (no navmesh). " +
		"The step starts and the sequencer awaits walk_completed. " +
		"We verify step_started fired, then abort the test via a short timeout on the seq.")
	var seq := _make_seq()
	var ch := _make_char("c17_12")
	var room := _make_room("test_room_17_12")

	var started := [false]
	seq.step_started.connect(func(_id: String) -> void: started[0] = true, CONNECT_ONE_SHOT)

	# Use a tiny step timeout so the step fails (times out) rather than hanging.
	var step := {
		"type": "character", "character": "c17_12",
		"command": "walk_to", "point": "door",
		"timeout": 0.05
	}
	seq.cutscene_fallback = "log_and_continue"
	await seq.run([step])

	assert_true(started[0], "step_started should fire — walk_to was dispatched")
	await _cleanup_nodes([seq, ch, room])


# ── UT-CUT17-13: run_to temporarily boosts move_speed during walk ─────────────

func test_13_run_to_boosts_speed() -> void:
	note("run_to in headless: navmesh missing, step times out after 0.05 s. " +
		"Speed restoration only happens when walk_to completes normally " +
		"(coroutine is orphaned on timeout). We only assert speed was boosted.")
	var seq := _make_seq()
	var ch := _make_char("c17_13")
	var room := _make_room("test_room_17_13")
	ch.move_speed = 4.0

	var speed_during_run := [0.0]
	# Sample speed after a brief delay so walk_to has started.
	_tree.create_timer(0.01).timeout.connect(
		func() -> void: speed_during_run[0] = ch.move_speed,
		CONNECT_ONE_SHOT
	)

	var step := {
		"type": "character", "character": "c17_13",
		"command": "run_to", "point": "exit", "speed": 12.0,
		"timeout": 0.05
	}
	seq.cutscene_fallback = "log_and_continue"
	await seq.run([step])

	assert_eq(speed_during_run[0], 12.0, "move_speed should be 12.0 during run_to")
	await _cleanup_nodes([seq, ch, room])
