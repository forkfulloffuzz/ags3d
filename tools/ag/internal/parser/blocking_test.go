package parser_test

// T11 — blocking call annotation tests.
//
// Covers:
//   - UT-M2-12  WalkTo method call is annotated IsBlocking=true
//   - UT-M2-13  Non-blocking call stays IsBlocking=false
//   - Global blocking built-ins (Wait, FadeIn, Display, …)
//   - Transitive propagation: user function calling a blocking function
//     is itself marked IsBlocking=true
//   - Deep transitive chain (A → B → C → blocking)
//   - All 22 valid fixtures produce no blocking-annotation panics/errors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// parseAndAnnotate is a test helper: parse, build symbol table, annotate blocking.
func parseAndAnnotate(src string) (*parser.File, *parser.SymbolTable) {
	s := scanner.New("test.agscript", src)
	p := parser.New(s)
	f, parseErrs := p.Parse("test.agscript")
	if len(parseErrs) > 0 {
		panic("parse errors: " + parseErrs[0].Error())
	}
	st, _ := parser.BuildSymbolTable(f)
	parser.AnnotateBlocking(f, st)
	return f, st
}

// firstCallExpr returns the first *CallExpr found in the first statement of
// the first function body in f.  Panics if not found (test misconfiguration).
func firstCallExprInFunc(f *parser.File) *parser.CallExpr {
	for _, d := range f.Decls {
		fd, ok := d.(*parser.FunctionDecl)
		if !ok || fd.Body == nil || len(fd.Body.Stmts) == 0 {
			continue
		}
		for _, stmt := range fd.Body.Stmts {
			if es, ok := stmt.(*parser.ExprStmt); ok {
				if ce, ok := es.X.(*parser.CallExpr); ok {
					return ce
				}
			}
		}
	}
	panic("no CallExpr found in first function body")
}

// funcByName returns the *FunctionDecl with the given name (top-level only).
func funcByName(f *parser.File, name string) *parser.FunctionDecl {
	for _, d := range f.Decls {
		if fd, ok := d.(*parser.FunctionDecl); ok && fd.Name == name {
			return fd
		}
	}
	return nil
}

// -------------------------------------------------------------------
// UT-M2-12 — blocking method calls are annotated
// -------------------------------------------------------------------

func TestBlocking_WalkTo_IsBlocking(t *testing.T) {
	f, _ := parseAndAnnotate(`
function go() {
    global.player.WalkTo(point.door);
}
`)
	fd := funcByName(f, "go")
	if fd == nil {
		t.Fatal("function 'go' not found")
	}
	if !fd.IsBlocking {
		t.Error("FunctionDecl.IsBlocking should be true when body calls WalkTo")
	}
	// Find the CallExpr for WalkTo.
	ce := firstCallExprInFunc(f)
	if !ce.IsBlocking {
		t.Error("CallExpr.IsBlocking should be true for WalkTo call")
	}
}

func TestBlocking_Say_IsBlocking(t *testing.T) {
	f, _ := parseAndAnnotate(`
function speak() {
    cChar.Say("hello");
}
`)
	fd := funcByName(f, "speak")
	if fd == nil {
		t.Fatal("function 'speak' not found")
	}
	if !fd.IsBlocking {
		t.Error("FunctionDecl.IsBlocking should be true for Say")
	}
}

func TestBlocking_AllMethodNames(t *testing.T) {
	methods := []string{
		"WalkTo", "WalkStraight", "Say", "Think", "PlayAnimation",
		"FaceDirection", "FaceCharacter", "FacePoint", "RunInteraction",
	}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			src := `function f() { obj.` + m + `(1); }`
			f, _ := parseAndAnnotate(src)
			fd := funcByName(f, "f")
			if fd == nil {
				t.Fatal("function 'f' not found")
			}
			if !fd.IsBlocking {
				t.Errorf("FunctionDecl.IsBlocking should be true for method %s", m)
			}
		})
	}
}

// -------------------------------------------------------------------
// Global blocking built-ins
// -------------------------------------------------------------------

