// Package parser implements the AGS-spirit recursive descent parser,
// symbol table, and blocking call annotation.
//
// The AST is defined in this file (ast.go). The parser that builds it
// lives in parser.go (T09). The symbol table and blocking annotation
// live in symbols.go and blocking.go (T10–T11).
//
// Grammar reference: docs/grammar.md
package parser

import "github.com/ags3d/ag/internal/scanner"

// -------------------------------------------------------------------
// Root
// -------------------------------------------------------------------

// File is the root node of a parsed .agscript file.
type File struct {
	Path  string
	Decls []Decl
}

// -------------------------------------------------------------------
// Declarations
// -------------------------------------------------------------------

// Decl is implemented by every top-level declaration node.
type Decl interface {
	Node
	declMarker()
	DeclPos() scanner.Token
}

// FunctionDecl is a function or event handler declaration.
//
//	[ Type ] "function" Ident "(" [ ParamList ] ")" Block
//
// IsExport marks "export function" — only valid inside a NamespaceDecl.
// Using export at file level is caught by the symbol table (T10).
// Engine event handlers (room_Load, hotspot_NAME_Interact, etc.) are always
// accessible to the engine regardless of IsExport.
//
// IsBlocking is propagated from blocking call sites in Body (T11).
type FunctionDecl struct {
	ReturnType string // "" means void
	Name       string
	Params     []Param
	Body       *Block
	IsExport   bool
	IsBlocking bool // set by T11
	Pos        scanner.Token
}

func (*FunctionDecl) nodeMarker()            {}
func (*FunctionDecl) declMarker()            {}
func (f *FunctionDecl) DeclPos() scanner.Token { return f.Pos }

// NamespaceDecl groups functions and enums under a named scope.
//
//	"namespace" Ident "{" { NamespaceMember } "}"
//
// Multiple files may contribute to the same namespace; the symbol table (T10)
// merges them and errors on duplicate exported names.
type NamespaceDecl struct {
	Name    string
	Members []Decl // *FunctionDecl | *EnumDecl | *TopVarDecl
	Pos     scanner.Token
}

func (*NamespaceDecl) nodeMarker()            {}
func (*NamespaceDecl) declMarker()            {}
func (n *NamespaceDecl) DeclPos() scanner.Token { return n.Pos }

// EnumDecl is a top-level enum declaration.
//
//	"enum" Ident "{" EnumMember { "," EnumMember } [ "," ] "}" ";"
type EnumDecl struct {
	Name    string
	Members []EnumMember
	Pos     scanner.Token
}

func (*EnumDecl) nodeMarker()            {}
func (*EnumDecl) declMarker()            {}
func (e *EnumDecl) DeclPos() scanner.Token { return e.Pos }

// EnumMember is a single named constant inside an EnumDecl.
// Value is nil when no explicit initialiser is given (auto-assigned by T10).
type EnumMember struct {
	Name  string
	Value Expr // nil if implicit
	Pos   scanner.Token
}

// Param is a single function parameter.
type Param struct {
	Type string
	Name string
	Pos  scanner.Token
}

// TopVarDecl wraps a VarDecl for use at file or namespace scope.
//
//	Type Ident [ "=" Expr ] ";"
type TopVarDecl struct {
	Decl *VarDecl
}

func (*TopVarDecl) nodeMarker()            {}
func (*TopVarDecl) declMarker()            {}
func (t *TopVarDecl) DeclPos() scanner.Token { return t.Decl.Pos }

// -------------------------------------------------------------------
// Statements
// -------------------------------------------------------------------

// Stmt is implemented by every statement node.
type Stmt interface {
	Node
	stmtMarker()
}

// Block is a brace-delimited sequence of statements.
//
//	"{" { Stmt } "}"
type Block struct {
	Stmts []Stmt
	Pos   scanner.Token
}

func (*Block) nodeMarker() {}
func (*Block) stmtMarker() {}

// VarDecl is a typed variable declaration with optional initialiser.
// Used both inside blocks (as Stmt) and at file/namespace scope (wrapped in TopVarDecl).
//
//	Type Ident [ "=" Expr ]
type VarDecl struct {
	Type string
	Name string
	Init Expr // nil if no initialiser
	Pos  scanner.Token
}

