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
	LocKey    string
	Source    string
	NodeTitle string
	Character string
	LineType  string
	File      string
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
