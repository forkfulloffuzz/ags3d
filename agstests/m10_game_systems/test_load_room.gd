## UT-M10-01..02 — AGSRuntime.load_room() / room_change_requested signal (T-GS10)
##
## load_room() is pure signal emission — actual scene swapping is handled by
## ags_room_manager.gd in the game project, so these tests only verify the C++
## side: that the signal fires and carries the correct room name.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: LoadRoom"


# UT-M10-01: load_room() emits room_change_requested with the exact room name supplied.
func test_01_load_room_emits_correct_name() -> void:
	var received := ""
	AGSRuntime.connect("room_change_requested", func(name: String) -> void: received = name, CONNECT_ONE_SHOT)
	AGSRuntime.load_room("test_lobby")
	assert_eq(received, "test_lobby", "room_change_requested did not carry the correct room name")


# UT-M10-02: load_room() with a different name carries that name unchanged.
func test_02_load_room_different_names_are_distinct() -> void:
	var received_a := ""
	var received_b := ""

	AGSRuntime.connect("room_change_requested", func(n: String) -> void: received_a = n, CONNECT_ONE_SHOT)
	AGSRuntime.load_room("library")
	assert_eq(received_a, "library", "first call: expected 'library', got '%s'" % received_a)

	AGSRuntime.connect("room_change_requested", func(n: String) -> void: received_b = n, CONNECT_ONE_SHOT)
	AGSRuntime.load_room("dungeon")
	assert_eq(received_b, "dungeon", "second call: expected 'dungeon', got '%s'" % received_b)

	assert_ne(received_a, received_b, "two distinct load_room calls produced the same received name")


# UT-M10-03: room_change_requested signal exists on AGSRuntime.
func test_03_room_change_requested_signal_exists() -> void:
	assert_true(
		AGSRuntime.has_signal("room_change_requested"),
		"AGSRuntime does not have a room_change_requested signal"
	)
