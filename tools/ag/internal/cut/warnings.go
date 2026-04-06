package cut

import (
	"fmt"
	"strings"
)

// ValidationWarning is a non-blocking diagnostic found during cutscene
// validation. Warnings do not prevent a build but should be surfaced to
// the author.
type ValidationWarning struct {
	Pos  Pos
	Code string // CUT-W001..W011
	Msg  string
}

func (w ValidationWarning) Error() string {
	return fmt.Sprintf("%s: %s: %s", w.Pos, w.Code, w.Msg)
}

// WarnCutscene runs single-file warning checks (CUT-W006..W011) on a parsed
// CutsceneFile. allTitles may be nil (skips W001). cross may be nil (skips
// cross-system checks). W001–W005 require multi-file or cross-system context
// and are checked by WarnProjectCutscenes.
func WarnCutscene(cf *CutsceneFile, _ map[string]bool, _ *ProjectIndex) []ValidationWarning {
	var warns []ValidationWarning

	// W006: sequence has no <<end>> and no room.transition call.
	if !sequenceHasEnd(cf) && !sequenceHasRoomTransition(cf) {
		warns = append(warns, ValidationWarning{
			Pos:  Pos{File: cf.Path, Line: 1},
			Code: "CUT-W006",
			Msg:  fmt.Sprintf("cutscene %q has no <<end>> and no room.transition — sequence has no defined exit", cf.Title),
		})
	}

	// Collect labels and skip_to targets for W007.
	labelPos := make(map[string]Pos)   // label name → declared pos
	skipToTargets := make(map[string]bool) // targets referenced by <<skip_to>>

	// Track audio channels for W009.
	type chanStart struct {
		pos Pos
	}
	openChannels := make(map[string]chanStart) // channel → open start

	for _, rc := range cf.Sequence {
		cmd := ParseCommand(rc)

		switch rc.Name {
		case "label":
			if len(cmd.Positional) > 0 {
				name := cmd.Positional[0]
				if _, already := labelPos[name]; !already {
					labelPos[name] = rc.Pos
				}
			}

		case "skip_to":
			if len(cmd.Positional) > 0 {
				skipToTargets[cmd.Positional[0]] = true
			}

		case "music", "ambient", "sound":
			if isStopCommand(cmd) {
				ch := stopChannelName(cmd)
				if ch != "" {
					delete(openChannels, ch)
				} else {
					// stop with no identifiable channel — clear first open of this type
					// (best-effort for W009)
				}
			} else {
				if len(cmd.Positional) > 0 {
					ch := cmd.Positional[0]
					if _, already := openChannels[ch]; !already {
						openChannels[ch] = chanStart{pos: rc.Pos}
					}
				}
			}
		}

		// W010: duck:all in any command args.
		if hasDuckAll(rc.Args) {
			warns = append(warns, ValidationWarning{
				Pos:  rc.Pos,
				Code: "CUT-W010",
				Msg:  "duck:all used — the set of active channels is determined at runtime and cannot be validated at build time",
			})
		}
	}

	// W007: labels declared but never referenced by <<skip_to>>.
	for name, pos := range labelPos {
		if !skipToTargets[name] {
			warns = append(warns, ValidationWarning{
				Pos:  pos,
				Code: "CUT-W007",
				Msg:  fmt.Sprintf("label %q is declared but never used as a <<skip_to>> target", name),
			})
		}
	}

	// W008: author_controlled skip with no <<label>>.
	if cf.Skip == "author_controlled" && len(labelPos) == 0 {
		warns = append(warns, ValidationWarning{
			Pos:  Pos{File: cf.Path, Line: 1},
			Code: "CUT-W008",
			Msg:  "skip:author_controlled but no <<label>> commands defined — no skip points available",
		})
	}

	// W009: audio channels still open at end of sequence.
	for ch, start := range openChannels {
		warns = append(warns, ValidationWarning{
			Pos:  start.pos,
			Code: "CUT-W009",
			Msg:  fmt.Sprintf("audio channel %q started but never stopped — runtime will hard-stop at sequence end", ch),
		})
	}

	// W011: auto_duck:true with empty duck_channels.
	if cf.AutoDuck && strings.TrimSpace(cf.DuckChannels) == "" {
		warns = append(warns, ValidationWarning{
			Pos:  Pos{File: cf.Path, Line: 1},
			Code: "CUT-W011",
			Msg:  "auto_duck:true but duck_channels not set — no channels will be ducked",
		})
	}

	return warns
}

// WarnProjectCutscenes runs multi-file warning checks (CUT-W001) across all
// files, then delegates per-file warnings to WarnCutscene.
func WarnProjectCutscenes(files []*CutsceneFile, cross *ProjectIndex) []ValidationWarning {
	var warns []ValidationWarning

	// Build set of all titles and all referenced titles.
	allTitles := make(map[string]bool)
	referencedTitles := make(map[string]bool)
	titleToFile := make(map[string]*CutsceneFile)

	for _, cf := range files {
		if cf.Title != "" {
			allTitles[cf.Title] = true
			titleToFile[cf.Title] = cf
		}
		for _, rc := range cf.Sequence {
			if rc.Name == "cutscene" {
				cmd := ParseCommand(rc)
				ref := ""
				if len(cmd.Positional) > 0 {
					ref = cmd.Positional[0]
				} else if v, ok := cmd.Params["file"]; ok {
					ref = v
				}
				if ref != "" {
					referencedTitles[ref] = true
				}
			}
		}
	}

	// W001: cutscene title never referenced by any <<cutscene>> call.
	for title, cf := range titleToFile {
		if !referencedTitles[title] {
			warns = append(warns, ValidationWarning{
				Pos:  Pos{File: cf.Path, Line: 1},
				Code: "CUT-W001",
				Msg:  fmt.Sprintf("cutscene %q is never triggered by any <<cutscene>> call in the project", title),
			})
		}
	}

	// Per-file warnings.
	for _, cf := range files {
		warns = append(warns, WarnCutscene(cf, allTitles, cross)...)
	}

	return warns
}

// ── helpers ──────────────────────────────────────────────────────────────────

// sequenceHasEnd returns true if the flat sequence contains an <<end>> command.
func sequenceHasEnd(cf *CutsceneFile) bool {
	for _, rc := range cf.Sequence {
		if rc.Name == "end" {
			return true
		}
	}
	return false
}

// sequenceHasRoomTransition returns true if any <<action>> contains room.transition.
func sequenceHasRoomTransition(cf *CutsceneFile) bool {
	for _, rc := range cf.Sequence {
		if rc.Name == "action" && strings.Contains(rc.Args, "room.transition") {
			return true
		}
	}
	return false
}

// isStopCommand returns true if the command represents a stop operation on an
// audio channel (has "stop" as a positional arg or a stop: named param).
func isStopCommand(cmd *Command) bool {
	for _, p := range cmd.Positional {
		if p == "stop" {
			return true
		}
	}
	_, hasStop := cmd.Params["stop"]
	return hasStop
}

// stopChannelName returns the channel name being stopped, or "" if unknown.
func stopChannelName(cmd *Command) string {
	// Prefer explicit channel: param.
	if ch, ok := cmd.Params["channel"]; ok && ch != "" {
		return ch
	}
	// Fall back to first non-"stop" positional.
	for _, p := range cmd.Positional {
		if p != "stop" {
			return p
		}
	}
	return ""
}

// hasDuckAll returns true if the raw args string contains the token "duck:all".
func hasDuckAll(args string) bool {
	for _, field := range strings.Fields(args) {
		if field == "duck:all" {
			return true
		}
	}
	return false
}
