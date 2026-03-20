// Package scanner implements the AGS-spirit lexer.
// Stub — full implementation in T07.
package scanner

// TokenKind identifies the type of a scanned token.
type TokenKind int

const (
	// Literals
	TokenEOF TokenKind = iota
	TokenIdent
	TokenIntLit
	TokenFloatLit
	TokenStringLit
	TokenBoolLit

	// Keywords
	TokenFunction
	TokenReturn
	TokenIf
	TokenElse
	TokenWhile
	TokenVoid
	TokenInt
	TokenFloat
	TokenBool
	TokenString

	// Punctuation / operators
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenSemicolon
	TokenComma
	TokenDot
	TokenAssign
	TokenEq
	TokenNeq
	TokenLt
	TokenLte
	TokenGt
	TokenGte
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenBang

	tokenKindCount
)

// Token is a single lexed token with source position.
type Token struct {
	Kind    TokenKind
	Lexeme  string
	File    string
	Line    int
	Column  int
}

// Scanner tokenises AGS-spirit source. Full implementation in T07.
type Scanner struct {
	src    []rune
	file   string
	pos    int
	line   int
	col    int
}

// New creates a Scanner over src from the given filename.
func New(file, src string) *Scanner {
	return &Scanner{
		src:  []rune(src),
		file: file,
		line: 1,
		col:  1,
	}
}

// Next returns the next token. Returns TokenEOF when exhausted.
// TODO(T07): implement full lexer.
func (s *Scanner) Next() Token {
	return Token{Kind: TokenEOF, File: s.file, Line: s.line, Column: s.col}
}
