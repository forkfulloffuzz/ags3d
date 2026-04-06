package dlg

import (
	"fmt"
	"strings"
)

// ProjectSymbolTable holds the project-wide symbol sets needed by the
// cross-system dialogue validator (T-DLG05, DLG-E020..E025).
//
// Populate by scanning all .agchar, .agroom, and .agitem files before
// calling ValidateCrossSystem.
type ProjectSymbolTable struct {
	// CharacterNames is the set of all defined character names (from .agchar).
	CharacterNames map[string]bool

	// RoomNames is the set of all defined room names (from .agroom).
	RoomNames map[string]bool

	// RoomPoints maps room_name → set of point names within that room.
	// Used for DLG-E024 (named point not in room).
	RoomPoints map[string]map[string]bool

	// ItemNames is the set of all defined inventory item names (from .agitem).
	ItemNames map[string]bool

	// FlagsEverSet is the set of all flag names that appear on the left-hand
	// side of an <<action flag.name = …>> or <<set>> command anywhere in the
	// project. Used for DLG-E023 (flag referenced but never set).
	FlagsEverSet map[string]bool

	// KnowledgeFlags is the set of flag names that are ever granted
	// ("flag.name = true" or similar) across the whole project.
	// Used for DLG-E025 (knowledge flag never granted).
	KnowledgeFlags map[string]bool
}

// CrossValidationError is an error found during cross-system validation.
// All errors in this set block the build.
type CrossValidationError struct {
	Pos  Pos
	Code string // DLG-E020..E025
	Msg  string
}

func (e CrossValidationError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Pos, e.Code, e.Msg)
}

// ValidateCrossSystem runs the cross-system validator (DLG-E020..E025) against
// a LinkedProject using the provided ProjectSymbolTable. It returns all errors
// found; an empty slice means cross-system validation passed.
//
// Errors checked:
//   - DLG-E020: inventory item referenced in condition does not exist
//   - DLG-E021: room referenced in action does not exist
//   - DLG-E022: character property referenced in condition/action does not exist
//   - DLG-E023: flag referenced in condition never set anywhere in project
//   - DLG-E024: named point referenced in dialogue action not in room
//   - DLG-E025: knowledge flag referenced but never granted anywhere
func ValidateCrossSystem(lp *LinkedProject, sym ProjectSymbolTable) []CrossValidationError {
	var errs []CrossValidationError
	for _, ln := range lp.NodesByTitle {
		errs = append(errs, crossCheckNode(ln.Node, sym)...)
	}
	return errs
}

// crossCheckNode runs cross-system checks on all statements in a node.
func crossCheckNode(n *DialogueNode, sym ProjectSymbolTable) []CrossValidationError {
	var errs []CrossValidationError
	errs = append(errs, crossCheckStmts(n.Body, sym)...)
	return errs
}

func crossCheckStmts(stmts []Statement, sym ProjectSymbolTable) []CrossValidationError {
	var errs []CrossValidationError
	for _, s := range stmts {
		switch st := s.(type) {
		case *CommandStatement:
			errs = append(errs, crossCheckCmd(st.Raw, st.SrcPos, sym)...)
		case *SpeakerLine:
			for _, c := range st.Commands {
				errs = append(errs, crossCheckCmd(c.Raw, c.SrcPos, sym)...)
			}
		case *NarrationLine:
			for _, c := range st.Commands {
				errs = append(errs, crossCheckCmd(c.Raw, c.SrcPos, sym)...)
			}
		case *OptionBranch:
			for _, c := range st.Commands {
				errs = append(errs, crossCheckCmd(c.Raw, c.SrcPos, sym)...)
			}
			errs = append(errs, crossCheckStmts(st.Body, sym)...)
		}
	}
	return errs
}

// crossCheckCmd analyses a single command expression (the content between << >>)
// for cross-system references.
func crossCheckCmd(raw string, pos Pos, sym ProjectSymbolTable) []CrossValidationError {
	var errs []CrossValidationError
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	// DLG-E021: room.transition("room_name") references a room that doesn't exist.
	if name, ok := extractRoomTransition(raw); ok {
		if len(sym.RoomNames) > 0 && !sym.RoomNames[name] {
			errs = append(errs, CrossValidationError{
				Pos:  pos,
				Code: "DLG-E021",
				Msg:  fmt.Sprintf("room %q referenced in action does not exist", name),
			})
		}
	}

	// DLG-E024: room.point("room", "point") or point.room_point style references.
	if room, point, ok := extractRoomPoint(raw); ok {
		if len(sym.RoomPoints) > 0 {
			pts, roomKnown := sym.RoomPoints[room]
			if roomKnown && !pts[point] {
				errs = append(errs, CrossValidationError{
					Pos:  pos,
					Code: "DLG-E024",
					Msg:  fmt.Sprintf("named point %q not defined in room %q", point, room),
				})
			}
		}
	}

	// DLG-E020: item.name in condition / item references.
	if item, ok := extractItemRef(raw); ok {
		if len(sym.ItemNames) > 0 && !sym.ItemNames[item] {
			errs = append(errs, CrossValidationError{
				Pos:  pos,
				Code: "DLG-E020",
				Msg:  fmt.Sprintf("inventory item %q referenced does not exist", item),
			})
		}
	}

	// DLG-E022: char.property references where char is not defined.
	if char, ok := extractCharRef(raw); ok {
		if len(sym.CharacterNames) > 0 && !sym.CharacterNames[char] {
			errs = append(errs, CrossValidationError{
				Pos:  pos,
				Code: "DLG-E022",
				Msg:  fmt.Sprintf("character %q referenced does not exist", char),
			})
		}
	}

	// DLG-E023: flag.name in visible_if / available_if — flag never set.
	if flag, ok := extractFlagConditionRef(raw); ok {
		if len(sym.FlagsEverSet) > 0 && !sym.FlagsEverSet[flag] {
			errs = append(errs, CrossValidationError{
				Pos:  pos,
				Code: "DLG-E023",
				Msg:  fmt.Sprintf("flag %q referenced in condition is never set anywhere in the project", flag),
			})
		}
	}

	// DLG-E025: knowledge flag referenced (in visible_if / available_if knowledge.X)
	// but never granted.
	if flag, ok := extractKnowledgeRef(raw); ok {
		if len(sym.KnowledgeFlags) > 0 && !sym.KnowledgeFlags[flag] {
			errs = append(errs, CrossValidationError{
				Pos:  pos,
				Code: "DLG-E025",
				Msg:  fmt.Sprintf("knowledge flag %q referenced but never granted anywhere", flag),
			})
		}
	}

	return errs
}

