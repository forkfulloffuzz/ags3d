package validate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/validate"
)

// scaffoldFiles creates a temp dir with given files and returns a
// []project.SourceFile list suitable for ValidateFiles.
func scaffoldFiles(t *testing.T, files map[string]string) []project.SourceFile {
	t.Helper()
	root := t.TempDir()
	var out []project.SourceFile
	for rel, content := range files {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		out = append(out, project.SourceFile{
			Path: abs,
			Rel:  rel,
			Ext:  filepath.Ext(rel),
		})
	}
	return out
}

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

// --------------------------------------------------------------------------
// Check 5: .agscript point-name cross-references
// --------------------------------------------------------------------------

func TestScriptWalkToKnownPoint(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			Point "door" { position = (1.0, 0.0, 0.0) }
		}`,
		"rooms/r/r.agscript": `function room_Enter() { player.WalkTo("door"); }`,
	})

	issues, _ := validate.ValidateProject(root, m)
	for _, i := range issues {
		if contains(i.String(), "WalkTo") || contains(i.String(), "door") {
			t.Errorf("unexpected point issue: %v", i)
		}
	}
}

func TestScriptWalkToUnknownPoint(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			Point "door" { position = (1.0, 0.0, 0.0) }
		}`,
		"rooms/r/r.agscript": `function room_Enter() { player.WalkTo("window"); }`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"window"`) {
		t.Errorf("expected unknown point issue for 'window', got %v", issues)
	}
	if !hasIssue(issues, "error") {
		t.Errorf("expected severity error")
	}
}

func TestScriptFaceToUnknownPoint(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {
			Point "door" { position = (1.0, 0.0, 0.0) }
		}`,
		"rooms/r/r.agscript": `function room_Enter() { player.FaceTo("nowhere"); }`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"nowhere"`) {
		t.Errorf("expected unknown point issue for 'nowhere', got %v", issues)
	}
}

func TestScriptNoRoomNoPanic(t *testing.T) {
	// Global script with no paired .agroom — no point checks, no crash.
	root, m := scaffold(t, map[string]string{
		"scripts/global.agscript": `function on_start() { player.WalkTo("anywhere"); }`,
	})

	issues, err := validate.ValidateProject(root, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No room paired → no point check → no issues from check 5.
	for _, i := range issues {
		if contains(i.String(), "WalkTo") {
			t.Errorf("unexpected WalkTo issue for unpairedscript: %v", i)
		}
	}
}

func TestScriptLineNumberReported(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agroom": `Room "r" {}`,
		"rooms/r/r.agscript": `function room_Enter() {
	player.WalkTo("missing_point");
}`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, "missing_point") {
		t.Errorf("expected issue for missing_point, got %v", issues)
	}
	// Line number should be included (line 2).
	if !hasIssue(issues, ":2:") {
		t.Errorf("expected line number in issue, got %v", issues)
	}
}

// --------------------------------------------------------------------------
// Check 6: .agscript inventory call cross-references
// --------------------------------------------------------------------------

func TestAddInventoryKnownItem(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"items/rusty_key.agitem": `Item "rusty_key" { display_name = "Rusty Key" }`,
		"rooms/r/r.agscript":    `function room_Enter() { AddInventory("rusty_key"); }`,
	})

	issues, _ := validate.ValidateProject(root, m)
	for _, i := range issues {
		if contains(i.String(), "rusty_key") {
			t.Errorf("unexpected item issue: %v", i)
		}
	}
}

func TestAddInventoryUnknownItem(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agscript": `function room_Enter() { AddInventory("magic_wand"); }`,
		// No .agitem for magic_wand
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"magic_wand"`) {
		t.Errorf("expected unknown item issue, got %v", issues)
	}
	if !hasIssue(issues, "error") {
		t.Errorf("expected severity error")
	}
}

func TestLoseInventoryUnknownItem(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agscript": `function room_Enter() { LoseInventory("ghost_item"); }`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"ghost_item"`) {
		t.Errorf("expected unknown item issue for LoseInventory, got %v", issues)
	}
}

