// Tests for the AGS-spirit parser and AST.
// T08: AST node construction and interface satisfaction.
// T09+: fixture-driven parsing tests added after the parser is implemented.
package parser_test

import (
	"testing"

	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func newScanner(src string) *scanner.Scanner {
	return scanner.New("test.agscript", src)
}

func parse(src string) (*parser.File, []*parser.ParseError) {
	s := newScanner(src)
	p := parser.New(s)
	return p.Parse("test.agscript")
}

func mustParse(t *testing.T, src string) *parser.File {
	t.Helper()
	f, errs := parse(src)
	if len(errs) != 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	return f
}

func assertErrors(t *testing.T, errs []*parser.ParseError, want int) {
	t.Helper()
	if len(errs) != want {
		t.Errorf("got %d errors, want %d", len(errs), want)
		for _, e := range errs {
			t.Logf("  %v", e)
		}
	}
}

func tok(line, col int, kind scanner.TokenKind, lexeme string) scanner.Token {
	return scanner.Token{Kind: kind, Lexeme: lexeme, Line: line, Column: col}
}

// -------------------------------------------------------------------
// Parser stub smoke tests (T08 — stub returns empty AST)
// -------------------------------------------------------------------

func TestParser_EmptyFileProducesNoDecls(t *testing.T) {
	s := scanner.New("empty.agscript", "")
	p := parser.New(s)
	f, errs := p.Parse("empty.agscript")
	if len(errs) != 0 {
		t.Errorf("unexpected errors on empty input: %v", errs)
	}
	if len(f.Decls) != 0 {
		t.Errorf("expected 0 decls on empty input, got %d", len(f.Decls))
	}
}

func TestParser_FilePathPreserved(t *testing.T) {
	s := scanner.New("rooms/market.agscript", "")
	p := parser.New(s)
	f, _ := p.Parse("rooms/market.agscript")
	if f.Path != "rooms/market.agscript" {
		t.Errorf("File.Path = %q, want rooms/market.agscript", f.Path)
	}
}

func TestParser_NoErrors_OnEmptyInput(t *testing.T) {
	_, errs := parse("")
	assertErrors(t, errs, 0)
}

// -------------------------------------------------------------------
// ParseError
// -------------------------------------------------------------------

func TestParseError_ErrorString(t *testing.T) {
	e := &parser.ParseError{
		File:    "rooms/hall.agscript",
		Line:    10,
		Column:  5,
		Message: "unexpected token",
	}
	got := e.Error()
	want := "rooms/hall.agscript:10:5: unexpected token"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// -------------------------------------------------------------------
// Interface satisfaction — Decl
// -------------------------------------------------------------------

func TestDecl_FunctionDeclSatisfiesDecl(t *testing.T) {
	var _ parser.Decl = (*parser.FunctionDecl)(nil)
}

func TestDecl_NamespaceDeclSatisfiesDecl(t *testing.T) {
	var _ parser.Decl = (*parser.NamespaceDecl)(nil)
}

func TestDecl_EnumDeclSatisfiesDecl(t *testing.T) {
	var _ parser.Decl = (*parser.EnumDecl)(nil)
}

func TestDecl_TopVarDeclSatisfiesDecl(t *testing.T) {
	var _ parser.Decl = (*parser.TopVarDecl)(nil)
}

// -------------------------------------------------------------------
// Interface satisfaction — Stmt
// -------------------------------------------------------------------

func TestStmt_BlockSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.Block)(nil)
}

func TestStmt_VarDeclSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.VarDecl)(nil)
}

func TestStmt_IfStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.IfStmt)(nil)
}

func TestStmt_WhileStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.WhileStmt)(nil)
}

func TestStmt_DoWhileStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.DoWhileStmt)(nil)
}

func TestStmt_ForStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.ForStmt)(nil)
}

func TestStmt_SwitchStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.SwitchStmt)(nil)
}

func TestStmt_ReturnStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.ReturnStmt)(nil)
}

func TestStmt_BreakStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.BreakStmt)(nil)
}

