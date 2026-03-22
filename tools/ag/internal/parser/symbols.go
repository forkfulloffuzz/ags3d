// symbols.go — T10: symbol table construction and name resolution for AGS-spirit.
//
// BuildSymbolTable runs two passes over a *File:
//
//	Pass 1 (collect): every declaration at file scope and inside namespace
//	blocks is inserted into SymbolTable.Globals or SymbolTable.Namespaces.
//
//	Pass 2 (resolve): identifier references inside function bodies are looked
//	up in a lexical scope chain.  Unresolvable names produce SymErrors.
//	The resolved type of every expression node is recorded in ExprTypes.
package parser

import (
	"fmt"

	"github.com/ags3d/ag/internal/scanner"
)

// -------------------------------------------------------------------
// SymKind
// -------------------------------------------------------------------

// SymKind classifies a symbol by how it was declared.
type SymKind int

const (
	KindFunction  SymKind = iota // function or event handler
	KindVar                      // file-scope or namespace-scope variable
	KindParam                    // function parameter
	KindLocal                    // block-local variable declared with VarDecl
	KindEnum                     // enum type name
	KindEnumMember               // individual named enum constant
	KindNamespace                // namespace name
)

func (k SymKind) String() string {
	switch k {
	case KindFunction:
		return "function"
	case KindVar:
		return "var"
	case KindParam:
		return "param"
	case KindLocal:
		return "local"
	case KindEnum:
		return "enum"
	case KindEnumMember:
		return "enum-member"
	case KindNamespace:
		return "namespace"
	}
	return fmt.Sprintf("SymKind(%d)", int(k))
}

// -------------------------------------------------------------------
// Symbol
// -------------------------------------------------------------------

// Symbol is a single entry in the symbol table.
type Symbol struct {
	Name      string
	Kind      SymKind
	Type      string        // declared type, or return type for functions; "" = void/unknown
	Namespace string        // "" = file scope; non-empty = owning namespace name
	IsExport  bool          // true for "export function" members
	IsBlocking bool         // propagated by T11
	Params    []Param       // parameter list for KindFunction symbols
	Decl      scanner.Token // position of the declaration
}

// -------------------------------------------------------------------
// SymbolTable
// -------------------------------------------------------------------

// SymbolTable is the output of BuildSymbolTable.
// It is safe to read from multiple goroutines after construction.
type SymbolTable struct {
	// Globals holds file-scope symbols: functions, top-level vars, enum types,
	// enum members, and namespace names.
	Globals map[string]*Symbol

	// Namespaces maps each namespace name to its member symbols.
	// A namespace name that appears in multiple files is merged into one entry.
	Namespaces map[string]map[string]*Symbol

	// ExprTypes maps each Expr node (by pointer identity) to its resolved type
	// string.  Entries are omitted for nodes whose type could not be inferred.
	ExprTypes map[Expr]string
}

// Lookup searches for a name at file scope, then in every namespace, returning
// the first match.  It is a convenience helper for callers that need a
// post-resolution lookup (e.g. the emitter, VIZ-03).
func (st *SymbolTable) Lookup(name string) (*Symbol, bool) {
	if sym, ok := st.Globals[name]; ok {
		return sym, true
	}
	for _, members := range st.Namespaces {
		if sym, ok := members[name]; ok {
			return sym, true
		}
	}
	return nil, false
}

// LookupMember looks up a member of a specific namespace.
func (st *SymbolTable) LookupMember(ns, name string) (*Symbol, bool) {
	members, ok := st.Namespaces[ns]
	if !ok {
		return nil, false
	}
	sym, ok := members[name]
	return sym, ok
}

// -------------------------------------------------------------------
// SymError
// -------------------------------------------------------------------

// SymError is a single name-resolution diagnostic produced by T10.
//
// Severity is "error" for definite problems (duplicate declarations, undefined
// member on a locally-known namespace) and "warning" for potentially valid
// cross-file references (bare identifier not found in the current file's scope).
type SymError struct {
	File     string
	Line     int
	Column   int
	Severity string // "error" | "warning"
	Message  string
}

func (e *SymError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", e.File, e.Line, e.Column, e.Severity, e.Message)
}

// -------------------------------------------------------------------
// BuildSymbolTable — public entry point
// -------------------------------------------------------------------

