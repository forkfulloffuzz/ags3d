// Package scanner implements the AGS-spirit lexer.
// Stub — full implementation in T07.
package scanner

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
	TokenString // string
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
	TokenExport    // export   (marks a namespace member as publicly callable)
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
	TokenAssign       // =
	TokenPlusAssign   // +=
	TokenMinusAssign  // -=
	TokenStarAssign   // *=
	TokenSlashAssign  // /=
	TokenPercentAssign // %=
	TokenAndAssign    // &=
	TokenOrAssign     // |=
	TokenXorAssign    // ^=
	TokenLShiftAssign // <<=
	TokenRShiftAssign // >>=

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
	TokenPlusPlus  // ++
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
	"enum":     TokenEnum,
	"function": TokenFunction,
	"export":    TokenExport,
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

// Scanner tokenises AGS-spirit source. Full implementation in T07.
type Scanner struct {
	src  []rune
	file string
	pos  int
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
// TODO(T07): implement full lexer.
func (s *Scanner) Next() Token {
	return Token{Kind: TokenEOF, File: s.file, Line: s.line, Column: s.col}
}
