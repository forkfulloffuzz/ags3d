## UT-M6-02..04 — AGSRuntime API integration tests.
##
## Verifies that AGSRuntime correctly indexes live scene nodes by name:
## get_room(), get_character(), and get_point().
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M6: RuntimeAPI"


# UT-M6-02: AGSRuntime.get_room() returns the correct room node after it enters the tree.
func test_02_get_room_returns_registered_room() -> void:
	var room := AGSRoom.new()
	room.room_name = "runtime_api_test_room"
	add_to_tree(room)
	room.notification(Node.NOTIFICATION_READY)

	var found: AGSRoom = AGSRuntime.get_room("runtime_api_test_room")
	assert_not_null(found, "AGSRuntime.get_room() returned null")
	assert_eq(found, room, "AGSRuntime.get_room() returned wrong node")

	room.free()


# UT-M6-03: AGSRuntime.get_character() returns the correct character node after it enters the tree.
func test_03_get_character_returns_registered_character() -> void:
	var room := AGSRoom.new()
	room.room_name = "char_lookup_room"
	add_to_tree(room)

	var ch := AGSCharacter3D.new()
	ch.character_name = "runtime_api_char"
	room.add_child(ch)
	ch.notification(Node.NOTIFICATION_READY)

	var found := AGSRuntime.get_character("runtime_api_char")
	assert_not_null(found, "AGSRuntime.get_character() returned null")
	assert_eq(found, ch, "AGSRuntime.get_character() returned wrong node")

	room.free()


# UT-M6-04: AGSRuntime.get_point() returns the correct Vector3 for a named point
# registered with an AGSRoom.
func test_04_get_point_returns_correct_vector3() -> void:
	var room := AGSRoom.new()
	room.room_name = "point_lookup_room"
	add_to_tree(room)
	room.notification(Node.NOTIFICATION_READY)

	var pt := AGSPoint.new()
	pt.point_name = "test_pos"
	pt.position = Vector3(4.0, 0.0, -2.0)
	room.add_child(pt)
	pt.notification(Node.NOTIFICATION_READY)

	var pos: Vector3 = AGSRuntime.get_point("point_lookup_room", "test_pos")
	assert_true(
		pos.is_equal_approx(Vector3(4.0, 0.0, -2.0)),
		"AGSRuntime.get_point() returned wrong vector: %s" % pos
	)

	room.free()
