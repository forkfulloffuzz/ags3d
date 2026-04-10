package loc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/loc"
)

// --------------------------------------------------------------------------
// Parse
// --------------------------------------------------------------------------

const minimalSrc = `[meta]
base_locale = en
locale      = fr

[strings]
guard_greeting:line0:a1b2c3d4 = "Vous. Arrêtez."
guard_greeting:line1:e5f6a7b8 = ""
`

func TestParse_Meta(t *testing.T) {
	sf, err := loc.Parse("test.agstrings", minimalSrc)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if sf.Meta.Locale != "fr" {
		t.Errorf("Locale = %q, want fr", sf.Meta.Locale)
	}
	if sf.Meta.BaseLocale != "en" {
		t.Errorf("BaseLocale = %q, want en", sf.Meta.BaseLocale)
	}
}

func TestParse_Entries(t *testing.T) {
	sf, err := loc.Parse("test.agstrings", minimalSrc)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(sf.Entries) != 2 {
		t.Fatalf("entries len = %d, want 2", len(sf.Entries))
	}
	if sf.Entries[0].Value != "Vous. Arrêtez." {
		t.Errorf("Entries[0].Value = %q", sf.Entries[0].Value)
	}
	if sf.Entries[1].Value != "" {
		t.Errorf("Entries[1].Value = %q, want empty", sf.Entries[1].Value)
	}
}

func TestParse_Get(t *testing.T) {
	sf, err := loc.Parse("test.agstrings", minimalSrc)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := sf.Get("guard_greeting:line0:a1b2c3d4"); got != "Vous. Arrêtez." {
		t.Errorf("Get = %q", got)
	}
	if got := sf.Get("nonexistent"); got != "" {
		t.Errorf("Get nonexistent = %q, want empty", got)
	}
}

func TestParse_RTL(t *testing.T) {
	src := "[meta]\nbase_locale = en\nlocale = ar\nrtl = true\n\n[strings]\n"
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !sf.Meta.RTL {
		t.Error("RTL should be true")
	}
}

func TestParse_FallbackChain(t *testing.T) {
	src := "[meta]\nbase_locale = en\nlocale = fr_CA\nfallback_chain = fr en\n\n[strings]\n"
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(sf.Meta.FallbackChain) != 2 || sf.Meta.FallbackChain[0] != "fr" {
		t.Errorf("FallbackChain = %v", sf.Meta.FallbackChain)
	}
}

func TestParse_StaleMarker(t *testing.T) {
	src := "[meta]\nbase_locale = en\nlocale = fr\n\n[strings]\n// [stale] key:line0:abc = \"old\"\n"
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(sf.Entries) != 1 || !sf.Entries[0].Stale {
		t.Errorf("expected 1 stale entry, got %+v", sf.Entries)
	}
}

func TestParse_OrphanMarker(t *testing.T) {
	src := "[meta]\nbase_locale = en\nlocale = fr\n\n[strings]\n// [orphan] old:key:abc = \"gone\"\n"
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(sf.Entries) != 1 || !sf.Entries[0].Orphan {
		t.Errorf("expected 1 orphan entry, got %+v", sf.Entries)
	}
}

func TestParse_ErrMissingLocale(t *testing.T) {
	src := "[meta]\nbase_locale = en\n\n[strings]\n"
	_, err := loc.Parse("test.agstrings", src)
	if err == nil || !strings.Contains(err.Error(), "locale") {
		t.Errorf("expected locale error, got %v", err)
	}
}

func TestParse_ErrMissingBaseLocale(t *testing.T) {
	src := "[meta]\nlocale = fr\n\n[strings]\n"
	_, err := loc.Parse("test.agstrings", src)
	if err == nil || !strings.Contains(err.Error(), "base_locale") {
		t.Errorf("expected base_locale error, got %v", err)
	}
}

func TestParse_ErrDuplicateKey(t *testing.T) {
	src := "[meta]\nbase_locale = en\nlocale = fr\n\n[strings]\nkey:a = \"v\"\nkey:a = \"v2\"\n"
	_, err := loc.Parse("test.agstrings", src)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected duplicate key error, got %v", err)
	}
}