// BuildSymbolTable runs the two-pass symbol resolution over a single parsed
// file and returns the populated SymbolTable plus any resolution errors.
// A partial table is always returned even when errors are present.
func BuildSymbolTable(f *File) (*SymbolTable, []*SymError) {
	st := &SymbolTable{
		Globals:    make(map[string]*Symbol),
		Namespaces: make(map[string]map[string]*Symbol),
		ExprTypes:  make(map[Expr]string),
	}
	b := &symBuilder{file: f.Path, st: st}
	b.collect(f)
	b.resolve(f)
	return st, b.errors
}

// -------------------------------------------------------------------
// AGS engine built-ins
// -------------------------------------------------------------------

// engineBuiltinScope returns a scope pre-populated with well-known AGS engine
// functions and global objects. This is the outermost scope in the chain so
// user declarations shadow built-ins without errors.
//
// Source: Adventure Game Studio manual — Script API reference.
func engineBuiltinScope() *scope {
	s := newScope(nil)
	fn := func(name, returnType string, params ...Param) {
		s.define(name, &Symbol{
			Name:   name,
			Kind:   KindFunction,
			Type:   returnType,
			Params: params,
		})
	}
	v := func(name, typ string) {
		s.define(name, &Symbol{Name: name, Kind: KindVar, Type: typ})
	}

	// --- Blocking game-flow functions ---
	fn("Wait", "", Param{Type: "int", Name: "ticks"})
	fn("WaitKey", "", Param{Type: "int", Name: "ticks"})
	fn("WaitMouse", "", Param{Type: "int", Name: "ticks"})
	fn("WaitInput", "", Param{Type: "int", Name: "ticks"})
	fn("WalkCharacterTo", "", Param{Type: "Character", Name: "char"}, Param{Type: "int", Name: "x"}, Param{Type: "int", Name: "y"})

	// --- Display / speech ---
	fn("Display", "", Param{Type: "string", Name: "message"})
	fn("DisplayMessage", "", Param{Type: "int", Name: "msgNum"})
	fn("DisplayAt", "", Param{Type: "int", Name: "x"}, Param{Type: "int", Name: "y"}, Param{Type: "int", Name: "width"}, Param{Type: "string", Name: "message"})

	// --- Screen effects ---
	fn("FadeIn", "", Param{Type: "int", Name: "speed"})
	fn("FadeOut", "", Param{Type: "int", Name: "speed"})
	fn("ShakeScreen", "", Param{Type: "int", Name: "amount"})
	fn("SetBackgroundFrame", "", Param{Type: "int", Name: "frame"})

	// --- Room / game state ---
	fn("GiveScore", "", Param{Type: "int", Name: "points"})
	fn("EndGame", "", Param{Type: "int", Name: "restart"})
	fn("RestartGame", "")
	fn("QuitGame", "", Param{Type: "int", Name: "prompt"})
	fn("ChangeRoom", "", Param{Type: "int", Name: "room"})
	fn("ChangeRoomAutoPosition", "", Param{Type: "int", Name: "room"})

	// --- Inventory ---
	fn("GiveInventory", "", Param{Type: "int", Name: "item"})
	fn("LoseInventory", "", Param{Type: "int", Name: "item"})
	fn("HasInventoryItem", "bool", Param{Type: "int", Name: "item"})

	// --- Random / math ---
	fn("Random", "int", Param{Type: "int", Name: "max"})
	fn("Abs", "int", Param{Type: "int", Name: "n"})

	// --- String ---
	fn("StrLen", "int", Param{Type: "string", Name: "s"})
	fn("StrContains", "int", Param{Type: "string", Name: "haystack"}, Param{Type: "string", Name: "needle"})
	fn("StrToInt", "int", Param{Type: "string", Name: "s"})
	fn("IntToFloat", "float", Param{Type: "int", Name: "n"})
	fn("FloatToInt", "int", Param{Type: "float", Name: "f"})

	// --- Global game objects (accessed as bare identifiers in older scripts) ---
	v("player", "Character")
	v("game", "GameState")
	v("mouse", "Mouse")
	v("system", "System")

	return s
}

// -------------------------------------------------------------------
// symBuilder — internal builder
// -------------------------------------------------------------------