func TestStmt_ContinueStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.ContinueStmt)(nil)
}

func TestStmt_ExprStmtSatisfiesStmt(t *testing.T) {
	var _ parser.Stmt = (*parser.ExprStmt)(nil)
}

// -------------------------------------------------------------------
// Interface satisfaction — Expr
// -------------------------------------------------------------------

func TestExpr_AssignExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.AssignExpr)(nil)
}

func TestExpr_BinaryExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.BinaryExpr)(nil)
}

func TestExpr_UnaryExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.UnaryExpr)(nil)
}

func TestExpr_PostfixExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.PostfixExpr)(nil)
}

func TestExpr_CallExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.CallExpr)(nil)
}

func TestExpr_IndexExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.IndexExpr)(nil)
}

func TestExpr_MemberExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.MemberExpr)(nil)
}

func TestExpr_GlobalExprSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.GlobalExpr)(nil)
}

func TestExpr_IdentifierSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.Identifier)(nil)
}

func TestExpr_LiteralSatisfiesExpr(t *testing.T) {
	var _ parser.Expr = (*parser.Literal)(nil)
}

// -------------------------------------------------------------------
// Node construction — Declarations
// -------------------------------------------------------------------

func TestFunctionDecl_Fields(t *testing.T) {
	pos := tok(1, 1, scanner.TokenFunction, "function")
	fn := &parser.FunctionDecl{
		ReturnType: "int",
		Name:       "getScore",
		Params: []parser.Param{
			{Type: "int", Name: "bonus", Pos: tok(1, 10, scanner.TokenIdent, "bonus")},
		},
		Body:       &parser.Block{},
		IsExport:   true,
		IsBlocking: false,
		Pos:        pos,
	}
	if fn.ReturnType != "int" {
		t.Errorf("ReturnType = %q, want \"int\"", fn.ReturnType)
	}
	if fn.Name != "getScore" {
		t.Errorf("Name = %q, want \"getScore\"", fn.Name)
	}
	if len(fn.Params) != 1 {
		t.Fatalf("len(Params) = %d, want 1", len(fn.Params))
	}
	if fn.Params[0].Type != "int" || fn.Params[0].Name != "bonus" {
		t.Errorf("Param = {%q %q}, want {int bonus}", fn.Params[0].Type, fn.Params[0].Name)
	}
	if !fn.IsExport {
		t.Error("IsExport should be true")
	}
	if fn.IsBlocking {
		t.Error("IsBlocking should be false")
	}
	if fn.DeclPos() != pos {
		t.Error("DeclPos() mismatch")
	}
}

func TestFunctionDecl_VoidReturnType(t *testing.T) {
	fn := &parser.FunctionDecl{ReturnType: "", Name: "room_Load"}
	if fn.ReturnType != "" {
		t.Errorf("void function should have empty ReturnType, got %q", fn.ReturnType)
	}
}

func TestNamespaceDecl_Fields(t *testing.T) {
	pos := tok(1, 1, scanner.TokenNamespace, "namespace")
	inner := &parser.FunctionDecl{Name: "helper", Pos: tok(2, 5, scanner.TokenFunction, "function")}
	ns := &parser.NamespaceDecl{
		Name:    "CharUtils",
		Members: []parser.Decl{inner},
		Pos:     pos,
	}
	if ns.Name != "CharUtils" {
		t.Errorf("Name = %q, want \"CharUtils\"", ns.Name)
	}
	if len(ns.Members) != 1 {
		t.Fatalf("len(Members) = %d, want 1", len(ns.Members))
	}
	if ns.DeclPos() != pos {
		t.Error("DeclPos() mismatch")
	}
}