func TestParse_ErrUnknownMetaField(t *testing.T) {
	src := "[meta]\nbase_locale = en\nlocale = fr\nunknown_field = x\n\n[strings]\n"
	_, err := loc.Parse("test.agstrings", src)
	if err == nil || !strings.Contains(err.Error(), "unknown meta field") {
		t.Errorf("expected unknown meta field error, got %v", err)
	}
}

func TestParse_MetadataComments(t *testing.T) {
	src := `[meta]
base_locale = en
locale = ar

[strings]
// type: spoken
// char: Guard
// scene: guard_greeting
// ctx: guard is suspicious, not yet hostile
guard_greeting:0:a1b2c3d4 = "Arrêtez !"
`
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(sf.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(sf.Entries))
	}
	e := sf.Entries[0]
	if e.Type != "spoken" {
		t.Errorf("Type = %q, want spoken", e.Type)
	}
	if e.Char != "Guard" {
		t.Errorf("Char = %q, want Guard", e.Char)
	}
	if e.Scene != "guard_greeting" {
		t.Errorf("Scene = %q, want guard_greeting", e.Scene)
	}
	if e.Ctx != "guard is suspicious, not yet hostile" {
		t.Errorf("Ctx = %q", e.Ctx)
	}
	if e.Value != "Arrêtez !" {
		t.Errorf("Value = %q", e.Value)
	}
}

func TestParse_MetadataPartial(t *testing.T) {
	src := `[meta]
base_locale = en
locale = fr

[strings]
// type: spoken
guard_greeting:0:a1b2c3d4 = "Bonjour."
// scene: intro
narrator:0:11111111 = "Il était une fois."
`
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(sf.Entries) != 2 {
		t.Fatalf("entries len = %d, want 2", len(sf.Entries))
	}
	if sf.Entries[0].Type != "spoken" {
		t.Errorf("Entries[0].Type = %q, want spoken", sf.Entries[0].Type)
	}
	if sf.Entries[0].Scene != "" {
		t.Errorf("Entries[0].Scene = %q, want empty", sf.Entries[0].Scene)
	}
	if sf.Entries[1].Scene != "intro" {
		t.Errorf("Entries[1].Scene = %q, want intro", sf.Entries[1].Scene)
	}
	if sf.Entries[1].Type != "" {
		t.Errorf("Entries[1].Type = %q, want empty", sf.Entries[1].Type)
	}
}

// --------------------------------------------------------------------------
// Write / round-trip
// --------------------------------------------------------------------------

func TestWrite_RoundTrip(t *testing.T) {
	sf, err := loc.Parse("test.agstrings", minimalSrc)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	out := loc.Write(sf)
	sf2, err := loc.Parse("test.agstrings", out)
	if err != nil {
		t.Fatalf("Parse round-trip: %v", err)
	}
	if sf2.Meta.Locale != sf.Meta.Locale {
		t.Errorf("Locale mismatch after round-trip")
	}
	if len(sf2.Entries) != len(sf.Entries) {
		t.Errorf("Entries len mismatch: %d vs %d", len(sf2.Entries), len(sf.Entries))
	}
	if sf2.Entries[0].Value != sf.Entries[0].Value {
		t.Errorf("Entry value mismatch: %q vs %q", sf2.Entries[0].Value, sf.Entries[0].Value)
	}
}

func TestWrite_ContainsMeta(t *testing.T) {
	sf := &loc.StringsFile{Meta: loc.Meta{BaseLocale: "en", Locale: "fr"}}
	out := loc.Write(sf)
	if !strings.Contains(out, "locale         = fr") {
		t.Errorf("Write missing locale line: %s", out)
	}
	if !strings.Contains(out, "[meta]") {
		t.Errorf("Write missing [meta] section: %s", out)
	}
}

