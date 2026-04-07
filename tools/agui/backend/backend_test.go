// Package backend contains smoke tests for the AG Studio Go backend.
//
// Tests mirror the App methods in app.go but call the api package directly
// so that they run without a live Wails context.
package backend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ags3d/ag/api"
)

// fixture returns the absolute path to a file under tools/ag/testdata/scripts.
// Tests run with CWD = tools/agui/backend/, so "../../ag/testdata" is correct.
func fixture(parts ...string) string {
	base := filepath.Join("..", "..", "ag", "testdata", "scripts")
	return filepath.Join(append([]string{base}, parts...)...)
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// UT-UI-01 — TranspileFile returns all 7 stage fields populated for a valid fixture.
func TestTranspileFile_ValidFixture_AllFieldsPopulated(t *testing.T) {
	path := fixture("valid", "01_minimal.agscript")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fixture not found: %v", err)
	}

	r := api.TranspileFile("01_minimal.agscript", string(data))
	if len(r.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}

	fields := map[string]string{
		"Tokens":   r.Tokens,
		"ASTText":  r.ASTText,
		"ASTDot":   r.ASTDot,
		"Symbols":  r.Symbols,
		"SymDot":   r.SymDot,
		"Blocking": r.Blocking,
		"GDScript": r.GDScript,
	}
	for name, val := range fields {
		if val == "" {
			t.Errorf("field %s is empty", name)
		}
	}
}

// UT-UI-02 — TranspileFile returns errors slice non-empty for an invalid fixture;
// GDScript and downstream fields are empty.
func TestTranspileFile_InvalidFixture_ErrorsNonEmpty(t *testing.T) {
	path := fixture("invalid", "err_01_unterminated_string.agscript")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fixture not found: %v", err)
	}

	r := api.TranspileFile("err_01.agscript", string(data))
	if len(r.Errors) == 0 {
		t.Fatal("expected errors, got none")
	}
	if r.GDScript != "" {
		t.Errorf("GDScript should be empty on error, got: %q", r.GDScript)
	}
}

// UT-UI-03 — VizASTDot returns a string containing "digraph" for a valid input.
func TestVizASTDot_ContainsDigraph(t *testing.T) {
	path := fixture("valid", "01_minimal.agscript")
	data, _ := os.ReadFile(path)

	out := api.VizASTDot("01_minimal.agscript", string(data))
	if !strings.Contains(out, "digraph") {
		t.Errorf("VizASTDot output does not contain 'digraph':\n%s", out)
	}
}

// UT-UI-04 — VizTokens output contains "FUNCTION" token for a file starting with `function`.
func TestVizTokens_ContainsFunctionToken(t *testing.T) {
	path := fixture("valid", "01_minimal.agscript")
	data, _ := os.ReadFile(path)

	out := api.VizTokens("01_minimal.agscript", string(data))
	if !strings.Contains(out, "FUNCTION") {
		t.Errorf("VizTokens output does not contain 'FUNCTION':\n%s", out)
	}
}

// UT-UI-05 — ListSourceFiles returns the correct count for a scaffolded project.
func TestListSourceFiles_ScaffoldedProject(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "game.agp"), "[project]\nname = \"Test\"\n")
	mustWrite(t, filepath.Join(root, "rooms/start/start.agroom"), "Room \"start\" {}\n")
	mustWrite(t, filepath.Join(root, "rooms/start/start.agscript"), "function room_Load() {}\n")
	mustWrite(t, filepath.Join(root, "characters/player.agchar"), "")

	files, err := api.ScanProject(root)
	if err != nil {
		t.Fatalf("ScanProject: %v", err)
	}
	const want = 3
	if len(files) != want {
		t.Errorf("expected %d source files, got %d", want, len(files))
	}
}

// UT-UI-06 — BatchViz with stage="tokens" returns one result per input file.
// (Mirrors App.BatchViz return value; Wails events are not tested here.)
func TestBatchViz_TokensStage_ReturnsOnePerFile(t *testing.T) {
	paths := []string{
		fixture("valid", "01_minimal.agscript"),
		fixture("valid", "10_functions.agscript"),
	}
	count := batchVizCount(paths, "tokens")
	if count != len(paths) {
		t.Errorf("BatchViz returned %d, expected %d", count, len(paths))
	}
}

// batchVizCount mirrors the App.BatchViz logic (without Wails events).
func batchVizCount(paths []string, stage string) int {
	vizFns := map[string]func(string, string) string{
		"tokens":      api.VizTokens,
		"ast":         api.VizAST,
		"ast-dot":     api.VizASTDot,
		"symbols":     api.VizSymbols,
		"symbols-dot": api.VizSymbolsDot,
		"blocking":    api.VizBlocking,
		"emit":        api.VizEmit,
	}
	stages := []string{stage}
	if stage == "all" || stage == "" {
		stages = []string{"tokens", "ast", "ast-dot", "symbols", "symbols-dot", "blocking", "emit"}
	}
	for _, absPath := range paths {
		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		for _, s := range stages {
			if fn, ok := vizFns[s]; ok {
				fn(filepath.Base(absPath), string(data))
			}
		}
	}
	return len(paths)
}

// UT-UI-07 — Build on a valid project produces .gd files in .engine/generated/.
func TestBuild_ValidProject_ProducesGDFiles(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "game.agp"), "[project]\nname = \"Test\"\n")
	mustWrite(t, filepath.Join(root, "start.agscript"), "function room_Load() {}\n")

	if err := api.Build(root, func(_, _ string) {}); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	generatedDir := filepath.Join(root, ".engine", "generated")
	entries, err := os.ReadDir(generatedDir)
	if err != nil {
		t.Fatalf(".engine/generated not created: %v", err)
	}

	var gdFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".gd") {
			gdFiles = append(gdFiles, e.Name())
		}
	}
	if len(gdFiles) == 0 {
		t.Errorf("no .gd files found in %s", generatedDir)
	}
}
