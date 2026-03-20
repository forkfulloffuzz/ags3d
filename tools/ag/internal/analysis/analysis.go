// Package analysis provides static analysis over a parsed AGS-spirit project.
// Shared by the ag validate command and the agls language server.
// Stub — full implementation in T12 and beyond.
package analysis

import "github.com/ags3d/ag/internal/parser"

// Diagnostic is a single analysis finding with source location.
type Diagnostic struct {
	File     string
	Line     int
	Column   int
	Severity string // "error" | "warning"
	Message  string
}

// Analyze runs all available checks on a parsed file and returns diagnostics.
// TODO(T12): implement broken-reference checks, unreachable option detection, etc.
func Analyze(f *parser.File) []Diagnostic {
	return nil
}
