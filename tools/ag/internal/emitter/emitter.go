// Package emitter generates GDScript from an AGS-spirit AST and writes
// source maps (.agmap sidecar files). Stub — full implementation in T13–T17.
package emitter

import (
	"github.com/ags3d/ag/internal/parser"
)

// Result holds the emitted GDScript and source map for one file.
type Result struct {
	// GDScript is the emitted source, ready to write to .engine/generated/.
	GDScript string
	// SourceMap maps each 1-based GDScript line to the originating
	// AGS-spirit file and line. Format mirrors the .agmap JSON schema:
	// [[gdscript_line, "rel/path.agscript", agscript_line], ...]
	SourceMap [][3]any
}

// Emitter walks an AGS-spirit AST and produces GDScript output.
// Full implementation in T13–T17.
type Emitter struct{}

// New creates a new Emitter.
func New() *Emitter { return &Emitter{} }

// Emit generates GDScript for the given parsed file.
// TODO(T13): implement visitor-based GDScript emission.
// TODO(T16): insert `await` for all IsBlocking call sites.
// TODO(T17): populate SourceMap with per-line origin data.
func (e *Emitter) Emit(f *parser.File) (*Result, error) {
	return &Result{
		GDScript:  "# stub — transpiler not yet implemented (T13+)\n",
		SourceMap: nil,
	}, nil
}
