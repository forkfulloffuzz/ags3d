package emitter_test

// TEST-M3-01 / TEST-M3-02 — Emitter unit tests with explicit UT IDs.
//
// These tests satisfy the acceptance criteria for GitHub issues:
//   - "TEST-M3-01 — Write emitter golden file tests (7 tests)"
//   - "TEST-M3-02 — Write await emission and source map tests (6 tests)"
//
// UT-M3-01 through UT-M3-09 are fully implemented below.
// UT-M3-10 through UT-M3-13 depend on T17/T18 and are implemented
// as part of the source map and build pipeline features.

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/emitter"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// emitUT is a helper that parses src and emits GDScript, failing on any error.
func emitUT(t *testing.T, src string) string {
	t.Helper()
	s := scanner.New("test.agscript", src)
	p := parser.New(s)
	f, errs := p.Parse("test.agscript")
	if len(errs) > 0 {
		t.Fatalf("parse error: %v", errs[0])
	}
	result, err := emitter.New().Emit(f)
	if err != nil {
		t.Fatalf("emit error: %v", err)
	}
	return result.GDScript
}

// emitUTWithResult returns the full Result so tests can inspect SourceMap.
func emitUTWithResult(t *testing.T, src string) *emitter.Result {
	t.Helper()
	s := scanner.New("test.agscript", src)
	p := parser.New(s)
	f, errs := p.Parse("test.agscript")
	if len(errs) > 0 {
		t.Fatalf("parse error: %v", errs[0])
	}
	result, err := emitter.New().Emit(f)
	if err != nil {
		t.Fatalf("emit error: %v", err)
	}
	return result
}

// UT-M3-01: Empty function emits a GDScript func with pass body.
func TestUT_M3_01_EmptyFunctionEmitsPass(t *testing.T) {
	got := emitUT(t, "function room_Load() {}")
	if !strings.Contains(got, "func room_load():") {
		t.Errorf("UT-M3-01: missing func header, got:\n%s", got)
	}
	if !strings.Contains(got, "pass") {
		t.Errorf("UT-M3-01: missing pass in empty body, got:\n%s", got)
	}
}

// UT-M3-02: Event handler name is converted to snake_case GDScript func name.
func TestUT_M3_02_EventHandlerSnakeCase(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{"function room_Load() {}", "func room_load():"},
		{"function room_AfterFadeIn() {}", "func room_after_fade_in():"},
		{"function hHotspot1_Look() {}", "func h_hotspot1_look():"},
	}
	for _, tc := range cases {
		got := emitUT(t, tc.src)
		if !strings.Contains(got, tc.want) {
			t.Errorf("UT-M3-02: %q → want %q in:\n%s", tc.src, tc.want, got)
		}
	}
}

// UT-M3-03: If statement emits correct GDScript (golden comparison).
func TestUT_M3_03_IfStatementEmit(t *testing.T) {
	got := emitUT(t, `function f() { if (x > 0) { return 1; } }`)
	if !strings.Contains(got, "if x > 0:") {
		t.Errorf("UT-M3-03: missing 'if x > 0:', got:\n%s", got)
	}
	if !strings.Contains(got, "return 1") {
		t.Errorf("UT-M3-03: missing 'return 1', got:\n%s", got)
	}
}

// UT-M3-04: While statement emits correct GDScript.
func TestUT_M3_04_WhileStatementEmit(t *testing.T) {
	got := emitUT(t, `function f() { while (running) { int x = 1; } }`)
	if !strings.Contains(got, "while running:") {
		t.Errorf("UT-M3-04: missing 'while running:', got:\n%s", got)
	}
	if !strings.Contains(got, "var x: int = 1") {
		t.Errorf("UT-M3-04: missing 'var x: int = 1', got:\n%s", got)
	}
}

// UT-M3-05: Assignment statement emits correctly.
func TestUT_M3_05_AssignmentEmit(t *testing.T) {
	got := emitUT(t, `function f() { int x = 0; x = 5; x += 3; }`)
	if !strings.Contains(got, "x = 5") {
		t.Errorf("UT-M3-05: missing 'x = 5', got:\n%s", got)
	}
	if !strings.Contains(got, "x += 3") {
		t.Errorf("UT-M3-05: missing 'x += 3', got:\n%s", got)
	}
}

// UT-M3-06: Member call emits correctly with snake_case method name.
func TestUT_M3_06_MemberCallEmit(t *testing.T) {
	got := emitUT(t, `function f() { obj.SetPosition(1, 2); }`)
	if !strings.Contains(got, "obj.set_position(1, 2)") {
		t.Errorf("UT-M3-06: missing 'obj.set_position(1, 2)', got:\n%s", got)
	}
}