func TestEnumDecl_Fields(t *testing.T) {
	pos := tok(1, 1, scanner.TokenEnum, "enum")
	ed := &parser.EnumDecl{
		Name: "Direction",
		Members: []parser.EnumMember{
			{Name: "North", Value: nil, Pos: tok(2, 3, scanner.TokenIdent, "North")},
			{Name: "South", Value: &parser.Literal{Kind: "int", Value: "1"}, Pos: tok(3, 3, scanner.TokenIdent, "South")},
		},
		Pos: pos,
	}
	if ed.Name != "Direction" {
		t.Errorf("Name = %q, want \"Direction\"", ed.Name)
	}
	if len(ed.Members) != 2 {
		t.Fatalf("len(Members) = %d, want 2", len(ed.Members))
	}
	if ed.Members[0].Value != nil {
		t.Error("first member should have nil Value (implicit)")
	}
	if ed.Members[1].Value == nil {
		t.Error("second member should have non-nil Value")
	}
	if ed.DeclPos() != pos {
		t.Error("DeclPos() mismatch")
	}
}

func TestTopVarDecl_Fields(t *testing.T) {
	inner := &parser.VarDecl{
		Type: "int",
		Name: "score",
		Pos:  tok(1, 1, scanner.TokenIdent, "int"),
	}
	tv := &parser.TopVarDecl{Decl: inner}
	if tv.Decl != inner {
		t.Error("Decl field mismatch")
	}
	if tv.DeclPos() != inner.Pos {
		t.Error("DeclPos() should delegate to inner VarDecl")
	}
}

// -------------------------------------------------------------------
// Node construction — Statements
// -------------------------------------------------------------------

func TestBlock_Empty(t *testing.T) {
	b := &parser.Block{Stmts: nil}
	if len(b.Stmts) != 0 {
		t.Errorf("expected 0 stmts, got %d", len(b.Stmts))
	}
}

func TestBlock_WithStmts(t *testing.T) {
	s1 := &parser.BreakStmt{}
	s2 := &parser.ContinueStmt{}
	b := &parser.Block{Stmts: []parser.Stmt{s1, s2}}
	if len(b.Stmts) != 2 {
		t.Errorf("expected 2 stmts, got %d", len(b.Stmts))
	}
}

func TestVarDecl_WithInit(t *testing.T) {
	init := &parser.Literal{Kind: "int", Value: "42"}
	vd := &parser.VarDecl{Type: "int", Name: "x", Init: init}
	if vd.Type != "int" || vd.Name != "x" {
		t.Errorf("VarDecl fields wrong: %q %q", vd.Type, vd.Name)
	}
	if vd.Init != init {
		t.Error("Init mismatch")
	}
}

func TestVarDecl_WithoutInit(t *testing.T) {
	vd := &parser.VarDecl{Type: "bool", Name: "flag", Init: nil}
	if vd.Init != nil {
		t.Error("Init should be nil when no initialiser")
	}
}

func TestIfStmt_ThenOnly(t *testing.T) {
	cond := &parser.Literal{Kind: "bool", Value: "true"}
	then := &parser.Block{}
	s := &parser.IfStmt{Cond: cond, Then: then, Else: nil}
	if s.Else != nil {
		t.Error("Else should be nil")
	}
}

func TestIfStmt_WithElseBlock(t *testing.T) {
	cond := &parser.Literal{Kind: "bool", Value: "true"}
	then := &parser.Block{}
	els := &parser.Block{}
	s := &parser.IfStmt{Cond: cond, Then: then, Else: els}
	_, ok := s.Else.(*parser.Block)
	if !ok {
		t.Error("Else should be *Block")
	}
}

func TestIfStmt_WithElseIf(t *testing.T) {
	inner := &parser.IfStmt{
		Cond: &parser.Literal{Kind: "bool", Value: "false"},
		Then: &parser.Block{},
		Else: nil,
	}
	outer := &parser.IfStmt{
		Cond: &parser.Literal{Kind: "bool", Value: "true"},
		Then: &parser.Block{},
		Else: inner,
	}
	_, ok := outer.Else.(*parser.IfStmt)
	if !ok {
		t.Error("Else should be *IfStmt for else-if chain")
	}
}

func TestWhileStmt_Fields(t *testing.T) {
	cond := &parser.Literal{Kind: "bool", Value: "true"}
	body := &parser.Block{}
	s := &parser.WhileStmt{Cond: cond, Body: body}
	if s.Cond != cond || s.Body != body {
		t.Error("WhileStmt field mismatch")
	}
}

