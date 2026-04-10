// Package dlg implements the lexer and parser for .agdlg dialogue files.
//
// A .agdlg file contains one or more dialogue nodes. Each node has a header
// section (key: value pairs) terminated by "---", a body section containing
// speaker lines, narration, options, and commands, and a "===" end marker.
//
// Example:
//
//	title: guard_greeting
//	character: guard
//	tags: [chapter:1]
//	---
//	Guard: You there. Stop.
//	-> I'm just passing through. <<visible_if not flag.guard_suspicious>>
//	   Guard: Move along then.
//	   <<action flag.guard_spoken = true>>
//	   <<end>>
//	===
package dlg

import (
	"fmt"
	"strings"
)

// TokenKind identifies the type of a lexed token.
type TokenKind int

const (
	TokEOF         TokenKind = iota
	TokComment               // content after "// " on a comment line
	TokHeaderKey             // key in "key: value" header pair (e.g. "title")
	TokHeaderValue           // value in "key: value" header pair (e.g. "guard_greeting")
	TokSeparator             // "---" — separates header from body
	TokNodeEnd               // "===" — ends a dialogue node
	TokSpeaker               // speaker name in "Speaker: dialogue text" body line
	TokLine                  // dialogue text: body text after a speaker colon, narration, or option text
	TokOption                // "->" option marker; Value holds option display text; Depth = indent level
	TokCommand               // content inside "<<...>>" (angle brackets stripped)
	TokTag                   // content inside "[...]" on a header value line (e.g. "chapter:1")
	TokLocKey                // loc key after "#loc:" annotation (e.g. "guard_hi_001")
	TokCtx                   // context note after "#ctx:" annotation (e.g. "guard is suspicious")
)

func (k TokenKind) String() string {
	switch k {
	case TokEOF:
		return "EOF"
	case TokComment:
		return "COMMENT"
	case TokHeaderKey:
		return "HEADER_KEY"
	case TokHeaderValue:
		return "HEADER_VALUE"
	case TokSeparator:
		return "SEPARATOR"
	case TokNodeEnd:
		return "NODE_END"
	case TokSpeaker:
		return "SPEAKER"
	case TokLine:
		return "LINE"
	case TokOption:
		return "OPTION"
	case TokCommand:
		return "COMMAND"
	case TokTag:
		return "TAG"
	case TokLocKey:
		return "LOC_KEY"
	case TokCtx:
		return "CTX"
	default:
		return fmt.Sprintf("TokenKind(%d)", int(k))
	}
}

// Pos identifies a source location within a .agdlg file.
type Pos struct {
	File string
	Line int // 1-based
	Col  int // 1-based, byte offset within line
}

func (p Pos) String() string {
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col)
}

// Token is a single lexed unit from a .agdlg source file.
type Token struct {
	Kind  TokenKind
	Value string // semantic text content (trimmed; delimiters excluded)
	Pos   Pos

	// Depth is the indent level (number of leading tab stops, where each
	// two-space group counts as one level). Only meaningful for TokOption
	// and TokLine tokens in the body section.
	Depth int
}

// LexError is returned when the lexer encounters malformed input.
type LexError struct {
	Pos Pos
	Msg string
}

func (e *LexError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
}

// lexState tracks whether the lexer is in a header or body section.
type lexState int

const (
	stateHeader lexState = iota
	stateBody
)

// Lex tokenises the full text of a .agdlg source file.
// filename is used only for error and position reporting.
// Returns all tokens followed by a single TokEOF, or a non-nil error.
func Lex(filename, src string) ([]Token, error) {
	lines := strings.Split(src, "\n")
	var tokens []Token
	state := stateHeader

	for lineIdx, raw := range lines {
		lineNum := lineIdx + 1
		pos := Pos{File: filename, Line: lineNum, Col: 1}

		// Strip trailing carriage return for Windows line endings.
		line := strings.TrimRight(raw, "\r")

		// Skip blank lines everywhere.
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Comments are valid in any section.
		if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "//") {
			content := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
			tokens = append(tokens, Token{Kind: TokComment, Value: content, Pos: pos})
			continue
		}

		trimmed := strings.TrimSpace(line)

		switch state {
		case stateHeader:
			toks, err := lexHeaderLine(filename, lineNum, trimmed)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, toks...)
			// Transition to body after separator.
			for _, t := range toks {
				if t.Kind == TokSeparator {
					state = stateBody
				}
			}

		case stateBody:
			if trimmed == "===" {
				tokens = append(tokens, Token{Kind: TokNodeEnd, Value: "===", Pos: pos})
				state = stateHeader
				continue
			}
			depth := indentDepth(line)
			toks, err := lexBodyLine(filename, lineNum, trimmed, depth)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, toks...)
		}
	}

	tokens = append(tokens, Token{Kind: TokEOF, Pos: Pos{File: filename, Line: len(lines) + 1, Col: 1}})
	return tokens, nil
}

