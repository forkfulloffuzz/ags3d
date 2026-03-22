package parser

import (
	"fmt"

	"github.com/ags3d/ag/internal/scanner"
)

// -------------------------------------------------------------------
// Parse errors
// -------------------------------------------------------------------

// ParseError is a single error with source location.
// All errors are collected rather than aborting — the caller receives
// a partial AST plus the full error list.
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
// Parser
// -------------------------------------------------------------------

// Parser is a single-pass recursive descent parser for AGS-spirit source.
//
// Construct with New(), then call Parse() exactly once.
// Errors are collected in p.errors — Parse() never panics on bad input.
// The returned *File may be partial when errors are present.
//
// Two-token lookahead (cur + next) is used to disambiguate type annotations
// from expression-starts at the top level and inside statement blocks.
type Parser struct {
	file   string
	s      *scanner.Scanner
	cur    scanner.Token // lookahead[0]
	next   scanner.Token // lookahead[1] — always valid
	errors []*ParseError
}

// New creates a Parser over an already-constructed Scanner.
// Both lookahead slots are primed so the first call to parseTopDecl
// can perform two-token disambiguation without extra buffering.
func New(s *scanner.Scanner) *Parser {
	p := &Parser{s: s}
	p.cur = p.s.Next()
	p.next = p.s.Next()
	return p
}

// Parse parses a complete .agscript file and returns the AST root.
// Errors are collected and returned alongside the (possibly partial) File.
// Never panics — unrecognised input is recorded as a ParseError.
func (p *Parser) Parse(file string) (*File, []*ParseError) {
	p.file = file
	f := &File{Path: file}
	for p.cur.Kind != scanner.TokenEOF {
		if d := p.parseTopDecl(); d != nil {
			f.Decls = append(f.Decls, d)
		}
	}
	return f, p.errors
}

// -------------------------------------------------------------------
// Top-level declarations
// -------------------------------------------------------------------

func (p *Parser) parseTopDecl() Decl {
	switch p.cur.Kind {
	case scanner.TokenNamespace:
		return p.parseNamespaceDecl()
	case scanner.TokenEnum:
		return p.parseEnumDecl()
	case scanner.TokenFunction:
		return p.parseFunctionDecl("", false)
	case scanner.TokenExport:
		p.errorf("'export' is only valid inside a namespace block")
		p.advance()
		p.sync()
		return nil
	default:
		typePos, typeName, ok := p.tryConsumeType(true)
		if !ok {
			p.errorf("expected declaration, got %q", p.cur.Lexeme)
			p.advance()
			p.sync()
			return nil
		}
		if p.cur.Kind == scanner.TokenFunction {
			return p.parseFunctionDecl(typeName, false)
		}
		// Top-level variable declaration: Type Ident [ "=" Expr ] ";"
		d := p.parseVarDeclTail(typeName, typePos)
		p.eat(scanner.TokenSemicolon, "top-level variable declaration")
		return &TopVarDecl{Decl: d}
	}
}

func (p *Parser) parseNamespaceDecl() *NamespaceDecl {
	pos := p.cur
	p.advance() // consume "namespace"
	nameTok := p.eat(scanner.TokenIdent, "namespace name")
	p.eat(scanner.TokenLBrace, "namespace body")

	ns := &NamespaceDecl{Name: nameTok.Lexeme, Pos: pos}
	for p.cur.Kind != scanner.TokenRBrace && p.cur.Kind != scanner.TokenEOF {
		isExport := false
		if p.cur.Kind == scanner.TokenExport {
			isExport = true
			p.advance()
		}
		switch p.cur.Kind {
		case scanner.TokenEnum:
			if isExport {
				p.errorf("'export' is not valid on enum declarations")
			}
			ns.Members = append(ns.Members, p.parseEnumDecl())
		case scanner.TokenFunction:
			ns.Members = append(ns.Members, p.parseFunctionDecl("", isExport))
		default:
			typePos, typeName, ok := p.tryConsumeType(true)
			if !ok {
				p.errorf("expected function, enum, or variable declaration in namespace, got %q", p.cur.Lexeme)
				p.advance()
				p.sync()
				continue
			}
			if p.cur.Kind == scanner.TokenFunction {
				ns.Members = append(ns.Members, p.parseFunctionDecl(typeName, isExport))
			} else {
				if isExport {
					p.errorf("'export' is not valid on variable declarations")
				}
				d := p.parseVarDeclTail(typeName, typePos)
				p.eat(scanner.TokenSemicolon, "namespace variable declaration")
				ns.Members = append(ns.Members, &TopVarDecl{Decl: d})
			}
		}
	}
	p.eat(scanner.TokenRBrace, "end of namespace")
	return ns
}

