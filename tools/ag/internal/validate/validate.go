// Package validate provides project-wide cross-reference checks for ag validate.
//
// Checks performed:
//
//  1. game.agp: start_room file exists
//  2. game.agp: start_character file exists
//  3. .agroom: initial_camera names a Camera block defined in the same room
//  4. .agroom: each SpawnPoint.character matches a known .agchar name
//  5. .agscript: WalkTo/FaceTo point-name args exist in the paired .agroom
//  6. .agscript: AddInventory/LoseInventory/HasInventory item-name args resolve to a known .agitem
//  7. .agscript: character receiver names in method calls resolve to a known .agchar
//  8. .agscript: HideRoomItem/ShowRoomItem item-name args exist as Hotspot blocks in the paired .agroom
//  9. .agscript: GoToRoom room-name args resolve to an existing rooms/<name>/<name.agroom>
//
// 10. .agroom + .agchar: billboard camera angle warnings (W1 elevation >30°, W3 arc >45° for 4-angle sprites)
package validate

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/ags3d/ag/internal/char"
	"github.com/ags3d/ag/internal/cut"
	"github.com/ags3d/ag/internal/dlg"
	"github.com/ags3d/ag/internal/item"
	"github.com/ags3d/ag/internal/loc"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/room"
	"github.com/ags3d/ag/internal/scanner"
)

// Issue is a single validation finding with source location.
type Issue struct {
	File     string
	Line     int    // 0 = no line info
	Severity string // "error" | "warning"
	Message  string
}

func (i Issue) String() string {
	if i.Line > 0 {
		return fmt.Sprintf("%s:%d: %s: %s", i.File, i.Line, i.Severity, i.Message)
	}
	return fmt.Sprintf("%s: %s: %s", i.File, i.Severity, i.Message)
}

// ValidateProject runs all cross-reference checks against the project at root.
// It returns a slice of Issues (may be empty if the project is clean) and any
// fatal infrastructure error that prevented validation from completing.
func ValidateProject(root string, manifest *project.Manifest) ([]Issue, error) {
	var issues []Issue

	// --- Check 1 & 2: game.agp references ---
	if manifest.Project.StartRoom != "" {
		if !fileExists(root, manifest.Project.StartRoom) {
			issues = append(issues, Issue{
				File:     "game.agp",
				Severity: "error",
				Message:  fmt.Sprintf("start_room %q not found", manifest.Project.StartRoom),
			})
		}
	}
	if manifest.Project.StartCharacter != "" {
		if !fileExists(root, manifest.Project.StartCharacter) {
			issues = append(issues, Issue{
				File:     "game.agp",
				Severity: "error",
				Message:  fmt.Sprintf("start_character %q not found", manifest.Project.StartCharacter),
			})
		}
	}

	// Scan all source files.
	files, err := project.Scan(root)
	if err != nil {
		return issues, err
	}

	fileIssues, err := ValidateFiles(files)
	issues = append(issues, fileIssues...)
	return issues, err
}

