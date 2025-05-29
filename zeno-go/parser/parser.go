package parser

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/linkalls/zeno-lang/ast"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/token"
)

// Precedence levels for operator precedence parsing
const (
	_          int = iota
	LOWEST         // 最低の優先度
	EQUALS         // ==, != 演算子の優先度
	COMPARISON     // <, >, <=, >= 演算子の優先度
	SUM            // +, - 演算子の優先度
	PRODUCT        // *, / 演算子の優先度
	PREFIX         // -X or !X
	CALL           // myFunction(X)
)

// precedences maps tokens to their precedence
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       COMPARISON,
	token.LTE:      COMPARISON,
	token.GT:       COMPARISON,
	token.GTE:      COMPARISON,
	token.AND:      EQUALS, // 論理AND
	token.OR:       EQUALS, // 論理OR
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.DIVIDE:   PRODUCT,
	token.MULTIPLY: PRODUCT,
	token.LPAREN:   CALL,
}

// Parser holds the state for parsing tokens into an AST
type Parser struct {
	l *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	currentUntil token.TokenType // 追加: 現在のuntilトークンを保持
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// tokenToBinaryOperator converts a token literal to BinaryOperator
func tokenToBinaryOperator(literal string) ast.BinaryOperator {
	switch literal {
	case "+":
		return ast.BinaryOpPlus
	case "-":
		return ast.BinaryOpMinus
	case "*":
		return ast.BinaryOpMultiply
	case "/":
		return ast.BinaryOpDivide
	case "%":
		return ast.BinaryOpModulo
	case "==":
		return ast.BinaryOpEq
	case "!=":
		return ast.BinaryOpNotEq
	case "<":
		return ast.BinaryOpLt
	case "<=":
		return ast.BinaryOpLte
	case ">":
		return ast.BinaryOpGt
	case ">=":
		return ast.BinaryOpGte
	case "&&":
		return ast.BinaryOpAnd
	case "||":
		return ast.BinaryOpOr
	default:
		return ast.BinaryOpPlus // デフォルト値
	}
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.currentToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// New creates a new parser instance
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:            l,
		errors:       []string{},
		currentUntil: token.SEMICOLON, // 追加: デフォルトのuntilトークン
	}

	// Initialize prefix parse functions
	p.prefixParseFns = map[token.TokenType]prefixParseFn{
		token.IDENT:  p.parseIdentifier,
		token.INT:    p.parseIntegerLiteral,
		token.STRING: p.parseStringLiteral,
		token.TRUE:   p.parseBooleanLiteral,
		token.FALSE:  p.parseBooleanLiteral,
	}

	// Initialize infix parse functions
	p.infixParseFns = map[token.TokenType]infixParseFn{
		token.PLUS:     p.parseInfixExpression,
		token.MINUS:    p.parseInfixExpression,
		token.MULTIPLY: p.parseInfixExpression,
		token.DIVIDE:   p.parseInfixExpression,
		token.EQ:       p.parseInfixExpression,
		token.NOT_EQ:   p.parseInfixExpression,
		token.LT:       p.parseInfixExpression,
		token.LTE:      p.parseInfixExpression,
		token.GT:       p.parseInfixExpression,
		token.GTE:      p.parseInfixExpression,
		token.AND:      p.parseInfixExpression,
		token.OR:       p.parseInfixExpression,
		token.LPAREN:   p.parseFunctionCall,
	}

	// Read two tokens, so currentToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currentToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.IMPORT:
		return p.parseImportStatement()
	case token.LET:
		return p.parseLetStatement()
	case token.PRINT:
		return p.parsePrintStatement(false)
	case token.PRINTLN:
		return p.parsePrintStatement(true)
	case token.IF:
		return p.parseIfStatement()
	case token.PUB:
		return p.parsePublicDeclaration()
	case token.FN:
		return p.parseFunctionDefinition()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.WHILE: // 追加: WHILEトークンに対する処理
		return p.parseWhileStatement()
	case token.IDENT:
		// Check if this is an assignment statement
		if p.peekToken.Type == token.ASSIGN {
			return p.parseAssignmentStatement()
		}
		// Otherwise, it's an expression statement
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetDeclaration {
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	name := p.currentToken.Literal

	// Check for optional type annotation
	var typeAnn *string
	if p.peekToken.Type == token.COLON {
		p.nextToken() // consume ':'
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		annotation := p.currentToken.Literal
		typeAnn = &annotation
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)

	return &ast.LetDeclaration{
		Name:            name,
		TypeAnn:         typeAnn,
		ValueExpression: value,
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{}
	stmt.Expression = p.parseExpression(LOWEST)

	return stmt
}

// parseAssignmentStatement parses assignment statements (x = value)
func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	// Current token is the identifier (variable name)
	name := p.currentToken.Literal

	// Expect the ASSIGN token
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// Move to the value expression
	p.nextToken()
	value := p.parseExpression(LOWEST)

	return &ast.AssignmentStatement{
		Name:  name,
		Value: value,
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	return p.parseExpressionUntil(precedence, token.SEMICOLON)
}

func (p *Parser) parseExpressionUntil(precedence int, until token.TokenType) ast.Expression {
	// 保存してuntilコンテキストを設定
	prev := p.currentUntil
	p.currentUntil = until
	defer func() { p.currentUntil = prev }()

	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}

	left := prefix()

	for p.peekToken.Type != until && p.peekToken.Type != token.EOF && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return left
		}
		p.nextToken()
		left = infix(left)
	}
	return left
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{}

	value, err := strconv.Atoi(p.currentToken.Literal)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	// Process escape sequences in the string literal
	processedValue := lexer.ProcessStringLiteral(p.currentToken.Literal)
	return &ast.StringLiteral{Value: processedValue}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Value: p.currentToken.Type == token.TRUE}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.BinaryExpression{
		Left:     left,
		Operator: tokenToBinaryOperator(p.currentToken.Literal),
	}

	prec := p.curPrecedence()
	p.nextToken()
	// 右辺をcurrentUntilコンテキストでパース
	expr.Right = p.parseExpressionUntil(prec, p.currentUntil)

	return expr
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// ParseExpression parses a single expression (useful for testing)
func ParseExpression(input string) (ast.Expression, error) {
	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)
	if len(p.errors) > 0 {
		return nil, errors.New(p.errors[0])
	}
	return expr, nil
}

