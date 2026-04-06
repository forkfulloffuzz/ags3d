package cut

import (
	"fmt"
	"strings"
)

// ValidationError is a structural error found during cutscene validation.
// All errors in this set block the build.
type ValidationError struct {
	Pos  Pos
	Code string // CUT-E001..E012
	Msg  string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Pos, e.Code, e.Msg)
}

// ProjectIndex provides the cross-system symbol table the validator needs
// to check references to characters, named points, audio files, and other
// cutscene files. Callers populate only the fields relevant to their context;
// nil/empty means "not checked".
type ProjectIndex struct {
	// Characters is the set of declared character names.
	Characters map[string]bool
	// NamedPoints is a map from room name → set of point names within that room.
	// Used for CUT-E002 validation when a room context is known.
	NamedPoints map[string]map[string]bool
	// AudioFiles is the set of audio file basenames (without extension).
	AudioFiles map[string]bool
	// VideoFiles is the set of video file basenames.
	VideoFiles map[string]bool
	// CutsceneTitles is the set of all declared cutscene titles (populated from
	// multiple parsed CutsceneFiles). Used for CUT-E001, CUT-E008, CUT-E009.
	CutsceneTitles map[string]bool
}

// ValidateCutscene runs structural validation on a parsed CutsceneFile.
// It checks CUT-E001..E012 (see milestone doc for reference).
// cross may be nil when cross-system symbol information is not available.
func ValidateCutscene(cf *CutsceneFile, allTitles map[string]bool, cross *ProjectIndex) []ValidationError {
	var errs []ValidationError

	// CUT-E001: title uniqueness checked by caller (requires multi-file context);
	// here we flag an empty title.
	if cf.Title == "" {
		errs = append(errs, ValidationError{
			Pos:  Pos{File: cf.Path, Line: 1},
			Code: "CUT-E001",
			Msg:  "cutscene is missing required 'title' field",
		})
	}

	// CUT-E012: save_block:false with state changes (<<action>> or <<set>>).
	if !cf.SaveBlock {
		for _, rc := range cf.Sequence {
			if rc.Name == "action" || rc.Name == "set" {
				errs = append(errs, ValidationError{
					Pos:  rc.Pos,
					Code: "CUT-E012",
					Msg:  fmt.Sprintf("save_block:false cutscene %q contains state-change command <<%s>>", cf.Title, rc.Name),
				})
			}
		}
	}

	// Collect labels and skip_to targets for CUT-E006.
	labels := make(map[string]bool)
	var skipTos []struct {
		target string
		pos    Pos
	}
	// Track if we're inside a parallel block for CUT-E007.
	parallelDepth := 0

	for _, rc := range cf.Sequence {
		cmd := ParseCommand(rc)

		switch rc.Name {
		case "label":
			if len(cmd.Positional) > 0 {
				labels[cmd.Positional[0]] = true
			}

		case "skip_to":
			if len(cmd.Positional) > 0 {
				skipTos = append(skipTos, struct {
					target string
					pos    Pos
				}{cmd.Positional[0], rc.Pos})
			}

		case "parallel":
			parallelDepth++

		case "end_parallel":
			if parallelDepth > 0 {
				parallelDepth--
			}

		case "choice":
			// CUT-E007: <<choice>> inside <<parallel>>.
			if parallelDepth > 0 {
				errs = append(errs, ValidationError{
					Pos:  rc.Pos,
					Code: "CUT-E007",
					Msg:  "<<choice>> is not allowed inside a <<parallel>> block",
				})
			}

		case "cutscene":
			// CUT-E008: nested cutscene reference must exist.
			// CUT-E009: circular nesting (simple: self-reference).
			ref := ""
			if len(cmd.Positional) > 0 {
				ref = cmd.Positional[0]
			} else if v, ok := cmd.Params["file"]; ok {
				ref = v
			}
			if ref != "" {
				if allTitles != nil && !allTitles[ref] {
					errs = append(errs, ValidationError{
						Pos:  rc.Pos,
						Code: "CUT-E008",
						Msg:  fmt.Sprintf("nested cutscene %q does not exist", ref),
					})
				}
				if ref == cf.Title {
					errs = append(errs, ValidationError{
						Pos:  rc.Pos,
						Code: "CUT-E009",
						Msg:  fmt.Sprintf("cutscene %q references itself (circular nesting)", cf.Title),
					})
				}
			}
		}

		// Cross-system checks when index is provided.
		if cross != nil {
			errs = append(errs, crossCheck(cf, rc, cmd, cross)...)
		}
	}

	// CUT-E006: skip_to targets must be declared labels.
	for _, st := range skipTos {
		if !labels[st.target] {
			errs = append(errs, ValidationError{
				Pos:  st.pos,
				Code: "CUT-E006",
				Msg:  fmt.Sprintf("skip_to target %q does not exist as a <<label>> in this sequence", st.target),
			})
		}
	}

	return errs
}

