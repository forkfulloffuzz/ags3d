// Tests for static analysis.
// Stubs only — full test cases added in T12.
package analysis_test

import (
	"testing"

	"github.com/ags3d/ag/internal/analysis"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// TODO(T12): add tests for broken reference detection, unreachable options, etc.

func TestAnalyze_EmptyFileNoDiagnostics(t *testing.T) {
	s := scanner.New("empty.agscript", "")
	p := parser.New(s)
	f, _ := p.Parse("empty.agscript")
	diags := analysis.Analyze(f)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics on empty file, got %v", diags)
	}
}
