@tool
extends EditorNode3DGizmoPlugin

## Gizmo plugin for AGSSpawnPoint.
## Draws a stylised character silhouette (head circle + body lines) at origin.

const HEAD_R  := 0.12
const BODY_H  := 0.5
const SHLDR_W := 0.18
const LEG_H   := 0.28


func _init() -> void:
	create_material("line", Color(0.2, 0.9, 0.7, 1.0))


func _get_gizmo_name() -> String:
	return "AGSSpawnPoint"


func _has_gizmo(node: Node3D) -> bool:
	return node.get_class() == "AGSSpawnPoint"


func _redraw(gizmo: EditorNode3DGizmo) -> void:
	gizmo.clear()
	var mat   := get_material("line", gizmo)
	var lines := PackedVector3Array()

	# Head: 8-segment circle in the XY plane
	var head_y: float = BODY_H + HEAD_R * 1.2
	const SEGS := 8
	for i in SEGS:
		var a0 := float(i)     / SEGS * TAU
		var a1 := float(i + 1) / SEGS * TAU
		lines.append(Vector3(cos(a0) * HEAD_R, head_y + sin(a0) * HEAD_R, 0))
		lines.append(Vector3(cos(a1) * HEAD_R, head_y + sin(a1) * HEAD_R, 0))

	# Spine
	lines.append(Vector3(0, BODY_H, 0));  lines.append(Vector3(0, 0, 0))

	# Shoulders
	lines.append(Vector3(-SHLDR_W, BODY_H, 0));  lines.append(Vector3(SHLDR_W, BODY_H, 0))

	# Arms
	lines.append(Vector3(-SHLDR_W, BODY_H,     0));  lines.append(Vector3(-SHLDR_W * 0.6, BODY_H * 0.4, 0))
	lines.append(Vector3( SHLDR_W, BODY_H,     0));  lines.append(Vector3( SHLDR_W * 0.6, BODY_H * 0.4, 0))

	# Legs
	lines.append(Vector3(0, 0, 0));  lines.append(Vector3(-SHLDR_W * 0.5, -LEG_H, 0))
	lines.append(Vector3(0, 0, 0));  lines.append(Vector3( SHLDR_W * 0.5, -LEG_H, 0))

	gizmo.add_lines(lines, mat)

	# Small capsule collision mesh covers the silhouette so the spawn point
	# is click-selectable anywhere within the figure bounds.
	var capsule := CapsuleMesh.new()
	capsule.radius = SHLDR_W
	capsule.height = BODY_H + HEAD_R * 2.0 + LEG_H
	gizmo.add_collision_triangles(capsule.generate_triangle_mesh())
