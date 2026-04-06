// Package fixtures_test provides a data-driven test harness for all files in
// tools/ag/testdata/. It walks the directory tree and dispatches each file to
// the appropriate parser (and validator where applicable) based on its extension,
// then asserts:
//
//   - Files under "valid/"   must produce zero errors.
//   - Files under "invalid/" must produce at least one error.
//
// Adding or editing a fixture file is all that is ever needed to add a test —
// no Go code changes required.
//
// # Skip tags
//
// Invalid fixture files may include one of the following tags in their first
// comment line to tell the harness why it cannot test them directly:
//
//	[cross-system]    — requires a cross-system ProjectIndex (characters, named
//	                    points, audio/video files). Tested in unit tests that
//	                    construct the index explicitly.
//	[multi-file]      — error only fires when multiple files are validated
//	                    together. Tested in package-level unit tests.
//	[not-implemented] — the validator that catches this error has not been built
//	                    yet. Remove the tag when the validator is implemented.
//
// The harness skips the file with a descriptive message in any of these cases.
//
// # Validators run per extension
//
//	.agscript → scanner + parser
//	.agdlg    → dlg.Parse + dlg.Link + dlg.Validate
//	.agcut    → cut.Parse + cut.ValidateCutscene (single-file, nil cross-index)
//	.agroom   → room.ParseRoom
//	.agitem   → item.ParseItem
//
// Extensions with no registered parser are skipped with t.Skip.
package fixtures_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/cut"
	"github.com/ags3d/ag/internal/dlg"
	"github.com/ags3d/ag/internal/item"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/room"
	"github.com/ags3d/ag/internal/scanner"
)

const testdataDir = "../../testdata"

// fixtureTag represents a skip annotation embedded in the fixture's first comment.
type fixtureTag int

const (
	tagNone         fixtureTag = iota
	tagCrossSystem             // [cross-system]
	tagMultiFile               // [multi-file]
	tagNotImplemented          // [not-implemented]
)

// readFixtureMeta reads the first non-blank line of src. If it is a comment
// (starts with "//") it extracts:
//   - whether it is EXPECT_ERROR or EXPECT_WARNING
//   - any [tag] present in the line
func readFixtureMeta(src string) (expectWarning bool, tag fixtureTag) {
	sc := bufio.NewScanner(strings.NewReader(src))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "//") {
			break
		}
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "EXPECT_WARNING") {
			expectWarning = true
		}
		if strings.Contains(line, "[cross-system]") {
			tag = tagCrossSystem
		} else if strings.Contains(line, "[multi-file]") {
			tag = tagMultiFile
		} else if strings.Contains(line, "[not-implemented]") {
			tag = tagNotImplemented
		}
		break
	}
	return
}

// runParser dispatches the file to its parser+validator and returns all error
// messages. Returns (nil, nil) when no parser is registered for the extension.
func runParser(filename, src, ext string) ([]string, error) {
	switch ext {

	case ".agscript":
		s := scanner.New(filename, src)
		p := parser.New(s)
		_, errs := p.Parse(filename)
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		return msgs, nil

	case ".agdlg":
		df, parseErr := dlg.Parse(filename, src)
		if parseErr != nil {
			return []string{parseErr.Error()}, nil
		}
		// Link + validate (single-file project).
		lp, linkErr := dlg.Link([]*dlg.DialogueFile{df})
		if linkErr != nil {
			return []string{linkErr.Error()}, nil
		}
		var msgs []string
		for _, le := range lp.LinkErrors {
			msgs = append(msgs, le.Error())
		}
		for _, ve := range dlg.Validate(lp) {
			msgs = append(msgs, ve.Error())
		}
		return msgs, nil

	case ".agcut":
		cf, parseErr := cut.Parse(filename, src)
		if parseErr != nil {
			return []string{parseErr.Error()}, nil
		}
		// Run single-file validator. Pass nil allTitles so that cross-file
		// existence checks (CUT-E008) are skipped — those require the full
		// project title set and are tested via [multi-file] fixtures instead.
		// nil cross-index also skips E002–E005 (character/point/asset existence).
		var msgs []string
		for _, ve := range cut.ValidateCutscene(cf, nil, nil) {
			msgs = append(msgs, ve.Error())
		}
		return msgs, nil

	case ".agroom":
		_, err := room.ParseRoom(filename, src)
		if err != nil {
			return []string{err.Error()}, nil
		}
		return nil, nil

	case ".agitem":
		_, err := item.ParseItem(filename, src)
		if err != nil {
			return []string{err.Error()}, nil
		}
		return nil, nil

	default:
		return nil, nil // no parser registered
	}
}

// TestFixtures is the top-level harness that walks testdata/ and tests every
// fixture file.
func TestFixtures(t *testing.T) {
	root, err := filepath.Abs(testdataDir)
	if err != nil {
		t.Fatalf("cannot resolve testdata path: %v", err)
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Fatalf("testdata directory not found: %s", root)
	}

	categories, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", root, err)
	}

	for _, cat := range categories {
		if !cat.IsDir() {
			continue
		}
		catName := cat.Name()

		for _, validity := range []string{"valid", "invalid"} {
			dir := filepath.Join(root, catName, validity)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Errorf("ReadDir(%s): %v", dir, err)
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				ext := strings.ToLower(filepath.Ext(name))
				fullPath := filepath.Join(dir, name)
				testName := fmt.Sprintf("%s/%s/%s", catName, validity, name)

				t.Run(testName, func(t *testing.T) {
					raw, err := os.ReadFile(fullPath)
					if err != nil {
						t.Fatalf("ReadFile: %v", err)
					}
					src := string(raw)

					// For invalid fixtures, check for skip tags before running.
					if validity == "invalid" {
						expectWarn, tag := readFixtureMeta(src)
						switch tag {
						case tagCrossSystem:
							t.Skip("requires cross-system ProjectIndex — tested in package unit tests")
						case tagMultiFile:
							t.Skip("requires multi-file context — tested in package unit tests")
						case tagNotImplemented:
							t.Skip("validator not yet implemented — remove [not-implemented] tag when done")
						}
						if expectWarn {
							t.Skip("warning-only fixture — warning validators not yet implemented")
						}
					}

					errors, err := runParser(fullPath, src, ext)
					if err != nil {
						t.Fatalf("parser dispatch error: %v", err)
					}
					if errors == nil && ext != ".agscript" && ext != ".agdlg" && ext != ".agcut" && ext != ".agroom" && ext != ".agitem" {
						t.Skipf("no parser registered for extension %q", ext)
						return
					}

					switch validity {
					case "valid":
						if len(errors) > 0 {
							t.Errorf("expected zero errors, got %d:\n%s",
								len(errors), strings.Join(errors, "\n"))
						}
					case "invalid":
						if len(errors) == 0 {
							t.Errorf("expected at least one error, got none")
						}
					}
				})
			}
		}
	}
}
