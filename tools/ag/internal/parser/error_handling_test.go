package parser_test

// T12 — parser error handling and recovery tests.
//
// Covers:
//   - UT-M2-14  Malformed input produces a non-empty error list with no crash
//   - UT-M2-15  Every error object has a line number > 0
//   - All 20 invalid fixtures are handled without panic
//   - Parser-level fixtures produce at least one error
//   - Semantic-only fixtures (T10) parse cleanly (zero parser errors)
//   - Error messages contain file, line, column, and human-readable text
//     (no raw TokenKind integers like "expected 35")

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/scanner"
)

// parseInvalidFixture is a helper that reads an invalid fixture and parses it.
func parseInvalidFixture(t *testing.T, name string) []*parser.ParseError {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "scripts", "invalid", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	s := scanner.New(name, string(data))
	p := parser.New(s)
	_, errs := p.Parse(name)
	return errs
}

// -------------------------------------------------------------------
// UT-M2-14 — malformed input produces errors, never panics
// -------------------------------------------------------------------

// TestErrorHandling_NoPanicOnAnyInput feeds every invalid fixture through the
// parser and asserts it returns without panicking.  This is the primary
// no-crash guarantee required by UT-M2-14.
func TestErrorHandling_NoPanicOnAnyInput(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "..", "testdata", "scripts", "invalid", "*.agscript"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no invalid fixture files found")
	}
	for _, path := range fixtures {
		name := filepath.Base(path)
		t.Run(strings.TrimSuffix(name, ".agscript"), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			s := scanner.New(name, string(data))
			p := parser.New(s)
			// Must not panic — partial AST + errors is acceptable.
			_, _ = p.Parse(name)
		})
	}
}

// TestErrorHandling_ParserErrors_NonEmpty verifies that each parser-level
// fixture (structural syntax errors) produces at least one error (UT-M2-14).
func TestErrorHandling_ParserErrors_NonEmpty(t *testing.T) {
	// These fixtures contain syntax errors the parser detects directly.
	// Each must produce ≥1 ParseError.
	parserErrorFiles := []string{
		"err_01_unterminated_string.agscript",
		"err_02_unclosed_brace.agscript",
		"err_03_unclosed_paren.agscript",
		"err_04_missing_semicolon.agscript",
		"err_05_export_outside_namespace.agscript",
		"err_06_invalid_token.agscript",
		"err_07_missing_function_body.agscript",
		"err_08_missing_if_condition.agscript",
		"err_09_bad_for_loop.agscript",
		"err_10_switch_no_brace.agscript",
		"err_11_enum_missing_brace.agscript",
		"err_12_namespace_missing_name.agscript",
		"err_13_unterminated_block_comment.agscript",
		"err_15_double_operator.agscript",
		"err_18_case_outside_switch.agscript",
		"err_19_function_inside_function.agscript",
		"err_20_global_assigned.agscript",
	}
	for _, name := range parserErrorFiles {
		name := name
		t.Run(strings.TrimSuffix(name, ".agscript"), func(t *testing.T) {
			errs := parseInvalidFixture(t, name)
			if len(errs) == 0 {
				t.Errorf("expected at least one parse error, got none")
			}
		})
	}
}

// TestErrorHandling_SemanticOnly_ParseClean verifies that fixtures whose errors
// are detected only by the symbol table (T10) parse with zero parser errors.
func TestErrorHandling_SemanticOnly_ParseClean(t *testing.T) {
	semanticFiles := []string{
		"err_14_break_outside_loop.agscript",
		"err_16_missing_return_value.agscript",
		"err_17_duplicate_export.agscript",
	}
	for _, name := range semanticFiles {
		name := name
		t.Run(strings.TrimSuffix(name, ".agscript"), func(t *testing.T) {
			errs := parseInvalidFixture(t, name)
			if len(errs) != 0 {
				t.Errorf("expected 0 parser errors (semantic-only fixture), got %d:", len(errs))
				for _, e := range errs {
					t.Logf("  %v", e)
				}
			}
		})
	}
}

// -------------------------------------------------------------------
// UT-M2-15 — every error has line > 0
// -------------------------------------------------------------------