func (*VarDecl) nodeMarker() {}
func (*VarDecl) stmtMarker() {}

// IfStmt is an if / else-if / else chain.
//
//	"if" "(" Expr ")" Block [ "else" ( IfStmt | Block ) ]
//
// Else is nil | *Block | *IfStmt.
// Representing else-if as *IfStmt makes the chain recursive and avoids
// a separate ElseIfStmt node.
type IfStmt struct {
	Cond Expr
	Then *Block
	Else Stmt // nil | *Block | *IfStmt
	Pos  scanner.Token
}

func (*IfStmt) nodeMarker() {}
func (*IfStmt) stmtMarker() {}

// WhileStmt is a while loop.
//
//	"while" "(" Expr ")" Block
type WhileStmt struct {
	Cond Expr
	Body *Block
	Pos  scanner.Token
}

func (*WhileStmt) nodeMarker() {}
func (*WhileStmt) stmtMarker() {}

// DoWhileStmt is a do..while loop.
//
//	"do" Block "while" "(" Expr ")" ";"
type DoWhileStmt struct {
	Body *Block
	Cond Expr
	Pos  scanner.Token
}

func (*DoWhileStmt) nodeMarker() {}
func (*DoWhileStmt) stmtMarker() {}

// ForStmt is a C-style for loop.
//
//	"for" "(" [ ForInit ] ";" [ Expr ] ";" [ Expr ] ")" Block
//
// Init is *VarDecl | *ExprStmt | nil.
// Cond is nil for an infinite loop (break required inside body).
// Post is nil when omitted.
type ForStmt struct {
	Init Stmt  // *VarDecl | *ExprStmt | nil
	Cond Expr  // nil = unconditional
	Post Expr  // nil if absent
	Body *Block
	Pos  scanner.Token
}

func (*ForStmt) nodeMarker() {}
func (*ForStmt) stmtMarker() {}

// SwitchStmt is a switch block.
//
//	"switch" "(" Expr ")" "{" { CaseClause } "}"
type SwitchStmt struct {
	Tag   Expr
	Cases []*CaseClause
	Pos   scanner.Token
}

func (*SwitchStmt) nodeMarker() {}
func (*SwitchStmt) stmtMarker() {}

// CaseClause is a single case or default arm inside a SwitchStmt.
//
//	( "case" Expr | "default" ) ":" { Stmt }
//
// Value is nil for the default arm.
// Fall-through is explicit — no implicit break between clauses.
type CaseClause struct {
	Value Expr // nil = default
	Body  []Stmt
	Pos   scanner.Token
}

func (*CaseClause) nodeMarker() {}

// ReturnStmt is a return statement.
//
//	"return" [ Expr ] ";"
//
// Value is nil for a bare return (only valid in void functions — checked at T10).
type ReturnStmt struct {
	Value Expr
	Pos   scanner.Token
}

func (*ReturnStmt) nodeMarker() {}
func (*ReturnStmt) stmtMarker() {}

// BreakStmt exits the nearest enclosing loop or switch.
// Valid only inside a loop or switch — checked by T10.
type BreakStmt struct{ Pos scanner.Token }

func (*BreakStmt) nodeMarker() {}
func (*BreakStmt) stmtMarker() {}

// ContinueStmt skips to the next iteration of the nearest enclosing loop.
// Valid only inside a loop — checked by T10.
type ContinueStmt struct{ Pos scanner.Token }

func (*ContinueStmt) nodeMarker() {}
func (*ContinueStmt) stmtMarker() {}

// ExprStmt is an expression used as a statement (typically a call).
//
//	Expr ";"
type ExprStmt struct {
	X   Expr
	Pos scanner.Token
}

func (*ExprStmt) nodeMarker() {}
func (*ExprStmt) stmtMarker() {}

// -------------------------------------------------------------------
// Expressions
// -------------------------------------------------------------------

// Expr is implemented by every expression node.
type Expr interface {
	Node
	exprMarker()
	ExprPos() scanner.Token
}

// AssignExpr is any assignment expression.
//
//	LogicOrExpr AssignOp AssignExpr   — right-associative
//
// Op is one of: = += -= *= /= %= &= |= ^= <<= >>=
type AssignExpr struct {
	Op     string
	Target Expr
	Value  Expr
	Pos    scanner.Token
}