func (p *Parser) parseFunctionDecl(returnType string, isExport bool) *FunctionDecl {
	pos := p.cur
	p.advance() // consume "function"
	nameTok := p.eat(scanner.TokenIdent, "function name")
	p.eat(scanner.TokenLParen, "function parameter list")
	var params []Param
	if p.cur.Kind != scanner.TokenRParen {
		params = p.parseParamList()
	}
	p.eat(scanner.TokenRParen, "end of parameter list")
	// Post-param return type: "function foo() int {}" — alternative to "int function foo() {}"
	if returnType == "" && p.cur.Kind != scanner.TokenLBrace {
		if isTypeTok(p.cur.Kind) {
			returnType = p.cur.Lexeme
			p.advance()
		} else if p.cur.Kind == scanner.TokenIdent && p.next.Kind == scanner.TokenLBrace {
			returnType = p.cur.Lexeme
			p.advance()
		}
	}
	body := p.parseBlock()
	return &FunctionDecl{
		ReturnType: returnType,
		Name:       nameTok.Lexeme,
		Params:     params,
		Body:       body,
		IsExport:   isExport,
		Pos:        pos,
	}
}

func (p *Parser) parseParamList() []Param {
	var params []Param
	for {
		typePos, typeName, ok := p.tryConsumeType(false)
		if !ok {
			p.errorf("expected parameter type, got %q", p.cur.Lexeme)
			break
		}
		nameTok := p.eat(scanner.TokenIdent, "parameter name")
		params = append(params, Param{
			Type: typeName,
			Name: nameTok.Lexeme,
			Pos:  typePos,
		})
		if p.cur.Kind != scanner.TokenComma {
			break
		}
		p.advance() // consume ","
	}
	return params
}

func (p *Parser) parseEnumDecl() *EnumDecl {
	pos := p.cur
	p.advance() // consume "enum"
	nameTok := p.eat(scanner.TokenIdent, "enum name")
	p.eat(scanner.TokenLBrace, "enum body")

	var members []EnumMember
	for p.cur.Kind != scanner.TokenRBrace && p.cur.Kind != scanner.TokenEOF {
		memPos := p.cur
		memTok := p.eat(scanner.TokenIdent, "enum member name")
		var val Expr
		if p.cur.Kind == scanner.TokenAssign {
			p.advance() // consume "="
			val = p.parseExpr()
		}
		members = append(members, EnumMember{
			Name:  memTok.Lexeme,
			Value: val,
			Pos:   memPos,
		})
		if p.cur.Kind == scanner.TokenComma {
			p.advance() // trailing comma is allowed
		} else {
			break
		}
	}
	p.eat(scanner.TokenRBrace, "end of enum")
	// Semicolon after enum closing brace is optional.
	if p.cur.Kind == scanner.TokenSemicolon {
		p.advance()
	}
	return &EnumDecl{Name: nameTok.Lexeme, Members: members, Pos: pos}
}

// -------------------------------------------------------------------
// Statements
// -------------------------------------------------------------------

// parseBlockOrSingleStmt parses either a braced block or a single statement.
// Fixtures use braceless C-style bodies: "if (cond) stmt;" without "{}".
func (p *Parser) parseBlockOrSingleStmt() *Block {
	if p.cur.Kind == scanner.TokenLBrace {
		return p.parseBlock()
	}
	pos := p.cur
	s := p.parseStmt()
	b := &Block{Pos: pos}
	if s != nil {
		b.Stmts = []Stmt{s}
	}
	return b
}

