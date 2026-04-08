package cut_test

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

// ── CollectLocEntries ────────────────────────────────────────────────────────

func TestCollect_Line_ExtractsText(t *testing.T) {
	cf, err := cut.Parse("test.agcut", "title: intro\nsequence:\n<<line narrator \"Hello world\">>\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	entries := cut.CollectLocEntries(cf)
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].Source != "Hello world" {
		t.Errorf("Source = %q, want %q", entries[0].Source, "Hello world")
	}
	if entries[0].CmdName != "line" {
		t.Errorf("CmdName = %q, want %q", entries[0].CmdName, "line")
	}
}

func TestCollect_TitleCard_ExtractsText(t *testing.T) {
	cf, err := cut.Parse("test.agcut", "title: ch1\nsequence:\n<<title_card \"Chapter One\" duration:2.0>>\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	entries := cut.CollectLocEntries(cf)
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].Source != "Chapter One" {
		t.Errorf("Source = %q, want %q", entries[0].Source, "Chapter One")
	}
	if entries[0].CmdName != "title_card" {
		t.Errorf("CmdName = %q, want %q", entries[0].CmdName, "title_card")
	}
}

func TestCollect_Subtitle_ExtractsText(t *testing.T) {
	cf, err := cut.Parse("test.agcut", "title: sc1\nsequence:\n<<subtitle \"[Sound of thunder]\">>\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	entries := cut.CollectLocEntries(cf)
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].Source != "[Sound of thunder]" {
		t.Errorf("Source = %q", entries[0].Source)
	}
}

func TestCollect_Choice_ExtractsText(t *testing.T) {
	// Choice commands in .agcut have a "choice" name.
	cf, err := cut.Parse("test.agcut", "title: dlg\nsequence:\n<<choice \"Yes, I'll go\">>\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	entries := cut.CollectLocEntries(cf)
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].Source != "Yes, I'll go" {
		t.Errorf("Source = %q", entries[0].Source)
	}
}

func TestCollect_ExplicitLocKey_Preserved(t *testing.T) {
	cf, err := cut.Parse("test.agcut", "title: sc\nsequence:\n<<line narrator \"Hello\" loc_key:intro_hello>>\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	entries := cut.CollectLocEntries(cf)
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].LocKey != "intro_hello" {
		t.Errorf("LocKey = %q, want %q", entries[0].LocKey, "intro_hello")
	}
}

func TestCollect_AutoLocKey_StableAcrossCalls(t *testing.T) {
	src := "title: opening\nsequence:\n<<line narrator \"Three years.\">>\n"
	cf1, _ := cut.Parse("a.agcut", src)
	cf2, _ := cut.Parse("b.agcut", src)
	e1 := cut.CollectLocEntries(cf1)
	e2 := cut.CollectLocEntries(cf2)
	// Keys differ only in the title prefix, not the hash.
	hash1 := strings.SplitN(e1[0].LocKey, ":", 3)
	hash2 := strings.SplitN(e2[0].LocKey, ":", 3)
	if len(hash1) < 3 || len(hash2) < 3 {
		t.Fatalf("loc key format unexpected: %q / %q", e1[0].LocKey, e2[0].LocKey)
	}
	// Hash (3rd segment) must be identical for same text.
	if hash1[2] != hash2[2] {
		t.Errorf("hash differs for same text: %q vs %q", hash1[2], hash2[2])
	}
}

func TestCollect_NonLocalizableCommands_Ignored(t *testing.T) {
	src := "title: x\nsequence:\n<<fade_in duration:1.0>>\n<<wait seconds:2>>\n<<camera set point.A>>\n"
	cf, _ := cut.Parse("x.agcut", src)
	entries := cut.CollectLocEntries(cf)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for non-localizable commands, got %d", len(entries))
	}
}

func TestCollect_MultipleLines_AllExtracted(t *testing.T) {
	src := "title: multi\nsequence:\n<<line narrator \"First line\">>\n<<line guard \"Second line\">>\n<<title_card \"Card\">>\n"
	cf, _ := cut.Parse("m.agcut", src)
	entries := cut.CollectLocEntries(cf)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

// ── WriteAgstringsTemplate ────────────────────────────────────────────────────

func TestWriteTemplate_ContainsMetaBlock(t *testing.T) {
	entries := []cut.LocEntry{
		{LocKey: "intro:0:abc12345", Source: "Hello"},
	}
	out := cut.WriteAgstringsTemplate("intro", entries)
	if !strings.Contains(out, "[meta]") {
		t.Error("template missing [meta] block")
	}
	if !strings.Contains(out, "[strings]") {
		t.Error("template missing [strings] block")
	}
}

func TestWriteTemplate_KeyAndValuePresent(t *testing.T) {
	entries := []cut.LocEntry{
		{LocKey: "intro:0:abc", Source: "Hello world"},
	}
	out := cut.WriteAgstringsTemplate("intro", entries)
	if !strings.Contains(out, "intro:0:abc") {
		t.Errorf("template missing loc key: %q", out)
	}
	if !strings.Contains(out, "Hello world") {
		t.Errorf("template missing source text: %q", out)
	}
}

func TestWriteTemplate_EmptyEntries(t *testing.T) {
	out := cut.WriteAgstringsTemplate("empty", nil)
	if !strings.Contains(out, "[strings]") {
		t.Error("template missing [strings] block even when empty")
	}
}

// ── ValidateLocKeys ────────────────────────────────────────────────────────────

func TestValidate_ExplicitLocKey_Present_NoError(t *testing.T) {
	cf, _ := cut.Parse("t.agcut", "title: t\nsequence:\n<<line narrator \"Hi\" loc_key:greeting>>\n")
	locale := map[string]string{"greeting": "Bonjour"}
	errs := cut.ValidateLocKeys(cf, locale)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidate_ExplicitLocKey_Missing_Error(t *testing.T) {
	cf, _ := cut.Parse("t.agcut", "title: t\nsequence:\n<<line narrator \"Hi\" loc_key:missing_key>>\n")
	locale := map[string]string{"other_key": "value"}
	errs := cut.ValidateLocKeys(cf, locale)
	if len(errs) == 0 {
		t.Error("expected error for missing loc_key, got none")
	}
	if !strings.Contains(errs[0], "missing_key") {
		t.Errorf("error should mention missing_key: %q", errs[0])
	}
}

func TestValidate_AutoGeneratedKey_NotValidated(t *testing.T) {
	// Lines without explicit loc_key: use auto keys — not checked against locale.
	cf, _ := cut.Parse("t.agcut", "title: t\nsequence:\n<<line narrator \"Hello\">>\n")
	errs := cut.ValidateLocKeys(cf, map[string]string{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for auto-keyed line, got %v", errs)
	}
}

func TestValidate_EmptyLocale_ExplicitKey_Error(t *testing.T) {
	cf, _ := cut.Parse("t.agcut", "title: t\nsequence:\n<<line narrator \"Bye\" loc_key:farewell>>\n")
	errs := cut.ValidateLocKeys(cf, map[string]string{})
	if len(errs) == 0 {
		t.Error("expected error when locale is empty and explicit loc_key used")
	}
}
