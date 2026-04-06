package dlg_test

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/dlg"
)

func mustCollect(t *testing.T, srcs ...string) []dlg.LocEntry {
	t.Helper()
	files := mustParseFiles(t, srcs...)
	lp, err := dlg.Link(files)
	if err != nil {
		t.Fatalf("Link error: %v", err)
	}
	return dlg.CollectLocEntries(lp)
}

// --- CollectLocEntries ---

func TestCollect_SpeakerLineProducesSpokenEntry(t *testing.T) {
	entries := mustCollect(t, "title: a\ncharacter: guard\n---\nGuard: You there.\n<<end>>\n===")
	if len(entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.LineType != "spoken" {
		t.Errorf("LineType = %q, want spoken", e.LineType)
	}
	if e.Character != "Guard" {
		t.Errorf("Character = %q, want Guard", e.Character)
	}
	if e.Source != "You there." {
		t.Errorf("Source = %q", e.Source)
	}
	if e.NodeTitle != "a" {
		t.Errorf("NodeTitle = %q, want a", e.NodeTitle)
	}
}

func TestCollect_OptionProducesChoiceEntry(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\n-> Let me pass.\n   <<end>>\n===")
	if len(entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(entries))
	}
	if entries[0].LineType != "choice" {
		t.Errorf("LineType = %q, want choice", entries[0].LineType)
	}
	if entries[0].Source != "Let me pass." {
		t.Errorf("Source = %q", entries[0].Source)
	}
}

func TestCollect_NarrationProducesNarrationEntry(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nThe door creaks open.\n<<end>>\n===")
	if len(entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(entries))
	}
	if entries[0].LineType != "narration" {
		t.Errorf("LineType = %q, want narration", entries[0].LineType)
	}
}

func TestCollect_CommandsNotCollected(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\n<<jump b>>\n===\ntitle: b\n---\n<<end>>\n===")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for command-only nodes, got %d", len(entries))
	}
}

func TestCollect_LocKeyStable(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hello.\n<<end>>\n===")
	if entries[0].LocKey == "" {
		t.Error("LocKey should not be empty")
	}
	// Running again should produce the same key.
	entries2 := mustCollect(t, "title: a\n---\nGuard: Hello.\n<<end>>\n===")
	if entries[0].LocKey != entries2[0].LocKey {
		t.Errorf("LocKey not stable: %q != %q", entries[0].LocKey, entries2[0].LocKey)
	}
}

// --- ExportPO ---

func TestExportPO_ContainsMsgid(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: You there.\n<<end>>\n===")
	out := dlg.ExportPO(entries, nil, false)
	if !strings.Contains(out, `msgid "You there."`) {
		t.Errorf("PO output missing msgid: %s", out)
	}
}

func TestExportPO_ContainsMsgctxt(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hello.\n<<end>>\n===")
	out := dlg.ExportPO(entries, nil, false)
	if !strings.Contains(out, "msgctxt") {
		t.Error("PO output missing msgctxt")
	}
}

func TestExportPO_ContainsTypeComment(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hi.\n<<end>>\n===")
	out := dlg.ExportPO(entries, nil, false)
	if !strings.Contains(out, "#. Type: spoken") {
		t.Errorf("PO output missing type comment: %s", out)
	}
}

func TestExportPO_EmptyMsgstrByDefault(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hi.\n<<end>>\n===")
	out := dlg.ExportPO(entries, nil, false)
	if !strings.Contains(out, `msgstr ""`) {
		t.Errorf("PO output should have empty msgstr: %s", out)
	}
}

func TestExportPO_DiffSkipsTranslated(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hi.\n<<end>>\n===")
	key := entries[0].LocKey
	existing := map[string]string{key: "Salut."}
	out := dlg.ExportPO(entries, existing, true)
	if strings.Contains(out, "Hi.") {
		t.Error("diff mode should skip already-translated entries")
	}
}

func TestExportPO_DiffIncludesUntranslated(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hi.\n<<end>>\n===")
	out := dlg.ExportPO(entries, map[string]string{}, true)
	if !strings.Contains(out, "Hi.") {
		t.Error("diff mode should include untranslated entries")
	}
}

// --- ExportCSV ---

func TestExportCSV_HeaderRow(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hi.\n<<end>>\n===")
	out := dlg.ExportCSV(entries, nil, false)
	if !strings.HasPrefix(out, "loc_key,node,character,type,source_text,translation") {
		t.Errorf("CSV missing header: %s", out[:min(80, len(out))])
	}
}

func TestExportCSV_DataRow(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hi.\n<<end>>\n===")
	out := dlg.ExportCSV(entries, nil, false)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines (header + data), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "Hi.") {
		t.Errorf("data row missing source text: %s", lines[1])
	}
}

func TestExportCSV_DiffSkipsTranslated(t *testing.T) {
	entries := mustCollect(t, "title: a\n---\nGuard: Hi.\n<<end>>\n===")
	key := entries[0].LocKey
	existing := map[string]string{key: "Salut."}
	out := dlg.ExportCSV(entries, existing, true)
	if strings.Contains(out, "Hi.") {
		t.Error("diff mode should skip already-translated entries")
	}
}

// --- ParsePOTranslations ---

func TestParsePOTranslations_ExtractsNonEmpty(t *testing.T) {
	po := `msgctxt "key1"
msgid "Hello."
msgstr "Bonjour."

msgctxt "key2"
msgid "Bye."
msgstr ""
`
	m := dlg.ParsePOTranslations(po)
	if m["key1"] != "Bonjour." {
		t.Errorf("key1 = %q, want Bonjour.", m["key1"])
	}
	if _, ok := m["key2"]; ok {
		t.Error("key2 should not be in map (empty msgstr)")
	}
}

// --- ParseCSVTranslations ---

func TestParseCSVTranslations_ExtractsNonEmpty(t *testing.T) {
	csv := "loc_key,node,character,type,source_text,translation\n" +
		"key1,a,guard,spoken,Hello.,Bonjour.\n" +
		"key2,a,guard,spoken,Bye.,\n"
	m := dlg.ParseCSVTranslations(csv)
	if m["key1"] != "Bonjour." {
		t.Errorf("key1 = %q, want Bonjour.", m["key1"])
	}
	if _, ok := m["key2"]; ok {
		t.Error("key2 should not be in map (empty translation)")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