type symBuilder struct {
	file   string
	st     *SymbolTable
	errors []*SymError
}

func (b *symBuilder) errorf(pos scanner.Token, format string, args ...any) {
	b.errors = append(b.errors, &SymError{
		File:     b.file,
		Line:     pos.Line,
		Column:   pos.Column,
		Severity: "error",
		Message:  fmt.Sprintf(format, args...),
	})
}

// warningf records a diagnostic at warning severity.
// Used for undefined bare identifiers that may be valid cross-file references.
func (b *symBuilder) warningf(pos scanner.Token, format string, args ...any) {
	b.errors = append(b.errors, &SymError{
		File:     b.file,
		Line:     pos.Line,
		Column:   pos.Column,
		Severity: "warning",
		Message:  fmt.Sprintf(format, args...),
	})
}

// -------------------------------------------------------------------
// Pass 1 — collect
// -------------------------------------------------------------------

func (b *symBuilder) collect(f *File) {
	for _, decl := range f.Decls {
		b.collectDecl(decl, "")
	}
}

func (b *symBuilder) collectDecl(decl Decl, ns string) {
	switch d := decl.(type) {

	case *FunctionDecl:
		sym := &Symbol{
			Name:      d.Name,
			Kind:      KindFunction,
			Type:      d.ReturnType,
			Namespace: ns,
			IsExport:  d.IsExport,
			Params:    d.Params,
			Decl:      d.Pos,
		}
		b.insertSymbol(d.Name, sym, d.Pos, ns)

	case *NamespaceDecl:
		// Register the namespace name in the global scope.
		if _, exists := b.st.Globals[d.Name]; !exists {
			b.st.Globals[d.Name] = &Symbol{
				Name: d.Name,
				Kind: KindNamespace,
				Decl: d.Pos,
			}
		}
		if _, exists := b.st.Namespaces[d.Name]; !exists {
			b.st.Namespaces[d.Name] = make(map[string]*Symbol)
		}
		for _, m := range d.Members {
			b.collectDecl(m, d.Name)
		}

	case *EnumDecl:
		enumSym := &Symbol{
			Name:      d.Name,
			Kind:      KindEnum,
			Namespace: ns,
			Decl:      d.Pos,
		}
		b.insertSymbol(d.Name, enumSym, d.Pos, ns)

		// Each member is a named constant visible in the same scope as the enum.
		for _, m := range d.Members {
			msym := &Symbol{
				Name:      m.Name,
				Kind:      KindEnumMember,
				Type:      d.Name, // member's type is the enum's name
				Namespace: ns,
				Decl:      m.Pos,
			}
			b.insertSymbol(m.Name, msym, m.Pos, ns)
		}

	case *TopVarDecl:
		if d.Decl == nil {
			return
		}
		sym := &Symbol{
			Name:      d.Decl.Name,
			Kind:      KindVar,
			Type:      d.Decl.Type,
			Namespace: ns,
			Decl:      d.Decl.Pos,
		}
		b.insertSymbol(d.Decl.Name, sym, d.Decl.Pos, ns)
	}
}

// insertSymbol puts sym into the correct map (Globals or a namespace members
// map) and reports a redeclaration error if the name is already present.
func (b *symBuilder) insertSymbol(name string, sym *Symbol, pos scanner.Token, ns string) {
	if ns == "" {
		if prev, exists := b.st.Globals[name]; exists {
			b.errorf(pos, "redeclared %s %q (previous declaration at %d:%d)",
				sym.Kind, name, prev.Decl.Line, prev.Decl.Column)
			return
		}
		b.st.Globals[name] = sym
	} else {
		members := b.st.Namespaces[ns]
		if prev, exists := members[name]; exists {
			b.errorf(pos, "redeclared %s %q in namespace %q (previous declaration at %d:%d)",
				sym.Kind, name, ns, prev.Decl.Line, prev.Decl.Column)
			return
		}
		members[name] = sym
	}
}

// -------------------------------------------------------------------
// Pass 2 — resolve
// -------------------------------------------------------------------

