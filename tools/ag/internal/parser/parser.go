// Package parser implements the AGS-spirit recursive descent parser,
// AST node types, symbol table, and blocking call annotation.
// Stub — full implementation in T08–T11.
package parser

import (
	"fmt"

	"github.com/ags3d/ag/internal/scanner"
)

// -------------------------------------------------------------------
// AST node hierarchy (T08)
// -------------------------------------------------------------------

// Node is the interface implemented by all AST nodes.
type Node interface{ nodeMarker() }

// File is the root node of a parsed .agscript file.
type File struct {
	Path  string
	Decls []Decl
}

func (*File) nodeMarker() {}

// Decl covers all top-level declarations.
type Decl interface {
	Node
	declMarker()
}

// FunctionDecl represents a function or event handler declaration.
// IsExport is set when the "export" modifier is present inside a namespace.
// Using export outside a namespace is a semantic error caught at T10.
// Engine event handlers (room_Load, hotspot_NAME_Interact, etc.) are always
// implicitly accessible to the engine regardless of this flag.
type FunctionDecl struct {
	Name       string
	Params     []Param
	ReturnType string
	Body       *Block
	IsExport   bool // "export function …" — valid only inside a NamespaceDecl
	IsBlocking bool // propagated from blocking call sites inside body (T11)
	Pos        scanner.Token
}

// NamespaceDecl groups exported and private functions under a named scope.
// Multiple files may contribute to the same namespace; the symbol table (T10)
// merges them and enforces that no two files export the same function name.
type NamespaceDecl struct {
	Name    string
	Members []Decl // *FunctionDecl | *EnumDecl | *TopVarDecl
	Pos     scanner.Token
}

func (*NamespaceDecl) nodeMarker() {}
func (*NamespaceDecl) declMarker() {}

func (*FunctionDecl) nodeMarker() {}
func (*FunctionDecl) declMarker() {}

// Param is a single function parameter.
type Param struct {
	Name string
	Type string
	Pos  scanner.Token
}

// EnumDecl is a top-level enum declaration.
//
//	enum Direction { eNorth = 0, eSouth, eEast, eWest }
type EnumDecl struct {
	Name    string
	Members []EnumMember
	Pos     scanner.Token
}

func (*EnumDecl) nodeMarker() {}
func (*EnumDecl) declMarker() {}

// EnumMember is a single named constant inside an EnumDecl.
type EnumMember struct {
	Name  string
	Value Expr // nil if no explicit initialiser
	Pos   scanner.Token
}

// TopVarDecl is a file-level variable declaration.
type TopVarDecl struct {
	Decl *VarDecl
}

func (*TopVarDecl) nodeMarker() {}
func (*TopVarDecl) declMarker() {}

// -------------------------------------------------------------------
// Statements
// -------------------------------------------------------------------

// Stmt covers all statement nodes.
type Stmt interface {
	Node
	stmtMarker()
}

// Block is a brace-delimited sequence of statements.
type Block struct {
	Stmts []Stmt
	Pos   scanner.Token
}

func (*Block) nodeMarker() {}
func (*Block) stmtMarker() {}

// VarDecl is a typed variable declaration with optional initialiser.
// Appears both as a top-level declaration and inside blocks.
type VarDecl struct {
	Type string
	Name string
	Init Expr // nil if no initialiser
	Pos  scanner.Token
}

func (*VarDecl) nodeMarker() {}
func (*VarDecl) stmtMarker() {}

// IfStmt is an if / else-if / else chain.
type IfStmt struct {
	Cond Expr
	Then *Block
	Else Stmt // nil | *Block | *IfStmt (else-if chain)
	Pos  scanner.Token
}

func (*IfStmt) nodeMarker() {}
func (*IfStmt) stmtMarker() {}

// WhileStmt is a while loop.
type WhileStmt struct {
	Cond Expr
	Body *Block
	Pos  scanner.Token
}

func (*WhileStmt) nodeMarker() {}
func (*WhileStmt) stmtMarker() {}

// DoWhileStmt is a do { } while ( cond ) loop.
type DoWhileStmt struct {
	Body *Block
	Cond Expr
	Pos  scanner.Token
}

func (*DoWhileStmt) nodeMarker() {}
func (*DoWhileStmt) stmtMarker() {}

// ForStmt is a C-style for loop.
//
//	for ( Init ; Cond ; Post ) Body
//
// Init and Post may be nil.
type ForStmt struct {
	Init Stmt // *VarDecl | *ExprStmt | nil
	Cond Expr // nil means loop forever (break required)
	Post Expr // nil if absent
	Body *Block
	Pos  scanner.Token
}

func (*ForStmt) nodeMarker() {}
func (*ForStmt) stmtMarker() {}

// SwitchStmt is a switch / case / default block.
type SwitchStmt struct {
	Tag     Expr
	Cases   []*CaseClause
	Pos     scanner.Token
}

func (*SwitchStmt) nodeMarker() {}
func (*SwitchStmt) stmtMarker() {}

// CaseClause is a single case or default arm inside a SwitchStmt.
// Value is nil for the default arm.
type CaseClause struct {
	Value Expr // nil = default
	Body  []Stmt
	Pos   scanner.Token
}