// ValidateFiles runs cross-reference checks on an explicit list of source files.
// Unlike ValidateProject it does not require a game.agp — use this when files
// are supplied via stdin or a file-list flag instead of a project directory.
func ValidateFiles(files []project.SourceFile) ([]Issue, error) {
	var issues []Issue

	// Build a set of known item names from all .agitem files.
	itemNames := make(map[string]bool)
	for _, f := range files {
		if f.Ext != ".agitem" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		it, parseErr := item.ParseItem(f.Rel, string(data))
		if parseErr != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  parseErr.Error(),
			})
			continue
		}
		itemNames[it.Name] = true
	}

	// Build a set of known character names from all .agchar files.
	// Map: character name → relative file path.
	charNames := make(map[string]string)
	charData := make(map[string]*char.CharData) // name → CharData for billboard checks
	for _, f := range files {
		if f.Ext != ".agchar" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		cd, parseErr := char.ParseChar(f.Rel, string(data))
		if parseErr != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  parseErr.Error(),
			})
			continue
		}
		charNames[cd.Name] = f.Rel
		charData[cd.Name] = cd
	}

	// Build a set of known room directory stems from all .agroom files.
	// "rooms/start/start.agroom" → stem "start".
	roomNames := make(map[string]bool)
	for _, f := range files {
		if f.Ext != ".agroom" {
			continue
		}
		// Stem is the directory name: rooms/start/ → "start".
		stem := filepath.Base(filepath.Dir(f.Rel))
		roomNames[stem] = true
	}

	// Build a map from .agroom relative path → parsed RoomData so scripts
	// can look up point names without re-parsing.
	// Key: relative path to .agroom (e.g. "rooms/start/start.agroom").
	roomData := make(map[string]*room.RoomData)

	// --- Check 3 & 4: .agroom internal consistency and cross-references ---
	for _, f := range files {
		if f.Ext != ".agroom" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		rd, parseErr := room.ParseRoom(f.Rel, string(data))
		if parseErr != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  parseErr.Error(),
			})
			continue
		}
		roomData[f.Rel] = rd

		// Check 3: initial_camera must name a Camera block in this room.
		if rd.InitialCamera != "" {
			if !hasCameraName(rd, rd.InitialCamera) {
				issues = append(issues, Issue{
					File:     f.Rel,
					Severity: "error",
					Message:  fmt.Sprintf("initial_camera %q is not defined in this room", rd.InitialCamera),
				})
			}
		}

		// Check 4: each SpawnPoint.character must name a known .agchar.
		for _, sp := range rd.SpawnPoints {
			if sp.Character == "" {
				continue
			}
			if _, ok := charNames[sp.Character]; !ok {
				issues = append(issues, Issue{
					File:     f.Rel,
					Severity: "error",
					Message:  fmt.Sprintf("SpawnPoint %q: unknown character %q (no matching .agchar found)", sp.Name, sp.Character),
				})
			}
		}
	}

	// --- Check 5 & 6: .agscript cross-references ---
	for _, f := range files {
		if f.Ext != ".agscript" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		src := string(data)

		// Check 5: WalkTo/FaceTo point names — only for scripts with a paired room.
		pairedRoom := pairedAgroom(f.Rel)
		if rd, ok := roomData[pairedRoom]; ok {
			issues = append(issues, checkScriptPointRefs(f.Rel, src, rd)...)
		}

		// Check 6: AddInventory/LoseInventory/HasInventory item names.
		issues = append(issues, checkScriptItemRefs(f.Rel, src, itemNames)...)

		// Check 7: character receiver names in method calls.
		issues = append(issues, checkScriptCharacterRefs(f.Rel, src, charNames)...)

		// Check 8: HideRoomItem/ShowRoomItem item names — only for scripts with a paired room.
		if rd, ok := roomData[pairedRoom]; ok {
			issues = append(issues, checkScriptRoomItemRefs(f.Rel, src, rd)...)
		}

		// Check 9: GoToRoom room names — validated against the known room set.
		issues = append(issues, checkScriptGoToRoomRefs(f.Rel, src, roomNames)...)
	}

	// --- Billboard camera warnings (W1, W3) ---
	issues = append(issues, validateBillboardCameraWarnings(roomData, charData)...)

	// --- Dialogue validation (DLG-E001..E025, DLG-W001..W012) ---
	issues = append(issues, validateDialogue(files, charNames, itemNames, roomData)...)

	// --- Localisation validation (DLG-LOC-E001, DLG-LOC-W001, DLG-LOC-W002) ---
	issues = append(issues, validateLocale(files)...)

	// --- Cutscene validation (CUT-E/W, SEQ-E/W) ---
	issues = append(issues, validateCutscenes(files, charNames, itemNames, roomData)...)

	// --- Cutscene localisation validation (T-CUT30) ---
	issues = append(issues, validateCutsceneLocKeys(files)...)

	// --- Dialogue localisation validation (T-LOC04) ---
	issues = append(issues, validateDialogueLocKeys(files)...)

	return issues, nil
}

