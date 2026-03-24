## UT-M4-05..07 — AGSPoint registration and get_point() tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M4: Points"

# UT-M4-05: AGSPoint self-registers with its parent AGSRoom when ready fires.
func test_05_point_registers_with_room() -> void:
	var room: AGSRoom = AGSRoom.new()
	var point: AGSPoint = AGSPoint.new()
	point.point_name = "door"
	point.position = Vector3(1.0, 0.0, 2.0)
	room.add_child(point)
	# Firing NOTIFICATION_READY on the point causes it to walk up to AGSRoom
	# and call register_point() — no scene tree required.
	point.notification(Node.NOTIFICATION_READY)

	var pos: Vector3 = room.get_point("door")
	assert_true(pos != Vector3.ZERO, "get_point('door') returned zero — point not registered")

	room.free()

# UT-M4-06: get_point() returns the correct world-space Vector3 position.
func test_06_get_point_returns_correct_position() -> void:
	var room: AGSRoom = AGSRoom.new()
	var point: AGSPoint = AGSPoint.new()
	point.point_name = "door"
	point.position = Vector3(1.0, 0.0, 2.0)
	room.add_child(point)
	point.notification(Node.NOTIFICATION_READY)

	var pos: Vector3 = room.get_point("door")
	assert_eq(pos, Vector3(1.0, 0.0, 2.0), "get_point('door') returned wrong position")

	room.free()

# UT-M4-07: get_point() with an unknown name returns Vector3.ZERO without crashing.
func test_07_get_point_unknown_name_no_crash() -> void:
	var room: AGSRoom = AGSRoom.new()
	var pos: Vector3 = room.get_point("nonexistent")
	assert_eq(pos, Vector3.ZERO, "get_point() with unknown name should return Vector3.ZERO")
	room.free()
