package parser_test

// TEST-M2-02 — Parser AST unit tests with explicit UT IDs.
//
// These tests satisfy the acceptance criteria for GitHub issue
// "TEST-M2-02 — Write parser AST tests and fixtures (5 tests)".

import (
	"testing"

	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

func parseOne(src string) (*parser.File, []*parser.ParseError) {
	s := scanner.New("test.agscript", src)
	p := parser.New(s)
	return p.Parse("test.agscript")
}

// UT-M2-05: Empty function produces a FunctionDecl node in the AST.
func TestUT_M2_05_EmptyFunctionIsFunctionDecl(t *testing.T) {
	f, errs := parseOne("function room_Load() {}")
	if len(errs) > 0 {
		t.Fatalf("UT-M2-05: parse errors: %v", errs[0])
	}
	if len(f.Decls) != 1 {
		t.Fatalf("UT-M2-05: want 1 decl, got %d", len(f.Decls))
	}
	fd, ok := f.Decls[0].(*parser.FunctionDecl)
	if !ok {
		t.Fatalf("UT-M2-05: decl type=%T, want *FunctionDecl", f.Decls[0])
	}
	if fd.Name != "room_Load" {
		t.Errorf("UT-M2-05: Name=%q, want room_Load", fd.Name)
	}
}

// UT-M2-06: Event handler function (room_Load, room_AfterFadeIn, etc.)
// is represented as a FunctionDecl node — AGS-spirit has no separate
// EventHandler node type; event handlers are functions with conventional names.
func TestUT_M2_06_EventHandlerIsFunctionDecl(t *testing.T) {
	handlers := []string{
		"function room_Load() {}",
		"function room_AfterFadeIn() {}",
		"function room_LeaveRoom() {}",
		"function hHotspot1_Look() {}",
	}
	for _, src := range handlers {
		f, errs := parseOne(src)
		if len(errs) > 0 {
			t.Errorf("UT-M2-06: parse error in %q: %v", src, errs[0])
			continue
		}
		if len(f.Decls) == 0 {
			t.Errorf("UT-M2-06: no decls for %q", src)
			continue
		}
		if _, ok := f.Decls[0].(*parser.FunctionDecl); !ok {
			t.Errorf("UT-M2-06: %q → decl type %T, want *FunctionDecl", src, f.Decls[0])
		}
	}
}

// UT-M2-07: If statement produces an IfStmt node with the correct condition.
func TestUT_M2_07_IfStmtNode(t *testing.T) {
	f, errs := parseOne(`function f() { if (x > 0) { return; } }`)
	if len(errs) > 0 {
		t.Fatalf("UT-M2-07: parse errors: %v", errs[0])
	}
	fd := f.Decls[0].(*parser.FunctionDecl)
	if len(fd.Body.Stmts) == 0 {
		t.Fatal("UT-M2-07: empty function body")
	}
	ifStmt, ok := fd.Body.Stmts[0].(*parser.IfStmt)
	if !ok {
		t.Fatalf("UT-M2-07: stmt type=%T, want *IfStmt", fd.Body.Stmts[0])
	}
	// Condition must be a BinaryExpr (x > 0).
	cond, ok := ifStmt.Cond.(*parser.BinaryExpr)
	if !ok {
		t.Fatalf("UT-M2-07: cond type=%T, want *BinaryExpr", ifStmt.Cond)
	}
	if cond.Op != ">" {
		t.Errorf("UT-M2-07: cond.Op=%q, want >", cond.Op)
	}
	if ifStmt.Then == nil {
		t.Error("UT-M2-07: Then block is nil")
	}
}

// UT-M2-08: While statement produces a WhileStmt node.
func TestUT_M2_08_WhileStmtNode(t *testing.T) {
	f, errs := parseOne(`function f() { while (running) { int x = 1; } }`)
	if len(errs) > 0 {
		t.Fatalf("UT-M2-08: parse errors: %v", errs[0])
	}
	fd := f.Decls[0].(*parser.FunctionDecl)
	ws, ok := fd.Body.Stmts[0].(*parser.WhileStmt)
	if !ok {
		t.Fatalf("UT-M2-08: stmt type=%T, want *WhileStmt", fd.Body.Stmts[0])
	}
	// Condition must be an identifier "running".
	ident, ok := ws.Cond.(*parser.Identifier)
	if !ok {
		t.Fatalf("UT-M2-08: cond type=%T, want *Identifier", ws.Cond)
	}
	if ident.Name != "running" {
		t.Errorf("UT-M2-08: cond name=%q, want running", ident.Name)
	}
	if ws.Body == nil {
		t.Error("UT-M2-08: Body is nil")
	}
}

// UT-M2-09: Member expression (character.WalkTo(point.door)) produces the
// expected nested AST: CallExpr → Callee=MemberExpr{Object=Ident, Field=WalkTo}
//                                   Arg=MemberExpr{Object=Ident, Field=door}
func TestUT_M2_09_MemberCallAST(t *testing.T) {
	f, errs := parseOne(`function f() { character.WalkTo(point.door); }`)
	if len(errs) > 0 {
		t.Fatalf("UT-M2-09: parse errors: %v", errs[0])
	}
	fd := f.Decls[0].(*parser.FunctionDecl)
	es, ok := fd.Body.Stmts[0].(*parser.ExprStmt)
	if !ok {
		t.Fatalf("UT-M2-09: stmt type=%T, want *ExprStmt", fd.Body.Stmts[0])
	}
	call, ok := es.X.(*parser.CallExpr)
	if !ok {
		t.Fatalf("UT-M2-09: expr type=%T, want *CallExpr", es.X)
	}

	// Callee must be MemberExpr{Object=Identifier{character}, Field=WalkTo}.
	callee, ok := call.Callee.(*parser.MemberExpr)
	if !ok {
		t.Fatalf("UT-M2-09: callee type=%T, want *MemberExpr", call.Callee)
	}
	if callee.Field != "WalkTo" {
		t.Errorf("UT-M2-09: callee.Field=%q, want WalkTo", callee.Field)
	}
	obj, ok := callee.Object.(*parser.Identifier)
	if !ok {
		t.Fatalf("UT-M2-09: callee.Object type=%T, want *Identifier", callee.Object)
	}
	if obj.Name != "character" {
		t.Errorf("UT-M2-09: callee.Object.Name=%q, want character", obj.Name)
	}

	// Single argument: MemberExpr{Object=Identifier{point}, Field=door}.
	if len(call.Args) != 1 {
		t.Fatalf("UT-M2-09: len(Args)=%d, want 1", len(call.Args))
	}
	arg, ok := call.Args[0].(*parser.MemberExpr)
	if !ok {
		t.Fatalf("UT-M2-09: arg type=%T, want *MemberExpr", call.Args[0])
	}
	if arg.Field != "door" {
		t.Errorf("UT-M2-09: arg.Field=%q, want door", arg.Field)
	}
	argObj, ok := arg.Object.(*parser.Identifier)
	if !ok {
		t.Fatalf("UT-M2-09: arg.Object type=%T, want *Identifier", arg.Object)
	}
	if argObj.Name != "point" {
		t.Errorf("UT-M2-09: arg.Object.Name=%q, want point", argObj.Name)
	}
}
