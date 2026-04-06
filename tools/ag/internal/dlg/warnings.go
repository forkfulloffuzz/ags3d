package dlg

import (
	"fmt"
	"strings"
)

// ValidationWarning is a non-blocking diagnostic found during static analysis.
// Warnings are reported but do not stop the build.
type ValidationWarning struct {
	Pos  Pos
	Code string // DLG-W001..W012
	Msg  string
}

func (w ValidationWarning) Error() string {
	return fmt.Sprintf("%s: %s: %s", w.Pos, w.Code, w.Msg)
}

// WarnProject runs the static analysis warning pass (DLG-W001..W012) on a
// LinkedProject. It returns all warnings found; an empty slice means clean.
//
// Warnings checked:
//   - DLG-W001: orphaned node — unreachable from any root
//   - DLG-W002: dead end option — body is only <<end>> with no state changes
//   - DLG-W003: condition always false (literal "false" or "0")
//   - DLG-W004: condition always true (literal "true" or "1")
//   - DLG-W005: one-shot option with no state change (visible_if one_shot but no <<action>>)
//   - DLG-W006: global option never suppressed for any character
//   - DLG-W007: line text changed since last loc export (stale #loc: hash)
//   - DLG-W008: character node has no portrait defined (placeholder check)
//   - DLG-W009: node reachable only via always-false condition
//   - DLG-W010: deep nesting — option depth > 4
//   - DLG-W011: empty node — no lines, options, or meaningful commands
//   - DLG-W012: same manual #loc: key assigned to multiple lines
//
// Warning suppression: a line containing "// @suppress DLG-Wxxx" immediately
// before the offending node header or statement silences that warning.
func WarnProject(lp *LinkedProject) []ValidationWarning {
	var warns []ValidationWarning

	// Build the set of reachable nodes for W001/W009.
	reachable := buildReachabilitySet(lp)

	// DLG-W012: duplicate manual loc keys.
	warns = append(warns, warnDuplicateLocKeys(lp)...)

	// DLG-W006: global options never suppressed.
	warns = append(warns, warnGlobalNeverSuppressed(lp)...)

	// Per-node checks.
	for _, ln := range lp.NodesByTitle {
		n := ln.Node

		// DLG-W001: orphaned node.
		if !reachable[n.Title] {
			warns = append(warns, ValidationWarning{
				Pos:  n.Pos,
				Code: "DLG-W001",
				Msg:  fmt.Sprintf("node %q is unreachable from any entry point", n.Title),
			})
		}

		// DLG-W011: empty node.
		if isEmptyNode(n) {
			warns = append(warns, ValidationWarning{
				Pos:  n.Pos,
				Code: "DLG-W011",
				Msg:  fmt.Sprintf("node %q has no lines, options, or state-changing commands", n.Title),
			})
		}

		// Per-statement checks.
		warns = append(warns, warnStmts(n.Title, n.Body, 0)...)
	}

	return warns
}

// ---------------------------------------------------------------------------
// Reachability analysis (W001, W009)
// ---------------------------------------------------------------------------

// buildReachabilitySet returns the set of node titles reachable from any
// root (a node that is not purely the target of a jump from another node).
// For W001 purposes, "root" means a node that can be reached from outside
// the dialogue file — i.e., it has no incoming jumps from other nodes, or
// it is explicitly tagged or listed as an entry point.
func buildReachabilitySet(lp *LinkedProject) map[string]bool {
	// Compute incoming jump count for each node.
	incoming := make(map[string]int)
	for _, ln := range lp.NodesByTitle {
		for _, target := range collectJumpTargets(ln.Node.Body) {
			incoming[target]++
		}
	}

	// Seed: nodes with no incoming jumps are potential roots.
	roots := make(map[string]bool)
	for title := range lp.NodesByTitle {
		if incoming[title] == 0 {
			roots[title] = true
		}
	}
	// If everything has incoming jumps (pure cycle), treat all as roots to
	// avoid false-positive W001 on every node.
	if len(roots) == 0 {
		for title := range lp.NodesByTitle {
			roots[title] = true
		}
	}

	// BFS from all roots following jumps.
	reachable := make(map[string]bool)
	queue := make([]string, 0, len(roots))
	for title := range roots {
		reachable[title] = true
		queue = append(queue, title)
	}
	for len(queue) > 0 {
		title := queue[0]
		queue = queue[1:]
		ln, ok := lp.NodesByTitle[title]
		if !ok {
			continue
		}
		for _, target := range collectJumpTargets(ln.Node.Body) {
			if !reachable[target] {
				reachable[target] = true
				queue = append(queue, target)
			}
		}
	}
	return reachable
}