// UT-M3-07: Blocking call emits await prefix.
func TestUT_M3_07_BlockingCallEmitsAwait(t *testing.T) {
	got := emitUT(t, `function f() { global.player.WalkTo(point.door); }`)
	if !strings.Contains(got, "await player.walk_to(") {
		t.Errorf("UT-M3-07: missing 'await player.walk_to(', got:\n%s", got)
	}
}

// UT-M3-08: Async (blocking) function emits correctly — func itself gets no
// special keyword, but its call sites get await.
func TestUT_M3_08_AsyncFunctionEmit(t *testing.T) {
	got := emitUT(t, `
function walkAndGreet() {
    global.player.WalkTo(point.door);
}
function room_AfterFadeIn() {
    walkAndGreet();
}`)
	// The function declaration itself is a plain func.
	if !strings.Contains(got, "func walk_and_greet():") {
		t.Errorf("UT-M3-08: missing 'func walk_and_greet():', got:\n%s", got)
	}
	// Its call site must be awaited because walkAndGreet is transitively blocking.
	if !strings.Contains(got, "await walk_and_greet()") {
		t.Errorf("UT-M3-08: missing 'await walk_and_greet()', got:\n%s", got)
	}
}

// UT-M3-09: Nested blocking calls emit a correct await chain.
func TestUT_M3_09_NestedBlockingChain(t *testing.T) {
	got := emitUT(t, `
function room_AfterFadeIn() {
    global.player.WalkTo(point.door);
    global.player.Say("Hello!");
    Wait(60);
}`)
	if !strings.Contains(got, "await player.walk_to(") {
		t.Errorf("UT-M3-09: missing walk_to await, got:\n%s", got)
	}
	if !strings.Contains(got, "await player.say(") {
		t.Errorf("UT-M3-09: missing say await, got:\n%s", got)
	}
	if !strings.Contains(got, "await wait(60)") {
		t.Errorf("UT-M3-09: missing Wait await, got:\n%s", got)
	}
}

// UT-M3-10: Source map line count is non-zero and ≤ GDScript line count.
func TestUT_M3_10_SourceMapLineCount(t *testing.T) {
	result := emitUTWithResult(t, `function f() { int x = 1; return x; }`)
	gdLines := strings.Count(result.GDScript, "\n")
	if len(result.SourceMap) == 0 {
		t.Fatal("UT-M3-10: SourceMap is empty")
	}
	if len(result.SourceMap) > gdLines {
		t.Errorf("UT-M3-10: SourceMap entries (%d) > GDScript lines (%d)",
			len(result.SourceMap), gdLines)
	}
}

// UT-M3-11: Source map maps a known AGS-spirit line to the correct GDScript line.
// Input: single-line "function room_Load() {}" — func header is line 1 in AGS-spirit.
// Expected: at least one source map entry pointing to agscript line 1.
func TestUT_M3_11_SourceMapKnownLine(t *testing.T) {
	result := emitUTWithResult(t, "function room_Load() {}")
	found := false
	for _, entry := range result.SourceMap {
		// entry = [gdLine, srcFile, srcLine]
		if srcLine, ok := entry[2].(int); ok && srcLine == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("UT-M3-11: no source map entry for AGS-spirit line 1; map=%v", result.SourceMap)
	}
}

// UT-M3-12: Source map entries reference the correct source file path.
func TestUT_M3_12_SourceMapFileRef(t *testing.T) {
	s := scanner.New("rooms/start.agscript", "function room_Load() {}")
	p := parser.New(s)
	f, errs := p.Parse("rooms/start.agscript")
	if len(errs) > 0 {
		t.Fatalf("UT-M3-12: parse error: %v", errs[0])
	}
	result, err := emitter.New().Emit(f)
	if err != nil {
		t.Fatalf("UT-M3-12: emit error: %v", err)
	}
	for _, entry := range result.SourceMap {
		if srcFile, ok := entry[1].(string); ok {
			if srcFile != "rooms/start.agscript" {
				t.Errorf("UT-M3-12: srcFile=%q, want rooms/start.agscript", srcFile)
			}
		}
	}
}

// UT-M3-13: Source map GDScript line numbers are positive and monotonically
// non-decreasing (lines are emitted in order).
func TestUT_M3_13_SourceMapLineOrder(t *testing.T) {
	result := emitUTWithResult(t, `function f() {
    int x = 1;
    int y = 2;
    return x + y;
}`)
	prev := 0
	for _, entry := range result.SourceMap {
		gdLine, ok := entry[0].(int)
		if !ok {
			t.Fatalf("UT-M3-13: GDScript line is not int: %T", entry[0])
		}
		if gdLine <= 0 {
			t.Errorf("UT-M3-13: non-positive GDScript line %d", gdLine)
		}
		if gdLine < prev {
			t.Errorf("UT-M3-13: source map out of order: %d after %d", gdLine, prev)
		}
		prev = gdLine
	}
}