// lexHeaderLine tokenises one line from the header section.
// Recognised patterns:
//
//	"---"            → TokSeparator
//	"key: value"     → TokHeaderKey + TokHeaderValue (+ TokTag items for [tag:val] syntax)
func lexHeaderLine(file string, lineNum int, line string) ([]Token, error) {
	pos := Pos{File: file, Line: lineNum, Col: 1}

	if line == "---" {
		return []Token{{Kind: TokSeparator, Value: "---", Pos: pos}}, nil
	}

	idx := strings.Index(line, ":")
	if idx <= 0 {
		return nil, &LexError{Pos: pos, Msg: fmt.Sprintf("malformed header line (expected 'key: value'): %q", line)}
	}

	key := strings.TrimSpace(line[:idx])
	rawVal := strings.TrimSpace(line[idx+1:])

	tokens := []Token{
		{Kind: TokHeaderKey, Value: key, Pos: pos},
	}

	// Parse tag list syntax: [tag1, tag2:val] — emit TokTag per entry plus TokHeaderValue.
	if strings.HasPrefix(rawVal, "[") && strings.HasSuffix(rawVal, "]") {
		inner := rawVal[1 : len(rawVal)-1]
		tokens = append(tokens, Token{Kind: TokHeaderValue, Value: rawVal, Pos: pos})
		for _, part := range strings.Split(inner, ",") {
			tag := strings.TrimSpace(part)
			if tag != "" {
				tokens = append(tokens, Token{Kind: TokTag, Value: tag, Pos: pos})
			}
		}
	} else {
		tokens = append(tokens, Token{Kind: TokHeaderValue, Value: rawVal, Pos: pos})
	}

	return tokens, nil
}

// lexBodyLine tokenises one line from the body section.
// Recognised patterns (all may carry inline <<commands>> and #loc: annotations):
//
//	"-> option text"         → TokOption (+ inline TokCommand, TokLocKey)
//	"Speaker: line text"     → TokSpeaker + TokLine (+ inline TokCommand, TokLocKey)
//	"<<command content>>"    → TokCommand (standalone command line)
//	"narration text"         → TokLine (+ inline TokCommand, TokLocKey)
func lexBodyLine(file string, lineNum int, trimmed string, depth int) ([]Token, error) {
	pos := Pos{File: file, Line: lineNum, Col: 1}

	// Option line: starts with "->"
	if strings.HasPrefix(trimmed, "->") {
		optText := strings.TrimSpace(trimmed[2:])
		text, inlines, err := extractInlines(file, lineNum, optText)
		if err != nil {
			return nil, err
		}
		toks := []Token{{Kind: TokOption, Value: text, Pos: pos, Depth: depth}}
		toks = append(toks, inlines...)
		return toks, nil
	}

	// Standalone command line: entire line is <<...>>
	if strings.HasPrefix(trimmed, "<<") {
		cmd, rest, err := extractCommand(file, lineNum, trimmed)
		if err != nil {
			return nil, err
		}
		_ = rest // standalone command lines have nothing after the closing >>
		return []Token{{Kind: TokCommand, Value: cmd, Pos: pos, Depth: depth}}, nil
	}

	// Speaker line: "Word(s): rest" — speaker name must not contain spaces
	// and must be followed by ": " (colon + space or colon at EOL).
	if speakerName, rest, ok := parseSpeakerPrefix(trimmed); ok {
		text, inlines, err := extractInlines(file, lineNum, rest)
		if err != nil {
			return nil, err
		}
		toks := []Token{
			{Kind: TokSpeaker, Value: speakerName, Pos: pos, Depth: depth},
			{Kind: TokLine, Value: text, Pos: pos, Depth: depth},
		}
		toks = append(toks, inlines...)
		return toks, nil
	}

	// Plain narration / continuation line.
	text, inlines, err := extractInlines(file, lineNum, trimmed)
	if err != nil {
		return nil, err
	}
	toks := []Token{{Kind: TokLine, Value: text, Pos: pos, Depth: depth}}
	toks = append(toks, inlines...)
	return toks, nil
}

