## UT-M4-11..12 — AGSHotspot signal and multi-hotspot disambiguation tests.
extends "res://utils/test_base.gd"

var _clicked_name: String = ""

func _on_hotspot_clicked(hotspot_name: String) -> void:
	_clicked_name = hotspot_name

# UT-M4-11: HotspotSurface fires hotspot_clicked with the correct name.
func test_11_hotspot_clicked_fires_with_correct_name() -> void:
	var scene := load("res://m4_room/scenes/test_room_two_hotspots.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_room_two_hotspots.tscn")

	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	var hotspot_a: AGSHotspot = root_node.get_node("HotspotA")
	assert_not_null(hotspot_a, "HotspotA not found in scene")

	_clicked_name = ""
	hotspot_a.connect("hotspot_clicked", _on_hotspot_clicked)
	hotspot_a.simulate_click()

	assert_eq(_clicked_name, "chest", "hotspot_clicked fired with wrong name for HotspotA")

	tree_root.remove_child(root_node)
	root_node.queue_free()

# UT-M4-12: Two hotspots in the same room — simulating hotspot_b fires hotspot_b, not hotspot_a.
func test_12_two_hotspots_do_not_cross_fire() -> void:
	var scene := load("res://m4_room/scenes/test_room_two_hotspots.tscn") as PackedScene
	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	var hotspot_b: AGSHotspot = root_node.get_node("HotspotB")
	assert_not_null(hotspot_b, "HotspotB not found in scene")

	_clicked_name = ""
	hotspot_b.connect("hotspot_clicked", _on_hotspot_clicked)
	hotspot_b.simulate_click()

	assert_eq(_clicked_name, "door", "hotspot_b fired wrong name — expected 'door'")
	assert_ne(_clicked_name, "chest", "hotspot_b incorrectly fired hotspot_a's name")

	tree_root.remove_child(root_node)
	root_node.queue_free()
