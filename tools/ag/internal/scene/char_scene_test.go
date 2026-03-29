package scene_test

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/char"
	"github.com/ags3d/ag/internal/scene"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func genChar(t *testing.T, src string) string {
	t.Helper()
	cd, err := char.ParseChar("test.agchar", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return scene.GenerateCharScene(cd)
}

// --------------------------------------------------------------------------
// 3D character — no mesh (default capsule)
// --------------------------------------------------------------------------

func TestChar3DDefaultCapsule(t *testing.T) {
	out := genChar(t, `Character "player" { display_name = "Player" }`)

	assertContains(t, out, "[gd_scene format=3]")
	assertContains(t, out, `type="Script" path="res://.engine/runtime/ags_character_3d.gd"`)
	assertContains(t, out, `[sub_resource type="CapsuleShape3D" id="BodyShape"]`)
	assertContains(t, out, `[sub_resource type="CapsuleMesh" id="BodyMesh"]`)

	assertContains(t, out, `[node name="Player" type="AGSCharacter3D"`)
	assertContains(t, out, `character_name = "player"`)
	assertContains(t, out, `display_name = "Player"`)
	assertContains(t, out, `script = ExtResource("CharScript")`)

	assertContains(t, out, `[node name="MeshInstance3D" type="MeshInstance3D" parent="Player"`)
	assertContains(t, out, `mesh = SubResource("BodyMesh")`)

	assertContains(t, out, `[node name="CollisionShape3D" type="CollisionShape3D" parent="Player"`)
	assertContains(t, out, `shape = SubResource("BodyShape")`)

	// No .glb ext_resource
	assertNotContains(t, out, `type="PackedScene"`)
}

func TestChar3DNoDisplayName(t *testing.T) {
	out := genChar(t, `Character "hero" {}`)
	// display_name should not appear if not set
	assertNotContains(t, out, "display_name")
}

// --------------------------------------------------------------------------
// 3D character — with .glb mesh
// --------------------------------------------------------------------------

func TestChar3DWithMesh(t *testing.T) {
	src := `Character "hero" {
		mesh = "characters/hero/hero.glb"
		animations = {
			idle = "Idle"
			walk = "Walk"
		}
	}`
	out := genChar(t, src)

	assertContains(t, out, `[ext_resource type="PackedScene" path="res://characters/hero/hero.glb" id="CharMesh"]`)
	// No CapsuleMesh when mesh is set
	assertNotContains(t, out, `CapsuleMesh`)
	// Instances the sub-scene
	assertContains(t, out, `instance=ExtResource("CharMesh")`)
	// CollisionShape still present
	assertContains(t, out, `[sub_resource type="CapsuleShape3D" id="BodyShape"]`)
	assertContains(t, out, `[node name="CollisionShape3D"`)
}

// --------------------------------------------------------------------------
// 2D billboard character
// --------------------------------------------------------------------------

func TestChar2DFull(t *testing.T) {
	src := `Character "guard" {
		type             = "2d"
		display_name     = "Guard"
		sprite_sheet     = "assets/sprites/guard.png"
		sprite_angles    = 8
		frames_per_angle = 6
	}`
	out := genChar(t, src)

	assertContains(t, out, `type="Script" path="res://.engine/runtime/ags_character_2d.gd"`)
	assertContains(t, out, `type="Texture2D" path="res://assets/sprites/guard.png"`)
	assertContains(t, out, `[sub_resource type="CapsuleShape3D" id="BodyShape"]`)

	assertContains(t, out, `[node name="Guard" type="AGSCharacter2D"`)
	assertContains(t, out, `character_name = "guard"`)
	assertContains(t, out, `display_name = "Guard"`)

	assertContains(t, out, `[node name="Sprite3D" type="Sprite3D" parent="Guard"`)
	assertContains(t, out, `billboard = 1`)
	assertContains(t, out, `texture = ExtResource("SpriteSheet")`)
	assertContains(t, out, `hframes = 6`)
	assertContains(t, out, `vframes = 8`)

	assertContains(t, out, `[node name="CollisionShape3D" type="CollisionShape3D" parent="Guard"`)
	assertContains(t, out, `shape = SubResource("BodyShape")`)
}

func TestChar2DNoSpriteSheet(t *testing.T) {
	out := genChar(t, `Character "npc" { type = "2d" }`)
	// No texture ext_resource when sprite_sheet is not set
	assertNotContains(t, out, `type="Texture2D"`)
	assertNotContains(t, out, `texture =`)
	// Still has Sprite3D and CollisionShape3D
	assertContains(t, out, `type="Sprite3D"`)
	assertContains(t, out, `type="CollisionShape3D"`)
}

// --------------------------------------------------------------------------
// Determinism
// --------------------------------------------------------------------------

func TestCharSceneDeterministic(t *testing.T) {
	src := `Character "player" { display_name = "Player" }`
	out1 := genChar(t, src)
	out2 := genChar(t, src)
	if out1 != out2 {
		t.Error("char scene output is not deterministic")
	}
}

// --------------------------------------------------------------------------
// PascalCase root node name
// --------------------------------------------------------------------------

func TestCharPascalCaseRootName(t *testing.T) {
	cases := []struct{ name, want string }{
		{"player", `name="Player"`},
		{"guard_npc", `name="GuardNpc"`},
		{"big_boss", `name="BigBoss"`},
	}
	for _, c := range cases {
		src := `Character "` + c.name + `" {}`
		out := genChar(t, src)
		if !strings.Contains(out, c.want) {
			t.Errorf("for character %q: output does not contain %q\n%s", c.name, c.want, out)
		}
	}
}
