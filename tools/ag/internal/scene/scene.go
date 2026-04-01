// Package scene serialises RoomData into Godot .tscn scene files.
//
// The generated scene contains AGSRoom as the root node with AGSCamera,
// AGSPoint, AGSWalkableSurface, AGSBlockerVolume, AGSSpawnPoint, and
// AGSHotspot child nodes derived from the .agroom source.
//
// All UIDs and unique_id values are deterministically derived from the
// node path so that rebuilds produce identical output for identical input.
package scene

import (
	"fmt"
	"hash/fnv"
	"math"
	"strings"
	"unicode"

	"github.com/ags3d/ag/internal/room"
)

// GenerateRoomScene returns the full text of a .tscn file for rd.
//
// scriptRelPath is the path to the .agscript file relative to the project
// root, e.g. "rooms/start/start.agscript". The generated scene references
// the corresponding compiled GDScript at
// res://.engine/generated/<scriptRelPath>.gd.
func GenerateRoomScene(rd *room.RoomData, scriptRelPath string) string {
	g := &generator{}
	return g.roomScene(rd, scriptRelPath)
}

// --------------------------------------------------------------------------
// generator
// --------------------------------------------------------------------------

type generator struct {
	subRes    strings.Builder
	nodes     strings.Builder
	target    *strings.Builder // current write target for prop()
	subResIDs map[string]bool  // tracks used sub-resource IDs for uniqueness
}

func (g *generator) init() {
	g.subResIDs = make(map[string]bool)
	g.target = &g.nodes // default write target
}

