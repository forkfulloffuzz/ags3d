// Package viz provides human-readable visualisers for each stage of the
// AGS-spirit → GDScript transpilation pipeline.
//
// Used by the `ag viz` subcommands to let authors and developers inspect
// what each stage produces. All visualisers write to an io.Writer so they
// can be directed to stdout, a file, or a test buffer.
//
// Stages and the tasks that implement them:
//
//	Tokens   — ag viz tokens <file>   — VIZ-01 (T07)
//	AST      — ag viz ast    <file>   — VIZ-02 (T09)
//	Blocking — ag viz blocking <file> — VIZ-03 (T11)
//	Emit     — ag viz emit   <file>   — VIZ-04 (T17)
package viz

import (
	"fmt"
	"io"

	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// --------------------------------------------------------------------
// VIZ-01 — Token stream
// --------------------------------------------------------------------

// Tokens scans src and writes a formatted token table to w.
//
// Example output:
//
//	Tokens — rooms/market/market.agscript
//	LINE  COL  KIND              LEXEME
//	   1    1  FUNCTION          "function"
//	   1   10  IDENT             "room_Load"
//	   ...
//	  12 tokens
//
// TODO(VIZ-01/T07): implement once scanner.Next() is complete.
func Tokens(w io.Writer, file, src string) {
	s := scanner.New(file, src)
	fmt.Fprintf(w, "Tokens — %s\n", file)
	fmt.Fprintf(w, "%-6s %-4s  %-18s  %s\n", "LINE", "COL", "KIND", "LEXEME")

	count := 0
	for {
		tok := s.Next()
		fmt.Fprintf(w, "%6d %4d  %-18s  %q\n", tok.Line, tok.Column, kindName(tok.Kind), tok.Lexeme)
		count++
		if tok.Kind == scanner.TokenEOF {
			break
		}
	}
	fmt.Fprintf(w, "%d tokens\n", count)
}

// --------------------------------------------------------------------
// VIZ-02 — AST tree
// --------------------------------------------------------------------

// AST parses src and writes an indented AST tree to w.
//
// Example output:
//
//	AST — rooms/market/market.agscript
//	File
//	└── FunctionDecl "room_Load" → void  [1:1]
//	    └── Block
//	        └── VarDecl "x": int  [2:5]
//	            └── Literal(int) "42"  [2:13]
//
// TODO(VIZ-02/T09): implement once parser.Parse() builds a real AST.
func AST(w io.Writer, file, src string) {
	s := scanner.New(file, src)
	p := parser.New(s)
	f, errs := p.Parse(file)

	fmt.Fprintf(w, "AST — %s\n", file)

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(w, "  error: %v\n", e)
		}
		return
	}

	fmt.Fprintf(w, "File\n")
	for _, decl := range f.Decls {
		printDecl(w, decl, "└── ", "    ")
	}
}

// --------------------------------------------------------------------
// VIZ-03 — Blocking call annotation
// --------------------------------------------------------------------

// Blocking parses src, runs symbol resolution (T10) and blocking annotation
// (T11), then writes a table of all call sites and their blocking status.
//
// Example output:
//
//	Blocking calls — rooms/market/market.agscript
//	LINE  COL  CALL                                        BLOCKING
//	   5    5  global.player.WalkTo(point.door_left)       YES → await
//	   8    5  getScore()                                  no
//
// TODO(VIZ-03/T11): implement once symbol table and blocking annotation exist.
func Blocking(w io.Writer, file, src string) {
	fmt.Fprintf(w, "Blocking calls — %s\n", file)
	fmt.Fprintf(w, "  (not yet implemented — available after T11)\n")
}

// --------------------------------------------------------------------
// VIZ-04 — Transpiler mapping (side-by-side)
// --------------------------------------------------------------------

// Emit runs the full pipeline (scan → parse → emit) and writes a
// side-by-side view of AGS-spirit source vs generated GDScript,
// with source-map line correspondences highlighted.
//
// Example output:
//
//	Transpile — rooms/market/market.agscript
//	  AGS-spirit                          │  GDScript
//	  ────────────────────────────────────┼────────────────────────────────────
//	  1│ function room_Load() {           │  1│ func room_load():
//	  2│     global.player.WalkTo(…)      │  2│     await AGSRuntime…walk_to(…)
//
// TODO(VIZ-04/T17): implement once emitter produces output and source maps.
func Emit(w io.Writer, file, src string) {
	fmt.Fprintf(w, "Transpile — %s\n", file)
	fmt.Fprintf(w, "  (not yet implemented — available after T17)\n")
}

