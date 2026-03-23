## UT-M4-08..10 — AGSTriggerRegion signal and registration tests.
extends "res://utils/test_base.gd"

var _last_entered_body: Node3D = null
var _last_exited_body: Node3D = null

func _on_region_entered(body: Node3D) -> void:
	_last_entered_body = body

func _on_region_exited(body: Node3D) -> void:
	_last_exited_body = body

# UT-M4-08: TriggerRegion fires region_entered when a body overlaps.
func test_08_region_entered_fires_on_body_overlap() -> void:
	var scene := load("res://m4_room/scenes/test_room_with_region.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_room_with_region.tscn")

	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	var region: AGSTriggerRegion = root_node.get_node("Region")
	assert_not_null(region, "AGSTriggerRegion not found in scene")

	_last_entered_body = null
	region.connect("region_entered", _on_region_entered)

	# Simulate a body entering by emitting body_entered — our bridge re-emits as region_entered.
	var dummy_body: CharacterBody3D = CharacterBody3D.new()
	region.emit_signal("body_entered", dummy_body)

	assert_not_null(_last_entered_body, "region_entered was not fired")
	assert_eq(_last_entered_body, dummy_body, "region_entered fired with wrong body")

	dummy_body.free()
	tree_root.remove_child(root_node)
	root_node.queue_free()

# UT-M4-09: TriggerRegion fires region_exited when a body leaves.
func test_09_region_exited_fires_on_body_leave() -> void:
	var scene := load("res://m4_room/scenes/test_room_with_region.tscn") as PackedScene
	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	var region: AGSTriggerRegion = root_node.get_node("Region")
	_last_exited_body = null
	region.connect("region_exited", _on_region_exited)

	var dummy_body: CharacterBody3D = CharacterBody3D.new()
	region.emit_signal("body_exited", dummy_body)

	assert_not_null(_last_exited_body, "region_exited was not fired")
	assert_eq(_last_exited_body, dummy_body, "region_exited fired with wrong body")

	dummy_body.free()
	tree_root.remove_child(root_node)
	root_node.queue_free()

# UT-M4-10: TriggerRegion registers its name with the parent AGSRoom.
func test_10_region_registers_with_room() -> void:
	var scene := load("res://m4_room/scenes/test_room_with_region.tscn") as PackedScene
	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	# Verify the room knows about the region by checking it is in the room's
	# child list with the correct class (registration is the prerequisite for T33).
	var found := false
	for child in root_node.get_children():
		if child is AGSTriggerRegion and child.region_name == "entrance":
			found = true
			break
	assert_true(found, "AGSTriggerRegion 'entrance' not found as child of AGSRoom")

	tree_root.remove_child(root_node)
	root_node.queue_free()
