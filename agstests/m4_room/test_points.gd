## UT-M4-05..07 — AGSPoint registration and get_point() tests.
extends "res://utils/test_base.gd"

# UT-M4-05: AGSPoint self-registers with its parent AGSRoom on ready.
func test_05_point_registers_with_room() -> void:
	var scene := load("res://m4_room/scenes/test_room_with_point.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_room_with_point.tscn")

	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	# get_point() would crash or return zero-Vector3 if point wasn't registered.
	var pos: Vector3 = root_node.get_point("door")
	# Position is non-zero — the point is at (1, 0, 2) in the fixture.
	assert_true(pos != Vector3.ZERO, "get_point('door') returned zero — point not registered")

	tree_root.remove_child(root_node)
	root_node.queue_free()

# UT-M4-06: get_point() returns the correct world-space Vector3 position.
func test_06_get_point_returns_correct_position() -> void:
	var scene := load("res://m4_room/scenes/test_room_with_point.tscn") as PackedScene
	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	var pos: Vector3 = root_node.get_point("door")
	assert_eq(pos, Vector3(1.0, 0.0, 2.0), "get_point('door') returned wrong position")

	tree_root.remove_child(root_node)
	root_node.queue_free()

# UT-M4-07: get_point() with an unknown name returns Vector3.ZERO without crashing.
func test_07_get_point_unknown_name_no_crash() -> void:
	var room: AGSRoom = AGSRoom.new()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(room)

	var pos: Vector3 = room.get_point("nonexistent")
	assert_eq(pos, Vector3.ZERO, "get_point() with unknown name should return Vector3.ZERO")

	tree_root.remove_child(room)
	room.queue_free()
