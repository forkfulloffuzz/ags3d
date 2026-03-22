package parser

import (
	"fmt"

	"github.com/ags3d/ag/internal/scanner"
)

// -------------------------------------------------------------------
// Parse errors
// -------------------------------------------------------------------

// ParseError is a single error with source location.
// All errors are collected rather than aborting — the caller receives
// a partial AST plus the full error list.
type ParseError struct {
	File    string
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
}

// -------------------------------------------------------------------
// Parser
// -------------------------------------------------------------------

// Parser is a single-pass recursive descent parser for AGS-spirit source.
//
// Construct with New(), then call Parse() exactly once.
// Errors are collected in p.errors — Parse() never panics on bad input.
// The returned *File may be partial when errors are present.
//
// Full implementation in T09. This file contains the struct and the
// stub entry point only.
type Parser struct {
	file   string
	s      *scanner.Scanner
	cur    scanner.Token // current lookahead token
	errors []*ParseError
}

// New creates a Parser over an already-constructed Scanner.
// The scanner is advanced once so p.cur holds the first real token.
func New(s *scanner.Scanner) *Parser {
	p := &Parser{s: s}
	p.advance()
	return p
}

// Parse parses a complete .agscript file and returns the AST root.
// Errors are collected and returned alongside the (possibly partial) File.
// Never panics — unrecognised input is recorded as a ParseError.
//
// TODO(T09): implement the full recursive descent parser.
func (p *Parser) Parse(file string) (*File, []*ParseError) {
	p.file = file
	f := &File{Path: file}
	// TODO(T09): parse top-level declarations
	return f, p.errors
}

// -------------------------------------------------------------------
// Token management (used by T09 parser methods)
// -------------------------------------------------------------------

// advance consumes the current token and loads the next one.
func (p *Parser) advance() {
	p.cur = p.s.Next()
}

// peek returns the current lookahead token without consuming it.
func (p *Parser) peek() scanner.Token {
	return p.cur
}

// check returns true if the current token has the given kind.
func (p *Parser) check(k scanner.TokenKind) bool {
	return p.cur.Kind == k
}

// eat consumes the current token and returns it if it matches kind k.
// If it does not match, it records an error and returns the current token
// without consuming it (panic-mode recovery keeps the token stream intact).
func (p *Parser) eat(k scanner.TokenKind, context string) scanner.Token {
	if p.cur.Kind == k {
		tok := p.cur
		p.advance()
		return tok
	}
	p.errorf("expected %v but found %q (%s)", k, p.cur.Lexeme, context)
	return p.cur
}

// match consumes and returns the current token if it matches any of the
// given kinds. Returns (tok, true) on match, (zero, false) otherwise.
func (p *Parser) match(kinds ...scanner.TokenKind) (scanner.Token, bool) {
	for _, k := range kinds {
		if p.cur.Kind == k {
			tok := p.cur
			p.advance()
			return tok, true
		}
	}
	return scanner.Token{}, false
}

// errorf records a ParseError at the current token position.
func (p *Parser) errorf(format string, args ...any) {
	p.errors = append(p.errors, &ParseError{
		File:    p.file,
		Line:    p.cur.Line,
		Column:  p.cur.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

// errorAt records a ParseError at a specific token's position.
func (p *Parser) errorAt(tok scanner.Token, format string, args ...any) {
	p.errors = append(p.errors, &ParseError{
		File:    p.file,
		Line:    tok.Line,
		Column:  tok.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

// sync advances past tokens until it finds one that can start a new
// top-level declaration or a statement boundary — used for panic-mode
// error recovery so parsing can continue after a bad token.
func (p *Parser) sync() {
	for {
		switch p.cur.Kind {
		case scanner.TokenEOF,
			scanner.TokenFunction,
			scanner.TokenNamespace,
			scanner.TokenEnum,
			scanner.TokenRBrace:
			return
		}
		p.advance()
	}
}
