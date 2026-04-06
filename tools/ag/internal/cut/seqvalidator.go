package cut

import (
	"fmt"
	"strings"
)

// ValidateSequence runs the sequencing validator (SEQ-E001..E007) on the flat
// command list of a parsed CutsceneFile. It returns all sequencing errors found.
//
// Checks implemented:
//
//	SEQ-E001: <<sync id>> references an id never declared as bg: on any command.
//	SEQ-E002: <<sync id>> references an id declared as a foreground id: (not bg:).
//	SEQ-E003: a bg:id step has no covering <<sync id>> or <<sync>> (all) later in
//	          the flat command list.
//	SEQ-E004: on_fail:jump_to:label references a label not declared in the sequence.
//	SEQ-E007: duplicate step id (bg: or id:) within the same sequence.
func ValidateSequence(cf *CutsceneFile) []ValidationError {
	var errs []ValidationError

	// First pass: collect all declared ids (bg: = background, id: = foreground)
	// and label names, and detect duplicates (SEQ-E007).
	bgIDs := make(map[string]Pos)   // id → first declaration pos
	fgIDs := make(map[string]Pos)   // foreground id: step identifiers
	allIDs := make(map[string]bool) // union of bg + fg ids (for SEQ-E001/E002 lookup)
	labels := make(map[string]bool)

	for _, rc := range cf.Sequence {
		cmd := ParseCommand(rc)

		if bgID, ok := cmd.Params["bg"]; ok && bgID != "" {
			if _, dup := bgIDs[bgID]; dup {
				errs = append(errs, ValidationError{
					Pos:  rc.Pos,
					Code: "SEQ-E007",
					Msg:  fmt.Sprintf("duplicate step id %q in same sequence", bgID),
				})
			} else {
				bgIDs[bgID] = rc.Pos
				allIDs[bgID] = true
			}
		}
		if fgID, ok := cmd.Params["id"]; ok && fgID != "" {
			if _, dup := fgIDs[fgID]; dup {
				errs = append(errs, ValidationError{
					Pos:  rc.Pos,
					Code: "SEQ-E007",
					Msg:  fmt.Sprintf("duplicate step id %q in same sequence", fgID),
				})
			} else {
				fgIDs[fgID] = rc.Pos
				allIDs[fgID] = true
			}
		}

		if rc.Name == "label" && len(cmd.Positional) > 0 {
			labels[cmd.Positional[0]] = true
		}
	}

	// Second pass: check <<sync>>, <<on_fail>>, and accumulate bg coverage.
	// synced tracks which bg ids have been covered by a <<sync>>.
	synced := make(map[string]bool)
	syncAll := false // true if we see <<sync>> with no ids (covers all bg steps)

	for _, rc := range cf.Sequence {
		cmd := ParseCommand(rc)

		switch rc.Name {
		case "sync":
			if len(cmd.Positional) == 0 {
				// <<sync>> with no args — covers all background steps.
				syncAll = true
				for id := range bgIDs {
					synced[id] = true
				}
			} else {
				for _, id := range cmd.Positional {
					if !allIDs[id] {
						// SEQ-E001: id never declared anywhere.
						errs = append(errs, ValidationError{
							Pos:  rc.Pos,
							Code: "SEQ-E001",
							Msg:  fmt.Sprintf("<<sync>> references %q which was never declared as a bg: step id", id),
						})
					} else if _, isFG := fgIDs[id]; isFG {
						// SEQ-E002: id is a foreground id:, not a background bg:.
						errs = append(errs, ValidationError{
							Pos:  rc.Pos,
							Code: "SEQ-E002",
							Msg:  fmt.Sprintf("<<sync>> references %q which is a foreground id: step, not a background bg: step", id),
						})
					} else {
						synced[id] = true
					}
				}
			}
		}

		// SEQ-E004: on_fail:jump_to:label — label must exist.
		for k, v := range cmd.Params {
			if k == "on_fail" && strings.HasPrefix(v, "jump_to:") {
				target := strings.TrimPrefix(v, "jump_to:")
				if !labels[target] {
					errs = append(errs, ValidationError{
						Pos:  rc.Pos,
						Code: "SEQ-E004",
						Msg:  fmt.Sprintf("on_fail:jump_to:%s references label %q not declared in this sequence", target, target),
					})
				}
			}
		}
	}

	// SEQ-E003: any bg: step with no covering <<sync>>.
	if !syncAll {
		for id, pos := range bgIDs {
			if !synced[id] {
				errs = append(errs, ValidationError{
					Pos:  pos,
					Code: "SEQ-E003",
					Msg:  fmt.Sprintf("background step %q has no covering <<sync>> — it will be abandoned at sequence end", id),
				})
			}
		}
	}

	return errs
}
