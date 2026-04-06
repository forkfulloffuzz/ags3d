package char_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/char"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func mustParse(t *testing.T, src string) *char.CharData {
	t.Helper()
	cd, err := char.ParseChar("test.agchar", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return cd
}

func mustFail(t *testing.T, src, wantSubstr string) {
	t.Helper()
	_, err := char.ParseChar("test.agchar", src)
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantSubstr)
	}
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("expected error containing %q, got: %v", wantSubstr, err)
	}
}

// --------------------------------------------------------------------------
// Minimal / defaults
// --------------------------------------------------------------------------

func TestMinimalChar(t *testing.T) {
	cd := mustParse(t, `Character "player" {}`)
	if cd.Name != "player" {
		t.Errorf("Name = %q", cd.Name)
	}
	if cd.Type != "3d" {
		t.Errorf("Type = %q, want default \"3d\"", cd.Type)
	}
}

func TestDisplayName(t *testing.T) {
	cd := mustParse(t, `Character "npc" { display_name = "The Guard" }`)
	if cd.DisplayName != "The Guard" {
		t.Errorf("DisplayName = %q", cd.DisplayName)
	}
}

// --------------------------------------------------------------------------
// 3D character
// --------------------------------------------------------------------------

func TestType3D(t *testing.T) {
	cd := mustParse(t, `Character "hero" { type = "3d" }`)
	if cd.Type != "3d" {
		t.Errorf("Type = %q", cd.Type)
	}
}

func TestMesh(t *testing.T) {
	cd := mustParse(t, `Character "hero" { mesh = "characters/hero/hero.glb" }`)
	if cd.Mesh != "characters/hero/hero.glb" {
		t.Errorf("Mesh = %q", cd.Mesh)
	}
}

func TestAnimations(t *testing.T) {
	src := `Character "hero" {
		animations = {
			idle = "Idle"
			walk = "Walk"
			talk = "Talk"
		}
	}`
	cd := mustParse(t, src)
	if cd.Animations["idle"] != "Idle" {
		t.Errorf("Animations[idle] = %q", cd.Animations["idle"])
	}
	if cd.Animations["walk"] != "Walk" {
		t.Errorf("Animations[walk] = %q", cd.Animations["walk"])
	}
	if cd.Animations["talk"] != "Talk" {
		t.Errorf("Animations[talk] = %q", cd.Animations["talk"])
	}
}

func TestAnimationsEmpty(t *testing.T) {
	cd := mustParse(t, `Character "hero" { animations = {} }`)
	if len(cd.Animations) != 0 {
		t.Errorf("Animations len = %d, want 0", len(cd.Animations))
	}
}

func TestFull3DChar(t *testing.T) {
	src := `Character "player" {
		display_name = "Player"
		type         = "3d"
		mesh         = "characters/player/player.glb"
		animations = {
			idle = "Idle"
			walk = "Walk"
			talk = "Talk"
		}
	}`
	cd := mustParse(t, src)
	if cd.Name != "player" {
		t.Errorf("Name = %q", cd.Name)
	}
	if cd.DisplayName != "Player" {
		t.Errorf("DisplayName = %q", cd.DisplayName)
	}
	if cd.Type != "3d" {
		t.Errorf("Type = %q", cd.Type)
	}
	if cd.Mesh != "characters/player/player.glb" {
		t.Errorf("Mesh = %q", cd.Mesh)
	}
	if len(cd.Animations) != 3 {
		t.Errorf("Animations len = %d", len(cd.Animations))
	}
}

// --------------------------------------------------------------------------
// 2D character
// --------------------------------------------------------------------------

func TestType2D(t *testing.T) {
	cd := mustParse(t, `Character "npc" { type = "2d" }`)
	if cd.Type != "2d" {
		t.Errorf("Type = %q", cd.Type)
	}
}

