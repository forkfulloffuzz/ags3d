package dlg_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ags3d/ag/internal/dlg"
)

func mustEmit(t *testing.T, srcs ...string) []*dlg.EmittedFile {
	t.Helper()
	files := mustParseFiles(t, srcs...)
	lp, err := dlg.Link(files)
	if err != nil {
		t.Fatalf("Link error: %v", err)
	}
	valErrs := dlg.Validate(lp)
	if len(valErrs) != 0 {
		t.Fatalf("Validate errors: %v", valErrs)
	}
	outDir := t.TempDir()
	emitErrs := dlg.EmitProject(lp, outDir)
	if len(emitErrs) != 0 {
		t.Fatalf("EmitProject errors: %v", emitErrs)
	}

	// Read back all emitted JSON files.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var emitted []*dlg.EmittedFile
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(outDir, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		var ef dlg.EmittedFile
		if err := json.Unmarshal(data, &ef); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		emitted = append(emitted, &ef)
	}
	return emitted
}

// --- Output structure ---

func TestEmit_ProducesOneFilePerSource(t *testing.T) {
	result := mustEmit(t,
		"title: a\n---\n<<end>>\n===",
		"title: b\n---\n<<end>>\n===",
	)
	if len(result) != 2 {
		t.Errorf("expected 2 emitted files, got %d", len(result))
	}
}

func TestEmit_NodeTitlePreserved(t *testing.T) {
	result := mustEmit(t, "title: guard_greeting\ncharacter: guard\n---\n<<end>>\n===")
	if len(result[0].Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result[0].Nodes))
	}
	n := result[0].Nodes[0]
	if n.Title != "guard_greeting" {
		t.Errorf("Title = %q, want guard_greeting", n.Title)
	}
	if n.Character != "guard" {
		t.Errorf("Character = %q, want guard", n.Character)
	}
}

func TestEmit_TagsPreserved(t *testing.T) {
	result := mustEmit(t, "title: t\ntags: [chapter:1, cinematic]\n---\n<<end>>\n===")
	n := result[0].Nodes[0]
	if len(n.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(n.Tags))
	}
}

// --- Statement types ---

func TestEmit_SpeakerLineType(t *testing.T) {
	result := mustEmit(t, "title: t\n---\nGuard: Hello.\n<<end>>\n===")
	stmts := result[0].Nodes[0].Body
	if stmts[0].Type != "speaker_line" {
		t.Errorf("Body[0].Type = %q, want speaker_line", stmts[0].Type)
	}
	if stmts[0].Speaker != "Guard" || stmts[0].Text != "Hello." {
		t.Errorf("Speaker=%q Text=%q", stmts[0].Speaker, stmts[0].Text)
	}
}

func TestEmit_NarrationLineType(t *testing.T) {
	result := mustEmit(t, "title: t\n---\nOnce upon a time.\n<<end>>\n===")
	if result[0].Nodes[0].Body[0].Type != "narration" {
		t.Errorf("Type = %q, want narration", result[0].Nodes[0].Body[0].Type)
	}
}

func TestEmit_OptionType(t *testing.T) {
	result := mustEmit(t, "title: t\n---\n-> Yes.\n   <<end>>\n===")
	if result[0].Nodes[0].Body[0].Type != "option" {
		t.Errorf("Type = %q, want option", result[0].Nodes[0].Body[0].Type)
	}
}

func TestEmit_CommandType(t *testing.T) {
	result := mustEmit(t, "title: t\n---\n<<end>>\n===")
	if result[0].Nodes[0].Body[0].Type != "command" {
		t.Errorf("Type = %q, want command", result[0].Nodes[0].Body[0].Type)
	}
	if result[0].Nodes[0].Body[0].Raw != "end" {
		t.Errorf("Raw = %q, want end", result[0].Nodes[0].Body[0].Raw)
	}
}

func TestEmit_OptionBodyNested(t *testing.T) {
	src := "title: t\n---\n-> Yes.\n   Guard: OK.\n   <<end>>\n==="
	result := mustEmit(t, src)
	opt := result[0].Nodes[0].Body[0]
	if len(opt.Body) != 2 {
		t.Fatalf("option body len = %d, want 2", len(opt.Body))
	}
	if opt.Body[0].Type != "speaker_line" {
		t.Errorf("opt.Body[0].Type = %q, want speaker_line", opt.Body[0].Type)
	}
}

func TestEmit_InlineCommandsOnSpeakerLine(t *testing.T) {
	src := "title: t\n---\nGuard: Hi. <<action flag.met = true>>\n<<end>>\n==="
	result := mustEmit(t, src)
	sl := result[0].Nodes[0].Body[0]
	if len(sl.Commands) != 1 || sl.Commands[0] != "action flag.met = true" {
		t.Errorf("Commands = %v", sl.Commands)
	}
}

// --- Loc key assignment ---

func TestEmit_AutoLocKeyAssigned(t *testing.T) {
	result := mustEmit(t, "title: t\n---\nGuard: Hello.\n<<end>>\n===")
	locKey := result[0].Nodes[0].Body[0].LocKey
	if locKey == "" {
		t.Error("expected auto loc key, got empty string")
	}
	// Auto key format: "nodeTitle:lineIndex:8hexchars"
	if locKey[:2] != "t:" {
		t.Errorf("loc key %q doesn't start with node title 't:'", locKey)
	}
}

func TestEmit_ManualLocKeyPreserved(t *testing.T) {
	src := "title: t\n---\nGuard: Hi. #loc:guard_hi_001\n<<end>>\n==="
	result := mustEmit(t, src)
	locKey := result[0].Nodes[0].Body[0].LocKey
	if locKey != "guard_hi_001" {
		t.Errorf("LocKey = %q, want guard_hi_001", locKey)
	}
}

func TestEmit_AutoLocKeyStable(t *testing.T) {
	// Same text should always produce the same hash.
	src := "title: t\n---\nGuard: Hello.\n<<end>>\n==="
	r1 := mustEmit(t, src)
	r2 := mustEmit(t, src)
	if r1[0].Nodes[0].Body[0].LocKey != r2[0].Nodes[0].Body[0].LocKey {
		t.Error("auto loc key is not stable across two runs")
	}
}

func TestEmit_LineIndexIncrements(t *testing.T) {
	src := "title: t\n---\nGuard: First.\nGuard: Second.\n<<end>>\n==="
	result := mustEmit(t, src)
	body := result[0].Nodes[0].Body
	lk0 := body[0].LocKey
	lk1 := body[1].LocKey
	if lk0 == lk1 {
		t.Error("consecutive lines have identical loc keys — line index not incrementing")
	}
	// Keys should differ in the index position: "t:0:..." vs "t:1:..."
	if lk0[:4] != "t:0:" {
		t.Errorf("first line key %q should start with t:0:", lk0)
	}
	if lk1[:4] != "t:1:" {
		t.Errorf("second line key %q should start with t:1:", lk1)
	}
}

// --- JSON output ---

func TestEmit_ValidJSON(t *testing.T) {
	files := mustParseFiles(t, "title: t\n---\nGuard: Hi.\n<<end>>\n===")
	lp, _ := dlg.Link(files)
	outDir := t.TempDir()
	dlg.EmitProject(lp, outDir)

	entries, _ := os.ReadDir(outDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 JSON file, got %d", len(entries))
	}
	data, _ := os.ReadFile(filepath.Join(outDir, entries[0].Name()))
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Errorf("emitted file is not valid JSON: %v", err)
	}
}
