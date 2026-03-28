## UT-M6-01, UT-M6-05 — Script wiring tests.
##
## UT-M6-01: Verifies that ag build produced a .gd alongside the .agscript fixture.
## UT-M6-05: Verifies that the generated .gd, when attached to AGSRoom, fires room_enter.
extends "res://utils/test_base.gd"

const AGSCRIPT_PATH := "res://m6_integration/scripts/test_room_enter.agscript"
const GENERATED_GD  := "res://m6_integration/scripts/test_room_enter.agscript.gd"

func suite_name() -> String:
	return "M6: ScriptWiring"


# UT-M6-01: The generated .gd file exists alongside the .agscript fixture.
# This verifies the ag build pipeline ran and produced output at the expected path.
func test_01_generated_gd_exists_for_agscript() -> void:
	assert_true(
		ResourceLoader.exists(GENERATED_GD),
		"Generated .gd not found at expected path: " + GENERATED_GD
	)


# UT-M6-05: Loading the generated .gd and attaching it to an AGSRoom causes
# room_enter() to fire when the room enters the scene tree.
func test_05_generated_script_fires_room_enter() -> void:
	var script := load(GENERATED_GD) as GDScript
	assert_not_null(script, "Could not load generated .gd: " + GENERATED_GD)

	var room := AGSRoom.new()
	room.room_name = "script_wiring_test"
	room.set_script(script)

	var fired := [false]
	room.room_enter.connect(func() -> void: fired[0] = true)

	var root := Node.new()
	add_to_tree(root)
	root.add_child(room)  # READY fires automatically; AGSRoom emits room_enter

	assert_true(fired[0], "room_enter signal did not fire after attaching generated script")
	root.free()