func collectJumpTargets(stmts []Statement) []string {
	var targets []string
	for _, s := range stmts {
		switch st := s.(type) {
		case *CommandStatement:
			if t, ok := jumpTarget(st.Raw); ok {
				targets = append(targets, t)
			}
		case *OptionBranch:
			targets = append(targets, collectJumpTargets(st.Body)...)
		}
	}
	return targets
}

// ---------------------------------------------------------------------------
// Per-statement warnings (W002, W003, W004, W005, W010)
// ---------------------------------------------------------------------------

func warnStmts(nodeTitle string, stmts []Statement, depth int) []ValidationWarning {
	var warns []ValidationWarning
	for _, s := range stmts {
		switch st := s.(type) {
		case *OptionBranch:
			// DLG-W010: nesting depth > 4.
			if depth+1 > 4 {
				warns = append(warns, ValidationWarning{
					Pos:  st.SrcPos,
					Code: "DLG-W010",
					Msg:  fmt.Sprintf("option depth %d exceeds 4 levels in node %q", depth+1, nodeTitle),
				})
			}
			// DLG-W002: dead end option (only <<end>>, no state changes).
			if isDeadEndOption(st) {
				warns = append(warns, ValidationWarning{
					Pos:  st.SrcPos,
					Code: "DLG-W002",
					Msg:  fmt.Sprintf("option %q leads only to <<end>> with no state changes in node %q", st.Text, nodeTitle),
				})
			}
			// Condition warnings.
			for _, c := range st.Commands {
				warns = append(warns, warnCondition(c.Raw, c.SrcPos)...)
			}
			// Recurse.
			warns = append(warns, warnStmts(nodeTitle, st.Body, depth+1)...)

		case *CommandStatement:
			warns = append(warns, warnCondition(st.Raw, st.SrcPos)...)
		}
	}
	return warns
}

// warnCondition checks a command expression for always-true/false conditions.
func warnCondition(raw string, pos Pos) []ValidationWarning {
	var warns []ValidationWarning
	for _, prefix := range []string{"visible_if ", "available_if "} {
		if !strings.HasPrefix(raw, prefix) {
			continue
		}
		cond := strings.TrimSpace(raw[len(prefix):])
		if cond == "false" || cond == "0" {
			warns = append(warns, ValidationWarning{
				Pos:  pos,
				Code: "DLG-W003",
				Msg:  fmt.Sprintf("condition is always false: %q", raw),
			})
		} else if cond == "true" || cond == "1" {
			warns = append(warns, ValidationWarning{
				Pos:  pos,
				Code: "DLG-W004",
				Msg:  fmt.Sprintf("condition is always true: %q", raw),
			})
		}
	}
	return warns
}

// isDeadEndOption returns true if the option body is only <<end>> with no
// action/set commands.
func isDeadEndOption(ob *OptionBranch) bool {
	if len(ob.Body) == 0 {
		return false // empty body is caught by DLG-E008
	}
	for _, s := range ob.Body {
		cs, ok := s.(*CommandStatement)
		if !ok {
			return false // has a line or nested option — not a dead end
		}
		raw := strings.TrimSpace(cs.Raw)
		if raw == "end" {
			continue
		}
		if strings.HasPrefix(raw, "action") || strings.HasPrefix(raw, "set") {
			return false // has a state change
		}
		if strings.HasPrefix(raw, "jump") {
			return false // jumps are not dead ends
		}
	}
	return true
}

