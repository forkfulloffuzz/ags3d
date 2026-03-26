// Package emitter generates GDScript from an AGS-spirit AST (T13–T17).
//
// The pipeline inside Emit():
//  1. Build symbol table (T10) to resolve names.
//  2. Annotate blocking calls (T11) so CallExpr.IsBlocking is populated.
//  3. Walk the AST with a printer (indentation-tracking output buffer) and
//     emit each node as GDScript.
//
// Key mappings:
//   - AGS-spirit PascalCase function/method names → GDScript snake_case
//   - `global.X` → X  (global namespace is flat in the runtime)
//   - Blocking CallExprs → `await <call>`  (T16)
//   - C-style for loop → var decl + while loop  (GDScript has no C-style for)
//   - switch/case → GDScript match/pattern
//   - namespace Foo { } → inner class Foo:
//   - `&&` / `||` / `!` → `and` / `or` / `not`
//   - Each emitted line is recorded in SourceMap as [gd_line, src_file, src_line] (T17)
package emitter

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ags3d/ag/internal/parser"
)

// Result holds the emitted GDScript and source map for one file.
type Result struct {
	// GDScript is the emitted source, ready to write to .engine/generated/.
	GDScript string
	// SourceMap maps each 1-based GDScript line to the originating
	// AGS-spirit file and line. Format mirrors the .agmap JSON schema:
	// [[gdscript_line, "rel/path.agscript", agscript_line], ...]
	// TODO(T17): populate during emission.
	SourceMap [][3]any
}

// Emitter walks an AGS-spirit AST and produces GDScript output.
type Emitter struct{}

// New creates a new Emitter.
func New() *Emitter { return &Emitter{} }

// Emit generates GDScript for the given parsed file.
func (em *Emitter) Emit(f *parser.File) (*Result, error) {
	// Build symbol table and annotate blocking calls so that
	// CallExpr.IsBlocking is set before we walk the tree.
	st, _ := parser.BuildSymbolTable(f)
	parser.AnnotateBlocking(f, st)

	p := &printer{srcFile: f.Path}
	p.emitFile(f)
	return &Result{GDScript: p.result(), SourceMap: p.sm}, nil
}

// -------------------------------------------------------------------
// printer — indentation-aware GDScript output buffer
// -------------------------------------------------------------------

type printer struct {
	buf     strings.Builder
	depth   int
	lineNum int    // 1-based GDScript output line counter
	srcFile string // AGS-spirit source file path
	srcLine int    // current AGS-spirit source line (0 = unknown)
	sm      [][3]any
}

func (p *printer) result() string { return p.buf.String() }

// setSrc records the AGS-spirit source line for subsequent emitted lines.
func (p *printer) setSrc(line int) { p.srcLine = line }

// line writes one indented line followed by a newline and records a source map entry.
func (p *printer) line(s string) {
	if p.depth > 0 {
		p.buf.WriteString(strings.Repeat("\t", p.depth))
	}
	p.buf.WriteString(s)
	p.buf.WriteByte('\n')
	p.lineNum++
	if p.srcFile != "" && p.srcLine > 0 {
		p.sm = append(p.sm, [3]any{p.lineNum, p.srcFile, p.srcLine})
	}
}

func (p *printer) linef(format string, args ...any) {
	p.line(fmt.Sprintf(format, args...))
}

func (p *printer) blank() { p.buf.WriteByte('\n'); p.lineNum++ }
func (p *printer) push()  { p.depth++ }
func (p *printer) pop()   { p.depth-- }

// -------------------------------------------------------------------
// File
// -------------------------------------------------------------------

func (p *printer) emitFile(f *parser.File) {
	for i, d := range f.Decls {
		if i > 0 {
			p.blank()
		}
		p.emitDecl(d)
	}
	if len(f.Decls) > 0 {
		p.blank()
	}
}

// -------------------------------------------------------------------
// Declarations
// -------------------------------------------------------------------

func (p *printer) emitDecl(d parser.Decl) {
	switch v := d.(type) {
	case *parser.FunctionDecl:
		p.emitFuncDecl(v, false)
	case *parser.NamespaceDecl:
		p.emitNamespaceDecl(v)
	case *parser.EnumDecl:
		p.emitEnumDecl(v)
	case *parser.TopVarDecl:
		p.emitTopVarDecl(v)
	}
}

// emitFuncDecl emits a GDScript func definition.
// static=true is used for namespace members exported as static methods.
func (p *printer) emitFuncDecl(fd *parser.FunctionDecl, static bool) {
	p.setSrc(fd.Pos.Line)
	prefix := ""
	if static {
		prefix = "static "
	}
	params := p.formatParams(fd.Params)
	ret := ""
	if fd.ReturnType != "" && fd.ReturnType != "void" {
		ret = " -> " + mapType(fd.ReturnType)
	}
	p.linef("%sfunc %s(%s)%s:", prefix, toSnakeCase(fd.Name), params, ret)
	p.push()
	if fd.Body == nil || len(fd.Body.Stmts) == 0 {
		p.line("pass")
	} else {
		p.emitBlock(fd.Body)
	}
	p.pop()
}

