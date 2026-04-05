// Package char provides types and a parser for .agchar config files.
//
// A .agchar file describes a character's visual type, mesh or sprite sheet,
// and animation mappings. The AGS-spirit API (WalkTo, Say, etc.) is the same
// regardless of type; only the rendering and animation machinery differs.
//
// Example 3D character:
//
//	Character "player" {
//	    display_name = "Player"
//	    type         = "3d"
//
//	    mesh = "characters/player/player.glb"
//	    animations = {
//	        idle = "Idle"
//	        walk = "Walk"
//	        talk = "Talk"
//	    }
//	}
//
// Example 2D billboard character:
//
//	Character "guard" {
//	    display_name = "Guard"
//	    type         = "2d"
//
//	    sprite_sheet     = "assets/sprites/guard.png"
//	    sprite_angles    = 8
//	    frame_size       = (64, 128)
//	    frames_per_angle = 6
//	}
package char

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// --------------------------------------------------------------------------
// Data types
// --------------------------------------------------------------------------

// CharData is the fully parsed representation of one .agchar file.
type CharData struct {
	Name        string
	DisplayName string
	Type        string // "3d" (default) | "2d"

	// 3D-specific
	Mesh       string            // relative path to .glb; empty = default capsule
	Animations map[string]string // state name → animation clip name in the .glb

	// 2D-specific
	SpriteSheet    string
	SpriteAngles   int    // number of direction variants (1, 4, or 8)
	FrameSize      [2]int // [width, height] in pixels
	FramesPerAngle int    // number of frames per direction
}

// --------------------------------------------------------------------------
// ParseError
// --------------------------------------------------------------------------

// ParseError is a parse error with file and line context.
type ParseError struct {
	File    string
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
}

// --------------------------------------------------------------------------
// Public entry point
// --------------------------------------------------------------------------

// ParseChar parses src as a .agchar file and returns CharData.
// filename is used only in error messages.
func ParseChar(filename, src string) (*CharData, error) {
	p := &agparser{src: []rune(src), file: filename, line: 1}
	return p.parseChar()
}

// --------------------------------------------------------------------------
// Parser
// --------------------------------------------------------------------------

type agparser struct {
	src  []rune
	pos  int
	file string
	line int
}

func (p *agparser) errorf(format string, args ...any) error {
	return &ParseError{File: p.file, Line: p.line, Message: fmt.Sprintf(format, args...)}
}

func (p *agparser) eof() bool { return p.pos >= len(p.src) }

func (p *agparser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *agparser) advance() rune {
	ch := p.src[p.pos]
	p.pos++
	if ch == '\n' {
		p.line++
	}
	return ch
}

// skipWS skips whitespace, // line comments, and # line comments.
func (p *agparser) skipWS() {
	for !p.eof() {
		ch := p.peek()
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			p.advance()
		case ch == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '/':
			for !p.eof() && p.peek() != '\n' {
				p.advance()
			}
		case ch == '#':
			for !p.eof() && p.peek() != '\n' {
				p.advance()
			}
		default:
			return
		}
	}
}

func (p *agparser) ident() (string, error) {
	p.skipWS()
	if p.eof() || (!unicode.IsLetter(p.peek()) && p.peek() != '_') {
		if p.eof() {
			return "", p.errorf("unexpected end of file, expected identifier")
		}
		return "", p.errorf("expected identifier, got %q", p.peek())
	}
	var b strings.Builder
	for !p.eof() && (unicode.IsLetter(p.peek()) || unicode.IsDigit(p.peek()) || p.peek() == '_') {
		b.WriteRune(p.advance())
	}
	return b.String(), nil
}