func (b *symBuilder) resolve(f *File) {
	// Outermost scope: AGS engine built-ins (Wait, Display, player, …).
	// User declarations in the file scope shadow these without conflict.
	builtins := engineBuiltinScope()

	// File scope sits on top of built-ins.
	global := newScope(builtins)
	for name, sym := range b.st.Globals {
		global.define(name, sym)
	}
	for _, decl := range f.Decls {
		b.resolveDecl(decl, global)
	}
}

func (b *symBuilder) resolveDecl(decl Decl, outer *scope) {
	switch d := decl.(type) {

	case *FunctionDecl:
		fnScope := newScope(outer)
		for _, p := range d.Params {
			fnScope.define(p.Name, &Symbol{
				Name: p.Name,
				Kind: KindParam,
				Type: p.Type,
				Decl: p.Pos,
			})
		}
		if d.Body != nil {
			b.resolveBlock(d.Body, fnScope)
		}

	case *NamespaceDecl:
		// Namespace members see both the global scope and all sibling members.
		nsScope := newScope(outer)
		for name, sym := range b.st.Namespaces[d.Name] {
			nsScope.define(name, sym)
		}
		for _, m := range d.Members {
			b.resolveDecl(m, nsScope)
		}

	case *EnumDecl:
		for _, m := range d.Members {
			if m.Value != nil {
				b.resolveExpr(m.Value, outer)
			}
		}

	case *TopVarDecl:
		if d.Decl != nil && d.Decl.Init != nil {
			b.resolveExpr(d.Decl.Init, outer)
		}
	}
}

func (b *symBuilder) resolveBlock(block *Block, outer *scope) {
	blockScope := newScope(outer)
	for _, stmt := range block.Stmts {
		b.resolveStmt(stmt, blockScope)
	}
}

func (b *symBuilder) resolveStmt(stmt Stmt, sc *scope) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {

	case *Block:
		b.resolveBlock(s, sc)

	case *VarDecl:
		// Resolve the initialiser before defining the variable so that
		// "int x = x;" correctly errors (x not yet in scope).
		if s.Init != nil {
			b.resolveExpr(s.Init, sc)
		}
		sc.define(s.Name, &Symbol{
			Name: s.Name,
			Kind: KindLocal,
			Type: s.Type,
			Decl: s.Pos,
		})

	case *IfStmt:
		b.resolveExpr(s.Cond, sc)
		if s.Then != nil {
			b.resolveBlock(s.Then, sc)
		}
		if s.Else != nil {
			b.resolveStmt(s.Else, sc)
		}

	case *WhileStmt:
		b.resolveExpr(s.Cond, sc)
		if s.Body != nil {
			b.resolveBlock(s.Body, sc)
		}

	case *DoWhileStmt:
		if s.Body != nil {
			b.resolveBlock(s.Body, sc)
		}
		b.resolveExpr(s.Cond, sc)

	case *ForStmt:
		// The for-init variable is scoped to the for statement itself.
		forScope := newScope(sc)
		if s.Init != nil {
			b.resolveStmt(s.Init, forScope)
		}
		if s.Cond != nil {
			b.resolveExpr(s.Cond, forScope)
		}
		if s.Post != nil {
			b.resolveExpr(s.Post, forScope)
		}
		if s.Body != nil {
			b.resolveBlock(s.Body, forScope)
		}

	case *SwitchStmt:
		b.resolveExpr(s.Tag, sc)
		for _, c := range s.Cases {
			if c.Value != nil {
				b.resolveExpr(c.Value, sc)
			}
			for _, cs := range c.Body {
				b.resolveStmt(cs, sc)
			}
		}

	case *ReturnStmt:
		if s.Value != nil {
			b.resolveExpr(s.Value, sc)
		}

	case *ExprStmt:
		if s.X != nil {
			b.resolveExpr(s.X, sc)
		}

	case *BreakStmt, *ContinueStmt:
		// No names to resolve.
	}
}

// resolveExpr resolves all identifier references inside expr, records the
// inferred type in st.ExprTypes, and returns the inferred type string.
func (b *symBuilder) resolveExpr(expr Expr, sc *scope) string {
	if expr == nil {
		return ""
	}
	typ := b.inferType(expr, sc)
	if typ != "" {
		b.st.ExprTypes[expr] = typ
	}
	return typ
}

