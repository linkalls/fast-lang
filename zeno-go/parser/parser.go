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
	LOWEST
	LOGICAL_OR     // ||
	LOGICAL_AND    // &&
	EQUALS         // ==, !=
	COMPARISON     // <, >, <=, >=
	SUM            // +, -
	PRODUCT        // *, /
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
	token.AND:      LOGICAL_AND,
	token.OR:       LOGICAL_OR,
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

	currentUntil token.TokenType
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func tokenToBinaryOperator(literal string) ast.BinaryOperator {
	switch literal {
	case "+": return ast.BinaryOpPlus
	case "-": return ast.BinaryOpMinus
	case "*": return ast.BinaryOpMultiply
	case "/": return ast.BinaryOpDivide
	case "%": return ast.BinaryOpModulo
	case "==": return ast.BinaryOpEq
	case "!=": return ast.BinaryOpNotEq
	case "<": return ast.BinaryOpLt
	case "<=": return ast.BinaryOpLte
	case ">": return ast.BinaryOpGt
	case ">=": return ast.BinaryOpGte
	case "&&": return ast.BinaryOpAnd
	case "||": return ast.BinaryOpOr
	default: return ast.BinaryOpPlus
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

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:            l,
		errors:       []string{},
		currentUntil: token.SEMICOLON,
	}
	p.prefixParseFns = map[token.TokenType]prefixParseFn{
		token.IDENT:    p.parseIdentifier,
		token.INT:      p.parseIntegerLiteral,
		token.STRING:   p.parseStringLiteral,
		token.TRUE:     p.parseBooleanLiteral,
		token.FALSE:    p.parseBooleanLiteral,
		token.BANG:     p.parsePrefixExpression,
		token.MINUS:    p.parsePrefixExpression,
		token.FLOAT:    p.parseFloatLiteral,
		token.LBRACKET: p.parseArrayLiteral, // Added for array literals
	}
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
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: []ast.Statement{}}
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
	var stmt ast.Statement
	switch p.currentToken.Type {
	case token.IMPORT:
		stmt = p.parseImportStatement()
	case token.LET:
		stmt = p.parseLetStatement()
	case token.IF:
		stmt = p.parseIfStatement()
	case token.PUB:
		stmt = p.parsePublicDeclaration()
	case token.FN:
		stmt = p.parseFunctionDefinition()
	case token.RETURN:
		stmt = p.parseReturnStatement()
	case token.WHILE:
		stmt = p.parseWhileStatement()
	case token.IDENT:
		if p.peekToken.Type == token.ASSIGN {
			stmt = p.parseAssignmentStatement()
		} else {
			stmt = p.parseExpressionStatement()
		}
	default:
		stmt = p.parseExpressionStatement()
	}
	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetDeclaration {
	if !p.expectPeek(token.IDENT) { return nil }
	name := p.currentToken.Literal
	var typeAnn *string
	if p.peekToken.Type == token.COLON {
		p.nextToken()
		if !p.expectPeek(token.IDENT) { return nil }
		annotation := p.currentToken.Literal
		typeAnn = &annotation
	}
	if !p.expectPeek(token.ASSIGN) { return nil }
	p.nextToken()
	value := p.parseExpression(LOWEST)
	return &ast.LetDeclaration{Name: name, TypeAnn: typeAnn, ValueExpression: value}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	name := p.currentToken.Literal
	if !p.expectPeek(token.ASSIGN) { return nil }
	p.nextToken()
	value := p.parseExpression(LOWEST)
	return &ast.AssignmentStatement{Name: name, Value: value}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	return p.parseExpressionUntil(precedence, token.SEMICOLON)
}

func (p *Parser) parseExpressionUntil(precedence int, until token.TokenType) ast.Expression {
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
		if infix == nil { return left }
		p.nextToken()
		left = infix(left)
	}
	return left
}

func (p *Parser) parseIdentifier() ast.Expression { return &ast.Identifier{Value: p.currentToken.Literal} }

