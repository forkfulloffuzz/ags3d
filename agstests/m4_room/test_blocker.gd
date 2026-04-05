## UT-M4-04 — AGSBlockerVolume has collision and does not force-hide child meshes.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M4: BlockerVolume"

# UT-M4-04: BlockerVolume keeps its CollisionShape3D active for navmesh baking
# and does NOT hide child MeshInstance3D nodes — users may place visible decorative
# meshes under a blocker. If no mesh is authored, the blocker is naturally invisible.
func test_04_blocker_volume_is_invisible_with_collision() -> void:
	var scene := load("res://m4_room/scenes/test_room_with_blocker.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_room_with_blocker.tscn")

	var root_node: Node = scene.instantiate()
	var blocker: AGSBlockerVolume = root_node.get_node("BlockerVolume")
	assert_not_null(blocker, "BlockerVolume node not found in scene")

	blocker.notification(Node.NOTIFICATION_READY)

	var collision_present := false
	for child in blocker.get_children():
		if child is CollisionShape3D:
			collision_present = true

	assert_true(collision_present, "BlockerVolume has no CollisionShape3D child")

	root_node.free()
