package parser_test

// T10 — symbol table and name resolution tests.
//
// Covers:
//   - UT-M2-10  Undefined reference produces an error
//   - UT-M2-11  Valid function call resolves with zero errors
//   - Collection: functions, vars, enums, namespaces
//   - Duplicate declaration detection
//   - Local variable scoping and shadowing
//   - Namespace member resolution
//   - Expression type annotation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// parseAndBuild is a test helper: parse source then run BuildSymbolTable.
func parseAndBuild(src string) (*parser.File, *parser.SymbolTable, []*parser.SymError) {
	s := scanner.New("test.agscript", src)
	p := parser.New(s)
	f, parseErrs := p.Parse("test.agscript")
	if len(parseErrs) > 0 {
		// Return parse errors as SymErrors so callers can fail uniformly.
		var symErrs []*parser.SymError
		for _, pe := range parseErrs {
			symErrs = append(symErrs, &parser.SymError{
				File:    pe.File,
				Line:    pe.Line,
				Column:  pe.Column,
				Message: "parse: " + pe.Message,
			})
		}
		return f, nil, symErrs
	}
	st, errs := parser.BuildSymbolTable(f)
	return f, st, errs
}

// requireNoErrors fails if any diagnostic with severity "error" is present.
// Warnings (cross-file references, unknown types) are allowed.
func requireNoErrors(t *testing.T, errs []*parser.SymError) {
	t.Helper()
	for _, e := range errs {
		if e.Severity == "error" {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

// requireError fails unless at least one diagnostic (any severity) contains substr.
func requireError(t *testing.T, errs []*parser.SymError, substr string) {
	t.Helper()
	for _, e := range errs {
		if strings.Contains(e.Message, substr) {
			return
		}
	}
	t.Errorf("expected a diagnostic containing %q, got: %v", substr, errs)
}

// -------------------------------------------------------------------
// Collection tests
// -------------------------------------------------------------------

func TestSymbols_FunctionCollected(t *testing.T) {
	_, st, errs := parseAndBuild(`function foo() { }`)
	requireNoErrors(t, errs)
	sym, ok := st.Globals["foo"]
	if !ok {
		t.Fatal("expected 'foo' in globals")
	}
	if sym.Kind != parser.KindFunction {
		t.Errorf("expected KindFunction, got %v", sym.Kind)
	}
	if sym.Type != "" {
		t.Errorf("expected void return type, got %q", sym.Type)
	}
}

func TestSymbols_FunctionWithReturnType(t *testing.T) {
	_, st, errs := parseAndBuild(`int function add(int a, int b) { return a; }`)
	requireNoErrors(t, errs)
	sym, ok := st.Globals["add"]
	if !ok {
		t.Fatal("expected 'add' in globals")
	}
	if sym.Type != "int" {
		t.Errorf("expected return type 'int', got %q", sym.Type)
	}
	if len(sym.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(sym.Params))
	}
}

func TestSymbols_TopVarCollected(t *testing.T) {
	_, st, errs := parseAndBuild(`int score;`)
	requireNoErrors(t, errs)
	sym, ok := st.Globals["score"]
	if !ok {
		t.Fatal("expected 'score' in globals")
	}
	if sym.Kind != parser.KindVar {
		t.Errorf("expected KindVar, got %v", sym.Kind)
	}
	if sym.Type != "int" {
		t.Errorf("expected type 'int', got %q", sym.Type)
	}
}

func TestSymbols_EnumCollected(t *testing.T) {
	_, st, errs := parseAndBuild(`enum Color { Red, Green, Blue }`)
	requireNoErrors(t, errs)

	enumSym, ok := st.Globals["Color"]
	if !ok {
		t.Fatal("expected 'Color' enum in globals")
	}
	if enumSym.Kind != parser.KindEnum {
		t.Errorf("expected KindEnum, got %v", enumSym.Kind)
	}
	for _, member := range []string{"Red", "Green", "Blue"} {
		sym, ok := st.Globals[member]
		if !ok {
			t.Errorf("expected enum member %q in globals", member)
			continue
		}
		if sym.Kind != parser.KindEnumMember {
			t.Errorf("%q: expected KindEnumMember, got %v", member, sym.Kind)
		}
		if sym.Type != "Color" {
			t.Errorf("%q: expected type 'Color', got %q", member, sym.Type)
		}
	}
}

func TestSymbols_NamespaceCollected(t *testing.T) {
	src := `
namespace Math {
    export int function square(int x) { return x; }
    export function noop() { }
}`
	_, st, errs := parseAndBuild(src)
	requireNoErrors(t, errs)

	nsSym, ok := st.Globals["Math"]
	if !ok {
		t.Fatal("expected 'Math' in globals as KindNamespace")
	}
	if nsSym.Kind != parser.KindNamespace {
		t.Errorf("expected KindNamespace, got %v", nsSym.Kind)
	}

	members, ok := st.Namespaces["Math"]
	if !ok {
		t.Fatal("expected 'Math' in Namespaces map")
	}
	if _, ok := members["square"]; !ok {
		t.Error("expected 'square' in Math namespace")
	}
	if _, ok := members["noop"]; !ok {
		t.Error("expected 'noop' in Math namespace")
	}
}

func TestSymbols_DuplicateGlobalFunction(t *testing.T) {
	src := `
function foo() { }
function foo() { }
`
	_, _, errs := parseAndBuild(src)
	requireError(t, errs, "redeclared")
}

func TestSymbols_DuplicateNamespaceMember(t *testing.T) {
	src := `
namespace NS {
    export function bar() { }
    export function bar() { }
}`
	_, _, errs := parseAndBuild(src)
	requireError(t, errs, "redeclared")
}

// -------------------------------------------------------------------
// Resolution tests — UT-M2-10 and UT-M2-11
// -------------------------------------------------------------------

// UT-M2-10: Undefined reference produces an error.
func TestSymbols_UndefinedReference(t *testing.T) {
	src := `
function foo() {
    int x = unknownVar;
}
`
	_, _, errs := parseAndBuild(src)
	requireError(t, errs, `undefined: "unknownVar"`)
}

// UT-M2-11: Valid function call resolves with zero errors.
func TestSymbols_ValidFunctionCall(t *testing.T) {
	src := `
int function add(int a, int b) { return a; }
function main() {
    int result = add(1, 2);
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_ParamVisibleInBody(t *testing.T) {
	src := `
function greet(int x) {
    int y = x;
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_LocalVarVisibleAfterDecl(t *testing.T) {
	src := `
function foo() {
    int a = 1;
    int b = a;
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_LocalVarNotVisibleBeforeDecl(t *testing.T) {
	src := `
function foo() {
    int b = a;
    int a = 1;
}
`
	_, _, errs := parseAndBuild(src)
	requireError(t, errs, `undefined: "a"`)
}

func TestSymbols_LocalVarShadowsGlobal(t *testing.T) {
	src := `
int x;
function foo() {
    int x = 5;
    int y = x;
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_ForInitVarScoped(t *testing.T) {
	src := `
function foo() {
    for (int i = 0; i < 10; i++) { }
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_NestedBlockScope(t *testing.T) {
	src := `
function foo() {
    {
        int inner = 1;
    }
    int x = inner;
}
`
	_, _, errs := parseAndBuild(src)
	requireError(t, errs, `undefined: "inner"`)
}

// -------------------------------------------------------------------
// Namespace resolution
// -------------------------------------------------------------------

func TestSymbols_NamespaceMemberCall(t *testing.T) {
	src := `
namespace Math {
    export int function square(int x) { return x; }
}
function main() {
    int s = Math.square(4);
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_UndefinedNamespaceMember(t *testing.T) {
	src := `
namespace Math {
    export int function square(int x) { return x; }
}
function main() {
    int s = Math.cube(4);
}
`
	_, _, errs := parseAndBuild(src)
	requireError(t, errs, `undefined member "cube" in namespace "Math"`)
}

func TestSymbols_NamespaceFunctionSeesGlobal(t *testing.T) {
	// A function inside a namespace can read file-scope variables.
	src := `
int counter;
namespace NS {
    export function bump() { counter = counter + 1; }
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_NamespaceFunctionSeesOwnMembers(t *testing.T) {
	src := `
namespace Util {
    export int function helper() { return 0; }
    export int function caller() { return helper(); }
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

func TestSymbols_EnumMemberResolvable(t *testing.T) {
	src := `
enum Dir { North, South, East, West }
function foo() {
    int d = North;
}
`
	_, _, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
}

// -------------------------------------------------------------------
// Expression type annotation
// -------------------------------------------------------------------

func TestSymbols_LiteralTypesAnnotated(t *testing.T) {
	src := `
function foo() {
    int a = 42;
    float b = 3.14;
    bool c = true;
    string d = "hi";
}
`
	_, st, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
	if st == nil {
		t.Fatal("nil symbol table")
	}
	// ExprTypes should contain entries for the literal nodes.
	found := map[string]bool{}
	for _, typ := range st.ExprTypes {
		found[typ] = true
	}
	for _, want := range []string{"int", "float", "bool", "string"} {
		if !found[want] {
			t.Errorf("expected type %q to appear in ExprTypes", want)
		}
	}
}

func TestSymbols_BinaryExprTypeBool(t *testing.T) {
	src := `
function foo() {
    bool ok = 1 == 2;
}
`
	_, st, errs := parseAndBuild(src)
	requireNoErrors(t, errs)

	found := false
	for _, typ := range st.ExprTypes {
		if typ == "bool" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'bool' type in ExprTypes for == expression")
	}
}

func TestSymbols_BinaryExprTypeFloat(t *testing.T) {
	src := `
function foo() {
    float f = 1 + 2.5;
}
`
	_, st, errs := parseAndBuild(src)
	requireNoErrors(t, errs)
	found := false
	for _, typ := range st.ExprTypes {
		if typ == "float" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'float' type in ExprTypes for int+float expression")
	}
}

// -------------------------------------------------------------------
// Lookup helpers
// -------------------------------------------------------------------

func TestSymbols_LookupGlobal(t *testing.T) {
	_, st, _ := parseAndBuild(`function bar() { }`)
	sym, ok := st.Lookup("bar")
	if !ok {
		t.Fatal("Lookup('bar') returned false")
	}
	if sym.Kind != parser.KindFunction {
		t.Errorf("expected KindFunction, got %v", sym.Kind)
	}
}

func TestSymbols_LookupMember(t *testing.T) {
	src := `namespace NS { export function fn() { } }`
	_, st, _ := parseAndBuild(src)
	sym, ok := st.LookupMember("NS", "fn")
	if !ok {
		t.Fatal("LookupMember('NS', 'fn') returned false")
	}
	if sym.Kind != parser.KindFunction {
		t.Errorf("expected KindFunction, got %v", sym.Kind)
	}
}

// -------------------------------------------------------------------
// Valid fixture files — all 22 must resolve without SymErrors
// -------------------------------------------------------------------

func TestSymbols_ValidFixtures(t *testing.T) {
	const testdataDir = "../../testdata"
	paths, err := filepath.Glob(filepath.Join(testdataDir, "scripts", "valid", "*.agscript"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no valid fixture files found")
	}
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			s := scanner.New(path, string(src))
			p := parser.New(s)
			f, parseErrs := p.Parse(path)
			if len(parseErrs) > 0 {
				t.Fatalf("parse errors: %v", parseErrs)
			}
			_, symErrs := parser.BuildSymbolTable(f)
			for _, e := range symErrs {
				if e.Severity == "error" {
					t.Errorf("sym error: %v", e)
				}
			}
		})
	}
}