func (p *printer) formatParams(params []parser.Param) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, param := range params {
		name := toSnakeCase(param.Name)
		if param.Type != "" {
			parts[i] = name + ": " + mapType(param.Type)
		} else {
			parts[i] = name
		}
	}
	return strings.Join(parts, ", ")
}

// emitNamespaceDecl emits a GDScript inner class for an AGS-spirit namespace.
func (p *printer) emitNamespaceDecl(nd *parser.NamespaceDecl) {
	p.setSrc(nd.Pos.Line)
	p.linef("class %s:", nd.Name)
	p.push()
	if len(nd.Members) == 0 {
		p.line("pass")
	} else {
		for i, m := range nd.Members {
			if i > 0 {
				p.blank()
			}
			if fd, ok := m.(*parser.FunctionDecl); ok {
				p.emitFuncDecl(fd, fd.IsExport)
			} else {
				p.emitDecl(m)
			}
		}
	}
	p.pop()
}

func (p *printer) emitEnumDecl(ed *parser.EnumDecl) {
	p.setSrc(ed.Pos.Line)
	if len(ed.Members) == 0 {
		p.linef("enum %s {}", ed.Name)
		return
	}
	parts := make([]string, 0, len(ed.Members))
	for _, m := range ed.Members {
		if m.Value != nil {
			parts = append(parts, m.Name+" = "+p.exprStr(m.Value))
		} else {
			parts = append(parts, m.Name)
		}
	}
	p.linef("enum %s { %s }", ed.Name, strings.Join(parts, ", "))
}

func (p *printer) emitTopVarDecl(tv *parser.TopVarDecl) {
	if tv.Decl == nil {
		return
	}
	vd := tv.Decl
	p.setSrc(vd.Pos.Line)
	if vd.Type != "" {
		if vd.Init != nil {
			p.linef("var %s: %s = %s", toSnakeCase(vd.Name), mapType(vd.Type), p.exprStr(vd.Init))
		} else {
			p.linef("var %s: %s", toSnakeCase(vd.Name), mapType(vd.Type))
		}
	} else if vd.Init != nil {
		p.linef("var %s = %s", toSnakeCase(vd.Name), p.exprStr(vd.Init))
	} else {
		p.linef("var %s", toSnakeCase(vd.Name))
	}
}

// -------------------------------------------------------------------
// Block
// -------------------------------------------------------------------

func (p *printer) emitBlock(b *parser.Block) {
	if b == nil || len(b.Stmts) == 0 {
		p.line("pass")
		return
	}
	for _, s := range b.Stmts {
		p.emitStmt(s)
	}
}

// -------------------------------------------------------------------
// Statements
// -------------------------------------------------------------------