// parseImportStatement parses import {identifier, ...} from "module" syntax
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	// Current token is IMPORT
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	var imports []string

	// Parse first identifier (accept both IDENT and keywords as identifiers in import context)
	p.nextToken()
	if !p.isValidImportIdentifier() {
		return nil
	}
	imports = append(imports, p.currentToken.Literal)

	// Parse additional identifiers separated by commas
	for p.peekToken.Type == token.COMMA {
		p.nextToken() // consume comma
		p.nextToken() // move to next token
		if !p.isValidImportIdentifier() {
			return nil
		}
		imports = append(imports, p.currentToken.Literal)
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	if !p.expectPeek(token.FROM) {
		return nil
	}

	if !p.expectPeek(token.STRING) {
		return nil
	}

	// Remove quotes from string literal
	module := p.currentToken.Literal
	if len(module) >= 2 && module[0] == '"' && module[len(module)-1] == '"' {
		module = module[1 : len(module)-1]
	}

	return &ast.ImportStatement{
		Imports: imports,
		Module:  module,
	}
}

// parsePrintStatement parses print(...) and println(...) statements
func (p *Parser) parsePrintStatement(newline bool) *ast.PrintStatement {
	// Current token is PRINT or PRINTLN
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	var arguments []ast.Expression

	// Handle empty argument list
	if p.peekToken.Type == token.RPAREN {
		p.nextToken() // consume ')'
	} else {
		// Parse first argument
		p.nextToken()
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			arguments = append(arguments, arg)
		}

		// Parse additional arguments separated by commas
		for p.peekToken.Type == token.COMMA {
			p.nextToken() // consume comma
			p.nextToken() // move to next expression
			arg := p.parseExpression(LOWEST)
			if arg != nil {
				arguments = append(arguments, arg)
			}
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	return &ast.PrintStatement{
		Arguments: arguments,
		Newline:   newline,
	}
}

// isValidImportIdentifier checks if the current token can be used as an identifier in import statements
func (p *Parser) isValidImportIdentifier() bool {
	return p.currentToken.Type == token.IDENT ||
		p.currentToken.Type == token.PRINTLN ||
		p.currentToken.Type == token.PRINT
}

// parseFunctionDefinition parses function definitions
func (p *Parser) parseFunctionDefinition() *ast.FunctionDefinition {
	return p.parseFunctionDefinitionWithVisibility(false)
}

// parsePublicDeclaration parses public declarations (pub fn)
func (p *Parser) parsePublicDeclaration() ast.Statement {
	// Current token is PUB
	if p.peekToken.Type != token.FN {
		p.errors = append(p.errors, "pub can only be used with function definitions")
		return nil
	}

	p.nextToken() // consume PUB, move to FN
	return p.parseFunctionDefinitionWithVisibility(true)
}

// parseFunctionDefinitionWithVisibility parses function definitions with optional visibility
func (p *Parser) parseFunctionDefinitionWithVisibility(isPublic bool) *ast.FunctionDefinition {
	// Current token is FN
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	name := p.currentToken.Literal

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Parse parameters
	var parameters []ast.Parameter
	if p.peekToken.Type != token.RPAREN {
		p.nextToken() // move to first parameter name

		for {
			if p.currentToken.Type != token.IDENT {
				p.errors = append(p.errors, "expected parameter name")
				return nil
			}

			paramName := p.currentToken.Literal

			if !p.expectPeek(token.COLON) {
				return nil
			}

			if !p.expectPeek(token.IDENT) {
				return nil
			}

			paramType := p.currentToken.Literal

			parameters = append(parameters, ast.Parameter{
				Name: paramName,
				Type: paramType,
			})

			if p.peekToken.Type == token.COMMA {
				p.nextToken() // consume comma
				p.nextToken() // move to next parameter
			} else {
				break
			}
		}
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Parse optional return type
	var returnType *string
	if p.peekToken.Type == token.COLON {
		p.nextToken() // consume ':'
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		retType := p.currentToken.Literal
		returnType = &retType
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse function body
	var body []ast.Statement
	for p.peekToken.Type != token.RBRACE && p.peekToken.Type != token.EOF {
		p.nextToken()
		stmt := p.parseStatement()
		if stmt != nil {
			body = append(body, stmt)
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return &ast.FunctionDefinition{
		Name:       name,
		Parameters: parameters,
		ReturnType: returnType,
		Body:       body,
		IsPublic:   isPublic,
	}
}

// parseReturnStatement parses return statements
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// Current token is RETURN

	var value ast.Expression
	if p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.EOF {
		p.nextToken()
		value = p.parseExpression(LOWEST)
	}

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return &ast.ReturnStatement{
		Value: value,
	}
}

// parseFunctionCall parses function calls
func (p *Parser) parseFunctionCall(fn ast.Expression) ast.Expression {
	call := &ast.FunctionCall{
		Name: fn.(*ast.Identifier).Value,
	}
	call.Arguments = p.parseCallArguments()
	return call
}

// parseCallArguments parses function call arguments
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

// parseIfStatement parses if/else if/else statements
func (p *Parser) parseIfStatement() *ast.IfStatement {
	// Current token is IF
	p.nextToken() // move to condition

	condition := p.parseExpressionUntil(LOWEST, token.LBRACE)
	if condition == nil {
		return nil
	}

	// After parsing expression, we should be at the end of condition
	// expectPeek will move us to LBRACE
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	thenBlock := p.parseBlockStatement()
	if thenBlock == nil {
		return nil
	}

	var elseIfClauses []ast.ElseIfClause
	var elseBlock *ast.Block

	// Check for else if and else clauses
	for p.peekToken.Type == token.ELSE {
		p.nextToken() // consume current token (RBRACE from previous block)
		p.nextToken() // consume ELSE

		if p.currentToken.Type == token.IF {
			// This is an else if clause
			p.nextToken() // consume IF, move to condition

			elseIfCondition := p.parseExpressionUntil(LOWEST, token.LBRACE)
			if elseIfCondition == nil {
				return nil
			}

			if !p.expectPeek(token.LBRACE) {
				return nil
			}

			elseIfBlock := p.parseBlockStatement()
			if elseIfBlock == nil {
				return nil
			}

			elseIfClauses = append(elseIfClauses, ast.ElseIfClause{
				Condition: elseIfCondition,
				Block:     elseIfBlock,
			})
		} else if p.currentToken.Type == token.LBRACE {
			// This is an else clause
			elseBlock = p.parseBlockStatement()
			if elseBlock == nil {
				return nil
			}
			break // else block is the last one
		} else {
			// else must be followed by if or {
			p.errors = append(p.errors, fmt.Sprintf("expected 'if' or '{' after 'else', got %s", p.currentToken.Type))
			return nil
		}
	}

	return &ast.IfStatement{
		Condition:     condition,
		ThenBlock:     thenBlock,
		ElseIfClauses: elseIfClauses,
		ElseBlock:     elseBlock,
	}
}

// parseBlockStatement parses a block statement
func (p *Parser) parseBlockStatement() *ast.Block {
	// Current token is LBRACE
	block := &ast.Block{}

	p.nextToken() // move past LBRACE

	for p.currentToken.Type != token.RBRACE && p.currentToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	if p.currentToken.Type != token.RBRACE {
		p.errors = append(p.errors, "expected '}' to close block")
		return nil
	}

	return block
}

// parseWhileStatement parses while loops
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	// Current token is WHILE
	p.nextToken() // move to condition

	condition := p.parseExpressionUntil(LOWEST, token.LBRACE)
	if condition == nil {
		return nil
	}

	// After parsing expression, we should be at the end of condition
	// expectPeek will move us to LBRACE
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	block := p.parseBlockStatement()
	if block == nil {
		return nil
	}

	return &ast.WhileStatement{
		Condition: condition,
		Block:     block,
	}
}
