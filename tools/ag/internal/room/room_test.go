package room_test

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/room"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func mustParse(t *testing.T, src string) *room.RoomData {
	t.Helper()
	rd, err := room.ParseRoom("test.agroom", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return rd
}

func mustFail(t *testing.T, src string, wantSubstr string) {
	t.Helper()
	_, err := room.ParseRoom("test.agroom", src)
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantSubstr)
	}
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("expected error containing %q, got: %v", wantSubstr, err)
	}
}

// --------------------------------------------------------------------------
// Valid inputs
// --------------------------------------------------------------------------

func TestMinimalRoom(t *testing.T) {
	rd := mustParse(t, `Room "lobby" {}`)
	if rd.Name != "lobby" {
		t.Errorf("Name = %q, want %q", rd.Name, "lobby")
	}
	if rd.InitialCamera != "" {
		t.Errorf("InitialCamera = %q, want empty", rd.InitialCamera)
	}
}

func TestInitialCamera(t *testing.T) {
	rd := mustParse(t, `Room "r" { initial_camera = "main" }`)
	if rd.InitialCamera != "main" {
		t.Errorf("InitialCamera = %q, want %q", rd.InitialCamera, "main")
	}
}

func TestCamera(t *testing.T) {
	src := `Room "r" {
		Camera "main" {
			position = (4.79, 5.52, 5.60)
			look_at  = (0.0, 0.0, 0.0)
		}
	}`
	rd := mustParse(t, src)
	if len(rd.Cameras) != 1 {
		t.Fatalf("Cameras len = %d, want 1", len(rd.Cameras))
	}
	c := rd.Cameras[0]
	if c.Name != "main" {
		t.Errorf("Camera.Name = %q, want %q", c.Name, "main")
	}
	if c.Position.X != 4.79 || c.Position.Y != 5.52 || c.Position.Z != 5.60 {
		t.Errorf("Camera.Position = %+v", c.Position)
	}
	if c.LookAt != (room.Vec3{}) {
		t.Errorf("Camera.LookAt = %+v, want zero", c.LookAt)
	}
}

func TestPoint(t *testing.T) {
	src := `Room "r" {
		Point "door_left" { position = (3.12, 0.18, 3.43) }
	}`
	rd := mustParse(t, src)
	if len(rd.Points) != 1 {
		t.Fatalf("Points len = %d, want 1", len(rd.Points))
	}
	pt := rd.Points[0]
	if pt.Name != "door_left" {
		t.Errorf("Point.Name = %q", pt.Name)
	}
	if pt.Position.X != 3.12 || pt.Position.Y != 0.18 || pt.Position.Z != 3.43 {
		t.Errorf("Point.Position = %+v", pt.Position)
	}
}

func TestWalkableSurface(t *testing.T) {
	src := `Room "r" {
		WalkableSurface "floor" {
			size   = (10.0, 10.0)
			offset = (0.0, -0.05, 0.0)
		}
	}`
	rd := mustParse(t, src)
	if len(rd.WalkableSurfaces) != 1 {
		t.Fatalf("WalkableSurfaces len = %d, want 1", len(rd.WalkableSurfaces))
	}
	ws := rd.WalkableSurfaces[0]
	if ws.Name != "floor" {
		t.Errorf("WalkableSurface.Name = %q", ws.Name)
	}
	if ws.Size.X != 10.0 || ws.Size.Z != 10.0 {
		t.Errorf("WalkableSurface.Size = %+v", ws.Size)
	}
	if ws.Offset.Y != -0.05 {
		t.Errorf("WalkableSurface.Offset.Y = %v, want -0.05", ws.Offset.Y)
	}
}

func TestBlockerVolume(t *testing.T) {
	src := `Room "r" {
		BlockerVolume "center_wall" {
			size     = (1.0, 2.0, 3.0)
			position = (0.0, 1.0, 1.15)
		}
	}`
	rd := mustParse(t, src)
	if len(rd.BlockerVolumes) != 1 {
		t.Fatalf("BlockerVolumes len = %d, want 1", len(rd.BlockerVolumes))
	}
	bv := rd.BlockerVolumes[0]
	if bv.Size != (room.Vec3{X: 1.0, Y: 2.0, Z: 3.0}) {
		t.Errorf("BlockerVolume.Size = %+v", bv.Size)
	}
	if bv.Position.Y != 1.0 || bv.Position.Z != 1.15 {
		t.Errorf("BlockerVolume.Position = %+v", bv.Position)
	}
}

func TestSpawnPoint(t *testing.T) {
	src := `Room "r" {
		SpawnPoint "player_start" {
			character = "player"
			position  = (-4.0, 0.0, -3.0)
		}
	}`
	rd := mustParse(t, src)
	if len(rd.SpawnPoints) != 1 {
		t.Fatalf("SpawnPoints len = %d, want 1", len(rd.SpawnPoints))
	}
	sp := rd.SpawnPoints[0]
	if sp.Character != "player" {
		t.Errorf("SpawnPoint.Character = %q", sp.Character)
	}
	if sp.Position.X != -4.0 || sp.Position.Z != -3.0 {
		t.Errorf("SpawnPoint.Position = %+v", sp.Position)
	}
}

func TestHotspot(t *testing.T) {
	src := `Room "r" {
		Hotspot "bookshelf" {
			size     = (1.0, 2.0, 0.3)
			position = (2.0, 1.0, -4.8)
		}
	}`
	rd := mustParse(t, src)
	if len(rd.Hotspots) != 1 {
		t.Fatalf("Hotspots len = %d, want 1", len(rd.Hotspots))
	}
	hs := rd.Hotspots[0]
	if hs.Name != "bookshelf" {
		t.Errorf("Hotspot.Name = %q", hs.Name)
	}
	if hs.Size.Z != 0.3 {
		t.Errorf("Hotspot.Size.Z = %v, want 0.3", hs.Size.Z)
	}
}

