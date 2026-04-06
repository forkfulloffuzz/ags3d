package cut

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EmittedCutscene is the compiled JSON representation of one .agcut source file.
// Written to .engine/generated/cutscenes/<name>.json.
type EmittedCutscene struct {
	Source       string           `json:"source"`                  // relative path to the .agcut source
	Title        string           `json:"title"`
	Skip         string           `json:"skip,omitempty"`
	SaveBlock    bool             `json:"save_block"`
	Tags         []string         `json:"tags,omitempty"`
	Fallback     string           `json:"fallback,omitempty"`
	LocGroup     string           `json:"loc_group,omitempty"`
	VoiceSession string           `json:"voice_session,omitempty"`
	AudioScope   string           `json:"audio_scope"`
	DuckChannels string           `json:"duck_channels,omitempty"`
	DuckLevel    float64          `json:"duck_level"`
	DuckFade     float64          `json:"duck_fade"`
	DuckRestore  float64          `json:"duck_restore"`
	AutoDuck     bool             `json:"auto_duck,omitempty"`
	Sequence     []*EmittedCommand `json:"sequence"`
}

// EmittedCommand is a compiled <<command>> step.
type EmittedCommand struct {
	Name         string            `json:"name"`
	Positional   []string          `json:"positional,omitempty"`
	Params       map[string]string `json:"params,omitempty"`
	IsBlockOpen  bool              `json:"is_block_open,omitempty"`
	IsBlockClose bool              `json:"is_block_close,omitempty"`
}

// EmitCutscenes compiles a slice of validated CutsceneFiles and writes one JSON
// file per source into outputDir (typically .engine/generated/cutscenes/).
// Errors from individual files are collected and returned together.
func EmitCutscenes(files []*CutsceneFile, outputDir string) []error {
	var errs []error
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return []error{fmt.Errorf("cut emit: cannot create output directory: %w", err)}
	}
	for _, cf := range files {
		if err := emitCutscene(cf, outputDir); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// emitCutscene compiles one CutsceneFile and writes its JSON representation.
func emitCutscene(cf *CutsceneFile, outputDir string) error {
	ec := &EmittedCutscene{
		Source:       cf.Path,
		Title:        cf.Title,
		Skip:         cf.Skip,
		SaveBlock:    cf.SaveBlock,
		Tags:         cf.Tags,
		Fallback:     cf.Fallback,
		LocGroup:     cf.LocGroup,
		VoiceSession: cf.VoiceSession,
		AudioScope:   cf.AudioScope,
		DuckChannels: cf.DuckChannels,
		DuckLevel:    cf.DuckLevel,
		DuckFade:     cf.DuckFade,
		DuckRestore:  cf.DuckRestore,
		AutoDuck:     cf.AutoDuck,
	}

	for _, rc := range cf.Sequence {
		cmd := ParseCommand(rc)
		ec.Sequence = append(ec.Sequence, &EmittedCommand{
			Name:         rc.Name,
			Positional:   nonNilSlice(cmd.Positional),
			Params:       nonNilMap(cmd.Params),
			IsBlockOpen:  rc.IsBlockOpen,
			IsBlockClose: rc.IsBlockClose,
		})
	}

	base := filepath.Base(cf.Path)
	base = strings.TrimSuffix(base, ".agcut") + ".json"
	outPath := filepath.Join(outputDir, base)

	data, err := json.MarshalIndent(ec, "", "  ")
	if err != nil {
		return fmt.Errorf("cut emit: json marshal %s: %w", cf.Path, err)
	}
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("cut emit: write %s: %w", outPath, err)
	}
	return nil
}

// nonNilSlice returns nil for an empty slice (omitempty-friendly).
func nonNilSlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}

// nonNilMap returns nil for an empty map (omitempty-friendly).
func nonNilMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	return m
}