// roomScene assembles the complete .tscn text.
func (g *generator) roomScene(rd *room.RoomData, scriptRelPath string) string {
	g.init()

	scriptGDPath := "res://.engine/generated/" + scriptRelPath + ".gd"

	// Collect sub-resources needed, then write all sections in order.

	// Walkable surface sub-resources (one shared material, per-surface mesh+shape).
	hasWalkable := len(rd.WalkableSurfaces) > 0
	if hasWalkable {
		g.subResource("StandardMaterial3D", "WalkMat", func() {
			g.prop("transparency", "1")
			g.prop("shading_mode", "0")
			g.prop("albedo_color", "Color(0, 0.8, 0.2, 0.35)")
		})
	}
	for _, ws := range rd.WalkableSurfaces {
		slug := slugify(ws.Name)
		meshID := "BoxMesh_" + slug
		shapeID := "BoxShape3D_" + slug
		g.subResource("BoxMesh", meshID, func() {
			g.prop("size", vec3Str(ws.Size.X, 0.1, ws.Size.Z))
		})
		g.subResource("BoxShape3D", shapeID, func() {
			g.prop("size", vec3Str(ws.Size.X, 0.1, ws.Size.Z))
		})
	}

	// BlockerVolume sub-resources (one shape per blocker).
	for _, bv := range rd.BlockerVolumes {
		slug := slugify(bv.Name)
		shapeID := "BoxShape3D_" + slug
		g.subResource("BoxShape3D", shapeID, func() {
			g.prop("size", vec3Str(bv.Size.X, bv.Size.Y, bv.Size.Z))
		})
	}

	// Hotspot sub-resources (one shape per hotspot).
	for _, hs := range rd.Hotspots {
		slug := slugify(hs.Name)
		shapeID := "BoxShape3D_hs_" + slug
		g.subResource("BoxShape3D", shapeID, func() {
			g.prop("size", vec3Str(hs.Size.X, hs.Size.Y, hs.Size.Z))
		})
	}

	// --- Nodes ---

	// Root: AGSRoom
	rootName := toPascalCase(rd.Name)
	g.node(rootName, "AGSRoom", "", func() {
		g.prop("room_name", strLit(rd.Name))
		if rd.InitialCamera != "" {
			g.prop("initial_camera", strLit(rd.InitialCamera))
		}
		g.prop("script", `ExtResource("RoomScript")`)
	})

	// WalkableSurfaces
	for _, ws := range rd.WalkableSurfaces {
		slug := slugify(ws.Name)
		nodeName := toPascalCase(ws.Name)
		g.node(nodeName, "AGSWalkableSurface", ".", func() {
			if ws.Offset != (room.Vec3{}) {
				g.prop("transform", identTransformAt(ws.Offset.X, ws.Offset.Y, ws.Offset.Z))
			}
		})
		meshID := "BoxMesh_" + slug
		shapeID := "BoxShape3D_" + slug
		g.node("MeshInstance3D", "MeshInstance3D", nodeName, func() {
			g.prop("material_overlay", subResRef("WalkMat"))
			g.prop("mesh", subResRef(meshID))
		})
		g.node("CollisionShape3D", "CollisionShape3D", nodeName, func() {
			g.prop("shape", subResRef(shapeID))
		})
	}

	// Points
	for _, pt := range rd.Points {
		nodeName := toPascalCase(pt.Name)
		g.node(nodeName, "AGSPoint", ".", func() {
			g.prop("transform", identTransformAt(pt.Position.X, pt.Position.Y, pt.Position.Z))
			g.prop("point_name", strLit(pt.Name))
		})
	}

	// BlockerVolumes
	for _, bv := range rd.BlockerVolumes {
		slug := slugify(bv.Name)
		nodeName := toPascalCase(bv.Name)
		shapeID := "BoxShape3D_" + slug
		g.node(nodeName, "AGSBlockerVolume", ".", func() {
			if bv.Position != (room.Vec3{}) {
				g.prop("transform", identTransformAt(bv.Position.X, bv.Position.Y, bv.Position.Z))
			}
		})
		g.node("CollisionShape3D", "CollisionShape3D", nodeName, func() {
			g.prop("shape", subResRef(shapeID))
		})
	}

	// SpawnPoints
	for _, sp := range rd.SpawnPoints {
		nodeName := toPascalCase(sp.Name)
		g.node(nodeName, "AGSSpawnPoint", ".", func() {
			g.prop("transform", identTransformAt(sp.Position.X, sp.Position.Y, sp.Position.Z))
			g.prop("spawn_character", strLit(sp.Character))
		})
	}

	// Hotspots
	for _, hs := range rd.Hotspots {
		slug := slugify(hs.Name)
		nodeName := toPascalCase(hs.Name)
		shapeID := "BoxShape3D_hs_" + slug
		g.node(nodeName, "AGSHotspot", ".", func() {
			if hs.Position != (room.Vec3{}) {
				g.prop("transform", identTransformAt(hs.Position.X, hs.Position.Y, hs.Position.Z))
			}
			g.prop("hotspot_name", strLit(hs.Name))
		})
		g.node("CollisionShape3D", "CollisionShape3D", nodeName, func() {
			g.prop("shape", subResRef(shapeID))
		})
	}

	// Cameras (last, after gameplay nodes)
	autoPos, autoLookAt := autoCamera(rd)
	for _, cam := range rd.Cameras {
		nodeName := toPascalCase(cam.Name)
		pos := cam.Position
		if !cam.HasPosition {
			pos = room.Vec3{X: autoPos.x, Y: autoPos.y, Z: autoPos.z}
		}
		lookAt := cam.LookAt
		if !cam.HasLookAt {
			lookAt = room.Vec3{X: autoLookAt.x, Y: autoLookAt.y, Z: autoLookAt.z}
		}
		tf := lookAtTransform(pos, lookAt)
		g.node(nodeName, "AGSCamera", ".", func() {
			g.prop("transform", tf)
			g.prop("camera_name", strLit(cam.Name))
		})
	}

	// --- Assemble ---
	var out strings.Builder
	fmt.Fprintln(&out, "[gd_scene format=3]")
	fmt.Fprintln(&out)
	fmt.Fprintf(&out, "[ext_resource type=\"Script\" path=%q id=\"RoomScript\"]\n", scriptGDPath)
	if g.subRes.Len() > 0 {
		fmt.Fprintln(&out)
		out.WriteString(g.subRes.String())
	}
	out.WriteString(g.nodes.String())
	return out.String()
}

// --------------------------------------------------------------------------
// Builders
// --------------------------------------------------------------------------

// subResource writes a [sub_resource ...] block.
func (g *generator) subResource(typ, id string, body func()) {
	fmt.Fprintf(&g.subRes, "[sub_resource type=%q id=%q]\n", typ, id)
	prev := g.target
	g.target = &g.subRes
	body()
	g.target = prev
	fmt.Fprintln(&g.subRes)
}

// node writes a [node ...] block.
func (g *generator) node(name, typ, parent string, body func()) {
	uid := nodeUID(parent + "/" + name)
	if parent == "" {
		fmt.Fprintf(&g.nodes, "[node name=%q type=%q unique_id=%d]\n", name, typ, uid)
	} else {
		fmt.Fprintf(&g.nodes, "[node name=%q type=%q parent=%q unique_id=%d]\n", name, typ, parent, uid)
	}
	prev := g.target
	g.target = &g.nodes
	body()
	g.target = prev
	fmt.Fprintln(&g.nodes)
}

// prop writes a key = value property line to the current target.
func (g *generator) prop(key, value string) {
	fmt.Fprintf(g.target, "%s = %s\n", key, value)
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// nodeUID returns a deterministic 32-bit int from a node path string.
func nodeUID(path string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(path))
	return h.Sum32()
}

// slugify converts a name to a safe identifier (underscores preserved).
func slugify(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "_")
}

