package dlg

import (
	"fmt"
	"os"
	"strings"

	"github.com/ags3d/ag/internal/cut"
)

// ParseFile lexes and parses a .agdlg source file from disk.
func ParseFile(path string) (*DialogueFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dlg: cannot read %s: %w", path, err)
	}
	return Parse(path, string(data))
}

// Parse lexes and parses a .agdlg source string.
// filename is used for position reporting only.
func Parse(filename, src string) (*DialogueFile, error) {
	tokens, err := Lex(filename, src)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens, pos: 0}
	return p.parseFile(filename)
}

// parser walks a token slice and builds the AST.
type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Kind: TokEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) consume() Token {
	t := p.peek()
	p.pos++
	return t
}

// skip advances past any TokComment tokens.
func (p *parser) skipComments() {
	for p.peek().Kind == TokComment {
		p.pos++
	}
}

// parseFile parses the full token stream into a DialogueFile.
func (p *parser) parseFile(filename string) (*DialogueFile, error) {
	df := &DialogueFile{Path: filename}
	for {
		p.skipComments()
		if p.peek().Kind == TokEOF {
			break
		}
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		df.Nodes = append(df.Nodes, node)
	}
	return df, nil
}

// parseNode parses one dialogue node: header section + body section + ===.
func (p *parser) parseNode() (*DialogueNode, error) {
	node := &DialogueNode{}

	// --- Header ---
	if err := p.parseHeader(node); err != nil {
		return nil, err
	}

	// title is required
	if node.Title == "" {
		return nil, &LexError{Pos: node.Pos, Msg: "dialogue node is missing required 'title' header field"}
	}

	// --- Body (top-level statements, depth 0) ---
	body, err := p.parseBody(0)
	if err != nil {
		return nil, err
	}
	node.Body = body

	// Expect NODE_END ("===")
	p.skipComments()
	if t := p.peek(); t.Kind != TokNodeEnd {
		return nil, &LexError{Pos: t.Pos, Msg: fmt.Sprintf("expected '===' to end node %q, got %s", node.Title, t.Kind)}
	}
	p.consume() // consume ===

	return node, nil
}

// parseHeader consumes header tokens until a TokSeparator ("---") is seen.
func (p *parser) parseHeader(node *DialogueNode) error {
	for {
		p.skipComments()
		t := p.peek()
		switch t.Kind {
		case TokEOF:
			return &LexError{Pos: t.Pos, Msg: "unexpected EOF in node header (missing '---')"}
		case TokSeparator:
			p.consume()
			return nil
		case TokHeaderKey:
			p.consume() // key
			key := t.Value
			// Next should be TokHeaderValue (required).
			vt := p.peek()
			if vt.Kind != TokHeaderValue {
				return &LexError{Pos: vt.Pos, Msg: fmt.Sprintf("expected header value after key %q, got %s", key, vt.Kind)}
			}
			p.consume() // value
			val := vt.Value

			switch key {
			case "title":
				node.Title = val
				node.Pos = t.Pos
			case "character":
				node.Character = val
			case "language":
				node.Language = val
			case "tags":
				// Consume any TokTag tokens that follow (already emitted by lexer).
				for p.peek().Kind == TokTag {
					node.Tags = append(node.Tags, p.consume().Value)
				}
			case "inherits":
				node.Inherits = splitCSV(val)
			case "suppress":
				node.Suppress = splitCSV(val)
			case "loc_id":
				node.LocID = val
			}
			// Unknown header keys are silently ignored (forward compatibility).
		default:
			return &LexError{Pos: t.Pos, Msg: fmt.Sprintf("unexpected token %s in header section", t.Kind)}
		}
	}
}

// parseBody collects statements at the given indent depth.
// It stops when it sees TokNodeEnd, TokEOF, or an option/line token at a
// shallower depth than minDepth.
func (p *parser) parseBody(minDepth int) ([]Statement, error) {
	var stmts []Statement
	for {
		p.skipComments()
		t := p.peek()
		switch t.Kind {
		case TokEOF, TokNodeEnd:
			return stmts, nil

		case TokHeaderKey:
			// This means we've run into the next node's header without seeing ===.
			return stmts, nil

		case TokSpeaker:
			if t.Depth < minDepth {
				return stmts, nil
			}
			stmt, err := p.parseSpeakerLine()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)

		case TokLine:
			if t.Depth < minDepth {
				return stmts, nil
			}
			stmt, err := p.parseNarrationLine()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)

		case TokOption:
			if t.Depth < minDepth {
				return stmts, nil
			}
			stmt, err := p.parseOption()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, stmt)

		case TokCommand:
			if t.Depth < minDepth {
				return stmts, nil
			}
			// Inline cutscene block: <<cutscene>> or <<cutscene skip:policy>>
			if t.Value == "cutscene" || strings.HasPrefix(t.Value, "cutscene ") {
				p.consume()
				block, err := p.parseInlineCutscene(t)
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, block)
				continue
			}
			p.consume()
			stmts = append(stmts, &CommandStatement{Raw: t.Value, SrcPos: t.Pos})

		default:
			// Unknown token at body level — skip to avoid infinite loop.
			p.consume()
		}
	}
}

