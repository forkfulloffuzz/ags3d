package validate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/validate"
)

// --------------------------------------------------------------------------
// Test fixture helpers
// --------------------------------------------------------------------------

// scaffold creates a temp project dir with the given files (path → content).
// Returns the project root and a minimal Manifest.
func scaffold(t *testing.T, files map[string]string) (string, *project.Manifest) {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	m := &project.Manifest{Root: root}
	return root, m
}

func hasIssue(issues []validate.Issue, substr string) bool {
	for _, i := range issues {
		if contains(i.String(), substr) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// --------------------------------------------------------------------------
// Clean project — no issues
// --------------------------------------------------------------------------

func TestCleanProject(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/start/start.agroom": `Room "start" {
			initial_camera = "main"
			Camera "main" { position = (0.0, 5.0, 5.0)  look_at = (0.0, 0.0, 0.0) }
			SpawnPoint "player_start" { character = "player"  position = (0.0, 0.0, 0.0) }
		}`,
		"characters/player.agchar": `Character "player" { display_name = "Player" }`,
	})
	m.Project.StartRoom = "rooms/start/start.agroom"
	m.Project.StartCharacter = "characters/player.agchar"

	issues, err := validate.ValidateProject(root, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d:", len(issues))
		for _, i := range issues {
			t.Logf("  %v", i)
		}
	}
}

// --------------------------------------------------------------------------
// game.agp checks
// --------------------------------------------------------------------------

func TestMissingStartRoom(t *testing.T) {
	root, m := scaffold(t, map[string]string{})
	m.Project.StartRoom = "rooms/missing/missing.agroom"

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, "start_room") {
		t.Errorf("expected start_room issue, got %v", issues)
	}
	if !hasIssue(issues, "error") {
		t.Errorf("expected severity error")
	}
}

func TestMissingStartCharacter(t *testing.T) {
	root, m := scaffold(t, map[string]string{})
	m.Project.StartCharacter = "characters/ghost.agchar"

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, "start_character") {
		t.Errorf("expected start_character issue, got %v", issues)
	}
}

func TestPresentStartRoomNoIssue(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/start/start.agroom": `Room "start" {}`,
	})
	m.Project.StartRoom = "rooms/start/start.agroom"

	issues, _ := validate.ValidateProject(root, m)
	for _, i := range issues {
		if contains(i.String(), "start_room") {
			t.Errorf("unexpected start_room issue: %v", i)
		}
	}
}

// --------------------------------------------------------------------------
// initial_camera check
// --------------------------------------------------------------------------

func TestInitialCameraMissing(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			initial_camera = "nonexistent"
		}`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"nonexistent"`) {
		t.Errorf("expected initial_camera issue, got %v", issues)
	}
	if !hasIssue(issues, "r.agroom") {
		t.Errorf("expected file r.agroom in issue, got %v", issues)
	}
}

func TestInitialCameraPresent(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			initial_camera = "main"
			Camera "main" { position = (0.0, 5.0, 5.0)  look_at = (0.0, 0.0, 0.0) }
		}`,
	})

	issues, _ := validate.ValidateProject(root, m)
	for _, i := range issues {
		if contains(i.String(), "initial_camera") {
			t.Errorf("unexpected initial_camera issue: %v", i)
		}
	}
}

func TestInitialCameraEmptyNoIssue(t *testing.T) {
	// No initial_camera set → no issue (it's optional)
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			Camera "main" { position = (0.0, 5.0, 5.0)  look_at = (0.0, 0.0, 0.0) }
		}`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

// --------------------------------------------------------------------------
// SpawnPoint.character check
// --------------------------------------------------------------------------

func TestSpawnPointUnknownCharacter(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			SpawnPoint "sp" { character = "ghost"  position = (0.0, 0.0, 0.0) }
		}`,
		// No ghost.agchar — deliberately missing
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"ghost"`) {
		t.Errorf("expected unknown character issue, got %v", issues)
	}
}

func TestSpawnPointKnownCharacter(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			SpawnPoint "sp" { character = "player"  position = (0.0, 0.0, 0.0) }
		}`,
		"characters/player.agchar": `Character "player" {}`,
	})

	issues, _ := validate.ValidateProject(root, m)
	for _, i := range issues {
		if contains(i.String(), "unknown character") {
			t.Errorf("unexpected unknown character issue: %v", i)
		}
	}
}

func TestSpawnPointEmptyCharacterNoIssue(t *testing.T) {
	// SpawnPoint with no character field → no issue
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			SpawnPoint "sp" { position = (0.0, 0.0, 0.0) }
		}`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

// --------------------------------------------------------------------------
// Multiple issues in one run
// --------------------------------------------------------------------------

func TestMultipleIssues(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			initial_camera = "no_such_cam"
			SpawnPoint "sp" { character = "nobody"  position = (0.0, 0.0, 0.0) }
		}`,
	})
	m.Project.StartRoom = "rooms/missing.agroom"

	issues, _ := validate.ValidateProject(root, m)
	if len(issues) < 3 {
		t.Errorf("expected at least 3 issues, got %d: %v", len(issues), issues)
	}
}

// --------------------------------------------------------------------------
// Broken .agchar / .agroom parse errors are reported as issues
// --------------------------------------------------------------------------

func TestBrokenAgcharReportedAsIssue(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"characters/bad.agchar": `NOT VALID SYNTAX {{{{`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, "bad.agchar") {
		t.Errorf("expected parse error issue for bad.agchar, got %v", issues)
	}
}

func TestBrokenAgroomReportedAsIssue(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" { BROKEN`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, "r.agroom") {
		t.Errorf("expected parse error issue for r.agroom, got %v", issues)
	}
}
