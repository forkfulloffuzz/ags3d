// Package gui provides types and a parser for .agui layout files.
//
// A .agui file describes an in-game HUD: a CanvasLayer that contains
// InventoryBar, VerbBar, and StatusLine widgets. ag build generates a
// CanvasLayer .tscn from each .agui file.
//
// Example:
//
//	GUI "main_hud" {
//	    layer = 10
//
//	    InventoryBar "inv_bar" {
//	        position   = (0, 0, bottom)
//	        item_size  = (48, 48)
//	        columns    = 8
//	    }
//
//	    VerbBar "verbs" {
//	        position = (0, 0, bottom_right)
//	        verbs    = ["Look", "Use", "Pick up", "Talk to"]
//	    }
//
//	    StatusLine "status" {
//	        position = (0, 0, top)
//	        font     = "assets/fonts/main.ttf"
//	    }
//	}
package gui

import (
	"fmt"
	"strings"
	"unicode"
)

// --------------------------------------------------------------------------
// Data types
// --------------------------------------------------------------------------

// GUIData is the fully parsed representation of one .agui file.
type GUIData struct {
	Name    string
	Layer   int      // CanvasLayer.layer value
	Widgets []Widget // ordered list of widgets
}

// Widget is one UI element within the GUI.
type Widget struct {
	Type   string // "InventoryBar" | "VerbBar" | "StatusLine"
	Name   string // widget's internal name
	Anchor string // anchor identifier: "top", "bottom", "bottom_right", etc.
	// Pixel offsets relative to the anchor position (from the position tuple).
	OffsetX int
	OffsetY int

	// InventoryBar
	ItemSizeW int // item slot width in pixels
	ItemSizeH int // item slot height in pixels
	Columns   int

	// VerbBar
	Verbs []string

	// StatusLine
	Font string // optional: path to a font asset
}

// --------------------------------------------------------------------------
// ParseError
// --------------------------------------------------------------------------

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

func ParseGUI(filename, src string) (*GUIData, error) {
	p := &guiParser{src: []rune(src), file: filename, line: 1}
	return p.parseGUI()
}

// --------------------------------------------------------------------------
// Parser
// --------------------------------------------------------------------------

type guiParser struct {
	src  []rune
	pos  int
	file string
	line int
}

func (p *guiParser) errorf(format string, args ...any) error {
	return &ParseError{File: p.file, Line: p.line, Message: fmt.Sprintf(format, args...)}
}

func (p *guiParser) eof() bool { return p.pos >= len(p.src) }

func (p *guiParser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *guiParser) advance() rune {
	ch := p.src[p.pos]
	p.pos++
	if ch == '\n' {
		p.line++
	}
	return ch
}

