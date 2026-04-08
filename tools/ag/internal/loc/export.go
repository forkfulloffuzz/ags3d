package loc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ags3d/ag/internal/cut"
	"github.com/ags3d/ag/internal/dlg"
	"github.com/ags3d/ag/internal/project"
)

func ExportLocale(root, locale string) (*StringsFile, error) {
	files, err := project.Scan(root)
	if err != nil {
		return nil, err
	}

	var dlgFiles []*dlg.DialogueFile
	var cutFiles []*cut.CutsceneFile

	for _, f := range files {
		switch f.Ext {
		case ".agdlg":
			df, err := dlg.ParseFile(f.Path)
			if err != nil {
				continue
			}
			dlgFiles = append(dlgFiles, df)
		}
	}

	if err := walkCutscenes(root, func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		cf, err := cut.Parse(rel, string(data))
		if err != nil {
			return nil
		}
		cutFiles = append(cutFiles, cf)
		return nil
	}); err != nil {
		return nil, err
	}

	sf := &StringsFile{
		Meta: Meta{
			BaseLocale: "en",
			Locale:     locale,
		},
		index: make(map[string]int),
	}

	if len(dlgFiles) > 0 {
		lp, err := dlg.Link(dlgFiles)
		if err == nil {
			for _, e := range dlg.CollectLocEntries(lp) {
				if e.LocKey == "" {
					continue
				}
				if _, exists := sf.index[e.LocKey]; !exists {
					sf.index[e.LocKey] = len(sf.Entries)
					sf.Entries = append(sf.Entries, Entry{
						Key:   e.LocKey,
						Value: e.Source,
					})
				}
			}
		}
	}

	for _, cf := range cutFiles {
		for _, e := range cut.CollectLocEntries(cf) {
			if e.LocKey == "" {
				continue
			}
			if _, exists := sf.index[e.LocKey]; !exists {
				sf.index[e.LocKey] = len(sf.Entries)
				sf.Entries = append(sf.Entries, Entry{
					Key:   e.LocKey,
					Value: e.Source,
				})
			}
		}
	}

	return sf, nil
}

func walkCutscenes(root string, fn func(path string) error) error {
	cutDir := filepath.Join(root, "cutscenes")
	info, err := os.Stat(cutDir)
	if err != nil || !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(cutDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".agcut" {
			return fn(path)
		}
		return nil
	})
}

func ExportLocaleFiles(root, locale string) ([]string, error) {
	sf, err := ExportLocale(root, locale)
	if err != nil {
		return nil, err
	}

	localeDir := filepath.Join(root, "locale")
	if err := os.MkdirAll(localeDir, 0755); err != nil {
		return nil, err
	}

	outPath := filepath.Join(localeDir, "strings."+locale+".agstrings")
	content := Write(sf)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return nil, err
	}

	rel, _ := filepath.Rel(root, outPath)
	return []string{rel}, nil
}

type LocaleEntryFull struct {
	LocKey     string
	Source     string
	Translated string // empty if untranslated
	NodeTitle  string
	Character  string
	LineType   string
	File       string
}

func CollectAllLocaleEntriesWithTranslations(root, locale string) ([]LocaleEntryFull, error) {
	entries, err := CollectAllLocaleEntries(root)
	if err != nil {
		return nil, err
	}

	localePath := findLocaleFile(root, locale)
	if localePath == "" {
		return entries, nil
	}

	data, err := os.ReadFile(localePath)
	if err != nil {
		return entries, nil
	}

	sf, err := Parse(localePath, string(data))
	if err != nil {
		return entries, nil
	}

	for i := range entries {
		entries[i].Translated = sf.Get(entries[i].LocKey)
	}

	return entries, nil
}

func CollectAllLocaleEntries(root string) ([]LocaleEntryFull, error) {
	files, err := project.Scan(root)
	if err != nil {
		return nil, err
	}

	var entries []LocaleEntryFull

	var dlgFiles []*dlg.DialogueFile
	var cutFiles []*cut.CutsceneFile

	for _, f := range files {
		switch f.Ext {
		case ".agdlg":
			df, err := dlg.ParseFile(f.Path)
			if err != nil {
				continue
			}
			dlgFiles = append(dlgFiles, df)
		}
	}

	if err := walkCutscenes(root, func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		cf, err := cut.Parse(rel, string(data))
		if err != nil {
			return nil
		}
		cutFiles = append(cutFiles, cf)
		return nil
	}); err != nil {
		return nil, err
	}

	if len(dlgFiles) > 0 {
		lp, err := dlg.Link(dlgFiles)
		if err == nil {
			for _, e := range dlg.CollectLocEntries(lp) {
				if e.LocKey == "" {
					continue
				}
				entries = append(entries, LocaleEntryFull{
					LocKey:    e.LocKey,
					Source:    e.Source,
					NodeTitle: e.NodeTitle,
					Character: e.Character,
					LineType:  e.LineType,
				})
			}
		}
	}

	for _, cf := range cutFiles {
		for _, e := range cut.CollectLocEntries(cf) {
			if e.LocKey == "" {
				continue
			}
			entries = append(entries, LocaleEntryFull{
				LocKey:    e.LocKey,
				Source:    e.Source,
				NodeTitle: cf.Title,
				Character: "",
				LineType:  e.CmdName,
			})
		}
	}

	return entries, nil
}

