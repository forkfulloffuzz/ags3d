## UT-M5-04..06 — AGSCharacter navigation setup tests.
##
## Navigation behaviour (walk_to, face_to, NavigationAgent3D) is implemented in
## the GDScript runtime (.engine/runtime/ags_character.gd), not in C++.
## The C++ AGSCharacter owns: character_name, move_speed, walk_completed, face_completed.
## Actual movement is covered by UT-M6-10..12 (end-to-end physics tests).
extends "res://utils/test_base.gd"

const CHAR_SCRIPT := "res://m6_integration/runtime/ags_character.gd"

func suite_name() -> String:
	return "M5: Navigation"

# UT-M5-04: AGSCharacter with the runtime script attached has a NavigationAgent3D
# child after _ready() runs. NavigationAgent3D creation is the runtime's responsibility.
func test_04_character_has_nav_agent_after_ready() -> void:
	var root := Node.new()
	add_to_tree(root)

	var ch: AGSCharacter = AGSCharacter.new()
	ch.character_name = "nav_test_char"
	ch.set_script(load(CHAR_SCRIPT))  # runtime script creates NavigationAgent3D in _ready()
	root.add_child(ch)  # tree propagates READY since root is live

	var nav_agent: NavigationAgent3D = null
	for child in ch.get_children():
		if child is NavigationAgent3D:
			nav_agent = child
			break

	assert_not_null(nav_agent, "No NavigationAgent3D child found after runtime script _ready()")

	root.free()

# UT-M5-05: AGSCharacter (C++) declares walk_completed and face_completed signals.
# walk_to / face_to are GDScript coroutines on the runtime script; this verifies
# the signals the coroutines emit are declared at the C++ level.
func test_05_walk_completed_and_face_completed_signals_declared() -> void:
	var ch := AGSCharacter.new()
	assert_true(ch.has_signal("walk_completed"),
			"AGSCharacter must declare walk_completed signal")
	assert_true(ch.has_signal("face_completed"),
			"AGSCharacter must declare face_completed signal")
	ch.free()

# UT-M5-06: Scene with WalkableSurface, BlockerVolume, and Character loads without crash.
# Full routing-around-blocker behaviour is tested in TEST-END-01 (IT-END-03).
func test_06_nav_room_scene_loads_without_crash() -> void:
	var scene := load("res://m5_character/scenes/test_nav_room.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_nav_room.tscn")

	var root_node: Node = scene.instantiate()
	assert_not_null(root_node, "Scene instantiation returned null")

	# Add to scene tree first so NavigationAgent3D has a valid viewport,
	# then fire ready manually (deferred _ready() never runs in _init()).
	add_to_tree(root_node)
	for child in root_node.get_children():
		child.notification(Node.NOTIFICATION_READY)

	root_node.free()