func (p *Parser) parseBlock() *Block {
	pos := p.cur
	p.eat(scanner.TokenLBrace, "block opening brace")
	b := &Block{Pos: pos}
	for p.cur.Kind != scanner.TokenRBrace && p.cur.Kind != scanner.TokenEOF {
		prev := p.cur
		s := p.parseStmt()
		if s != nil {
			b.Stmts = append(b.Stmts, s)
		}
		// Safety: if parseStmt made no progress, skip one token to avoid an infinite loop.
		if p.cur.Kind == prev.Kind && p.cur.Line == prev.Line && p.cur.Column == prev.Column {
			p.advance()
		}
	}
	p.eat(scanner.TokenRBrace, "block closing brace")
	return b
}

func (p *Parser) parseStmt() Stmt {
	switch p.cur.Kind {
	case scanner.TokenIf:
		return p.parseIfStmt()
	case scanner.TokenWhile:
		return p.parseWhileStmt()
	case scanner.TokenDo:
		return p.parseDoWhileStmt()
	case scanner.TokenFor:
		return p.parseForStmt()
	case scanner.TokenSwitch:
		return p.parseSwitchStmt()
	case scanner.TokenLBrace:
		return p.parseBlock()
	case scanner.TokenReturn:
		s := p.parseReturnStmt()
		p.eat(scanner.TokenSemicolon, "return statement")
		return s
	case scanner.TokenBreak:
		pos := p.cur
		p.advance()
		p.eat(scanner.TokenSemicolon, "break statement")
		return &BreakStmt{Pos: pos}
	case scanner.TokenContinue:
		pos := p.cur
		p.advance()
		p.eat(scanner.TokenSemicolon, "continue statement")
		return &ContinueStmt{Pos: pos}
	default:
		if p.startsVarDeclInStmt() {
			typePos, typeName, _ := p.tryConsumeType(false)
			d := p.parseVarDeclTail(typeName, typePos)
			p.eat(scanner.TokenSemicolon, "variable declaration")
			return d
		}
		pos := p.cur
		x := p.parseExpr()
		// Detect qualified-type variable declarations: "Namespace.Type varName".
		// parseExpr() consumes "Namespace.Type" as a MemberExpr; if the next
		// token is an Ident the expression must really be a type annotation.
		if p.cur.Kind == scanner.TokenIdent {
			if qn := memberExprTypeName(x); qn != "" {
				d := p.parseVarDeclTail(qn, pos)
				p.eat(scanner.TokenSemicolon, "variable declaration")
				return d
			}
		}
		s := &ExprStmt{X: x, Pos: pos}
		p.eat(scanner.TokenSemicolon, "expression statement")
		return s
	}
}

// startsVarDeclInStmt returns true when the current position looks like
// the beginning of a typed variable declaration inside a block.
// Built-in type keywords are unambiguous; a user-defined type (Ident) is
// only a type if followed by another Ident (the variable name).
func (p *Parser) startsVarDeclInStmt() bool {
	if isTypeTok(p.cur.Kind) {
		return true
	}
	return p.cur.Kind == scanner.TokenIdent && p.next.Kind == scanner.TokenIdent
}

func (p *Parser) parseIfStmt() *IfStmt {
	pos := p.cur
	p.advance() // consume "if"
	p.eat(scanner.TokenLParen, "if condition")
	cond := p.parseExpr()
	p.eat(scanner.TokenRParen, "if condition")
	then := p.parseBlockOrSingleStmt()
	var els Stmt
	if p.cur.Kind == scanner.TokenElse {
		p.advance() // consume "else"
		if p.cur.Kind == scanner.TokenIf {
			els = p.parseIfStmt()
		} else {
			els = p.parseBlockOrSingleStmt()
		}
	}
	return &IfStmt{Cond: cond, Then: then, Else: els, Pos: pos}
}

func (p *Parser) parseWhileStmt() *WhileStmt {
	pos := p.cur
	p.advance() // consume "while"
	p.eat(scanner.TokenLParen, "while condition")
	cond := p.parseExpr()
	p.eat(scanner.TokenRParen, "while condition")
	body := p.parseBlockOrSingleStmt()
	return &WhileStmt{Cond: cond, Body: body, Pos: pos}
}