func FormatLocaleReport(entries []LocaleEntryFull, locale string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# AGS3D Locale Report — %s\n\n", locale)
	fmt.Fprintf(&sb, "[meta]\nbase_locale    = en\nlocale         = %s\n\n[strings]\n\n", locale)

	for _, e := range entries {
		fmt.Fprintf(&sb, "# file: %s | node: %s | char: %s | type: %s\n",
			e.LocKey, e.NodeTitle, e.Character, e.LineType)
		fmt.Fprintf(&sb, "%s = %q\n\n", e.LocKey, e.Source)
	}

	return sb.String()
}

type LocaleValidationIssue struct {
	Code    string
	LocKey  string
	Message string
}

func FindOrphanKeys(sf *StringsFile, usedKeys map[string]bool) []LocaleValidationIssue {
	var issues []LocaleValidationIssue
	for _, e := range sf.Entries {
		if e.Orphan {
			continue
		}
		if !usedKeys[e.Key] {
			issues = append(issues, LocaleValidationIssue{
				Code:    "LOC-W002",
				LocKey:  e.Key,
				Message: fmt.Sprintf("orphan loc_key %q (in locale file but never referenced in .agdlg or .agcut)", e.Key),
			})
		}
	}
	return issues
}

func CollectUsedLocKeys(root string) (map[string]bool, error) {
	usedKeys := make(map[string]bool)

	files, err := project.Scan(root)
	if err != nil {
		return nil, err
	}

	var dlgFiles []*dlg.DialogueFile

	for _, f := range files {
		switch f.Ext {
		case ".agdlg":
			df, err := dlg.ParseFile(f.Path)
			if err != nil {
				continue
			}
			dlgFiles = append(dlgFiles, df)
		}
	}

	if err := walkCutscenes(root, func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		cf, err := cut.Parse(rel, string(data))
		if err != nil {
			return nil
		}
		for _, e := range cut.CollectLocEntries(cf) {
			if e.LocKey != "" {
				usedKeys[e.LocKey] = true
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(dlgFiles) > 0 {
		lp, err := dlg.Link(dlgFiles)
		if err == nil {
			for _, e := range dlg.CollectLocEntries(lp) {
				if e.LocKey != "" {
					usedKeys[e.LocKey] = true
				}
			}
		}
	}

	return usedKeys, nil
}

type ImportResult struct {
	Imported int
	Skipped  int
	Invalid  int
	Errors   []string
}

func ImportPO(root, locale, poPath string) (*ImportResult, error) {
	data, err := os.ReadFile(poPath)
	if err != nil {
		return nil, fmt.Errorf("read PO file: %w", err)
	}

	translations := dlg.ParsePOTranslations(string(data))
	usedKeys, err := CollectUsedLocKeys(root)
	if err != nil {
		return nil, fmt.Errorf("collect used keys: %w", err)
	}

	result := &ImportResult{}
	localePath := findLocaleFile(root, locale)
	var sf *StringsFile
	if localePath != "" {
		existing, err := os.ReadFile(localePath)
		if err == nil {
			sf, err = Parse(localePath, string(existing))
			if err != nil {
				sf = nil
			}
		}
	}
	if sf == nil {
		sf = &StringsFile{
			Meta: Meta{
				BaseLocale: "en",
				Locale:     locale,
			},
			index: make(map[string]int),
		}
	}

	for key, translation := range translations {
		if !usedKeys[key] {
			result.Invalid++
			result.Errors = append(result.Errors, fmt.Sprintf("key %q not found in project", key))
			continue
		}
		if translation == "" {
			continue
		}
		if idx, ok := sf.index[key]; ok {
			sf.Entries[idx].Value = translation
		} else {
			sf.index[key] = len(sf.Entries)
			sf.Entries = append(sf.Entries, Entry{Key: key, Value: translation})
		}
		result.Imported++
	}

	if localePath == "" {
		localeDir := filepath.Join(root, "locale")
		if err := os.MkdirAll(localeDir, 0755); err != nil {
			return nil, fmt.Errorf("create locale dir: %w", err)
		}
		localePath = filepath.Join(localeDir, "strings."+locale+".agstrings")
	}

	if err := os.WriteFile(localePath, []byte(Write(sf)), 0644); err != nil {
		return nil, fmt.Errorf("write locale file: %w", err)
	}

	return result, nil
}

func ImportCSV(root, locale, csvPath string) (*ImportResult, error) {
	data, err := os.ReadFile(csvPath)
	if err != nil {
		return nil, fmt.Errorf("read CSV file: %w", err)
	}

	translations := dlg.ParseCSVTranslations(string(data))
	usedKeys, err := CollectUsedLocKeys(root)
	if err != nil {
		return nil, fmt.Errorf("collect used keys: %w", err)
	}

	result := &ImportResult{}
	localePath := findLocaleFile(root, locale)
	var sf *StringsFile
	if localePath != "" {
		existing, err := os.ReadFile(localePath)
		if err == nil {
			sf, err = Parse(localePath, string(existing))
			if err != nil {
				sf = nil
			}
		}
	}
	if sf == nil {
		sf = &StringsFile{
			Meta: Meta{
				BaseLocale: "en",
				Locale:     locale,
			},
			index: make(map[string]int),
		}
	}

	for key, translation := range translations {
		if !usedKeys[key] {
			result.Invalid++
			result.Errors = append(result.Errors, fmt.Sprintf("key %q not found in project", key))
			continue
		}
		if translation == "" {
			continue
		}
		if idx, ok := sf.index[key]; ok {
			sf.Entries[idx].Value = translation
		} else {
			sf.index[key] = len(sf.Entries)
			sf.Entries = append(sf.Entries, Entry{Key: key, Value: translation})
		}
		result.Imported++
	}

	if localePath == "" {
		localeDir := filepath.Join(root, "locale")
		if err := os.MkdirAll(localeDir, 0755); err != nil {
			return nil, fmt.Errorf("create locale dir: %w", err)
		}
		localePath = filepath.Join(localeDir, "strings."+locale+".agstrings")
	}

	if err := os.WriteFile(localePath, []byte(Write(sf)), 0644); err != nil {
		return nil, fmt.Errorf("write locale file: %w", err)
	}

	return result, nil
}

func findLocaleFile(root, locale string) string {
	patterns := []string{
		filepath.Join(root, "locale", "strings."+locale+".agstrings"),
		filepath.Join(root, "locale", locale+".agstrings"),
		filepath.Join(root, "locale", locale+".po"),
		filepath.Join(root, "locale", locale+".csv"),
	}
	for _, p := range patterns {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

type FilterOptions struct {
	Untranslated bool
	Type         string
	Char         string
	Node         string
}

func FilterLocaleEntries(entries []LocaleEntryFull, opts FilterOptions) []LocaleEntryFull {
	var result []LocaleEntryFull
	for _, e := range entries {
		if opts.Untranslated && e.Translated != "" {
			continue
		}
		if opts.Type != "" && e.LineType != opts.Type {
			continue
		}
		if opts.Char != "" && e.Character != opts.Char {
			continue
		}
		if opts.Node != "" && e.NodeTitle != opts.Node {
			continue
		}
		result = append(result, e)
	}
	return result
}

func FindLocaleEntries(entries []LocaleEntryFull, pattern string) []LocaleEntryFull {
	var result []LocaleEntryFull
	for _, e := range entries {
		if globMatch(e.LocKey, pattern) || globMatch(e.Source, pattern) {
			result = append(result, e)
		}
	}
	return result
}

func GroupLocaleEntries(entries []LocaleEntryFull, by string) map[string][]LocaleEntryFull {
	groups := make(map[string][]LocaleEntryFull)
	for _, e := range entries {
		var key string
		switch by {
		case "character":
			key = e.Character
			if key == "" {
				key = "(no character)"
			}
		case "node":
			key = e.NodeTitle
		case "type":
			key = e.LineType
		default:
			key = "(all)"
		}
		groups[key] = append(groups[key], e)
	}
	return groups
}

func globMatch(s, pattern string) bool {
	if pattern == "" {
		return true
	}
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(s, strings.Trim(pattern, "*"))
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(s, strings.Trim(pattern, "*"))
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(s, strings.Trim(pattern, "*"))
	}
	return s == pattern
}

func FormatLocaleFind(entries []LocaleEntryFull, groupBy string) string {
	if groupBy == "" {
		return formatEntriesList(entries)
	}
	groups := GroupLocaleEntries(entries, groupBy)
	var sb strings.Builder
	for key, group := range groups {
		fmt.Fprintf(&sb, "## %s (%d)\n\n", key, len(group))
		sb.WriteString(formatEntriesList(group))
		sb.WriteString("\n")
	}
	return sb.String()
}

func formatEntriesList(entries []LocaleEntryFull) string {
	var sb strings.Builder
	for _, e := range entries {
		translation := e.Translated
		if translation == "" {
			translation = "(untranslated)"
		}
		fmt.Fprintf(&sb, "%s  [%s | %s | %s]\n  %s → %s\n\n",
			e.LocKey, e.NodeTitle, e.Character, e.LineType, e.Source, translation)
	}
	return sb.String()
}