func TestBlocking_Wait_IsBlocking(t *testing.T) {
	f, _ := parseAndAnnotate(`
function pause() {
    Wait(60);
}
`)
	fd := funcByName(f, "pause")
	if fd == nil {
		t.Fatal("function not found")
	}
	if !fd.IsBlocking {
		t.Error("FunctionDecl.IsBlocking should be true for Wait")
	}
}

func TestBlocking_AllGlobalBuiltins(t *testing.T) {
	builtins := []string{
		"Wait", "WaitKey", "WaitMouse", "WaitInput",
		"FadeIn", "FadeOut", "Display", "DisplayMessage",
	}
	for _, b := range builtins {
		t.Run(b, func(t *testing.T) {
			src := `function f() { ` + b + `(1); }`
			f, _ := parseAndAnnotate(src)
			fd := funcByName(f, "f")
			if fd == nil {
				t.Fatal("function 'f' not found")
			}
			if !fd.IsBlocking {
				t.Errorf("FunctionDecl.IsBlocking should be true for builtin %s", b)
			}
		})
	}
}

// -------------------------------------------------------------------
// UT-M2-13 — non-blocking calls stay false
// -------------------------------------------------------------------

func TestBlocking_NonBlocking_IsFalse(t *testing.T) {
	f, _ := parseAndAnnotate(`
function update() {
    SetGlobalInt(0, 42);
    int x;
    x = getScore();
}
`)
	fd := funcByName(f, "update")
	if fd == nil {
		t.Fatal("function not found")
	}
	if fd.IsBlocking {
		t.Error("FunctionDecl.IsBlocking should be false for non-blocking calls")
	}
}

func TestBlocking_EmptyFunction_IsFalse(t *testing.T) {
	f, _ := parseAndAnnotate(`function noop() {}`)
	fd := funcByName(f, "noop")
	if fd == nil {
		t.Fatal("function not found")
	}
	if fd.IsBlocking {
		t.Error("empty function should not be blocking")
	}
}

// -------------------------------------------------------------------
// Transitive propagation
// -------------------------------------------------------------------

// TestBlocking_Transitive_Direct: walkAndGreet contains WalkTo → blocking.
// room_AfterFadeIn calls walkAndGreet → also blocking.
func TestBlocking_Transitive_WalkAndGreet(t *testing.T) {
	f, _ := parseAndAnnotate(`
function walkAndGreet() {
    global.player.WalkTo(point.npc_guard);
    global.player.Say("Good day!");
}
function room_AfterFadeIn() {
    walkAndGreet();
}
`)
	wag := funcByName(f, "walkAndGreet")
	if wag == nil {
		t.Fatal("walkAndGreet not found")
	}
	if !wag.IsBlocking {
		t.Error("walkAndGreet should be blocking (contains WalkTo + Say)")
	}

	rafi := funcByName(f, "room_AfterFadeIn")
	if rafi == nil {
		t.Fatal("room_AfterFadeIn not found")
	}
	if !rafi.IsBlocking {
		t.Error("room_AfterFadeIn should be blocking (calls walkAndGreet)")
	}
}

// TestBlocking_Transitive_Deep: A → B → C → blocking built-in.
// All three should become blocking via fixed-point propagation.
func TestBlocking_Transitive_Deep(t *testing.T) {
	f, _ := parseAndAnnotate(`
function c() { FadeOut(15); }
function b() { c(); }
function a() { b(); }
`)
	for _, name := range []string{"a", "b", "c"} {
		fd := funcByName(f, name)
		if fd == nil {
			t.Fatalf("function %q not found", name)
		}
		if !fd.IsBlocking {
			t.Errorf("function %q should be blocking via transitive propagation", name)
		}
	}
}

// TestBlocking_Transitive_Nonblocking_Chain: A → B, neither is blocking.
func TestBlocking_Transitive_NeitherBlocking(t *testing.T) {
	f, _ := parseAndAnnotate(`
function helper() { int x; x = 1 + 2; }
function caller() { helper(); }
`)
	for _, name := range []string{"helper", "caller"} {
		fd := funcByName(f, name)
		if fd == nil {
			t.Fatalf("function %q not found", name)
		}
		if fd.IsBlocking {
			t.Errorf("function %q should not be blocking", name)
		}
	}
}