func TestWrite_StaleEntryCommented(t *testing.T) {
	sf := &loc.StringsFile{
		Meta:    loc.Meta{BaseLocale: "en", Locale: "fr"},
		Entries: []loc.Entry{{Key: "k", Value: "v", Stale: true}},
	}
	out := loc.Write(sf)
	if !strings.Contains(out, "// [stale]") {
		t.Errorf("Write missing stale comment: %s", out)
	}
}

func TestWrite_OrphanEntryCommented(t *testing.T) {
	sf := &loc.StringsFile{
		Meta:    loc.Meta{BaseLocale: "en", Locale: "fr"},
		Entries: []loc.Entry{{Key: "k", Value: "v", Orphan: true}},
	}
	out := loc.Write(sf)
	if !strings.Contains(out, "// [orphan]") {
		t.Errorf("Write missing orphan comment: %s", out)
	}
}

func TestWrite_MetadataComments(t *testing.T) {
	sf := &loc.StringsFile{
		Meta: loc.Meta{BaseLocale: "en", Locale: "fr"},
		Entries: []loc.Entry{
			{Key: "k1", Value: "v1", Type: "spoken", Char: "Guard", Scene: "greet", Ctx: "suspicious"},
			{Key: "k2", Value: "v2", Type: "narration"},
		},
	}
	out := loc.Write(sf)
	if !strings.Contains(out, "// type: spoken") {
		t.Errorf("Write missing type comment: %s", out)
	}
	if !strings.Contains(out, "// char: Guard") {
		t.Errorf("Write missing char comment: %s", out)
	}
	if !strings.Contains(out, "// scene: greet") {
		t.Errorf("Write missing scene comment: %s", out)
	}
	if !strings.Contains(out, "// ctx: suspicious") {
		t.Errorf("Write missing ctx comment: %s", out)
	}
	if !strings.Contains(out, "// type: narration") {
		t.Errorf("Write missing narration type: %s", out)
	}
}

func TestWrite_MetadataRoundTrip(t *testing.T) {
	sf := &loc.StringsFile{
		Meta: loc.Meta{BaseLocale: "en", Locale: "fr"},
		Entries: []loc.Entry{
			{Key: "k1", Value: "v1", Type: "spoken", Char: "Guard", Scene: "greet", Ctx: "suspicious"},
		},
	}
	out := loc.Write(sf)
	sf2, err := loc.Parse("test.agstrings", out)
	if err != nil {
		t.Fatalf("Parse after Write: %v\n%s", err, out)
	}
	if len(sf2.Entries) != 1 {
		t.Fatalf("Entries len = %d, want 1", len(sf2.Entries))
	}
	e := sf2.Entries[0]
	if e.Type != "spoken" {
		t.Errorf("Type = %q, want spoken", e.Type)
	}
	if e.Char != "Guard" {
		t.Errorf("Char = %q, want Guard", e.Char)
	}
	if e.Scene != "greet" {
		t.Errorf("Scene = %q, want greet", e.Scene)
	}
	if e.Ctx != "suspicious" {
		t.Errorf("Ctx = %q, want suspicious", e.Ctx)
	}
}

// --------------------------------------------------------------------------
// Diff / Apply
// --------------------------------------------------------------------------

func mustParse(t *testing.T, src string) *loc.StringsFile {
	t.Helper()
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return sf
}

func TestDiff_Unchanged(t *testing.T) {
	base := mustParse(t, "[meta]\nbase_locale=en\nlocale=fr\n\n[strings]\nk:a = \"v\"\n")
	diff := loc.Diff(base, []string{"k:a"})
	if len(diff) != 1 || diff[0].Kind != loc.DiffUnchanged {
		t.Errorf("expected 1 Unchanged, got %+v", diff)
	}
}

func TestDiff_Added(t *testing.T) {
	base := mustParse(t, "[meta]\nbase_locale=en\nlocale=fr\n\n[strings]\n")
	diff := loc.Diff(base, []string{"new:key:abc"})
	if len(diff) != 1 || diff[0].Kind != loc.DiffAdded {
		t.Errorf("expected 1 Added, got %+v", diff)
	}
}

