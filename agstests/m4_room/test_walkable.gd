## UT-M4-03 — AGSWalkableSurface bakes a valid navmesh on scene load.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M4: WalkableSurface"

# UT-M4-03: WalkableSurface creates a NavigationRegion3D child and sets a
# NavigationMesh on it when NOTIFICATION_READY fires.
func test_03_walkable_surface_bakes_navmesh() -> void:
	var scene := load("res://m4_room/scenes/test_room_walkable.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_room_walkable.tscn")

	var root_node: Node = scene.instantiate()
	var surface: AGSWalkableSurface = root_node.get_node("WalkableSurface")
	assert_not_null(surface, "WalkableSurface node not found in scene")

	# Fire _notification(READY) directly — no scene tree needed.
	surface.notification(Node.NOTIFICATION_READY)

	var nav_region: NavigationRegion3D = null
	for child in surface.get_children():
		if child is NavigationRegion3D:
			nav_region = child
			break
	assert_not_null(nav_region, "No NavigationRegion3D child found on WalkableSurface after ready")
	assert_not_null(nav_region.navigation_mesh, "NavigationRegion3D has no NavigationMesh set")

	root_node.free()