// -------------------------------------------------------------------
// SymbolTable sync
// -------------------------------------------------------------------

func TestBlocking_SymbolTableSync(t *testing.T) {
	f, st := parseAndAnnotate(`
function doFade() { FadeIn(30); }
`)
	_ = f
	sym, ok := st.Globals["doFade"]
	if !ok {
		t.Fatal("doFade not in symbol table globals")
	}
	if !sym.IsBlocking {
		t.Error("Symbol.IsBlocking should be synced to true after AnnotateBlocking")
	}
}

// -------------------------------------------------------------------
// IsBlockingBuiltin / IsBlockingMethod helpers
// -------------------------------------------------------------------

func TestIsBlockingBuiltin(t *testing.T) {
	yes := []string{"Wait", "WaitKey", "WaitMouse", "WaitInput", "FadeIn", "FadeOut", "Display", "DisplayMessage"}
	no := []string{"Random", "GetGlobalInt", "SetGlobalInt", "GetTime", "IsKeyPressed"}
	for _, name := range yes {
		if !parser.IsBlockingBuiltin(name) {
			t.Errorf("IsBlockingBuiltin(%q) should be true", name)
		}
	}
	for _, name := range no {
		if parser.IsBlockingBuiltin(name) {
			t.Errorf("IsBlockingBuiltin(%q) should be false", name)
		}
	}
}

func TestIsBlockingMethod(t *testing.T) {
	yes := []string{"WalkTo", "WalkStraight", "Say", "Think", "PlayAnimation",
		"FaceDirection", "FaceCharacter", "FacePoint", "RunInteraction"}
	no := []string{"SetPosition", "GetX", "GetY", "IsEnabled", "ChangeView"}
	for _, name := range yes {
		if !parser.IsBlockingMethod(name) {
			t.Errorf("IsBlockingMethod(%q) should be true", name)
		}
	}
	for _, name := range no {
		if parser.IsBlockingMethod(name) {
			t.Errorf("IsBlockingMethod(%q) should be false", name)
		}
	}
}

// -------------------------------------------------------------------
// Fixture 15 — blocking_calls.agscript
// -------------------------------------------------------------------

func TestBlocking_Fixture15(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "scripts", "valid", "15_blocking_calls.agscript"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	src := string(data)
	s := scanner.New("15_blocking_calls.agscript", src)
	p := parser.New(s)
	f, parseErrs := p.Parse("15_blocking_calls.agscript")
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs[0])
	}
	st, _ := parser.BuildSymbolTable(f)
	parser.AnnotateBlocking(f, st)

	blocking := []string{"testCharacterBlocking", "testGlobalBlocking", "testBlockingChain", "walkAndGreet", "room_AfterFadeIn"}
	for _, name := range blocking {
		fd := funcByName(f, name)
		if fd == nil {
			t.Errorf("function %q not found in fixture 15", name)
			continue
		}
		if !fd.IsBlocking {
			t.Errorf("function %q should be IsBlocking=true in fixture 15", name)
		}
	}
}

// -------------------------------------------------------------------
// All valid fixtures — AnnotateBlocking must not panic
// -------------------------------------------------------------------

func TestBlocking_ValidFixtures_NoPanic(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "..", "testdata", "scripts", "valid", "*.agscript"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no fixture files found")
	}
	for _, path := range fixtures {
		name := filepath.Base(path)
		t.Run(strings.TrimSuffix(name, ".agscript"), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			s := scanner.New(name, string(data))
			p := parser.New(s)
			f, parseErrs := p.Parse(name)
			if len(parseErrs) > 0 {
				t.Fatalf("parse errors: %v", parseErrs[0])
			}
			st, _ := parser.BuildSymbolTable(f)
			// Must not panic.
			parser.AnnotateBlocking(f, st)
		})
	}
}