func TestDiff_Removed(t *testing.T) {
	base := mustParse(t, "[meta]\nbase_locale=en\nlocale=fr\n\n[strings]\nold:key:abc = \"old\"\n")
	diff := loc.Diff(base, []string{})
	if len(diff) != 1 || diff[0].Kind != loc.DiffRemoved {
		t.Errorf("expected 1 Removed, got %+v", diff)
	}
}

func TestApply_NewKeyAdded(t *testing.T) {
	base := mustParse(t, "[meta]\nbase_locale=en\nlocale=fr\n\n[strings]\n")
	diff := loc.Diff(base, []string{"new:key:abc"})
	updated := loc.Apply(base, diff)
	if len(updated.Entries) != 1 || updated.Entries[0].Key != "new:key:abc" {
		t.Errorf("Apply: expected new key, got %+v", updated.Entries)
	}
	if updated.Entries[0].Value != "" {
		t.Errorf("Apply: new key should have empty value, got %q", updated.Entries[0].Value)
	}
}

func TestApply_RemovedMarkedOrphan(t *testing.T) {
	base := mustParse(t, "[meta]\nbase_locale=en\nlocale=fr\n\n[strings]\nold:key = \"v\"\n")
	diff := loc.Diff(base, []string{})
	updated := loc.Apply(base, diff)
	if len(updated.Entries) != 1 || !updated.Entries[0].Orphan {
		t.Errorf("Apply: expected orphan entry, got %+v", updated.Entries)
	}
}

func TestApply_ExistingTranslationPreserved(t *testing.T) {
	base := mustParse(t, "[meta]\nbase_locale=en\nlocale=fr\n\n[strings]\nk:a = \"Bonjour.\"\n")
	diff := loc.Diff(base, []string{"k:a"})
	updated := loc.Apply(base, diff)
	if updated.Entries[0].Value != "Bonjour." {
		t.Errorf("Apply: translation lost, got %q", updated.Entries[0].Value)
	}
}

func TestApply_RoundTripWriteParse(t *testing.T) {
	base := mustParse(t, "[meta]\nbase_locale=en\nlocale=fr\n\n[strings]\nk:a = \"v\"\n")
	diff := loc.Diff(base, []string{"k:a", "k:b"})
	updated := loc.Apply(base, diff)
	written := loc.Write(updated)
	reparsed, err := loc.Parse("test.agstrings", written)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(reparsed.Entries) != 2 {
		t.Errorf("expected 2 entries after round-trip, got %d", len(reparsed.Entries))
	}
}

// --------------------------------------------------------------------------
// ExportLocale
// --------------------------------------------------------------------------

func TestExportLocale_DialogueEntries(t *testing.T) {
	tmp := t.TempDir()

	subDir := tmp + "/dialogue"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gameAGP := `name = "test"
start_room = "room1"
start_character = "player"
`
	if err := os.WriteFile(tmp+"/game.agp", []byte(gameAGP), 0644); err != nil {
		t.Fatalf("write game.agp: %v", err)
	}

	dlgSrc := `title: test_node
character: guard
---
Guard: Hello there. #loc:guard_hello
<<end>>
===
`
	if err := os.WriteFile(subDir+"/test.agdlg", []byte(dlgSrc), 0644); err != nil {
		t.Fatalf("write agdlg: %v", err)
	}

	sf, err := loc.ExportLocale(tmp, "fr")
	if err != nil {
		t.Fatalf("ExportLocale: %v", err)
	}
	if sf.Meta.Locale != "fr" {
		t.Errorf("Meta.Locale = %q, want fr", sf.Meta.Locale)
	}
	if len(sf.Entries) == 0 {
		t.Error("expected at least one entry")
	}
}

