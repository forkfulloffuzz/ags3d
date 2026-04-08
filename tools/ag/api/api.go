// Package api is the public facade for the AGS-spirit transpiler pipeline.
//
// It exists so external binaries (e.g. tools/agui) can call into the
// transpiler without depending on internal/* packages, which Go's module
// system prohibits across module boundaries.
package api

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ags3d/ag/internal/char"
	"github.com/ags3d/ag/internal/emitter"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/room"
	"github.com/ags3d/ag/internal/scene"
	"github.com/ags3d/ag/internal/scanner"
	"github.com/ags3d/ag/internal/validate"
	"github.com/ags3d/ag/internal/viz"
)

// -------------------------------------------------------------------
// Project
// -------------------------------------------------------------------

// Manifest holds the parsed game.agp metadata.
type Manifest struct {
	Root           string
	Name           string
	StartRoom      string
	StartCharacter string
	RenderingMode  string
}

// LoadProject finds and loads game.agp from root.
func LoadProject(root string) (Manifest, error) {
	m, err := project.Load(root)
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{
		Root:           root,
		Name:           m.Project.Name,
		StartRoom:      m.Project.StartRoom,
		StartCharacter: m.Project.StartCharacter,
		RenderingMode:  m.Settings.RenderingMode,
	}, nil
}

// SourceFile is a discovered adventure-game source file.
type SourceFile struct {
	Path string
	Rel  string
	Ext  string
}

// ScanProject returns all source files in root (requires game.agp).
func ScanProject(root string) ([]SourceFile, error) {
	files, err := project.Scan(root)
	if err != nil {
		return nil, err
	}
	out := make([]SourceFile, len(files))
	for i, f := range files {
		out[i] = SourceFile{Path: f.Path, Rel: f.Rel, Ext: f.Ext}
	}
	return out, nil
}

// ScanFolder returns all AG source files under any directory (no game.agp needed).
func ScanFolder(root string) ([]SourceFile, error) {
	files, err := project.Scan(root)
	if err != nil {
		return nil, err
	}
	out := make([]SourceFile, len(files))
	for i, f := range files {
		out[i] = SourceFile{Path: f.Path, Rel: f.Rel, Ext: f.Ext}
	}
	return out, nil
}

// -------------------------------------------------------------------
// Build
// -------------------------------------------------------------------

