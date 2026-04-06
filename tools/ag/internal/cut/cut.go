// Package cut implements the parser for .agcut cutscene files.
//
// A .agcut file has a header section (key: value pairs) terminated by a
// "sequence:" line, followed by a body of <<command>> steps.
//
// Example:
//
//	title: chapter1_opening
//	skip: after_first_view
//	save_block: true
//	tags: [chapter:1, cinematic]
//	fallback: halt
//	sequence:
//	<<fade_in duration:2.0>>
//	<<camera set point.rooftop_wide fov:60>>
//	<<line narrator "Three years. And it still felt foreign.">>
//	<<end>>
package cut

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ---- Token types -------------------------------------------------------

// TokenKind identifies the type of a lexed token from a .agcut file.
type TokenKind int

const (
	TokEOF           TokenKind = iota
	TokComment                 // // comment text
	TokHeaderKey               // key in "key: value" header pair
	TokHeaderValue             // value in "key: value" header pair
	TokTag                     // tag entry from [tag1, tag2] syntax
	TokSequenceStart           // "sequence:" line — body begin marker

	// Command-body token kinds.
	TokCommandName  // first word after << (e.g. "fade_in", "camera", "character")
	TokNamedParam   // word:value pair inside << >> (e.g. "duration:2.0", "bg:cam_move")
	TokStringValue  // "quoted string" inside << >>
	TokIdentifier   // unquoted word or dotted path inside << >> (e.g. "narrator", "point.street_level")
	TokBlockOpen    // <<parallel>>, <<if condition>>, <<on event:…>> (opening block command)
	TokBlockClose   // <<end_parallel>>, <<end_if>>, <<end_on>>    (closing block command)
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
	case TokTag:
		return "TAG"
	case TokSequenceStart:
		return "SEQUENCE_START"
	case TokCommandName:
		return "COMMAND_NAME"
	case TokNamedParam:
		return "NAMED_PARAM"
	case TokStringValue:
		return "STRING_VALUE"
	case TokIdentifier:
		return "IDENTIFIER"
	case TokBlockOpen:
		return "BLOCK_OPEN"
	case TokBlockClose:
		return "BLOCK_CLOSE"
	default:
		return fmt.Sprintf("TokenKind(%d)", int(k))
	}
}

// ---- Position -------------------------------------------------------

// Pos identifies a source location within a .agcut file.
type Pos struct {
	File string
	Line int // 1-based
	Col  int // 1-based
}

func (p Pos) String() string {
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col)
}

// ---- Error -------------------------------------------------------

// ParseError is returned when the parser encounters malformed input.
type ParseError struct {
	Pos Pos
	Msg string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
}

// ---- AST -------------------------------------------------------

// CutsceneFile is the parsed result of a single .agcut source file.
type CutsceneFile struct {
	Path         string
	Title        string
	Skip         string   // "always" | "never" | "after_first_view" | "author_controlled"
	SaveBlock    bool     // default true
	Tags         []string
	Fallback     string   // "halt" | "skip_and_continue" | "log_and_continue" | "retry_once"
	LocGroup     string
	VoiceSession string

	// Audio scope and dialogue ducking fields.
	AudioScope   string  // "keep" | "pause" | "stop"; default "keep"
	DuckChannels string  // space-separated channel names to duck during dialogue
	DuckLevel    float64 // target volume level while ducked; default 0.25
	DuckFade     float64 // seconds to ramp down to DuckLevel; default 0.3
	DuckRestore  float64 // seconds to ramp back up after dialogue; default 0.5
	AutoDuck     bool    // automatically duck during all <<line>>/<<dialogue>> commands

	// Sequence is the flat ordered list of commands in the sequence body.
	// Block structure (parallel, if, on) is resolved in T-CUT02.
	Sequence []*RawCommand
}

// RawCommand is a single <<command args>> step from the sequence body.
// The command name is parsed; the argument string is left raw for T-CUT02.
type RawCommand struct {
	Name        string // command verb, e.g. "fade_in", "camera", "line"
	Args        string // raw argument content (everything after Name inside << >>)
	IsBlockOpen bool   // true for <<parallel>>, <<if>>, <<on>>
	IsBlockClose bool  // true for <<end_parallel>>, <<end_if>>, <<end_on>>
	Pos         Pos
}

// ---- Valid header values -------------------------------------------------------

// ValidSkipPolicies is the set of recognised skip: values.
var ValidSkipPolicies = map[string]bool{
	"always":           true,
	"never":            true,
	"after_first_view": true,
	"author_controlled": true,
}

// ValidAudioScopes is the set of recognised audio_scope: values.
var ValidAudioScopes = map[string]bool{
	"keep":  true,
	"pause": true,
	"stop":  true,
}

// ValidFallbacks is the set of recognised fallback: values.
var ValidFallbacks = map[string]bool{
	"halt":              true,
	"skip_and_continue": true,
	"log_and_continue":  true,
	"retry_once":        true,
}

// blockOpenNames are <<command>> names that open a nested block.
var blockOpenNames = map[string]bool{
	"parallel": true,
	"if":       true,
	"on":       true,
	"cutscene": true, // inline <<cutscene>> block
}

