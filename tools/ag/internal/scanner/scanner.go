// Package scanner implements the AGS-spirit lexer.
package scanner

import "fmt"

// TokenKind identifies the type of a scanned token.
type TokenKind int

const (
	// Special
	TokenEOF     TokenKind = iota
	TokenInvalid           // unrecognised character

	// Literals
	TokenIdent
	TokenIntLit
	TokenFloatLit
	TokenStringLit

	// Keywords — types
	TokenBool   // bool
	TokenChar   // char   (maps to int semantically)
	TokenFloat  // float
	TokenInt    // int
	TokenShort  // short  (maps to int semantically)
	TokenString // string / String
	TokenVoid   // void

	// Keywords — control flow
	TokenBreak    // break
	TokenCase     // case
	TokenContinue // continue
	TokenDefault  // default
	TokenDo       // do
	TokenElse     // else
	TokenFor      // for
	TokenIf       // if
	TokenReturn   // return
	TokenSwitch   // switch
	TokenWhile    // while

	// Keywords — declarations
	TokenEnum      // enum
	TokenExport    // export
	TokenFunction  // function
	TokenNamespace // namespace

	// Keywords — built-in namespaces
	TokenGlobal // global

	// Keywords — values
	TokenFalse // false
	TokenNull  // null
	TokenTrue  // true

	// Punctuation
	TokenLParen    // (
	TokenRParen    // )
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenSemicolon // ;
	TokenComma     // ,
	TokenDot       // .

	// Assignment operators
	TokenAssign        // =
	TokenPlusAssign    // +=
	TokenMinusAssign   // -=
	TokenStarAssign    // *=
	TokenSlashAssign   // /=
	TokenPercentAssign // %=
	TokenAndAssign     // &=
	TokenOrAssign      // |=
	TokenXorAssign     // ^=
	TokenLShiftAssign  // <<=
	TokenRShiftAssign  // >>=

	// Comparison operators
	TokenEq  // ==
	TokenNeq // !=
	TokenLt  // <
	TokenLte // <=
	TokenGt  // >
	TokenGte // >=

	// Arithmetic operators
	TokenPlus    // +
	TokenMinus   // -
	TokenStar    // *
	TokenSlash   // /
	TokenPercent // %

	// Increment / decrement
	TokenPlusPlus   // ++
	TokenMinusMinus // --

	// Logical operators
	TokenBang // !
	TokenAnd  // &&
	TokenOr   // ||

	// Bitwise operators
	TokenAmpersand // &
	TokenPipe      // |
	TokenCaret     // ^
	TokenTilde     // ~
	TokenLShift    // <<
	TokenRShift    // >>

	tokenKindCount
)

// Keywords maps keyword strings to their TokenKind.
// Both "string" and "String" map to TokenString (AGS uses capital S; AGS-spirit accepts both).
var Keywords = map[string]TokenKind{
	// Types
	"bool":   TokenBool,
	"char":   TokenChar,
	"float":  TokenFloat,
	"int":    TokenInt,
	"short":  TokenShort,
	"string": TokenString,
	"String": TokenString, // AGS capital-S alias
	"void":   TokenVoid,

	// Control flow
	"break":    TokenBreak,
	"case":     TokenCase,
	"continue": TokenContinue,
	"default":  TokenDefault,
	"do":       TokenDo,
	"else":     TokenElse,
	"for":      TokenFor,
	"if":       TokenIf,
	"return":   TokenReturn,
	"switch":   TokenSwitch,
	"while":    TokenWhile,

	// Declarations
	"enum":      TokenEnum,
	"export":    TokenExport,
	"function":  TokenFunction,
	"namespace": TokenNamespace,

	// Built-in namespaces
	"global": TokenGlobal,

	// Values
	"false": TokenFalse,
	"null":  TokenNull,
	"true":  TokenTrue,
}

// Token is a single lexed token with source position.
type Token struct {
	Kind   TokenKind
	Lexeme string
	File   string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%s:%d:%d %v %q", t.File, t.Line, t.Column, t.Kind, t.Lexeme)
}

