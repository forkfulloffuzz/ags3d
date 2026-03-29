package scene

import (
	"fmt"
	"strings"

	"github.com/ags3d/ag/internal/char"
)

// GenerateCharScene returns the .tscn text for a character.
//
// The generated scene is a standalone file that can be instantiated by the
// runtime at a SpawnPoint. It does not include a world position — that is set
// by the SpawnPoint at load time.
func GenerateCharScene(cd *char.CharData) string {
	switch cd.Type {
	case "2d":
		return generate2DCharScene(cd)
	default:
		return generate3DCharScene(cd)
	}
}

// --------------------------------------------------------------------------
// 3D character
// --------------------------------------------------------------------------

func generate3DCharScene(cd *char.CharData) string {
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

	return out.String()
}

// --------------------------------------------------------------------------
// 2D billboard character
// --------------------------------------------------------------------------

func generate2DCharScene(cd *char.CharData) string {
	var out strings.Builder

	rootName := toPascalCase(cd.Name)

	fmt.Fprintln(&out, "[gd_scene format=3]")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, `[ext_resource type="Script" path="res://.engine/runtime/ags_character.gd" id="CharScript"]`)

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
	if cd.DisplayName != "" {
		fmt.Fprintf(&out, "display_name = %q\n", cd.DisplayName)
	}
	fmt.Fprintln(&out, `script = ExtResource("CharScript")`)
	fmt.Fprintln(&out)

	// Sprite3D billboard
	spriteUID := nodeUID("/" + rootName + "/Sprite3D")
	fmt.Fprintf(&out, "[node name=\"Sprite3D\" type=\"Sprite3D\" parent=%q unique_id=%d]\n", rootName, spriteUID)
	fmt.Fprintln(&out, `billboard = 1`)
	if cd.SpriteSheet != "" {
		fmt.Fprintln(&out, `texture = ExtResource("SpriteSheet")`)
	}
	if cd.FramesPerAngle > 0 {
		fmt.Fprintf(&out, "hframes = %d\n", cd.FramesPerAngle)
	}
	if cd.SpriteAngles > 0 {
		fmt.Fprintf(&out, "vframes = %d\n", cd.SpriteAngles)
	}
	fmt.Fprintln(&out)

	// CollisionShape3D
	colUID := nodeUID("/" + rootName + "/CollisionShape3D")
	fmt.Fprintf(&out, "[node name=\"CollisionShape3D\" type=\"CollisionShape3D\" parent=%q unique_id=%d]\n", rootName, colUID)
	fmt.Fprintln(&out, `transform = Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0.9, 0)`)
	fmt.Fprintln(&out, `shape = SubResource("BodyShape")`)
	fmt.Fprintln(&out)

	return out.String()
}
