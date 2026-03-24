## UT-M5-04..06 — AGSCharacter navigation setup tests.
##
## NOTE: UT-M5-04 and UT-M5-05 verify structural setup only (NavigationAgent3D
## child exists, walk_to() returns the correct Signal). Actual movement behaviour
## — character arriving at destination, routing around BlockerVolume — requires
## multi-frame physics processing and is covered by TEST-END-01 (IT-END-02,03).
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M5: Navigation"

# UT-M5-04: AGSCharacter has a NavigationAgent3D child after NOTIFICATION_READY.
func test_04_character_has_nav_agent_after_ready() -> void:
	var ch: AGSCharacter = AGSCharacter.new()
	ch.character_name = "nav_test_char"
	ch.notification(Node.NOTIFICATION_READY)

	var nav_agent: NavigationAgent3D = null
	for child in ch.get_children():
		if child is NavigationAgent3D:
			nav_agent = child
			break

	assert_not_null(nav_agent, "No NavigationAgent3D child found on AGSCharacter after ready")

	ch.notification(Node.NOTIFICATION_EXIT_TREE)
	ch.free()

# UT-M5-05: walk_to() returns a Signal named "walk_completed".
func test_05_walk_to_returns_walk_completed_signal() -> void:
	var room: AGSRoom = AGSRoom.new()
	room.room_name = "nav_test_room"

	var point: AGSPoint = AGSPoint.new()
	point.point_name = "target"
	point.position = Vector3(5, 0, 0)
	room.add_child(point)
	point.notification(Node.NOTIFICATION_READY)

	var ch: AGSCharacter = AGSCharacter.new()
	ch.character_name = "walker"
	room.add_child(ch)
	ch.notification(Node.NOTIFICATION_READY)

	var sig: Signal = ch.walk_to("target")
	assert_true(sig.get_name() == "walk_completed", \
		"walk_to() returned signal with wrong name: '%s'" % sig.get_name())

	ch.notification(Node.NOTIFICATION_EXIT_TREE)
	room.free()

# UT-M5-06: Scene with WalkableSurface, BlockerVolume, and Character loads without crash.
# Full routing-around-blocker behaviour is tested in TEST-END-01 (IT-END-03).
func test_06_nav_room_scene_loads_without_crash() -> void:
	var scene := load("res://m5_character/scenes/test_nav_room.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_nav_room.tscn")

	var root_node: Node = scene.instantiate()
	assert_not_null(root_node, "Scene instantiation returned null")

	# Fire ready on key nodes to verify no crashes during setup.
	for child in root_node.get_children():
		child.notification(Node.NOTIFICATION_READY)

	root_node.free()
