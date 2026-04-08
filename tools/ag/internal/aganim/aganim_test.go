package aganim_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ags3d/ag/internal/aganim"
)

// ── ParseFile ────────────────────────────────────────────────────────────────

func TestParseFile_Valid(t *testing.T) {
	data := `{
		"character": "player",
		"clips": [
			{
				"name": "Walk",
				"frame_tags": [
					{"name": "footstep_left", "frame": 12},
					{"name": "footstep_right", "frame": 24}
				]
			}
		]
	}`
	tmp := t.TempDir()
	path := filepath.Join(tmp, "player.aganim")
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	af, err := aganim.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if af.Character != "player" {
		t.Errorf("Character = %q, want %q", af.Character, "player")
	}
	if len(af.Clips) != 1 {
		t.Fatalf("len(Clips) = %d, want 1", len(af.Clips))
	}
	clip := af.Clips[0]
	if clip.Name != "Walk" {
		t.Errorf("Clips[0].Name = %q, want Walk", clip.Name)
	}
	if len(clip.FrameTags) != 2 {
		t.Fatalf("len(FrameTags) = %d, want 2", len(clip.FrameTags))
	}
	if clip.FrameTags[0].Name != "footstep_left" || clip.FrameTags[0].Frame != 12 {
		t.Errorf("FrameTags[0] = %+v, want {Name:footstep_left Frame:12}", clip.FrameTags[0])
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := aganim.ParseFile("/nonexistent/player.aganim")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestParseFile_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.aganim")
	os.WriteFile(path, []byte("not json"), 0644)
	_, err := aganim.ParseFile(path)
	if err == nil {
		t.Error("expected JSON parse error, got nil")
	}
}

// ── SidecarPath ───────────────────────────────────────────────────────────────

func TestSidecarPath(t *testing.T) {
	cases := []struct{ in, want string }{
		{"characters/player/player.glb", "characters/player/player.aganim"},
		{"player.glb", "player.aganim"},
		{"no_ext", "no_ext.aganim"},
	}
	for _, c := range cases {
		got := aganim.SidecarPath(c.in)
		if got != c.want {
			t.Errorf("SidecarPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── GDScriptLiteral ──────────────────────────────────────────────────────────

func TestGDScriptLiteral_Empty(t *testing.T) {
	var af *aganim.AnimFile
	got := af.GDScriptLiteral()
	if got != "{}" {
		t.Errorf("nil.GDScriptLiteral() = %q, want {}", got)
	}
}

func TestGDScriptLiteral_NoClips(t *testing.T) {
	af := &aganim.AnimFile{Character: "hero", Clips: nil}
	got := af.GDScriptLiteral()
	if got != "{}" {
		t.Errorf("GDScriptLiteral() = %q, want {}", got)
	}
}

func TestGDScriptLiteral_OneClipNoTags(t *testing.T) {
	af := &aganim.AnimFile{
		Character: "hero",
		Clips: []aganim.Clip{
			{Name: "Idle", FrameTags: nil},
		},
	}
	got := af.GDScriptLiteral()
	want := `{"Idle": []}`
	if got != want {
		t.Errorf("GDScriptLiteral() = %q, want %q", got, want)
	}
}

func TestGDScriptLiteral_WithTags(t *testing.T) {
	af := &aganim.AnimFile{
		Character: "player",
		Clips: []aganim.Clip{
			{
				Name: "Walk",
				FrameTags: []aganim.FrameTag{
					{Name: "footstep_left", Frame: 12},
					{Name: "footstep_right", Frame: 24},
				},
			},
		},
	}
	got := af.GDScriptLiteral()
	want := `{"Walk": [{"frame": 12, "name": "footstep_left"}, {"frame": 24, "name": "footstep_right"}]}`
	if got != want {
		t.Errorf("GDScriptLiteral() = %q, want %q", got, want)
	}
}

func TestGDScriptLiteral_MultipleClips(t *testing.T) {
	af := &aganim.AnimFile{
		Character: "guard",
		Clips: []aganim.Clip{
			{Name: "Walk", FrameTags: []aganim.FrameTag{{Name: "step", Frame: 10}}},
			{Name: "Idle", FrameTags: nil},
		},
	}
	got := af.GDScriptLiteral()
	// Verify it's parseable as a Go map (structural check).
	_ = got // The literal is GDScript, not JSON; just check it contains both clip names.
	if !contains(got, `"Walk"`) || !contains(got, `"Idle"`) {
		t.Errorf("GDScriptLiteral() missing clip names: %q", got)
	}
}

func TestGDScriptLiteral_EscapesQuotes(t *testing.T) {
	af := &aganim.AnimFile{
		Clips: []aganim.Clip{
			{Name: `Say "Hello"`, FrameTags: []aganim.FrameTag{{Name: `it's "done"`, Frame: 1}}},
		},
	}
	got := af.GDScriptLiteral()
	if !contains(got, `\"Hello\"`) {
		t.Errorf("GDScriptLiteral() did not escape quotes in clip name: %q", got)
	}
	if !contains(got, `\"done\"`) {
		t.Errorf("GDScriptLiteral() did not escape quotes in tag name: %q", got)
	}
}

// ── Round-trip: build + parse ─────────────────────────────────────────────────

func TestRoundTrip(t *testing.T) {
	original := &aganim.AnimFile{
		Character: "knight",
		Clips: []aganim.Clip{
			{
				Name: "Attack",
				FrameTags: []aganim.FrameTag{
					{Name: "hit", Frame: 15},
					{Name: "recover", Frame: 30},
				},
			},
		},
	}

	tmp := t.TempDir()
	path := filepath.Join(tmp, "knight.aganim")
	data, _ := json.Marshal(original)
	os.WriteFile(path, data, 0644)

	loaded, err := aganim.ParseFile(path)
	if err != nil {
		t.Fatalf("round-trip ParseFile: %v", err)
	}
	if loaded.Character != original.Character {
		t.Errorf("Character: got %q, want %q", loaded.Character, original.Character)
	}
	if len(loaded.Clips) != 1 || loaded.Clips[0].Name != "Attack" {
		t.Errorf("Clips not preserved after round-trip")
	}
	if loaded.Clips[0].FrameTags[0].Name != "hit" {
		t.Errorf("FrameTags not preserved after round-trip")
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
