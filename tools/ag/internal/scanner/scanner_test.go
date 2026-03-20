// Tests for the AGS-spirit lexer.
// Stubs only — full test cases added in T07.
package scanner_test

import (
	"testing"

	"github.com/ags3d/ag/internal/scanner"
)

// TODO(T07): expand with full token-type coverage once the lexer is implemented.

func TestScanner_EOFOnEmpty(t *testing.T) {
	s := scanner.New("test.agscript", "")
	tok := s.Next()
	if tok.Kind != scanner.TokenEOF {
		t.Errorf("expected EOF on empty input, got %v", tok.Kind)
	}
}

func TestScanner_FileNamePreserved(t *testing.T) {
	s := scanner.New("rooms/start.agscript", "")
	tok := s.Next()
	if tok.File != "rooms/start.agscript" {
		t.Errorf("File = %q, want rooms/start.agscript", tok.File)
	}
}