// parseSpeakerLine consumes TokSpeaker + TokLine + optional inline tokens.
func (p *parser) parseSpeakerLine() (*SpeakerLine, error) {
	speakerTok := p.consume() // TokSpeaker
	sl := &SpeakerLine{Speaker: speakerTok.Value, SrcPos: speakerTok.Pos}

	// TokLine must follow immediately.
	if p.peek().Kind == TokLine {
		sl.Text = p.consume().Value
	}

	// Collect any inline TokCommand and TokLocKey at the same position (same source line).
	sl.Commands, sl.LocKey = p.collectInlines(speakerTok.Pos.Line)
	return sl, nil
}

// parseNarrationLine consumes TokLine + optional inline tokens.
func (p *parser) parseNarrationLine() (*NarrationLine, error) {
	lineTok := p.consume() // TokLine
	nl := &NarrationLine{Text: lineTok.Value, SrcPos: lineTok.Pos}
	nl.Commands, nl.LocKey = p.collectInlines(lineTok.Pos.Line)
	return nl, nil
}

// parseOption consumes TokOption + its body block.
func (p *parser) parseOption() (*OptionBranch, error) {
	optTok := p.consume() // TokOption
	ob := &OptionBranch{
		Text:   optTok.Value,
		Depth:  optTok.Depth,
		SrcPos: optTok.Pos,
	}
	ob.Commands, ob.LocKey = p.collectInlines(optTok.Pos.Line)

	// Parse the indented body belonging to this option.
	// Body statements must be at depth > ob.Depth.
	body, err := p.parseBody(ob.Depth + 1)
	if err != nil {
		return nil, err
	}
	ob.Body = body
	return ob, nil
}

// collectInlines drains any TokCommand and TokLocKey tokens that originated
// from the same source line number as lineNum.
func (p *parser) collectInlines(lineNum int) (cmds []CommandExpr, locKey string) {
	for {
		t := p.peek()
		if t.Pos.Line != lineNum {
			break
		}
		switch t.Kind {
		case TokCommand:
			p.consume()
			cmds = append(cmds, CommandExpr{Raw: t.Value, SrcPos: t.Pos})
		case TokLocKey:
			p.consume()
			locKey = t.Value
		default:
			return
		}
	}
	return
}

// parseInlineCutscene parses a <<cutscene [skip:policy]>> ... <<end_cutscene>>
// block. openTok is the already-consumed opening TokCommand.
func (p *parser) parseInlineCutscene(openTok Token) (*InlineCutsceneBlock, error) {
	block := &InlineCutsceneBlock{SrcPos: openTok.Pos}

	// Extract skip: param from opening command args (everything after "cutscene ").
	inner := openTok.Value
	if idx := strings.Index(inner, "skip:"); idx >= 0 {
		rest := inner[idx+5:]
		if end := strings.IndexAny(rest, " \t"); end >= 0 {
			block.SkipPolicy = rest[:end]
		} else {
			block.SkipPolicy = rest
		}
	}

	// Collect body commands until <<end_cutscene>>.
	var rawCmds []*cut.RawCommand
	for {
		p.skipComments()
		t := p.peek()
		if t.Kind == TokEOF || t.Kind == TokNodeEnd {
			return nil, &LexError{Pos: openTok.Pos, Msg: "unclosed <<cutscene>> block (missing <<end_cutscene>>)"}
		}
		if t.Kind != TokCommand {
			p.consume() // skip non-command tokens (shouldn't normally appear here)
			continue
		}
		p.consume()
		if t.Value == "end_cutscene" {
			break
		}
		rc, err := cut.ParseInlineCommand(t.Pos.File, t.Pos.Line, t.Value)
		if err != nil {
			return nil, &LexError{Pos: t.Pos, Msg: err.Error()}
		}
		rawCmds = append(rawCmds, rc)
	}

	cmds := cut.ParseSequence(rawCmds)
	block.Sequence = cut.ParseSequenceTree(cmds)
	return block, nil
}

// splitCSV splits a comma-separated string, trimming whitespace.
func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