// validateCutsceneLocKeys parses all .agcut files and validates that every
// explicit #loc: argument is present in at least one .agstrings locale file
// found in the same file set. Missing keys are reported as warnings (not errors)
// so the build is not blocked in dev; they become errors in release mode.
func validateCutsceneLocKeys(files []project.SourceFile) []Issue {
	var issues []Issue

	// Build a merged locale map from all .agstrings files in the project.
	localeMap := make(map[string]string)
	for _, f := range files {
		if f.Ext != ".agstrings" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		sf, parseErr := loc.Parse(f.Rel, string(data))
		if parseErr != nil {
			continue
		}
		for _, e := range sf.Entries {
			if !e.Orphan {
				localeMap[e.Key] = e.Value
			}
		}
	}
	if len(localeMap) == 0 {
		// No locale files in this project — nothing to validate against.
		return nil
	}

	releaseMode := os.Getenv("AGSBUILD") == "release"

	for _, f := range files {
		if f.Ext != ".agcut" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		cf, parseErr := cut.Parse(f.Rel, string(data))
		if parseErr != nil {
			continue // parse errors are reported by validateCutscenes
		}
		for _, msg := range cut.ValidateLocKeys(cf, localeMap) {
			sev := "warning"
			if releaseMode {
				sev = "error"
			}
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: sev,
				Message:  fmt.Sprintf("CUT-LOC-W001: %s", msg),
			})
		}
	}
	return issues
}

// validateDialogueLocKeys parses all .agdlg files and validates that every
// explicit #loc: annotation is present in at least one .agstrings locale file.
// Missing keys are reported as warnings (not errors) in dev mode and errors in
// release mode.
func validateDialogueLocKeys(files []project.SourceFile) []Issue {
	var issues []Issue

	localeMap := buildLocaleMap(files)
	if len(localeMap) == 0 {
		return nil
	}

	releaseMode := os.Getenv("AGSBUILD") == "release"

	var dlgFiles []*dlg.DialogueFile
	for _, f := range files {
		if f.Ext != ".agdlg" {
			continue
		}
		if _, err := os.Stat(f.Path); err != nil {
			continue
		}
		df, err := dlg.ParseFile(f.Path)
		if err != nil {
			continue
		}
		dlgFiles = append(dlgFiles, df)
	}
	if len(dlgFiles) == 0 {
		return nil
	}

	lp, err := dlg.Link(dlgFiles)
	if err != nil {
		return nil
	}

	for _, vi := range dlg.ValidateLocKeys(lp, localeMap) {
		sev := "warning"
		if releaseMode {
			sev = "error"
		}
		issues = append(issues, Issue{
			File:     vi.Pos.File,
			Line:     vi.Pos.Line,
			Severity: sev,
			Message:  fmt.Sprintf("DLG-LOC-W001: %s", vi.Msg),
		})
	}

	return issues
}

// buildLocaleMap merges all loc_key → translation entries from every .agstrings
// file in the project into a single map.
func buildLocaleMap(files []project.SourceFile) map[string]string {
	localeMap := make(map[string]string)
	for _, f := range files {
		if f.Ext != ".agstrings" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		sf, parseErr := loc.Parse(f.Rel, string(data))
		if parseErr != nil {
			continue
		}
		for _, e := range sf.Entries {
			if !e.Orphan {
				localeMap[e.Key] = e.Value
			}
		}
	}
	return localeMap
}