func TestExportLocale_CutsceneEntries(t *testing.T) {
	tmp := t.TempDir()

	cutDir := tmp + "/cutscenes"
	if err := os.MkdirAll(cutDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gameAGP := `name = "test"
start_room = "room1"
start_character = "player"
`
	if err := os.WriteFile(tmp+"/game.agp", []byte(gameAGP), 0644); err != nil {
		t.Fatalf("write game.agp: %v", err)
	}

	cutSrc := `title: intro
sequence:
<<line narrator "Once upon a time." #loc:narrator_intro>>
<<end>>
`
	if err := os.WriteFile(cutDir+"/intro.agcut", []byte(cutSrc), 0644); err != nil {
		t.Fatalf("write agcut: %v", err)
	}

	sf, err := loc.ExportLocale(tmp, "de")
	if err != nil {
		t.Fatalf("ExportLocale: %v", err)
	}
	if sf.Meta.Locale != "de" {
		t.Errorf("Meta.Locale = %q, want de", sf.Meta.Locale)
	}
	found := false
	for _, e := range sf.Entries {
		if e.Key == "narrator_intro" {
			found = true
			if e.Value != "Once upon a time." {
				t.Errorf("entry value = %q, want %q", e.Value, "Once upon a time.")
			}
		}
	}
	if !found {
		t.Errorf("expected narrator_intro key, got %+v", sf.Entries)
	}
}

func TestExportLocale_DeduplicatesKeys(t *testing.T) {
	tmp := t.TempDir()

	subDir := tmp + "/dialogue"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cutDir := tmp + "/cutscenes"
	if err := os.MkdirAll(cutDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gameAGP := `name = "test"
start_room = "room1"
start_character = "player"
`
	if err := os.WriteFile(tmp+"/game.agp", []byte(gameAGP), 0644); err != nil {
		t.Fatalf("write game.agp: %v", err)
	}

	dlgSrc := `title: test_node
character: guard
---
Guard: Hello. #loc:shared_key
<<end>>
===
`
	if err := os.WriteFile(subDir+"/test.agdlg", []byte(dlgSrc), 0644); err != nil {
		t.Fatalf("write agdlg: %v", err)
	}

	cutSrc := `title: intro
sequence:
<<line narrator "Hello." #loc:shared_key>>
<<end>>
`
	if err := os.WriteFile(cutDir+"/intro.agcut", []byte(cutSrc), 0644); err != nil {
		t.Fatalf("write agcut: %v", err)
	}

	sf, err := loc.ExportLocale(tmp, "fr")
	if err != nil {
		t.Fatalf("ExportLocale: %v", err)
	}

	count := 0
	for _, e := range sf.Entries {
		if e.Key == "shared_key" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 entry for shared_key, got %d", count)
	}
}

// --------------------------------------------------------------------------
// FilterLocaleEntries / FindLocaleEntries / GroupLocaleEntries (T-LOC12)
// --------------------------------------------------------------------------

func TestFilterLocaleEntries_Untranslated(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Source: "Hello", Translated: "Bonjour"},
		{LocKey: "b", Source: "Goodbye", Translated: ""},
	}
	filtered := loc.FilterLocaleEntries(entries, loc.FilterOptions{Untranslated: true})
	if len(filtered) != 1 {
		t.Errorf("len = %d, want 1", len(filtered))
	}
	if filtered[0].LocKey != "b" {
		t.Errorf("key = %q, want b", filtered[0].LocKey)
	}
}

func TestFilterLocaleEntries_ByCharacter(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Character: "Guard"},
		{LocKey: "b", Character: "Player"},
		{LocKey: "c", Character: "Guard"},
	}
	filtered := loc.FilterLocaleEntries(entries, loc.FilterOptions{Char: "Guard"})
	if len(filtered) != 2 {
		t.Errorf("len = %d, want 2", len(filtered))
	}
}

func TestFilterLocaleEntries_ByType(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", LineType: "spoken"},
		{LocKey: "b", LineType: "narration"},
	}
	filtered := loc.FilterLocaleEntries(entries, loc.FilterOptions{Type: "narration"})
	if len(filtered) != 1 {
		t.Errorf("len = %d, want 1", len(filtered))
	}
}

func TestFilterLocaleEntries_ByNode(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", NodeTitle: "guard_greeting"},
		{LocKey: "b", NodeTitle: "player_intro"},
	}
	filtered := loc.FilterLocaleEntries(entries, loc.FilterOptions{Node: "guard_greeting"})
	if len(filtered) != 1 {
		t.Errorf("len = %d, want 1", len(filtered))
	}
}