func TestDoWhileStmt_Fields(t *testing.T) {
	cond := &parser.Literal{Kind: "bool", Value: "true"}
	body := &parser.Block{}
	s := &parser.DoWhileStmt{Body: body, Cond: cond}
	if s.Body != body || s.Cond != cond {
		t.Error("DoWhileStmt field mismatch")
	}
}

func TestForStmt_AllFields(t *testing.T) {
	init := &parser.VarDecl{Type: "int", Name: "i"}
	cond := &parser.BinaryExpr{Op: "<"}
	post := &parser.PostfixExpr{Op: "++"}
	body := &parser.Block{}
	s := &parser.ForStmt{Init: init, Cond: cond, Post: post, Body: body}
	if s.Init != init {
		t.Error("Init mismatch")
	}
	if s.Cond != cond {
		t.Error("Cond mismatch")
	}
	if s.Post != post {
		t.Error("Post mismatch")
	}
}

func TestForStmt_NilFields(t *testing.T) {
	// Infinite loop: for (;;)
	s := &parser.ForStmt{Init: nil, Cond: nil, Post: nil, Body: &parser.Block{}}
	if s.Init != nil || s.Cond != nil || s.Post != nil {
		t.Error("nil for-loop fields should be nil")
	}
}

func TestSwitchStmt_Fields(t *testing.T) {
	tag := &parser.Identifier{Name: "dir"}
	c1 := &parser.CaseClause{
		Value: &parser.Literal{Kind: "int", Value: "0"},
		Body:  []parser.Stmt{&parser.BreakStmt{}},
	}
	c2 := &parser.CaseClause{
		Value: nil, // default
		Body:  []parser.Stmt{},
	}
	s := &parser.SwitchStmt{Tag: tag, Cases: []*parser.CaseClause{c1, c2}}
	if len(s.Cases) != 2 {
		t.Fatalf("len(Cases) = %d, want 2", len(s.Cases))
	}
	if s.Cases[1].Value != nil {
		t.Error("default case should have nil Value")
	}
}

func TestReturnStmt_WithValue(t *testing.T) {
	val := &parser.Literal{Kind: "int", Value: "0"}
	s := &parser.ReturnStmt{Value: val}
	if s.Value != val {
		t.Error("Value mismatch")
	}
}

func TestReturnStmt_BareReturn(t *testing.T) {
	s := &parser.ReturnStmt{Value: nil}
	if s.Value != nil {
		t.Error("bare return should have nil Value")
	}
}

func TestBreakStmt_Construction(t *testing.T) {
	pos := tok(5, 3, scanner.TokenBreak, "break")
	s := &parser.BreakStmt{Pos: pos}
	if s.Pos != pos {
		t.Error("Pos mismatch")
	}
}

func TestContinueStmt_Construction(t *testing.T) {
	pos := tok(6, 3, scanner.TokenContinue, "continue")
	s := &parser.ContinueStmt{Pos: pos}
	if s.Pos != pos {
		t.Error("Pos mismatch")
	}
}

func TestExprStmt_Construction(t *testing.T) {
	x := &parser.CallExpr{Callee: &parser.Identifier{Name: "doThing"}}
	s := &parser.ExprStmt{X: x}
	if s.X != x {
		t.Error("X mismatch")
	}
}

// -------------------------------------------------------------------
// Node construction — Expressions
// -------------------------------------------------------------------