// validateCutscenes parses all .agcut files and runs CUT-E/W and SEQ-E/W checks.
func validateCutscenes(files []project.SourceFile, charNames map[string]string, itemNames map[string]bool, roomData map[string]*room.RoomData) []Issue {
	var issues []Issue

	// Parse all .agcut files.
	var cutFiles []*cut.CutsceneFile
	for _, f := range files {
		if f.Ext != ".agcut" {
			continue
		}
		cf, err := cut.ParseFile(f.Path)
		if err != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  err.Error(),
			})
			continue
		}
		cutFiles = append(cutFiles, cf)
	}
	if len(cutFiles) == 0 {
		return issues
	}

	// Build ProjectIndex from already-resolved project symbols.
	cross := &cut.ProjectIndex{
		Characters:     make(map[string]bool, len(charNames)),
		NamedPoints:    make(map[string]map[string]bool, len(roomData)),
		CutsceneTitles: make(map[string]bool, len(cutFiles)),
	}
	for name := range charNames {
		cross.Characters[name] = true
	}
	for _, rd := range roomData {
		pts := make(map[string]bool, len(rd.Points))
		for _, p := range rd.Points {
			pts[p.Name] = true
		}
		cross.NamedPoints[rd.Name] = pts
	}
	for _, cf := range cutFiles {
		if cf.Title != "" {
			cross.CutsceneTitles[cf.Title] = true
		}
	}

	// CUT-E/W: project-wide and per-file structural validation.
	for _, e := range cut.ValidateProjectCutscenes(cutFiles, cross) {
		issues = append(issues, Issue{
			File:     e.Pos.File,
			Line:     e.Pos.Line,
			Severity: "error",
			Message:  fmt.Sprintf("%s: %s", e.Code, e.Msg),
		})
	}
	for _, w := range cut.WarnProjectCutscenes(cutFiles, cross) {
		issues = append(issues, Issue{
			File:     w.Pos.File,
			Line:     w.Pos.Line,
			Severity: "warning",
			Message:  fmt.Sprintf("%s: %s", w.Code, w.Msg),
		})
	}

	// SEQ-E/W: per-file sequencing validation.
	for _, cf := range cutFiles {
		for _, e := range cut.ValidateSequence(cf) {
			issues = append(issues, Issue{
				File:     e.Pos.File,
				Line:     e.Pos.Line,
				Severity: "error",
				Message:  fmt.Sprintf("%s: %s", e.Code, e.Msg),
			})
		}
		for _, w := range cut.WarnSequence(cf) {
			issues = append(issues, Issue{
				File:     w.Pos.File,
				Line:     w.Pos.Line,
				Severity: "warning",
				Message:  fmt.Sprintf("%s: %s", w.Code, w.Msg),
			})
		}
	}

	return issues
}

// validateLocale scans all .agstrings files in the provided file list and
// reports missing translations (DLG-LOC-E001), stale keys (DLG-LOC-W001),
// and orphaned keys (DLG-LOC-W002).
//
// DLG-LOC-E001 (error) is only reported when the --release flag is set via
// the AGSBUILD environment variable; in dev builds it is downgraded to a
// warning so untranslated strings don't block iteration.
func validateLocale(files []project.SourceFile) []Issue {
	var issues []Issue
	releaseMode := os.Getenv("AGSBUILD") == "release"

	for _, f := range files {
		if f.Ext != ".agstrings" {
			continue
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		sf, parseErr := loc.Parse(f.Rel, string(data))
		if parseErr != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  parseErr.Error(),
			})
			continue
		}
		for _, e := range sf.Entries {
			if e.Orphan {
				issues = append(issues, Issue{
					File:     f.Rel,
					Severity: "warning",
					Message:  fmt.Sprintf("DLG-LOC-W002: orphaned key %q (no longer in source)", e.Key),
				})
				continue
			}
			if e.Stale {
				issues = append(issues, Issue{
					File:     f.Rel,
					Severity: "warning",
					Message:  fmt.Sprintf("DLG-LOC-W001: stale key %q (source text changed since last export)", e.Key),
				})
			}
			if e.Value == "" {
				sev := "warning"
				code := "DLG-LOC-W003"
				if releaseMode {
					sev = "error"
					code = "DLG-LOC-E001"
				}
				issues = append(issues, Issue{
					File:     f.Rel,
					Severity: sev,
					Message:  fmt.Sprintf("%s: untranslated key %q", code, e.Key),
				})
			}
		}
	}
	return issues
}