func (p *Parser) parseDoWhileStmt() *DoWhileStmt {
	pos := p.cur
	p.advance() // consume "do"
	body := p.parseBlock()
	p.eat(scanner.TokenWhile, "do-while condition")
	p.eat(scanner.TokenLParen, "do-while condition")
	cond := p.parseExpr()
	p.eat(scanner.TokenRParen, "do-while condition")
	p.eat(scanner.TokenSemicolon, "do-while statement")
	return &DoWhileStmt{Body: body, Cond: cond, Pos: pos}
}

func (p *Parser) parseForStmt() *ForStmt {
	pos := p.cur
	p.advance() // consume "for"
	p.eat(scanner.TokenLParen, "for statement")

	// ForInit: VarDecl | Expr | ε
	var init Stmt
	if p.cur.Kind != scanner.TokenSemicolon {
		if p.startsVarDeclInStmt() {
			typePos, typeName, _ := p.tryConsumeType(false)
			init = p.parseVarDeclTail(typeName, typePos)
		} else {
			initPos := p.cur
			init = &ExprStmt{X: p.parseExpr(), Pos: initPos}
		}
	}
	p.eat(scanner.TokenSemicolon, "for init separator")

	var cond Expr
	if p.cur.Kind != scanner.TokenSemicolon {
		cond = p.parseExpr()
	}
	p.eat(scanner.TokenSemicolon, "for condition separator")

	var post Expr
	if p.cur.Kind != scanner.TokenRParen {
		post = p.parseExpr()
	}
	p.eat(scanner.TokenRParen, "for post")
	body := p.parseBlockOrSingleStmt()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body, Pos: pos}
}

func (p *Parser) parseSwitchStmt() *SwitchStmt {
	pos := p.cur
	p.advance() // consume "switch"
	p.eat(scanner.TokenLParen, "switch tag")
	tag := p.parseExpr()
	p.eat(scanner.TokenRParen, "switch tag")
	p.eat(scanner.TokenLBrace, "switch body")
	var cases []*CaseClause
	for p.cur.Kind != scanner.TokenRBrace && p.cur.Kind != scanner.TokenEOF {
		cases = append(cases, p.parseCaseClause())
	}
	p.eat(scanner.TokenRBrace, "end of switch")
	return &SwitchStmt{Tag: tag, Cases: cases, Pos: pos}
}

func (p *Parser) parseCaseClause() *CaseClause {
	pos := p.cur
	var val Expr
	switch p.cur.Kind {
	case scanner.TokenCase:
		p.advance() // consume "case"
		val = p.parseExpr()
		p.eat(scanner.TokenColon, "case label")
	case scanner.TokenDefault:
		p.advance() // consume "default"
		p.eat(scanner.TokenColon, "default label")
	default:
		p.errorf("expected 'case' or 'default', got %q", p.cur.Lexeme)
		p.sync()
		return &CaseClause{Pos: pos}
	}
	var stmts []Stmt
	for p.cur.Kind != scanner.TokenCase &&
		p.cur.Kind != scanner.TokenDefault &&
		p.cur.Kind != scanner.TokenRBrace &&
		p.cur.Kind != scanner.TokenEOF {
		if s := p.parseStmt(); s != nil {
			stmts = append(stmts, s)
		}
	}
	return &CaseClause{Value: val, Body: stmts, Pos: pos}
}

func (p *Parser) parseReturnStmt() *ReturnStmt {
	pos := p.cur
	p.advance() // consume "return"
	var val Expr
	if p.cur.Kind != scanner.TokenSemicolon {
		val = p.parseExpr()
	}
	return &ReturnStmt{Value: val, Pos: pos}
}

// -------------------------------------------------------------------
// Expressions — operator-precedence grammar (lowest → highest)
// -------------------------------------------------------------------

func (p *Parser) parseExpr() Expr {
	return p.parseAssignExpr()
}

