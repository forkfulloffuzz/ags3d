## T33 — Room script event binding tests.
##
## Verifies that AGSRoom auto-connects its signals to AGS-spirit event handler
## functions (room_enter, hotspot_interact, region_walked_into/off) on READY.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M6: EventBinding"


# T33-01: AGSRoom declares a room_enter signal.
func test_01_room_enter_signal_declared() -> void:
	var room := AGSRoom.new()
	assert_true(room.has_signal("room_enter"), "AGSRoom must declare a room_enter signal")
	room.free()


# T33-02: room_enter signal fires when the room enters the scene tree.
# Use an Array for reference semantics so the lambda can update the outer value.
func test_02_room_enter_signal_fires_on_ready() -> void:
	var room := AGSRoom.new()
	room.room_name = "enter_signal_test"
	# Connect before add_to_tree so the listener is in place when READY fires.
	var fired := [false]
	room.room_enter.connect(func() -> void: fired[0] = true)
	add_to_tree(room)  # READY fires automatically here, emitting room_enter

	assert_true(fired[0], "room_enter signal should fire when room enters the scene tree")
	room.free()


# T33-03: hotspot_clicked is connected to hotspot_interact on READY when the
# attached script defines that method.
func test_03_hotspot_interact_called_when_hotspot_clicked() -> void:
	var root := Node.new()
	add_to_tree(root)

	var script := GDScript.new()
	script.source_code = """
extends AGSRoom
var last_clicked := ""
func hotspot_interact(name: String) -> void:
	last_clicked = name
"""
	var err := script.reload()
	assert_eq(err, OK, "Inline script should compile without error")

	var room := AGSRoom.new()
	room.room_name = "hotspot_bind_test"
	room.set_script(script)
	root.add_child(room)  # READY fires automatically, making the signal connection

	room.emit_signal("hotspot_clicked", "painting")
	assert_eq(room.get("last_clicked"), "painting",
			"hotspot_interact should be called with the hotspot name")

	root.free()


# T33-04: room_enter() in an attached script is called on READY.
func test_04_room_enter_handler_called_on_ready() -> void:
	var root := Node.new()
	add_to_tree(root)

	var script := GDScript.new()
	script.source_code = """
extends AGSRoom
var entered := false
func room_enter() -> void:
	entered = true
"""
	assert_eq(script.reload(), OK, "Inline script should compile")

	var room := AGSRoom.new()
	room.room_name = "room_enter_handler_test"
	room.set_script(script)
	root.add_child(room)  # READY fires automatically, connecting and emitting room_enter

	assert_true(room.get("entered"), "room_enter() should have been called on READY")
	root.free()


# T33-05: room_load() in an attached script is also called on READY.
func test_05_room_load_handler_called_on_ready() -> void:
	var root := Node.new()
	add_to_tree(root)

	var script := GDScript.new()
	script.source_code = """
extends AGSRoom
var loaded := false
func room_load() -> void:
	loaded = true
"""
	assert_eq(script.reload(), OK, "Inline script should compile")

	var room := AGSRoom.new()
	room.room_name = "room_load_handler_test"
	room.set_script(script)
	root.add_child(room)  # READY fires automatically, connecting room_load to room_enter signal

	assert_true(room.get("loaded"), "room_load() should have been called on READY")
	root.free()