func TestSpriteSheet(t *testing.T) {
	cd := mustParse(t, `Character "npc" { sprite_sheet = "assets/sprites/guard.png" }`)
	if cd.SpriteSheet != "assets/sprites/guard.png" {
		t.Errorf("SpriteSheet = %q", cd.SpriteSheet)
	}
}

func TestSpriteAnglesValid(t *testing.T) {
	for _, n := range []int{1, 4, 8} {
		src := fmt.Sprintf(`Character "npc" { sprite_angles = %d }`, n)
		cd := mustParse(t, src)
		if cd.SpriteAngles != n {
			t.Errorf("SpriteAngles = %d, want %d", cd.SpriteAngles, n)
		}
	}
}

func TestFrameSize(t *testing.T) {
	cd := mustParse(t, `Character "npc" { frame_size = (64, 128) }`)
	if cd.FrameSize[0] != 64 || cd.FrameSize[1] != 128 {
		t.Errorf("FrameSize = %v", cd.FrameSize)
	}
}

func TestFramesPerAngle(t *testing.T) {
	cd := mustParse(t, `Character "npc" { frames_per_angle = 6 }`)
	if cd.FramesPerAngle != 6 {
		t.Errorf("FramesPerAngle = %d", cd.FramesPerAngle)
	}
}

func TestFull2DChar(t *testing.T) {
	src := `Character "guard" {
		display_name     = "Guard"
		type             = "2d"
		sprite_sheet     = "assets/sprites/guard.png"
		sprite_angles    = 8
		frame_size       = (64, 128)
		frames_per_angle = 6
	}`
	cd := mustParse(t, src)
	if cd.Type != "2d" {
		t.Errorf("Type = %q", cd.Type)
	}
	if cd.SpriteSheet != "assets/sprites/guard.png" {
		t.Errorf("SpriteSheet = %q", cd.SpriteSheet)
	}
	if cd.SpriteAngles != 8 {
		t.Errorf("SpriteAngles = %d", cd.SpriteAngles)
	}
	if cd.FrameSize != [2]int{64, 128} {
		t.Errorf("FrameSize = %v", cd.FrameSize)
	}
	if cd.FramesPerAngle != 6 {
		t.Errorf("FramesPerAngle = %d", cd.FramesPerAngle)
	}
}

// --------------------------------------------------------------------------
// Comments
// --------------------------------------------------------------------------

func TestSlashComments(t *testing.T) {
	src := `// top comment
Character "hero" {
	// inline
	display_name = "Hero" // trailing
}`
	cd := mustParse(t, src)
	if cd.DisplayName != "Hero" {
		t.Errorf("DisplayName = %q", cd.DisplayName)
	}
}

func TestHashComments(t *testing.T) {
	src := `# top comment
Character "hero" {
	# inline
	display_name = "Hero" # trailing
}`
	cd := mustParse(t, src)
	if cd.DisplayName != "Hero" {
		t.Errorf("DisplayName = %q", cd.DisplayName)
	}
}

// --------------------------------------------------------------------------
// Error cases
// --------------------------------------------------------------------------

func TestErrNotCharacter(t *testing.T) {
	mustFail(t, `Room "r" {}`, `expected 'Character'`)
}

func TestErrInvalidType(t *testing.T) {
	mustFail(t, `Character "x" { type = "puppet" }`, `type must be "3d" or "2d"`)
}

func TestVisualModeMesh(t *testing.T) {
	cd := mustParse(t, `Character "hero" { visual_mode = "mesh" }`)
	if cd.Type != "3d" {
		t.Errorf("visual_mode=\"mesh\" should set Type=\"3d\", got %q", cd.Type)
	}
}

func TestVisualModeBillboard(t *testing.T) {
	cd := mustParse(t, `Character "guard" { visual_mode = "billboard" }`)
	if cd.Type != "2d" {
		t.Errorf("visual_mode=\"billboard\" should set Type=\"2d\", got %q", cd.Type)
	}
}

