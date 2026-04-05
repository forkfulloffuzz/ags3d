// Tests for the AGS-spirit → GDScript emitter (T13–T16).
//
// Test structure:
//   - Golden file tests: parse a known fixture, emit, compare to .gd file
//   - Inline unit tests: specific constructs verified with embedded source
//   - Sweep tests: all valid fixtures parse + emit without panic/error
package emitter_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/emitter"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func emit(t *testing.T, src string) string {
	t.Helper()
	s := scanner.New("test.agscript", src)
	p := parser.New(s)
	f, parseErrs := p.Parse("test.agscript")
	if len(parseErrs) > 0 {
		t.Fatalf("parse error: %v", parseErrs[0])
	}
	e := emitter.New()
	result, err := e.Emit(f)
	if err != nil {
		t.Fatalf("emit error: %v", err)
	}
	return result.GDScript
}

func emitFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "valid", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	s := scanner.New(name, string(data))
	p := parser.New(s)
	f, parseErrs := p.Parse(name)
	if len(parseErrs) > 0 {
		t.Fatalf("parse error in %s: %v", name, parseErrs[0])
	}
	e := emitter.New()
	result, err := e.Emit(f)
	if err != nil {
		t.Fatalf("emit error in %s: %v", name, err)
	}
	return result.GDScript
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("output does not contain %q\ngot:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Errorf("output should not contain %q\ngot:\n%s", want, got)
	}
}

// -------------------------------------------------------------------
// Golden file tests
// -------------------------------------------------------------------