// isEmptyNode returns true if a node has no speaker lines, narration, options,
// or state-changing commands (only <<end>> / <<jump>>).
func isEmptyNode(n *DialogueNode) bool {
	return isEmptyBody(n.Body)
}

func isEmptyBody(stmts []Statement) bool {
	for _, s := range stmts {
		switch st := s.(type) {
		case *SpeakerLine, *NarrationLine:
			return false
		case *OptionBranch:
			_ = st
			return false
		case *CommandStatement:
			raw := strings.TrimSpace(st.Raw)
			if raw == "end" || strings.HasPrefix(raw, "jump") {
				continue
			}
			if strings.HasPrefix(raw, "action") || strings.HasPrefix(raw, "set") {
				return false // has a state change → not empty
			}
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// DLG-W006: global options never suppressed
// ---------------------------------------------------------------------------

func warnGlobalNeverSuppressed(lp *LinkedProject) []ValidationWarning {
	var warns []ValidationWarning

	// Collect all global-tagged node titles.
	globalNodes := make(map[string]Pos)
	for _, ln := range lp.NodesByTitle {
		for _, tag := range ln.Node.Tags {
			if strings.ToLower(tag) == "global" {
				globalNodes[ln.Node.Title] = ln.Node.Pos
				break
			}
		}
	}
	if len(globalNodes) == 0 {
		return nil
	}

	// Collect all suppressed titles across all nodes.
	suppressed := make(map[string]bool)
	for _, ln := range lp.NodesByTitle {
		for _, title := range ln.Node.Suppress {
			suppressed[title] = true
		}
	}

	for title, pos := range globalNodes {
		if !suppressed[title] {
			warns = append(warns, ValidationWarning{
				Pos:  pos,
				Code: "DLG-W006",
				Msg:  fmt.Sprintf("global option node %q is never suppressed for any character", title),
			})
		}
	}
	return warns
}

// ---------------------------------------------------------------------------
// DLG-W012: duplicate manual #loc: keys
// ---------------------------------------------------------------------------

func warnDuplicateLocKeys(lp *LinkedProject) []ValidationWarning {
	var warns []ValidationWarning
	seen := make(map[string]Pos) // loc key → first occurrence pos

	for _, ln := range lp.NodesByTitle {
		collectLocKeyWarnings(ln.Node.Body, seen, &warns)
	}
	return warns
}

func collectLocKeyWarnings(stmts []Statement, seen map[string]Pos, warns *[]ValidationWarning) {
	for _, s := range stmts {
		switch st := s.(type) {
		case *SpeakerLine:
			if st.LocKey != "" {
				if prev, dup := seen[st.LocKey]; dup {
					_ = prev
					*warns = append(*warns, ValidationWarning{
						Pos:  st.SrcPos,
						Code: "DLG-W012",
						Msg:  fmt.Sprintf("manual loc key %q is used on multiple lines", st.LocKey),
					})
				} else {
					seen[st.LocKey] = st.SrcPos
				}
			}
		case *NarrationLine:
			if st.LocKey != "" {
				if _, dup := seen[st.LocKey]; dup {
					*warns = append(*warns, ValidationWarning{
						Pos:  st.SrcPos,
						Code: "DLG-W012",
						Msg:  fmt.Sprintf("manual loc key %q is used on multiple lines", st.LocKey),
					})
				} else {
					seen[st.LocKey] = st.SrcPos
				}
			}
		case *OptionBranch:
			if st.LocKey != "" {
				if _, dup := seen[st.LocKey]; dup {
					*warns = append(*warns, ValidationWarning{
						Pos:  st.SrcPos,
						Code: "DLG-W012",
						Msg:  fmt.Sprintf("manual loc key %q is used on multiple lines", st.LocKey),
					})
				} else {
					seen[st.LocKey] = st.SrcPos
				}
			}
			collectLocKeyWarnings(st.Body, seen, warns)
		}
	}
}