func TestHasInventoryUnknownItem(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"rooms/r/r.agscript": `function room_Enter() { if (HasInventory("nope")) {} }`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"nope"`) {
		t.Errorf("expected unknown item issue for HasInventory, got %v", issues)
	}
}

func TestInventoryCheckGlobalScript(t *testing.T) {
	// Script with no paired room should still get inventory checks.
	root, m := scaffold(t, map[string]string{
		"scripts/global.agscript": `function on_start() { AddInventory("no_such_item"); }`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, `"no_such_item"`) {
		t.Errorf("expected inventory issue for global script, got %v", issues)
	}
}

func TestBrokenAgitemReportedAsIssue(t *testing.T) {
	root, m := scaffold(t, map[string]string{
		"items/bad.agitem": `NOT VALID SYNTAX`,
	})

	issues, _ := validate.ValidateProject(root, m)
	if !hasIssue(issues, "bad.agitem") {
		t.Errorf("expected parse error issue for bad.agitem, got %v", issues)
	}
}

// --------------------------------------------------------------------------
// T-LOC04 — Localisation validation
// --------------------------------------------------------------------------

const agstringsClean = `[meta]
base_locale = en
locale      = fr

[strings]
node1:line0:aabb1122 = "Bonjour."
node1:line1:ccdd3344 = "Au revoir."
`

const agstringsStale = `[meta]
base_locale = en
locale      = fr

[strings]
// [stale] node1:line0:aabb1122 = "Bonjour."
node1:line0:eeff5566 = ""
`

const agstringsOrphan = `[meta]
base_locale = en
locale      = fr

[strings]
// [orphan] node1:line0:aabb1122 = "Bonjour."
`

const agstringsUntranslated = `[meta]
base_locale = en
locale      = fr

[strings]
node1:line0:aabb1122 = ""
`

func TestLocaleClean(t *testing.T) {
	files := scaffoldFiles(t, map[string]string{
		"locale/fr.agstrings": agstringsClean,
	})
	issues, _ := validate.ValidateFiles(files)
	for _, iss := range issues {
		if contains(iss.File, ".agstrings") {
			t.Errorf("expected no locale issues for clean file, got %v", iss)
		}
	}
}

func TestLocaleStaleKeyReportsWarning(t *testing.T) {
	files := scaffoldFiles(t, map[string]string{
		"locale/fr.agstrings": agstringsStale,
	})
	issues, _ := validate.ValidateFiles(files)
	if !hasIssue(issues, "DLG-LOC-W001") {
		t.Errorf("expected DLG-LOC-W001 for stale key, got %v", issues)
	}
}

func TestLocaleOrphanKeyReportsWarning(t *testing.T) {
	files := scaffoldFiles(t, map[string]string{
		"locale/fr.agstrings": agstringsOrphan,
	})
	issues, _ := validate.ValidateFiles(files)
	if !hasIssue(issues, "DLG-LOC-W002") {
		t.Errorf("expected DLG-LOC-W002 for orphan key, got %v", issues)
	}
}

func TestLocaleUntranslatedDevMode(t *testing.T) {
	// In dev mode (default), untranslated keys are warnings, not errors.
	t.Setenv("AGSBUILD", "")
	files := scaffoldFiles(t, map[string]string{
		"locale/fr.agstrings": agstringsUntranslated,
	})
	issues, _ := validate.ValidateFiles(files)
	if !hasIssue(issues, "DLG-LOC-W003") {
		t.Errorf("expected DLG-LOC-W003 warning for untranslated key in dev mode, got %v", issues)
	}
	for _, iss := range issues {
		if contains(iss.String(), "DLG-LOC-E001") {
			t.Errorf("unexpected error DLG-LOC-E001 in dev mode, got %v", iss)
		}
	}
}

func TestLocaleUntranslatedReleaseMode(t *testing.T) {
	// In release mode, untranslated keys escalate to errors.
	t.Setenv("AGSBUILD", "release")
	files := scaffoldFiles(t, map[string]string{
		"locale/fr.agstrings": agstringsUntranslated,
	})
	issues, _ := validate.ValidateFiles(files)
	if !hasIssue(issues, "DLG-LOC-E001") {
		t.Errorf("expected DLG-LOC-E001 error for untranslated key in release mode, got %v", issues)
	}
}

func TestLocaleMultipleFilesEachChecked(t *testing.T) {
	files := scaffoldFiles(t, map[string]string{
		"locale/fr.agstrings": agstringsStale,
		"locale/de.agstrings": agstringsOrphan,
	})
	issues, _ := validate.ValidateFiles(files)
	if !hasIssue(issues, "DLG-LOC-W001") {
		t.Errorf("expected DLG-LOC-W001 for fr.agstrings, got %v", issues)
	}
	if !hasIssue(issues, "DLG-LOC-W002") {
		t.Errorf("expected DLG-LOC-W002 for de.agstrings, got %v", issues)
	}
}
