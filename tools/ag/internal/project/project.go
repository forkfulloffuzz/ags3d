// Package project handles game.agp parsing, project directory scanning,
// and the incremental build manifest (.engine/cache/build_manifest.json).
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest is the parsed content of game.agp.
type Manifest struct {
	Project  ProjectSection  `toml:"project"`
	Settings SettingsSection `toml:"settings"`
	// Root is the directory containing game.agp (set after parsing).
	Root string `toml:"-"`
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
		switch section + "." + k {
		case "project.name":
			m.Project.Name = v
		case "project.start_room":
			m.Project.StartRoom = v
		case "project.start_character":
			m.Project.StartCharacter = v
		case "settings.rendering_mode":
			m.Settings.RenderingMode = v
		case "settings.autosave":
			m.Settings.Autosave = v == "true"
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