func TestEmitter_Golden_Minimal(t *testing.T) {
	got := emitFixture(t, "01_minimal.agscript")
	golden, err := os.ReadFile(filepath.Join("testdata", "01_minimal.gd"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	want := strings.TrimRight(string(golden), "\n") + "\n"
	got = strings.TrimRight(got, "\n") + "\n"
	if got != want {
		t.Errorf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestEmitter_Golden_BlockingCalls(t *testing.T) {
	got := emitFixture(t, "15_blocking_calls.agscript")
	golden, err := os.ReadFile(filepath.Join("testdata", "15_blocking_calls.gd"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	want := strings.TrimRight(string(golden), "\n") + "\n"
	got = strings.TrimRight(got, "\n") + "\n"
	if got != want {
		t.Errorf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

// -------------------------------------------------------------------
// T13 — scaffolding: printer + indentation
// -------------------------------------------------------------------

func TestEmitter_EmptyFileProducesOutput(t *testing.T) {
	s := scanner.New("empty.agscript", "")
	p := parser.New(s)
	f, _ := p.Parse("empty.agscript")
	e := emitter.New()
	result, err := e.Emit(f)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	if result == nil {
		t.Fatal("Emit returned nil result")
	}
}

func TestEmitter_EmptyFunctionBody_EmitsPass(t *testing.T) {
	got := emit(t, "function noop() {}")
	assertContains(t, got, "pass")
}

func TestEmitter_Indentation_NestedBlocks(t *testing.T) {
	got := emit(t, `function f() {
    if (true) {
        int x = 1;
    }
}`)
	assertContains(t, got, "\tif true:\n\t\tvar x: int = 1")
}

func TestEmitter_ResultIsNotNil(t *testing.T) {
	result, err := emitter.New().Emit(&parser.File{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

// -------------------------------------------------------------------
// T14 — function and event handler declarations
// -------------------------------------------------------------------

func TestEmitter_FuncDecl_VoidReturn(t *testing.T) {
	got := emit(t, "function room_Load() {}")
	assertContains(t, got, "func room_load():")
	assertNotContains(t, got, "->")
}

func TestEmitter_FuncDecl_WithReturnType(t *testing.T) {
	got := emit(t, "int function getScore() { return 0; }")
	assertContains(t, got, "func get_score() -> int:")
}

func TestEmitter_FuncDecl_WithParams(t *testing.T) {
	got := emit(t, "function move(int x, int y) {}")
	assertContains(t, got, "func move(x: int, y: int):")
}

func TestEmitter_FuncDecl_SnakeCase(t *testing.T) {
	got := emit(t, "function room_AfterFadeIn() {}")
	assertContains(t, got, "func room_after_fade_in():")
}

func TestEmitter_FuncDecl_TypeMapping_String(t *testing.T) {
	got := emit(t, `string function greet(string name) { return name; }`)
	assertContains(t, got, "func greet(name: String) -> String:")
}

func TestEmitter_FuncDecl_TypeMapping_IntVariants(t *testing.T) {
	got := emit(t, "function f(short a, char b) {}")
	assertContains(t, got, "a: int")
	assertContains(t, got, "b: int")
}

func TestEmitter_Namespace_EmitsClass(t *testing.T) {
	got := emit(t, `namespace Utils {
    export function square(int x) int {
        return x * x;
    }
}`)
	assertContains(t, got, "class Utils:")
	assertContains(t, got, "static func square(x: int) -> int:")
	assertContains(t, got, "\t\treturn x * x")
}

func TestEmitter_Enum_EmitsGDScriptEnum(t *testing.T) {
	got := emit(t, "enum Direction { eNorth, eSouth, eEast, eWest }")
	assertContains(t, got, "enum Direction { eNorth, eSouth, eEast, eWest }")
}

func TestEmitter_Enum_WithValues(t *testing.T) {
	got := emit(t, "enum State { sRunning = 0, sPaused = 1 }")
	assertContains(t, got, "sRunning = 0")
	assertContains(t, got, "sPaused = 1")
}

func TestEmitter_TopVar_WithType(t *testing.T) {
	got := emit(t, "int score;")
	assertContains(t, got, "var score: int")
}

func TestEmitter_TopVar_WithInit(t *testing.T) {
	got := emit(t, "int score = 100;")
	assertContains(t, got, "var score: int = 100")
}

// -------------------------------------------------------------------
// T15 — statements and expressions
// -------------------------------------------------------------------

func TestEmitter_VarDecl_LocalWithType(t *testing.T) {
	got := emit(t, `function f() { int x = 42; }`)
	assertContains(t, got, "var x: int = 42")
}

func TestEmitter_VarDecl_LocalNoInit(t *testing.T) {
	got := emit(t, `function f() { int x; }`)
	assertContains(t, got, "var x: int")
}

func TestEmitter_If_Simple(t *testing.T) {
	got := emit(t, `function f() { if (x > 0) { return 1; } }`)
	assertContains(t, got, "if x > 0:")
	assertContains(t, got, "return 1")
}

func TestEmitter_If_Else(t *testing.T) {
	got := emit(t, `function f() { if (x > 0) { return 1; } else { return 0; } }`)
	assertContains(t, got, "if x > 0:")
	assertContains(t, got, "else:")
}

func TestEmitter_If_ElseIf_Chain(t *testing.T) {
	got := emit(t, `function f() {
    if (x == 1) { return 1; }
    else if (x == 2) { return 2; }
    else { return 0; }
}`)
	assertContains(t, got, "if x == 1:")
	assertContains(t, got, "elif x == 2:")
	assertContains(t, got, "else:")
}

func TestEmitter_While(t *testing.T) {
	got := emit(t, `function f() { while (running) { int x = 1; } }`)
	assertContains(t, got, "while running:")
}

func TestEmitter_DoWhile(t *testing.T) {
	got := emit(t, `function f() { do { int x = 1; } while (x > 0); }`)
	assertContains(t, got, "while true:")
	assertContains(t, got, "if not (x > 0):")
	assertContains(t, got, "break")
}

func TestEmitter_For_EmitsWhileLoop(t *testing.T) {
	got := emit(t, `function f() { for (int i = 0; i < 10; i++) { int x = i; } }`)
	assertContains(t, got, "var i: int = 0")
	assertContains(t, got, "while i < 10:")
	assertContains(t, got, "i += 1")
}

func TestEmitter_Switch_EmitsMatch(t *testing.T) {
	got := emit(t, `function f() {
    int x = 1;
    switch (x) {
        case 1:
            int y = 1;
            break;
        default:
            int z = 0;
            break;
    }
}`)
	assertContains(t, got, "match x:")
	assertContains(t, got, "1:")
	assertContains(t, got, "_:")
	assertNotContains(t, got, "break")
}

func TestEmitter_Return_WithValue(t *testing.T) {
	got := emit(t, `int function f() { return 42; }`)
	assertContains(t, got, "return 42")
}

func TestEmitter_Return_Void(t *testing.T) {
	got := emit(t, `function f() { return; }`)
	assertContains(t, got, "return")
}

func TestEmitter_Break_Continue(t *testing.T) {
	got := emit(t, `function f() {
    while (true) { break; continue; }
}`)
	assertContains(t, got, "break")
	assertContains(t, got, "continue")
}

func TestEmitter_BinaryExpr_ArithmeticPassThrough(t *testing.T) {
	got := emit(t, `function f() { int x = 1 + 2 * 3; }`)
	assertContains(t, got, "1 + 2 * 3")
}

func TestEmitter_BinaryExpr_LogicalOps(t *testing.T) {
	got := emit(t, `function f() { bool b = x > 0 && y > 0; }`)
	assertContains(t, got, "and")
	assertNotContains(t, got, "&&")
}

func TestEmitter_BinaryExpr_OrOp(t *testing.T) {
	got := emit(t, `function f() { bool b = a || b; }`)
	assertContains(t, got, "or")
	assertNotContains(t, got, "||")
}

func TestEmitter_UnaryExpr_Not(t *testing.T) {
	got := emit(t, `function f() { bool b = !flag; }`)
	assertContains(t, got, "not ")
	assertNotContains(t, got, "!flag")
}

func TestEmitter_PostfixExpr_Increment(t *testing.T) {
	got := emit(t, `function f() { int x = 0; x++; }`)
	assertContains(t, got, "x += 1")
}

func TestEmitter_PostfixExpr_Decrement(t *testing.T) {
	got := emit(t, `function f() { int x = 5; x--; }`)
	assertContains(t, got, "x -= 1")
}

func TestEmitter_StringLiteral_HasQuotes(t *testing.T) {
	got := emit(t, `function f() { Display("Hello, world!"); }`)
	assertContains(t, got, `"Hello, world!"`)
}

func TestEmitter_GlobalExpr_EmitsGetGlobal(t *testing.T) {
	got := emit(t, `function f() { global.player.SetPosition(0, 0); }`)
	assertNotContains(t, got, "global.")
	assertContains(t, got, `AGSRuntime.get_global("player").set_position(0, 0)`)
}

func TestEmitter_MemberExpr_SnakeCase(t *testing.T) {
	got := emit(t, `function f() { obj.SetPosition(1, 2); }`)
	assertContains(t, got, "obj.set_position(1, 2)")
}

func TestEmitter_IndexExpr(t *testing.T) {
	got := emit(t, `function f() { int x = arr[0]; }`)
	assertContains(t, got, "arr[0]")
}

func TestEmitter_AssignExpr(t *testing.T) {
	got := emit(t, `function f() { int x = 0; x = 5; }`)
	assertContains(t, got, "x = 5")
}

func TestEmitter_CompoundAssign(t *testing.T) {
	got := emit(t, `function f() { int x = 0; x += 3; }`)
	assertContains(t, got, "x += 3")
}

// -------------------------------------------------------------------
// T16 — blocking calls with await
// -------------------------------------------------------------------

func TestEmitter_BlockingCall_MethodGetsAwait(t *testing.T) {
	got := emit(t, `function f() { global.player.WalkTo(point.door); }`)
	assertContains(t, got, `await AGSRuntime.get_character("player").walk_to("door")`)
}

func TestEmitter_BlockingCall_GlobalBuiltin(t *testing.T) {
	got := emit(t, `function f() { Wait(60); }`)
	assertContains(t, got, "await AGSCutscene.wait(60)")
}

func TestEmitter_BlockingCall_AllBuiltins(t *testing.T) {
	builtins := []string{"Wait", "WaitKey", "WaitMouse", "WaitInput", "FadeIn", "FadeOut", "Display", "DisplayMessage", "GoToRoom"}
	for _, b := range builtins {
		t.Run(b, func(t *testing.T) {
			got := emit(t, `function f() { `+b+`(1); }`)
			assertContains(t, got, "await ")
		})
	}
}

func TestEmitter_BlockingCall_TransitiveFunctionGetsAwait(t *testing.T) {
	got := emit(t, `
function walkAndGreet() {
    global.player.WalkTo(point.door);
}
function room_AfterFadeIn() {
    walkAndGreet();
}`)
	assertContains(t, got, "await walk_and_greet()")
}

func TestEmitter_NonBlockingCall_NoAwait(t *testing.T) {
	got := emit(t, `function f() { SetGlobalInt(0, 1); }`)
	assertNotContains(t, got, "await ")
}

// -------------------------------------------------------------------
// T32 — map AGS-spirit built-in names to AGSRuntime calls
// -------------------------------------------------------------------

func TestEmitter_T32_WalkTo_MapsToRuntimeCall(t *testing.T) {
	got := emit(t, `function f() { global.player.WalkTo(point.door); }`)
	assertContains(t, got, `AGSRuntime.get_character("player").walk_to("door")`)
	assertNotContains(t, got, "player.walk_to")
}

func TestEmitter_T32_FaceTo_MapsToRuntimeCall(t *testing.T) {
	got := emit(t, `function f() { global.player.FaceTo(point.entrance); }`)
	assertContains(t, got, `AGSRuntime.get_character("player").face_to("entrance")`)
}

func TestEmitter_T32_WalkTo_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { global.player.WalkTo(point.door); }`)
	assertContains(t, got, `await AGSRuntime.get_character("player").walk_to("door")`)
}

func TestEmitter_T32_FaceTo_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { global.player.FaceTo(point.door); }`)
	assertContains(t, got, `await AGSRuntime.get_character("player").face_to("door")`)
}

func TestEmitter_T32_NonBuiltin_UsesGetGlobal(t *testing.T) {
	// WalkStraight is not in T32's table — falls through to generic method call.
	// global.player resolves via AGSRuntime.get_global("player").
	got := emit(t, `function f() { global.player.WalkStraight(point.window); }`)
	assertContains(t, got, `AGSRuntime.get_global("player").walk_straight`)
	assertNotContains(t, got, "AGSRuntime.get_character")
}

func TestEmitter_T32_CharacterName_FromIdentifier(t *testing.T) {
	// Non-global receiver: identifier name is used as the character name string.
	got := emit(t, `function f() { cGuard.WalkTo(point.door); }`)
	assertContains(t, got, `AGSRuntime.get_character("c_guard").walk_to("door")`)
}

// -------------------------------------------------------------------
// All valid fixtures — no panic, non-empty output
// -------------------------------------------------------------------

func TestEmitter_ValidFixtures_NoPanic(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "..", "testdata", "valid", "*.agscript"))
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
				t.Fatalf("parse error: %v", parseErrs[0])
			}
			result, err := emitter.New().Emit(f)
			if err != nil {
				t.Fatalf("emit error: %v", err)
			}
			if result == nil {
				t.Fatal("nil result")
			}
		})
	}
}

// -------------------------------------------------------------------
// T-GS07 — global variable read/write
// -------------------------------------------------------------------

func TestEmitter_GlobalRead_EmitsGetGlobal(t *testing.T) {
	got := emit(t, `function f() { int x = global.score; }`)
	assertContains(t, got, `AGSRuntime.get_global("score")`)
}

func TestEmitter_GlobalAssign_EmitsSetGlobal(t *testing.T) {
	got := emit(t, `function f() { global.score = 10; }`)
	assertContains(t, got, `AGSRuntime.set_global("score", 10)`)
	assertNotContains(t, got, "global.score")
}

func TestEmitter_GlobalCompoundAdd_Expanded(t *testing.T) {
	got := emit(t, `function f() { global.score += 5; }`)
	assertContains(t, got, `AGSRuntime.set_global("score", AGSRuntime.get_global("score") + 5)`)
}

func TestEmitter_GlobalCompoundSub_Expanded(t *testing.T) {
	got := emit(t, `function f() { global.score -= 1; }`)
	assertContains(t, got, `AGSRuntime.set_global("score", AGSRuntime.get_global("score") - 1)`)
}

func TestEmitter_GlobalBoolAssign(t *testing.T) {
	got := emit(t, `function f() { global.door_unlocked = true; }`)
	assertContains(t, got, `AGSRuntime.set_global("door_unlocked", true)`)
}

func TestEmitter_GlobalInCondition(t *testing.T) {
	got := emit(t, `function f() { if (global.door_unlocked) { } }`)
	assertContains(t, got, `AGSRuntime.get_global("door_unlocked")`)
}

// -------------------------------------------------------------------
// T-GS09 — GoToRoom room transition
// -------------------------------------------------------------------

func TestEmitter_GoToRoom_MapsToLoadRoom(t *testing.T) {
	got := emit(t, `function f() { GoToRoom("library"); }`)
	assertContains(t, got, `AGSRuntime.load_room("library")`)
}

func TestEmitter_GoToRoom_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { GoToRoom("library"); }`)
	assertContains(t, got, `await AGSRuntime.load_room("library")`)
}

// -------------------------------------------------------------------
// T-GS05 — Say, Think, AddInventory, LoseInventory, HasInventory
// -------------------------------------------------------------------

func TestEmitter_Say_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { global.player.Say("Hello!"); }`)
	assertContains(t, got, `await AGSRuntime.get_character("player").say("Hello!")`)
}

func TestEmitter_Think_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { global.player.Think("Hmm..."); }`)
	assertContains(t, got, `await AGSRuntime.get_character("player").think("Hmm...")`)
}

func TestEmitter_AddInventory_NonBlocking(t *testing.T) {
	got := emit(t, `function f() { global.player.AddInventory("rusty_key"); }`)
	assertContains(t, got, `AGSRuntime.get_character("player").add_inventory("rusty_key")`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_LoseInventory_NonBlocking(t *testing.T) {
	got := emit(t, `function f() { global.player.LoseInventory("rusty_key"); }`)
	assertContains(t, got, `AGSRuntime.get_character("player").lose_inventory("rusty_key")`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_HasInventory_NonBlocking(t *testing.T) {
	got := emit(t, `function f() { if (global.player.HasInventory("rusty_key")) { } }`)
	assertContains(t, got, `AGSRuntime.get_character("player").has_inventory("rusty_key")`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_Say_IdentifierReceiver(t *testing.T) {
	got := emit(t, `function f() { cGuard.Say("Halt!"); }`)
	assertContains(t, got, `await AGSRuntime.get_character("c_guard").say("Halt!")`)
}

// -------------------------------------------------------------------
// T-GS06 — HideRoomItem, ShowRoomItem, item_interact handler
// -------------------------------------------------------------------

func TestEmitter_HideRoomItem_NonBlocking(t *testing.T) {
	got := emit(t, `function f() { HideRoomItem("old_chest"); }`)
	assertContains(t, got, `AGSRuntime.hide_room_item("old_chest")`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_ShowRoomItem_NonBlocking(t *testing.T) {
	got := emit(t, `function f() { ShowRoomItem("old_chest"); }`)
	assertContains(t, got, `AGSRuntime.show_room_item("old_chest")`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_ItemInteract_EmitsFunc(t *testing.T) {
	got := emit(t, `function item_interact(string name) {
    if (name == "old_chest") {
        global.player.AddInventory("rusty_key");
        HideRoomItem("old_chest");
    }
}`)
	assertContains(t, got, "func item_interact(name: String):")
	assertContains(t, got, `AGSRuntime.get_character("player").add_inventory("rusty_key")`)
	assertContains(t, got, `AGSRuntime.hide_room_item("old_chest")`)
}

// -------------------------------------------------------------------
// T-GS19 — SetPlayerControl, FadeIn, FadeOut, Wait
// -------------------------------------------------------------------

func TestEmitter_Wait_MapsToAGSCutscene(t *testing.T) {
	got := emit(t, `function f() { Wait(2.0); }`)
	assertContains(t, got, `AGSCutscene.wait(2.0)`)
}

func TestEmitter_Wait_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { Wait(2.0); }`)
	assertContains(t, got, `await AGSCutscene.wait(2.0)`)
}

