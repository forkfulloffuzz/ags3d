// blocking.go — T11: blocking call identification and propagation.
//
// AnnotateBlocking performs two passes over a parsed *File:
//
//  1. Direct annotation: every CallExpr whose callee is a known blocking
//     built-in function or blocking method name gets CallExpr.IsBlocking=true.
//
//  2. Transitive propagation: if a FunctionDecl body contains any blocking
//     call (directly or through a call to another user-defined blocking
//     function), FunctionDecl.IsBlocking is set to true.  The propagation
//     repeats to fixed-point so arbitrarily deep call chains are covered.
//
// After both passes the IsBlocking flag on the matching Symbol entries in the
// SymbolTable is synced, so the emitter (T16) can query either the AST or the
// symbol table.
package parser

// -------------------------------------------------------------------
// Blocking built-in registries
// -------------------------------------------------------------------

// blockingGlobalFuncs is the set of globally-callable AGS functions that
// suspend execution until they complete.  Calls to these must be preceded
// by `await` in the emitted GDScript.
var blockingGlobalFuncs = map[string]bool{
	"Wait":           true,
	"WaitKey":        true,
	"WaitMouse":      true,
	"WaitInput":      true,
	"FadeIn":         true,
	"FadeOut":        true,
	"Display":        true,
	"DisplayMessage": true,
}

// blockingMethodNames is the set of method names (called on any object,
// typically a Character) that are always blocking.
var blockingMethodNames = map[string]bool{
	"WalkTo":         true,
	"WalkStraight":   true,
	"FaceTo":         true,
	"Say":            true,
	"Think":          true,
	"PlayAnimation":  true,
	"FaceDirection":  true,
	"FaceCharacter":  true,
	"FacePoint":      true,
	"RunInteraction": true,
}

// IsBlockingBuiltin reports whether name is a known blocking global function.
func IsBlockingBuiltin(name string) bool { return blockingGlobalFuncs[name] }

// IsBlockingMethod reports whether name is a known blocking method name.
func IsBlockingMethod(name string) bool { return blockingMethodNames[name] }

// -------------------------------------------------------------------
// AnnotateBlocking — public entry point
// -------------------------------------------------------------------

