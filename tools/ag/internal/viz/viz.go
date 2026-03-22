// Package viz provides human-readable visualisers for each stage of the
// AGS-spirit → GDScript transpilation pipeline.
//
// Used by the `ag viz` subcommands to let authors and developers inspect
// what each stage produces. All visualisers write to an io.Writer so they
// can be directed to stdout, a file, or a test buffer.
//
// Stages and the tasks that implement them:
//
//	Tokens   — ag viz tokens   <file>  — VIZ-01 (T07)
//	AST      — ag viz ast      <file>  — VIZ-02 (T09)
//	Symbols  — ag viz symbols  <file>  — VIZ-T10 (T10)
//	Blocking — ag viz blocking <file>  — VIZ-03 (T11)
//	Emit     — ag viz emit     <file>  — VIZ-04 (T17)
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
// VIZ-02b — AST Graphviz DOT export
// --------------------------------------------------------------------

// ASTDot parses src and writes a Graphviz DOT graph to w.
// Render with: ag viz ast-dot file.agscript | dot -Tsvg -o ast.svg
//
// Node colour key:
//
//	#AED6F1  File root
//	#A9DFBF  Declarations  (FunctionDecl, NamespaceDecl, EnumDecl, TopVarDecl)
//	#F9E79F  Statements    (Block, IfStmt, WhileStmt, …)
//	#FDEBD0  Expressions   (BinaryExpr, CallExpr, Literal, …)
func ASTDot(w io.Writer, file, src string) {
	s := scanner.New(file, src)
	p := parser.New(s)
	f, errs := p.Parse(file)

	if len(errs) > 0 {
		fmt.Fprintf(w, "// parse errors in %s:\n", file)
		for _, e := range errs {
			fmt.Fprintf(w, "//   %v\n", e)
		}
		return
	}

	dg := &dotGen{w: w, id: 1}
	fmt.Fprintf(w, "digraph AST {\n")
	fmt.Fprintf(w, "  graph [fontname=\"Helvetica\" rankdir=\"TB\" label=%q fontsize=12]\n", "AST — "+file)
	fmt.Fprintf(w, "  node  [fontname=\"Helvetica\" shape=\"box\" style=\"rounded,filled\" fontsize=11]\n")
	fmt.Fprintf(w, "  edge  [fontname=\"Helvetica\" fontsize=9]\n\n")

	root := dg.node("File", "#AED6F1")
	for _, decl := range f.Decls {
		child := dg.emitDecl(decl)
		dg.edge(root, child, "")
	}
	fmt.Fprintf(w, "}\n")
}

// dotGen tracks node IDs and writes DOT statements to w.
type dotGen struct {
	w  io.Writer
	id int
}

func (d *dotGen) node(label, color string) int {
	id := d.id
	d.id++
	fmt.Fprintf(d.w, "  n%d [label=%q fillcolor=%q]\n", id, label, color)
	return id
}

func (d *dotGen) edge(from, to int, label string) {
	if label == "" {
		fmt.Fprintf(d.w, "  n%d -> n%d\n", from, to)
	} else {
		fmt.Fprintf(d.w, "  n%d -> n%d [label=%q fontcolor=\"#555555\"]\n", from, to, label)
	}
}

