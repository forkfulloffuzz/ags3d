package dlg

// DialogueFile is the parsed result of a single .agdlg source file.
// It contains one or more dialogue nodes.
type DialogueFile struct {
	Path  string          // absolute path to the source file
	Nodes []*DialogueNode // in source order
}

// DialogueNode corresponds to one titled section between "---" and "===".
type DialogueNode struct {
	// Header fields.
	Title     string   // required — unique across the project
	Character string   // character name (may be empty for narrator-only nodes)
	Tags      []string // e.g. ["chapter:1", "cinematic"]
	Inherits  []string // list of node titles whose global options this node inherits
	Suppress  []string // list of global option titles suppressed in this node
	LocID     string   // stable loc namespace prefix for this node

	// Source location of the title declaration (for error reporting).
	Pos Pos

	// Body — top-level statements in the node body.
	Body []Statement
}

// Statement is implemented by all body-level AST node types.
type Statement interface {
	statPos() Pos
}

// SpeakerLine is a "Speaker: dialogue text" line.
type SpeakerLine struct {
	Speaker  string
	Text     string
	Commands []CommandExpr // inline <<commands>> on this line
	LocKey   string        // #loc: annotation if present
	SrcPos   Pos
}

func (s *SpeakerLine) statPos() Pos { return s.SrcPos }

// NarrationLine is a plain text line with no speaker prefix.
type NarrationLine struct {
	Text     string
	Commands []CommandExpr
	LocKey   string
	SrcPos   Pos
}

func (n *NarrationLine) statPos() Pos { return n.SrcPos }

// OptionBranch is a "-> option text" line plus its indented body.
type OptionBranch struct {
	Text     string        // display text of the option
	Commands []CommandExpr // inline <<conditions>> on the option line (e.g. <<visible_if …>>)
	LocKey   string
	Body     []Statement // nested statements under this option (indented block)
	Depth    int         // indent depth of the "-> " marker
	SrcPos   Pos
}

func (o *OptionBranch) statPos() Pos { return o.SrcPos }

// CommandStatement is a standalone <<command>> line.
type CommandStatement struct {
	Raw    string // full content inside << >>
	SrcPos Pos
}

func (c *CommandStatement) statPos() Pos { return c.SrcPos }

// CommandExpr is an inline <<command>> extracted from a line.
type CommandExpr struct {
	Raw    string
	SrcPos Pos
}