// Scanner tokenises AGS-spirit source one token at a time.
// Call Next() repeatedly; it returns TokenEOF once the source is exhausted.
// The Scanner never panics on bad input — unrecognised characters produce
// TokenInvalid tokens so the parser can report them as errors.
type Scanner struct {
	src  []rune
	file string
	pos  int // current read position in src
	line int
	col  int
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
// Every returned token carries File, Line, Column from the start of the token.
func (s *Scanner) Next() Token {
	s.skipWhitespaceAndComments()

	if s.pos >= len(s.src) {
		return s.tok(TokenEOF, "")
	}

	line, col := s.line, s.col
	ch := s.src[s.pos]

	// --- identifiers and keywords ---
	if isLetter(ch) {
		return s.scanIdent(line, col)
	}

	// --- numeric literals ---
	if isDigit(ch) {
		return s.scanNumber(line, col)
	}

	// --- string literals ---
	if ch == '"' {
		return s.scanString(line, col)
	}

	// --- operators and punctuation ---
	return s.scanSymbol(line, col)
}

// -------------------------------------------------------------------
// Whitespace / comment skipping
// -------------------------------------------------------------------

func (s *Scanner) skipWhitespaceAndComments() {
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r':
			s.advance()
		case ch == '\n':
			s.advanceLine()
		case ch == '/' && s.peek(1) == '/':
			s.skipLineComment()
		case ch == '/' && s.peek(1) == '*':
			s.skipBlockComment()
		default:
			return
		}
	}
}

func (s *Scanner) skipLineComment() {
	for s.pos < len(s.src) && s.src[s.pos] != '\n' {
		s.advance()
	}
}

func (s *Scanner) skipBlockComment() {
	s.advance() // /
	s.advance() // *
	for s.pos < len(s.src) {
		if s.src[s.pos] == '*' && s.peek(1) == '/' {
			s.advance() // *
			s.advance() // /
			return
		}
		if s.src[s.pos] == '\n' {
			s.advanceLine()
		} else {
			s.advance()
		}
	}
	// unterminated block comment — the parser will surface this as an error
	// when it finds EOF unexpectedly; no need to emit a special token here.
}

// -------------------------------------------------------------------
// Identifier / keyword scanning
// -------------------------------------------------------------------

func (s *Scanner) scanIdent(line, col int) Token {
	start := s.pos
	for s.pos < len(s.src) && isLetterOrDigit(s.src[s.pos]) {
		s.advance()
	}
	lexeme := string(s.src[start:s.pos])
	kind, ok := Keywords[lexeme]
	if !ok {
		kind = TokenIdent
	}
	return Token{Kind: kind, Lexeme: lexeme, File: s.file, Line: line, Column: col}
}

// -------------------------------------------------------------------
// Numeric literal scanning
// -------------------------------------------------------------------

func (s *Scanner) scanNumber(line, col int) Token {
	start := s.pos
	for s.pos < len(s.src) && isDigit(s.src[s.pos]) {
		s.advance()
	}
	if s.pos < len(s.src) && s.src[s.pos] == '.' && s.pos+1 < len(s.src) && isDigit(s.src[s.pos+1]) {
		s.advance() // consume '.'
		for s.pos < len(s.src) && isDigit(s.src[s.pos]) {
			s.advance()
		}
		return Token{Kind: TokenFloatLit, Lexeme: string(s.src[start:s.pos]), File: s.file, Line: line, Column: col}
	}
	return Token{Kind: TokenIntLit, Lexeme: string(s.src[start:s.pos]), File: s.file, Line: line, Column: col}
}

// -------------------------------------------------------------------
// String literal scanning
// -------------------------------------------------------------------

func (s *Scanner) scanString(line, col int) Token {
	s.advance() // opening "
	start := s.pos
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		if ch == '"' {
			lexeme := string(s.src[start:s.pos])
			s.advance() // closing "
			return Token{Kind: TokenStringLit, Lexeme: lexeme, File: s.file, Line: line, Column: col}
		}
		if ch == '\\' {
			s.advance() // backslash
			if s.pos < len(s.src) {
				s.advance() // escaped char
			}
			continue
		}
		if ch == '\n' {
			// unterminated string — return what we have as Invalid
			break
		}
		s.advance()
	}
	return Token{Kind: TokenInvalid, Lexeme: "unterminated string", File: s.file, Line: line, Column: col}
}

// -------------------------------------------------------------------
// Operator / punctuation scanning
// -------------------------------------------------------------------

