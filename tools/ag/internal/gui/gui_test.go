package gui_test

import (
	"testing"

	"github.com/ags3d/ag/internal/gui"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func mustParse(t *testing.T, src string) *gui.GUIData {
	t.Helper()
	g, err := gui.ParseGUI("test.agui", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return g
}

func mustFail(t *testing.T, src string, want string) {
	t.Helper()
	_, err := gui.ParseGUI("test.agui", src)
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !containsStr(err.Error(), want) {
		t.Fatalf("error %q does not contain %q", err.Error(), want)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// --------------------------------------------------------------------------
// Basic parsing
// --------------------------------------------------------------------------

func TestGUIName(t *testing.T) {
	g := mustParse(t, `GUI "main_hud" {}`)
	if g.Name != "main_hud" {
		t.Errorf("Name = %q, want %q", g.Name, "main_hud")
	}
}

func TestGUILayerDefault(t *testing.T) {
	g := mustParse(t, `GUI "hud" {}`)
	if g.Layer != 1 {
		t.Errorf("Layer = %d, want 1", g.Layer)
	}
}

func TestGUILayer(t *testing.T) {
	g := mustParse(t, `GUI "hud" { layer = 10 }`)
	if g.Layer != 10 {
		t.Errorf("Layer = %d, want 10", g.Layer)
	}
}

func TestGUINoWidgets(t *testing.T) {
	g := mustParse(t, `GUI "hud" {}`)
	if len(g.Widgets) != 0 {
		t.Errorf("Widgets = %d, want 0", len(g.Widgets))
	}
}

// --------------------------------------------------------------------------
// InventoryBar
// --------------------------------------------------------------------------

func TestInventoryBarDefaults(t *testing.T) {
	g := mustParse(t, `GUI "hud" {
		InventoryBar "inv" {}
	}`)
	if len(g.Widgets) != 1 {
		t.Fatalf("len(Widgets) = %d", len(g.Widgets))
	}
	w := g.Widgets[0]
	if w.Type != "InventoryBar" {
		t.Errorf("Type = %q", w.Type)
	}
	if w.Name != "inv" {
		t.Errorf("Name = %q", w.Name)
	}
	if w.Columns != 8 {
		t.Errorf("Columns = %d, want 8", w.Columns)
	}
	if w.ItemSizeW != 48 || w.ItemSizeH != 48 {
		t.Errorf("ItemSize = (%d, %d), want (48, 48)", w.ItemSizeW, w.ItemSizeH)
	}
}

func TestInventoryBarFull(t *testing.T) {
	g := mustParse(t, `GUI "hud" {
		InventoryBar "inv_bar" {
			position  = (0, 0, bottom)
			item_size = (64, 64)
			columns   = 10
		}
	}`)
	w := g.Widgets[0]
	if w.Anchor != "bottom" {
		t.Errorf("Anchor = %q", w.Anchor)
	}
	if w.Columns != 10 {
		t.Errorf("Columns = %d", w.Columns)
	}
	if w.ItemSizeW != 64 || w.ItemSizeH != 64 {
		t.Errorf("ItemSize = (%d, %d)", w.ItemSizeW, w.ItemSizeH)
	}
}

// --------------------------------------------------------------------------
// VerbBar
// --------------------------------------------------------------------------

func TestVerbBarVerbs(t *testing.T) {
	g := mustParse(t, `GUI "hud" {
		VerbBar "verbs" {
			position = (0, 0, bottom_right)
			verbs    = ["Look", "Use", "Pick up"]
		}
	}`)
	w := g.Widgets[0]
	if w.Type != "VerbBar" {
		t.Errorf("Type = %q", w.Type)
	}
	if w.Anchor != "bottom_right" {
		t.Errorf("Anchor = %q", w.Anchor)
	}
	if len(w.Verbs) != 3 {
		t.Fatalf("len(Verbs) = %d", len(w.Verbs))
	}
	if w.Verbs[0] != "Look" || w.Verbs[2] != "Pick up" {
		t.Errorf("Verbs = %v", w.Verbs)
	}
}

func TestVerbBarEmptyVerbs(t *testing.T) {
	g := mustParse(t, `GUI "hud" { VerbBar "v" { verbs = [] } }`)
	if len(g.Widgets[0].Verbs) != 0 {
		t.Errorf("Verbs = %v", g.Widgets[0].Verbs)
	}
}

// --------------------------------------------------------------------------
// StatusLine
// --------------------------------------------------------------------------

func TestStatusLineFont(t *testing.T) {
	g := mustParse(t, `GUI "hud" {
		StatusLine "status" {
			position = (0, 0, top)
			font     = "assets/fonts/main.ttf"
		}
	}`)
	w := g.Widgets[0]
	if w.Type != "StatusLine" {
		t.Errorf("Type = %q", w.Type)
	}
	if w.Font != "assets/fonts/main.ttf" {
		t.Errorf("Font = %q", w.Font)
	}
	if w.Anchor != "top" {
		t.Errorf("Anchor = %q", w.Anchor)
	}
}

func TestStatusLineNoFont(t *testing.T) {
	g := mustParse(t, `GUI "hud" { StatusLine "s" {} }`)
	if g.Widgets[0].Font != "" {
		t.Errorf("Font = %q, want empty", g.Widgets[0].Font)
	}
}

// --------------------------------------------------------------------------
// Multiple widgets
// --------------------------------------------------------------------------

func TestMultipleWidgets(t *testing.T) {
	src := `GUI "main_hud" {
		layer = 10
		InventoryBar "inv" { position = (0, 0, bottom) }
		VerbBar "verbs" { position = (0, 0, bottom_right) verbs = ["Look"] }
		StatusLine "status" { position = (0, 0, top) }
	}`
	g := mustParse(t, src)
	if len(g.Widgets) != 3 {
		t.Fatalf("len(Widgets) = %d, want 3", len(g.Widgets))
	}
	if g.Widgets[0].Type != "InventoryBar" {
		t.Errorf("widget 0 type = %q", g.Widgets[0].Type)
	}
	if g.Widgets[1].Type != "VerbBar" {
		t.Errorf("widget 1 type = %q", g.Widgets[1].Type)
	}
	if g.Widgets[2].Type != "StatusLine" {
		t.Errorf("widget 2 type = %q", g.Widgets[2].Type)
	}
}

// --------------------------------------------------------------------------
// Position tuple
// --------------------------------------------------------------------------

func TestPositionOffsets(t *testing.T) {
	g := mustParse(t, `GUI "hud" {
		StatusLine "s" { position = (20, 10, top_right) }
	}`)
	w := g.Widgets[0]
	if w.OffsetX != 20 || w.OffsetY != 10 {
		t.Errorf("Offset = (%d, %d), want (20, 10)", w.OffsetX, w.OffsetY)
	}
	if w.Anchor != "top_right" {
		t.Errorf("Anchor = %q", w.Anchor)
	}
}

// --------------------------------------------------------------------------
// Comments and whitespace
// --------------------------------------------------------------------------

func TestComments(t *testing.T) {
	src := `// A GUI file
GUI "hud" {
    layer = 5 // main layer
    # hash comment
    InventoryBar "inv" {}
}`
	g := mustParse(t, src)
	if g.Layer != 5 {
		t.Errorf("Layer = %d", g.Layer)
	}
	if len(g.Widgets) != 1 {
		t.Errorf("len(Widgets) = %d", len(g.Widgets))
	}
}

// --------------------------------------------------------------------------
// Error cases
// --------------------------------------------------------------------------

func TestErrNotGUI(t *testing.T) {
	mustFail(t, `Room "r" {}`, `expected 'GUI'`)
}

func TestErrMissingBrace(t *testing.T) {
	mustFail(t, `GUI "hud" {`, `missing '}'`)
}

func TestErrUnknownGUIProperty(t *testing.T) {
	mustFail(t, `GUI "hud" { fov = 90 }`, `unknown GUI property`)
}

func TestErrUnknownWidgetProperty(t *testing.T) {
	mustFail(t, `GUI "hud" { InventoryBar "inv" { zindex = 5 } }`, `unknown InventoryBar property`)
}

func TestErrUnterminatedWidgetBlock(t *testing.T) {
	mustFail(t, `GUI "hud" { InventoryBar "inv" {`, `unexpected end of file`)
}
