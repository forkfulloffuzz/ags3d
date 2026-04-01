package scene_test

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/room"
	"github.com/ags3d/ag/internal/scene"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func generate(t *testing.T, src string) string {
	t.Helper()
	rd, err := room.ParseRoom("test.agroom", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return scene.GenerateRoomScene(rd, "rooms/test/test.agscript")
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("output does not contain %q\ngot:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Errorf("output unexpectedly contains %q\ngot:\n%s", want, got)
	}
}

// --------------------------------------------------------------------------
// Header and ext_resource
// --------------------------------------------------------------------------

func TestHeader(t *testing.T) {
	out := generate(t, `Room "r" {}`)
	assertContains(t, out, "[gd_scene format=3]")
	assertContains(t, out, `[ext_resource type="Script" path="res://.engine/generated/rooms/test/test.agscript.gd" id="RoomScript"]`)
}

// --------------------------------------------------------------------------
// Root AGSRoom node
// --------------------------------------------------------------------------

func TestRoomNode(t *testing.T) {
	out := generate(t, `Room "start" { initial_camera = "main" }`)
	assertContains(t, out, `[node name="Start" type="AGSRoom"`)
	assertContains(t, out, `room_name = "start"`)
	assertContains(t, out, `initial_camera = "main"`)
	assertContains(t, out, `script = ExtResource("RoomScript")`)
}

func TestRoomNodeNoCamera(t *testing.T) {
	out := generate(t, `Room "lobby" {}`)
	assertNotContains(t, out, "initial_camera")
}

// --------------------------------------------------------------------------
// WalkableSurface
// --------------------------------------------------------------------------

func TestWalkableSurface(t *testing.T) {
	src := `Room "r" {
		WalkableSurface "floor" {
			size   = (10.0, 8.0)
			offset = (0.0, -0.05, 0.0)
		}
	}`
	out := generate(t, src)

	// Sub-resources
	assertContains(t, out, `[sub_resource type="StandardMaterial3D" id="WalkMat"]`)
	assertContains(t, out, `[sub_resource type="BoxMesh" id="BoxMesh_floor"]`)
	assertContains(t, out, `size = Vector3(10, 0.1, 8)`)
	assertContains(t, out, `[sub_resource type="BoxShape3D" id="BoxShape3D_floor"]`)

	// Node
	assertContains(t, out, `[node name="Floor" type="AGSWalkableSurface" parent="."`)
	assertContains(t, out, `Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 0, -0.05, 0)`)

	// Children
	assertContains(t, out, `[node name="MeshInstance3D" type="MeshInstance3D" parent="Floor"`)
	assertContains(t, out, `material_overlay = SubResource("WalkMat")`)
	assertContains(t, out, `mesh = SubResource("BoxMesh_floor")`)
	assertContains(t, out, `[node name="CollisionShape3D" type="CollisionShape3D" parent="Floor"`)
	assertContains(t, out, `shape = SubResource("BoxShape3D_floor")`)
}

func TestWalkableSurfaceNoOffset(t *testing.T) {
	src := `Room "r" { WalkableSurface "floor" { size = (10.0, 10.0) } }`
	out := generate(t, src)
	// When offset is zero, no transform property is emitted for the surface node
	lines := strings.Split(out, "\n")
	inFloor := false
	for _, line := range lines {
		if strings.Contains(line, `[node name="Floor"`) {
			inFloor = true
			continue
		}
		if inFloor && strings.HasPrefix(line, "[") {
			break
		}
		if inFloor && strings.Contains(line, "transform") {
			t.Errorf("expected no transform for zero-offset WalkableSurface, got: %s", line)
		}
	}
}

func TestNoWalkMatWithoutWalkableSurface(t *testing.T) {
	out := generate(t, `Room "r" {}`)
	assertNotContains(t, out, "WalkMat")
}

// --------------------------------------------------------------------------
// Point
// --------------------------------------------------------------------------

func TestPoint(t *testing.T) {
	src := `Room "r" { Point "door_left" { position = (3.12, 0.18, 3.43) } }`
	out := generate(t, src)
	assertContains(t, out, `[node name="DoorLeft" type="AGSPoint" parent="."`)
	assertContains(t, out, `Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 3.12, 0.18, 3.43)`)
	assertContains(t, out, `point_name = "door_left"`)
}

func TestPointNegativeCoords(t *testing.T) {
	src := `Room "r" { Point "exit" { position = (-4.0, 0.0, -3.0) } }`
	out := generate(t, src)
	assertContains(t, out, `Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, -4, 0, -3)`)
}

// --------------------------------------------------------------------------
// BlockerVolume
// --------------------------------------------------------------------------

func TestBlockerVolume(t *testing.T) {
	src := `Room "r" {
		BlockerVolume "center_wall" {
			size     = (1.0, 2.0, 3.0)
			position = (0.0, 1.0, 1.15)
		}
	}`
	out := generate(t, src)
	assertContains(t, out, `[sub_resource type="BoxShape3D" id="BoxShape3D_center_wall"]`)
	assertContains(t, out, `size = Vector3(1, 2, 3)`)
	assertContains(t, out, `[node name="CenterWall" type="AGSBlockerVolume" parent="."`)
	assertContains(t, out, `Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1.15)`)
	assertContains(t, out, `[node name="CollisionShape3D" type="CollisionShape3D" parent="CenterWall"`)
	assertContains(t, out, `shape = SubResource("BoxShape3D_center_wall")`)
}

// --------------------------------------------------------------------------
// SpawnPoint
// --------------------------------------------------------------------------

func TestSpawnPoint(t *testing.T) {
	src := `Room "r" {
		SpawnPoint "player_start" {
			character = "player"
			position  = (-4.0, 0.0, -3.0)
		}
	}`
	out := generate(t, src)
	assertContains(t, out, `[node name="PlayerStart" type="AGSSpawnPoint" parent="."`)
	assertContains(t, out, `Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, -4, 0, -3)`)
	assertContains(t, out, `spawn_character = "player"`)
}

// --------------------------------------------------------------------------
// Hotspot
// --------------------------------------------------------------------------

func TestHotspot(t *testing.T) {
	src := `Room "r" {
		Hotspot "bookshelf" {
			size     = (1.0, 2.0, 0.3)
			position = (2.0, 1.0, -4.8)
		}
	}`
	out := generate(t, src)
	assertContains(t, out, `[sub_resource type="BoxShape3D" id="BoxShape3D_hs_bookshelf"]`)
	assertContains(t, out, `[node name="Bookshelf" type="AGSHotspot" parent="."`)
	assertContains(t, out, `hotspot_name = "bookshelf"`)
	assertContains(t, out, `[node name="CollisionShape3D" type="CollisionShape3D" parent="Bookshelf"`)
	assertContains(t, out, `shape = SubResource("BoxShape3D_hs_bookshelf")`)
}

// --------------------------------------------------------------------------
// Camera
// --------------------------------------------------------------------------

func TestCamera(t *testing.T) {
	src := `Room "r" {
		Camera "main" {
			position = (0.0, 0.0, 5.0)
			look_at  = (0.0, 0.0, 0.0)
		}
	}`
	out := generate(t, src)
	assertContains(t, out, `[node name="Main" type="AGSCamera" parent="."`)
	assertContains(t, out, `camera_name = "main"`)
	// Camera at (0,0,5) looking at origin: right=(1,0,0), up=(0,1,0), back=(0,0,1)
	// Row-major format: (right.x,up.x,back.x, right.y,up.y,back.y, right.z,up.z,back.z, ox,oy,oz)
	// = (1,0,0, 0,1,0, 0,0,1, 0,0,5) — identity basis so rows equal columns here.
	assertContains(t, out, `Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 5)`)
}

func TestCameraRowMajorNonTrivial(t *testing.T) {
	// Camera at (0,8,5.5) looking at origin.
	// back = (0,8,5.5) normalised ≈ (0, 0.8240, 0.5665)
	// right = cross(up,back) normalised = (1, 0, 0)
	// up    = cross(back,right) = (0, 0.5665, -0.8240)
	//
	// Row-major output (what Godot expects):
	//   row0 = (right.x, up.x, back.x) = (1,        0,       0      )
	//   row1 = (right.y, up.y, back.y) = (0,        0.5665,  0.8240 )
	//   row2 = (right.z, up.z, back.z) = (0,       -0.8240,  0.5665 )
	//
	// Basis.z col = (row0[2],row1[2],row2[2]) = (0, 0.8240, 0.5665) = back → camera looks down ✓
	// Column-major (WRONG) would give back.y and back.z swapped sign.
	src := `Room "r" {
		WalkableSurface "floor" { size = (10.0, 10.0) }
		Camera "main" { position = (0.0, 8.0, 5.5)  look_at = (0.0, 0.0, 0.0) }
	}`
	out := generate(t, src)
	// Row-major correct:   row1 = (right.y, up.y, back.y) = (0, 0.5665,  0.8240) → "0.566529, 0.824042"
	// Column-major wrong:  row1 would be (right.y, right.z, ...)-equivalent → "0.566529, -0.824042"
	// The 5th and 6th args are row1[1]=up.y and row1[2]=back.y — both must be positive.
	assertContains(t, out, "0.566529, 0.824042")
	assertNotContains(t, out, "0.566529, -0.824042")
}

func TestCameraAutoPosition(t *testing.T) {
	// Camera with no position: auto-computed from 10x10 floor.
	// pos = (0, 10*0.8, 0+10*0.55) = (0, 8, 5.5)
	src := `Room "r" {
		WalkableSurface "floor" { size = (10.0, 10.0) }
		Camera "main" {}
	}`
	out := generate(t, src)
	assertContains(t, out, `[node name="Main" type="AGSCamera"`)
	// Position (0, 8, 5.5) must appear in the transform.
	assertContains(t, out, "0, 8, 5.5)")
}

func TestCameraExplicitPositionOverridesAuto(t *testing.T) {
	// Explicit position must be used as-is.
	src := `Room "r" {
		WalkableSurface "floor" { size = (10.0, 10.0) }
		Camera "main" { position = (1.0, 2.0, 3.0) }
	}`
	out := generate(t, src)
	assertContains(t, out, "1, 2, 3)")
}

func TestCameraAutoLookAtUsesFloorCenter(t *testing.T) {
	// When look_at is omitted the camera looks at the floor center (Y=0).
	// Camera at explicit (0,10,0) with no look_at: back vector should point
	// straight down, same as explicit look_at=(0,0,0).
	src := `Room "r" {
		WalkableSurface "floor" { size = (10.0, 10.0) }
		Camera "main" { position = (0.0, 10.0, 0.0) }
	}`
	srcExplicit := `Room "r" {
		WalkableSurface "floor" { size = (10.0, 10.0) }
		Camera "main" { position = (0.0, 10.0, 0.0)  look_at = (0.0, 0.0, 0.0) }
	}`
	if generate(t, src) != generate(t, srcExplicit) {
		t.Error("auto look_at with floor center should match explicit look_at=(0,0,0)")
	}
}

func TestCameraLookAtNonTrivial(t *testing.T) {
	// Camera at (0,5,0) looking straight down at origin — degenerate case (back = up)
	// Should not panic; uses fallback right vector.
	src := `Room "r" {
		Camera "top" {
			position = (0.0, 5.0, 0.0)
			look_at  = (0.0, 0.0, 0.0)
		}
	}`
	out := generate(t, src)
	assertContains(t, out, `[node name="Top" type="AGSCamera" parent="."`)
}

// --------------------------------------------------------------------------
// Determinism
// --------------------------------------------------------------------------

func TestDeterministicOutput(t *testing.T) {
	src := `Room "start" {
		initial_camera = "main"
		Camera "main" { position = (4.79, 5.52, 5.60)  look_at = (0.0, 0.0, 0.0) }
		Point "door_left" { position = (3.12, 0.18, 3.43) }
		WalkableSurface "floor" { size = (10.0, 10.0)  offset = (0.0, -0.05, 0.0) }
		BlockerVolume "center_wall" { size = (1.0, 2.0, 3.0)  position = (0.0, 1.0, 1.15) }
		SpawnPoint "player_start" { character = "player"  position = (-4.0, 0.0, -3.0) }
	}`
	out1 := generate(t, src)
	out2 := generate(t, src)
	if out1 != out2 {
		t.Error("output is not deterministic")
	}
}

// --------------------------------------------------------------------------
// Node ordering
// --------------------------------------------------------------------------

func TestNodeOrder(t *testing.T) {
	// WalkableSurfaces before Points before Blockers before SpawnPoints before Cameras
	src := `Room "r" {
		Camera "cam" { position = (1.0, 1.0, 1.0)  look_at = (0.0, 0.0, 0.0) }
		SpawnPoint "sp" { character = "p"  position = (0.0, 0.0, 0.0) }
		Point "pt" { position = (1.0, 0.0, 0.0) }
		WalkableSurface "floor" { size = (5.0, 5.0) }
	}`
	out := generate(t, src)
	floorPos := strings.Index(out, "AGSWalkableSurface")
	ptPos := strings.Index(out, "AGSPoint")
	spPos := strings.Index(out, "AGSSpawnPoint")
	camPos := strings.Index(out, "AGSCamera")
	if floorPos > ptPos {
		t.Error("WalkableSurface should appear before Point")
	}
	if ptPos > spPos {
		t.Error("Point should appear before SpawnPoint")
	}
	if spPos > camPos {
		t.Error("SpawnPoint should appear before Camera")
	}
}

// --------------------------------------------------------------------------
// unique_id determinism
// --------------------------------------------------------------------------

func TestUniqueIDsDeterministic(t *testing.T) {
	src := `Room "start" {
		Point "a" { position = (1.0, 0.0, 0.0) }
		Point "b" { position = (2.0, 0.0, 0.0) }
	}`
	out := generate(t, src)
	// The same node paths must always produce the same unique_id.
	// Run twice and compare.
	out2 := generate(t, src)
	if out != out2 {
		t.Error("unique_ids are not deterministic across runs")
	}
}

// --------------------------------------------------------------------------
// PascalCase node naming
// --------------------------------------------------------------------------

func TestPascalCaseNames(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{`Room "r" { Point "door_left" { position = (0.0, 0.0, 0.0) } }`, `name="DoorLeft"`},
		{`Room "r" { Point "window" { position = (0.0, 0.0, 0.0) } }`, `name="Window"`},
		{`Room "r" { Point "player_spawn_point" { position = (0.0, 0.0, 0.0) } }`, `name="PlayerSpawnPoint"`},
	}
	for _, c := range cases {
		out := generate(t, c.src)
		assertContains(t, out, c.want)
	}
}
