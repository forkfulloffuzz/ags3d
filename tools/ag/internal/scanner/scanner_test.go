package scanner_test

import (
	"testing"

	"github.com/ags3d/ag/internal/scanner"
)

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func scan(src string) []scanner.Token {
	s := scanner.New("test.agscript", src)
	var toks []scanner.Token
	for {
		t := s.Next()
		toks = append(toks, t)
		if t.Kind == scanner.TokenEOF {
			break
		}
	}
	return toks
}

func kinds(src string) []scanner.TokenKind {
	toks := scan(src)
	ks := make([]scanner.TokenKind, len(toks))
	for i, t := range toks {
		ks[i] = t.Kind
	}
	return ks
}

func assertTokens(t *testing.T, src string, want []scanner.TokenKind) {
	t.Helper()
	got := kinds(src)
	if len(got) != len(want) {
		t.Fatalf("token count: got %d, want %d\nsrc: %q\ngot kinds: %v", len(got), len(want), src, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d]: got %v, want %v (src=%q)", i, got[i], want[i], src)
		}
	}
}

// -------------------------------------------------------------------
// EOF / empty
// -------------------------------------------------------------------

func TestScanner_EOFOnEmpty(t *testing.T) {
	toks := scan("")
	if len(toks) != 1 || toks[0].Kind != scanner.TokenEOF {
		t.Errorf("expected single EOF, got %v", toks)
	}
}

func TestScanner_FileNamePreserved(t *testing.T) {
	s := scanner.New("rooms/start.agscript", "")
	tok := s.Next()
	if tok.File != "rooms/start.agscript" {
		t.Errorf("File = %q, want rooms/start.agscript", tok.File)
	}
}

// -------------------------------------------------------------------
// Whitespace and comments
// -------------------------------------------------------------------

func TestScanner_WhitespaceIgnored(t *testing.T) {
	assertTokens(t, "   \t\r\n  ", []scanner.TokenKind{scanner.TokenEOF})
}

func TestScanner_LineComment(t *testing.T) {
	assertTokens(t, "// this is a comment\n", []scanner.TokenKind{scanner.TokenEOF})
}

func TestScanner_LineCommentDoesNotConsumeNextLine(t *testing.T) {
	assertTokens(t, "// comment\nint", []scanner.TokenKind{scanner.TokenInt, scanner.TokenEOF})
}

func TestScanner_BlockComment(t *testing.T) {
	assertTokens(t, "/* block comment */", []scanner.TokenKind{scanner.TokenEOF})
}

func TestScanner_BlockCommentMultiline(t *testing.T) {
	assertTokens(t, "/* line one\n   line two\n*/int", []scanner.TokenKind{scanner.TokenInt, scanner.TokenEOF})
}

// -------------------------------------------------------------------
// Identifiers
// -------------------------------------------------------------------

func TestScanner_Identifier(t *testing.T) {
	toks := scan("myVar")
	if toks[0].Kind != scanner.TokenIdent || toks[0].Lexeme != "myVar" {
		t.Errorf("got %v", toks[0])
	}
}

func TestScanner_IdentifierWithUnderscore(t *testing.T) {
	toks := scan("room_Load")
	if toks[0].Kind != scanner.TokenIdent || toks[0].Lexeme != "room_Load" {
		t.Errorf("got %v", toks[0])
	}
}

func TestScanner_IdentifierWithDigits(t *testing.T) {
	toks := scan("var123")
	if toks[0].Kind != scanner.TokenIdent || toks[0].Lexeme != "var123" {
		t.Errorf("got %v", toks[0])
	}
}

// -------------------------------------------------------------------
// Keywords
// -------------------------------------------------------------------