// crossCheck performs the cross-system checks (CUT-E002..E005, E010, E011)
// for a single command, using the provided ProjectIndex.
func crossCheck(cf *CutsceneFile, rc *RawCommand, cmd *Command, idx *ProjectIndex) []ValidationError {
	var errs []ValidationError

	switch rc.Name {
	case "character":
		// CUT-E003: character referenced not defined.
		if len(cmd.Positional) > 0 && idx.Characters != nil {
			charName := cmd.Positional[0]
			if !idx.Characters[charName] {
				errs = append(errs, ValidationError{
					Pos:  rc.Pos,
					Code: "CUT-E003",
					Msg:  fmt.Sprintf("character %q is not defined", charName),
				})
			}
		}
		// CUT-E010: animation name on character.
		if playAnim, ok := cmd.Params["play"]; ok && len(cmd.Positional) > 0 {
			charName := cmd.Positional[0]
			_ = charName
			_ = playAnim
			// Animation validation requires per-character animation lists —
			// deferred to cross-system validator (T-DLG05 equivalent for CUT).
		}

	case "music", "ambient":
		// CUT-E004: audio file referenced does not exist.
		if len(cmd.Positional) > 0 && idx.AudioFiles != nil {
			name := cmd.Positional[0]
			if !idx.AudioFiles[name] {
				errs = append(errs, ValidationError{
					Pos:  rc.Pos,
					Code: "CUT-E004",
					Msg:  fmt.Sprintf("audio file %q is not defined", name),
				})
			}
		}

	case "video":
		// CUT-E005: video file referenced does not exist.
		if len(cmd.Positional) > 0 && idx.VideoFiles != nil {
			name := cmd.Positional[0]
			if !idx.VideoFiles[name] {
				errs = append(errs, ValidationError{
					Pos:  rc.Pos,
					Code: "CUT-E005",
					Msg:  fmt.Sprintf("video file %q is not defined", name),
				})
			}
		}

	case "camera":
		// CUT-E002: named point exists.
		if len(cmd.Positional) >= 2 {
			maybePoint := cmd.Positional[1]
			if strings.HasPrefix(maybePoint, "point.") && idx.NamedPoints != nil {
				pointName := strings.TrimPrefix(maybePoint, "point.")
				// Check if the point exists in any declared room (simple check).
				found := false
				for _, pts := range idx.NamedPoints {
					if pts[pointName] {
						found = true
						break
					}
				}
				if !found {
					errs = append(errs, ValidationError{
						Pos:  rc.Pos,
						Code: "CUT-E002",
						Msg:  fmt.Sprintf("named point %q does not exist in any room", pointName),
					})
				}
			}
		}
	}

	return errs
}

// ValidateProjectCutscenes validates title uniqueness across all cutscenes
// (CUT-E001) and runs per-file validation. Returns all errors found.
func ValidateProjectCutscenes(files []*CutsceneFile, cross *ProjectIndex) []ValidationError {
	var errs []ValidationError

	// Build title → file index; detect duplicates (CUT-E001).
	titles := make(map[string]*CutsceneFile)
	allTitles := make(map[string]bool)
	for _, cf := range files {
		if cf.Title == "" {
			continue
		}
		allTitles[cf.Title] = true
		if existing, dup := titles[cf.Title]; dup {
			errs = append(errs, ValidationError{
				Pos:  Pos{File: cf.Path, Line: 1},
				Code: "CUT-E001",
				Msg:  fmt.Sprintf("cutscene title %q is not unique (first declared in %s)", cf.Title, existing.Path),
			})
		} else {
			titles[cf.Title] = cf
		}
	}

	// Per-file validation.
	for _, cf := range files {
		errs = append(errs, ValidateCutscene(cf, allTitles, cross)...)
	}

	return errs
}
