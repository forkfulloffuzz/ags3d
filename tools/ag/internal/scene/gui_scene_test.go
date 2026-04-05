package scene_test

import (
	"testing"

	"github.com/ags3d/ag/internal/gui"
	"github.com/ags3d/ag/internal/scene"
)

func genGUI(t *testing.T, src string) string {
	t.Helper()
	g, err := gui.ParseGUI("test.agui", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return scene.GenerateGUIScene(g)
}

// --------------------------------------------------------------------------
// Root CanvasLayer
// --------------------------------------------------------------------------

func TestGUISceneRootNode(t *testing.T) {
	out := genGUI(t, `GUI "main_hud" { layer = 10 }`)
	assertContains(t, out, "[gd_scene format=3]")
	assertContains(t, out, `type="Script" path="res://.engine/runtime/ags_gui.gd"`)
	assertContains(t, out, `[node name="MainHud" type="CanvasLayer"`)
	assertContains(t, out, `layer = 10`)
	assertContains(t, out, `metadata/ags_gui_name = "main_hud"`)
	assertContains(t, out, `script = ExtResource("GUIScript")`)
}

func TestGUISceneDefaultLayer(t *testing.T) {
	out := genGUI(t, `GUI "hud" {}`)
	assertContains(t, out, `layer = 1`)
}

// --------------------------------------------------------------------------
// InventoryBar → GridContainer
// --------------------------------------------------------------------------

func TestGUISceneInventoryBar(t *testing.T) {
	out := genGUI(t, `GUI "hud" {
		InventoryBar "inv_bar" {
			position  = (0, 0, bottom)
			item_size = (64, 64)
			columns   = 6
		}
	}`)
	assertContains(t, out, `[node name="InvBar" type="GridContainer" parent="Hud"`)
	assertContains(t, out, `anchors_preset = 12`) // PRESET_BOTTOM_WIDE
	assertContains(t, out, `columns = 6`)
	assertContains(t, out, `metadata/ags_widget = "InventoryBar"`)
	assertContains(t, out, `metadata/item_size = Vector2i(64, 64)`)
}

// --------------------------------------------------------------------------
// VerbBar → HBoxContainer + Button children
// --------------------------------------------------------------------------

func TestGUISceneVerbBar(t *testing.T) {
	out := genGUI(t, `GUI "hud" {
		VerbBar "verbs" {
			position = (0, 0, bottom_right)
			verbs    = ["Look", "Use", "Pick up"]
		}
	}`)
	assertContains(t, out, `[node name="Verbs" type="HBoxContainer" parent="Hud"`)
	assertContains(t, out, `anchors_preset = 3`) // PRESET_BOTTOM_RIGHT
	assertContains(t, out, `metadata/ags_widget = "VerbBar"`)
	// Button children
	assertContains(t, out, `[node name="Look" type="Button"`)
	assertContains(t, out, `text = "Look"`)
	assertContains(t, out, `[node name="PickUp" type="Button"`)
	assertContains(t, out, `text = "Pick up"`)
	assertContains(t, out, `metadata/ags_verb = "Pick up"`)
}

// --------------------------------------------------------------------------
// StatusLine → Label
// --------------------------------------------------------------------------

func TestGUISceneStatusLine(t *testing.T) {
	out := genGUI(t, `GUI "hud" {
		StatusLine "status" {
			position = (0, 0, top)
		}
	}`)
	assertContains(t, out, `[node name="Status" type="Label" parent="Hud"`)
	assertContains(t, out, `anchors_preset = 10`) // PRESET_TOP_WIDE
	assertContains(t, out, `metadata/ags_widget = "StatusLine"`)
}

func TestGUISceneStatusLineFont(t *testing.T) {
	out := genGUI(t, `GUI "hud" {
		StatusLine "status" {
			position = (0, 0, top)
			font     = "assets/fonts/main.ttf"
		}
	}`)
	assertContains(t, out, `type="FontFile" path="res://assets/fonts/main.ttf"`)
	assertContains(t, out, `label_settings/font`)
}

// --------------------------------------------------------------------------
// Determinism
// --------------------------------------------------------------------------

func TestGUISceneDeterministic(t *testing.T) {
	src := `GUI "main_hud" {
		layer = 5
		InventoryBar "inv" { position = (0, 0, bottom) }
	}`
	out1 := genGUI(t, src)
	out2 := genGUI(t, src)
	if out1 != out2 {
		t.Error("gui scene output is not deterministic")
	}
}
