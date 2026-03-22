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

	"github.com/ags3d/ag/internal/emitter"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/scanner"
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

// ScanProject returns all source files in root.
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
