package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ags3d/ag/api"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application backend. Every exported method becomes
// callable from the Svelte frontend via window.go.<MethodName>().
type App struct {
	ctx         context.Context
	projectRoot string
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// -------------------------------------------------------------------
// Project management
// -------------------------------------------------------------------

// ProjectInfo is returned by OpenProject / LoadProject.
type ProjectInfo struct {
	Root           string `json:"root"`
	Name           string `json:"name"`
	StartRoom      string `json:"startRoom"`
	StartCharacter string `json:"startCharacter"`
	RenderingMode  string `json:"renderingMode"`
	Error          string `json:"error,omitempty"`
}

// SourceFile is a single adventure-game source file.
type SourceFile struct {
	Path string `json:"path"`
	Rel  string `json:"rel"`
	Ext  string `json:"ext"`
}

// OpenProject opens a native folder-picker dialog then loads game.agp.
func (a *App) OpenProject() ProjectInfo {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open AGS3D Project",
	})
	if err != nil || dir == "" {
		return ProjectInfo{Error: "cancelled"}
	}
	return a.LoadProject(dir)
}

// LoadProject loads a project from an explicit path (used by recent-projects list).
func (a *App) LoadProject(root string) ProjectInfo {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ProjectInfo{Error: err.Error()}
	}
	m, err := api.LoadProject(absRoot)
	if err != nil {
		return ProjectInfo{Root: absRoot, Error: err.Error()}
	}
	a.projectRoot = absRoot
	return ProjectInfo{
		Root:           m.Root,
		Name:           m.Name,
		StartRoom:      m.StartRoom,
		StartCharacter: m.StartCharacter,
		RenderingMode:  m.RenderingMode,
	}
}

// ListSourceFiles returns all adventure-game source files in the open project.
func (a *App) ListSourceFiles() []SourceFile {
	if a.projectRoot == "" {
		return nil
	}
	files, err := api.ScanProject(a.projectRoot)
	if err != nil {
		return nil
	}
	out := make([]SourceFile, len(files))
	for i, f := range files {
		out[i] = SourceFile{Path: f.Path, Rel: f.Rel, Ext: f.Ext}
	}
	return out
}

// -------------------------------------------------------------------
// Reference folders (arbitrary directories, no game.agp required)
// -------------------------------------------------------------------

// RefFolderInfo is returned by OpenRefFolder.
type RefFolderInfo struct {
	Root  string `json:"root"`
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

// OpenRefFolder opens a native folder-picker and returns the chosen folder.
func (a *App) OpenRefFolder() RefFolderInfo {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Reference Folder",
	})
	if err != nil || dir == "" {
		return RefFolderInfo{Error: "cancelled"}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return RefFolderInfo{Error: err.Error()}
	}
	return RefFolderInfo{Root: abs, Name: filepath.Base(abs)}
}

// ListRefFiles returns all AG source files under any arbitrary folder.
func (a *App) ListRefFiles(root string) []SourceFile {
	files, err := api.ScanFolder(root)
	if err != nil {
		return nil
	}
	out := make([]SourceFile, len(files))
	for i, f := range files {
		out[i] = SourceFile{Path: f.Path, Rel: f.Rel, Ext: f.Ext}
	}
	return out
}

// ReadFile returns the raw source of a file inside the project.
func (a *App) ReadFile(absPath string) string {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return ""
	}
	return string(data)
}

// -------------------------------------------------------------------
// Build
// -------------------------------------------------------------------

// Build runs the ag build pipeline and streams log lines as Wails events.
func (a *App) Build() string {
	if a.projectRoot == "" {
		return "no project open"
	}
	runtime.EventsEmit(a.ctx, "log:clear", "")
	err := api.Build(a.projectRoot, func(kind, msg string) {
		runtime.EventsEmit(a.ctx, "log:"+kind, msg)
	})
	if err != nil {
		return err.Error()
	}
	return "ok"
}

// -------------------------------------------------------------------
// Transpile — full per-file pipeline result
// -------------------------------------------------------------------

// TranspileResult holds every pipeline stage's output for a single file.
type TranspileResult struct {
	Source    string   `json:"source"`
	Tokens    string   `json:"tokens"`
	ASTText   string   `json:"astText"`
	ASTDot    string   `json:"astDot"`
	Symbols   string   `json:"symbols"`
	SymDot    string   `json:"symDot"`
	Blocking  string   `json:"blocking"`
	GDScript  string   `json:"gdscript"`
	EmitView  string   `json:"emitView"`
	SourceMap [][3]any `json:"sourceMap"`
	Errors    []string `json:"errors"`
}

// TranspileFile runs all pipeline stages on the file at absPath.
func (a *App) TranspileFile(absPath string) TranspileResult {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return TranspileResult{Errors: []string{err.Error()}}
	}
	src := string(data)
	rel := a.relPath(absPath)

	r := api.TranspileFile(rel, src)
	return TranspileResult{
		Source:    r.Source,
		Tokens:    r.Tokens,
		ASTText:   r.ASTText,
		ASTDot:    r.ASTDot,
		Symbols:   r.Symbols,
		SymDot:    r.SymDot,
		Blocking:  r.Blocking,
		GDScript:  r.GDScript,
		EmitView:  r.EmitView,
		SourceMap: r.SourceMap,
		Errors:    r.Errors,
	}
}