// parseAssignExpr handles right-associative assignment:
//
//	LogicOrExpr [ AssignOp AssignExpr ]
func (p *Parser) parseAssignExpr() Expr {
	left := p.parseLogicOrExpr()
	op, ok := p.matchAssignOp()
	if !ok {
		return left
	}
	right := p.parseAssignExpr() // right-associative
	return &AssignExpr{Op: op.Lexeme, Target: left, Value: right, Pos: op}
}

func (p *Parser) matchAssignOp() (scanner.Token, bool) {
	return p.match(
		scanner.TokenAssign,
		scanner.TokenPlusAssign, scanner.TokenMinusAssign,
		scanner.TokenStarAssign, scanner.TokenSlashAssign,
		scanner.TokenPercentAssign,
		scanner.TokenAndAssign, scanner.TokenOrAssign,
		scanner.TokenXorAssign,
		scanner.TokenLShiftAssign, scanner.TokenRShiftAssign,
	)
}

func (p *Parser) parseLogicOrExpr() Expr {
	left := p.parseLogicAndExpr()
	for p.cur.Kind == scanner.TokenOr {
		op := p.cur
		p.advance()
		right := p.parseLogicAndExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseLogicAndExpr() Expr {
	left := p.parseBitOrExpr()
	for p.cur.Kind == scanner.TokenAnd {
		op := p.cur
		p.advance()
		right := p.parseBitOrExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseBitOrExpr() Expr {
	left := p.parseBitXorExpr()
	for p.cur.Kind == scanner.TokenPipe {
		op := p.cur
		p.advance()
		right := p.parseBitXorExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseBitXorExpr() Expr {
	left := p.parseBitAndExpr()
	for p.cur.Kind == scanner.TokenCaret {
		op := p.cur
		p.advance()
		right := p.parseBitAndExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseBitAndExpr() Expr {
	left := p.parseEqualExpr()
	for p.cur.Kind == scanner.TokenAmpersand {
		op := p.cur
		p.advance()
		right := p.parseEqualExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseEqualExpr() Expr {
	left := p.parseRelExpr()
	for p.cur.Kind == scanner.TokenEq || p.cur.Kind == scanner.TokenNeq {
		op := p.cur
		p.advance()
		right := p.parseRelExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseRelExpr() Expr {
	left := p.parseShiftExpr()
	for p.cur.Kind == scanner.TokenLt || p.cur.Kind == scanner.TokenLte ||
		p.cur.Kind == scanner.TokenGt || p.cur.Kind == scanner.TokenGte {
		op := p.cur
		p.advance()
		right := p.parseShiftExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseShiftExpr() Expr {
	left := p.parseAddExpr()
	for p.cur.Kind == scanner.TokenLShift || p.cur.Kind == scanner.TokenRShift {
		op := p.cur
		p.advance()
		right := p.parseAddExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseAddExpr() Expr {
	left := p.parseMulExpr()
	for p.cur.Kind == scanner.TokenPlus || p.cur.Kind == scanner.TokenMinus {
		op := p.cur
		p.advance()
		right := p.parseMulExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseMulExpr() Expr {
	left := p.parseUnaryExpr()
	for p.cur.Kind == scanner.TokenStar || p.cur.Kind == scanner.TokenSlash ||
		p.cur.Kind == scanner.TokenPercent {
		op := p.cur
		p.advance()
		right := p.parseUnaryExpr()
		left = &BinaryExpr{Op: op.Lexeme, Left: left, Right: right, Pos: op}
	}
	return left
}

func (p *Parser) parseUnaryExpr() Expr {
	switch p.cur.Kind {
	case scanner.TokenBang, scanner.TokenMinus, scanner.TokenTilde,
		scanner.TokenPlusPlus, scanner.TokenMinusMinus:
		op := p.cur
		p.advance()
		x := p.parseUnaryExpr()
		return &UnaryExpr{Op: op.Lexeme, X: x, Pos: op}
	}
	return p.parsePostfixExpr()
}

// parsePostfixExpr applies zero or more postfix suffixes to a primary:
//
//	PrimaryExpr { "++" | "--" | "." Ident | "(" [ArgList] ")" | "[" Expr "]" }
func (p *Parser) parsePostfixExpr() Expr {
	x := p.parsePrimaryExpr()
	for {
		switch p.cur.Kind {
		case scanner.TokenPlusPlus, scanner.TokenMinusMinus:
			op := p.cur
			p.advance()
			x = &PostfixExpr{Op: op.Lexeme, X: x, Pos: op}
		case scanner.TokenDot:
			dotPos := p.cur
			p.advance()
			fieldTok := p.eat(scanner.TokenIdent, "field name")
			x = &MemberExpr{Object: x, Field: fieldTok.Lexeme, Pos: dotPos}
		case scanner.TokenLParen:
			callPos := p.cur
			p.advance()
			var args []Expr
			if p.cur.Kind != scanner.TokenRParen {
				args = p.parseArgList()
			}
			p.eat(scanner.TokenRParen, "end of argument list")
			x = &CallExpr{Callee: x, Args: args, Pos: callPos}
		case scanner.TokenLBracket:
			idxPos := p.cur
			p.advance()
			idx := p.parseExpr()
			p.eat(scanner.TokenRBracket, "end of index expression")
			x = &IndexExpr{Object: x, Index: idx, Pos: idxPos}
		default:
			return x
		}
	}
}

// parsePrimaryExpr handles atomic expressions:
//
//	Ident | Literal | "global" "." Ident | "(" Expr ")"
func (p *Parser) parsePrimaryExpr() Expr {
	switch p.cur.Kind {
	case scanner.TokenGlobal:
		// "global" "." Ident — GlobalExpr; postfix loop handles further suffixes.
		pos := p.cur
		p.advance()
		p.eat(scanner.TokenDot, "global property access (expected '.')")
		propTok := p.eat(scanner.TokenIdent, "global property name")
		return &GlobalExpr{Property: propTok.Lexeme, Pos: pos}

	case scanner.TokenIdent:
		tok := p.cur
		p.advance()
		return &Identifier{Name: tok.Lexeme, Pos: tok}

	case scanner.TokenIntLit:
		tok := p.cur
		p.advance()
		return &Literal{Kind: "int", Value: tok.Lexeme, Pos: tok}

	case scanner.TokenFloatLit:
		tok := p.cur
		p.advance()
		return &Literal{Kind: "float", Value: tok.Lexeme, Pos: tok}

	case scanner.TokenStringLit:
		tok := p.cur
		p.advance()
		return &Literal{Kind: "string", Value: tok.Lexeme, Pos: tok}

	case scanner.TokenTrue:
		tok := p.cur
		p.advance()
		return &Literal{Kind: "bool", Value: "true", Pos: tok}

	case scanner.TokenFalse:
		tok := p.cur
		p.advance()
		return &Literal{Kind: "bool", Value: "false", Pos: tok}

	case scanner.TokenNull:
		tok := p.cur
		p.advance()
		return &Literal{Kind: "null", Value: "null", Pos: tok}

	case scanner.TokenLParen:
		p.advance()
		x := p.parseExpr()
		p.eat(scanner.TokenRParen, "closing parenthesis")
		return x

	default:
		p.errorf("expected expression, got %q", p.cur.Lexeme)
		tok := p.cur
		p.advance() // consume bad token so parsing can continue
		return &Literal{Kind: "int", Value: "0", Pos: tok}
	}
}

func (p *Parser) parseArgList() []Expr {
	args := []Expr{p.parseExpr()}
	for p.cur.Kind == scanner.TokenComma {
		p.advance()
		args = append(args, p.parseExpr())
	}
	return args
}

// -------------------------------------------------------------------
// Shared helpers
// -------------------------------------------------------------------

// parseVarDeclTail parses the name and optional initialiser of a variable
// declaration after the type has already been consumed.
//
//	Ident [ "=" Expr ]
func (p *Parser) parseVarDeclTail(typeName string, typePos scanner.Token) *VarDecl {
	nameTok := p.eat(scanner.TokenIdent, "variable name")
	var init Expr
	if p.cur.Kind == scanner.TokenAssign {
		p.advance() // consume "="
		init = p.parseExpr()
	}
	return &VarDecl{Type: typeName, Name: nameTok.Lexeme, Init: init, Pos: typePos}
}

// tryConsumeType attempts to consume a type annotation at the current position.
// Returns (typePos, typeName, true) on success.
//
// Built-in type keywords (int, float, bool, string, void, …) are always types.
// A bare Ident is treated as a user-defined type when:
//   - forTopLevel is true and next is "function" (return-type annotation), OR
//   - next is an Ident (variable or parameter name).
func (p *Parser) tryConsumeType(forTopLevel bool) (scanner.Token, string, bool) {
	if isTypeTok(p.cur.Kind) {
		pos, name := p.cur, p.cur.Lexeme
		p.advance()
		return pos, name, true
	}
	if p.cur.Kind == scanner.TokenIdent {
		next := p.next.Kind
		if next == scanner.TokenIdent || (forTopLevel && next == scanner.TokenFunction) {
			pos, name := p.cur, p.cur.Lexeme
			p.advance()
			return pos, name, true
		}
	}
	return scanner.Token{}, "", false
}

// memberExprTypeName extracts a dotted type name from a MemberExpr chain,
// e.g. MemberExpr{Ident{"InventoryUtils"}, "PickupResult"} → "InventoryUtils.PickupResult".
// Returns "" if x is not a pure Ident/MemberExpr chain (e.g. has a CallExpr inside).
func memberExprTypeName(x Expr) string {
	switch e := x.(type) {
	case *Identifier:
		return e.Name
	case *MemberExpr:
		base := memberExprTypeName(e.Object)
		if base == "" {
			return ""
		}
		return base + "." + e.Field
	default:
		return ""
	}
}

// isTypeTok reports whether k is a built-in type keyword.
func isTypeTok(k scanner.TokenKind) bool {
	switch k {
	case scanner.TokenBool, scanner.TokenChar, scanner.TokenFloat,
		scanner.TokenInt, scanner.TokenShort, scanner.TokenString, scanner.TokenVoid:
		return true
	}
	return false
}

// -------------------------------------------------------------------
// Token management
// -------------------------------------------------------------------

// advance shifts the two-token lookahead window forward by one position.
func (p *Parser) advance() {
	p.cur = p.next
	p.next = p.s.Next()
}

// peek returns the current lookahead token without consuming it.
func (p *Parser) peek() scanner.Token {
	return p.cur
}

// check returns true if the current token has the given kind.
func (p *Parser) check(k scanner.TokenKind) bool {
	return p.cur.Kind == k
}

// eat consumes the current token and returns it if it matches kind k.
// On mismatch, records an error and returns the current token without consuming.
func (p *Parser) eat(k scanner.TokenKind, context string) scanner.Token {
	if p.cur.Kind == k {
		tok := p.cur
		p.advance()
		return tok
	}
	p.errorf("expected %v but found %q (%s)", k, p.cur.Lexeme, context)
	return p.cur
}

// match consumes and returns the current token if it matches any given kind.
func (p *Parser) match(kinds ...scanner.TokenKind) (scanner.Token, bool) {
	for _, k := range kinds {
		if p.cur.Kind == k {
			tok := p.cur
			p.advance()
			return tok, true
		}
	}
	return scanner.Token{}, false
}

// errorf records a ParseError at the current token position.
func (p *Parser) errorf(format string, args ...any) {
	p.errors = append(p.errors, &ParseError{
		File:    p.file,
		Line:    p.cur.Line,
		Column:  p.cur.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

// errorAt records a ParseError at a specific token's position.
func (p *Parser) errorAt(tok scanner.Token, format string, args ...any) {
	p.errors = append(p.errors, &ParseError{
		File:    p.file,
		Line:    tok.Line,
		Column:  tok.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

// sync advances past tokens until it finds one that can start a new
// top-level declaration or a block boundary — panic-mode error recovery.
func (p *Parser) sync() {
	for {
		switch p.cur.Kind {
		case scanner.TokenEOF,
			scanner.TokenFunction,
			scanner.TokenNamespace,
			scanner.TokenEnum,
			scanner.TokenRBrace:
			return
		}
		p.advance()
	}
}