// validateDialogue parses all .agdlg files, links them, and runs the structural
// validator (DLG-E001..E011), cross-system validator (DLG-E020..E025), and
// static analysis warnings (DLG-W001..W012).
func validateDialogue(files []project.SourceFile, charNames map[string]string, itemNames map[string]bool, roomData map[string]*room.RoomData) []Issue {
	var issues []Issue

	// Parse all .agdlg files.
	var dlgFiles []*dlg.DialogueFile
	for _, f := range files {
		if f.Ext != ".agdlg" {
			continue
		}
		if _, err := os.Stat(f.Path); err != nil {
			continue
		}
		df, parseErr := dlg.ParseFile(f.Path)
		if parseErr != nil {
			issues = append(issues, Issue{
				File:     f.Rel,
				Severity: "error",
				Message:  parseErr.Error(),
			})
			continue
		}
		dlgFiles = append(dlgFiles, df)
	}
	if len(dlgFiles) == 0 {
		return issues
	}

	// Link.
	lp, linkErr := dlg.Link(dlgFiles)
	if linkErr != nil {
		issues = append(issues, Issue{
			File:     "dialogue",
			Severity: "error",
			Message:  linkErr.Error(),
		})
		return issues
	}

	// Structural validation (DLG-E001..E011).
	for _, e := range dlg.Validate(lp) {
		issues = append(issues, Issue{
			File:     e.Pos.File,
			Line:     e.Pos.Line,
			Severity: "error",
			Message:  fmt.Sprintf("%s: %s", e.Code, e.Msg),
		})
	}

	// Cross-system validation (DLG-E020..E025).
	sym := buildDlgSymbolTable(charNames, itemNames, roomData)
	for _, e := range dlg.ValidateCrossSystem(lp, sym) {
		issues = append(issues, Issue{
			File:     e.Pos.File,
			Line:     e.Pos.Line,
			Severity: "error",
			Message:  fmt.Sprintf("%s: %s", e.Code, e.Msg),
		})
	}

	// Static analysis warnings (DLG-W001..W012).
	for _, w := range dlg.WarnProject(lp) {
		issues = append(issues, Issue{
			File:     w.Pos.File,
			Line:     w.Pos.Line,
			Severity: "warning",
			Message:  fmt.Sprintf("%s: %s", w.Code, w.Msg),
		})
	}

	return issues
}

// buildDlgSymbolTable constructs the ProjectSymbolTable for cross-system
// dialogue validation from the already-parsed project symbols.
func buildDlgSymbolTable(charNames map[string]string, itemNames map[string]bool, roomData map[string]*room.RoomData) dlg.ProjectSymbolTable {
	sym := dlg.ProjectSymbolTable{
		CharacterNames: make(map[string]bool, len(charNames)),
		ItemNames:      itemNames,
		RoomNames:      make(map[string]bool, len(roomData)),
		RoomPoints:     make(map[string]map[string]bool, len(roomData)),
		FlagsEverSet:   make(map[string]bool),
		KnowledgeFlags: make(map[string]bool),
	}
	for name := range charNames {
		sym.CharacterNames[name] = true
	}
	for relPath, rd := range roomData {
		sym.RoomNames[rd.Name] = true
		pts := make(map[string]bool, len(rd.Points))
		for _, p := range rd.Points {
			pts[p.Name] = true
		}
		sym.RoomPoints[rd.Name] = pts
		_ = relPath
	}
	return sym
}

// checkScriptPointRefs parses src as an .agscript and returns Issues for any
// WalkTo/FaceTo call whose first string argument does not name a Point in rd.
func checkScriptPointRefs(rel, src string, rd *room.RoomData) []Issue {
	s := scanner.New(rel, src)
	p := parser.New(s)
	f, _ := p.Parse(rel) // parse errors are emitter's job; we tolerate them here

	pointNames := make(map[string]bool, len(rd.Points))
	for _, pt := range rd.Points {
		pointNames[strings.ToLower(pt.Name)] = true
	}

	var issues []Issue
	for _, decl := range f.Decls {
		walkDecl(decl, func(call *parser.CallExpr) {
			method, ok := methodName(call)
			if !ok {
				return
			}
			if method != "WalkTo" && method != "FaceTo" {
				return
			}
			if len(call.Args) == 0 {
				return
			}
			lit, ok := call.Args[0].(*parser.Literal)
			if !ok || lit.Kind != "string" {
				return
			}
			name := lit.Value
			if !pointNames[strings.ToLower(name)] {
				tok := call.ExprPos()
				issues = append(issues, Issue{
					File:     rel,
					Line:     tok.Line,
					Severity: "error",
					Message:  fmt.Sprintf("%s(%q): point %q is not defined in the room", method, name, name),
				})
			}
		})
	}
	return issues
}

