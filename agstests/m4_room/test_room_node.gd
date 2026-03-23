## UT-M4-01..02 — AGSRoom node instantiation and property tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M4: RoomNode"

# UT-M4-01: AGSRoom instantiates and is a Node3D subclass.
func test_01_ags_room_is_node3d() -> void:
	var room: AGSRoom = AGSRoom.new()
	assert_not_null(room, "AGSRoom.new() returned null")
	assert_true(room is Node3D, "AGSRoom is not a Node3D subclass")
	room.free()

# UT-M4-02: room_name property reads and writes correctly.
func test_02_room_name_property() -> void:
	var room: AGSRoom = AGSRoom.new()
	room.room_name = "library"
	assert_eq(room.room_name, "library", "room_name did not round-trip correctly")
	room.free()
