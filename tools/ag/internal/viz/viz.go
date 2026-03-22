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
//	    └── Block  [1:19]
//	        └── VarDecl "x": int  [2:5]
//	            └── Literal(int) "42"  [2:9]
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
	for i, decl := range f.Decls {
		last := i == len(f.Decls)-1
		nodePrefix, childPrefix := branchPrefixes("", last)
		printDecl(w, decl, nodePrefix, childPrefix)
	}
}

// --------------------------------------------------------------------
// VIZ-03 — Blocking call annotation
// --------------------------------------------------------------------

// Blocking parses src, runs symbol resolution (T10) and blocking annotation
// (T11), then writes a table of all call sites and their blocking status.
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
// side-by-side view of AGS-spirit source vs generated GDScript.
//
// TODO(VIZ-04/T17): implement once emitter produces output and source maps.
func Emit(w io.Writer, file, src string) {
	fmt.Fprintf(w, "Transpile — %s\n", file)
	fmt.Fprintf(w, "  (not yet implemented — available after T17)\n")
}

// --------------------------------------------------------------------
// AST tree printer helpers
// --------------------------------------------------------------------

// branchPrefixes returns the (nodePrefix, childPrefix) pair for a tree node.
// last=true → final child in a list (└──); last=false → non-final (├──).
func branchPrefixes(indent string, last bool) (nodePrefix, childPrefix string) {
	if last {
		return indent + "└── ", indent + "    "
	}
	return indent + "├── ", indent + "│   "
}

func printDecl(w io.Writer, decl parser.Decl, prefix, indent string) {
	switch d := decl.(type) {
	case *parser.FunctionDecl:
		retType := d.ReturnType
		if retType == "" {
			retType = "void"
		}
		flags := ""
		if d.IsExport {
			flags += " [export]"
		}
		if d.IsBlocking {
			flags += " [blocking]"
		}
		paramsStr := ""
		for i, param := range d.Params {
			if i > 0 {
				paramsStr += ", "
			}
			paramsStr += param.Type + " " + param.Name
		}
		fmt.Fprintf(w, "%sFunctionDecl %q → %s%s  (%s)  [%d:%d]\n",
			prefix, d.Name, retType, flags, paramsStr, d.Pos.Line, d.Pos.Column)
		if d.Body != nil {
			bp, bcp := branchPrefixes(indent, true)
			printBlock(w, d.Body, bp, bcp)
		}

	case *parser.NamespaceDecl:
		fmt.Fprintf(w, "%sNamespaceDecl %q  [%d:%d]\n", prefix, d.Name, d.Pos.Line, d.Pos.Column)
		for i, m := range d.Members {
			mp, mcp := branchPrefixes(indent, i == len(d.Members)-1)
			printDecl(w, m, mp, mcp)
		}

	case *parser.EnumDecl:
		fmt.Fprintf(w, "%sEnumDecl %q  [%d:%d]\n", prefix, d.Name, d.Pos.Line, d.Pos.Column)
		for i, m := range d.Members {
			mp, _ := branchPrefixes(indent, i == len(d.Members)-1)
			if m.Value != nil {
				fmt.Fprintf(w, "%sEnumMember %q = ...\n", mp, m.Name)
			} else {
				fmt.Fprintf(w, "%sEnumMember %q\n", mp, m.Name)
			}
		}

	case *parser.TopVarDecl:
		if d.Decl != nil {
			fmt.Fprintf(w, "%sTopVarDecl %q: %s  [%d:%d]\n",
				prefix, d.Decl.Name, d.Decl.Type, d.Decl.Pos.Line, d.Decl.Pos.Column)
		}

	default:
		fmt.Fprintf(w, "%s%T\n", prefix, decl)
	}
}

func printBlock(w io.Writer, b *parser.Block, prefix, indent string) {
	fmt.Fprintf(w, "%sBlock  [%d:%d]\n", prefix, b.Pos.Line, b.Pos.Column)
	for i, s := range b.Stmts {
		sp, scp := branchPrefixes(indent, i == len(b.Stmts)-1)
		printStmt(w, s, sp, scp)
	}
}

