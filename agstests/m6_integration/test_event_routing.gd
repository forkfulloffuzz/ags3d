## UT-M6-06..09 — Event handler routing and source map integration tests.
##
## UT-M6-06: hotspot_interact() receives the correct hotspot name when hotspot is clicked.
## UT-M6-07: region_walked_into() handler fires when AGSRoom emits region_walked_into.
## UT-M6-08: AGSRuntime.translate_script_error() returns the .agscript file path, not the .gd path.
## UT-M6-09: AGSRuntime.translate_script_error() returns the correct AGS-spirit line number.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M6: EventRouting"


# UT-M6-06: When hotspot_clicked("name") fires on AGSRoom, the attached script's
# hotspot_interact(name) handler is called with the correct hotspot name.
func test_06_hotspot_interact_receives_correct_name() -> void:
	var root := Node.new()
	add_to_tree(root)

	var script := GDScript.new()
	script.source_code = """
extends AGSRoom
var last_hotspot := ""
func hotspot_interact(name: String) -> void:
	last_hotspot = name
"""
	assert_eq(script.reload(), OK, "Inline script should compile")

	var room := AGSRoom.new()
	room.room_name = "hotspot_routing_test"
	room.set_script(script)
	root.add_child(room)

	room.emit_signal("hotspot_clicked", "ancient_tome")
	assert_eq(
		room.get("last_hotspot"),
		"ancient_tome",
		"hotspot_interact did not receive the correct hotspot name"
	)

	root.free()


# UT-M6-07: When a body enters an AGSTriggerRegion child, the room's attached
# script receives region_walked_into(region_name: String) with the correct name.
func test_07_region_walked_into_handler_fires() -> void:
	var root := Node.new()
	add_to_tree(root)

	var script := GDScript.new()
	script.source_code = """
extends AGSRoom
var last_region := ""
func region_walked_into(region_name: String) -> void:
	last_region = region_name
"""
	assert_eq(script.reload(), OK, "Inline script should compile")

	# Build room with region child BEFORE adding to tree so NOTIFICATION_READY
	# connects the region_entered signal to region_walked_into.
	var region := AGSTriggerRegion.new()
	region.region_name = "exit_zone"

	var room := AGSRoom.new()
	room.room_name = "region_routing_test"
	room.set_script(script)
	room.add_child(region)
	# Add to tree WITHOUT manually calling notification(READY) on region — tree
	# propagates READY to children before parents, so AGSTriggerRegion sets up its
	# body_entered bridge before AGSRoom connects region_entered to the script handler.
	root.add_child(room)  # READY fires on region (body_entered bridge) then on room

	# Simulate a body entering the region.
	var dummy_body := CharacterBody3D.new()
	region.emit_signal("region_entered", dummy_body)

	assert_eq(
		room.get("last_region"),
		"exit_zone",
		"region_walked_into handler did not receive the correct region name"
	)

	dummy_body.free()
	root.free()


# UT-M6-08: translate_script_error() returns the .agscript file path, not the
# .gd path, proving error messages surface the AGS-spirit source location.
func test_08_translate_returns_agscript_path() -> void:
	var gd_path := "res://m6_integration/scripts/test_room_enter.agscript.gd"
	AGSRuntime.register_source_map(gd_path, [
		[2, "m6_integration/scripts/test_room_enter.agscript", 1],
	])

	var loc: Dictionary = AGSRuntime.translate_script_error(gd_path, 2)
	assert_false(loc.is_empty(), "translate_script_error returned empty dict")

	var file_path: String = loc.get("file", "")
	assert_true(
		file_path.ends_with(".agscript"),
		"Translated path should end with .agscript, got: " + file_path
	)
	assert_false(
		file_path.ends_with(".gd"),
		"Translated path must not be the .gd path, got: " + file_path
	)


# UT-M6-09: translate_script_error() returns the correct AGS-spirit line number
# that corresponds to the failing GDScript line.
func test_09_translate_returns_correct_agscript_line() -> void:
	var gd_path := "res://m6_integration/scripts/test_walkto.agscript.gd"
	# test_walkto.agscript.gd line 3 = "await ...walk_to(...)"
	# which was generated from test_walkto.agscript line 2.
	AGSRuntime.register_source_map(gd_path, [
		[1, "m6_integration/scripts/test_walkto.agscript", 1],
		[3, "m6_integration/scripts/test_walkto.agscript", 2],
		[4, "m6_integration/scripts/test_walkto.agscript", 3],
	])

	var loc: Dictionary = AGSRuntime.translate_script_error(gd_path, 3)
	assert_eq(
		loc.get("line"),
		2,
		"translate_script_error should map .gd line 3 → .agscript line 2"
	)
