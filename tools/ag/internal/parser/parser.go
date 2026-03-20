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
type FunctionDecl struct {
	Name       string
	Params     []Param
	ReturnType string
	Body       *Block
	IsBlocking bool // propagated from blocking call sites inside body
	Pos        scanner.Token
}

func (*FunctionDecl) nodeMarker() {}
func (*FunctionDecl) declMarker() {}

// Param is a single function parameter.
type Param struct {
	Name string
	Type string
}

// Stmt covers all statement nodes.
type Stmt interface {
	Node
	stmtMarker()
}

type Block struct{ Stmts []Stmt }

func (*Block) nodeMarker() {}
func (*Block) stmtMarker() {}

type IfStmt struct {
	Cond     Expr
	Then     *Block
	Else     *Block // nil if no else branch
}

func (*IfStmt) nodeMarker() {}
func (*IfStmt) stmtMarker() {}

type WhileStmt struct {
	Cond Expr
	Body *Block
}

func (*WhileStmt) nodeMarker() {}
func (*WhileStmt) stmtMarker() {}

type AssignStmt struct {
	Target Expr
	Value  Expr
}

func (*AssignStmt) nodeMarker() {}
func (*AssignStmt) stmtMarker() {}

type ReturnStmt struct{ Value Expr }

func (*ReturnStmt) nodeMarker() {}
func (*ReturnStmt) stmtMarker() {}

type ExprStmt struct{ X Expr }

func (*ExprStmt) nodeMarker() {}
func (*ExprStmt) stmtMarker() {}

// Expr covers all expression nodes.
type Expr interface {
	Node
	exprMarker()
}

type CallExpr struct {
	Callee     Expr
	Args       []Expr
	IsBlocking bool // set during symbol resolution (T11)
}

func (*CallExpr) nodeMarker() {}
func (*CallExpr) exprMarker() {}

type MemberExpr struct {
	Object Expr
	Field  string
}

func (*MemberExpr) nodeMarker() {}
func (*MemberExpr) exprMarker() {}

type BinaryExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

func (*BinaryExpr) nodeMarker() {}
func (*BinaryExpr) exprMarker() {}

type UnaryExpr struct {
	Op string
	X  Expr
}

func (*UnaryExpr) nodeMarker() {}
func (*UnaryExpr) exprMarker() {}

type Identifier struct {
	Name string
	Pos  scanner.Token
}

func (*Identifier) nodeMarker() {}
func (*Identifier) exprMarker() {}

type Literal struct {
	Kind  string // "int" | "float" | "string" | "bool"
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

// New creates a Parser over already-created Scanner.
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