func TestFilterLocaleEntries_AllCriteria(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Character: "Guard", LineType: "spoken", Translated: ""},
		{LocKey: "b", Character: "Guard", LineType: "narration", Translated: ""},
		{LocKey: "c", Character: "Guard", LineType: "spoken", Translated: "Bonjour"},
	}
	filtered := loc.FilterLocaleEntries(entries, loc.FilterOptions{
		Char:         "Guard",
		Type:         "spoken",
		Untranslated: true,
	})
	if len(filtered) != 1 {
		t.Errorf("len = %d, want 1", len(filtered))
	}
	if filtered[0].LocKey != "a" {
		t.Errorf("key = %q, want a", filtered[0].LocKey)
	}
}

func TestFindLocaleEntries_Wildcard(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "guard_greeting:line0", Source: "Hello"},
		{LocKey: "player_intro:line0", Source: "Bye"},
	}
	found := loc.FindLocaleEntries(entries, "guard_*")
	if len(found) != 1 {
		t.Errorf("len = %d, want 1", len(found))
	}
	if found[0].LocKey != "guard_greeting:line0" {
		t.Errorf("key = %q", found[0].LocKey)
	}
}

func TestFindLocaleEntries_PrefixGlob(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "guard_greeting:line0", Source: "Hello"},
		{LocKey: "guard_farewell:line0", Source: "Bye"},
	}
	found := loc.FindLocaleEntries(entries, "*farewell*")
	if len(found) != 1 {
		t.Errorf("len = %d, want 1", len(found))
	}
}

func TestFindLocaleEntries_SuffixGlob(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "guard_greeting:line0", Source: "Hello"},
		{LocKey: "player_intro:line0", Source: "Hello"},
	}
	found := loc.FindLocaleEntries(entries, "*:line0")
	if len(found) != 2 {
		t.Errorf("len = %d, want 2", len(found))
	}
}

func TestFindLocaleEntries_SourceMatch(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Source: "Hello world"},
		{LocKey: "b", Source: "Goodbye world"},
	}
	found := loc.FindLocaleEntries(entries, "*world*")
	if len(found) != 2 {
		t.Errorf("len = %d, want 2", len(found))
	}
}

func TestGroupLocaleEntries_ByCharacter(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Character: "Guard"},
		{LocKey: "b", Character: "Player"},
		{LocKey: "c", Character: "Guard"},
	}
	groups := loc.GroupLocaleEntries(entries, "character")
	if len(groups) != 2 {
		t.Errorf("group count = %d, want 2 (Guard, Player)", len(groups))
	}
	if len(groups["Guard"]) != 2 {
		t.Errorf("Guard group len = %d, want 2", len(groups["Guard"]))
	}
}

func TestGroupLocaleEntries_ByNode(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", NodeTitle: "guard_greeting"},
		{LocKey: "b", NodeTitle: "guard_greeting"},
		{LocKey: "c", NodeTitle: "player_intro"},
	}
	groups := loc.GroupLocaleEntries(entries, "node")
	if len(groups["guard_greeting"]) != 2 {
		t.Errorf("guard_greeting len = %d, want 2", len(groups["guard_greeting"]))
	}
}

func TestGroupLocaleEntries_NoCharacter(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Character: ""},
		{LocKey: "b", Character: "Guard"},
	}
	groups := loc.GroupLocaleEntries(entries, "character")
	if _, ok := groups["(no character)"]; !ok {
		t.Errorf("missing (no character) group")
	}
}

func TestFormatLocaleFind_Grouped(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "guard_greeting:line0", Source: "Hello", NodeTitle: "greet", Character: "Guard", LineType: "spoken"},
		{LocKey: "guard_farewell:line0", Source: "Bye", NodeTitle: "farewell", Character: "Guard", LineType: "spoken"},
	}
	out := loc.FormatLocaleFind(entries, "character")
	if !strings.Contains(out, "## Guard") {
		t.Errorf("expected grouped output, got: %s", out[:200])
	}
	if !strings.Contains(out, "Hello") {
		t.Errorf("expected source text in output")
	}
}