// AnnotateBlocking annotates the AST and updates st in-place.
// Call this after BuildSymbolTable.
func AnnotateBlocking(f *File, st *SymbolTable) {
	// Collect all *FunctionDecl nodes keyed by bare name for transitive lookup.
	allFuncs := collectAllFuncDecls(f)

	// Iterative fixed-point: repeat until no new FunctionDecl becomes blocking.
	for {
		changed := false
		for _, fd := range allFuncs {
			if annotateBlockingInFunc(fd, allFuncs) {
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Sync IsBlocking from AST FunctionDecl nodes back to SymbolTable entries.
	for _, fd := range allFuncs {
		if !fd.IsBlocking {
			continue
		}
		if sym, ok := st.Globals[fd.Name]; ok {
			sym.IsBlocking = true
			continue
		}
		for _, members := range st.Namespaces {
			if sym, ok := members[fd.Name]; ok {
				sym.IsBlocking = true
			}
		}
	}
}

// -------------------------------------------------------------------
// FunctionDecl collection
// -------------------------------------------------------------------

func collectAllFuncDecls(f *File) map[string]*FunctionDecl {
	out := make(map[string]*FunctionDecl)
	for _, d := range f.Decls {
		collectFuncDeclsFromDecl(d, out)
	}
	return out
}

func collectFuncDeclsFromDecl(d Decl, out map[string]*FunctionDecl) {
	switch v := d.(type) {
	case *FunctionDecl:
		out[v.Name] = v
	case *NamespaceDecl:
		for _, m := range v.Members {
			collectFuncDeclsFromDecl(m, out)
		}
	}
}

// -------------------------------------------------------------------
// Single-function annotation
// -------------------------------------------------------------------

// annotateBlockingInFunc walks fd.Body, marks any blocking CallExprs, and
// sets fd.IsBlocking if a blocking call was found.  Returns true if
// fd.IsBlocking was newly set during this call (used by the fixed-point loop).
func annotateBlockingInFunc(fd *FunctionDecl, allFuncs map[string]*FunctionDecl) bool {
	if fd.Body == nil {
		return false
	}
	wasBlocking := fd.IsBlocking
	if walkBlock(fd.Body, allFuncs) {
		fd.IsBlocking = true
	}
	return fd.IsBlocking && !wasBlocking
}

// -------------------------------------------------------------------
// AST walkers — return true if a blocking call was found/annotated
// -------------------------------------------------------------------

func walkBlock(b *Block, fns map[string]*FunctionDecl) bool {
	if b == nil {
		return false
	}
	found := false
	for _, s := range b.Stmts {
		if walkStmt(s, fns) {
			found = true
		}
	}
	return found
}

func walkStmt(s Stmt, fns map[string]*FunctionDecl) bool {
	if s == nil {
		return false
	}
	switch v := s.(type) {
	case *Block:
		return walkBlock(v, fns)
	case *VarDecl:
		return walkExpr(v.Init, fns)
	case *ExprStmt:
		return walkExpr(v.X, fns)
	case *ReturnStmt:
		return walkExpr(v.Value, fns)
	case *IfStmt:
		a := walkExpr(v.Cond, fns)
		b := walkBlock(v.Then, fns)
		c := walkStmt(v.Else, fns)
		return a || b || c
	case *WhileStmt:
		return walkExpr(v.Cond, fns) || walkBlock(v.Body, fns)
	case *DoWhileStmt:
		return walkBlock(v.Body, fns) || walkExpr(v.Cond, fns)
	case *ForStmt:
		a := walkStmt(v.Init, fns)
		b := walkExpr(v.Cond, fns)
		c := walkExpr(v.Post, fns)
		d := walkBlock(v.Body, fns)
		return a || b || c || d
	case *SwitchStmt:
		found := walkExpr(v.Tag, fns)
		for _, cl := range v.Cases {
			if walkExpr(cl.Value, fns) {
				found = true
			}
			for _, cs := range cl.Body {
				if walkStmt(cs, fns) {
					found = true
				}
			}
		}
		return found
	}
	return false
}

// walkExpr recurses into expr, annotates any blocking CallExprs it finds,
// and returns true if the outermost expression is itself a blocking call.
// Nested blocking calls (e.g. inside arguments) are annotated but do NOT
// make the outer call blocking — only the direct callee determines that.
func walkExpr(e Expr, fns map[string]*FunctionDecl) bool {
	if e == nil {
		return false
	}
	switch v := e.(type) {
	case *CallExpr:
		blocking := calleeIsBlocking(v.Callee, fns)
		// Recurse into callee expression (handles chained calls in callee pos).
		walkExpr(v.Callee, fns)
		// Recurse into args to annotate any nested blocking sub-calls.
		for _, arg := range v.Args {
			walkExpr(arg, fns)
		}
		if blocking {
			v.IsBlocking = true
		}
		return blocking

	case *AssignExpr:
		a := walkExpr(v.Target, fns)
		b := walkExpr(v.Value, fns)
		return a || b
	case *BinaryExpr:
		return walkExpr(v.Left, fns) || walkExpr(v.Right, fns)
	case *UnaryExpr:
		return walkExpr(v.X, fns)
	case *PostfixExpr:
		return walkExpr(v.X, fns)
	case *MemberExpr:
		return walkExpr(v.Object, fns)
	case *IndexExpr:
		return walkExpr(v.Object, fns) || walkExpr(v.Index, fns)
	}
	return false
}

// calleeIsBlocking reports whether callee refers to a blocking function.
//
//   - Identifier → check global blocking built-ins, then user-defined functions.
//   - MemberExpr → check if the field name is a blocking method.
func calleeIsBlocking(callee Expr, fns map[string]*FunctionDecl) bool {
	switch c := callee.(type) {
	case *Identifier:
		if blockingGlobalFuncs[c.Name] {
			return true
		}
		if fd, ok := fns[c.Name]; ok && fd.IsBlocking {
			return true
		}
	case *MemberExpr:
		return blockingMethodNames[c.Field]
	}
	return false
}
