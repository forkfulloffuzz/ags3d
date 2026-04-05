## UT-M5-10..11 — AGSSpawnPoint placement and error-safety tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M5: SpawnPoint"

# UT-M5-10: SpawnPoint places the named character at its world position on load.
func test_10_spawn_point_places_character_at_position() -> void:
	var room: AGSRoom = AGSRoom.new()
	room.room_name = "spawn_test_room"
	add_to_tree(room)

	var ch := AGSCharacter3D.new()
	ch.character_name = "spawn_player"
	room.add_child(ch)
	ch.notification(Node.NOTIFICATION_READY)  # registers with AGSRuntime

	var spawn: AGSSpawnPoint = AGSSpawnPoint.new()
	spawn.spawn_character = "spawn_player"
	spawn.position = Vector3(3.0, 0.0, 2.0)
	room.add_child(spawn)
	spawn.notification(Node.NOTIFICATION_READY)  # places character

	# Check local position: room is at origin so local == global, and
	# AGSSpawnPoint falls back to set_position() when outside the scene tree.
	var char_pos: Vector3 = ch.position
	assert_true(
		char_pos.is_equal_approx(Vector3(3.0, 0.0, 2.0)),
		"Character not placed at spawn position. Got: %s" % char_pos
	)

	room.free()

# UT-M5-11: SpawnPoint with unknown character name — scene loads without crash.
func test_11_unknown_spawn_character_no_crash() -> void:
	var room: AGSRoom = AGSRoom.new()

	var spawn: AGSSpawnPoint = AGSSpawnPoint.new()
	spawn.spawn_character = "does_not_exist"
	room.add_child(spawn)

	assert_no_crash(func() -> void:
		spawn.notification(Node.NOTIFICATION_READY),
		"AGSSpawnPoint with unknown character name crashed on ready"
	)

	room.free()
