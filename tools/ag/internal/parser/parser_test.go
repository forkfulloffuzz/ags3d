// Tests for the AGS-spirit parser and AST.
// Stubs only — full test cases added in T08–T11.
package parser_test

import (
	"testing"

	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// TODO(T09): expand with fixture-driven tests once the parser is implemented.

func TestParser_EmptyFileProducesNoDecls(t *testing.T) {
	s := scanner.New("empty.agscript", "")
	p := parser.New(s)
	f, errs := p.Parse("empty.agscript")
	if len(errs) != 0 {
		t.Errorf("unexpected errors on empty input: %v", errs)
	}
	if len(f.Decls) != 0 {
		t.Errorf("expected 0 decls on empty input, got %d", len(f.Decls))
	}
}

func TestParser_FilePathPreserved(t *testing.T) {
	s := scanner.New("rooms/market.agscript", "")
	p := parser.New(s)
	f, _ := p.Parse("rooms/market.agscript")
	if f.Path != "rooms/market.agscript" {
		t.Errorf("File.Path = %q, want rooms/market.agscript", f.Path)
	}
}