func TestScanner_Keywords(t *testing.T) {
	cases := []struct {
		src  string
		kind scanner.TokenKind
	}{
		{"bool", scanner.TokenBool},
		{"break", scanner.TokenBreak},
		{"case", scanner.TokenCase},
		{"char", scanner.TokenChar},
		{"continue", scanner.TokenContinue},
		{"default", scanner.TokenDefault},
		{"do", scanner.TokenDo},
		{"else", scanner.TokenElse},
		{"enum", scanner.TokenEnum},
		{"export", scanner.TokenExport},
		{"false", scanner.TokenFalse},
		{"float", scanner.TokenFloat},
		{"for", scanner.TokenFor},
		{"function", scanner.TokenFunction},
		{"global", scanner.TokenGlobal},
		{"if", scanner.TokenIf},
		{"int", scanner.TokenInt},
		{"namespace", scanner.TokenNamespace},
		{"null", scanner.TokenNull},
		{"return", scanner.TokenReturn},
		{"short", scanner.TokenShort},
		{"string", scanner.TokenString},
		{"String", scanner.TokenString}, // capital-S alias
		{"switch", scanner.TokenSwitch},
		{"true", scanner.TokenTrue},
		{"void", scanner.TokenVoid},
		{"while", scanner.TokenWhile},
	}
	for _, tc := range cases {
		toks := scan(tc.src)
		if toks[0].Kind != tc.kind {
			t.Errorf("keyword %q: got %v, want %v", tc.src, toks[0].Kind, tc.kind)
		}
		if toks[0].Lexeme != tc.src {
			t.Errorf("keyword %q: lexeme = %q, want %q", tc.src, toks[0].Lexeme, tc.src)
		}
	}
}

func TestScanner_KeywordNotSubstringOfIdent(t *testing.T) {
	// "integer" should be TokenIdent, not TokenInt + ident
	toks := scan("integer")
	if toks[0].Kind != scanner.TokenIdent {
		t.Errorf("expected TokenIdent for 'integer', got %v", toks[0].Kind)
	}
}

// -------------------------------------------------------------------
// Integer literals
// -------------------------------------------------------------------

func TestScanner_IntLiteral(t *testing.T) {
	toks := scan("42")
	if toks[0].Kind != scanner.TokenIntLit || toks[0].Lexeme != "42" {
		t.Errorf("got %v", toks[0])
	}
}

func TestScanner_IntLiteralZero(t *testing.T) {
	toks := scan("0")
	if toks[0].Kind != scanner.TokenIntLit || toks[0].Lexeme != "0" {
		t.Errorf("got %v", toks[0])
	}
}

// -------------------------------------------------------------------
// Float literals
// -------------------------------------------------------------------

func TestScanner_FloatLiteral(t *testing.T) {
	toks := scan("3.14")
	if toks[0].Kind != scanner.TokenFloatLit || toks[0].Lexeme != "3.14" {
		t.Errorf("got %v", toks[0])
	}
}

func TestScanner_FloatLiteralLeadingZero(t *testing.T) {
	toks := scan("0.5")
	if toks[0].Kind != scanner.TokenFloatLit || toks[0].Lexeme != "0.5" {
		t.Errorf("got %v", toks[0])
	}
}

func TestScanner_IntNotFloat_TrailingDot(t *testing.T) {
	// "5." with no digit after dot is two tokens: IntLit "5" and Dot "."
	assertTokens(t, "5.", []scanner.TokenKind{scanner.TokenIntLit, scanner.TokenDot, scanner.TokenEOF})
}

// -------------------------------------------------------------------
// String literals
// -------------------------------------------------------------------

func TestScanner_StringLiteral(t *testing.T) {
	toks := scan(`"hello"`)
	if toks[0].Kind != scanner.TokenStringLit || toks[0].Lexeme != "hello" {
		t.Errorf("got %v", toks[0])
	}
}

func TestScanner_StringLiteralEmpty(t *testing.T) {
	toks := scan(`""`)
	if toks[0].Kind != scanner.TokenStringLit || toks[0].Lexeme != "" {
		t.Errorf("got %v", toks[0])
	}
}

func TestScanner_StringLiteralEscapeSequences(t *testing.T) {
	toks := scan(`"line\nnewline"`)
	if toks[0].Kind != scanner.TokenStringLit {
		t.Errorf("got kind %v, want TokenStringLit", toks[0].Kind)
	}
}