func (p *Parser) parseIntegerLiteral() ast.Expression {
	value, err := strconv.Atoi(p.currentToken.Literal)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal))
		return nil
	}
	return &ast.IntegerLiteral{Value: value}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{}
	value, err := strconv.ParseFloat(p.currentToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression { return &ast.StringLiteral{Value: lexer.ProcessStringLiteral(p.currentToken.Literal)} }
func (p *Parser) parseBooleanLiteral() ast.Expression { return &ast.BooleanLiteral{Value: p.currentToken.Type == token.TRUE} }

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.UnaryExpression{Operator: tokenToUnaryOperator(p.currentToken.Type)}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func tokenToUnaryOperator(tok token.TokenType) ast.UnaryOperator {
	switch tok {
	case token.BANG: return ast.UnaryOpBang
	case token.MINUS: return ast.UnaryOpMinus
	}
	panic(fmt.Sprintf("tokenToUnaryOperator called with non-unary token: %s", tok))
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.BinaryExpression{Left: left, Operator: tokenToBinaryOperator(p.currentToken.Literal)}
	prec := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpressionUntil(prec, p.currentUntil)
	return expr
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) { p.errors = append(p.errors, fmt.Sprintf("no prefix parse function for %s found", t)) }

func ParseExpression(input string) (ast.Expression, error) {
	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)
	if len(p.errors) > 0 { return nil, errors.New(p.errors[0]) }
	return expr, nil
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	if !p.expectPeek(token.LBRACE) { return nil }
	var imports []string
	p.nextToken()
	if !p.isValidImportIdentifier() { return nil }
	imports = append(imports, p.currentToken.Literal)
	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		if !p.isValidImportIdentifier() { return nil }
		imports = append(imports, p.currentToken.Literal)
	}
	if !p.expectPeek(token.RBRACE) { return nil }
	if !p.expectPeek(token.FROM) { return nil }
	if !p.expectPeek(token.STRING) { return nil }
	module := p.currentToken.Literal
	if len(module) >= 2 && module[0] == '"' && module[len(module)-1] == '"' {
		module = module[1 : len(module)-1]
	}
	return &ast.ImportStatement{Imports: imports, Module: module}
}

func (p *Parser) isValidImportIdentifier() bool { return p.currentToken.Type == token.IDENT }

func (p *Parser) parseFunctionDefinition() *ast.FunctionDefinition { return p.parseFunctionDefinitionWithVisibility(false) }

func (p *Parser) parsePublicDeclaration() ast.Statement {
	if p.peekToken.Type != token.FN {
		p.errors = append(p.errors, "pub can only be used with function definitions")
		return nil
	}
	p.nextToken()
	return p.parseFunctionDefinitionWithVisibility(true)
}

func (p *Parser) parseFunctionDefinitionWithVisibility(isPublic bool) *ast.FunctionDefinition {
	if !p.expectPeek(token.IDENT) { return nil }
	name := p.currentToken.Literal
	if !p.expectPeek(token.LPAREN) { return nil }
	var parameters []ast.Parameter
	if p.peekToken.Type != token.RPAREN {
		p.nextToken()
		for {
			if p.currentToken.Type != token.IDENT {
				p.errors = append(p.errors, "expected parameter name")
				return nil
			}
			paramName := p.currentToken.Literal
			if !p.expectPeek(token.COLON) { return nil }
			if !p.expectPeek(token.IDENT) { return nil }
			paramType := p.currentToken.Literal
			parameters = append(parameters, ast.Parameter{Name: paramName, Type: paramType})
			if p.peekToken.Type == token.COMMA {
				p.nextToken()
				p.nextToken()
			} else { break }
		}
	}
	if !p.expectPeek(token.RPAREN) { return nil }
	var returnType *string
	if p.peekToken.Type == token.COLON {
		p.nextToken()
		if !p.expectPeek(token.IDENT) { return nil }
		retType := p.currentToken.Literal
		returnType = &retType
	}
	if !p.expectPeek(token.LBRACE) { return nil }
	bodyBlock := p.parseBlockStatement()
	if bodyBlock == nil { return nil }
	return &ast.FunctionDefinition{Name: name, Parameters: parameters, ReturnType: returnType, Body: bodyBlock.Statements, IsPublic: isPublic}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	var value ast.Expression
	if p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.EOF && p.peekToken.Type != token.RBRACE {
		p.nextToken()
		value = p.parseExpression(LOWEST)
	}
	if p.peekToken.Type == token.SEMICOLON { p.nextToken() }
	return &ast.ReturnStatement{Value: value}
}

