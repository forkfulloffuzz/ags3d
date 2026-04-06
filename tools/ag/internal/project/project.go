// Package project handles game.agp parsing, project directory scanning,
// and the incremental build manifest (.engine/cache/build_manifest.json).
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Manifest is the parsed content of game.agp.
type Manifest struct {
	Project       ProjectSection      `toml:"project"`
	Settings      SettingsSection     `toml:"settings"`
	Localisation  LocalisationSection `toml:"localisation"`
	Cutscenes     CutsceneSection     `toml:"cutscenes"`
	Input         InputSection        `toml:"input"`
	// Locales maps BCP 47 locale codes to their declarations.
	// Populated from [locale.xx] subsections in game.agp.
	Locales map[string]*LocaleEntry `toml:"locales"`
	// Globals holds the user-defined global variables from the [globals] section.
	// Key: variable name, Value: default value string (e.g. "0", "false", `""`).
	Globals map[string]string `toml:"globals"`
	// Root is the directory containing game.agp (set after parsing).
	Root string `toml:"-"`
}

// CutsceneSection holds project-wide cutscene defaults from [cutscenes].
type CutsceneSection struct {
	// FallbackDebug is the default fallback policy during debug builds.
	// One of: "halt", "skip_and_continue", "log_and_continue", "retry_once".
	FallbackDebug string
	// FallbackRelease is the default fallback policy for release builds.
	FallbackRelease string
	// FallbackQA is the default fallback policy for QA builds.
	FallbackQA string
	// StepTimeoutDefault is the default step timeout in seconds (0 = no timeout).
	StepTimeoutDefault float64
}

// InputSection holds input action bindings for dialogue and cutscenes from [input].
type InputSection struct {
	// DialogueAdvance is the input action name for advancing dialogue.
	DialogueAdvance string
	// CutsceneSkip is the input action name for skipping a cutscene.
	CutsceneSkip string
	// DialogueHoldAdvance is the input action name for hold-to-rapid-advance dialogue.
	DialogueHoldAdvance string
}

// LocaleEntry declares a supported locale for the project.
type LocaleEntry struct {
	Code string // BCP 47 code, e.g. "en", "fr", "zh-TW" (set from section name)
	Name string // Human-readable display name, e.g. "English"
	RTL  bool   // True for right-to-left scripts (Arabic, Hebrew, etc.)
}

// LocalisationSection holds project-wide localisation settings from [localisation].
type LocalisationSection struct {
	BaseLocale    string   // Locale code used for authoring (source strings)
	FallbackChain []string // Ordered list of locale codes tried when a string is missing
}

// localeCodeRE matches valid BCP 47 locale codes: "en", "fr", "zh-TW", "sr-Latn-RS".
var localeCodeRE = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{1,8})*$`)

// ValidateLocales checks that all locale codes are valid BCP 47, that base_locale
// exists in the declared locales, and that every entry in fallback_chain is declared.
// Returns a slice of human-readable error strings (empty if valid).
func (m *Manifest) ValidateLocales() []string {
	var errs []string
	for code := range m.Locales {
		if !localeCodeRE.MatchString(code) {
			errs = append(errs, fmt.Sprintf("invalid locale code %q (expected BCP 47, e.g. \"en\", \"zh-TW\")", code))
		}
	}
	if bl := m.Localisation.BaseLocale; bl != "" {
		if _, ok := m.Locales[bl]; !ok {
			errs = append(errs, fmt.Sprintf("base_locale %q is not declared in any [locale.*] section", bl))
		}
	}
	for _, code := range m.Localisation.FallbackChain {
		if _, ok := m.Locales[code]; !ok {
			errs = append(errs, fmt.Sprintf("fallback_chain entry %q is not declared in any [locale.*] section", code))
		}
	}
	return errs
}

type ProjectSection struct {
	Name           string `toml:"name"`
	StartRoom      string `toml:"start_room"`
	StartCharacter string `toml:"start_character"`
}

type SettingsSection struct {
	RenderingMode string `toml:"rendering_mode"`
	Autosave      bool   `toml:"autosave"`
}

// SourceFile represents a discovered adventure-game source file.
type SourceFile struct {
	Path string // absolute path
	Rel  string // relative to project root
	Ext  string // .agscript | .agroom | .agchar | .agitem | .agdlg
}

// sourceExtensions lists all file extensions the scanner recognises.
var sourceExtensions = map[string]bool{
	".agscript": true,
	".agroom":   true,
	".agchar":   true,
	".agitem":   true,
	".agdlg":    true,
}

// Find walks upward from start looking for game.agp and returns its directory.
// Returns ("", false) if no game.agp is found.
func Find(start string) (string, bool) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, "game.agp")); err == nil {
			return abs, true
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", false
		}
		abs = parent
	}
}

// Load reads and parses game.agp from the given project root directory.
// It uses a minimal hand-rolled TOML parser sufficient for the game.agp schema.
func Load(root string) (*Manifest, error) {
	path := filepath.Join(root, "game.agp")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	m := &Manifest{Root: root}
	if err := parseAGP(string(data), m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return m, nil
}

// parseAGP is a minimal TOML parser for the game.agp schema.
func parseAGP(src string, m *Manifest) error {
	section := ""
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"`)
		switch {
		case section == "project":
			switch k {
			case "name":
				m.Project.Name = v
			case "start_room":
				m.Project.StartRoom = v
			case "start_character":
				m.Project.StartCharacter = v
			}
		case section == "settings":
			switch k {
			case "rendering_mode":
				m.Settings.RenderingMode = v
			case "autosave":
				m.Settings.Autosave = v == "true"
			}
		case section == "globals":
			if m.Globals == nil {
				m.Globals = make(map[string]string)
			}
			m.Globals[k] = v
		case section == "cutscenes":
			switch k {
			case "fallback_debug":
				m.Cutscenes.FallbackDebug = v
			case "fallback_release":
				m.Cutscenes.FallbackRelease = v
			case "fallback_qa":
				m.Cutscenes.FallbackQA = v
			case "step_timeout_default":
				var f float64
				_, _ = fmt.Sscanf(v, "%f", &f)
				m.Cutscenes.StepTimeoutDefault = f
			}
		case section == "input":
			switch k {
			case "dialogue_advance":
				m.Input.DialogueAdvance = v
			case "cutscene_skip":
				m.Input.CutsceneSkip = v
			case "dialogue_hold_advance":
				m.Input.DialogueHoldAdvance = v
			}
		case strings.HasPrefix(section, "locale."):
			// [locale.en], [locale.fr], etc.
			code := section[len("locale."):]
			if m.Locales == nil {
				m.Locales = make(map[string]*LocaleEntry)
			}
			entry := m.Locales[code]
			if entry == nil {
				entry = &LocaleEntry{Code: code}
				m.Locales[code] = entry
			}
			switch k {
			case "name":
				entry.Name = v
			case "rtl":
				entry.RTL = v == "true"
			}
		case section == "localisation":
			switch k {
			case "base_locale":
				m.Localisation.BaseLocale = v
			case "fallback_chain":
				// Comma-separated list, e.g. fallback_chain = "en, fr"
				for _, part := range strings.Split(v, ",") {
					if code := strings.TrimSpace(part); code != "" {
						m.Localisation.FallbackChain = append(m.Localisation.FallbackChain, code)
					}
				}
			}
		}
	}
	return nil
}