func printStmt(w io.Writer, stmt parser.Stmt, prefix, indent string) {
	if stmt == nil {
		fmt.Fprintf(w, "%s<nil stmt>\n", prefix)
		return
	}
	switch s := stmt.(type) {
	case *parser.Block:
		printBlock(w, s, prefix, indent)

	case *parser.VarDecl:
		fmt.Fprintf(w, "%sVarDecl %q: %s  [%d:%d]\n", prefix, s.Name, s.Type, s.Pos.Line, s.Pos.Column)
		if s.Init != nil {
			ip, icp := branchPrefixes(indent, true)
			printExpr(w, s.Init, ip, icp)
		}

	case *parser.IfStmt:
		fmt.Fprintf(w, "%sIfStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)
		hasElse := s.Else != nil
		condP, condCP := branchPrefixes(indent, false)
		fmt.Fprintf(w, "%sCond:\n", condP)
		printExpr(w, s.Cond, condCP+"└── ", condCP+"    ")
		thenP, thenCP := branchPrefixes(indent, !hasElse)
		fmt.Fprintf(w, "%sThen:\n", thenP)
		printBlock(w, s.Then, thenCP+"└── ", thenCP+"    ")
		if hasElse {
			elseP, elseCP := branchPrefixes(indent, true)
			fmt.Fprintf(w, "%sElse:\n", elseP)
			printStmt(w, s.Else, elseCP+"└── ", elseCP+"    ")
		}

	case *parser.WhileStmt:
		fmt.Fprintf(w, "%sWhileStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)
		condP, condCP := branchPrefixes(indent, false)
		fmt.Fprintf(w, "%sCond:\n", condP)
		printExpr(w, s.Cond, condCP+"└── ", condCP+"    ")
		bodyP, bodyCP := branchPrefixes(indent, true)
		printBlock(w, s.Body, bodyP, bodyCP)

	case *parser.DoWhileStmt:
		fmt.Fprintf(w, "%sDoWhileStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)
		bodyP, bodyCP := branchPrefixes(indent, false)
		printBlock(w, s.Body, bodyP, bodyCP)
		condP, condCP := branchPrefixes(indent, true)
		fmt.Fprintf(w, "%sCond:\n", condP)
		printExpr(w, s.Cond, condCP+"└── ", condCP+"    ")

	case *parser.ForStmt:
		fmt.Fprintf(w, "%sForStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)
		if s.Init != nil {
			ip, icp := branchPrefixes(indent, s.Cond == nil && s.Post == nil)
			fmt.Fprintf(w, "%sInit:\n", ip)
			printStmt(w, s.Init, icp+"└── ", icp+"    ")
		}
		if s.Cond != nil {
			cp, ccp := branchPrefixes(indent, s.Post == nil)
			fmt.Fprintf(w, "%sCond:\n", cp)
			printExpr(w, s.Cond, ccp+"└── ", ccp+"    ")
		}
		if s.Post != nil {
			pp, pcp := branchPrefixes(indent, true)
			fmt.Fprintf(w, "%sPost:\n", pp)
			printExpr(w, s.Post, pcp+"└── ", pcp+"    ")
		}
		bp, bcp := branchPrefixes(indent, true)
		printBlock(w, s.Body, bp, bcp)

	case *parser.SwitchStmt:
		fmt.Fprintf(w, "%sSwitchStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)
		tagP, tagCP := branchPrefixes(indent, len(s.Cases) == 0)
		fmt.Fprintf(w, "%sTag:\n", tagP)
		printExpr(w, s.Tag, tagCP+"└── ", tagCP+"    ")
		for i, c := range s.Cases {
			cp, ccp := branchPrefixes(indent, i == len(s.Cases)-1)
			if c.Value == nil {
				fmt.Fprintf(w, "%sDefault:  [%d:%d]\n", cp, c.Pos.Line, c.Pos.Column)
			} else {
				fmt.Fprintf(w, "%sCase:  [%d:%d]\n", cp, c.Pos.Line, c.Pos.Column)
				vp, vcp := branchPrefixes(ccp, len(c.Body) == 0)
				printExpr(w, c.Value, vp, vcp)
			}
			for j, cs := range c.Body {
				sp, scp := branchPrefixes(ccp, j == len(c.Body)-1)
				printStmt(w, cs, sp, scp)
			}
		}

	case *parser.ReturnStmt:
		fmt.Fprintf(w, "%sReturnStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)
		if s.Value != nil {
			vp, vcp := branchPrefixes(indent, true)
			printExpr(w, s.Value, vp, vcp)
		}

	case *parser.BreakStmt:
		fmt.Fprintf(w, "%sBreakStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)

	case *parser.ContinueStmt:
		fmt.Fprintf(w, "%sContinueStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)

	case *parser.ExprStmt:
		fmt.Fprintf(w, "%sExprStmt  [%d:%d]\n", prefix, s.Pos.Line, s.Pos.Column)
		if s.X != nil {
			xp, xcp := branchPrefixes(indent, true)
			printExpr(w, s.X, xp, xcp)
		}

	default:
		fmt.Fprintf(w, "%s%T\n", prefix, stmt)
	}
}

func printExpr(w io.Writer, expr parser.Expr, prefix, indent string) {
	if expr == nil {
		fmt.Fprintf(w, "%s<nil expr>\n", prefix)
		return
	}
	switch e := expr.(type) {
	case *parser.Literal:
		fmt.Fprintf(w, "%sLiteral(%s) %q  [%d:%d]\n", prefix, e.Kind, e.Value, e.Pos.Line, e.Pos.Column)

	case *parser.Identifier:
		fmt.Fprintf(w, "%sIdent %q  [%d:%d]\n", prefix, e.Name, e.Pos.Line, e.Pos.Column)

	case *parser.GlobalExpr:
		fmt.Fprintf(w, "%sGlobal.%s  [%d:%d]\n", prefix, e.Property, e.Pos.Line, e.Pos.Column)

	case *parser.AssignExpr:
		fmt.Fprintf(w, "%sAssignExpr %q  [%d:%d]\n", prefix, e.Op, e.Pos.Line, e.Pos.Column)
		lp, lcp := branchPrefixes(indent, false)
		rp, rcp := branchPrefixes(indent, true)
		printExpr(w, e.Target, lp, lcp)
		printExpr(w, e.Value, rp, rcp)

	case *parser.BinaryExpr:
		fmt.Fprintf(w, "%sBinaryExpr %q  [%d:%d]\n", prefix, e.Op, e.Pos.Line, e.Pos.Column)
		lp, lcp := branchPrefixes(indent, false)
		rp, rcp := branchPrefixes(indent, true)
		printExpr(w, e.Left, lp, lcp)
		printExpr(w, e.Right, rp, rcp)

	case *parser.UnaryExpr:
		fmt.Fprintf(w, "%sUnaryExpr %q  [%d:%d]\n", prefix, e.Op, e.Pos.Line, e.Pos.Column)
		xp, xcp := branchPrefixes(indent, true)
		printExpr(w, e.X, xp, xcp)

	case *parser.PostfixExpr:
		fmt.Fprintf(w, "%sPostfixExpr %q  [%d:%d]\n", prefix, e.Op, e.Pos.Line, e.Pos.Column)
		xp, xcp := branchPrefixes(indent, true)
		printExpr(w, e.X, xp, xcp)

	case *parser.CallExpr:
		fmt.Fprintf(w, "%sCallExpr  [%d:%d]\n", prefix, e.Pos.Line, e.Pos.Column)
		calleeP, calleeCP := branchPrefixes(indent, len(e.Args) == 0)
		printExpr(w, e.Callee, calleeP, calleeCP)
		for i, arg := range e.Args {
			ap, acp := branchPrefixes(indent, i == len(e.Args)-1)
			printExpr(w, arg, ap, acp)
		}

	case *parser.MemberExpr:
		fmt.Fprintf(w, "%sMemberExpr .%s  [%d:%d]\n", prefix, e.Field, e.Pos.Line, e.Pos.Column)
		op, ocp := branchPrefixes(indent, true)
		printExpr(w, e.Object, op, ocp)

	case *parser.IndexExpr:
		fmt.Fprintf(w, "%sIndexExpr  [%d:%d]\n", prefix, e.Pos.Line, e.Pos.Column)
		op, ocp := branchPrefixes(indent, false)
		ip, icp := branchPrefixes(indent, true)
		printExpr(w, e.Object, op, ocp)
		printExpr(w, e.Index, ip, icp)

	default:
		fmt.Fprintf(w, "%s%T\n", prefix, expr)
	}
}

// --------------------------------------------------------------------
// kindName — human-readable token kind labels
// --------------------------------------------------------------------

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
		scanner.TokenColon:         "COLON",
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
