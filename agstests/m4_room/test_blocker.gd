## UT-M4-04 — AGSBlockerVolume is invisible at runtime and has collision.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M4: BlockerVolume"

# UT-M4-04: BlockerVolume hides its MeshInstance3D children at runtime and
# keeps its CollisionShape3D active so navmesh baking excludes the volume.
func test_04_blocker_volume_is_invisible_with_collision() -> void:
	var scene := load("res://m4_room/scenes/test_room_with_blocker.tscn") as PackedScene
	assert_not_null(scene, "Could not load test_room_with_blocker.tscn")

	var root_node: Node = scene.instantiate()
	var blocker: AGSBlockerVolume = root_node.get_node("BlockerVolume")
	assert_not_null(blocker, "BlockerVolume node not found in scene")

	# Fire _notification(READY) directly — hides mesh, keeps collision active.
	blocker.notification(Node.NOTIFICATION_READY)

	var mesh_hidden := false
	var collision_present := false
	for child in blocker.get_children():
		if child is MeshInstance3D:
			mesh_hidden = not child.visible
		if child is CollisionShape3D:
			collision_present = true

	assert_true(mesh_hidden, "BlockerVolume MeshInstance3D is still visible at runtime")
	assert_true(collision_present, "BlockerVolume has no CollisionShape3D child")

	root_node.free()