func TestEmitter_FadeOut_MapsToAGSCutscene(t *testing.T) {
	got := emit(t, `function f() { FadeOut(); }`)
	assertContains(t, got, `AGSCutscene.fade_out()`)
}

func TestEmitter_FadeOut_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { FadeOut(0.5); }`)
	assertContains(t, got, `await AGSCutscene.fade_out(0.5)`)
}

func TestEmitter_FadeIn_MapsToAGSCutscene(t *testing.T) {
	got := emit(t, `function f() { FadeIn(); }`)
	assertContains(t, got, `AGSCutscene.fade_in()`)
}

func TestEmitter_FadeIn_IsBlocking(t *testing.T) {
	got := emit(t, `function f() { FadeIn(1.0); }`)
	assertContains(t, got, `await AGSCutscene.fade_in(1.0)`)
}

func TestEmitter_SetPlayerControl_MapsToAGSRuntime(t *testing.T) {
	got := emit(t, `function f() { SetPlayerControl(false); }`)
	assertContains(t, got, `AGSRuntime.set_player_control(false)`)
}

func TestEmitter_SetPlayerControl_NotBlocking(t *testing.T) {
	got := emit(t, `function f() { SetPlayerControl(false); }`)
	assertNotContains(t, got, "await ")
}

// -------------------------------------------------------------------
// T-GS11 — PlayMusic, StopMusic, PlaySound
// -------------------------------------------------------------------

func TestEmitter_PlayMusic_MapsToAGSRuntime(t *testing.T) {
	got := emit(t, `function f() { PlayMusic("theme_main"); }`)
	assertContains(t, got, `AGSRuntime.play_music("theme_main")`)
}

func TestEmitter_PlayMusic_NotBlocking(t *testing.T) {
	got := emit(t, `function f() { PlayMusic("theme_main"); }`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_StopMusic_MapsToAGSRuntime(t *testing.T) {
	got := emit(t, `function f() { StopMusic(); }`)
	assertContains(t, got, `AGSRuntime.stop_music()`)
}

func TestEmitter_StopMusic_NotBlocking(t *testing.T) {
	got := emit(t, `function f() { StopMusic(); }`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_PlaySound_MapsToAGSRuntime(t *testing.T) {
	got := emit(t, `function f() { PlaySound("door_creak"); }`)
	assertContains(t, got, `AGSRuntime.play_sound("door_creak")`)
}

func TestEmitter_PlaySound_NotBlocking(t *testing.T) {
	got := emit(t, `function f() { PlaySound("door_creak"); }`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_AudioSequence(t *testing.T) {
	// Typical room-load: start music, play a sound effect, stop music at end.
	src := `function room_Load() {
    PlayMusic("theme_main");
    PlaySound("door_creak");
    StopMusic();
}`
	got := emit(t, src)
	assertContains(t, got, `AGSRuntime.play_music("theme_main")`)
	assertContains(t, got, `AGSRuntime.play_sound("door_creak")`)
	assertContains(t, got, `AGSRuntime.stop_music()`)
	assertNotContains(t, got, "await ")
}

func TestEmitter_CutsceneSequence(t *testing.T) {
	// Full cutscene pattern: disable control → fade out → action → fade in → re-enable.
	src := `function room_Enter() {
    SetPlayerControl(false);
    FadeOut(0.5);
    global.player.Say("Darkness falls.");
    FadeIn(0.5);
    SetPlayerControl(true);
}`
	got := emit(t, src)
	assertContains(t, got, `AGSRuntime.set_player_control(false)`)
	assertContains(t, got, `await AGSCutscene.fade_out(0.5)`)
	assertContains(t, got, `await AGSRuntime.get_character("player").say("Darkness falls.")`)
	assertContains(t, got, `await AGSCutscene.fade_in(0.5)`)
	assertContains(t, got, `AGSRuntime.set_player_control(true)`)
	// SetPlayerControl is not blocking — no await before it.
	assertNotContains(t, got, `await AGSRuntime.set_player_control`)
}
