package dlg

import (
	"fmt"
	"regexp"
	"strings"
)

// dlgIdentRe is the naming rule for all author-assigned dialogue identifiers.
var dlgIdentRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// isValidDlgIdent returns true if s satisfies the naming rule.
func isValidDlgIdent(s string) bool { return dlgIdentRe.MatchString(s) }

// dlgIdentErr returns a DLG-E011 ValidationError for field at pos.
func dlgIdentErr(pos Pos, field, value string) ValidationError {
	return ValidationError{
		Pos:  pos,
		Code: "DLG-E011",
		Msg:  fmt.Sprintf("identifier %q in %s violates naming rule (must match ^[a-z][a-z0-9_]*$)", value, field),
	}
}

// ValidationError is a structural error found during the validate stage.
// All errors in this set block the build.
type ValidationError struct {
	Pos  Pos
	Code string // DLG-E001..E010
	Msg  string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Pos, e.Code, e.Msg)
}

// Validate runs the structural dialogue validator (DLG-E001..E010) on a
// LinkedProject. It returns all errors found; an empty slice means the
// dialogue structure is valid.
//
// Errors checked here:
//   - DLG-E001: duplicate node title (already caught in Link; re-checked here
//     in case Validate is called standalone)
//   - DLG-E002: jump target does not exist (promoted from LinkErrors)
//   - DLG-E003: node has no reachable <<end>> or <<jump>>
//   - DLG-E007: malformed <<command>> expression (empty command content)
//   - DLG-E008: option with no body content and no <<end>> or <<jump>>
//   - DLG-E009: circular jump chain with no exit
//   - DLG-E010: loc_id collision between nodes
func Validate(lp *LinkedProject) []ValidationError {
	var errs []ValidationError

	// Promote link errors that are structural errors.
	for _, le := range lp.LinkErrors {
		switch le.Code {
		case "DLG-E001", "DLG-E002", "DLG-E006":
			errs = append(errs, ValidationError{Pos: le.Pos, Code: le.Code, Msg: le.Msg})
		}
	}

	// DLG-E010: loc_id collision.
	locIDs := make(map[string]*LinkedNode)
	for _, ln := range lp.NodesByTitle {
		if ln.Node.LocID != "" {
			if existing, dup := locIDs[ln.Node.LocID]; dup {
				errs = append(errs, ValidationError{
					Pos:  ln.Node.Pos,
					Code: "DLG-E010",
					Msg:  fmt.Sprintf("loc_id %q collides with node %q (at %s)", ln.Node.LocID, existing.Node.Title, existing.Node.Pos),
				})
			} else {
				locIDs[ln.Node.LocID] = ln
			}
		}
	}

	// Per-node checks.
	for _, ln := range lp.NodesByTitle {
		n := ln.Node
		errs = append(errs, validateNode(lp, n)...)
	}

	// DLG-E009: circular jump detection (nodes that only jump to each other with no <<end>>).
	errs = append(errs, detectCircularJumps(lp)...)

	return errs
}

// validateNode runs per-node structural checks.
func validateNode(lp *LinkedProject, n *DialogueNode) []ValidationError {
	var errs []ValidationError

	// DLG-E011: identifier naming rule.
	if n.Title != "" && !isValidDlgIdent(n.Title) {
		errs = append(errs, dlgIdentErr(n.Pos, "title", n.Title))
	}
	if n.Character != "" && !isValidDlgIdent(n.Character) {
		errs = append(errs, dlgIdentErr(n.Pos, "character", n.Character))
	}
	if n.LocID != "" && !isValidDlgIdent(n.LocID) {
		errs = append(errs, dlgIdentErr(n.Pos, "loc_id", n.LocID))
	}
	errs = append(errs, checkJumpTargetIdents(n.Body)...)

	// DLG-E007: empty <<>> command content.
	errs = append(errs, checkEmptyCommands(n.Body)...)

	// DLG-E008: options with no body and no end/jump.
	errs = append(errs, checkOptionBodies(n.Body)...)

	// DLG-E003: node has no reachable terminal (<<end>> or <<jump>>).
	if !bodyHasTerminal(n.Body) {
		errs = append(errs, ValidationError{
			Pos:  n.Pos,
			Code: "DLG-E003",
			Msg:  fmt.Sprintf("node %q has no reachable <<end>> or <<jump>>", n.Title),
		})
	}

	return errs
}

// checkJumpTargetIdents walks body statements and flags any <<jump target>>
// where target violates the identifier naming rule (DLG-E011).
func checkJumpTargetIdents(stmts []Statement) []ValidationError {
	var errs []ValidationError
	for _, s := range stmts {
		switch st := s.(type) {
		case *CommandStatement:
			if target, ok := jumpTarget(st.Raw); ok && !isValidDlgIdent(target) {
				errs = append(errs, dlgIdentErr(st.SrcPos, "jump target", target))
			}
		case *OptionBranch:
			errs = append(errs, checkJumpTargetIdents(st.Body)...)
		}
	}
	return errs
}

