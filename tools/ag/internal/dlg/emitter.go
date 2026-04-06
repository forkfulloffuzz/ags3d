package dlg

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
)

// EmittedFile is the compiled JSON representation of one .agdlg source file.
// Written to .engine/generated/dialogue/<rel_path>.json.
type EmittedFile struct {
	Source string         `json:"source"` // relative path to the .agdlg source
	Nodes  []*EmittedNode `json:"nodes"`
}

// EmittedNode is one compiled dialogue node.
type EmittedNode struct {
	Title     string          `json:"title"`
	Character string          `json:"character,omitempty"`
	Tags      []string        `json:"tags,omitempty"`
	Inherits  []string        `json:"inherits,omitempty"`
	Suppress  []string        `json:"suppress,omitempty"`
	LocID     string          `json:"loc_id,omitempty"`
	Body      []EmittedStmt   `json:"body"`
}

// EmittedStmt is a tagged union; the "type" field discriminates.
type EmittedStmt struct {
	Type     string        `json:"type"`
	Speaker  string        `json:"speaker,omitempty"`
	Text     string        `json:"text,omitempty"`
	LocKey   string        `json:"loc_key,omitempty"`
	Commands []string      `json:"commands,omitempty"`
	Raw      string        `json:"raw,omitempty"`    // for command statements
	Body     []EmittedStmt `json:"body,omitempty"`  // for option branches
}

// EmitProject compiles a validated LinkedProject and writes one JSON file per
// source .agdlg file into outputDir (typically .engine/generated/dialogue/).
// Errors from individual files are collected and returned together.
func EmitProject(lp *LinkedProject, outputDir string) []error {
	var errs []error
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return []error{fmt.Errorf("dlg emit: cannot create output directory: %w", err)}
	}
	for _, f := range lp.Files {
		if err := emitFile(lp, f, outputDir); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// emitFile compiles one DialogueFile and writes its JSON representation.
func emitFile(lp *LinkedProject, df *DialogueFile, outputDir string) error {
	ef := &EmittedFile{
		Source: df.Path,
	}
	for _, n := range df.Nodes {
		en := emitNode(n)
		ef.Nodes = append(ef.Nodes, en)
	}

	// Derive output path: preserve relative directory structure.
	// df.Path may be absolute; take the base name for simple layout.
	base := filepath.Base(df.Path)
	base = strings.TrimSuffix(base, ".agdlg") + ".json"
	outPath := filepath.Join(outputDir, base)

	data, err := json.MarshalIndent(ef, "", "  ")
	if err != nil {
		return fmt.Errorf("dlg emit: json marshal %s: %w", df.Path, err)
	}
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("dlg emit: write %s: %w", outPath, err)
	}
	return nil
}

// emitNode converts a DialogueNode into an EmittedNode, assigning automatic
// loc keys to any lines or options that don't already have one.
func emitNode(n *DialogueNode) *EmittedNode {
	en := &EmittedNode{
		Title:     n.Title,
		Character: n.Character,
		Tags:      n.Tags,
		Inherits:  n.Inherits,
		Suppress:  n.Suppress,
		LocID:     n.LocID,
	}
	var lineIdx int
	en.Body = emitStmts(n.Title, n.Body, &lineIdx)
	return en
}

// emitStmts converts a slice of Statements into EmittedStmts.
// lineIdx is incremented for each line/option that gets a loc key.
func emitStmts(nodeTitle string, stmts []Statement, lineIdx *int) []EmittedStmt {
	var out []EmittedStmt
	for _, s := range stmts {
		out = append(out, emitStmt(nodeTitle, s, lineIdx))
	}
	return out
}

func emitStmt(nodeTitle string, s Statement, lineIdx *int) EmittedStmt {
	switch st := s.(type) {
	case *SpeakerLine:
		locKey := st.LocKey
		if locKey == "" {
			locKey = autoLocKey(nodeTitle, *lineIdx, st.Text)
		}
		*lineIdx++
		return EmittedStmt{
			Type:     "speaker_line",
			Speaker:  st.Speaker,
			Text:     st.Text,
			LocKey:   locKey,
			Commands: cmdRaws(st.Commands),
		}
	case *NarrationLine:
		locKey := st.LocKey
		if locKey == "" {
			locKey = autoLocKey(nodeTitle, *lineIdx, st.Text)
		}
		*lineIdx++
		return EmittedStmt{
			Type:     "narration",
			Text:     st.Text,
			LocKey:   locKey,
			Commands: cmdRaws(st.Commands),
		}
	case *OptionBranch:
		locKey := st.LocKey
		if locKey == "" {
			locKey = autoLocKey(nodeTitle, *lineIdx, st.Text)
		}
		*lineIdx++
		body := emitStmts(nodeTitle, st.Body, lineIdx)
		return EmittedStmt{
			Type:     "option",
			Text:     st.Text,
			LocKey:   locKey,
			Commands: cmdRaws(st.Commands),
			Body:     body,
		}
	case *CommandStatement:
		return EmittedStmt{
			Type: "command",
			Raw:  st.Raw,
		}
	default:
		return EmittedStmt{Type: "unknown"}
	}
}

// cmdRaws extracts the Raw strings from a CommandExpr slice.
func cmdRaws(cmds []CommandExpr) []string {
	if len(cmds) == 0 {
		return nil
	}
	out := make([]string, len(cmds))
	for i, c := range cmds {
		out[i] = c.Raw
	}
	return out
}

// autoLocKey generates a stable loc key in the format
// "{nodeTitle}:{lineIndex}:{textHash8}" where textHash8 is the first 8 hex
// digits of an FNV-32a hash of the text.
func autoLocKey(nodeTitle string, lineIdx int, text string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(text))
	return fmt.Sprintf("%s:%d:%08x", nodeTitle, lineIdx, h.Sum32())
}