// inventoryBuiltins are the global inventory calls whose first string arg names an item.
var inventoryBuiltins = map[string]bool{
	"AddInventory":  true,
	"LoseInventory": true,
	"HasInventory":  true,
}

// checkScriptCharacterRefs parses src as an .agscript and returns Issues for any
// method call whose receiver (object) does not name a known character.
// For example: player.Say(...) — "player" must be a known .agchar.
func checkScriptCharacterRefs(rel, src string, charNames map[string]string) []Issue {
	s := scanner.New(rel, src)
	p := parser.New(s)
	f, _ := p.Parse(rel)

	var issues []Issue
	for _, decl := range f.Decls {
		walkDecl(decl, func(call *parser.CallExpr) {
			mem, ok := call.Callee.(*parser.MemberExpr)
			if !ok {
				return
			}
			ident, ok := mem.Object.(*parser.Identifier)
			if !ok {
				return
			}
			if _, ok := charNames[ident.Name]; !ok {
				tok := call.ExprPos()
				issues = append(issues, Issue{
					File:     rel,
					Line:     tok.Line,
					Severity: "error",
					Message:  fmt.Sprintf("%s.%s(...): unknown character %q (no matching .agchar found)", ident.Name, mem.Field, ident.Name),
				})
			}
		})
	}
	return issues
}

// checkScriptItemRefs parses src as an .agscript and returns Issues for any
// AddInventory/LoseInventory/HasInventory call whose first string argument
// does not name a known item.
func checkScriptItemRefs(rel, src string, itemNames map[string]bool) []Issue {
	s := scanner.New(rel, src)
	p := parser.New(s)
	f, _ := p.Parse(rel)

	var issues []Issue
	for _, decl := range f.Decls {
		walkDecl(decl, func(call *parser.CallExpr) {
			// These are bare function calls, not method calls.
			ident, ok := call.Callee.(*parser.Identifier)
			if !ok {
				return
			}
			if !inventoryBuiltins[ident.Name] {
				return
			}
			if len(call.Args) == 0 {
				return
			}
			lit, ok := call.Args[0].(*parser.Literal)
			if !ok || lit.Kind != "string" {
				return
			}
			name := lit.Value
			if !itemNames[name] {
				tok := call.ExprPos()
				issues = append(issues, Issue{
					File:     rel,
					Line:     tok.Line,
					Severity: "error",
					Message:  fmt.Sprintf("%s(%q): item %q is not defined (no matching .agitem found)", ident.Name, name, name),
				})
			}
		})
	}
	return issues
}

// roomItemBuiltins are the global calls whose first string arg names a room Hotspot.
var roomItemBuiltins = map[string]bool{
	"HideRoomItem": true,
	"ShowRoomItem": true,
}

// checkScriptRoomItemRefs parses src as an .agscript and returns Issues for any
// HideRoomItem/ShowRoomItem call whose first string argument does not name a
// Hotspot block in rd.
func checkScriptRoomItemRefs(rel, src string, rd *room.RoomData) []Issue {
	s := scanner.New(rel, src)
	p := parser.New(s)
	f, _ := p.Parse(rel)

	hotspotNames := make(map[string]bool, len(rd.Hotspots))
	for _, hs := range rd.Hotspots {
		hotspotNames[strings.ToLower(hs.Name)] = true
	}

	var issues []Issue
	for _, decl := range f.Decls {
		walkDecl(decl, func(call *parser.CallExpr) {
			ident, ok := call.Callee.(*parser.Identifier)
			if !ok {
				return
			}
			if !roomItemBuiltins[ident.Name] {
				return
			}
			if len(call.Args) == 0 {
				return
			}
			lit, ok := call.Args[0].(*parser.Literal)
			if !ok || lit.Kind != "string" {
				return
			}
			name := lit.Value
			if !hotspotNames[strings.ToLower(name)] {
				tok := call.ExprPos()
				issues = append(issues, Issue{
					File:     rel,
					Line:     tok.Line,
					Severity: "error",
					Message:  fmt.Sprintf("%s(%q): hotspot %q is not defined in this room", ident.Name, name, name),
				})
			}
		})
	}
	return issues
}

