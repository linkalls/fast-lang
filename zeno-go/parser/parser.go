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
	_ int = iota
	LOWEST
	SUM     // +, -
	PRODUCT // *, /
	PREFIX  // -X or !X
	CALL    // myFunction(X)
)

// precedences maps tokens to their precedence
var precedences = map[token.TokenType]int{
	token.EQ:       LOWEST,
	token.NOT_EQ:   LOWEST,
	token.LT:       LOWEST,
	token.GT:       LOWEST,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.DIVIDE:   PRODUCT,
	token.MULTIPLY: PRODUCT,
}

// Parser represents the parser
type Parser struct {
	l *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
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
		l:      l,
		errors: []string{},
	}

	// Initialize prefix parse functions
	p.prefixParseFns = map[token.TokenType]prefixParseFn{
		token.IDENT:  p.parseIdentifier,
		token.INT:    p.parseIntegerLiteral,
		token.STRING: p.parseStringLiteral,
	}

	// Initialize infix parse functions
	p.infixParseFns = map[token.TokenType]infixParseFn{
		token.PLUS:     p.parseInfixExpression,
		token.MINUS:    p.parseInfixExpression,
		token.MULTIPLY: p.parseInfixExpression,
		token.DIVIDE:   p.parseInfixExpression,
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
	case token.LET:
		return p.parseLetStatement()
	case token.MUT:
		return p.parseMutStatement()
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

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return &ast.LetDeclaration{
		Name:            name,
		TypeAnn:         typeAnn,
		Mutable:         false,
		ValueExpression: value,
	}
}

func (p *Parser) parseMutStatement() *ast.LetDeclaration {
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

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return &ast.LetDeclaration{
		Name:            name,
		TypeAnn:         typeAnn,
		Mutable:         true,
		ValueExpression: value,
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}

	leftExp := prefix()

	for p.peekToken.Type != token.SEMICOLON && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Value: p.currentToken.Literal}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.BinaryExpression{
		Left:     left,
		Operator: tokenToBinaryOperator(p.currentToken.Literal),
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)

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
