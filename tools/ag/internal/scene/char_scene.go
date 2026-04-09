package scene

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ags3d/ag/internal/aganim"
	"github.com/ags3d/ag/internal/char"
)

// GenerateCharScene returns the .tscn text for a character.
//
// The generated scene is a standalone file that can be instantiated by the
// runtime at a SpawnPoint. It does not include a world position — that is set
// by the SpawnPoint at load time.
//
// If anim is non-nil the frame tags from the .aganim sidecar are injected as
// metadata/anim_frame_tags on the root node (T-CUT28).
func GenerateCharScene(cd *char.CharData, anim *aganim.AnimFile) string {
	switch cd.Type {
	case "2d":
		return generate2DCharScene(cd, anim)
	default:
		return generate3DCharScene(cd, anim)
	}
}

// --------------------------------------------------------------------------
// 3D character
// --------------------------------------------------------------------------

func generate3DCharScene(cd *char.CharData, anim *aganim.AnimFile) string {
	var out strings.Builder

	rootName := toPascalCase(cd.Name)

	fmt.Fprintln(&out, "[gd_scene format=3]")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, `[ext_resource type="Script" path="res://.engine/runtime/ags_character.gd" id="CharScript"]`)

	hasMesh := cd.Mesh != ""

	if hasMesh {
		fmt.Fprintf(&out, "[ext_resource type=\"PackedScene\" path=%q id=\"CharMesh\"]\n", "res://"+cd.Mesh)
	}

	// CapsuleShape3D for the CollisionShape (always present)
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, `[sub_resource type="CapsuleShape3D" id="BodyShape"]`)
	fmt.Fprintln(&out, `radius = 0.3`)
	fmt.Fprintln(&out, `height = 1.8`)
	fmt.Fprintln(&out)

	if !hasMesh {
		// Default capsule visual mesh
		fmt.Fprintln(&out, `[sub_resource type="CapsuleMesh" id="BodyMesh"]`)
		fmt.Fprintln(&out, `radius = 0.3`)
		fmt.Fprintln(&out, `height = 1.8`)
		fmt.Fprintln(&out)
	}

	// Root node
	rootUID := nodeUID("/" + rootName)
	fmt.Fprintf(&out, "[node name=%q type=\"AGSCharacter3D\" unique_id=%d]\n", rootName, rootUID)
	fmt.Fprintf(&out, "character_name = %q\n", cd.Name)
	fmt.Fprintln(&out, `visual_mode = "mesh"`)
	if len(cd.Animations) > 0 {
		keys := make([]string, 0, len(cd.Animations))
		for k := range cd.Animations {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&out, "anim_%s = %q\n", k, cd.Animations[k])
		}
	}
	if cd.DisplayName != "" {
		fmt.Fprintf(&out, "display_name = %q\n", cd.DisplayName)
	}
	fmt.Fprintln(&out, `script = ExtResource("CharScript")`)
	fmt.Fprintln(&out)

	if hasMesh {
		// Instance the .glb sub-scene (carries Skeleton3D + AnimationPlayer)
		meshUID := nodeUID("/" + rootName + "/Mesh")
		fmt.Fprintf(&out, "[node name=\"Mesh\" parent=%q instance=ExtResource(\"CharMesh\") unique_id=%d]\n", rootName, meshUID)
		fmt.Fprintln(&out)
	} else {
		// Default visual: capsule mesh centered at half-height
		meshUID := nodeUID("/" + rootName + "/MeshInstance3D")
		fmt.Fprintf(&out, "[node name=\"MeshInstance3D\" type=\"MeshInstance3D\" parent=%q unique_id=%d]\n", rootName, meshUID)
		fmt.Fprintln(&out, `transform = Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0.9, 0)`)
		fmt.Fprintln(&out, `mesh = SubResource("BodyMesh")`)
		fmt.Fprintln(&out)
	}

	// CollisionShape3D centered at half-height
	colUID := nodeUID("/" + rootName + "/CollisionShape3D")
	fmt.Fprintf(&out, "[node name=\"CollisionShape3D\" type=\"CollisionShape3D\" parent=%q unique_id=%d]\n", rootName, colUID)
	fmt.Fprintln(&out, `transform = Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0.9, 0)`)
	fmt.Fprintln(&out, `shape = SubResource("BodyShape")`)
	fmt.Fprintln(&out)

	// T-CUT28: inject frame tags metadata on the root node.
	if anim != nil && len(anim.Clips) > 0 {
		fmt.Fprintf(&out, "[node name=%q type=\"AGSCharacter3D\"]\n", rootName)
		fmt.Fprintf(&out, "metadata/anim_frame_tags = %s\n", anim.GDScriptLiteral())
		fmt.Fprintln(&out)
	}

	return out.String()
}

// --------------------------------------------------------------------------
// 2D billboard character
// --------------------------------------------------------------------------