// Scan returns all adventure-game source files under root, excluding .engine/.
func Scan(root string) ([]SourceFile, error) {
	var files []SourceFile
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".engine" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if !sourceExtensions[ext] {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		files = append(files, SourceFile{Path: path, Rel: rel, Ext: ext})
		return nil
	})
	return files, err
}

// BuildManifest tracks per-file mtimes for incremental builds.
type BuildManifest map[string]float64

// LoadManifest reads the build manifest from .engine/cache/build_manifest.json.
// Returns an empty manifest if the file does not exist.
func LoadManifest(root string) (BuildManifest, error) {
	path := filepath.Join(root, ".engine", "cache", "build_manifest.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(BuildManifest), nil
	}
	if err != nil {
		return nil, err
	}
	m := make(BuildManifest)
	if err := json.Unmarshal(data, &m); err != nil {
		return make(BuildManifest), nil // treat corrupt manifest as empty
	}
	return m, nil
}

// SaveManifest writes the build manifest to disk.
func SaveManifest(root string, m BuildManifest) error {
	dir := filepath.Join(root, ".engine", "cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "build_manifest.json"), data, 0644)
}

// Changed returns the subset of files whose mtime differs from the manifest.
func Changed(files []SourceFile, manifest BuildManifest) []SourceFile {
	var changed []SourceFile
	for _, f := range files {
		info, err := os.Stat(f.Path)
		if err != nil {
			continue
		}
		mtime := float64(info.ModTime().UnixNano()) / 1e9
		if manifest[f.Path] != mtime {
			changed = append(changed, f)
		}
	}
	return changed
}

// RecordMtimes updates the manifest with current mtimes for the given files.
func RecordMtimes(files []SourceFile, manifest BuildManifest) {
	for _, f := range files {
		info, err := os.Stat(f.Path)
		if err != nil {
			continue
		}
		manifest[f.Path] = float64(info.ModTime().UnixNano()) / 1e9
	}
}

// Scaffold creates a new project directory with the standard layout and game.agp.
func Scaffold(dest, name string) error {
	dirs := []string{
		"characters", "rooms/start", "dialogue",
		"inventory", "scripts", "audio", "assets",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dest, d), 0755); err != nil {
			return err
		}
	}

	agp := fmt.Sprintf(`[project]
name = %q
start_room = "rooms/start/start.agroom"
start_character = "characters/player.agchar"

[settings]
rendering_mode = "full_3d"
autosave = true
`, name)
	if err := os.WriteFile(filepath.Join(dest, "game.agp"), []byte(agp), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dest, ".gitignore"), []byte(".engine/\n"), 0644); err != nil {
		return err
	}

	room := "Room \"start\" {\n    // Starter room — edit me\n}\n"
	if err := os.WriteFile(filepath.Join(dest, "rooms/start/start.agroom"), []byte(room), 0644); err != nil {
		return err
	}

	script := "// start.agscript — room script for 'start'\n\nfunction room_Load() {\n    // Called when the room loads\n}\n"
	if err := os.WriteFile(filepath.Join(dest, "rooms/start/start.agscript"), []byte(script), 0644); err != nil {
		return err
	}

	char := "Character \"player\" {\n    display_name = \"Player\"\n}\n"
	if err := os.WriteFile(filepath.Join(dest, "characters/player.agchar"), []byte(char), 0644); err != nil {
		return err
	}

	return nil
}