// emitDecl adds nodes for a top-level or namespace-member declaration.
func (d *dotGen) emitDecl(decl parser.Decl) int {
	const declColor = "#A9DFBF"
	switch v := decl.(type) {
	case *parser.FunctionDecl:
		ret := v.ReturnType
		if ret == "" {
			ret = "void"
		}
		flags := ""
		if v.IsExport {
			flags += " export"
		}
		if v.IsBlocking {
			flags += " blocking"
		}
		params := ""
		for i, p := range v.Params {
			if i > 0 {
				params += ", "
			}
			params += p.Type + " " + p.Name
		}
		label := fmt.Sprintf("FunctionDecl\n%s → %s%s\n(%s)", v.Name, ret, flags, params)
		n := d.node(label, declColor)
		if v.Body != nil {
			child := d.emitBlock(v.Body)
			d.edge(n, child, "body")
		}
		return n

	case *parser.NamespaceDecl:
		n := d.node("NamespaceDecl\n"+v.Name, declColor)
		for _, m := range v.Members {
			child := d.emitDecl(m)
			d.edge(n, child, "")
		}
		return n

	case *parser.EnumDecl:
		n := d.node("EnumDecl\n"+v.Name, declColor)
		for _, m := range v.Members {
			label := "EnumMember\n" + m.Name
			mn := d.node(label, "#D7BDE2")
			d.edge(n, mn, "")
			if m.Value != nil {
				child := d.emitExpr(m.Value)
				d.edge(mn, child, "=")
			}
		}
		return n

	case *parser.TopVarDecl:
		if v.Decl == nil {
			return d.node("TopVarDecl", declColor)
		}
		n := d.node(fmt.Sprintf("TopVarDecl\n%s: %s", v.Decl.Name, v.Decl.Type), declColor)
		if v.Decl.Init != nil {
			child := d.emitExpr(v.Decl.Init)
			d.edge(n, child, "=")
		}
		return n

	default:
		return d.node(fmt.Sprintf("%T", decl), declColor)
	}
}

// emitBlock adds a Block node and its child statement nodes.
func (d *dotGen) emitBlock(b *parser.Block) int {
	n := d.node(fmt.Sprintf("Block\n[%d:%d]", b.Pos.Line, b.Pos.Column), "#E8DAEF")
	for _, s := range b.Stmts {
		child := d.emitStmt(s)
		d.edge(n, child, "")
	}
	return n
}