func (p *printer) emitStmt(s parser.Stmt) {
	if s == nil {
		return
	}
	switch v := s.(type) {

	case *parser.Block:
		p.emitBlock(v)

	case *parser.VarDecl:
		p.setSrc(v.Pos.Line)
		p.emitVarDecl(v)

	case *parser.ExprStmt:
		if v.X == nil {
			return
		}
		p.setSrc(v.X.ExprPos().Line)
		// Postfix ++ / -- as a standalone statement → x += 1 / x -= 1
		if post, ok := v.X.(*parser.PostfixExpr); ok {
			switch post.Op {
			case "++":
				p.linef("%s += 1", p.exprStr(post.X))
				return
			case "--":
				p.linef("%s -= 1", p.exprStr(post.X))
				return
			}
		}
		p.line(p.exprStr(v.X))

	case *parser.ReturnStmt:
		p.setSrc(v.Pos.Line)
		if v.Value != nil {
			p.linef("return %s", p.exprStr(v.Value))
		} else {
			p.line("return")
		}

	case *parser.IfStmt:
		p.setSrc(v.Pos.Line)
		p.emitIfStmt(v, "if")

	case *parser.WhileStmt:
		p.setSrc(v.Pos.Line)
		p.linef("while %s:", p.exprStr(v.Cond))
		p.push()
		p.emitBlock(v.Body)
		p.pop()

	case *parser.DoWhileStmt:
		p.setSrc(v.Pos.Line)
		// do { body } while (cond)  →  while true: body; if not cond: break
		p.line("while true:")
		p.push()
		p.emitBlock(v.Body)
		p.linef("if not (%s):", p.exprStr(v.Cond))
		p.push()
		p.line("break")
		p.pop()
		p.pop()

	case *parser.ForStmt:
		p.setSrc(v.Pos.Line)
		// C-style for → optional init + while loop (GDScript has no C-style for)
		if v.Init != nil {
			p.emitStmt(v.Init)
		}
		p.setSrc(v.Pos.Line) // reset after init emission
		cond := "true"
		if v.Cond != nil {
			cond = p.exprStr(v.Cond)
		}
		p.linef("while %s:", cond)
		p.push()
		if v.Body != nil && len(v.Body.Stmts) > 0 {
			p.emitBlock(v.Body)
		} else {
			p.line("pass")
		}
		if v.Post != nil {
			p.setSrc(v.Pos.Line)
			if post, ok := v.Post.(*parser.PostfixExpr); ok {
				switch post.Op {
				case "++":
					p.linef("%s += 1", p.exprStr(post.X))
				case "--":
					p.linef("%s -= 1", p.exprStr(post.X))
				default:
					p.line(p.exprStr(v.Post))
				}
			} else {
				p.line(p.exprStr(v.Post))
			}
		}
		p.pop()

	case *parser.SwitchStmt:
		p.setSrc(v.Pos.Line)
		p.linef("match %s:", p.exprStr(v.Tag))
		p.push()
		if len(v.Cases) == 0 {
			p.line("pass")
		}
		for _, cl := range v.Cases {
			p.setSrc(cl.Pos.Line)
			if cl.Value == nil {
				p.line("_:")
			} else {
				p.linef("%s:", p.exprStr(cl.Value))
			}
			p.push()
			// Skip break statements — GDScript match doesn't fall through.
			emitted := false
			for _, cs := range cl.Body {
				if _, isBreak := cs.(*parser.BreakStmt); isBreak {
					continue
				}
				p.emitStmt(cs)
				emitted = true
			}
			if !emitted {
				p.line("pass")
			}
			p.pop()
		}
		p.pop()

	case *parser.BreakStmt:
		p.setSrc(v.Pos.Line)
		p.line("break")

	case *parser.ContinueStmt:
		p.setSrc(v.Pos.Line)
		p.line("continue")
	}
}

func (p *printer) emitVarDecl(v *parser.VarDecl) {
	name := toSnakeCase(v.Name)
	if v.Type != "" {
		if v.Init != nil {
			p.linef("var %s: %s = %s", name, mapType(v.Type), p.exprStr(v.Init))
		} else {
			p.linef("var %s: %s", name, mapType(v.Type))
		}
	} else if v.Init != nil {
		p.linef("var %s = %s", name, p.exprStr(v.Init))
	} else {
		p.linef("var %s", name)
	}
}

// emitIfStmt emits an if / elif chain recursively.
func (p *printer) emitIfStmt(v *parser.IfStmt, keyword string) {
	p.setSrc(v.Pos.Line)
	p.linef("%s %s:", keyword, p.exprStr(v.Cond))
	p.push()
	p.emitBlock(v.Then)
	p.pop()
	if v.Else == nil {
		return
	}
	if elseIf, ok := v.Else.(*parser.IfStmt); ok {
		p.emitIfStmt(elseIf, "elif")
	} else {
		p.line("else:")
		p.push()
		p.emitStmt(v.Else)
		p.pop()
	}
}

// -------------------------------------------------------------------
// Expressions
// -------------------------------------------------------------------

// exprStr returns the GDScript representation of an expression.
func (p *printer) exprStr(e parser.Expr) string {
	if e == nil {
		return ""
	}
	switch v := e.(type) {

	case *parser.Literal:
		if v.Kind == "string" {
			// Scanner strips surrounding quotes from string lexemes; restore them.
			return `"` + v.Value + `"`
		}
		return v.Value // int, float, bool, null — already the right lexeme

	case *parser.Identifier:
		return toSnakeCase(v.Name)

	case *parser.GlobalExpr:
		// global.player → player  (global namespace is flat in the runtime)
		return toSnakeCase(v.Property)

	case *parser.AssignExpr:
		return p.exprStr(v.Target) + " " + v.Op + " " + p.exprStr(v.Value)

	case *parser.BinaryExpr:
		return p.exprStr(v.Left) + " " + mapBinaryOp(v.Op) + " " + p.exprStr(v.Right)

	case *parser.UnaryExpr:
		return mapUnaryOp(v.Op) + p.exprStr(v.X)

	case *parser.PostfixExpr:
		// In a pure expression context: emit inline (best-effort).
		// Statement context handles ++ / -- as += 1 / -= 1.
		switch v.Op {
		case "++":
			return p.exprStr(v.X) + " + 1"
		case "--":
			return p.exprStr(v.X) + " - 1"
		}
		return p.exprStr(v.X)

	case *parser.CallExpr:
		// T32: character built-in methods are rewritten to AGSRuntime calls.
		if recv, gdMethod, ok := characterBuiltinCallee(v.Callee); ok {
			charName := p.receiverName(recv)
			args := make([]string, len(v.Args))
			for i, a := range v.Args {
				args[i] = p.pointArgStr(a)
			}
			call := fmt.Sprintf(`AGSRuntime.get_character(%q).%s(%s)`, charName, gdMethod, strings.Join(args, ", "))
			if v.IsBlocking {
				return "await " + call
			}
			return call
		}
		callee := p.calleeStr(v.Callee)
		args := make([]string, len(v.Args))
		for i, arg := range v.Args {
			args[i] = p.exprStr(arg)
		}
		call := callee + "(" + strings.Join(args, ", ") + ")"
		if v.IsBlocking {
			return "await " + call
		}
		return call

	case *parser.MemberExpr:
		return p.exprStr(v.Object) + "." + toSnakeCase(v.Field)

	case *parser.IndexExpr:
		return p.exprStr(v.Object) + "[" + p.exprStr(v.Index) + "]"
	}
	return "null"
}

