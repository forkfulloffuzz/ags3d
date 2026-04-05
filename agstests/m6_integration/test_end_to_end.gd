## UT-M6-10..12 — End-to-end script execution tests.
##
## These tests require multi-frame physics processing. They are run via
## run_suite_async() in run_tests.gd so each test_* method can use await.
##
## UT-M6-10: Character.walk_to() in a generated script drives the character to its destination.
## UT-M6-11: Character.face_to() in a generated script drives the character's rotation.
## UT-M6-12: Full prototype — generated script calls walk_to then face_to;
##           character reaches door_left (< 0.5 m) then faces window (dot > 0.95).
extends "res://utils/test_base.gd"

const CHAR_SCRIPT  := "res://m6_integration/runtime/ags_character.gd"
const E2E_SCENE    := "res://m6_integration/scenes/test_e2e_room.tscn"
const WALKTO_SCRIPT := "res://m6_integration/scripts/test_walkto.agscript.gd"

## Physics frames to wait per second of timeout (Godot default = 60 ticks/s).
const FRAMES_PER_SECOND := 60
## Maximum wall-clock seconds for a movement test to complete.
const WALK_TIMEOUT_SEC  := 12

func suite_name() -> String:
	return "M6: EndToEnd"


## Await sig for up to timeout_sec seconds (measured in physics frames).
## Returns true if the signal fired, false on timeout.
func _await_signal(sig: Signal, timeout_sec: int) -> bool:
	var fired := [false]
	sig.connect(func() -> void: fired[0] = true, CONNECT_ONE_SHOT)
	var max_frames := timeout_sec * FRAMES_PER_SECOND
	for _i in max_frames:
		if fired[0]:
			return true
		await _tree.physics_frame
	return fired[0]


## Fire NOTIFICATION_READY on every child of p_node, then on p_node itself.
## Needed because _ready() is deferred and never fires in headless test code.
func _fire_ready_recursive(p_node: Node) -> void:
	for child in p_node.get_children():
		_fire_ready_recursive(child)
	p_node.notification(Node.NOTIFICATION_READY)


## Load the e2e scene, add to the real scene tree, and fire READY on all nodes.
## Returns the instantiated root node (AGSRoom). Caller must free() it.
func _setup_e2e_room() -> AGSRoom:
	var packed := load(E2E_SCENE) as PackedScene
	var room := packed.instantiate() as AGSRoom
	add_to_tree(room)
	_fire_ready_recursive(room)
	return room


# UT-M6-10: Calling walk_to() on a character with the runtime script attached
# drives it to within 0.5 m of the target point within the timeout.
func test_10_walk_to_reaches_destination() -> void:
	var room := _setup_e2e_room()

	# Wait two physics frames: one for the nav map to sync, one for stability.
	await _tree.physics_frame
	await _tree.physics_frame

	var ch: AGSCharacterBase = room.get_node("e2e_player")
	var door_pos := Vector3(-4.0, 0.0, 0.0)

	ch.walk_to("door_left")  # coroutine runs in background
	var reached := await _await_signal(ch.walk_completed, WALK_TIMEOUT_SEC)

	assert_true(reached, "walk_to('door_left') did not complete within %d seconds" % WALK_TIMEOUT_SEC)
	if reached:
		var dist := ch.global_position.distance_to(door_pos)
		assert_true(dist < 0.5, "Character did not reach door_left — distance: %.2f m" % dist)

	room.free()


# UT-M6-11: Calling face_to() on a character with the runtime script attached
# rotates it to face the target point (forward dot ≥ 0.95) within 2 seconds.
func test_11_face_to_rotates_character() -> void:
	var room := _setup_e2e_room()

	await _tree.physics_frame

	var ch: AGSCharacterBase = room.get_node("e2e_player")
	var window_pos := Vector3(4.0, 0.0, 4.0)

	ch.face_to("window")  # coroutine runs in background
	var done := await _await_signal(ch.face_completed, 5)

	assert_true(done, "face_to('window') did not complete within 5 seconds")
	if done:
		var dir := (window_pos - ch.global_position)
		dir.y = 0.0
		dir = dir.normalized()
		var forward := -ch.global_transform.basis.z  # Godot -Z is forward
		var dot := forward.dot(dir)
		assert_true(dot > 0.95, "Character is not facing window — dot product: %.3f" % dot)

	room.free()


# UT-M6-12: Full prototype success criterion as an automated test.
# The generated script (test_walkto.agscript.gd) is attached to the room.
# On room_enter it walks the character to door_left then faces the window.
# Pass condition: character within 0.5 m of door_left AND facing window (dot ≥ 0.95).
func test_12_full_prototype_script_drives_walk_and_face() -> void:
	var room_script := load(WALKTO_SCRIPT) as GDScript
	assert_not_null(room_script, "Could not load generated room script: " + WALKTO_SCRIPT)

	var packed := load(E2E_SCENE) as PackedScene
	var room := packed.instantiate() as AGSRoom

	# Set script BEFORE entering the tree so that AGSRoom's NOTIFICATION_READY
	# sets up the room_enter() call into the generated script.
	room.set_script(room_script)

	add_to_tree(room)
	_fire_ready_recursive(room)  # READY fires: sets up signals, calls room_enter()
	# room_enter() starts walk_to which awaits physics_frame internally.

	# Let the navigation system initialise (nav server processes the baked mesh).
	await _tree.physics_frame
	await _tree.physics_frame

	var ch: AGSCharacterBase = room.get_node("e2e_player")

	# Await face_completed — the last signal in the walk_to → face_to chain.
	var done := await _await_signal(ch.face_completed, WALK_TIMEOUT_SEC)

	assert_true(done, "Prototype script did not complete within %d seconds" % WALK_TIMEOUT_SEC)

	if done:
		var door_pos   := Vector3(-4.0, 0.0, 0.0)
		var window_pos := Vector3(4.0, 0.0, 4.0)

		var dist := ch.global_position.distance_to(door_pos)
		assert_true(dist < 0.5, "Character did not reach door_left — distance: %.2f m" % dist)

		var dir := (window_pos - ch.global_position)
		dir.y = 0.0
		dir = dir.normalized()
		var forward := -ch.global_transform.basis.z
		var dot := forward.dot(dir)
		assert_true(dot > 0.95, "Character not facing window after face_to — dot: %.3f" % dot)

	room.free()
