package cut

import (
	"strings"
)

// ---- Argument tokenizer -----------------------------------------------
//
// Tokenizes the Args string of a RawCommand into a flat []ArgToken list.
// This is used by T-CUT02's command parsers to extract named parameters,
// string values, and identifiers without the caller needing to hand-parse
// the raw string.
//
// Rules (in priority order):
//   - "quoted string"      → TokStringValue (content without quotes)
//   - word:value           → TokNamedParam  (colon present, no spaces around colon)
//   - any other word       → TokIdentifier

// ArgToken is a single token from a command's argument string.
type ArgToken struct {
	Kind  TokenKind
	Value string
}

// TokenizeArgs splits a raw argument string into ArgTokens.
func TokenizeArgs(args string) []ArgToken {
	var tokens []ArgToken
	s := strings.TrimSpace(args)

	for len(s) > 0 {
		s = strings.TrimLeft(s, " \t")
		if len(s) == 0 {
			break
		}

		// Quoted string: "..."
		if s[0] == '"' {
			end := strings.Index(s[1:], `"`)
			if end < 0 {
				// Unterminated string — consume remainder.
				tokens = append(tokens, ArgToken{Kind: TokStringValue, Value: s[1:]})
				break
			}
			tokens = append(tokens, ArgToken{Kind: TokStringValue, Value: s[1 : end+1]})
			s = s[end+2:]
			continue
		}

		// Read until next whitespace or end.
		end := strings.IndexAny(s, " \t")
		var word string
		if end < 0 {
			word = s
			s = ""
		} else {
			word = s[:end]
			s = s[end:]
		}

		// Classify: named param if it contains ":" and the part before ":" has
		// no special characters (just an identifier or dotted name).
		if isNamedParam(word) {
			tokens = append(tokens, ArgToken{Kind: TokNamedParam, Value: word})
		} else {
			tokens = append(tokens, ArgToken{Kind: TokIdentifier, Value: word})
		}
	}
	return tokens
}

// isNamedParam returns true when word looks like "key:value".
// A named param has exactly one colon, with a non-empty part on each side,
// and the key part contains only letters, digits, underscores, or hyphens.
func isNamedParam(word string) bool {
	idx := strings.Index(word, ":")
	if idx <= 0 || idx == len(word)-1 {
		return false
	}
	key := word[:idx]
	for _, ch := range key {
		if !isIdentChar(ch) {
			return false
		}
	}
	return true
}

func isIdentChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '-'
}

// NamedParamValue extracts the value from a "key:value" ArgToken.
func NamedParamValue(t ArgToken) (key, value string) {
	idx := strings.Index(t.Value, ":")
	if idx < 0 {
		return t.Value, ""
	}
	return t.Value[:idx], t.Value[idx+1:]
}

// ---- Typed command vocabulary -----------------------------------------

// Command is the fully-parsed representation of one cutscene step.
type Command struct {
	// Name is the command verb (same as RawCommand.Name).
	Name string
	// Params holds named key:value parameters (e.g. duration:"2.0", bg:"cam_move").
	Params map[string]string
	// Positional holds un-named identifier tokens in order.
	Positional []string
	// Text is the first STRING_VALUE argument (for line, title_card, etc.).
	Text string
	// IsBlockOpen / IsBlockClose mirror the RawCommand flags.
	IsBlockOpen  bool
	IsBlockClose bool
	// Condition holds the raw condition expression for <<if condition>>.
	Condition string
	// Expr holds the raw AGS-spirit expression for <<action expr>>.
	Expr string
	// Raw is the original RawCommand for error reporting.
	Raw *RawCommand
}

// ParseCommand converts a RawCommand into a fully-parsed Command.
// For commands whose Args contain a raw expression (action, set, if),
// the expression is stored verbatim in Cmd.Expr / Cmd.Condition.
func ParseCommand(rc *RawCommand) *Command {
	cmd := &Command{
		Name:         rc.Name,
		IsBlockOpen:  rc.IsBlockOpen,
		IsBlockClose: rc.IsBlockClose,
		Params:       make(map[string]string),
		Raw:          rc,
	}

	// Commands whose Args field is a raw expression (not token-parseable).
	switch rc.Name {
	case "action", "set":
		cmd.Expr = rc.Args
		return cmd
	case "if", "else_if":
		cmd.Condition = rc.Args
		return cmd
	}

	// Tokenize the argument string.
	for _, tok := range TokenizeArgs(rc.Args) {
		switch tok.Kind {
		case TokStringValue:
			if cmd.Text == "" {
				cmd.Text = tok.Value
			} else {
				// Multiple string values — append positional.
				cmd.Positional = append(cmd.Positional, tok.Value)
			}
		case TokNamedParam:
			k, v := NamedParamValue(tok)
			cmd.Params[k] = v
		case TokIdentifier:
			cmd.Positional = append(cmd.Positional, tok.Value)
		}
	}
	return cmd
}

// ParseSequence converts a flat []RawCommand list into fully-parsed Commands.
func ParseSequence(seq []*RawCommand) []*Command {
	out := make([]*Command, len(seq))
	for i, rc := range seq {
		out[i] = ParseCommand(rc)
	}
	return out
}

// ---- CutsceneSequence: tree structure --------------------------------

// Sequence is the parsed, tree-structured representation of a cutscene sequence.
// Block commands (parallel, if, on) are nested into their body Steps.
type Sequence struct {
	Steps []*Step
}

// Step is a single element in a Sequence.
type Step struct {
	Cmd  *Command
	Body *Sequence // non-nil for block commands (parallel, if, on)
	Else *Sequence // non-nil for <<else>> branch of <<if>>
}

// ParseSequenceTree takes a flat parsed command list and builds the nested
// Step tree. Mismatched block open/close counts are handled gracefully.
func ParseSequenceTree(cmds []*Command) *Sequence {
	seq, _, _ := parseBlock(cmds, 0)
	return seq
}

// parseBlock parses from index i until a block-close or else/else_if command,
// or end of slice. Returns (sequence, terminating command or nil, next index).
// The terminating command is consumed and returned so the caller can inspect it
// (e.g. to detect <<else>> vs <<end_if>>).
func parseBlock(cmds []*Command, i int) (*Sequence, *Command, int) {
	seq := &Sequence{}
	for i < len(cmds) {
		cmd := cmds[i]

		// Block close or else — stop and return the terminator to the caller.
		if cmd.IsBlockClose || cmd.Name == "else" || cmd.Name == "else_if" {
			return seq, cmd, i + 1
		}

		step := &Step{Cmd: cmd}

		if cmd.IsBlockOpen {
			// Parse the body block.
			var body *Sequence
			var termCmd *Command
			body, termCmd, i = parseBlock(cmds, i+1)
			step.Body = body

			// Handle <<if>> / <<else>> / <<end_if>> structure.
			if (cmd.Name == "if" || cmd.Name == "else_if") &&
				termCmd != nil && (termCmd.Name == "else" || termCmd.Name == "else_if") {
				var elseBranch *Sequence
				elseBranch, _, i = parseBlock(cmds, i)
				step.Else = elseBranch
			}
		} else {
			i++
		}
		seq.Steps = append(seq.Steps, step)
	}
	return seq, nil, i
}