// -------------------------------------------------------------------
// Viz helpers (single-stage, used by the Viz panel)
// -------------------------------------------------------------------

func (a *App) VizTokens(absPath string) string     { return a.vizFile(absPath, api.VizTokens) }
func (a *App) VizAST(absPath string) string        { return a.vizFile(absPath, api.VizAST) }
func (a *App) VizASTDot(absPath string) string     { return a.vizFile(absPath, api.VizASTDot) }
func (a *App) VizSymbols(absPath string) string    { return a.vizFile(absPath, api.VizSymbols) }
func (a *App) VizSymbolsDot(absPath string) string { return a.vizFile(absPath, api.VizSymbolsDot) }
func (a *App) VizBlocking(absPath string) string   { return a.vizFile(absPath, api.VizBlocking) }
func (a *App) VizEmit(absPath string) string       { return a.vizFile(absPath, api.VizEmit) }

func (a *App) vizFile(absPath string, fn func(file, src string) string) string {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "error: " + err.Error()
	}
	return fn(a.relPath(absPath), string(data))
}

// -------------------------------------------------------------------
// Batch viz
// -------------------------------------------------------------------

// BatchVizResult is one file's output from a batch viz run.
type BatchVizResult struct {
	File   string            `json:"file"`
	Rel    string            `json:"rel"`
	Stages map[string]string `json:"stages"`
	Error  string            `json:"error,omitempty"`
}

// BatchViz runs stage(s) on every file in paths, emitting "batchviz:result"
// events as each file completes. Returns the total file count.
func (a *App) BatchViz(paths []string, stage string) int {
	stages := resolveBatchStages(stage)
	vizFns := map[string]func(string, string) string{
		"tokens":      api.VizTokens,
		"ast":         api.VizAST,
		"ast-dot":     api.VizASTDot,
		"symbols":     api.VizSymbols,
		"symbols-dot": api.VizSymbolsDot,
		"blocking":    api.VizBlocking,
		"emit":        api.VizEmit,
	}
	for _, absPath := range paths {
		data, err := os.ReadFile(absPath)
		r := BatchVizResult{
			File:   absPath,
			Rel:    a.relPath(absPath),
			Stages: make(map[string]string),
		}
		if err != nil {
			r.Error = err.Error()
			runtime.EventsEmit(a.ctx, "batchviz:result", r)
			continue
		}
		src := string(data)
		for _, s := range stages {
			if fn, ok := vizFns[s]; ok {
				r.Stages[s] = fn(r.Rel, src)
			}
		}
		runtime.EventsEmit(a.ctx, "batchviz:result", r)
	}
	runtime.EventsEmit(a.ctx, "batchviz:done", len(paths))
	return len(paths)
}

// -------------------------------------------------------------------
// Room / Char inspection
// -------------------------------------------------------------------

// ParsedRoom is the structured result of parsing a .agroom file.
type ParsedRoom = api.ParsedRoom

// ParsedChar is the structured result of parsing a .agchar file.
type ParsedChar = api.ParsedChar

// ValidateResult is the output of ValidateProject.
type ValidateResult = api.ValidateResult

// ParseRoom reads and parses the .agroom at absPath.
func (a *App) ParseRoom(absPath string) api.ParsedRoom {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return api.ParsedRoom{Error: err.Error()}
	}
	return api.ParseRoom(a.relPath(absPath), string(data))
}

// ParseChar reads and parses the .agchar at absPath.
func (a *App) ParseChar(absPath string) api.ParsedChar {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return api.ParsedChar{Error: err.Error()}
	}
	return api.ParseChar(a.relPath(absPath), string(data))
}

// GenerateRoomScene parses the .agroom at absPath and returns generated .tscn text.
func (a *App) GenerateRoomScene(absPath string) string {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "-- error: " + err.Error()
	}
	rel := a.relPath(absPath)
	scriptRelPath := strings.TrimSuffix(rel, ".agroom") + ".agscript"
	return api.GenerateRoomScene(rel, string(data), scriptRelPath)
}

// GenerateCharScene parses the .agchar at absPath and returns generated .tscn text.
func (a *App) GenerateCharScene(absPath string) string {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "-- error: " + err.Error()
	}
	return api.GenerateCharScene(a.relPath(absPath), string(data))
}

// ValidateProject runs ag validate on the open project.
func (a *App) ValidateProject() api.ValidateResult {
	if a.projectRoot == "" {
		return api.ValidateResult{Error: "no project open"}
	}
	return api.ValidateProjectDir(a.projectRoot)
}

// -------------------------------------------------------------------
// Utilities
// -------------------------------------------------------------------

func (a *App) relPath(absPath string) string {
	if a.projectRoot != "" {
		if r, err := filepath.Rel(a.projectRoot, absPath); err == nil {
			return r
		}
	}
	return filepath.Base(absPath)
}

func resolveBatchStages(stage string) []string {
	if stage == "all" || stage == "" {
		return []string{"tokens", "ast", "ast-dot", "symbols", "symbols-dot", "blocking", "emit"}
	}
	return []string{stage}
}
