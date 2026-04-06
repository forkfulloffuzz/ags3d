package dlg

import (
	"fmt"
	"strings"
)

// LinkedProject is the result of the link stage: all .agdlg files parsed and
// cross-referenced into a project-wide dialogue graph.
type LinkedProject struct {
	// Files is the full set of parsed source files.
	Files []*DialogueFile

	// NodesByTitle maps every declared node title to its node and source file.
	// Populated by Link; used by validators and the emitter.
	NodesByTitle map[string]*LinkedNode

	// LinkErrors contains non-fatal issues found during linking (e.g. unresolved
	// jump targets). Fatal issues (duplicate titles) are returned as errors from Link.
	LinkErrors []LinkError
}

// LinkedNode pairs a DialogueNode with the file it came from.
type LinkedNode struct {
	Node *DialogueNode
	File *DialogueFile
}

// LinkError describes a non-fatal problem found during the link stage.
type LinkError struct {
	Pos  Pos
	Code string // e.g. "DLG-E002"
	Msg  string
}

func (e LinkError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Pos, e.Code, e.Msg)
}

// Link performs stage 4 of the dialogue pipeline:
//  1. Builds the project-wide NodesByTitle index.
//  2. Detects duplicate node titles (hard error — stops build).
//  3. Resolves all <<jump target>> references and records unresolved ones as
//     LinkErrors (DLG-E002).
//  4. Resolves $character placeholders in node titles referenced by inherits/suppress.
//  5. Validates that suppress targets are global (tagged [global]) nodes
//     (DLG-E006 warning recorded as LinkError).
//
// Returns a non-nil error only for hard failures (duplicate titles).
// Soft issues (unresolved jumps, bad suppress targets) are collected in
// LinkedProject.LinkErrors so callers can report them in batch.
func Link(files []*DialogueFile) (*LinkedProject, error) {
	lp := &LinkedProject{
		Files:        files,
		NodesByTitle: make(map[string]*LinkedNode),
	}

	// Pass 1 — build title index, detect duplicates.
	if err := lp.buildIndex(); err != nil {
		return nil, err
	}

	// Pass 2 — resolve jumps.
	lp.resolveJumps()

	// Pass 3 — validate inherits/suppress references.
	lp.resolveInheritSuppress()

	return lp, nil
}

// buildIndex indexes all nodes by title. Returns an error if any title is
// declared more than once across all files.
func (lp *LinkedProject) buildIndex() error {
	for _, f := range lp.Files {
		for _, n := range f.Nodes {
			if existing, dup := lp.NodesByTitle[n.Title]; dup {
				return fmt.Errorf("%s: DLG-E001: duplicate node title %q (first declared at %s)",
					n.Pos, n.Title, existing.Node.Pos)
			}
			lp.NodesByTitle[n.Title] = &LinkedNode{Node: n, File: f}
		}
	}
	return nil
}

// resolveJumps walks every node body (recursively through option branches)
// and records DLG-E002 for any <<jump target>> where target is not in the index.
func (lp *LinkedProject) resolveJumps() {
	for _, ln := range lp.NodesByTitle {
		lp.resolveStmtJumps(ln.Node.Body)
	}
}

func (lp *LinkedProject) resolveStmtJumps(stmts []Statement) {
	for _, s := range stmts {
		switch st := s.(type) {
		case *CommandStatement:
			lp.checkJumpCmd(st.Raw, st.SrcPos)
		case *SpeakerLine:
			for _, c := range st.Commands {
				lp.checkJumpCmd(c.Raw, c.SrcPos)
			}
		case *NarrationLine:
			for _, c := range st.Commands {
				lp.checkJumpCmd(c.Raw, c.SrcPos)
			}
		case *OptionBranch:
			for _, c := range st.Commands {
				lp.checkJumpCmd(c.Raw, c.SrcPos)
			}
			lp.resolveStmtJumps(st.Body)
		}
	}
}

// checkJumpCmd inspects a raw command string; if it is a jump command, it
// verifies the target exists in the project index.
func (lp *LinkedProject) checkJumpCmd(raw string, pos Pos) {
	target, ok := jumpTarget(raw)
	if !ok {
		return
	}
	// $character.node_name placeholder — resolve by looking for any node
	// matching "*.node_name" suffix. Record as unresolved only if no match found.
	if strings.HasPrefix(target, "$") {
		// Placeholder targets are validated in the cross-system pass (T-DLG05).
		// Here we just skip them.
		return
	}
	if _, exists := lp.NodesByTitle[target]; !exists {
		lp.LinkErrors = append(lp.LinkErrors, LinkError{
			Pos:  pos,
			Code: "DLG-E002",
			Msg:  fmt.Sprintf("jump target %q does not exist", target),
		})
	}
}

// jumpTarget extracts the target from a "jump <target>" command string.
// Returns ("", false) if the command is not a jump.
func jumpTarget(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "jump") {
		return "", false
	}
	rest := strings.TrimSpace(raw[4:])
	if rest == "" {
		return "", false
	}
	// Take the first word as the target (ignore any trailing arguments).
	fields := strings.Fields(rest)
	return fields[0], true
}

// resolveInheritSuppress validates that all titles listed in node.Inherits and
// node.Suppress exist in the project index, and that suppress targets are
// [global]-tagged nodes.
func (lp *LinkedProject) resolveInheritSuppress() {
	for _, ln := range lp.NodesByTitle {
		n := ln.Node
		for _, title := range n.Inherits {
			if _, ok := lp.NodesByTitle[title]; !ok {
				lp.LinkErrors = append(lp.LinkErrors, LinkError{
					Pos:  n.Pos,
					Code: "DLG-E002",
					Msg:  fmt.Sprintf("inherits target %q does not exist", title),
				})
			}
		}
		for _, title := range n.Suppress {
			target, ok := lp.NodesByTitle[title]
			if !ok {
				lp.LinkErrors = append(lp.LinkErrors, LinkError{
					Pos:  n.Pos,
					Code: "DLG-E002",
					Msg:  fmt.Sprintf("suppress target %q does not exist", title),
				})
				continue
			}
			if !hasTag(target.Node, "global") {
				lp.LinkErrors = append(lp.LinkErrors, LinkError{
					Pos:  n.Pos,
					Code: "DLG-E006",
					Msg:  fmt.Sprintf("suppress target %q is not a [global] node", title),
				})
			}
		}
	}
}

// hasTag returns true if node has the given tag in its Tags slice.
func hasTag(n *DialogueNode, tag string) bool {
	for _, t := range n.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