// checkScriptGoToRoomRefs parses src as an .agscript and returns Issues for any
// GoToRoom call whose first string argument does not name a known room directory
// (rooms/<name>/<name.agroom>).
func checkScriptGoToRoomRefs(rel, src string, roomNames map[string]bool) []Issue {
	s := scanner.New(rel, src)
	p := parser.New(s)
	f, _ := p.Parse(rel)

	var issues []Issue
	for _, decl := range f.Decls {
		walkDecl(decl, func(call *parser.CallExpr) {
			ident, ok := call.Callee.(*parser.Identifier)
			if !ok {
				return
			}
			if ident.Name != "GoToRoom" {
				return
			}
			if len(call.Args) == 0 {
				return
			}
			lit, ok := call.Args[0].(*parser.Literal)
			if !ok || lit.Kind != "string" {
				return
			}
			roomName := lit.Value
			if !roomNames[roomName] {
				tok := call.ExprPos()
				issues = append(issues, Issue{
					File:     rel,
					Line:     tok.Line,
					Severity: "error",
					Message:  fmt.Sprintf("GoToRoom(%q): room %q has no .agroom file (rooms/%s/%s.agroom)", roomName, roomName, roomName, roomName),
				})
			}
		})
	}
	return issues
}

// validateBillboardCameraWarnings checks each room for billboard camera configuration
// issues and returns warnings for W1 (elevation > 30°) and W3 (arc > 45° for 4-angle sprites).
func validateBillboardCameraWarnings(roomData map[string]*room.RoomData, charData map[string]*char.CharData) []Issue {
	var issues []Issue

	// Build a set of billboard character names per room by scanning SpawnPoints.
	// A SpawnPoint names a character; if that character is billboard (Type=="2d"),
	// the room is considered a billboard room.
	type roomBillboards struct {
		hasBillboard bool
		maxAngles    int // max SpriteAngles across all billboard chars in room
	}
	roomBB := make(map[string]*roomBillboards)

	for roomRel, rd := range roomData {
		bb := &roomBillboards{}
		for _, sp := range rd.SpawnPoints {
			if sp.Character == "" {
				continue
			}
			cd, ok := charData[sp.Character]
			if !ok {
				continue
			}
			if cd.Type == "2d" {
				bb.hasBillboard = true
				if cd.SpriteAngles > bb.maxAngles {
					bb.maxAngles = cd.SpriteAngles
				}
			}
		}
		if bb.hasBillboard || bb.maxAngles > 0 {
			roomBB[roomRel] = bb
		}
	}

	for roomRel, rd := range roomData {
		bb, ok := roomBB[roomRel]
		if !ok {
			continue // no billboard characters in this room
		}
		for _, cam := range rd.Cameras {
			if !cam.HasPosition || !cam.HasLookAt {
				continue
			}
			// W1: Camera elevation angle > 30°.
			// Elevation = angle between XZ plane and the vector to look_at.
			dx := cam.Position.X - cam.LookAt.X
			dy := cam.Position.Y - cam.LookAt.Y
			dz := cam.Position.Z - cam.LookAt.Z
			horizontal := math.Sqrt(dx*dx + dz*dz)
			if horizontal > 0 {
				elevDeg := math.Atan2(math.Abs(dy), horizontal) * 180 / math.Pi
				if elevDeg > 30 {
					issues = append(issues, Issue{
						File:     roomRel,
						Severity: "warning",
						Message:  fmt.Sprintf("W1: Camera %q elevation (%.0f°) may clip billboard character sprites. Recommended: keep below 30°.", cam.Name, elevDeg),
					})
				}
			}
			// W3: Camera horizontal arc > 45° relative to room origin AND room has 4-angle sprites.
			if bb.maxAngles == 4 {
				// Camera arc: horizontal displacement from origin (look_at is treated as floor centre ~0,0,0).
				// Use the XZ displacement as the arc metric.
				arcDeg := math.Atan2(math.Abs(dx), math.Abs(dz)) * 180 / math.Pi
				// Total arc span is roughly 2 * arcDeg for a camera at (x, 0, z) looking at origin.
				// We warn when the camera is positioned more than 45° from a cardinal axis.
				if arcDeg > 45 {
					issues = append(issues, Issue{
						File:     roomRel,
						Severity: "warning",
						Message:  fmt.Sprintf("W3: Camera %q horizontal arc (%.0f°) may cause visible direction snapping for 4-angle billboard sprites. Recommended: keep arc below 45° from cardinal axes.", cam.Name, arcDeg),
					})
				}
			}
		}
	}
	return issues
}

