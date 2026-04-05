## UT-M5-07..09 — Sequential WalkTo and FaceTo Signal contract tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M5: FaceTo"

# Returns [room, character] — caller must free room (frees children).
func _make_room_with_character_and_points() -> Array:
	var room: AGSRoom = AGSRoom.new()
	room.room_name = "faceto_test_room"

	var pt_a: AGSPoint = AGSPoint.new()
	pt_a.point_name = "door"
	pt_a.position = Vector3(4, 0, 0)
	room.add_child(pt_a)
	pt_a.notification(Node.NOTIFICATION_READY)

	var pt_b: AGSPoint = AGSPoint.new()
	pt_b.point_name = "window"
	pt_b.position = Vector3(-4, 0, 0)
	room.add_child(pt_b)
	pt_b.notification(Node.NOTIFICATION_READY)

	var ch := AGSCharacter3D.new()
	ch.character_name = "face_test_char"
	room.add_child(ch)
	ch.notification(Node.NOTIFICATION_READY)

	return [room, ch]

# UT-M5-07: Two sequential walk_to() calls each return a valid "walk_completed" Signal.
func test_07_sequential_walk_to_returns_valid_signals() -> void:
	var parts: Array = _make_room_with_character_and_points()
	var room: AGSRoom = parts[0]
	var ch: AGSCharacterBase = parts[1]

	var sig_a: Signal = ch.walk_to("door")
	var sig_b: Signal = ch.walk_to("window")

	assert_true(sig_a.get_name() == "walk_completed",
		"First walk_to() signal has wrong name: '%s'" % sig_a.get_name())
	assert_true(sig_b.get_name() == "walk_completed",
		"Second walk_to() signal has wrong name: '%s'" % sig_b.get_name())
	assert_true(sig_a.get_object() == ch, "Signal is not on the character")

	room.free()

# UT-M5-08: face_to() returns a Signal named "face_completed".
func test_08_face_to_returns_face_completed_signal() -> void:
	var parts: Array = _make_room_with_character_and_points()
	var room: AGSRoom = parts[0]
	var ch: AGSCharacterBase = parts[1]

	var sig: Signal = ch.face_to("window")
	assert_true(sig.get_name() == "face_completed",
		"face_to() returned signal with wrong name: '%s'" % sig.get_name())
	assert_true(sig.get_object() == ch, "Signal is not on the character")

	room.free()

# UT-M5-09: face_to() applies rotation immediately in headless mode
# (is_inside_tree() == false → instant rotation, no tween).
func test_09_face_to_applies_rotation_headless() -> void:
	var parts: Array = _make_room_with_character_and_points()
	var room: AGSRoom = parts[0]
	var ch: AGSCharacterBase = parts[1]

	var initial_y: float = ch.rotation.y
	ch.face_to("window")  # window is at (-4, 0, 0) — requires non-zero rotation

	assert_true(not is_equal_approx(ch.rotation.y, initial_y),
		"face_to() did not rotate the character in headless mode")

	room.free()