func TestFormatLocaleFind_Ungrouped(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "guard_greeting:line0", Source: "Hello", Translated: "Bonjour", NodeTitle: "greet", Character: "Guard", LineType: "spoken"},
	}
	out := loc.FormatLocaleFind(entries, "")
	if !strings.Contains(out, "guard_greeting:line0") {
		t.Errorf("expected loc_key in output, got: %s", out)
	}
	if !strings.Contains(out, "Bonjour") {
		t.Errorf("expected translation in output")
	}
}

func TestFormatLocaleFind_Untranslated(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Source: "Hello", Translated: "", NodeTitle: "n", Character: "", LineType: "narration"},
	}
	out := loc.FormatLocaleFind(entries, "")
	if !strings.Contains(out, "(untranslated)") {
		t.Errorf("expected (untranslated) marker, got: %s", out)
	}
}

func TestFormatLocaleReport_ShowsTranslation(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Source: "Hello", Translated: "Bonjour", NodeTitle: "n", Character: "Guard", LineType: "spoken"},
	}
	out := loc.FormatLocaleReport(entries, "fr")
	if !strings.Contains(out, "Bonjour") {
		t.Errorf("expected translated text in report, got: %s", out)
	}
	if !strings.Contains(out, "translated: true") {
		t.Errorf("expected translated: true flag in report")
	}
}

func TestFormatLocaleReport_ShowsUntranslated(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Source: "Hello", Translated: "", NodeTitle: "n", Character: "Guard", LineType: "spoken"},
	}
	out := loc.FormatLocaleReport(entries, "fr")
	if !strings.Contains(out, "Hello") {
		t.Errorf("expected source text when untranslated, got: %s", out)
	}
	if !strings.Contains(out, "translated: false") {
		t.Errorf("expected translated: false flag in report")
	}
}

func TestFormatLocaleReportGrouped_ByCharacter(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Source: "Hello", Translated: "Bonjour", NodeTitle: "n", Character: "Guard", LineType: "spoken"},
		{LocKey: "b", Source: "Bye", Translated: "", NodeTitle: "n", Character: "Guard", LineType: "spoken"},
	}
	out := loc.FormatLocaleReportGrouped(entries, "character", "fr")
	if !strings.Contains(out, "## Guard") {
		t.Errorf("expected group header ## Guard, got: %s", out[:200])
	}
	if !strings.Contains(out, "Bonjour") {
		t.Errorf("expected translated text in group")
	}
}

func TestFormatLocaleReportGrouped_ByNode(t *testing.T) {
	entries := []loc.LocaleEntryFull{
		{LocKey: "a", Source: "Hello", Translated: "", NodeTitle: "greeting", Character: "Guard", LineType: "spoken"},
		{LocKey: "b", Source: "Bye", Translated: "", NodeTitle: "farewell", Character: "Guard", LineType: "spoken"},
	}
	out := loc.FormatLocaleReportGrouped(entries, "node", "fr")
	if !strings.Contains(out, "## greeting") {
		t.Errorf("expected group header ## greeting, got: %s", out[:200])
	}
	if !strings.Contains(out, "## farewell") {
		t.Errorf("expected group header ## farewell, got: %s", out[:200])
	}
}

// ── testdata/locale/ fixture tests (T-LOC02) ─────────────────────────────────

func TestParse_ValidFixtures(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "testdata", "locale", "valid", "*.agstrings"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no locale/valid fixtures found")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			_, err = loc.Parse(path, string(data))
			if err != nil {
				t.Errorf("unexpected parse error: %v", err)
			}
		})
	}
}

func TestParse_InvalidFixtures(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "testdata", "locale", "invalid", "*.agstrings"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no locale/invalid fixtures found")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			_, err = loc.Parse(path, string(data))
			if err == nil {
				t.Errorf("expected parse error for %q, got none", filepath.Base(path))
			}
		})
	}
}
