package loc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ags3d/ag/internal/loc"
)

func TestExportLocaleForProject_MultiLocale(t *testing.T) {
	dir := t.TempDir()

	gameAGP := `[project]
name = "Test"
start_room = "rooms/start.agroom"
start_character = "characters/player.agchar"

[locale.en]
name = "English"

[locale.fr]
name = "French"

[locale.de]
name = "German"

[localisation]
default_author_locale = "en"
supported_locales = "en fr de"
fallback_chain = "en"
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(gameAGP), 0644); err != nil {
		t.Fatalf("write game.agp: %v", err)
	}

	dlgDir := filepath.Join(dir, "dialogue")
	if err := os.MkdirAll(dlgDir, 0755); err != nil {
		t.Fatalf("mkdir dialogue: %v", err)
	}

	dlgSrc := `title: greeting
language: en
---
Guard: Hello.
===
`
	if err := os.WriteFile(filepath.Join(dlgDir, "greeting.agdlg"), []byte(dlgSrc), 0644); err != nil {
		t.Fatalf("write greeting.agdlg: %v", err)
	}

	cutDir := filepath.Join(dir, "cutscenes")
	if err := os.MkdirAll(cutDir, 0755); err != nil {
		t.Fatalf("mkdir cutscenes: %v", err)
	}

	cutSrc := `title: intro
language: en
sequence:
<<line narrator "Once upon a time." #loc_key:intro_narration>>
<<end>>
`
	if err := os.WriteFile(filepath.Join(cutDir, "intro.agcut"), []byte(cutSrc), 0644); err != nil {
		t.Fatalf("write intro.agcut: %v", err)
	}

	localeDir := filepath.Join(dir, "locale")
	if err := os.MkdirAll(localeDir, 0755); err != nil {
		t.Fatalf("mkdir locale: %v", err)
	}

	written, err := loc.ExportLocaleForProject(dir)
	if err != nil {
		t.Fatalf("ExportLocaleForProject: %v", err)
	}
	if len(written) != 3 {
		t.Fatalf("expected 3 locale files, got %d: %v", len(written), written)
	}

	enPath := filepath.Join(localeDir, "strings.en.agstrings")
	frPath := filepath.Join(localeDir, "strings.fr.agstrings")
	dePath := filepath.Join(localeDir, "strings.de.agstrings")

	for _, path := range []string{enPath, frPath, dePath} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", path, err)
		}
	}

	enData, err := os.ReadFile(enPath)
	if err != nil {
		t.Fatalf("read en: %v", err)
	}
	enSF, err := loc.Parse(enPath, string(enData))
	if err != nil {
		t.Fatalf("parse en: %v", err)
	}

	if len(enSF.Entries) == 0 {
		t.Fatal("en agstrings should have entries")
	}
	if enSF.Entries[0].Value == "" {
		t.Error("en locale should have source text filled in")
	}

	frData, err := os.ReadFile(frPath)
	if err != nil {
		t.Fatalf("read fr: %v", err)
	}
	frSF, err := loc.Parse(frPath, string(frData))
	if err != nil {
		t.Fatalf("parse fr: %v", err)
	}
	if len(frSF.Entries) == 0 {
		t.Fatal("fr agstrings should have entries")
	}
	if frSF.Entries[0].Value != "" {
		t.Error("fr locale should have empty stub for untranslated entry")
	}
}

func TestExportLocaleForProject_SingleLocale(t *testing.T) {
	dir := t.TempDir()

	gameAGP := `[project]
name = "Test"

[locale.en]
name = "English"

[localisation]
default_author_locale = "en"
supported_locales = "en"
fallback_chain = "en"
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(gameAGP), 0644); err != nil {
		t.Fatalf("write game.agp: %v", err)
	}

	dlgDir := filepath.Join(dir, "dialogue")
	if err := os.MkdirAll(dlgDir, 0755); err != nil {
		t.Fatalf("mkdir dialogue: %v", err)
	}

	dlgSrc := `title: greeting
---
Guard: Hello.
===
`
	if err := os.WriteFile(filepath.Join(dlgDir, "greeting.agdlg"), []byte(dlgSrc), 0644); err != nil {
		t.Fatalf("write greeting.agdlg: %v", err)
	}

	localeDir := filepath.Join(dir, "locale")
	if err := os.MkdirAll(localeDir, 0755); err != nil {
		t.Fatalf("mkdir locale: %v", err)
	}

	written, err := loc.ExportLocaleForProject(dir)
	if err != nil {
		t.Fatalf("ExportLocaleForProject: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 locale file, got %d: %v", len(written), written)
	}

	enPath := filepath.Join(localeDir, "strings.en.agstrings")
	enData, err := os.ReadFile(enPath)
	if err != nil {
		t.Fatalf("read en: %v", err)
	}
	enSF, err := loc.Parse(enPath, string(enData))
	if err != nil {
		t.Fatalf("parse en: %v", err)
	}
	if len(enSF.Entries) == 0 {
		t.Fatal("en agstrings should have entries")
	}
	if enSF.Entries[0].Value == "" {
		t.Error("en locale should have source text filled in")
	}
}

func TestCollectAllLocaleEntries_HasSourceLanguage(t *testing.T) {
	dir := t.TempDir()

	gameAGP := `[project]
name = "Test"

[locale.en]
name = "English"

[locale.fr]
name = "French"

[localisation]
default_author_locale = "en"
supported_locales = "en fr"
fallback_chain = "en"
`
	if err := os.WriteFile(filepath.Join(dir, "game.agp"), []byte(gameAGP), 0644); err != nil {
		t.Fatalf("write game.agp: %v", err)
	}

	dlgDir := filepath.Join(dir, "dialogue")
	if err := os.MkdirAll(dlgDir, 0755); err != nil {
		t.Fatalf("mkdir dialogue: %v", err)
	}

	dlgSrc := `title: greeting
language: fr
---
Guard: Bonjour.
===
`
	if err := os.WriteFile(filepath.Join(dlgDir, "greeting.agdlg"), []byte(dlgSrc), 0644); err != nil {
		t.Fatalf("write greeting.agdlg: %v", err)
	}

	cutDir := filepath.Join(dir, "cutscenes")
	if err := os.MkdirAll(cutDir, 0755); err != nil {
		t.Fatalf("mkdir cutscenes: %v", err)
	}

	cutSrc := `title: intro
language: en
sequence:
<<line narrator "Once upon a time." #loc_key:intro_narration>>
<<end>>
`
	if err := os.WriteFile(filepath.Join(cutDir, "intro.agcut"), []byte(cutSrc), 0644); err != nil {
		t.Fatalf("write intro.agcut: %v", err)
	}

	entries, err := loc.CollectAllLocaleEntries(dir)
	if err != nil {
		t.Fatalf("CollectAllLocaleEntries: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	var hasFR, hasEN bool
	for _, e := range entries {
		if e.SourceLanguage == "fr" {
			hasFR = true
		}
		if e.SourceLanguage == "en" {
			hasEN = true
		}
	}
	if !hasFR {
		t.Error("expected at least one entry with SourceLanguage=fr")
	}
	if !hasEN {
		t.Error("expected at least one entry with SourceLanguage=en")
	}
}
