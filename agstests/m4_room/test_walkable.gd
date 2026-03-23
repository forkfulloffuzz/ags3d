## UT-M4-03 — AGSWalkableSurface bakes a valid navmesh on scene load.
extends "res://utils/test_base.gd"

# UT-M4-03: WalkableSurface creates a NavigationRegion3D child and bakes a
# valid NavigationMesh when added to the scene tree.
func test_03_walkable_surface_bakes_navmesh() -> void:
	var scene := load("res://m4_room/scenes/test_room_walkable.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_room_walkable.tscn")

	var root_node: Node = scene.instantiate()
	var tree_root: Window = Engine.get_main_loop().root
	tree_root.add_child(root_node)

	# The WalkableSurface is a direct child of the room.
	var surface: AGSWalkableSurface = root_node.get_node("WalkableSurface")
	assert_not_null(surface, "WalkableSurface node not found in scene")

	# After _ready(), a NavigationRegion3D must have been added as a child.
	var nav_region: NavigationRegion3D = null
	for child in surface.get_children():
		if child is NavigationRegion3D:
			nav_region = child
			break
	assert_not_null(nav_region, "No NavigationRegion3D child found on WalkableSurface")
	assert_not_null(nav_region.navigation_mesh, "NavigationRegion3D has no NavigationMesh set")

	tree_root.remove_child(root_node)
	root_node.queue_free()