func generate2DCharScene(cd *char.CharData, anim *aganim.AnimFile) string {
	var out strings.Builder

	rootName := toPascalCase(cd.Name)

	fmt.Fprintln(&out, "[gd_scene format=3]")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, `[ext_resource type="Script" path="res://.engine/runtime/ags_character.gd" id="CharScript"]`)
	fmt.Fprintln(&out, `[ext_resource type="Script" path="res://.engine/runtime/ags_billboard_controller.gd" id="BillboardController"]`)
	fmt.Fprintln(&out, `[ext_resource type="Script" path="res://.engine/runtime/ags_animation_player_2d.gd" id="AnimPlayer2D"]`)

	if cd.SpriteSheet != "" {
		fmt.Fprintf(&out, "[ext_resource type=\"Texture2D\" path=%q id=\"SpriteSheet\"]\n", "res://"+cd.SpriteSheet)
	}

	// CapsuleShape3D (always present)
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, `[sub_resource type="CapsuleShape3D" id="BodyShape"]`)
	fmt.Fprintln(&out, `radius = 0.3`)
	fmt.Fprintln(&out, `height = 1.8`)
	fmt.Fprintln(&out)

	// Root node
	rootUID := nodeUID("/" + rootName)
	fmt.Fprintf(&out, "[node name=%q type=\"AGSCharacter2D\" unique_id=%d]\n", rootName, rootUID)
	fmt.Fprintf(&out, "character_name = %q\n", cd.Name)
	fmt.Fprintln(&out, `visual_mode = "billboard"`)
	if cd.DisplayName != "" {
		fmt.Fprintf(&out, "display_name = %q\n", cd.DisplayName)
	}
	fmt.Fprintln(&out, `script = ExtResource("CharScript")`)
	fmt.Fprintln(&out)

	// Sprite3D billboard
	spriteUID := nodeUID("/" + rootName + "/Sprite3D")
	fmt.Fprintf(&out, "[node name=%q type=\"Sprite3D\" parent=%q unique_id=%d]\n", "Sprite3D", rootName, spriteUID)
	fmt.Fprintln(&out, `billboard = 1`)
	if cd.SpriteSheet != "" {
		fmt.Fprintln(&out, `texture = ExtResource("SpriteSheet")`)
	}
	// hframes = frames_per_angle (columns per direction row); vframes = sprite_angles (direction rows).
	if cd.FramesPerAngle > 0 {
		fmt.Fprintf(&out, "hframes = %d\n", cd.FramesPerAngle)
	}
	if cd.SpriteAngles > 0 {
		fmt.Fprintf(&out, "vframes = %d\n", cd.SpriteAngles)
	}
	fmt.Fprintln(&out)

	// AGSBillboardController — handles direction quantization + frame cycling.
	// NodePath to Sprite3D sibling is relative to parent (this node's parent is the root).
	billboardUID := nodeUID("/" + rootName + "/AGSBillboardController")
	fmt.Fprintf(&out, "[node name=%q type=\"Node\" parent=%q unique_id=%d]\n", "AGSBillboardController", rootName, billboardUID)
	fmt.Fprintf(&out, "script = ExtResource(\"BillboardController\")\n")
	fmt.Fprintf(&out, "sprite_angles = %d\n", cd.SpriteAngles)
	fmt.Fprintf(&out, "frames_per_angle = %d\n", cd.FramesPerAngle)
	fmt.Fprintf(&out, "sprite_locked = false\n")
	// sprite_path is relative to AGSBillboardController's parent (the root),
	// so "../Sprite3D" resolves to the sibling Sprite3D node.
	fmt.Fprintln(&out, `sprite_path = NodePath("../Sprite3D")`)
	fmt.Fprintln(&out)

	// AGSAnimationPlayer2D — handles state transitions (idle/walk/talk).
	// Depends on AGSBillboardController for per-frame row selection.
	animPlayerUID := nodeUID("/" + rootName + "/AGSAnimationPlayer2D")
	fmt.Fprintf(&out, "[node name=%q type=\"Node\" parent=%q unique_id=%d]\n", "AGSAnimationPlayer2D", rootName, animPlayerUID)
	fmt.Fprintf(&out, "script = ExtResource(\"AnimPlayer2D\")\n")
	fmt.Fprintf(&out, "frames_per_state = %d\n", cd.FramesPerAngle)
	// fps is not emitted — GDScript default (8.0) is used when absent.
	// controller_path is relative to AGSAnimationPlayer2D's parent (the root),
	// so "../AGSBillboardController" resolves to the sibling AGSBillboardController.
	fmt.Fprintln(&out, `controller_path = NodePath("../AGSBillboardController")`)
	fmt.Fprintln(&out)

	// CollisionShape3D
	colUID := nodeUID("/" + rootName + "/CollisionShape3D")
	fmt.Fprintf(&out, "[node name=%q type=\"CollisionShape3D\" parent=%q unique_id=%d]\n", "CollisionShape3D", rootName, colUID)
	fmt.Fprintln(&out, `transform = Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0.9, 0)`)
	fmt.Fprintln(&out, `shape = SubResource("BodyShape")`)
	fmt.Fprintln(&out)

	// T-CUT28: inject frame tags metadata on the root node.
	if anim != nil && len(anim.Clips) > 0 {
		fmt.Fprintf(&out, "[node name=%q type=\"AGSCharacter2D\"]\n", rootName)
		fmt.Fprintf(&out, "metadata/anim_frame_tags = %s\n", anim.GDScriptLiteral())
		fmt.Fprintln(&out)
	}

	return out.String()
}