// emitStmt adds nodes for a statement.
func (d *dotGen) emitStmt(stmt parser.Stmt) int {
	const stmtColor = "#F9E79F"
	if stmt == nil {
		return d.node("<nil>", "#CCCCCC")
	}
	switch s := stmt.(type) {
	case *parser.Block:
		return d.emitBlock(s)

	case *parser.VarDecl:
		n := d.node(fmt.Sprintf("VarDecl\n%s: %s", s.Name, s.Type), stmtColor)
		if s.Init != nil {
			child := d.emitExpr(s.Init)
			d.edge(n, child, "=")
		}
		return n

	case *parser.IfStmt:
		n := d.node(fmt.Sprintf("IfStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)
		cond := d.emitExpr(s.Cond)
		d.edge(n, cond, "cond")
		then := d.emitBlock(s.Then)
		d.edge(n, then, "then")
		if s.Else != nil {
			els := d.emitStmt(s.Else)
			d.edge(n, els, "else")
		}
		return n

	case *parser.WhileStmt:
		n := d.node(fmt.Sprintf("WhileStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)
		cond := d.emitExpr(s.Cond)
		d.edge(n, cond, "cond")
		body := d.emitBlock(s.Body)
		d.edge(n, body, "body")
		return n

	case *parser.DoWhileStmt:
		n := d.node(fmt.Sprintf("DoWhileStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)
		body := d.emitBlock(s.Body)
		d.edge(n, body, "body")
		cond := d.emitExpr(s.Cond)
		d.edge(n, cond, "cond")
		return n

	case *parser.ForStmt:
		n := d.node(fmt.Sprintf("ForStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)
		if s.Init != nil {
			child := d.emitStmt(s.Init)
			d.edge(n, child, "init")
		}
		if s.Cond != nil {
			child := d.emitExpr(s.Cond)
			d.edge(n, child, "cond")
		}
		if s.Post != nil {
			child := d.emitExpr(s.Post)
			d.edge(n, child, "post")
		}
		body := d.emitBlock(s.Body)
		d.edge(n, body, "body")
		return n

	case *parser.SwitchStmt:
		n := d.node(fmt.Sprintf("SwitchStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)
		tag := d.emitExpr(s.Tag)
		d.edge(n, tag, "tag")
		for _, c := range s.Cases {
			var cn int
			if c.Value == nil {
				cn = d.node("Default", stmtColor)
			} else {
				cn = d.node(fmt.Sprintf("Case\n[%d:%d]", c.Pos.Line, c.Pos.Column), stmtColor)
				val := d.emitExpr(c.Value)
				d.edge(cn, val, "val")
			}
			d.edge(n, cn, "")
			for _, cs := range c.Body {
				child := d.emitStmt(cs)
				d.edge(cn, child, "")
			}
		}
		return n

	case *parser.ReturnStmt:
		n := d.node(fmt.Sprintf("ReturnStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)
		if s.Value != nil {
			child := d.emitExpr(s.Value)
			d.edge(n, child, "")
		}
		return n

	case *parser.BreakStmt:
		return d.node(fmt.Sprintf("BreakStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)

	case *parser.ContinueStmt:
		return d.node(fmt.Sprintf("ContinueStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)

	case *parser.ExprStmt:
		n := d.node(fmt.Sprintf("ExprStmt\n[%d:%d]", s.Pos.Line, s.Pos.Column), stmtColor)
		if s.X != nil {
			child := d.emitExpr(s.X)
			d.edge(n, child, "")
		}
		return n

	default:
		return d.node(fmt.Sprintf("%T", stmt), stmtColor)
	}
}

// emitExpr adds nodes for an expression.
func (d *dotGen) emitExpr(expr parser.Expr) int {
	const exprColor = "#FDEBD0"
	if expr == nil {
		return d.node("<nil>", "#CCCCCC")
	}
	switch e := expr.(type) {
	case *parser.Literal:
		return d.node(fmt.Sprintf("Literal(%s)\n%q", e.Kind, e.Value), exprColor)

	case *parser.Identifier:
		return d.node("Ident\n"+e.Name, exprColor)

	case *parser.GlobalExpr:
		return d.node("Global\n."+e.Property, exprColor)

	case *parser.AssignExpr:
		n := d.node("AssignExpr\n"+e.Op, exprColor)
		target := d.emitExpr(e.Target)
		val := d.emitExpr(e.Value)
		d.edge(n, target, "target")
		d.edge(n, val, "value")
		return n

	case *parser.BinaryExpr:
		n := d.node("BinaryExpr\n"+e.Op, exprColor)
		left := d.emitExpr(e.Left)
		right := d.emitExpr(e.Right)
		d.edge(n, left, "L")
		d.edge(n, right, "R")
		return n

	case *parser.UnaryExpr:
		n := d.node("UnaryExpr\n"+e.Op, exprColor)
		child := d.emitExpr(e.X)
		d.edge(n, child, "")
		return n

	case *parser.PostfixExpr:
		n := d.node("PostfixExpr\n"+e.Op, exprColor)
		child := d.emitExpr(e.X)
		d.edge(n, child, "")
		return n

	case *parser.CallExpr:
		n := d.node(fmt.Sprintf("CallExpr\n[%d:%d]", e.Pos.Line, e.Pos.Column), exprColor)
		callee := d.emitExpr(e.Callee)
		d.edge(n, callee, "callee")
		for i, arg := range e.Args {
			child := d.emitExpr(arg)
			d.edge(n, child, fmt.Sprintf("arg%d", i))
		}
		return n

	case *parser.MemberExpr:
		n := d.node("MemberExpr\n."+e.Field, exprColor)
		obj := d.emitExpr(e.Object)
		d.edge(n, obj, "obj")
		return n

	case *parser.IndexExpr:
		n := d.node(fmt.Sprintf("IndexExpr\n[%d:%d]", e.Pos.Line, e.Pos.Column), exprColor)
		obj := d.emitExpr(e.Object)
		idx := d.emitExpr(e.Index)
		d.edge(n, obj, "obj")
		d.edge(n, idx, "idx")
		return n

	default:
		return d.node(fmt.Sprintf("%T", expr), exprColor)
	}
}

// --------------------------------------------------------------------
// VIZ-T10 — Symbol table (text)
// --------------------------------------------------------------------

// Symbols parses src, builds the symbol table, and writes a formatted
// table of all collected symbols to w.
//
// Example output:
//
//	Symbols — rooms/market/market.agscript
//	Globals (3 symbols)
//	  function  room_Load          → void          [1:1]
//	  var       score              : int           [3:1]
//	  enum      Direction                          [5:1]
//	    member  North              : Direction
//	    member  South              : Direction
//	Namespace MathUtils (2 members)
//	  function  square(int x)      → int  [export] [10:5]
//	  function  clamp(int v, …)    → int  [export] [15:5]
//	2 diagnostic(s): 0 error(s), 2 warning(s)
func Symbols(w io.Writer, file, src string) {
	s := scanner.New(file, src)
	p := parser.New(s)
	f, parseErrs := p.Parse(file)

	fmt.Fprintf(w, "Symbols — %s\n", file)

	if len(parseErrs) > 0 {
		for _, e := range parseErrs {
			fmt.Fprintf(w, "  parse error: %v\n", e)
		}
		return
	}

	st, diags := parser.BuildSymbolTable(f)

	// ---- Globals ----
	globals := symbolsSorted(st.Globals)
	fmt.Fprintf(w, "\nGlobals (%d symbol(s))\n", len(globals))
	for _, sym := range globals {
		if sym.Kind == parser.KindNamespace {
			continue // printed in the Namespaces section
		}
		printSymbolLine(w, "  ", sym)
		// Inline enum members indented under the enum.
		if sym.Kind == parser.KindEnum {
			for _, msym := range symbolsSorted(st.Globals) {
				if msym.Kind == parser.KindEnumMember && msym.Type == sym.Name {
					fmt.Fprintf(w, "      member  %-20s : %s\n", msym.Name, msym.Type)
				}
			}
		}
	}

	// ---- Namespaces ----
	for _, nsName := range sortedKeys(st.Namespaces) {
		members := st.Namespaces[nsName]
		fmt.Fprintf(w, "\nNamespace %s (%d member(s))\n", nsName, len(members))
		for _, sym := range symbolsSorted(members) {
			printSymbolLine(w, "  ", sym)
		}
	}

	// ---- Diagnostics summary ----
	errCount, warnCount := 0, 0
	for _, d := range diags {
		if d.Severity == "error" {
			errCount++
		} else {
			warnCount++
		}
	}
	fmt.Fprintf(w, "\n%d diagnostic(s): %d error(s), %d warning(s)\n",
		len(diags), errCount, warnCount)
	for _, d := range diags {
		if d.Severity == "error" {
			fmt.Fprintf(w, "  error:   %s\n", d.Message)
		}
	}
}

// printSymbolLine writes one formatted symbol row.
func printSymbolLine(w io.Writer, indent string, sym *parser.Symbol) {
	flags := ""
	if sym.IsExport {
		flags += " [export]"
	}
	if sym.IsBlocking {
		flags += " [blocking]"
	}

	switch sym.Kind {
	case parser.KindFunction:
		params := formatParams(sym.Params)
		ret := sym.Type
		if ret == "" {
			ret = "void"
		}
		fmt.Fprintf(w, "%s%-10s %-22s → %-10s%s [%d:%d]\n",
			indent, "function", sym.Name+"("+params+")", ret, flags,
			sym.Decl.Line, sym.Decl.Column)
	case parser.KindVar:
		fmt.Fprintf(w, "%s%-10s %-22s : %-10s%s [%d:%d]\n",
			indent, "var", sym.Name, sym.Type, flags,
			sym.Decl.Line, sym.Decl.Column)
	case parser.KindEnum:
		fmt.Fprintf(w, "%s%-10s %-22s%s [%d:%d]\n",
			indent, "enum", sym.Name, flags,
			sym.Decl.Line, sym.Decl.Column)
	case parser.KindEnumMember:
		// printed inline under the owning enum; skip here
	case parser.KindNamespace:
		// printed as its own section; skip here
	default:
		fmt.Fprintf(w, "%s%-10s %s\n", indent, sym.Kind.String(), sym.Name)
	}
}

func formatParams(params []parser.Param) string {
	if len(params) == 0 {
		return ""
	}
	s := ""
	for i, p := range params {
		if i > 0 {
			s += ", "
		}
		s += p.Type + " " + p.Name
	}
	return s
}

// --------------------------------------------------------------------
// VIZ-T10b — Symbol table (Graphviz DOT)
// --------------------------------------------------------------------

// SymbolsDot parses src, builds the symbol table, and writes a Graphviz
// DOT graph to w. Namespaces are rendered as clusters; enum members are
// nested inside their enum node.
// Render with: ag viz symbols-dot file.agscript | dot -Tsvg -o symbols.svg
func SymbolsDot(w io.Writer, file, src string) {
	s := scanner.New(file, src)
	p := parser.New(s)
	f, parseErrs := p.Parse(file)

	if len(parseErrs) > 0 {
		fmt.Fprintf(w, "// parse errors in %s:\n", file)
		for _, e := range parseErrs {
			fmt.Fprintf(w, "//   %v\n", e)
		}
		return
	}

	st, _ := parser.BuildSymbolTable(f)

	fmt.Fprintf(w, "digraph Symbols {\n")
	fmt.Fprintf(w, "  graph [fontname=\"Helvetica\" rankdir=\"LR\" label=%q fontsize=13 compound=true]\n",
		"Symbols — "+file)
	fmt.Fprintf(w, "  node  [fontname=\"Helvetica\" shape=\"box\" style=\"rounded,filled\" fontsize=11]\n")
	fmt.Fprintf(w, "  edge  [fontname=\"Helvetica\" fontsize=9 style=\"dashed\" color=\"#888888\"]\n\n")

	id := 1
	nextID := func() int { i := id; id++; return i }

	// File-scope globals cluster.
	fmt.Fprintf(w, "  subgraph cluster_globals {\n")
	fmt.Fprintf(w, "    label=\"File scope\" style=\"filled\" fillcolor=\"#EBF5FB\" fontsize=12\n")

	globalAnchor := nextID()
	fmt.Fprintf(w, "    n%d [label=\"\" shape=point width=0]\n", globalAnchor)

	for _, sym := range symbolsSorted(st.Globals) {
		if sym.Kind == parser.KindNamespace {
			continue
		}
		nid := nextID()
		label, color := symDotLabel(sym)
		fmt.Fprintf(w, "    n%d [label=%q fillcolor=%q]\n", nid, label, color)

		// Enum members as child nodes inside their own mini-cluster.
		if sym.Kind == parser.KindEnum {
			fmt.Fprintf(w, "    subgraph cluster_enum_%d {\n", nid)
			fmt.Fprintf(w, "      label=%q style=\"filled\" fillcolor=\"#F9F3FF\" fontsize=10\n", sym.Name+" members")
			for _, msym := range symbolsSorted(st.Globals) {
				if msym.Kind == parser.KindEnumMember && msym.Type == sym.Name {
					mid := nextID()
					fmt.Fprintf(w, "      n%d [label=%q fillcolor=\"#E8DAEF\"]\n", mid, msym.Name)
				}
			}
			fmt.Fprintf(w, "    }\n")
		}
	}
	fmt.Fprintf(w, "  }\n\n")

	// One cluster per namespace.
	for _, nsName := range sortedKeys(st.Namespaces) {
		members := st.Namespaces[nsName]
		fmt.Fprintf(w, "  subgraph cluster_ns_%s {\n", nsName)
		fmt.Fprintf(w, "    label=%q style=\"filled\" fillcolor=\"#EAFAF1\" fontsize=12\n",
			"namespace "+nsName)
		for _, sym := range symbolsSorted(members) {
			nid := nextID()
			label, color := symDotLabel(sym)
			fmt.Fprintf(w, "    n%d [label=%q fillcolor=%q]\n", nid, label, color)
		}
		fmt.Fprintf(w, "  }\n\n")
	}

	fmt.Fprintf(w, "}\n")
}

// symDotLabel returns a DOT label and fill colour for a symbol.
func symDotLabel(sym *parser.Symbol) (label, color string) {
	flags := ""
	if sym.IsExport {
		flags = " ●export"
	}
	switch sym.Kind {
	case parser.KindFunction:
		ret := sym.Type
		if ret == "" {
			ret = "void"
		}
		params := formatParams(sym.Params)
		return fmt.Sprintf("fn %s(%s)\n→ %s%s", sym.Name, params, ret, flags), "#A9DFBF"
	case parser.KindVar:
		return fmt.Sprintf("var %s\n: %s", sym.Name, sym.Type), "#AED6F1"
	case parser.KindEnum:
		return fmt.Sprintf("enum %s%s", sym.Name, flags), "#D7BDE2"
	case parser.KindEnumMember:
		return sym.Name, "#E8DAEF"
	}
	return sym.Name, "#F0F0F0"
}

// -------------------------------------------------------------------
// Sorting helpers
// -------------------------------------------------------------------

func symbolsSorted(m map[string]*parser.Symbol) []*parser.Symbol {
	keys := sortedKeys2(m)
	out := make([]*parser.Symbol, 0, len(keys))
	for _, k := range keys {
		out = append(out, m[k])
	}
	return out
}

func sortedKeys(m map[string]map[string]*parser.Symbol) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

func sortedKeys2(m map[string]*parser.Symbol) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// --------------------------------------------------------------------
// VIZ-03 — Blocking call annotation
// --------------------------------------------------------------------

// Blocking parses src, runs symbol resolution (T10) and blocking annotation
// (T11), then writes a table of:
//   - each function and whether it is blocking
//   - each blocking call site within blocking functions (line, callee)
//
// Example output:
//
//	Blocking — rooms/market/market.agscript
//	Functions (4)
//	  [blocking] walkAndGreet         — 2 blocking call(s)
//	      L8   global.player.WalkTo(…)
//	      L9   global.player.Say(…)
//	  [blocking] room_AfterFadeIn     — 1 blocking call(s)
//	      L13  walkAndGreet(…)
//	  [-]        room_Load            — 0 blocking call(s)
func Blocking(w io.Writer, file, src string) {
	s := scanner.New(file, src)
	p := parser.New(s)
	f, parseErrs := p.Parse(file)

	fmt.Fprintf(w, "Blocking — %s\n", file)

	if len(parseErrs) > 0 {
		for _, e := range parseErrs {
			fmt.Fprintf(w, "  parse error: %v\n", e)
		}
		return
	}

	st, _ := parser.BuildSymbolTable(f)
	parser.AnnotateBlocking(f, st)

	// Collect top-level and namespace function decls.
	var funcs []*parser.FunctionDecl
	for _, d := range f.Decls {
		switch v := d.(type) {
		case *parser.FunctionDecl:
			funcs = append(funcs, v)
		case *parser.NamespaceDecl:
			for _, m := range v.Members {
				if fd, ok := m.(*parser.FunctionDecl); ok {
					funcs = append(funcs, fd)
				}
			}
		}
	}

	fmt.Fprintf(w, "Functions (%d)\n", len(funcs))
	for _, fd := range funcs {
		sites := collectBlockingCalls(fd)
		flag := "[-]       "
		if fd.IsBlocking {
			flag = "[blocking]"
		}
		fmt.Fprintf(w, "  %s %-24s — %d blocking call(s)\n", flag, fd.Name, len(sites))
		for _, site := range sites {
			fmt.Fprintf(w, "      L%-4d %s\n", site.line, site.callee)
		}
	}
}

// blockingCallSite records a single blocking call expression for display.
type blockingCallSite struct {
	line   int
	callee string
}

// collectBlockingCalls walks fd.Body and returns all CallExprs with IsBlocking=true.
func collectBlockingCalls(fd *parser.FunctionDecl) []blockingCallSite {
	if fd.Body == nil {
		return nil
	}
	var out []blockingCallSite
	collectBlockingInBlock(fd.Body, &out)
	return out
}

func collectBlockingInBlock(b *parser.Block, out *[]blockingCallSite) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		collectBlockingInStmt(s, out)
	}
}

func collectBlockingInStmt(s parser.Stmt, out *[]blockingCallSite) {
	if s == nil {
		return
	}
	switch v := s.(type) {
	case *parser.Block:
		collectBlockingInBlock(v, out)
	case *parser.VarDecl:
		collectBlockingInExpr(v.Init, out)
	case *parser.ExprStmt:
		collectBlockingInExpr(v.X, out)
	case *parser.ReturnStmt:
		collectBlockingInExpr(v.Value, out)
	case *parser.IfStmt:
		collectBlockingInExpr(v.Cond, out)
		collectBlockingInBlock(v.Then, out)
		collectBlockingInStmt(v.Else, out)
	case *parser.WhileStmt:
		collectBlockingInExpr(v.Cond, out)
		collectBlockingInBlock(v.Body, out)
	case *parser.DoWhileStmt:
		collectBlockingInBlock(v.Body, out)
		collectBlockingInExpr(v.Cond, out)
	case *parser.ForStmt:
		collectBlockingInStmt(v.Init, out)
		collectBlockingInExpr(v.Cond, out)
		collectBlockingInExpr(v.Post, out)
		collectBlockingInBlock(v.Body, out)
	case *parser.SwitchStmt:
		collectBlockingInExpr(v.Tag, out)
		for _, cl := range v.Cases {
			collectBlockingInExpr(cl.Value, out)
			for _, cs := range cl.Body {
				collectBlockingInStmt(cs, out)
			}
		}
	}
}

func collectBlockingInExpr(e parser.Expr, out *[]blockingCallSite) {
	if e == nil {
		return
	}
	switch v := e.(type) {
	case *parser.CallExpr:
		if v.IsBlocking {
			*out = append(*out, blockingCallSite{
				line:   v.Pos.Line,
				callee: calleeString(v.Callee) + "(…)",
			})
		}
		collectBlockingInExpr(v.Callee, out)
		for _, arg := range v.Args {
			collectBlockingInExpr(arg, out)
		}
	case *parser.AssignExpr:
		collectBlockingInExpr(v.Target, out)
		collectBlockingInExpr(v.Value, out)
	case *parser.BinaryExpr:
		collectBlockingInExpr(v.Left, out)
		collectBlockingInExpr(v.Right, out)
	case *parser.UnaryExpr:
		collectBlockingInExpr(v.X, out)
	case *parser.PostfixExpr:
		collectBlockingInExpr(v.X, out)
	case *parser.MemberExpr:
		collectBlockingInExpr(v.Object, out)
	case *parser.IndexExpr:
		collectBlockingInExpr(v.Object, out)
		collectBlockingInExpr(v.Index, out)
	}
}

// calleeString returns a short human-readable string for a callee expression.
func calleeString(e parser.Expr) string {
	if e == nil {
		return "?"
	}
	switch v := e.(type) {
	case *parser.Identifier:
		return v.Name
	case *parser.MemberExpr:
		return calleeString(v.Object) + "." + v.Field
	case *parser.GlobalExpr:
		return "global." + v.Property
	}
	return "?"
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