// inferType performs type inference for one expression node.
// It recurses via resolveExpr so every sub-expression is also annotated.
func (b *symBuilder) inferType(expr Expr, sc *scope) string {
	switch e := expr.(type) {

	case *Literal:
		return e.Kind // "int" | "float" | "string" | "bool" | "null"

	case *Identifier:
		sym, ok := sc.lookup(e.Name)
		if !ok {
			// Warn rather than error: the name may be defined in another file
			// (cross-file reference) or be an engine built-in not yet in scope.
			b.warningf(e.Pos, "undefined: %q", e.Name)
			return ""
		}
		return sym.Type

	case *GlobalExpr:
		// global.X is always valid; the actual type is resolved at T32.
		return ""

	case *AssignExpr:
		// Resolve target (may itself be a MemberExpr, IndexExpr, etc.)
		b.resolveExpr(e.Target, sc)
		return b.resolveExpr(e.Value, sc)

	case *BinaryExpr:
		left := b.resolveExpr(e.Left, sc)
		right := b.resolveExpr(e.Right, sc)
		return binaryResultType(e.Op, left, right)

	case *UnaryExpr:
		inner := b.resolveExpr(e.X, sc)
		return unaryResultType(e.Op, inner)

	case *PostfixExpr:
		return b.resolveExpr(e.X, sc)

	case *CallExpr:
		calleeType := b.resolveExpr(e.Callee, sc)
		for _, arg := range e.Args {
			b.resolveExpr(arg, sc)
		}
		// The type of a call expression is the return type of the callee.
		// resolveExpr on the callee (Identifier or MemberExpr) already
		// returns the function's declared return type.
		return calleeType

	case *MemberExpr:
		return b.resolveMemberExpr(e, sc)

	case *IndexExpr:
		b.resolveExpr(e.Object, sc)
		b.resolveExpr(e.Index, sc)
		return "" // element type requires a full type system

	}
	return ""
}

// resolveMemberExpr resolves "Object.Field" access.
//
// Special case: when Object is an Identifier naming a known namespace, the
// field is looked up inside that namespace.  For all other objects the field
// type is unknown (would need a complete type system) and resolution is
// best-effort — the object expression is still resolved for side-effects.
func (b *symBuilder) resolveMemberExpr(e *MemberExpr, sc *scope) string {
	if ident, ok := e.Object.(*Identifier); ok {
		if members, isNS := b.st.Namespaces[ident.Name]; isNS {
			// Object names a namespace — mark it and look up the field.
			b.st.ExprTypes[e.Object] = "namespace"
			sym, found := members[e.Field]
			if !found {
				b.errorf(e.Pos, "undefined member %q in namespace %q", e.Field, ident.Name)
				return ""
			}
			return sym.Type
		}
	}
	// General case — resolve the object and leave the field type unknown.
	b.resolveExpr(e.Object, sc)
	return ""
}

// -------------------------------------------------------------------
// Type inference helpers
// -------------------------------------------------------------------

// binaryResultType returns the result type of a binary operation.
func binaryResultType(op, left, right string) string {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=", "&&", "||":
		return "bool"
	case "+", "-", "*", "/", "%":
		if left == "float" || right == "float" {
			return "float"
		}
		return "int"
	case "&", "|", "^", "<<", ">>":
		return "int"
	}
	if left != "" {
		return left
	}
	return right
}

// unaryResultType returns the result type of a unary operation.
func unaryResultType(op, inner string) string {
	switch op {
	case "!":
		return "bool"
	case "~":
		return "int"
	}
	return inner // -, prefix++/--, postfix++/-- preserve the operand type
}

// -------------------------------------------------------------------
// Lexical scope chain
// -------------------------------------------------------------------

// scope is a single frame in the lexical scope chain built during pass 2.
type scope struct {
	syms   map[string]*Symbol
	parent *scope
}

func newScope(parent *scope) *scope {
	return &scope{syms: make(map[string]*Symbol), parent: parent}
}

func (s *scope) define(name string, sym *Symbol) {
	s.syms[name] = sym
}

// lookup searches this scope then its ancestors, returning the innermost match.
func (s *scope) lookup(name string) (*Symbol, bool) {
	for cur := s; cur != nil; cur = cur.parent {
		if sym, ok := cur.syms[name]; ok {
			return sym, true
		}
	}
	return nil, false
}