// calleeStr handles the callee of a call expression, applying snake_case to
// the final name while preserving the receiver chain intact.
func (p *printer) calleeStr(e parser.Expr) string {
	switch v := e.(type) {
	case *parser.Identifier:
		return toSnakeCase(v.Name)
	case *parser.MemberExpr:
		return p.exprStr(v.Object) + "." + toSnakeCase(v.Field)
	case *parser.GlobalExpr:
		return toSnakeCase(v.Property)
	}
	return p.exprStr(e)
}

// -------------------------------------------------------------------
// T32 — AGS-spirit built-in → AGSRuntime call mapping
// -------------------------------------------------------------------

// builtinCharacterMethods maps AGS-spirit PascalCase character method names
// to their GDScript snake_case counterparts on AGSCharacter.  Calls to these
// are rewritten from  obj.Method(arg)  to
//   AGSRuntime.get_character("obj").method("point_name")
var builtinCharacterMethods = map[string]string{
	"WalkTo": "walk_to",
	"FaceTo": "face_to",
}

// characterBuiltinCallee returns the receiver expression, the GDScript method
// name, and true when callee is a character built-in member call.
func characterBuiltinCallee(e parser.Expr) (recv parser.Expr, gdMethod string, ok bool) {
	m, isMember := e.(*parser.MemberExpr)
	if !isMember {
		return nil, "", false
	}
	gdMethod, ok = builtinCharacterMethods[m.Field]
	if !ok {
		return nil, "", false
	}
	return m.Object, gdMethod, true
}

// receiverName extracts the character name string from a receiver expression.
// global.player → "player", player → "player".
func (p *printer) receiverName(recv parser.Expr) string {
	switch v := recv.(type) {
	case *parser.GlobalExpr:
		return toSnakeCase(v.Property)
	case *parser.Identifier:
		return toSnakeCase(v.Name)
	}
	return p.exprStr(recv)
}

// pointArgStr extracts a point name string from a point argument.
// point.door → "door", point.door_left → "door_left".
// Any other expression falls back to the default expression renderer.
func (p *printer) pointArgStr(arg parser.Expr) string {
	if m, ok := arg.(*parser.MemberExpr); ok {
		if id, ok := m.Object.(*parser.Identifier); ok && id.Name == "point" {
			return fmt.Sprintf("%q", m.Field)
		}
	}
	return p.exprStr(arg)
}

// -------------------------------------------------------------------
// Utilities
// -------------------------------------------------------------------

// toSnakeCase converts PascalCase or camelCase to snake_case.
// Examples: WalkTo → walk_to, room_Load → room_load, getScore → get_score.
func toSnakeCase(s string) string {
	var buf strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && runes[i-1] != '_' {
				buf.WriteByte('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// mapType maps an AGS-spirit type name to its GDScript equivalent.
func mapType(t string) string {
	switch strings.ToLower(t) {
	case "int", "short", "char":
		return "int"
	case "float":
		return "float"
	case "bool":
		return "bool"
	case "string":
		return "String"
	case "void":
		return "void"
	}
	return t // custom types pass through unchanged
}

// mapBinaryOp maps AGS-spirit binary operators to GDScript equivalents.
func mapBinaryOp(op string) string {
	switch op {
	case "&&":
		return "and"
	case "||":
		return "or"
	}
	return op // +, -, *, /, %, ==, !=, <, <=, >, >=, &, |, ^, <<, >> unchanged
}

// mapUnaryOp maps AGS-spirit unary operators to GDScript equivalents.
func mapUnaryOp(op string) string {
	if op == "!" {
		return "not "
	}
	return op // -, ~ unchanged
}