// --------------------------------------------------------------------
// Internal helpers
// --------------------------------------------------------------------

func printDecl(w io.Writer, decl parser.Decl, prefix, indent string) {
	// TODO(VIZ-02): recursive tree printer — implement alongside T09.
	fmt.Fprintf(w, "%s%T\n", prefix, decl)
}

// kindName returns a short human-readable name for a TokenKind.
func kindName(k scanner.TokenKind) string {
	names := map[scanner.TokenKind]string{
		scanner.TokenEOF:           "EOF",
		scanner.TokenInvalid:       "INVALID",
		scanner.TokenIdent:         "IDENT",
		scanner.TokenIntLit:        "INT_LIT",
		scanner.TokenFloatLit:      "FLOAT_LIT",
		scanner.TokenStringLit:     "STRING_LIT",
		scanner.TokenBool:          "BOOL",
		scanner.TokenChar:          "CHAR",
		scanner.TokenFloat:         "FLOAT",
		scanner.TokenInt:           "INT",
		scanner.TokenShort:         "SHORT",
		scanner.TokenString:        "STRING",
		scanner.TokenVoid:          "VOID",
		scanner.TokenBreak:         "BREAK",
		scanner.TokenCase:          "CASE",
		scanner.TokenContinue:      "CONTINUE",
		scanner.TokenDefault:       "DEFAULT",
		scanner.TokenDo:            "DO",
		scanner.TokenElse:          "ELSE",
		scanner.TokenFor:           "FOR",
		scanner.TokenIf:            "IF",
		scanner.TokenReturn:        "RETURN",
		scanner.TokenSwitch:        "SWITCH",
		scanner.TokenWhile:         "WHILE",
		scanner.TokenEnum:          "ENUM",
		scanner.TokenExport:        "EXPORT",
		scanner.TokenFunction:      "FUNCTION",
		scanner.TokenNamespace:     "NAMESPACE",
		scanner.TokenGlobal:        "GLOBAL",
		scanner.TokenFalse:         "FALSE",
		scanner.TokenNull:          "NULL",
		scanner.TokenTrue:          "TRUE",
		scanner.TokenLParen:        "LPAREN",
		scanner.TokenRParen:        "RPAREN",
		scanner.TokenLBrace:        "LBRACE",
		scanner.TokenRBrace:        "RBRACE",
		scanner.TokenLBracket:      "LBRACKET",
		scanner.TokenRBracket:      "RBRACKET",
		scanner.TokenSemicolon:     "SEMICOLON",
		scanner.TokenComma:         "COMMA",
		scanner.TokenDot:           "DOT",
		scanner.TokenAssign:        "ASSIGN",
		scanner.TokenPlusAssign:    "PLUS_ASSIGN",
		scanner.TokenMinusAssign:   "MINUS_ASSIGN",
		scanner.TokenStarAssign:    "STAR_ASSIGN",
		scanner.TokenSlashAssign:   "SLASH_ASSIGN",
		scanner.TokenPercentAssign: "PCT_ASSIGN",
		scanner.TokenAndAssign:     "AND_ASSIGN",
		scanner.TokenOrAssign:      "OR_ASSIGN",
		scanner.TokenXorAssign:     "XOR_ASSIGN",
		scanner.TokenLShiftAssign:  "LSHIFT_ASSIGN",
		scanner.TokenRShiftAssign:  "RSHIFT_ASSIGN",
		scanner.TokenEq:            "EQ",
		scanner.TokenNeq:           "NEQ",
		scanner.TokenLt:            "LT",
		scanner.TokenLte:           "LTE",
		scanner.TokenGt:            "GT",
		scanner.TokenGte:           "GTE",
		scanner.TokenPlus:          "PLUS",
		scanner.TokenMinus:         "MINUS",
		scanner.TokenStar:          "STAR",
		scanner.TokenSlash:         "SLASH",
		scanner.TokenPercent:       "PERCENT",
		scanner.TokenPlusPlus:      "PLUS_PLUS",
		scanner.TokenMinusMinus:    "MINUS_MINUS",
		scanner.TokenBang:          "BANG",
		scanner.TokenAnd:           "AND",
		scanner.TokenOr:            "OR",
		scanner.TokenAmpersand:     "AMPERSAND",
		scanner.TokenPipe:          "PIPE",
		scanner.TokenCaret:         "CARET",
		scanner.TokenTilde:         "TILDE",
		scanner.TokenLShift:        "LSHIFT",
		scanner.TokenRShift:        "RSHIFT",
	}
	if name, ok := names[k]; ok {
		return name
	}
	return fmt.Sprintf("TOKEN(%d)", int(k))
}