// parseSpeakerPrefix checks whether line begins with "SpeakerName: ".
// The speaker name must be a single word (no spaces) to avoid false positives
// on narration lines that happen to contain a colon (e.g. "Time: 3:00 PM").
func parseSpeakerPrefix(line string) (speaker, rest string, ok bool) {
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return "", "", false
	}
	candidate := line[:idx]
	// Reject if the candidate contains spaces — it's not a speaker name.
	if strings.ContainsAny(candidate, " \t") {
		return "", "", false
	}
	return candidate, strings.TrimSpace(line[idx+1:]), true
}

// extractInlines extracts #loc:, #ctx:, and <<commands>> from text.
// Returns the cleaned text and inline tokens.
func extractInlines(file string, lineNum int, text string) (string, []Token, error) {
	pos := Pos{File: file, Line: lineNum}
	var inlines []Token

	// Extract #ctx: annotation last (extends to end of line).
	if idx := strings.LastIndex(text, "#ctx:"); idx >= 0 {
		ctx := strings.TrimSpace(text[idx+5:])
		if ctx != "" {
			inlines = append(inlines, Token{Kind: TokCtx, Value: ctx, Pos: pos})
		}
		text = strings.TrimSpace(text[:idx])
	}

	// Extract #loc: annotation (before #ctx: if both present).
	if idx := strings.LastIndex(text, "#loc:"); idx >= 0 {
		locKey := strings.TrimSpace(text[idx+5:])
		// Strip any trailing <<...>> from the loc key (shouldn't happen, but be safe).
		if end := strings.Index(locKey, " "); end >= 0 {
			locKey = locKey[:end]
		}
		if locKey != "" {
			inlines = append(inlines, Token{Kind: TokLocKey, Value: locKey, Pos: pos})
		}
		text = strings.TrimSpace(text[:idx])
	}

	// Extract <<commands>> — may appear anywhere in text.
	var cleaned strings.Builder
	remaining := text
	for {
		open := strings.Index(remaining, "<<")
		if open < 0 {
			cleaned.WriteString(remaining)
			break
		}
		cleaned.WriteString(remaining[:open])
		afterOpen := remaining[open+2:]
		close := strings.Index(afterOpen, ">>")
		if close < 0 {
			return "", nil, &LexError{
				Pos: Pos{File: file, Line: lineNum},
				Msg: "unclosed '<<' command",
			}
		}
		cmdContent := strings.TrimSpace(afterOpen[:close])
		inlines = append(inlines, Token{Kind: TokCommand, Value: cmdContent, Pos: pos})
		remaining = afterOpen[close+2:]
	}

	return strings.TrimSpace(cleaned.String()), inlines, nil
}

// extractCommand parses a standalone "<<content>>" line, returning
// the command content and any remaining text after the closing ">>".
func extractCommand(file string, lineNum int, line string) (cmd, rest string, err error) {
	if !strings.HasPrefix(line, "<<") {
		return "", line, nil
	}
	afterOpen := line[2:]
	idx := strings.Index(afterOpen, ">>")
	if idx < 0 {
		return "", "", &LexError{
			Pos: Pos{File: file, Line: lineNum},
			Msg: "unclosed '<<' command",
		}
	}
	return strings.TrimSpace(afterOpen[:idx]), strings.TrimSpace(afterOpen[idx+2:]), nil
}

// indentDepth converts the leading whitespace of a raw (un-trimmed) line into
// a nesting depth. Two spaces or one tab each count as one indent level.
func indentDepth(raw string) int {
	spaces := 0
	for _, ch := range raw {
		if ch == '\t' {
			spaces += 2 // treat tab as 2-space indent
		} else if ch == ' ' {
			spaces++
		} else {
			break
		}
	}
	return spaces / 2
}
