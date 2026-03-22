package scanner_test

// TEST-M2-01 — Lexer unit tests with explicit UT IDs.
//
// These tests satisfy the acceptance criteria for GitHub issue
// "TEST-M2-01 — Write lexer tests and fixtures (4 tests)".

import (
	"testing"

	"github.com/ags3d/ag/internal/scanner"
)

// UT-M2-01: Keywords tokenised correctly (function, if, while, return).
func TestUT_M2_01_KeywordsTokenisedCorrectly(t *testing.T) {
	cases := []struct {
		src  string
		kind scanner.TokenKind
	}{
		{"function", scanner.TokenFunction},
		{"if", scanner.TokenIf},
		{"while", scanner.TokenWhile},
		{"return", scanner.TokenReturn},
	}
	for _, tc := range cases {
		s := scanner.New("test", tc.src)
		tok := s.Next()
		if tok.Kind != tc.kind {
			t.Errorf("UT-M2-01: %q → kind %v, want %v", tc.src, tok.Kind, tc.kind)
		}
		if tok.Lexeme != tc.src {
			t.Errorf("UT-M2-01: %q → lexeme %q, want %q", tc.src, tok.Lexeme, tc.src)
		}
	}
}

// UT-M2-02: Line and column tracked correctly across multi-line input.
func TestUT_M2_02_LineColumnTracked(t *testing.T) {
	src := "function foo()\nint x"
	s := scanner.New("test", src)

	// Line 1
	tok := s.Next() // function
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("UT-M2-02: 'function' at %d:%d, want 1:1", tok.Line, tok.Column)
	}
	tok = s.Next() // foo
	if tok.Line != 1 {
		t.Errorf("UT-M2-02: 'foo' line=%d, want 1", tok.Line)
	}
	tok = s.Next() // (
	tok = s.Next() // )

	// Line 2
	tok = s.Next() // int
	if tok.Line != 2 || tok.Column != 1 {
		t.Errorf("UT-M2-02: 'int' at %d:%d, want 2:1", tok.Line, tok.Column)
	}
	tok = s.Next() // x
	if tok.Line != 2 || tok.Column != 5 {
		t.Errorf("UT-M2-02: 'x' at %d:%d, want 2:5", tok.Line, tok.Column)
	}
}

// UT-M2-03: String literals tokenised with correct value (content without quotes).
func TestUT_M2_03_StringLiteralValue(t *testing.T) {
	s := scanner.New("test", `"Hello, world!"`)
	tok := s.Next()
	if tok.Kind != scanner.TokenStringLit {
		t.Fatalf("UT-M2-03: kind=%v, want TokenStringLit", tok.Kind)
	}
	if tok.Lexeme != "Hello, world!" {
		t.Errorf("UT-M2-03: lexeme=%q, want %q", tok.Lexeme, "Hello, world!")
	}
}

// UT-M2-04: Comments are skipped; the identifier after a comment is tokenised.
func TestUT_M2_04_CommentsSkipped(t *testing.T) {
	// Line comment followed by identifier on next line.
	s := scanner.New("test", "// this is a comment\nidentifier")
	tok := s.Next()
	if tok.Kind != scanner.TokenIdent {
		t.Fatalf("UT-M2-04: after line comment, kind=%v, want TokenIdent", tok.Kind)
	}
	if tok.Lexeme != "identifier" {
		t.Errorf("UT-M2-04: lexeme=%q, want %q", tok.Lexeme, "identifier")
	}

	// Block comment followed by identifier on same line.
	s2 := scanner.New("test", "/* block comment */ identifier2")
	tok2 := s2.Next()
	if tok2.Kind != scanner.TokenIdent {
		t.Fatalf("UT-M2-04: after block comment, kind=%v, want TokenIdent", tok2.Kind)
	}
	if tok2.Lexeme != "identifier2" {
		t.Errorf("UT-M2-04: lexeme=%q, want %q", tok2.Lexeme, "identifier2")
	}
}