func (s *Scanner) scanSymbol(line, col int) Token {
	ch := s.src[s.pos]
	s.advance()

	next := s.peek(0) // character after ch (already advanced past ch)

	switch ch {
	case '(':
		return s.makeTok(TokenLParen, "(", line, col)
	case ')':
		return s.makeTok(TokenRParen, ")", line, col)
	case '{':
		return s.makeTok(TokenLBrace, "{", line, col)
	case '}':
		return s.makeTok(TokenRBrace, "}", line, col)
	case '[':
		return s.makeTok(TokenLBracket, "[", line, col)
	case ']':
		return s.makeTok(TokenRBracket, "]", line, col)
	case ';':
		return s.makeTok(TokenSemicolon, ";", line, col)
	case ',':
		return s.makeTok(TokenComma, ",", line, col)
	case '.':
		return s.makeTok(TokenDot, ".", line, col)
	case '~':
		return s.makeTok(TokenTilde, "~", line, col)

	case '+':
		if next == '+' {
			s.advance()
			return s.makeTok(TokenPlusPlus, "++", line, col)
		}
		if next == '=' {
			s.advance()
			return s.makeTok(TokenPlusAssign, "+=", line, col)
		}
		return s.makeTok(TokenPlus, "+", line, col)

	case '-':
		if next == '-' {
			s.advance()
			return s.makeTok(TokenMinusMinus, "--", line, col)
		}
		if next == '=' {
			s.advance()
			return s.makeTok(TokenMinusAssign, "-=", line, col)
		}
		return s.makeTok(TokenMinus, "-", line, col)

	case '*':
		if next == '=' {
			s.advance()
			return s.makeTok(TokenStarAssign, "*=", line, col)
		}
		return s.makeTok(TokenStar, "*", line, col)

	case '/':
		if next == '=' {
			s.advance()
			return s.makeTok(TokenSlashAssign, "/=", line, col)
		}
		return s.makeTok(TokenSlash, "/", line, col)

	case '%':
		if next == '=' {
			s.advance()
			return s.makeTok(TokenPercentAssign, "%=", line, col)
		}
		return s.makeTok(TokenPercent, "%", line, col)

	case '!':
		if next == '=' {
			s.advance()
			return s.makeTok(TokenNeq, "!=", line, col)
		}
		return s.makeTok(TokenBang, "!", line, col)

	case '=':
		if next == '=' {
			s.advance()
			return s.makeTok(TokenEq, "==", line, col)
		}
		return s.makeTok(TokenAssign, "=", line, col)

	case '<':
		if next == '<' {
			s.advance()
			if s.peek(0) == '=' {
				s.advance()
				return s.makeTok(TokenLShiftAssign, "<<=", line, col)
			}
			return s.makeTok(TokenLShift, "<<", line, col)
		}
		if next == '=' {
			s.advance()
			return s.makeTok(TokenLte, "<=", line, col)
		}
		return s.makeTok(TokenLt, "<", line, col)

	case '>':
		if next == '>' {
			s.advance()
			if s.peek(0) == '=' {
				s.advance()
				return s.makeTok(TokenRShiftAssign, ">>=", line, col)
			}
			return s.makeTok(TokenRShift, ">>", line, col)
		}
		if next == '=' {
			s.advance()
			return s.makeTok(TokenGte, ">=", line, col)
		}
		return s.makeTok(TokenGt, ">", line, col)

	case '&':
		if next == '&' {
			s.advance()
			return s.makeTok(TokenAnd, "&&", line, col)
		}
		if next == '=' {
			s.advance()
			return s.makeTok(TokenAndAssign, "&=", line, col)
		}
		return s.makeTok(TokenAmpersand, "&", line, col)

	case '|':
		if next == '|' {
			s.advance()
			return s.makeTok(TokenOr, "||", line, col)
		}
		if next == '=' {
			s.advance()
			return s.makeTok(TokenOrAssign, "|=", line, col)
		}
		return s.makeTok(TokenPipe, "|", line, col)

	case '^':
		if next == '=' {
			s.advance()
			return s.makeTok(TokenXorAssign, "^=", line, col)
		}
		return s.makeTok(TokenCaret, "^", line, col)
	}

	return Token{Kind: TokenInvalid, Lexeme: string(ch), File: s.file, Line: line, Column: col}
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func (s *Scanner) advance() {
	s.pos++
	s.col++
}

func (s *Scanner) advanceLine() {
	s.pos++
	s.line++
	s.col = 1
}

// peek returns the character at pos+offset without consuming it.
// Returns 0 if out of bounds.
func (s *Scanner) peek(offset int) rune {
	i := s.pos + offset
	if i >= len(s.src) {
		return 0
	}
	return s.src[i]
}

func (s *Scanner) tok(kind TokenKind, lexeme string) Token {
	return Token{Kind: kind, Lexeme: lexeme, File: s.file, Line: s.line, Column: s.col}
}

func (s *Scanner) makeTok(kind TokenKind, lexeme string, line, col int) Token {
	return Token{Kind: kind, Lexeme: lexeme, File: s.file, Line: line, Column: col}
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isLetterOrDigit(ch rune) bool {
	return isLetter(ch) || isDigit(ch)
}