func (p *agparser) str() (string, error) {
	p.skipWS()
	if p.eof() || p.peek() != '"' {
		if p.eof() {
			return "", p.errorf("unexpected end of file, expected string")
		}
		return "", p.errorf("expected string literal, got %q", p.peek())
	}
	p.advance() // opening "
	var b strings.Builder
	for {
		if p.eof() {
			return "", p.errorf("unterminated string literal")
		}
		ch := p.advance()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			if p.eof() {
				return "", p.errorf("unterminated escape sequence in string")
			}
			switch p.advance() {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			case '"':
				b.WriteRune('"')
			case '\\':
				b.WriteRune('\\')
			default:
				return "", p.errorf("unknown escape sequence")
			}
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String(), nil
}

func (p *agparser) expect(ch rune) error {
	p.skipWS()
	if p.eof() {
		return p.errorf("unexpected end of file, expected %q", ch)
	}
	if p.peek() != ch {
		return p.errorf("expected %q, got %q", ch, p.peek())
	}
	p.advance()
	return nil
}

// integer reads a non-negative decimal integer.
func (p *agparser) integer() (int, error) {
	p.skipWS()
	if p.eof() || !unicode.IsDigit(p.peek()) {
		return 0, p.errorf("expected integer")
	}
	var b strings.Builder
	for !p.eof() && unicode.IsDigit(p.peek()) {
		b.WriteRune(p.advance())
	}
	v, err := strconv.Atoi(b.String())
	if err != nil {
		return 0, p.errorf("invalid integer %q", b.String())
	}
	return v, nil
}

// intTuple parses (n1, n2) and returns two integers.
func (p *agparser) intTuple2() ([2]int, error) {
	if err := p.expect('('); err != nil {
		return [2]int{}, err
	}
	a, err := p.integer()
	if err != nil {
		return [2]int{}, err
	}
	if err := p.expect(','); err != nil {
		return [2]int{}, err
	}
	b, err := p.integer()
	if err != nil {
		return [2]int{}, err
	}
	if err := p.expect(')'); err != nil {
		return [2]int{}, err
	}
	return [2]int{a, b}, nil
}

// --------------------------------------------------------------------------
// Top-level Character block
// --------------------------------------------------------------------------

func (p *agparser) parseChar() (*CharData, error) {
	p.skipWS()
	kw, err := p.ident()
	if err != nil {
		return nil, err
	}
	if kw != "Character" {
		return nil, p.errorf("expected 'Character', got %q", kw)
	}
	name, err := p.str()
	if err != nil {
		return nil, err
	}
	if err := p.expect('{'); err != nil {
		return nil, err
	}

	cd := &CharData{
		Name:       name,
		Type:       "3d", // default
		Animations: make(map[string]string),
	}

	for {
		p.skipWS()
		if p.eof() {
			return nil, p.errorf("unexpected end of file inside Character block — missing '}'")
		}
		if p.peek() == '}' {
			p.advance()
			break
		}

		key, err := p.ident()
		if err != nil {
			return nil, err
		}
		if err := p.expect('='); err != nil {
			return nil, err
		}

		switch key {
		case "display_name":
			v, err := p.str()
			if err != nil {
				return nil, err
			}
			cd.DisplayName = v

		case "type":
			v, err := p.str()
			if err != nil {
				return nil, err
			}
			if v != "3d" && v != "2d" {
				return nil, p.errorf("type must be \"3d\" or \"2d\", got %q", v)
			}
			cd.Type = v

		case "visual_mode":
			v, err := p.str()
			if err != nil {
				return nil, err
			}
			switch v {
			case "mesh":
				cd.Type = "3d"
			case "billboard":
				cd.Type = "2d"
			default:
				return nil, p.errorf("visual_mode must be \"mesh\" or \"billboard\", got %q", v)
			}

		case "mesh":
			v, err := p.str()
			if err != nil {
				return nil, err
			}
			cd.Mesh = v

		case "animations":
			if err := p.expect('{'); err != nil {
				return nil, err
			}
			if err := p.parseAnimations(cd); err != nil {
				return nil, err
			}

		case "sprite_sheet":
			v, err := p.str()
			if err != nil {
				return nil, err
			}
			cd.SpriteSheet = v

		case "sprite_angles":
			v, err := p.integer()
			if err != nil {
				return nil, err
			}
			if v != 1 && v != 4 && v != 8 {
				return nil, p.errorf("sprite_angles must be 1, 4, or 8, got %d", v)
			}
			cd.SpriteAngles = v

		case "frame_size":
			fs, err := p.intTuple2()
			if err != nil {
				return nil, err
			}
			cd.FrameSize = fs

		case "frames_per_angle":
			v, err := p.integer()
			if err != nil {
				return nil, err
			}
			cd.FramesPerAngle = v

		default:
			return nil, p.errorf("unknown Character property %q", key)
		}
	}

	p.skipWS()
	if !p.eof() {
		return nil, p.errorf("unexpected content after Character block")
	}
	return cd, nil
}

// parseAnimations reads key = "value" pairs until '}'.
func (p *agparser) parseAnimations(cd *CharData) error {
	for {
		p.skipWS()
		if p.eof() {
			return p.errorf("unterminated animations block — missing '}'")
		}
		if p.peek() == '}' {
			p.advance()
			return nil
		}
		key, err := p.ident()
		if err != nil {
			return err
		}
		if err := p.expect('='); err != nil {
			return err
		}
		v, err := p.str()
		if err != nil {
			return err
		}
		cd.Animations[key] = v
	}
}