// blockCloseNames are <<command>> names that close a nested block.
var blockCloseNames = map[string]bool{
	"end_parallel": true,
	"end_if":       true,
	"end_on":       true,
	"end_cutscene": true,
}

// IsBlockOpenName returns true if name is a block-opening command.
func IsBlockOpenName(name string) bool { return blockOpenNames[name] }

// IsBlockCloseName returns true if name is a block-closing command.
func IsBlockCloseName(name string) bool { return blockCloseNames[name] }

// ---- Parser -------------------------------------------------------

// ParseFile reads and parses a .agcut file from disk.
func ParseFile(path string) (*CutsceneFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cut: cannot read %s: %w", path, err)
	}
	return Parse(path, string(data))
}

// Parse parses a .agcut source string.
// filename is used for position reporting only.
func Parse(filename, src string) (*CutsceneFile, error) {
	cf := &CutsceneFile{
		Path:        filename,
		SaveBlock:   true,   // default
		AudioScope:  "keep", // default
		DuckLevel:   0.25,   // default
		DuckFade:    0.3,    // default
		DuckRestore: 0.5,    // default
	}

	lines := strings.Split(src, "\n")
	inSequence := false

	for lineIdx, raw := range lines {
		lineNum := lineIdx + 1
		pos := Pos{File: filename, Line: lineNum, Col: 1}
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		// Skip blank lines everywhere.
		if trimmed == "" {
			continue
		}

		// Comments are valid anywhere.
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		if !inSequence {
			// --- Header section ---
			// "sequence:" with no value → body start
			if trimmed == "sequence:" {
				inSequence = true
				continue
			}
			if err := parseHeaderLine(cf, pos, trimmed); err != nil {
				return nil, err
			}
		} else {
			// --- Sequence body ---
			cmd, err := parseCommandLine(pos, trimmed)
			if err != nil {
				return nil, err
			}
			if cmd != nil {
				cf.Sequence = append(cf.Sequence, cmd)
			}
		}
	}

	return cf, nil
}

// parseHeaderLine parses one header key: value line into cf.
func parseHeaderLine(cf *CutsceneFile, pos Pos, line string) error {
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return &ParseError{Pos: pos, Msg: fmt.Sprintf("malformed header line (expected 'key: value'): %q", line)}
	}
	key := strings.TrimSpace(line[:idx])
	rawVal := strings.TrimSpace(line[idx+1:])

	switch key {
	case "title":
		cf.Title = rawVal
	case "skip":
		cf.Skip = rawVal
	case "save_block":
		cf.SaveBlock = rawVal != "false"
	case "tags":
		// Parse [tag1, tag2] list.
		if strings.HasPrefix(rawVal, "[") && strings.HasSuffix(rawVal, "]") {
			inner := rawVal[1 : len(rawVal)-1]
			for _, part := range strings.Split(inner, ",") {
				if tag := strings.TrimSpace(part); tag != "" {
					cf.Tags = append(cf.Tags, tag)
				}
			}
		}
	case "fallback":
		cf.Fallback = rawVal
	case "loc_group":
		cf.LocGroup = rawVal
	case "voice_session":
		cf.VoiceSession = rawVal
	case "audio_scope":
		cf.AudioScope = rawVal
	case "duck_channels":
		cf.DuckChannels = rawVal
	case "duck_level":
		if v, err := strconv.ParseFloat(rawVal, 64); err == nil {
			cf.DuckLevel = v
		}
	case "duck_fade":
		if v, err := strconv.ParseFloat(rawVal, 64); err == nil {
			cf.DuckFade = v
		}
	case "duck_restore":
		if v, err := strconv.ParseFloat(rawVal, 64); err == nil {
			cf.DuckRestore = v
		}
	case "auto_duck":
		cf.AutoDuck = rawVal == "true"
	// Unknown header keys silently ignored (forward compatibility).
	}
	return nil
}

// parseCommandLine parses one <<command args>> body line into a RawCommand.
// Returns nil if the line is not a command (e.g. blank or comment — handled above).
func parseCommandLine(pos Pos, line string) (*RawCommand, error) {
	if !strings.HasPrefix(line, "<<") || !strings.HasSuffix(line, ">>") {
		return nil, &ParseError{Pos: pos, Msg: fmt.Sprintf("sequence body line is not a <<command>>: %q", line)}
	}
	inner := strings.TrimSpace(line[2 : len(line)-2])
	if inner == "" {
		return nil, &ParseError{Pos: pos, Msg: "empty <<>> command"}
	}

	// Split off the command name (first word).
	name, args := splitCommandName(inner)

	cmd := &RawCommand{
		Name:         name,
		Args:         args,
		IsBlockOpen:  blockOpenNames[name],
		IsBlockClose: blockCloseNames[name],
		Pos:          pos,
	}
	return cmd, nil
}

// splitCommandName splits "name rest args" into (name, "rest args").
func splitCommandName(s string) (name, args string) {
	idx := strings.IndexAny(s, " \t")
	if idx < 0 {
		return s, ""
	}
	return s[:idx], strings.TrimSpace(s[idx+1:])
}