// methodName returns the method name if call is a method call (obj.Method(...)).
func methodName(call *parser.CallExpr) (string, bool) {
	mem, ok := call.Callee.(*parser.MemberExpr)
	if !ok {
		return "", false
	}
	return mem.Field, true
}

// walkDecl visits every CallExpr reachable from a top-level declaration node.
func walkDecl(node parser.Node, fn func(*parser.CallExpr)) {
	switch n := node.(type) {
	case *parser.FunctionDecl:
		walkBlock(n.Body, fn)
	case *parser.NamespaceDecl:
		for _, m := range n.Members {
			walkDecl(m, fn)
		}
	}
}

func walkBlock(b *parser.Block, fn func(*parser.CallExpr)) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		walkStmt(s, fn)
	}
}

func walkStmt(s parser.Stmt, fn func(*parser.CallExpr)) {
	switch n := s.(type) {
	case *parser.ExprStmt:
		walkExpr(n.X, fn)
	case *parser.VarDecl:
		walkExpr(n.Init, fn)
	case *parser.IfStmt:
		walkExpr(n.Cond, fn)
		walkBlock(n.Then, fn)
		if n.Else != nil {
			walkStmt(n.Else, fn)
		}
	case *parser.Block:
		walkBlock(n, fn)
	case *parser.WhileStmt:
		walkExpr(n.Cond, fn)
		walkBlock(n.Body, fn)
	case *parser.DoWhileStmt:
		walkBlock(n.Body, fn)
		walkExpr(n.Cond, fn)
	case *parser.ForStmt:
		if n.Init != nil {
			walkStmt(n.Init, fn)
		}
		walkExpr(n.Cond, fn)
		walkExpr(n.Post, fn)
		walkBlock(n.Body, fn)
	case *parser.SwitchStmt:
		walkExpr(n.Tag, fn)
		for _, c := range n.Cases {
			walkExpr(c.Value, fn)
			for _, cs := range c.Body {
				walkStmt(cs, fn)
			}
		}
	case *parser.ReturnStmt:
		walkExpr(n.Value, fn)
	}
}

func walkExpr(e parser.Expr, fn func(*parser.CallExpr)) {
	if e == nil {
		return
	}
	switch n := e.(type) {
	case *parser.CallExpr:
		fn(n)
		walkExpr(n.Callee, fn)
		for _, a := range n.Args {
			walkExpr(a, fn)
		}
	case *parser.MemberExpr:
		walkExpr(n.Object, fn)
	case *parser.BinaryExpr:
		walkExpr(n.Left, fn)
		walkExpr(n.Right, fn)
	case *parser.UnaryExpr:
		walkExpr(n.X, fn)
	case *parser.PostfixExpr:
		walkExpr(n.X, fn)
	case *parser.AssignExpr:
		walkExpr(n.Target, fn)
		walkExpr(n.Value, fn)
	case *parser.IndexExpr:
		walkExpr(n.Object, fn)
		walkExpr(n.Index, fn)
	}
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// pairedAgroom returns the .agroom path that pairs with a given .agscript path.
// "rooms/start/start.agscript" → "rooms/start/start.agroom"
func pairedAgroom(scriptRel string) string {
	dir := filepath.Dir(scriptRel)
	stem := strings.TrimSuffix(filepath.Base(scriptRel), ".agscript")
	return filepath.Join(dir, stem+".agroom")
}

func fileExists(root, rel string) bool {
	_, err := os.Stat(root + string(os.PathSeparator) + rel)
	return err == nil
}

func hasCameraName(rd *room.RoomData, name string) bool {
	for _, cam := range rd.Cameras {
		if cam.Name == name {
			return true
		}
	}
	return false
}