// TestErrorHandling_LineNumbers_Positive asserts that every ParseError produced
// by any invalid fixture has a Line field greater than zero (UT-M2-15).
func TestErrorHandling_LineNumbers_Positive(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "..", "testdata", "scripts", "invalid", "*.agscript"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	for _, path := range fixtures {
		name := filepath.Base(path)
		t.Run(strings.TrimSuffix(name, ".agscript"), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			s := scanner.New(name, string(data))
			p := parser.New(s)
			_, errs := p.Parse(name)
			for _, e := range errs {
				if e.Line <= 0 {
					t.Errorf("error has line=%d (want >0): %v", e.Line, e)
				}
			}
		})
	}
}

// -------------------------------------------------------------------
// Error message quality — no raw token integers
// -------------------------------------------------------------------

// TestErrorHandling_NoRawTokenIntegers verifies that error messages do not
// contain raw TokenKind integers like "expected 35" or "expected 41".
// These would appear only if TokenKind lacks a String() method.
func TestErrorHandling_NoRawTokenIntegers(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("..", "..", "testdata", "scripts", "invalid", "*.agscript"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	for _, path := range fixtures {
		name := filepath.Base(path)
		t.Run(strings.TrimSuffix(name, ".agscript"), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			s := scanner.New(name, string(data))
			p := parser.New(s)
			_, errs := p.Parse(name)
			for _, e := range errs {
				msg := e.Message
				// "expected 35 …" pattern means TokenKind printed as integer.
				for i := 0; i <= 80; i++ {
					suspect := "expected " + itoa(i) + " "
					if strings.Contains(msg, suspect) || msg == "expected "+itoa(i) {
						t.Errorf("error message looks like raw token integer %d: %q", i, msg)
					}
				}
			}
		})
	}
}

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// -------------------------------------------------------------------
// ParseError struct — file and column are populated
// -------------------------------------------------------------------

func TestErrorHandling_ParseError_Fields(t *testing.T) {
	_, errs := parse("function f( { }")
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	e := errs[0]
	if e.File == "" {
		t.Error("ParseError.File should not be empty")
	}
	if e.Line <= 0 {
		t.Errorf("ParseError.Line = %d, want > 0", e.Line)
	}
	if e.Column <= 0 {
		t.Errorf("ParseError.Column = %d, want > 0", e.Column)
	}
	if e.Message == "" {
		t.Error("ParseError.Message should not be empty")
	}
	// Error() output should include file:line:col
	got := e.Error()
	if !strings.Contains(got, ":") {
		t.Errorf("Error() = %q, want file:line:col: msg format", got)
	}
}

// -------------------------------------------------------------------
// Recovery: parsing continues after an error
// -------------------------------------------------------------------

// TestErrorHandling_Recovery_ContinuesAfterError verifies that after a parse
// error the parser recovers and continues, so the partial AST still contains
// declarations that appear after the error site.
func TestErrorHandling_Recovery_ContinuesAfterError(t *testing.T) {
	// The first function has a bad body; the second is valid.
	// We expect errors AND at least one decl from recovery.
	src := `
function broken( {
}
function valid() {}
`
	f, errs := parse(src)
	if len(errs) == 0 {
		t.Fatal("expected at least one parse error")
	}
	// After recovery, 'valid' should appear in the parsed declarations.
	found := false
	for _, d := range f.Decls {
		if fd, ok := d.(*parser.FunctionDecl); ok && fd.Name == "valid" {
			found = true
		}
	}
	if !found {
		t.Error("parser should recover and parse 'valid' function after the error")
	}
}

// -------------------------------------------------------------------
// TokenKind.String() — human-readable names
// -------------------------------------------------------------------

func TestTokenKind_String_HumanReadable(t *testing.T) {
	cases := []struct {
		kind scanner.TokenKind
		want string
	}{
		{scanner.TokenRBrace, "'}'"},
		{scanner.TokenDot, "'.'"},
		{scanner.TokenSemicolon, "';'"},
		{scanner.TokenLParen, "'('"},
		{scanner.TokenRParen, "')'"},
		{scanner.TokenFunction, "'function'"},
		{scanner.TokenIf, "'if'"},
		{scanner.TokenEOF, "end of file"},
		{scanner.TokenIdent, "identifier"},
		{scanner.TokenIntLit, "integer literal"},
	}
	for _, tc := range cases {
		got := tc.kind.String()
		if got != tc.want {
			t.Errorf("TokenKind(%d).String() = %q, want %q", int(tc.kind), got, tc.want)
		}
	}
}