func TestErrInvalidVisualMode(t *testing.T) {
	mustFail(t, `Character "x" { visual_mode = "puppet" }`, `visual_mode must be "mesh" or "billboard"`)
}

func TestErrSpriteAnglesInvalid(t *testing.T) {
	mustFail(t, `Character "x" { sprite_angles = 3 }`, `sprite_angles must be 1, 4, or 8`)
}

func TestErrUnknownProperty(t *testing.T) {
	mustFail(t, `Character "x" { fov = "wide" }`, `unknown Character property`)
}

func TestErrUnterminatedAnimations(t *testing.T) {
	mustFail(t, `Character "x" { animations = { idle = "Idle" }`, "unexpected end of file")
}

func TestErrMissingBrace(t *testing.T) {
	mustFail(t, `Character "x" {`, `missing '}'`)
}

func TestErrLineNumbers(t *testing.T) {
	src := "Character \"x\" {\n\n\n    bad_key = \"v\"\n}"
	_, err := char.ParseChar("test.agchar", src)
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*char.ParseError)
	if !ok {
		t.Fatalf("expected *char.ParseError, got %T", err)
	}
	if pe.Line != 4 {
		t.Errorf("error line = %d, want 4", pe.Line)
	}
}

func TestErrTrailingContent(t *testing.T) {
	mustFail(t, `Character "x" {} extra`, `unexpected content after Character block`)
}

// --------------------------------------------------------------------------
// [dialogue] block — T-DLG10
// --------------------------------------------------------------------------

func TestDialogue_Roots(t *testing.T) {
	cd := mustParse(t, `Character "guard" {
		dialogue {
			roots = ["guard_greeting", "guard_farewell"]
		}
	}`)
	if cd.Dialogue == nil {
		t.Fatal("Dialogue is nil")
	}
	if len(cd.Dialogue.Roots) != 2 || cd.Dialogue.Roots[0] != "guard_greeting" || cd.Dialogue.Roots[1] != "guard_farewell" {
		t.Errorf("Roots = %v", cd.Dialogue.Roots)
	}
}

func TestDialogue_InheritsGlobalsDefault(t *testing.T) {
	cd := mustParse(t, `Character "guard" { dialogue {} }`)
	if !cd.Dialogue.InheritsGlobals {
		t.Error("InheritsGlobals should default to true")
	}
}

func TestDialogue_InheritsGlobalsFalse(t *testing.T) {
	cd := mustParse(t, `Character "guard" { dialogue { inherits_globals = false } }`)
	if cd.Dialogue.InheritsGlobals {
		t.Error("InheritsGlobals should be false")
	}
}

func TestDialogue_SuppressGlobals(t *testing.T) {
	cd := mustParse(t, `Character "guard" { dialogue { suppress_globals = ["global_help"] } }`)
	if len(cd.Dialogue.SuppressGlobals) != 1 || cd.Dialogue.SuppressGlobals[0] != "global_help" {
		t.Errorf("SuppressGlobals = %v", cd.Dialogue.SuppressGlobals)
	}
}

func TestDialogue_NilWhenAbsent(t *testing.T) {
	cd := mustParse(t, `Character "guard" {}`)
	if cd.Dialogue != nil {
		t.Error("Dialogue should be nil when block is absent")
	}
}

func TestDialogue_ValidateRefsOK(t *testing.T) {
	cd := mustParse(t, `Character "guard" { dialogue { roots = ["guard_greeting"] } }`)
	errs := cd.ValidateDialogueRefs(map[string]bool{"guard_greeting": true})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestDialogue_ValidateRefsMissing(t *testing.T) {
	cd := mustParse(t, `Character "guard" { dialogue { roots = ["guard_greeting"] } }`)
	errs := cd.ValidateDialogueRefs(map[string]bool{})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestDialogue_UnknownField(t *testing.T) {
	mustFail(t, `Character "x" { dialogue { unknown_field = "v" } }`, `unknown dialogue field`)
}
