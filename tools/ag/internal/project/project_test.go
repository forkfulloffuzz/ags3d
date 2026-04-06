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

// --- T-DLG13: [locale.*] and [localisation] block tests ---

func TestLoad_LocaleEntries(t *testing.T) {
	dir := t.TempDir()
	agp := `[project]
name = "Test"
start_room = "rooms/start/start.agroom"
start_character = "characters/player.agchar"

[locale.en]
name = "English"
rtl = false

[locale.fr]
name = "French"
rtl = false

[locale.ar]
name = "Arabic"
rtl = true
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(m.Locales) != 3 {
		t.Fatalf("expected 3 locales, got %d", len(m.Locales))
	}
	en := m.Locales["en"]
	if en == nil {
		t.Fatal("locale 'en' not found")
	}
	if en.Name != "English" {
		t.Errorf("en.Name = %q, want English", en.Name)
	}
	if en.RTL {
		t.Error("en.RTL = true, want false")
	}
	ar := m.Locales["ar"]
	if ar == nil {
		t.Fatal("locale 'ar' not found")
	}
	if !ar.RTL {
		t.Error("ar.RTL = false, want true")
	}
}

func TestLoad_LocaleCodeSetFromSectionName(t *testing.T) {
	dir := t.TempDir()
	agp := "[project]\nname = \"T\"\n[locale.zh-TW]\nname = \"Traditional Chinese\"\n"
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	entry, ok := m.Locales["zh-TW"]
	if !ok {
		t.Fatal("locale 'zh-TW' not found")
	}
	if entry.Code != "zh-TW" {
		t.Errorf("Code = %q, want zh-TW", entry.Code)
	}
}

func TestLoad_LocalisationSection(t *testing.T) {
	dir := t.TempDir()
	agp := `[project]
name = "Test"

[locale.en]
name = "English"

[locale.fr]
name = "French"

[localisation]
base_locale = "en"
fallback_chain = "en, fr"
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if m.Localisation.BaseLocale != "en" {
		t.Errorf("BaseLocale = %q, want en", m.Localisation.BaseLocale)
	}
	if len(m.Localisation.FallbackChain) != 2 {
		t.Fatalf("FallbackChain len = %d, want 2", len(m.Localisation.FallbackChain))
	}
	if m.Localisation.FallbackChain[0] != "en" || m.Localisation.FallbackChain[1] != "fr" {
		t.Errorf("FallbackChain = %v, want [en fr]", m.Localisation.FallbackChain)
	}
}

func TestValidateLocales_ValidProject(t *testing.T) {
	dir := t.TempDir()
	agp := `[project]
name = "Test"

[locale.en]
name = "English"

[locale.fr]
name = "French"

[localisation]
base_locale = "en"
fallback_chain = "en, fr"
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	errs := m.ValidateLocales()
	if len(errs) != 0 {
		t.Errorf("unexpected validation errors: %v", errs)
	}
}

func TestValidateLocales_InvalidCode(t *testing.T) {
	dir := t.TempDir()
	agp := "[project]\nname = \"T\"\n[locale.INVALID]\nname = \"Bad\"\n"
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	errs := m.ValidateLocales()
	if len(errs) == 0 {
		t.Error("expected validation error for invalid locale code, got none")
	}
}

func TestValidateLocales_BaseLocaleMissing(t *testing.T) {
	dir := t.TempDir()
	agp := "[project]\nname = \"T\"\n[locale.en]\nname = \"English\"\n[localisation]\nbase_locale = \"de\"\n"
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	errs := m.ValidateLocales()
	if len(errs) == 0 {
		t.Error("expected error for undeclared base_locale, got none")
	}
}

func TestValidateLocales_FallbackChainMissing(t *testing.T) {
	dir := t.TempDir()
	agp := "[project]\nname = \"T\"\n[locale.en]\nname = \"English\"\n[localisation]\nbase_locale = \"en\"\nfallback_chain = \"en, xx\"\n"
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	errs := m.ValidateLocales()
	if len(errs) == 0 {
		t.Error("expected error for undeclared fallback_chain entry, got none")
	}
}

func TestLoad_NoLocalesSection(t *testing.T) {
	dir := t.TempDir()
	agp := "[project]\nname = \"T\"\nstart_room = \"r\"\nstart_character = \"c\"\n"
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(m.Locales) != 0 {
		t.Errorf("expected 0 locales, got %d", len(m.Locales))
	}
	errs := m.ValidateLocales()
	if len(errs) != 0 {
		t.Errorf("expected no errors on project without locales, got %v", errs)
	}
}

// --- T-CUT04: [cutscenes] and [input] block tests ---

func TestLoad_CutsceneSection(t *testing.T) {
	dir := t.TempDir()
	agp := `[project]
name = "Test"

[cutscenes]
fallback_debug = "halt"
fallback_release = "skip_and_continue"
fallback_qa = "log_and_continue"
step_timeout_default = "30"
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if m.Cutscenes.FallbackDebug != "halt" {
		t.Errorf("FallbackDebug = %q, want halt", m.Cutscenes.FallbackDebug)
	}
	if m.Cutscenes.FallbackRelease != "skip_and_continue" {
		t.Errorf("FallbackRelease = %q, want skip_and_continue", m.Cutscenes.FallbackRelease)
	}
	if m.Cutscenes.FallbackQA != "log_and_continue" {
		t.Errorf("FallbackQA = %q, want log_and_continue", m.Cutscenes.FallbackQA)
	}
	if m.Cutscenes.StepTimeoutDefault != 30 {
		t.Errorf("StepTimeoutDefault = %v, want 30", m.Cutscenes.StepTimeoutDefault)
	}
}

func TestLoad_InputSection(t *testing.T) {
	dir := t.TempDir()
	agp := `[project]
name = "Test"

[input]
dialogue_advance = "ui_accept"
cutscene_skip = "cutscene_skip"
dialogue_hold_advance = "ui_accept_hold"
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if m.Input.DialogueAdvance != "ui_accept" {
		t.Errorf("DialogueAdvance = %q, want ui_accept", m.Input.DialogueAdvance)
	}
	if m.Input.CutsceneSkip != "cutscene_skip" {
		t.Errorf("CutsceneSkip = %q, want cutscene_skip", m.Input.CutsceneSkip)
	}
	if m.Input.DialogueHoldAdvance != "ui_accept_hold" {
		t.Errorf("DialogueHoldAdvance = %q, want ui_accept_hold", m.Input.DialogueHoldAdvance)
	}
}

func TestLoad_NoCutsceneOrInputSection(t *testing.T) {
	dir := t.TempDir()
	agp := "[project]\nname = \"T\"\n"
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(agp), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if m.Cutscenes.FallbackDebug != "" {
		t.Errorf("expected empty FallbackDebug, got %q", m.Cutscenes.FallbackDebug)
	}
	if m.Input.DialogueAdvance != "" {
		t.Errorf("expected empty DialogueAdvance, got %q", m.Input.DialogueAdvance)
	}
}