func (*AssignExpr) nodeMarker()            {}
func (*AssignExpr) exprMarker()            {}
func (a *AssignExpr) ExprPos() scanner.Token { return a.Pos }

// BinaryExpr is any infix binary operation.
//
// Op is one of: + - * / % == != < <= > >= && || & | ^ << >>
type BinaryExpr struct {
	Op    string
	Left  Expr
	Right Expr
	Pos   scanner.Token
}

func (*BinaryExpr) nodeMarker()            {}
func (*BinaryExpr) exprMarker()            {}
func (b *BinaryExpr) ExprPos() scanner.Token { return b.Pos }

// UnaryExpr is a prefix unary operation.
//
// Op is one of: ! - ~ ++ --
type UnaryExpr struct {
	Op  string
	X   Expr
	Pos scanner.Token
}

func (*UnaryExpr) nodeMarker()            {}
func (*UnaryExpr) exprMarker()            {}
func (u *UnaryExpr) ExprPos() scanner.Token { return u.Pos }

// PostfixExpr is a postfix ++ or -- applied to an addressable expression.
//
// Op is "++" or "--".
type PostfixExpr struct {
	Op  string
	X   Expr
	Pos scanner.Token
}

func (*PostfixExpr) nodeMarker()            {}
func (*PostfixExpr) exprMarker()            {}
func (p *PostfixExpr) ExprPos() scanner.Token { return p.Pos }

// CallExpr is a function or method call.
//
//	PostfixExpr "(" [ ArgList ] ")"
//
// IsBlocking is set during symbol resolution (T11) — true when the callee
// is a known blocking built-in or a user function marked IsBlocking.
type CallExpr struct {
	Callee     Expr
	Args       []Expr
	IsBlocking bool
	Pos        scanner.Token
}

func (*CallExpr) nodeMarker()            {}
func (*CallExpr) exprMarker()            {}
func (c *CallExpr) ExprPos() scanner.Token { return c.Pos }

// IndexExpr is a subscript operation.
//
//	PostfixExpr "[" Expr "]"
type IndexExpr struct {
	Object Expr
	Index  Expr
	Pos    scanner.Token
}

func (*IndexExpr) nodeMarker()            {}
func (*IndexExpr) exprMarker()            {}
func (i *IndexExpr) ExprPos() scanner.Token { return i.Pos }

// MemberExpr is a field access.
//
//	PostfixExpr "." Ident
type MemberExpr struct {
	Object Expr
	Field  string
	Pos    scanner.Token
}

func (*MemberExpr) nodeMarker()            {}
func (*MemberExpr) exprMarker()            {}
func (m *MemberExpr) ExprPos() scanner.Token { return m.Pos }

// GlobalExpr is an access into the engine-owned global namespace.
//
//	"global" "." Ident
//
// The emitter maps this to the appropriate AGSRuntime property (T32).
// global itself cannot be assigned to — checked by T10.
type GlobalExpr struct {
	Property string
	Pos      scanner.Token
}

func (*GlobalExpr) nodeMarker()            {}
func (*GlobalExpr) exprMarker()            {}
func (g *GlobalExpr) ExprPos() scanner.Token { return g.Pos }

// Identifier is a bare name reference resolved by the symbol table (T10).
type Identifier struct {
	Name string
	Pos  scanner.Token
}

func (*Identifier) nodeMarker()            {}
func (*Identifier) exprMarker()            {}
func (i *Identifier) ExprPos() scanner.Token { return i.Pos }

// Literal is an integer, float, string, bool, or null literal.
//
// Kind is one of: "int" | "float" | "string" | "bool" | "null"
type Literal struct {
	Kind  string
	Value string
	Pos   scanner.Token
}

func (*Literal) nodeMarker()            {}
func (*Literal) exprMarker()            {}
func (l *Literal) ExprPos() scanner.Token { return l.Pos }

// -------------------------------------------------------------------
// Node (base interface)
// -------------------------------------------------------------------

// Node is the marker interface implemented by all AST nodes.
type Node interface{ nodeMarker() }
