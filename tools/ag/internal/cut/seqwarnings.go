package cut

import (
	"fmt"
	"strconv"
	"strings"
)

// WarnSequence runs the sequencing warning checks (SEQ-W001..W006) on a parsed
// CutsceneFile. It returns all sequencing warnings found.
//
// Checks implemented:
//
//	SEQ-W001: background step with duration: > 10 s before its covering <<sync>>.
//	SEQ-W002: foreground long-running step (walk_to, run_to, camera move_to) with
//	          no timeout: parameter.
//	SEQ-W003: <<sync>> (all) called when no background steps are currently pending.
//	SEQ-W004: (not yet implemented — requires .aganim sidecar data from T-CUT27)
//	SEQ-W005: on_fail:skip on a step that contains a state change (<<action>>/<<set>>).
//	SEQ-W006: <<wait_for event:name>> where no command in the same sequence emits
//	          that event name.
func WarnSequence(cf *CutsceneFile) []ValidationWarning {
	var warns []ValidationWarning

	// Pre-pass: collect all events emitted within the sequence (for SEQ-W006).
	emittedEvents := collectEmittedEvents(cf)

	// Walk sequence once, tracking pending bg: steps.
	// pending maps bg-id → RawCommand (for W001 check on sync coverage).
	pending := make(map[string]*RawCommand) // bg id → step

	for _, rc := range cf.Sequence {
		cmd := ParseCommand(rc)

		// Track new bg: steps.
		if bgID, ok := cmd.Params["bg"]; ok && bgID != "" {
			pending[bgID] = rc

			// SEQ-W001: bg step with duration: > 10 s.
			if dur, ok := cmd.Params["duration"]; ok {
				if f, err := strconv.ParseFloat(dur, 64); err == nil && f > 10.0 {
					warns = append(warns, ValidationWarning{
						Pos:  rc.Pos,
						Code: "SEQ-W001",
						Msg:  fmt.Sprintf("background step %q has duration %.1f s — consider a shorter timeout or restructuring to avoid blocking the sequence", bgID, f),
					})
				}
			}
		}

		switch rc.Name {
		case "walk_to", "run_to":
			// SEQ-W002: long-running foreground movement with no timeout:.
			if _, hasTimeout := cmd.Params["timeout"]; !hasTimeout {
				warns = append(warns, ValidationWarning{
					Pos:  rc.Pos,
					Code: "SEQ-W002",
					Msg:  fmt.Sprintf("<<%s>> has no timeout: parameter — if the character cannot reach the destination the sequence will hang", rc.Name),
				})
			}

		case "camera":
			// SEQ-W002: camera move_to with no timeout:.
			if len(cmd.Positional) > 0 && cmd.Positional[0] == "move_to" {
				if _, hasTimeout := cmd.Params["timeout"]; !hasTimeout {
					warns = append(warns, ValidationWarning{
						Pos:  rc.Pos,
						Code: "SEQ-W002",
						Msg:  "<<camera move_to>> has no timeout: parameter — if the tween cannot complete the sequence will hang",
					})
				}
			}

		case "sync":
			if len(cmd.Positional) == 0 {
				// SEQ-W003: <<sync>> (all) with no pending backgrounds.
				if len(pending) == 0 {
					warns = append(warns, ValidationWarning{
						Pos:  rc.Pos,
						Code: "SEQ-W003",
						Msg:  "<<sync>> (all) called but no background steps are currently pending — this sync point is a no-op",
					})
				}
				// All pending bg steps are now considered synced.
				pending = make(map[string]*RawCommand)
			} else {
				// Named sync — remove covered ids from pending.
				for _, id := range cmd.Positional {
					delete(pending, id)
				}
			}

		case "action", "set":
			// SEQ-W005: on_fail:skip on a state-change step.
			// action/set args are raw expressions so Params is empty;
			// scan the raw args string for the on_fail:skip token.
			if containsOnFailSkip(rc.Args) {
				warns = append(warns, ValidationWarning{
					Pos:  rc.Pos,
					Code: "SEQ-W005",
					Msg:  fmt.Sprintf("<<%s>> uses on_fail:skip but contains a state change — skipping it may leave the game in an inconsistent state", rc.Name),
				})
			}

		case "wait_for":
			// SEQ-W006: wait_for event: not emitted anywhere in the same sequence.
			if eventName, ok := cmd.Params["event"]; ok && eventName != "" {
				if !emittedEvents[eventName] {
					warns = append(warns, ValidationWarning{
						Pos:  rc.Pos,
						Code: "SEQ-W006",
						Msg:  fmt.Sprintf("<<wait_for event:%s>> — event %q is not emitted by any command in this cutscene file", eventName, eventName),
					})
				}
			}
		}
	}

	return warns
}

// containsOnFailSkip reports whether the raw args string contains on_fail:skip
// as a standalone token (not as part of a longer value).
func containsOnFailSkip(args string) bool {
	for _, tok := range strings.Fields(args) {
		if tok == "on_fail:skip" {
			return true
		}
	}
	return false
}

// collectEmittedEvents returns a set of event names emitted within the sequence.
// Events are emitted by <<emit event:name>> commands or <<on event:name>> handlers.
func collectEmittedEvents(cf *CutsceneFile) map[string]bool {
	emitted := make(map[string]bool)
	for _, rc := range cf.Sequence {
		cmd := ParseCommand(rc)
		switch rc.Name {
		case "emit":
			if name, ok := cmd.Params["event"]; ok && name != "" {
				emitted[name] = true
			}
			// also accept positional: <<emit event_name>>
			if len(cmd.Positional) > 0 {
				emitted[cmd.Positional[0]] = true
			}
		case "on":
			// <<on event:name>> — the event name in the param
			for k, v := range cmd.Params {
				if k == "event" || strings.HasPrefix(k, "event") {
					emitted[v] = true
				}
			}
		}
	}
	return emitted
}