func (*CaseClause) nodeMarker() {}

// ReturnStmt is a return statement with optional value.
type ReturnStmt struct {
	Value Expr // nil for bare return
	Pos   scanner.Token
}

func (*ReturnStmt) nodeMarker() {}
func (*ReturnStmt) stmtMarker() {}

// BreakStmt is a break statement (loop exit or switch exit).
type BreakStmt struct{ Pos scanner.Token }

func (*BreakStmt) nodeMarker() {}
func (*BreakStmt) stmtMarker() {}

// ContinueStmt is a continue statement (next loop iteration).
type ContinueStmt struct{ Pos scanner.Token }

func (*ContinueStmt) nodeMarker() {}
func (*ContinueStmt) stmtMarker() {}

// ExprStmt is an expression used as a statement.
type ExprStmt struct{ X Expr }

func (*ExprStmt) nodeMarker() {}
func (*ExprStmt) stmtMarker() {}

// -------------------------------------------------------------------
// Expressions
// -------------------------------------------------------------------

// Expr covers all expression nodes.
type Expr interface {
	Node
	exprMarker()
}

// AssignExpr is any assignment: =, +=, -=, *=, /=, %=, &=, |=, ^=, <<=, >>=
type AssignExpr struct {
	Op     string // "=", "+=", "-=", etc.
	Target Expr
	Value  Expr
	Pos    scanner.Token
}

func (*AssignExpr) nodeMarker() {}
func (*AssignExpr) exprMarker() {}

// BinaryExpr is any infix binary operation.
type BinaryExpr struct {
	Op    string // "+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=", "&&", "||", "&", "|", "^", "<<", ">>"
	Left  Expr
	Right Expr
	Pos   scanner.Token
}

func (*BinaryExpr) nodeMarker() {}
func (*BinaryExpr) exprMarker() {}

// UnaryExpr is a prefix unary operation: !, -, ~, ++, --
type UnaryExpr struct {
	Op  string // "!", "-", "~", "++", "--"
	X   Expr
	Pos scanner.Token
}

func (*UnaryExpr) nodeMarker() {}
func (*UnaryExpr) exprMarker() {}

// PostfixExpr is a postfix ++ or -- applied to an expression.
type PostfixExpr struct {
	Op  string // "++" | "--"
	X   Expr
	Pos scanner.Token
}

func (*PostfixExpr) nodeMarker() {}
func (*PostfixExpr) exprMarker() {}

// CallExpr is a function or method invocation.
type CallExpr struct {
	Callee     Expr
	Args       []Expr
	IsBlocking bool // annotated during symbol resolution (T11)
	Pos        scanner.Token
}

func (*CallExpr) nodeMarker() {}
func (*CallExpr) exprMarker() {}

// IndexExpr is an array subscript: expr[index].
type IndexExpr struct {
	Object Expr
	Index  Expr
	Pos    scanner.Token
}

func (*IndexExpr) nodeMarker() {}
func (*IndexExpr) exprMarker() {}

// MemberExpr is a field access: object.field
type MemberExpr struct {
	Object Expr
	Field  string
	Pos    scanner.Token
}

func (*MemberExpr) nodeMarker() {}
func (*MemberExpr) exprMarker() {}

// Identifier is a bare name reference.
type Identifier struct {
	Name string
	Pos  scanner.Token
}

// GlobalExpr is an access into the engine-owned global namespace: global.NAME
// The emitter translates this to the appropriate AGSRuntime property call.
type GlobalExpr struct {
	Property string // "player", "room", "score", "camera", etc.
	Pos      scanner.Token
}

func (*GlobalExpr) nodeMarker() {}
func (*GlobalExpr) exprMarker() {}

func (*Identifier) nodeMarker() {}
func (*Identifier) exprMarker() {}

// Literal is an integer, float, string, bool, or null literal.
type Literal struct {
	Kind  string // "int" | "float" | "string" | "bool" | "null"
	Value string
	Pos   scanner.Token
}

func (*Literal) nodeMarker() {}
func (*Literal) exprMarker() {}

// -------------------------------------------------------------------
// Errors
// -------------------------------------------------------------------

// ParseError is a single parser error with source location.
type ParseError struct {
	File    string
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
}

// -------------------------------------------------------------------
// Parser (T09)
// -------------------------------------------------------------------

// Parser is a recursive descent parser for AGS-spirit source.
// Full implementation in T09.
type Parser struct {
	s      *scanner.Scanner
	cur    scanner.Token
	errors []*ParseError
}

// New creates a Parser over an already-created Scanner.
func New(s *scanner.Scanner) *Parser {
	p := &Parser{s: s}
	p.advance()
	return p
}

// Parse parses a complete .agscript file. Errors are collected; the
// returned File may be partial on error. Never panics on bad input.
// TODO(T09): implement full recursive descent parser.
func (p *Parser) Parse(file string) (*File, []*ParseError) {
	f := &File{Path: file}
	// TODO(T09): parse declarations
	return f, p.errors
}

func (p *Parser) advance() {
	p.cur = p.s.Next()
}