// checkEmptyCommands walks statements and flags any CommandStatement or
// CommandExpr with empty Raw content (DLG-E007).
func checkEmptyCommands(stmts []Statement) []ValidationError {
	var errs []ValidationError
	for _, s := range stmts {
		switch st := s.(type) {
		case *CommandStatement:
			if strings.TrimSpace(st.Raw) == "" {
				errs = append(errs, ValidationError{
					Pos:  st.SrcPos,
					Code: "DLG-E007",
					Msg:  "empty <<>> command expression",
				})
			}
		case *SpeakerLine:
			for _, c := range st.Commands {
				if strings.TrimSpace(c.Raw) == "" {
					errs = append(errs, ValidationError{
						Pos:  c.SrcPos,
						Code: "DLG-E007",
						Msg:  "empty <<>> command expression",
					})
				}
			}
		case *NarrationLine:
			for _, c := range st.Commands {
				if strings.TrimSpace(c.Raw) == "" {
					errs = append(errs, ValidationError{
						Pos:  c.SrcPos,
						Code: "DLG-E007",
						Msg:  "empty <<>> command expression",
					})
				}
			}
		case *OptionBranch:
			for _, c := range st.Commands {
				if strings.TrimSpace(c.Raw) == "" {
					errs = append(errs, ValidationError{
						Pos:  c.SrcPos,
						Code: "DLG-E007",
						Msg:  "empty <<>> command expression",
					})
				}
			}
			errs = append(errs, checkEmptyCommands(st.Body)...)
		}
	}
	return errs
}

// checkOptionBodies flags options whose body is empty or contains no terminal
// (DLG-E008: option with no content/jump).
func checkOptionBodies(stmts []Statement) []ValidationError {
	var errs []ValidationError
	for _, s := range stmts {
		ob, ok := s.(*OptionBranch)
		if !ok {
			continue
		}
		if len(ob.Body) == 0 {
			errs = append(errs, ValidationError{
				Pos:  ob.SrcPos,
				Code: "DLG-E008",
				Msg:  fmt.Sprintf("option %q has no body content", ob.Text),
			})
		}
		// Recurse into nested options.
		errs = append(errs, checkOptionBodies(ob.Body)...)
	}
	return errs
}

// bodyHasTerminal returns true if any path through the statement list reaches
// a <<end>> or <<jump …>> command. A node with no options and no such command
// is considered to fall off the end (no terminal).
func bodyHasTerminal(stmts []Statement) bool {
	for _, s := range stmts {
		switch st := s.(type) {
		case *CommandStatement:
			if isTerminalCmd(st.Raw) {
				return true
			}
		case *OptionBranch:
			if bodyHasTerminal(st.Body) {
				return true
			}
		}
	}
	return false
}

// isTerminalCmd returns true for <<end>> and <<jump …>> commands.
func isTerminalCmd(raw string) bool {
	raw = strings.TrimSpace(raw)
	return raw == "end" || strings.HasPrefix(raw, "jump")
}

// detectCircularJumps finds nodes that form pure jump cycles with no <<end>>
// reachable from any node in the cycle (DLG-E009).
func detectCircularJumps(lp *LinkedProject) []ValidationError {
	var errs []ValidationError

	// For each node that has only <<jump>> as its terminal and no <<end>>,
	// follow the jump chain to detect a cycle.
	visited := make(map[string]bool)

	for title := range lp.NodesByTitle {
		if visited[title] {
			continue
		}
		chain := []string{}
		inChain := make(map[string]bool)
		if cycle, cycleStart := followJumpChain(lp, title, chain, inChain, visited); cycle {
			ln := lp.NodesByTitle[cycleStart]
			errs = append(errs, ValidationError{
				Pos:  ln.Node.Pos,
				Code: "DLG-E009",
				Msg:  fmt.Sprintf("circular jump with no exit starting at node %q", cycleStart),
			})
		}
	}
	return errs
}

// followJumpChain walks the jump chain from start, returning (true, cycleNode)
// if a cycle is detected. All nodes visited are marked in the global `visited`
// set to avoid redundant checks.
func followJumpChain(lp *LinkedProject, title string, chain []string, inChain, visited map[string]bool) (bool, string) {
	ln, ok := lp.NodesByTitle[title]
	if !ok || visited[title] {
		return false, ""
	}

	// If this node has any <<end>> reachable, the chain can terminate here.
	if bodyHasEnd(ln.Node.Body) {
		visited[title] = true
		return false, ""
	}

	// Detect cycle.
	if inChain[title] {
		return true, title
	}

	inChain[title] = true
	chain = append(chain, title)

	// Find the sole jump target (if the node only jumps).
	target := soleJumpTarget(ln.Node.Body)
	if target == "" {
		// No jump — node falls off end (caught by DLG-E003, not E009).
		visited[title] = true
		inChain[title] = false
		return false, ""
	}

	cycle, cycleStart := followJumpChain(lp, target, chain, inChain, visited)
	inChain[title] = false
	visited[title] = true
	return cycle, cycleStart
}

// bodyHasEnd returns true if the body contains any <<end>> command at any depth.
func bodyHasEnd(stmts []Statement) bool {
	for _, s := range stmts {
		switch st := s.(type) {
		case *CommandStatement:
			if strings.TrimSpace(st.Raw) == "end" {
				return true
			}
		case *OptionBranch:
			if bodyHasEnd(st.Body) {
				return true
			}
		}
	}
	return false
}

// soleJumpTarget returns the jump target if the node body contains exactly one
// terminal and it is a <<jump>>, otherwise returns "".
func soleJumpTarget(stmts []Statement) string {
	for _, s := range stmts {
		if cs, ok := s.(*CommandStatement); ok {
			if target, ok := jumpTarget(cs.Raw); ok {
				return target
			}
		}
	}
	return ""
}
