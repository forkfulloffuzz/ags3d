package scene

import (
	"fmt"
	"strings"

	"github.com/ags3d/ag/internal/gui"
)

// anchorPresets maps .agui anchor names to Godot 4 LayoutPreset integer values.
//
// Godot 4 Control.LayoutPreset:
//
//	0  PRESET_TOP_LEFT
//	1  PRESET_TOP_RIGHT
//	2  PRESET_BOTTOM_LEFT
//	3  PRESET_BOTTOM_RIGHT
//	8  PRESET_CENTER
//	10 PRESET_TOP_WIDE   (full width, anchored at top)
//	12 PRESET_BOTTOM_WIDE (full width, anchored at bottom)
var anchorPresets = map[string]int{
	"top":          10, // full-width top bar
	"top_left":      0,
	"top_right":     1,
	"bottom":       12, // full-width bottom bar
	"bottom_left":   2,
	"bottom_right":  3,
	"center":        8,
}

// GenerateGUIScene returns the .tscn text for a GUI definition.
//
// The root node is a CanvasLayer. Each widget becomes a child Control node:
//   - InventoryBar → GridContainer (columns property, slot size via theme override)
//   - VerbBar      → HBoxContainer (one Button per verb)
//   - StatusLine   → Label (optional font override)
//
// A GDScript at res://.engine/runtime/ags_gui.gd is attached to the root.
func GenerateGUIScene(g *gui.GUIData) string {
	var out strings.Builder

	rootName := toPascalCase(g.Name)
	rootUID := nodeUID("/" + rootName)

	fmt.Fprintln(&out, "[gd_scene format=3]")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, `[ext_resource type="Script" path="res://.engine/runtime/ags_gui.gd" id="GUIScript"]`)

	// Font ext_resources for StatusLine widgets.
	fontIDs := map[string]string{} // font path → resource id
	for i, w := range g.Widgets {
		if w.Type == "StatusLine" && w.Font != "" {
			if _, ok := fontIDs[w.Font]; !ok {
				id := fmt.Sprintf("Font_%d", i)
				fontIDs[w.Font] = id
				fmt.Fprintf(&out, "[ext_resource type=\"FontFile\" path=%q id=%q]\n", "res://"+w.Font, id)
			}
		}
	}

	fmt.Fprintln(&out)

	// Root CanvasLayer
	fmt.Fprintf(&out, "[node name=%q type=\"CanvasLayer\" unique_id=%d]\n", rootName, rootUID)
	fmt.Fprintf(&out, "layer = %d\n", g.Layer)
	fmt.Fprintf(&out, "metadata/ags_gui_name = %q\n", g.Name)
	fmt.Fprintln(&out, `script = ExtResource("GUIScript")`)
	fmt.Fprintln(&out)

	// Child widgets
	for _, w := range g.Widgets {
		childName := toPascalCase(w.Name)
		childUID := nodeUID("/" + rootName + "/" + childName)
		preset := anchorPresets[w.Anchor]
		// Unknown anchors fall back to 0 (top-left).

		switch w.Type {
		case "InventoryBar":
			fmt.Fprintf(&out, "[node name=%q type=\"GridContainer\" parent=%q unique_id=%d]\n",
				childName, rootName, childUID)
			fmt.Fprintf(&out, "anchors_preset = %d\n", preset)
			if w.OffsetX != 0 || w.OffsetY != 0 {
				fmt.Fprintf(&out, "offset_left = %d\n", w.OffsetX)
				fmt.Fprintf(&out, "offset_top = %d\n", w.OffsetY)
			}
			fmt.Fprintf(&out, "columns = %d\n", w.Columns)
			fmt.Fprintf(&out, "metadata/ags_widget = \"InventoryBar\"\n")
			fmt.Fprintf(&out, "metadata/item_size = Vector2i(%d, %d)\n", w.ItemSizeW, w.ItemSizeH)
			fmt.Fprintln(&out)

		case "VerbBar":
			fmt.Fprintf(&out, "[node name=%q type=\"HBoxContainer\" parent=%q unique_id=%d]\n",
				childName, rootName, childUID)
			fmt.Fprintf(&out, "anchors_preset = %d\n", preset)
			if w.OffsetX != 0 || w.OffsetY != 0 {
				fmt.Fprintf(&out, "offset_left = %d\n", w.OffsetX)
				fmt.Fprintf(&out, "offset_top = %d\n", w.OffsetY)
			}
			fmt.Fprintf(&out, "metadata/ags_widget = \"VerbBar\"\n")
			fmt.Fprintln(&out)
			// One Button child per verb.
			for _, verb := range w.Verbs {
				btnName := toPascalCase(strings.ReplaceAll(verb, " ", "_"))
				btnUID := nodeUID("/" + rootName + "/" + childName + "/" + btnName)
				fmt.Fprintf(&out, "[node name=%q type=\"Button\" parent=%q unique_id=%d]\n",
					btnName, rootName+"/"+childName, btnUID)
				fmt.Fprintf(&out, "text = %q\n", verb)
				fmt.Fprintf(&out, "metadata/ags_verb = %q\n", verb)
				fmt.Fprintln(&out)
			}

		case "StatusLine":
			fmt.Fprintf(&out, "[node name=%q type=\"Label\" parent=%q unique_id=%d]\n",
				childName, rootName, childUID)
			fmt.Fprintf(&out, "anchors_preset = %d\n", preset)
			if w.OffsetX != 0 || w.OffsetY != 0 {
				fmt.Fprintf(&out, "offset_left = %d\n", w.OffsetX)
				fmt.Fprintf(&out, "offset_top = %d\n", w.OffsetY)
			}
			fmt.Fprintf(&out, "metadata/ags_widget = \"StatusLine\"\n")
			if id, ok := fontIDs[w.Font]; ok {
				fmt.Fprintf(&out, "label_settings = LabelSettings.new()\n")
				fmt.Fprintf(&out, "label_settings/font = ExtResource(%q)\n", id)
			}
			fmt.Fprintln(&out)
		}
	}

	return out.String()
}