// ---------------------------------------------------------------------------
// Reference extractors
// ---------------------------------------------------------------------------

// extractRoomTransition returns ("room_name", true) for commands of the form:
//
//	action room.transition("room_name")
//	action room.transition('room_name')
func extractRoomTransition(raw string) (string, bool) {
	const needle = "room.transition("
	idx := strings.Index(raw, needle)
	if idx < 0 {
		return "", false
	}
	rest := raw[idx+len(needle):]
	name, ok := extractQuotedArg(rest)
	return name, ok
}

// extractRoomPoint returns ("room", "point", true) for patterns like:
//
//	action char.walk_to(room.point("market", "stall_left"))
//	point.market.stall_left
func extractRoomPoint(raw string) (string, string, bool) {
	const needle = "room.point("
	idx := strings.Index(raw, needle)
	if idx < 0 {
		return "", "", false
	}
	rest := raw[idx+len(needle):]
	// Expect two quoted args: room_name, point_name
	room, rest2, ok := extractQuotedArgWithRest(rest)
	if !ok {
		return "", "", false
	}
	// Skip comma and whitespace.
	rest2 = strings.TrimLeft(rest2, ", \t")
	point, _, ok := extractQuotedArgWithRest(rest2)
	if !ok {
		return "", "", false
	}
	return room, point, true
}

// extractItemRef returns ("item_name", true) for patterns like:
//
//	visible_if item.gate_pass in player.inventory
//	available_if not item.key in player.inventory
func extractItemRef(raw string) (string, bool) {
	// Look for "item." prefix in a visible_if / available_if condition.
	if !strings.HasPrefix(raw, "visible_if") && !strings.HasPrefix(raw, "available_if") {
		return "", false
	}
	const prefix = "item."
	idx := strings.Index(raw, prefix)
	if idx < 0 {
		return "", false
	}
	rest := raw[idx+len(prefix):]
	name := readIdentifier(rest)
	if name == "" {
		return "", false
	}
	return name, true
}

// extractCharRef returns ("char_name", true) for patterns like:
//
//	visible_if char.guard.suspicious
//	action char.elara.mood = "happy"
func extractCharRef(raw string) (string, bool) {
	const prefix = "char."
	idx := strings.Index(raw, prefix)
	if idx < 0 {
		return "", false
	}
	rest := raw[idx+len(prefix):]
	name := readIdentifier(rest)
	if name == "" {
		return "", false
	}
	return name, true
}

// extractFlagConditionRef returns ("flag_name", true) for patterns like:
//
//	visible_if flag.guard_spoken
//	available_if not flag.gate_open
func extractFlagConditionRef(raw string) (string, bool) {
	if !strings.HasPrefix(raw, "visible_if") && !strings.HasPrefix(raw, "available_if") {
		return "", false
	}
	const prefix = "flag."
	idx := strings.Index(raw, prefix)
	if idx < 0 {
		return "", false
	}
	rest := raw[idx+len(prefix):]
	name := readIdentifier(rest)
	if name == "" {
		return "", false
	}
	return name, true
}

// extractKnowledgeRef returns ("knowledge_name", true) for patterns like:
//
//	visible_if knowledge.guard_secret
func extractKnowledgeRef(raw string) (string, bool) {
	if !strings.HasPrefix(raw, "visible_if") && !strings.HasPrefix(raw, "available_if") {
		return "", false
	}
	const prefix = "knowledge."
	idx := strings.Index(raw, prefix)
	if idx < 0 {
		return "", false
	}
	rest := raw[idx+len(prefix):]
	name := readIdentifier(rest)
	if name == "" {
		return "", false
	}
	return name, true
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

// extractQuotedArg returns the string content of the first "..." or '...'
// starting at the beginning of s.
func extractQuotedArg(s string) (string, bool) {
	v, _, ok := extractQuotedArgWithRest(s)
	return v, ok
}

// extractQuotedArgWithRest returns (value, remainder, ok). remainder starts
// after the closing quote.
func extractQuotedArgWithRest(s string) (string, string, bool) {
	s = strings.TrimLeft(s, " \t")
	if len(s) == 0 {
		return "", "", false
	}
	quote := rune(s[0])
	if quote != '"' && quote != '\'' {
		return "", "", false
	}
	end := strings.IndexRune(s[1:], quote)
	if end < 0 {
		return "", "", false
	}
	return s[1 : end+1], s[end+2:], true
}

// readIdentifier reads a [a-z0-9_] identifier from the start of s.
func readIdentifier(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			break
		}
	}
	return b.String()
}