// Build runs the full build pipeline over changed files in root.
// logFn receives each log line as it is produced (may be called from a
// goroutine — the caller is responsible for any synchronisation needed).
func Build(root string, logFn func(kind, msg string)) error {
	files, err := project.Scan(root)
	if err != nil {
		return err
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}
	changed := project.Changed(files, manifest)
	if len(changed) == 0 {
		logFn("info", "ag build: nothing to do (no changed source files)")
		return nil
	}
	logFn("info", "ag build: processing "+itoa(len(changed))+" file(s)")

	generatedDir := filepath.Join(root, ".engine", "generated")
	_ = os.MkdirAll(generatedDir, 0755)

	em := emitter.New()
	var errs []error

	for _, src := range changed {
		data, err := os.ReadFile(src.Path)
		if err != nil {
			logFn("error", err.Error())
			errs = append(errs, err)
			continue
		}
		s := scanner.New(src.Rel, string(data))
		p := parser.New(s)
		f, parseErrs := p.Parse(src.Rel)
		for _, pe := range parseErrs {
			logFn("error", pe.Error())
			errs = append(errs, pe)
		}
		if len(parseErrs) > 0 {
			continue
		}
		result, err := em.Emit(f)
		if err != nil {
			logFn("error", src.Rel+": "+err.Error())
			errs = append(errs, err)
			continue
		}
		outPath := filepath.Join(generatedDir, src.Rel+".gd")
		_ = os.MkdirAll(filepath.Dir(outPath), 0755)
		if err := os.WriteFile(outPath, []byte(result.GDScript), 0644); err != nil {
			logFn("error", err.Error())
			errs = append(errs, err)
			continue
		}
		project.RecordMtimes([]project.SourceFile{src}, manifest)
		logFn("info", "  "+src.Rel+" → .engine/generated/"+src.Rel+".gd")
	}
	_ = project.SaveManifest(root, manifest)

	if len(errs) > 0 {
		logFn("info", "ag build: "+itoa(len(errs))+" error(s)")
	} else {
		logFn("info", "ag build: done")
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// -------------------------------------------------------------------
// Transpile — full per-file pipeline result
// -------------------------------------------------------------------

// TranspileResult holds every pipeline stage's output for a single file.
type TranspileResult struct {
	Source    string
	Tokens    string
	ASTText   string
	ASTDot    string
	Symbols   string
	SymDot    string
	Blocking  string
	GDScript  string
	EmitView  string    // side-by-side AGS-spirit ↔ GDScript (viz emit)
	SourceMap [][3]any
	Errors    []string
}

// TranspileFile runs all pipeline stages on file and returns each stage's output.
func TranspileFile(file, src string) TranspileResult {
	res := TranspileResult{Source: src}
	var buf strings.Builder

	buf.Reset(); viz.Tokens(&buf, file, src);   res.Tokens = buf.String()
	buf.Reset(); viz.AST(&buf, file, src);      res.ASTText = buf.String()
	buf.Reset(); viz.ASTDot(&buf, file, src);   res.ASTDot = buf.String()
	buf.Reset(); viz.Symbols(&buf, file, src);  res.Symbols = buf.String()
	buf.Reset(); viz.SymbolsDot(&buf, file, src); res.SymDot = buf.String()
	buf.Reset(); viz.Blocking(&buf, file, src); res.Blocking = buf.String()
	buf.Reset(); viz.Emit(&buf, file, src);     res.EmitView = buf.String()

	// Parse + emit to get GDScript + SourceMap directly.
	s := scanner.New(file, src)
	p := parser.New(s)
	f, parseErrs := p.Parse(file)
	if len(parseErrs) > 0 {
		for _, pe := range parseErrs {
			res.Errors = append(res.Errors, pe.Error())
		}
		return res
	}
	emitResult, err := emitter.New().Emit(f)
	if err != nil {
		res.Errors = append(res.Errors, err.Error())
		return res
	}
	res.GDScript = emitResult.GDScript
	res.SourceMap = emitResult.SourceMap
	return res
}

// -------------------------------------------------------------------
// Individual viz stages
// -------------------------------------------------------------------

func VizTokens(file, src string) string     { return vizRun(file, src, viz.Tokens) }
func VizAST(file, src string) string        { return vizRun(file, src, viz.AST) }
func VizASTDot(file, src string) string     { return vizRun(file, src, viz.ASTDot) }
func VizSymbols(file, src string) string    { return vizRun(file, src, viz.Symbols) }
func VizSymbolsDot(file, src string) string { return vizRun(file, src, viz.SymbolsDot) }
func VizBlocking(file, src string) string   { return vizRun(file, src, viz.Blocking) }
func VizEmit(file, src string) string       { return vizRun(file, src, viz.Emit) }

func vizRun(file, src string, fn func(io.Writer, string, string)) string {
	var buf strings.Builder
	fn(&buf, file, src)
	return buf.String()
}

// -------------------------------------------------------------------
// Room / Char parsing and scene generation
// -------------------------------------------------------------------

// Vec3 is an XYZ coordinate.
type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Vec2 is an XZ coordinate (horizontal plane).
type Vec2 struct {
	X float64 `json:"x"`
	Z float64 `json:"z"`
}

// ParsedCamera is one Camera block from a .agroom.
type ParsedCamera struct {
	Name     string `json:"name"`
	Position Vec3   `json:"position"`
	LookAt   Vec3   `json:"lookAt"`
}

// ParsedPoint is one Point block.
type ParsedPoint struct {
	Name     string `json:"name"`
	Position Vec3   `json:"position"`
}

// ParsedWalkable is one WalkableSurface block.
type ParsedWalkable struct {
	Name     string `json:"name"`
	Position Vec3   `json:"position"`
	Size     Vec2   `json:"size"`
}

// ParsedBlocker is one BlockerVolume block.
type ParsedBlocker struct {
	Name     string `json:"name"`
	Position Vec3   `json:"position"`
	Size     Vec3   `json:"size"`
}

// ParsedSpawnPoint is one SpawnPoint block.
type ParsedSpawnPoint struct {
	Name      string `json:"name"`
	Character string `json:"character,omitempty"`
	Position  Vec3   `json:"position"`
}

// ParsedHotspot is one Hotspot block.
type ParsedHotspot struct {
	Name     string `json:"name"`
	Position Vec3   `json:"position"`
	Size     Vec3   `json:"size"`
}

// ParsedRoom is the result of parsing a .agroom file.
type ParsedRoom struct {
	Name             string             `json:"name"`
	InitialCamera    string             `json:"initialCamera,omitempty"`
	Cameras          []ParsedCamera     `json:"cameras,omitempty"`
	Points           []ParsedPoint      `json:"points,omitempty"`
	WalkableSurfaces []ParsedWalkable   `json:"walkableSurfaces,omitempty"`
	BlockerVolumes   []ParsedBlocker    `json:"blockerVolumes,omitempty"`
	SpawnPoints      []ParsedSpawnPoint `json:"spawnPoints,omitempty"`
	Hotspots         []ParsedHotspot    `json:"hotspots,omitempty"`
	Error            string             `json:"error,omitempty"`
}

// ParseRoom parses .agroom source text and returns the structured result.
func ParseRoom(filename, src string) ParsedRoom {
	rd, err := room.ParseRoom(filename, src)
	if err != nil {
		return ParsedRoom{Error: err.Error()}
	}
	pr := ParsedRoom{
		Name:          rd.Name,
		InitialCamera: rd.InitialCamera,
	}
	for _, c := range rd.Cameras {
		pr.Cameras = append(pr.Cameras, ParsedCamera{
			Name:     c.Name,
			Position: Vec3{c.Position.X, c.Position.Y, c.Position.Z},
			LookAt:   Vec3{c.LookAt.X, c.LookAt.Y, c.LookAt.Z},
		})
	}
	for _, p := range rd.Points {
		pr.Points = append(pr.Points, ParsedPoint{
			Name:     p.Name,
			Position: Vec3{p.Position.X, p.Position.Y, p.Position.Z},
		})
	}
	for _, w := range rd.WalkableSurfaces {
		pr.WalkableSurfaces = append(pr.WalkableSurfaces, ParsedWalkable{
			Name:     w.Name,
			Position: Vec3{w.Offset.X, w.Offset.Y, w.Offset.Z},
			Size:     Vec2{w.Size.X, w.Size.Z},
		})
	}
	for _, b := range rd.BlockerVolumes {
		pr.BlockerVolumes = append(pr.BlockerVolumes, ParsedBlocker{
			Name:     b.Name,
			Position: Vec3{b.Position.X, b.Position.Y, b.Position.Z},
			Size:     Vec3{b.Size.X, b.Size.Y, b.Size.Z},
		})
	}
	for _, sp := range rd.SpawnPoints {
		pr.SpawnPoints = append(pr.SpawnPoints, ParsedSpawnPoint{
			Name:      sp.Name,
			Character: sp.Character,
			Position:  Vec3{sp.Position.X, sp.Position.Y, sp.Position.Z},
		})
	}
	for _, h := range rd.Hotspots {
		pr.Hotspots = append(pr.Hotspots, ParsedHotspot{
			Name:     h.Name,
			Position: Vec3{h.Position.X, h.Position.Y, h.Position.Z},
			Size:     Vec3{h.Size.X, h.Size.Y, h.Size.Z},
		})
	}
	return pr
}

// GenerateRoomScene parses a .agroom and returns the generated .tscn text.
// scriptRelPath is the companion .agscript path (may be empty).
// glbRelPath is the optional .glb visual mesh path relative to the project
// root (pass "" when no .glb exists).
func GenerateRoomScene(filename, src, scriptRelPath, glbRelPath string) string {
	rd, err := room.ParseRoom(filename, src)
	if err != nil {
		return "-- parse error: " + err.Error()
	}
	return scene.GenerateRoomScene(rd, scriptRelPath, glbRelPath)
}

// ParsedChar is the result of parsing a .agchar file.
type ParsedChar struct {
	Name            string            `json:"name"`
	DisplayName     string            `json:"displayName,omitempty"`
	Type            string            `json:"type"`
	Mesh            string            `json:"mesh,omitempty"`
	Animations      map[string]string `json:"animations,omitempty"`
	SpriteSheet     string            `json:"spriteSheet,omitempty"`
	SpriteAngles    int               `json:"spriteAngles,omitempty"`
	FramesPerAngle  int               `json:"framesPerAngle,omitempty"`
	FrameSize       [2]int            `json:"frameSize,omitempty"`
	Error           string            `json:"error,omitempty"`
}

// ParseChar parses .agchar source text and returns the structured result.
func ParseChar(filename, src string) ParsedChar {
	cd, err := char.ParseChar(filename, src)
	if err != nil {
		return ParsedChar{Error: err.Error()}
	}
	return ParsedChar{
		Name:           cd.Name,
		DisplayName:    cd.DisplayName,
		Type:           cd.Type,
		Mesh:           cd.Mesh,
		Animations:     cd.Animations,
		SpriteSheet:    cd.SpriteSheet,
		SpriteAngles:   cd.SpriteAngles,
		FramesPerAngle: cd.FramesPerAngle,
		FrameSize:      cd.FrameSize,
	}
}

// GenerateCharScene parses a .agchar and returns the generated .tscn text.
func GenerateCharScene(filename, src string) string {
	cd, err := char.ParseChar(filename, src)
	if err != nil {
		return "-- parse error: " + err.Error()
	}
	return scene.GenerateCharScene(cd, nil)
}

// -------------------------------------------------------------------
// Validate
// -------------------------------------------------------------------

// ValidateIssue is one finding from ag validate.
type ValidateIssue struct {
	File     string `json:"file"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// ValidateResult is the full output of ValidateProject.
type ValidateResult struct {
	Issues []ValidateIssue `json:"issues"`
	Error  string          `json:"error,omitempty"`
}

// ValidateProjectDir runs all cross-reference checks on the project at root.
func ValidateProjectDir(root string) ValidateResult {
	m, err := project.Load(root)
	if err != nil {
		return ValidateResult{Error: err.Error()}
	}
	issues, err := validate.ValidateProject(root, m)
	if err != nil {
		return ValidateResult{Error: err.Error()}
	}
	out := ValidateResult{Issues: make([]ValidateIssue, len(issues))}
	for i, iss := range issues {
		out.Issues[i] = ValidateIssue{File: iss.File, Severity: iss.Severity, Message: iss.Message}
	}
	return out
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