func TestScanner_UnterminatedString(t *testing.T) {
	toks := scan(`"oops`)
	if toks[0].Kind != scanner.TokenInvalid {
		t.Errorf("expected TokenInvalid for unterminated string, got %v", toks[0].Kind)
	}
}

// -------------------------------------------------------------------
// Operators
// -------------------------------------------------------------------

func TestScanner_ArithmeticOperators(t *testing.T) {
	assertTokens(t, "+ - * / %", []scanner.TokenKind{
		scanner.TokenPlus, scanner.TokenMinus, scanner.TokenStar,
		scanner.TokenSlash, scanner.TokenPercent, scanner.TokenEOF,
	})
}

func TestScanner_ComparisonOperators(t *testing.T) {
	assertTokens(t, "== != < <= > >=", []scanner.TokenKind{
		scanner.TokenEq, scanner.TokenNeq,
		scanner.TokenLt, scanner.TokenLte,
		scanner.TokenGt, scanner.TokenGte,
		scanner.TokenEOF,
	})
}

func TestScanner_LogicalOperators(t *testing.T) {
	assertTokens(t, "! && ||", []scanner.TokenKind{
		scanner.TokenBang, scanner.TokenAnd, scanner.TokenOr, scanner.TokenEOF,
	})
}

func TestScanner_BitwiseOperators(t *testing.T) {
	assertTokens(t, "& | ^ ~ << >>", []scanner.TokenKind{
		scanner.TokenAmpersand, scanner.TokenPipe, scanner.TokenCaret,
		scanner.TokenTilde, scanner.TokenLShift, scanner.TokenRShift,
		scanner.TokenEOF,
	})
}

func TestScanner_IncrementDecrement(t *testing.T) {
	assertTokens(t, "++ --", []scanner.TokenKind{
		scanner.TokenPlusPlus, scanner.TokenMinusMinus, scanner.TokenEOF,
	})
}

func TestScanner_AssignmentOperators(t *testing.T) {
	assertTokens(t, "= += -= *= /= %= &= |= ^= <<= >>=", []scanner.TokenKind{
		scanner.TokenAssign,
		scanner.TokenPlusAssign, scanner.TokenMinusAssign,
		scanner.TokenStarAssign, scanner.TokenSlashAssign, scanner.TokenPercentAssign,
		scanner.TokenAndAssign, scanner.TokenOrAssign, scanner.TokenXorAssign,
		scanner.TokenLShiftAssign, scanner.TokenRShiftAssign,
		scanner.TokenEOF,
	})
}

// -------------------------------------------------------------------
// Punctuation
// -------------------------------------------------------------------

func TestScanner_Punctuation(t *testing.T) {
	assertTokens(t, "( ) { } [ ] ; , .", []scanner.TokenKind{
		scanner.TokenLParen, scanner.TokenRParen,
		scanner.TokenLBrace, scanner.TokenRBrace,
		scanner.TokenLBracket, scanner.TokenRBracket,
		scanner.TokenSemicolon, scanner.TokenComma, scanner.TokenDot,
		scanner.TokenEOF,
	})
}

// -------------------------------------------------------------------
// Source positions
// -------------------------------------------------------------------

func TestScanner_LineAndColumn(t *testing.T) {
	toks := scan("int x")
	// "int" starts at line 1, col 1
	if toks[0].Line != 1 || toks[0].Column != 1 {
		t.Errorf("int: line=%d col=%d, want 1:1", toks[0].Line, toks[0].Column)
	}
	// "x" starts at line 1, col 5
	if toks[1].Line != 1 || toks[1].Column != 5 {
		t.Errorf("x: line=%d col=%d, want 1:5", toks[1].Line, toks[1].Column)
	}
}

func TestScanner_LineTracking(t *testing.T) {
	toks := scan("int\nx")
	if toks[1].Line != 2 || toks[1].Column != 1 {
		t.Errorf("x after newline: line=%d col=%d, want 2:1", toks[1].Line, toks[1].Column)
	}
}