// toPascalCase converts snake_case or lowercase to PascalCase.
// "door_left" → "DoorLeft", "floor" → "Floor", "player_start" → "PlayerStart"
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		b.WriteRune(unicode.ToUpper(runes[0]))
		b.WriteString(string(runes[1:]))
	}
	return b.String()
}

// strLit wraps s in double quotes for a GDScript string literal.
func strLit(s string) string {
	return fmt.Sprintf("%q", s)
}

// subResRef returns SubResource("id").
func subResRef(id string) string {
	return fmt.Sprintf("SubResource(%q)", id)
}

// fmtF formats a float for .tscn output: integers as integers, others as %g.
// Rounds to 6 decimal places first to eliminate floating-point noise from
// computed values (e.g. 11.200000000000001 → 11.2).
func fmtF(f float64) string {
	f = math.Round(f*1e6) / 1e6
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}

// vec3Str formats a Vector3 for .tscn.
func vec3Str(x, y, z float64) string {
	return fmt.Sprintf("Vector3(%s, %s, %s)", fmtF(x), fmtF(y), fmtF(z))
}

// identTransformAt returns a Transform3D with identity rotation at position (x,y,z).
func identTransformAt(x, y, z float64) string {
	return fmt.Sprintf("Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, %s, %s, %s)",
		fmtF(x), fmtF(y), fmtF(z))
}

// --------------------------------------------------------------------------
// Look-at transform
// --------------------------------------------------------------------------

type vec3 struct{ x, y, z float64 }

func (v vec3) sub(o vec3) vec3    { return vec3{v.x - o.x, v.y - o.y, v.z - o.z} }
func (v vec3) length() float64    { return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z) }
func (v vec3) normalize() vec3 {
	l := v.length()
	if l == 0 {
		return vec3{0, 0, 1}
	}
	return vec3{v.x / l, v.y / l, v.z / l}
}
func cross(a, b vec3) vec3 {
	return vec3{
		a.y*b.z - a.z*b.y,
		a.z*b.x - a.x*b.z,
		a.x*b.y - a.y*b.x,
	}
}

// autoCamera derives a sensible default camera position and look-at from the
// room's walkable floor. The camera is placed above and slightly in front of
// the floor center so the whole floor is visible at a gentle downward angle.
//
//	position  = floor_center + (0, maxSize*0.8, maxSize*0.55)
//	look_at   = floor_center (at Y=0)
//
// Falls back to a 10-unit default when no WalkableSurface is defined.
func autoCamera(rd *room.RoomData) (pos vec3, lookAt vec3) {
	var cx, cz float64
	maxSize := 10.0
	if len(rd.WalkableSurfaces) > 0 {
		ws := rd.WalkableSurfaces[0]
		cx = ws.Offset.X
		cz = ws.Offset.Z
		if ws.Size.X > ws.Size.Z {
			maxSize = ws.Size.X
		} else {
			maxSize = ws.Size.Z
		}
	}
	lookAt = vec3{cx, 0, cz}
	pos = vec3{cx, maxSize * 0.8, cz + maxSize*0.55}
	return
}

// lookAtTransform returns a Transform3D string for a camera at eye looking at target.
// The camera's -Z axis points toward target; world up is (0, 1, 0).
func lookAtTransform(pos, target room.Vec3) string {
	eye := vec3{pos.X, pos.Y, pos.Z}
	tgt := vec3{target.X, target.Y, target.Z}
	worldUp := vec3{0, 1, 0}

	// back = direction from target to eye (+Z in Godot camera basis)
	back := eye.sub(tgt).normalize()

	// Handle degenerate case: eye directly above/below target
	if math.Abs(back.x) < 1e-6 && math.Abs(back.z) < 1e-6 {
		worldUp = vec3{0, 0, 1}
	}

	right := cross(worldUp, back).normalize()
	up := cross(back, right)

	// Godot Transform3D text format stores the basis ROW by ROW:
	//   Basis(xx,xy,xz, yx,yy,yz, zx,zy,zz) → rows[0]=(xx,xy,xz), rows[1]=(yx,...), rows[2]=(zx,...)
	// Each column of the stored matrix is an axis vector:
	//   col0 = right (X), col1 = up (Y), col2 = back (Z)
	// So the 9 values are written as (right.x, up.x, back.x,  right.y, up.y, back.y,  right.z, up.z, back.z).
	return fmt.Sprintf("Transform3D(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
		fmtF(right.x), fmtF(up.x), fmtF(back.x),
		fmtF(right.y), fmtF(up.y), fmtF(back.y),
		fmtF(right.z), fmtF(up.z), fmtF(back.z),
		fmtF(pos.X), fmtF(pos.Y), fmtF(pos.Z),
	)
}
