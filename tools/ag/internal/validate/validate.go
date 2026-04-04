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
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ags3d/ag/internal/char"
	"github.com/ags3d/ag/internal/item"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/room"
	"github.com/ags3d/ag/internal/scanner"
)

// Issue is a single validation finding with source location.
type Issue struct {
	File     string
	Line     int // 0 = no line info
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
	}

	return issues, nil
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
