// Package item provides types and a parser for .agitem config files.
//
// A .agitem file describes an inventory item — its internal name, display name,
// and description. No scene is generated for items; they are data only.
//
// Example:
//
//	Item "rusty_key" {
//	    display_name = "Rusty Key"
//	    description  = "An old iron key."
//	    sprite       = "assets/items/rusty_key.png"
//	}
package item

import (
	"fmt"
	"strings"
	"unicode"
)


// ItemDialogue holds the parsed content of a `dialogue { ... }` block inside
// a .agitem file.
type ItemDialogue struct {
	OnExamine  string // node title triggered when player examines this item
	OnUseFailed string // node title triggered when item use has no valid target
}

// ItemData is the fully parsed representation of one .agitem file.
type ItemData struct {
	Name        string
	DisplayName string
	Description string
	Sprite      string        // optional: relative path to a sprite image
	Dialogue    *ItemDialogue // nil if no dialogue block present
}

// ValidateDialogueRefs checks that all node titles listed in the dialogue block
// exist in knownTitles. Returns one error per unknown title.
func (id *ItemData) ValidateDialogueRefs(knownTitles map[string]bool) []error {
	if id.Dialogue == nil {
		return nil
	}
	var errs []error
	check := func(field, title string) {
		if title != "" && !knownTitles[title] {
			errs = append(errs, fmt.Errorf("%s: dialogue.%s: node title %q not found", id.Name, field, title))
		}
	}
	check("on_examine", id.Dialogue.OnExamine)
	check("on_use_failed", id.Dialogue.OnUseFailed)
	return errs
}

// ParseItem parses src (the text content of a .agitem file) and returns
// the ItemData. filename is used only in error messages.
func ParseItem(filename, src string) (*ItemData, error) {
	p := &itemParser{src: []rune(src), file: filename}
	return p.parse()
}

// --------------------------------------------------------------------------
// Parser
// --------------------------------------------------------------------------

type itemParser struct {
	src  []rune
	pos  int
	file string
	line int
}

func (p *itemParser) parse() (*ItemData, error) {
	p.line = 1
	p.skipWS()

	// Expect: Item "name" { ... }
	if err := p.expectKeyword("Item"); err != nil {
		return nil, err
	}
	p.skipWS()
	name, err := p.readString()
	if err != nil {
		return nil, fmt.Errorf("%s: expected item name string: %w", p.file, err)
	}
	p.skipWS()
	if err := p.expectChar('{'); err != nil {
		return nil, err
	}

	item := &ItemData{Name: name}

	for {
		p.skipWS()
		if p.peek() == '}' {
			p.pos++
			break
		}
		if p.pos >= len(p.src) {
			return nil, fmt.Errorf("%s: unexpected end of file inside Item block", p.file)
		}
		key, err := p.readIdent()
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", p.file, p.line, err)
		}
		p.skipWS()
		if key == "dialogue" && p.peek() == '{' {
			p.pos++ // consume '{'
			dlg, err := p.parseItemDialogue()
			if err != nil {
				return nil, err
			}
			item.Dialogue = dlg
			continue
		}
		if err := p.expectChar('='); err != nil {
			return nil, err
		}
		p.skipWS()
		val, err := p.readValue()
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", p.file, p.line, err)
		}
		switch key {
		case "display_name":
			item.DisplayName = val
		case "description":
			item.Description = val
		case "sprite":
			item.Sprite = val
		}
	}
	return item, nil
}

func (p *itemParser) parseItemDialogue() (*ItemDialogue, error) {
	dlg := &ItemDialogue{}
	for {
		p.skipWS()
		if p.pos >= len(p.src) {
			return nil, fmt.Errorf("%s:%d: unterminated dialogue block — missing '}'", p.file, p.line)
		}
		if p.peek() == '}' {
			p.pos++
			return dlg, nil
		}
		key, err := p.readIdent()
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", p.file, p.line, err)
		}
		p.skipWS()
		if err := p.expectChar('='); err != nil {
			return nil, err
		}
		p.skipWS()
		val, err := p.readValue()
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", p.file, p.line, err)
		}
		switch key {
		case "on_examine":
			dlg.OnExamine = val
		case "on_use_failed":
			dlg.OnUseFailed = val
		default:
			return nil, fmt.Errorf("%s:%d: unknown dialogue field %q", p.file, p.line, key)
		}
	}
}

// --------------------------------------------------------------------------
// Lexer helpers
// --------------------------------------------------------------------------

func (p *itemParser) peek() rune {
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *itemParser) skipWS() {
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == '\n' {
			p.line++
			p.pos++
		} else if ch == '\r' || ch == '\t' || ch == ' ' {
			p.pos++
		} else if ch == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '/' {
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
		} else {
			break
		}
	}
}

func (p *itemParser) expectKeyword(kw string) error {
	start := p.pos
	for _, r := range []rune(kw) {
		if p.pos >= len(p.src) || p.src[p.pos] != r {
			return fmt.Errorf("%s:%d: expected %q", p.file, p.line, kw)
		}
		p.pos++
	}
	// Must not be followed by an identifier character.
	if p.pos < len(p.src) && isIdentRune(p.src[p.pos]) {
		return fmt.Errorf("%s:%d: expected keyword %q, got %q", p.file, p.line, kw, string(p.src[start:p.pos+1]))
	}
	return nil
}

func (p *itemParser) expectChar(ch rune) error {
	if p.pos >= len(p.src) || p.src[p.pos] != ch {
		return fmt.Errorf("%s:%d: expected %q", p.file, p.line, string(ch))
	}
	p.pos++
	return nil
}

func (p *itemParser) readString() (string, error) {
	if p.pos >= len(p.src) || p.src[p.pos] != '"' {
		return "", fmt.Errorf("expected string literal")
	}
	p.pos++ // consume opening quote
	var b strings.Builder
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == '"' {
			p.pos++
			return b.String(), nil
		}
		if ch == '\n' {
			p.line++
		}
		b.WriteRune(ch)
		p.pos++
	}
	return "", fmt.Errorf("unterminated string literal")
}

func (p *itemParser) readIdent() (string, error) {
	if p.pos >= len(p.src) || !isIdentStart(p.src[p.pos]) {
		return "", fmt.Errorf("expected identifier")
	}
	start := p.pos
	for p.pos < len(p.src) && isIdentRune(p.src[p.pos]) {
		p.pos++
	}
	return string(p.src[start:p.pos]), nil
}

// readValue reads a quoted string and strips the surrounding quotes.
func (p *itemParser) readValue() (string, error) {
	return p.readString()
}

func isIdentStart(r rune) bool { return unicode.IsLetter(r) || r == '_' }
func isIdentRune(r rune) bool  { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }
