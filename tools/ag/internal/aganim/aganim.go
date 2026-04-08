// Package aganim parses .aganim sidecar files produced by the Blender
// frame-tag exporter (T-CUT27).
//
// .aganim files are JSON with the structure:
//
//	{
//	  "character": "player",
//	  "clips": [
//	    {
//	      "name": "Walk",
//	      "frame_tags": [
//	        {"name": "footstep_left",  "frame": 12},
//	        {"name": "footstep_right", "frame": 24}
//	      ]
//	    }
//	  ]
//	}
//
// Each clip in "clips" corresponds to one NLA track / GLTF animation clip.
// frame_tags within each clip are sorted by frame ascending by the exporter.
package aganim

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// FrameTag is a single named tag at a specific frame number (1-based).
type FrameTag struct {
	Name  string `json:"name"`
	Frame int    `json:"frame"`
}

// Clip is one animation clip with its associated frame tags.
type Clip struct {
	Name      string     `json:"name"`
	FrameTags []FrameTag `json:"frame_tags"`
}

// AnimFile is the parsed content of a .aganim sidecar file.
type AnimFile struct {
	Character string `json:"character"`
	Clips     []Clip `json:"clips"`
}

// ParseFile reads and parses the .aganim file at path.
func ParseFile(path string) (*AnimFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var af AnimFile
	if err := json.Unmarshal(data, &af); err != nil {
		return nil, err
	}
	return &af, nil
}

// SidecarPath returns the .aganim sidecar path for a .glb file path.
// E.g. "characters/player/player.glb" → "characters/player/player.aganim".
func SidecarPath(glbPath string) string {
	ext := filepath.Ext(glbPath)
	return strings.TrimSuffix(glbPath, ext) + ".aganim"
}

// GDScriptLiteral returns a GDScript dictionary literal representing all
// clip frame tags, suitable for embedding in a .tscn metadata property.
//
// Output format (one line):
//
//	{"Walk": [{"frame": 12, "name": "footstep_left"}], "Idle": []}
func (af *AnimFile) GDScriptLiteral() string {
	if af == nil || len(af.Clips) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, clip := range af.Clips {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(`"`)
		b.WriteString(escapeGDString(clip.Name))
		b.WriteString(`": [`)
		for j, tag := range clip.FrameTags {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(`{"frame": `)
			writeInt(&b, tag.Frame)
			b.WriteString(`, "name": "`)
			b.WriteString(escapeGDString(tag.Name))
			b.WriteString(`"}`)
		}
		b.WriteByte(']')
	}
	b.WriteByte('}')
	return b.String()
}

func escapeGDString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func writeInt(b *strings.Builder, n int) {
	if n == 0 {
		b.WriteByte('0')
		return
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		b.WriteByte('-')
	}
	for i := len(buf) - 1; i >= 0; i-- {
		b.WriteByte(buf[i])
	}
}