func TestComments(t *testing.T) {
	src := `// Top comment
Room "r" {
	// inline comment
	initial_camera = "cam" // trailing comment
	Camera "cam" {
		position = (0.0, 0.0, 0.0) // camera pos
		look_at  = (0.0, 0.0, 0.0)
	}
}`
	rd := mustParse(t, src)
	if rd.InitialCamera != "cam" {
		t.Errorf("InitialCamera = %q, want %q", rd.InitialCamera, "cam")
	}
	if len(rd.Cameras) != 1 {
		t.Errorf("Cameras len = %d, want 1", len(rd.Cameras))
	}
}

func TestMultipleBlocks(t *testing.T) {
	src := `Room "r" {
		Point "a" { position = (1.0, 0.0, 0.0) }
		Point "b" { position = (2.0, 0.0, 0.0) }
		Point "c" { position = (3.0, 0.0, 0.0) }
	}`
	rd := mustParse(t, src)
	if len(rd.Points) != 3 {
		t.Errorf("Points len = %d, want 3", len(rd.Points))
	}
	if rd.Points[2].Name != "c" {
		t.Errorf("Points[2].Name = %q, want %q", rd.Points[2].Name, "c")
	}
}

// TestFullRoom parses the example from the project prototype and validates key fields.
func TestFullRoom(t *testing.T) {
	src := `Room "start" {
    initial_camera = "main"

    Camera "main" {
        position = (4.79, 5.52, 5.60)
        look_at  = (0.0, 0.0, 0.0)
    }

    Point "door_left" {
        position = (3.12, 0.18, 3.43)
    }

    Point "window" {
        position = (3.15, 0.0, -3.55)
    }

    WalkableSurface "floor" {
        size   = (10.0, 10.0)
        offset = (0.0, -0.05, 0.0)
    }

    BlockerVolume "center_wall" {
        size     = (1.0, 2.0, 3.0)
        position = (0.0, 1.0, 1.15)
    }

    SpawnPoint "player_start" {
        character = "player"
        position  = (-4.0, 0.0, -3.0)
    }
}`
	rd := mustParse(t, src)
	if rd.Name != "start" {
		t.Errorf("Name = %q", rd.Name)
	}
	if rd.InitialCamera != "main" {
		t.Errorf("InitialCamera = %q", rd.InitialCamera)
	}
	if len(rd.Cameras) != 1 {
		t.Errorf("Cameras len = %d", len(rd.Cameras))
	}
	if len(rd.Points) != 2 {
		t.Errorf("Points len = %d", len(rd.Points))
	}
	if len(rd.WalkableSurfaces) != 1 {
		t.Errorf("WalkableSurfaces len = %d", len(rd.WalkableSurfaces))
	}
	if len(rd.BlockerVolumes) != 1 {
		t.Errorf("BlockerVolumes len = %d", len(rd.BlockerVolumes))
	}
	if len(rd.SpawnPoints) != 1 {
		t.Errorf("SpawnPoints len = %d", len(rd.SpawnPoints))
	}
	if rd.SpawnPoints[0].Character != "player" {
		t.Errorf("SpawnPoint.Character = %q", rd.SpawnPoints[0].Character)
	}
}

// --------------------------------------------------------------------------
// Error cases
// --------------------------------------------------------------------------

func TestErrNotRoom(t *testing.T) {
	mustFail(t, `Scene "r" {}`, `expected 'Room'`)
}

func TestErrMissingBrace(t *testing.T) {
	mustFail(t, `Room "r" {`, `missing '}'`)
}

func TestErrUnterminatedString(t *testing.T) {
	mustFail(t, `Room "r" { initial_camera = "unclosed }`, `unterminated string`)
}

func TestErrUnknownRoomProperty(t *testing.T) {
	mustFail(t, `Room "r" { unknown_key = "v" }`, `unknown Room property`)
}

func TestErrUnknownBlockType(t *testing.T) {
	mustFail(t, `Room "r" { Lamp "x" {} }`, `unknown block type`)
}

func TestErrUnknownCameraProperty(t *testing.T) {
	mustFail(t, `Room "r" { Camera "c" { fov = (1.0, 2.0, 3.0) } }`, `unknown Camera property`)
}

func TestErrWalkableSizeMustBe2(t *testing.T) {
	mustFail(t,
		`Room "r" { WalkableSurface "f" { size = (1.0, 2.0, 3.0) } }`,
		`requires 2 components`,
	)
}

func TestErrBlockerSizeMustBe3(t *testing.T) {
	mustFail(t,
		`Room "r" { BlockerVolume "w" { size = (1.0, 2.0) } }`,
		`requires 3 components`,
	)
}

func TestErrTupleWrongCount(t *testing.T) {
	mustFail(t,
		`Room "r" { Camera "c" { position = (1.0) } }`,
		`tuple must have 2 or 3 components`,
	)
}

func TestErrTrailingContent(t *testing.T) {
	mustFail(t, `Room "r" {} extra`, `unexpected content after Room block`)
}

func TestErrLineNumbers(t *testing.T) {
	src := "Room \"r\" {\n\n\n    invalid_key = \"v\"\n}"
	_, err := room.ParseRoom("test.agroom", src)
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*room.ParseError)
	if !ok {
		t.Fatalf("expected *room.ParseError, got %T", err)
	}
	if pe.Line != 4 {
		t.Errorf("error line = %d, want 4", pe.Line)
	}
}
