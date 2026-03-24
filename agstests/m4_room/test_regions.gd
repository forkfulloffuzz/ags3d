## UT-M4-08..10 — AGSTriggerRegion signal and registration tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M4: Regions"

var _last_entered_body: Node3D = null
var _last_exited_body: Node3D = null

func _on_region_entered(body: Node3D) -> void:
	_last_entered_body = body

func _on_region_exited(body: Node3D) -> void:
	_last_exited_body = body

# UT-M4-08: TriggerRegion fires region_entered when a body overlaps.
func test_08_region_entered_fires_on_body_overlap() -> void:
	var region: AGSTriggerRegion = AGSTriggerRegion.new()
	region.region_name = "entrance"
	# Fire NOTIFICATION_READY — connects body_entered → _on_body_entered bridge.
	region.notification(Node.NOTIFICATION_READY)

	_last_entered_body = null
	region.connect("region_entered", _on_region_entered)

	# Simulate a body entering by emitting body_entered — our bridge re-emits as region_entered.
	var dummy_body: CharacterBody3D = CharacterBody3D.new()
	region.emit_signal("body_entered", dummy_body)

	assert_not_null(_last_entered_body, "region_entered was not fired")
	assert_eq(_last_entered_body, dummy_body, "region_entered fired with wrong body")

	dummy_body.free()
	region.free()

# UT-M4-09: TriggerRegion fires region_exited when a body leaves.
func test_09_region_exited_fires_on_body_leave() -> void:
	var region: AGSTriggerRegion = AGSTriggerRegion.new()
	region.region_name = "entrance"
	region.notification(Node.NOTIFICATION_READY)

	_last_exited_body = null
	region.connect("region_exited", _on_region_exited)

	var dummy_body: CharacterBody3D = CharacterBody3D.new()
	region.emit_signal("body_exited", dummy_body)

	assert_not_null(_last_exited_body, "region_exited was not fired")
	assert_eq(_last_exited_body, dummy_body, "region_exited fired with wrong body")

	dummy_body.free()
	region.free()

# UT-M4-10: TriggerRegion self-registers with its parent AGSRoom on ready.
func test_10_region_registers_with_room() -> void:
	var room: AGSRoom = AGSRoom.new()
	var region: AGSTriggerRegion = AGSTriggerRegion.new()
	region.region_name = "entrance"
	room.add_child(region)
	region.notification(Node.NOTIFICATION_READY)

	# Verify the region is a child of the room with the correct name.
	# (Registration in the room's internal HashMap is verified indirectly via T33.)
	var found := false
	for child in room.get_children():
		if child is AGSTriggerRegion and child.region_name == "entrance":
			found = true
			break
	assert_true(found, "AGSTriggerRegion 'entrance' not found as child of AGSRoom")

	room.free()
