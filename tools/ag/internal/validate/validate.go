// Package validate provides project-wide cross-reference checks for ag validate.
//
// Checks performed:
//
//  1. game.agp: start_room file exists
//  2. game.agp: start_character file exists
//  3. .agroom: initial_camera names a Camera block defined in the same room
//  4. .agroom: each SpawnPoint.character matches a known .agchar name
//
// Script-level checks (point names, character names referenced in .agscript)
// are deferred until the AGS-spirit analysis package is fully implemented.
package validate

import (
	"fmt"
	"os"

	"github.com/ags3d/ag/internal/char"
	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/room"
)

// Issue is a single validation finding with source location.
type Issue struct {
	File     string
	Severity string // "error" | "warning"
	Message  string
}

func (i Issue) String() string {
	return fmt.Sprintf("%s: %s: %s", i.File, i.Severity, i.Message)
}

// ValidateProject runs all cross-reference checks against the project at root.
// It returns a slice of Issues (may be empty if the project is clean) and any
// fatal infrastructure error that prevented validation from completing.
func ValidateProject(root string, manifest *project.Manifest) ([]Issue, error) {
	var issues []Issue

	// --- Check 1 & 2: game.agp references ---
	if manifest.Project.StartRoom != "" {
		if !fileExists(root, manifest.Project.StartRoom) {
			issues = append(issues, Issue{
				File:     "game.agp",
				Severity: "error",
				Message:  fmt.Sprintf("start_room %q not found", manifest.Project.StartRoom),
			})
		}
	}
	if manifest.Project.StartCharacter != "" {
		if !fileExists(root, manifest.Project.StartCharacter) {
			issues = append(issues, Issue{
				File:     "game.agp",
				Severity: "error",
				Message:  fmt.Sprintf("start_character %q not found", manifest.Project.StartCharacter),
			})
		}
	}

	// Scan all source files.
	files, err := project.Scan(root)
	if err != nil {
		return issues, err
	}

	// Build a set of known character names from all .agchar files.
	// Map: character name (from "Character "name"" block) → relative file path.
	charNames := make(map[string]string)
	for _, f := range files {
		if f.Ext != ".agchar" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue // file I/O errors are not validation issues
		}
		cd, parseErr := char.ParseChar(f.Rel, string(data))
		if parseErr != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  parseErr.Error(),
			})
			continue
		}
		charNames[cd.Name] = f.Rel
	}

	// --- Check 3 & 4: .agroom internal consistency and cross-references ---
	for _, f := range files {
		if f.Ext != ".agroom" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		rd, parseErr := room.ParseRoom(f.Rel, string(data))
		if parseErr != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  parseErr.Error(),
			})
			continue
		}

		// Check 3: initial_camera must name a Camera block in this room.
		if rd.InitialCamera != "" {
			if !hasCameraName(rd, rd.InitialCamera) {
				issues = append(issues, Issue{
					File:     f.Rel,
					Severity: "error",
					Message:  fmt.Sprintf("initial_camera %q is not defined in this room", rd.InitialCamera),
				})
			}
		}

		// Check 4: each SpawnPoint.character must name a known .agchar.
		for _, sp := range rd.SpawnPoints {
			if sp.Character == "" {
				continue
			}
			if _, ok := charNames[sp.Character]; !ok {
				issues = append(issues, Issue{
					File:     f.Rel,
					Severity: "error",
					Message:  fmt.Sprintf("SpawnPoint %q: unknown character %q (no matching .agchar found)", sp.Name, sp.Character),
				})
			}
		}
	}

	return issues, nil
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func fileExists(root, rel string) bool {
	_, err := os.Stat(root + string(os.PathSeparator) + rel)
	return err == nil
}

func hasCameraName(rd *room.RoomData, name string) bool {
	for _, cam := range rd.Cameras {
		if cam.Name == name {
			return true
		}
	}
	return false
}