func (p *guiParser) skipWS() {
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

func (p *guiParser) ident() (string, error) {
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

func (p *guiParser) str() (string, error) {
	p.skipWS()
	if p.eof() || p.peek() != '"' {
		if p.eof() {
			return "", p.errorf("unexpected end of file, expected string literal")
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
				return "", p.errorf("unterminated escape in string")
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

func (p *guiParser) expect(ch rune) error {
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

func (p *guiParser) integer() (int, error) {
	p.skipWS()
	negative := false
	if !p.eof() && p.peek() == '-' {
		negative = true
		p.advance()
	}
	if p.eof() || !unicode.IsDigit(p.peek()) {
		return 0, p.errorf("expected integer")
	}
	var b strings.Builder
	for !p.eof() && unicode.IsDigit(p.peek()) {
		b.WriteRune(p.advance())
	}
	v := 0
	for _, ch := range b.String() {
		v = v*10 + int(ch-'0')
	}
	if negative {
		v = -v
	}
	return v, nil
}

// --------------------------------------------------------------------------
// Top-level GUI block
// --------------------------------------------------------------------------

func (p *guiParser) parseGUI() (*GUIData, error) {
	p.skipWS()
	kw, err := p.ident()
	if err != nil {
		return nil, err
	}
	if kw != "GUI" {
		return nil, p.errorf("expected 'GUI', got %q", kw)
	}
	name, err := p.str()
	if err != nil {
		return nil, err
	}
	if err := p.expect('{'); err != nil {
		return nil, err
	}

	g := &GUIData{Name: name, Layer: 1}

	for {
		p.skipWS()
		if p.eof() {
			return nil, p.errorf("unexpected end of file inside GUI block — missing '}'")
		}
		if p.peek() == '}' {
			p.advance()
			break
		}

		key, err := p.ident()
		if err != nil {
			return nil, err
		}

		switch key {
		case "layer":
			if err := p.expect('='); err != nil {
				return nil, err
			}
			v, err := p.integer()
			if err != nil {
				return nil, err
			}
			g.Layer = v

		case "InventoryBar", "VerbBar", "StatusLine":
			w, err := p.parseWidget(key)
			if err != nil {
				return nil, err
			}
			g.Widgets = append(g.Widgets, *w)

		default:
			return nil, p.errorf("unknown GUI property or widget type %q", key)
		}
	}

	p.skipWS()
	if !p.eof() {
		return nil, p.errorf("unexpected content after GUI block")
	}
	return g, nil
}

// --------------------------------------------------------------------------
// Widget block
// --------------------------------------------------------------------------

func (p *guiParser) parseWidget(widgetType string) (*Widget, error) {
	name, err := p.str()
	if err != nil {
		return nil, err
	}
	if err := p.expect('{'); err != nil {
		return nil, err
	}

	w := &Widget{
		Type:      widgetType,
		Name:      name,
		Anchor:    "top_left",
		ItemSizeW: 48,
		ItemSizeH: 48,
		Columns:   8,
	}

	for {
		p.skipWS()
		if p.eof() {
			return nil, p.errorf("unexpected end of file inside %s block", widgetType)
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
		case "position":
			if err := p.expect('('); err != nil {
				return nil, err
			}
			x, err := p.integer()
			if err != nil {
				return nil, err
			}
			if err := p.expect(','); err != nil {
				return nil, err
			}
			y, err := p.integer()
			if err != nil {
				return nil, err
			}
			if err := p.expect(','); err != nil {
				return nil, err
			}
			anchor, err := p.ident()
			if err != nil {
				return nil, err
			}
			if err := p.expect(')'); err != nil {
				return nil, err
			}
			w.OffsetX = x
			w.OffsetY = y
			w.Anchor = anchor

		case "item_size":
			if err := p.expect('('); err != nil {
				return nil, err
			}
			sw, err := p.integer()
			if err != nil {
				return nil, err
			}
			if err := p.expect(','); err != nil {
				return nil, err
			}
			sh, err := p.integer()
			if err != nil {
				return nil, err
			}
			if err := p.expect(')'); err != nil {
				return nil, err
			}
			w.ItemSizeW = sw
			w.ItemSizeH = sh

		case "columns":
			v, err := p.integer()
			if err != nil {
				return nil, err
			}
			w.Columns = v

		case "verbs":
			verbs, err := p.parseStringList()
			if err != nil {
				return nil, err
			}
			w.Verbs = verbs

		case "font":
			v, err := p.str()
			if err != nil {
				return nil, err
			}
			w.Font = v

		default:
			return nil, p.errorf("unknown %s property %q", widgetType, key)
		}
	}

	return w, nil
}

// parseStringList reads: [ "a", "b", "c" ]
func (p *guiParser) parseStringList() ([]string, error) {
	if err := p.expect('['); err != nil {
		return nil, err
	}
	var out []string
	for {
		p.skipWS()
		if p.eof() {
			return nil, p.errorf("unterminated list literal")
		}
		if p.peek() == ']' {
			p.advance()
			break
		}
		s, err := p.str()
		if err != nil {
			return nil, err
		}
		out = append(out, s)
		p.skipWS()
		if !p.eof() && p.peek() == ',' {
			p.advance()
		}
	}
	return out, nil
}