func TestScanner_ColumnResetAfterNewline(t *testing.T) {
	toks := scan("aaa\nbb")
	// "bb" is on line 2
	if toks[1].Line != 2 {
		t.Errorf("want line 2, got %d", toks[1].Line)
	}
	if toks[1].Column != 1 {
		t.Errorf("want col 1, got %d", toks[1].Column)
	}
}

// -------------------------------------------------------------------
// Unknown characters
// -------------------------------------------------------------------

func TestScanner_UnknownCharProducesInvalid(t *testing.T) {
	toks := scan("@")
	if toks[0].Kind != scanner.TokenInvalid {
		t.Errorf("expected TokenInvalid for '@', got %v", toks[0].Kind)
	}
}

func TestScanner_ScanContinuesAfterInvalid(t *testing.T) {
	// Scanner should not get stuck on an unknown char
	toks := scan("@ int")
	if len(toks) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d: %v", len(toks), toks)
	}
	if toks[1].Kind != scanner.TokenInt {
		t.Errorf("expected TokenInt after invalid, got %v", toks[1])
	}
}

// -------------------------------------------------------------------
// Real AGS-spirit snippets
// -------------------------------------------------------------------

func TestScanner_FunctionDeclaration(t *testing.T) {
	assertTokens(t, "function room_Load()", []scanner.TokenKind{
		scanner.TokenFunction, scanner.TokenIdent,
		scanner.TokenLParen, scanner.TokenRParen,
		scanner.TokenEOF,
	})
}

func TestScanner_ForLoop(t *testing.T) {
	assertTokens(t, "for (int i = 0; i < 3; i++)", []scanner.TokenKind{
		scanner.TokenFor, scanner.TokenLParen,
		scanner.TokenInt, scanner.TokenIdent, scanner.TokenAssign, scanner.TokenIntLit, scanner.TokenSemicolon,
		scanner.TokenIdent, scanner.TokenLt, scanner.TokenIntLit, scanner.TokenSemicolon,
		scanner.TokenIdent, scanner.TokenPlusPlus,
		scanner.TokenRParen, scanner.TokenEOF,
	})
}

func TestScanner_GlobalNamespace(t *testing.T) {
	assertTokens(t, "global.player", []scanner.TokenKind{
		scanner.TokenGlobal, scanner.TokenDot, scanner.TokenIdent, scanner.TokenEOF,
	})
}

func TestScanner_NamespaceBlock(t *testing.T) {
	assertTokens(t, "namespace CharUtils { export function Foo() {} }", []scanner.TokenKind{
		scanner.TokenNamespace, scanner.TokenIdent, scanner.TokenLBrace,
		scanner.TokenExport, scanner.TokenFunction, scanner.TokenIdent,
		scanner.TokenLParen, scanner.TokenRParen,
		scanner.TokenLBrace, scanner.TokenRBrace,
		scanner.TokenRBrace, scanner.TokenEOF,
	})
}

func TestScanner_EnumDeclaration(t *testing.T) {
	assertTokens(t, "enum Dir { eNorth = 0, eSouth };", []scanner.TokenKind{
		scanner.TokenEnum, scanner.TokenIdent, scanner.TokenLBrace,
		scanner.TokenIdent, scanner.TokenAssign, scanner.TokenIntLit, scanner.TokenComma,
		scanner.TokenIdent,
		scanner.TokenRBrace, scanner.TokenSemicolon, scanner.TokenEOF,
	})
}

func TestScanner_BlockingCallChain(t *testing.T) {
	assertTokens(t, `global.player.WalkTo(point.door_left);`, []scanner.TokenKind{
		scanner.TokenGlobal, scanner.TokenDot, scanner.TokenIdent,
		scanner.TokenDot, scanner.TokenIdent,
		scanner.TokenLParen,
		scanner.TokenIdent, scanner.TokenDot, scanner.TokenIdent,
		scanner.TokenRParen, scanner.TokenSemicolon, scanner.TokenEOF,
	})
}
