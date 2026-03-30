## UT-M10-01..02 — AGSRuntime.load_room() / room_change_requested signal (T-GS10)
##
## load_room() is pure signal emission — actual scene swapping is handled by
## ags_room_manager.gd in the game project, so these tests only verify the C++
## side: that the signal fires and carries the correct room name.
##
## NOTE: lambdas capture primitives (String, bool, int) by value in GDScript 4.
## All captured signal payloads must use Array wrappers so the lambda can mutate
## the outer variable. This is the same pattern as test_end_to_end.gd.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: LoadRoom"


# UT-M10-01: load_room() emits room_change_requested with the exact room name supplied.
func test_01_load_room_emits_correct_name() -> void:
	var received := [""]
	AGSRuntime.connect("room_change_requested", func(name: String) -> void: received[0] = name, CONNECT_ONE_SHOT)
	AGSRuntime.load_room("test_lobby")
	assert_eq(received[0], "test_lobby", "room_change_requested did not carry the correct room name")


# UT-M10-02: load_room() with a different name carries that name unchanged.
func test_02_load_room_different_names_are_distinct() -> void:
	var received_a := [""]
	var received_b := [""]

	AGSRuntime.connect("room_change_requested", func(n: String) -> void: received_a[0] = n, CONNECT_ONE_SHOT)
	AGSRuntime.load_room("library")
	assert_eq(received_a[0], "library", "first call: expected 'library', got '%s'" % received_a[0])

	AGSRuntime.connect("room_change_requested", func(n: String) -> void: received_b[0] = n, CONNECT_ONE_SHOT)
	AGSRuntime.load_room("dungeon")
	assert_eq(received_b[0], "dungeon", "second call: expected 'dungeon', got '%s'" % received_b[0])

	assert_ne(received_a[0], received_b[0], "two distinct load_room calls produced the same received name")


# UT-M10-03: room_change_requested signal exists on AGSRuntime.
func test_03_room_change_requested_signal_exists() -> void:
	assert_true(
		AGSRuntime.has_signal("room_change_requested"),
		"AGSRuntime does not have a room_change_requested signal"
	)
