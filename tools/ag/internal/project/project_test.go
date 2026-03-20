package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ags3d/ag/internal/project"
)

func TestFind_FindsGameAGP(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte("[project]\nname = \"Test\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, ok := project.Find(dir)
	if !ok {
		t.Fatal("Find returned false, expected true")
	}
	if got != dir {
		t.Fatalf("Find returned %q, want %q", got, dir)
	}
}

func TestFind_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, ok := project.Find(dir)
	if ok {
		t.Fatal("Find returned true on directory with no game.agp")
	}
}

func TestLoad_ParsesManifest(t *testing.T) {
	dir := t.TempDir()
	agp := `[project]
name = "My Game"
start_room = "rooms/start/start.agroom"
start_character = "characters/player.agchar"

[settings]
rendering_mode = "full_3d"
autosave = true
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if m.Project.Name != "My Game" {
		t.Errorf("Project.Name = %q, want %q", m.Project.Name, "My Game")
	}
	if m.Project.StartRoom != "rooms/start/start.agroom" {
		t.Errorf("Project.StartRoom = %q", m.Project.StartRoom)
	}
	if !m.Settings.Autosave {
		t.Error("Settings.Autosave = false, want true")
	}
	if m.Settings.RenderingMode != "full_3d" {
		t.Errorf("Settings.RenderingMode = %q, want full_3d", m.Settings.RenderingMode)
	}
}

func TestScan_FindsSourceFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"rooms/start/start.agroom",
		"rooms/start/start.agscript",
		"characters/player.agchar",
		"dialogue/intro.agdlg",
		"inventory/key.agitem",
		"audio/music.ogg",  // should be excluded
		"assets/bg.png",    // should be excluded
	}
	for _, f := range files {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// .engine/ should be excluded
	engineFile := filepath.Join(dir, ".engine", "generated", "foo.gd")
	if err := os.MkdirAll(filepath.Dir(engineFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(engineFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := project.Scan(dir)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("Scan returned %d files, want 5: %v", len(got), got)
	}
}

func TestScan_ExcludesEngineDir(t *testing.T) {
	dir := t.TempDir()
	engineScript := filepath.Join(dir, ".engine", "generated", "room.agscript")
	if err := os.MkdirAll(filepath.Dir(engineScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(engineScript, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := project.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("Scan should exclude .engine/, got %v", got)
	}
}

func TestBuildManifest_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := project.BuildManifest{"foo.agscript": 1234567890.5}
	if err := project.SaveManifest(dir, m); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}
	loaded, err := project.LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if loaded["foo.agscript"] != 1234567890.5 {
		t.Errorf("mtime round-trip failed: got %v", loaded["foo.agscript"])
	}
}

func TestBuildManifest_EmptyOnMissing(t *testing.T) {
	dir := t.TempDir()
	m, err := project.LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty manifest, got %v", m)
	}
}

func TestScaffold_CreatesExpectedLayout(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "mygame")
	if err := project.Scaffold(dest, "mygame"); err != nil {
		t.Fatalf("Scaffold error: %v", err)
	}
	expectedDirs := []string{
		"characters", "rooms/start", "dialogue",
		"inventory", "scripts", "audio", "assets",
	}
	for _, d := range expectedDirs {
		p := filepath.Join(dest, d)
		if info, err := os.Stat(p); err != nil || !info.IsDir() {
			t.Errorf("expected directory %q to exist", d)
		}
	}
	expectedFiles := []string{
		"game.agp",
		".gitignore",
		"rooms/start/start.agroom",
		"rooms/start/start.agscript",
		"characters/player.agchar",
	}
	for _, f := range expectedFiles {
		p := filepath.Join(dest, f)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected file %q to exist", f)
		}
	}
	// Verify game.agp is valid
	m, err := project.Load(dest)
	if err != nil {
		t.Fatalf("scaffolded project failed to load: %v", err)
	}
	if m.Project.Name != "mygame" {
		t.Errorf("scaffolded name = %q, want mygame", m.Project.Name)
	}
}
