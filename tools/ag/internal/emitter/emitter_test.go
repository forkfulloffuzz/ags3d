// Tests for the AGS-spirit → GDScript emitter.
// Uses golden files in testdata/ for output comparison.
// Stubs only — full test cases added in T13–T17.
package emitter_test

import (
	"testing"

	"github.com/ags3d/ag/internal/emitter"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// TODO(T13): add golden-file tests once the emitter is implemented.
// Pattern: parse a fixture .agscript, emit, compare to testdata/*.gd.

func TestEmitter_EmptyFileProducesOutput(t *testing.T) {
	s := scanner.New("empty.agscript", "")
	p := parser.New(s)
	f, _ := p.Parse("empty.agscript")
	e := emitter.New()
	result, err := e.Emit(f)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	if result == nil {
		t.Fatal("Emit returned nil result")
	}
	if result.GDScript == "" {
		t.Error("Emit returned empty GDScript string")
	}
}