func TestAssignExpr_Fields(t *testing.T) {
	pos := tok(3, 5, scanner.TokenAssign, "=")
	target := &parser.Identifier{Name: "x"}
	value := &parser.Literal{Kind: "int", Value: "10"}
	e := &parser.AssignExpr{Op: "=", Target: target, Value: value, Pos: pos}
	if e.Op != "=" || e.Target != target || e.Value != value {
		t.Error("AssignExpr field mismatch")
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestAssignExpr_CompoundOps(t *testing.T) {
	ops := []string{"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>="}
	for _, op := range ops {
		e := &parser.AssignExpr{Op: op}
		if e.Op != op {
			t.Errorf("Op = %q, want %q", e.Op, op)
		}
	}
}

func TestBinaryExpr_Fields(t *testing.T) {
	pos := tok(2, 8, scanner.TokenPlus, "+")
	left := &parser.Literal{Kind: "int", Value: "1"}
	right := &parser.Literal{Kind: "int", Value: "2"}
	e := &parser.BinaryExpr{Op: "+", Left: left, Right: right, Pos: pos}
	if e.Op != "+" || e.Left != left || e.Right != right {
		t.Error("BinaryExpr field mismatch")
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestBinaryExpr_AllOps(t *testing.T) {
	ops := []string{"+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=", "&&", "||", "&", "|", "^", "<<", ">>"}
	for _, op := range ops {
		e := &parser.BinaryExpr{Op: op}
		if e.Op != op {
			t.Errorf("Op = %q, want %q", e.Op, op)
		}
	}
}

func TestUnaryExpr_Fields(t *testing.T) {
	pos := tok(1, 1, scanner.TokenBang, "!")
	x := &parser.Identifier{Name: "flag"}
	e := &parser.UnaryExpr{Op: "!", X: x, Pos: pos}
	if e.Op != "!" || e.X != x {
		t.Error("UnaryExpr field mismatch")
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestUnaryExpr_AllPrefixOps(t *testing.T) {
	ops := []string{"!", "-", "~", "++", "--"}
	for _, op := range ops {
		e := &parser.UnaryExpr{Op: op}
		if e.Op != op {
			t.Errorf("Op = %q, want %q", e.Op, op)
		}
	}
}

func TestPostfixExpr_Fields(t *testing.T) {
	pos := tok(4, 3, scanner.TokenPlusPlus, "++")
	x := &parser.Identifier{Name: "i"}
	e := &parser.PostfixExpr{Op: "++", X: x, Pos: pos}
	if e.Op != "++" || e.X != x {
		t.Error("PostfixExpr field mismatch")
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestPostfixExpr_BothOps(t *testing.T) {
	for _, op := range []string{"++", "--"} {
		e := &parser.PostfixExpr{Op: op}
		if e.Op != op {
			t.Errorf("Op = %q, want %q", e.Op, op)
		}
	}
}

func TestCallExpr_NoArgs(t *testing.T) {
	pos := tok(5, 1, scanner.TokenIdent, "doThing")
	callee := &parser.Identifier{Name: "doThing"}
	e := &parser.CallExpr{Callee: callee, Args: nil, Pos: pos}
	if e.Callee != callee {
		t.Error("Callee mismatch")
	}
	if len(e.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(e.Args))
	}
	if e.IsBlocking {
		t.Error("IsBlocking should default to false")
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestCallExpr_WithArgs(t *testing.T) {
	arg1 := &parser.Identifier{Name: "x"}
	arg2 := &parser.Literal{Kind: "int", Value: "3"}
	e := &parser.CallExpr{
		Callee: &parser.Identifier{Name: "move"},
		Args:   []parser.Expr{arg1, arg2},
	}
	if len(e.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(e.Args))
	}
	if e.Args[0] != arg1 || e.Args[1] != arg2 {
		t.Error("Args mismatch")
	}
}

func TestCallExpr_BlockingFlag(t *testing.T) {
	e := &parser.CallExpr{
		Callee:     &parser.Identifier{Name: "WalkTo"},
		IsBlocking: true,
	}
	if !e.IsBlocking {
		t.Error("IsBlocking should be true for blocking calls")
	}
}

func TestIndexExpr_Fields(t *testing.T) {
	pos := tok(3, 5, scanner.TokenLBracket, "[")
	obj := &parser.Identifier{Name: "arr"}
	idx := &parser.Literal{Kind: "int", Value: "0"}
	e := &parser.IndexExpr{Object: obj, Index: idx, Pos: pos}
	if e.Object != obj || e.Index != idx {
		t.Error("IndexExpr field mismatch")
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestMemberExpr_Fields(t *testing.T) {
	pos := tok(2, 3, scanner.TokenDot, ".")
	obj := &parser.Identifier{Name: "player"}
	e := &parser.MemberExpr{Object: obj, Field: "x", Pos: pos}
	if e.Object != obj || e.Field != "x" {
		t.Error("MemberExpr field mismatch")
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestGlobalExpr_Fields(t *testing.T) {
	pos := tok(1, 1, scanner.TokenGlobal, "global")
	e := &parser.GlobalExpr{Property: "player", Pos: pos}
	if e.Property != "player" {
		t.Errorf("Property = %q, want \"player\"", e.Property)
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestGlobalExpr_KnownProperties(t *testing.T) {
	for _, prop := range []string{"player", "room", "score", "mouse", "game"} {
		e := &parser.GlobalExpr{Property: prop}
		if e.Property != prop {
			t.Errorf("Property = %q, want %q", e.Property, prop)
		}
	}
}

func TestIdentifier_Fields(t *testing.T) {
	pos := tok(2, 5, scanner.TokenIdent, "myVar")
	e := &parser.Identifier{Name: "myVar", Pos: pos}
	if e.Name != "myVar" {
		t.Errorf("Name = %q, want \"myVar\"", e.Name)
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestLiteral_IntKind(t *testing.T) {
	pos := tok(1, 10, scanner.TokenIntLit, "42")
	e := &parser.Literal{Kind: "int", Value: "42", Pos: pos}
	if e.Kind != "int" || e.Value != "42" {
		t.Errorf("Literal{int,42} got {%q,%q}", e.Kind, e.Value)
	}
	if e.ExprPos() != pos {
		t.Error("ExprPos() mismatch")
	}
}

func TestLiteral_AllKinds(t *testing.T) {
	cases := []struct {
		kind  string
		value string
	}{
		{"int", "42"},
		{"float", "3.14"},
		{"string", "hello"},
		{"bool", "true"},
		{"null", "null"},
	}
	for _, c := range cases {
		e := &parser.Literal{Kind: c.kind, Value: c.value}
		if e.Kind != c.kind || e.Value != c.value {
			t.Errorf("Literal{%q,%q} got {%q,%q}", c.kind, c.value, e.Kind, e.Value)
		}
	}
}

// -------------------------------------------------------------------
// DeclPos / ExprPos delegation
// -------------------------------------------------------------------

func TestDeclPos_DelegatesCorrectly(t *testing.T) {
	pos := tok(5, 1, scanner.TokenFunction, "function")

	fn := &parser.FunctionDecl{Pos: pos}
	if fn.DeclPos() != pos {
		t.Error("FunctionDecl.DeclPos() wrong")
	}

	ns := &parser.NamespaceDecl{Pos: pos}
	if ns.DeclPos() != pos {
		t.Error("NamespaceDecl.DeclPos() wrong")
	}

	ed := &parser.EnumDecl{Pos: pos}
	if ed.DeclPos() != pos {
		t.Error("EnumDecl.DeclPos() wrong")
	}

	inner := &parser.VarDecl{Pos: pos}
	tv := &parser.TopVarDecl{Decl: inner}
	if tv.DeclPos() != pos {
		t.Error("TopVarDecl.DeclPos() wrong")
	}
}

func TestExprPos_DelegatesCorrectly(t *testing.T) {
	pos := tok(7, 3, scanner.TokenIdent, "x")

	cases := []parser.Expr{
		&parser.AssignExpr{Pos: pos},
		&parser.BinaryExpr{Pos: pos},
		&parser.UnaryExpr{Pos: pos},
		&parser.PostfixExpr{Pos: pos},
		&parser.CallExpr{Pos: pos},
		&parser.IndexExpr{Pos: pos},
		&parser.MemberExpr{Pos: pos},
		&parser.GlobalExpr{Pos: pos},
		&parser.Identifier{Pos: pos},
		&parser.Literal{Pos: pos},
	}
	for _, e := range cases {
		if e.ExprPos() != pos {
			t.Errorf("%T.ExprPos() wrong", e)
		}
	}
}