// parseCommaSeparatedExpressions parses a list of comma-separated expressions until an endToken.
// It's called when p.currentToken is the opening delimiter (e.g., LBRACKET, LPAREN for function calls).
func (p *Parser) parseCommaSeparatedExpressions(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// If the next token is the end token, it's an empty list (e.g., "[]" or "()").
	if p.peekToken.Type == end {
		p.nextToken() // Consume the opening token (e.g., LBRACKET or LPAREN).
		p.nextToken() // Consume the closing end token (e.g., RBRACKET or RPAREN).
		return list
	}

	p.nextToken() // Consume the opening token. Current token is now the first token of the first expression.
	list = append(list, p.parseExpression(LOWEST)) // Parse the first expression.

	// Loop for subsequent comma-separated expressions.
	for p.peekToken.Type == token.COMMA {
		p.nextToken() // Consume the last token of the previous expression.
		p.nextToken() // Consume the COMMA. Current token is now COMMA.
		p.nextToken() // Advance to the start of the next expression.
		list = append(list, p.parseExpression(LOWEST))
	}

	// Expect and consume the endToken.
	if !p.expectPeek(end) {
		return nil // Error already registered by expectPeek.
	}
	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	// currentToken is token.LBRACKET when this prefixParseFn is called.
	array := &ast.ArrayLiteral{}
	array.Elements = p.parseCommaSeparatedExpressions(token.RBRACKET)
	return array
}

func (p *Parser) parseFunctionCall(functionExpression ast.Expression) ast.Expression {
	// currentToken is LPAREN when this (infixParseFn) is called.
	call := &ast.FunctionCall{Name: functionExpression.(*ast.Identifier).Value}
	call.Arguments = p.parseCommaSeparatedExpressions(token.RPAREN)
	return call
}

// // parseCallArguments was replaced by parseCommaSeparatedExpressions
// func (p *Parser) parseCallArguments() []ast.Expression { ... }


func (p *Parser) parseIfStatement() *ast.IfStatement {
	p.nextToken()
	condition := p.parseExpressionUntil(LOWEST, token.LBRACE)
	if condition == nil { return nil }
	if !p.expectPeek(token.LBRACE) { return nil }
	thenBlock := p.parseBlockStatement()
	if thenBlock == nil { return nil }
	var elseIfClauses []ast.ElseIfClause
	var elseBlock *ast.Block
	for p.peekToken.Type == token.ELSE {
		p.nextToken()
		p.nextToken()
		if p.currentToken.Type == token.IF {
			p.nextToken()
			elseIfCondition := p.parseExpressionUntil(LOWEST, token.LBRACE)
			if elseIfCondition == nil { return nil }
			if !p.expectPeek(token.LBRACE) { return nil }
			elseIfBlock := p.parseBlockStatement()
			if elseIfBlock == nil { return nil }
			elseIfClauses = append(elseIfClauses, ast.ElseIfClause{Condition: elseIfCondition, Block: elseIfBlock})
		} else if p.currentToken.Type == token.LBRACE {
			elseBlock = p.parseBlockStatement()
			if elseBlock == nil { return nil }
			break
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected 'if' or '{' after 'else', got %s", p.currentToken.Type))
			return nil
		}
	}
	return &ast.IfStatement{Condition: condition, ThenBlock: thenBlock, ElseIfClauses: elseIfClauses, ElseBlock: elseBlock}
}

func (p *Parser) parseBlockStatement() *ast.Block {
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

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	p.nextToken()
	condition := p.parseExpressionUntil(LOWEST, token.LBRACE)
	if condition == nil { return nil }
	if !p.expectPeek(token.LBRACE) { return nil }
	block := p.parseBlockStatement()
	if block == nil { return nil }
	return &ast.WhileStatement{Condition: condition, Block: block}
}
