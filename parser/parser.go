package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/linkalls/zeno-lang/ast"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/token"
)

// ParseError represents a detailed parser error with position information
type ParseError struct {
	Message    string
	Line       int
	Column     int
	Token      token.Token
	Expected   string
	Got        string
	Context    string
	Suggestion string
}

func (e ParseError) String() string {
	var builder strings.Builder

	// Error header with position
	builder.WriteString(fmt.Sprintf("error: %s\n", e.Message))

	if e.Line > 0 {
		builder.WriteString(fmt.Sprintf("  --> line %d, column %d\n", e.Line, e.Column))
	}

	// Context information
	if e.Context != "" {
		builder.WriteString(fmt.Sprintf("   | %s\n", e.Context))
	}

	// Expected vs Got
	if e.Expected != "" && e.Got != "" {
		builder.WriteString(fmt.Sprintf("   = expected %s, but got %s\n", e.Expected, e.Got))
	}

	// Suggestion
	if e.Suggestion != "" {
		builder.WriteString(fmt.Sprintf("help: %s\n", e.Suggestion))
	}

	return builder.String()
}

// Precedence levels for operator precedence parsing
const (
	_ int = iota
	LOWEST
	LOGICAL_OR  // ||
	LOGICAL_AND // &&
	EQUALS      // ==, !=
	COMPARISON  // <, >, <=, >=
	SUM         // +, -
	PRODUCT     // *, /
	PREFIX      // -X or !X
	CALL        // myFunction(X)
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
	token.LBRACE:   CALL, // For struct literals
}

// Parser holds the state for parsing tokens into an AST
type Parser struct {
	l *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token

	errors         []string
	detailedErrors []ParseError
	filename       string
	input          string

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
		return ast.BinaryOpPlus
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
		l:              l,
		errors:         []string{},
		detailedErrors: []ParseError{},
		currentUntil:   token.SEMICOLON,
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
		token.LBRACE:   p.parseMapLiteral,   // Added for map literals
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
		token.LBRACE:   p.parseStructLiteral, // Added for struct literals
	}
	p.nextToken()
	p.nextToken()
	return p
}

// NewWithInput creates a parser with filename and input for better error reporting
func NewWithInput(l *lexer.Lexer, filename, input string) *Parser {
	p := New(l)
	p.filename = filename
	p.input = input
	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string { return p.errors }

// DetailedErrors returns the list of detailed ParseError structs
func (p *Parser) DetailedErrors() []ParseError { return p.detailedErrors }

// addDetailedError adds a detailed error with position information
func (p *Parser) addDetailedError(message, expected, got, context, suggestion string) {
	// Calculate line and column from current token position
	line, column := p.getTokenPosition()

	detailedErr := ParseError{
		Message:    message,
		Line:       line,
		Column:     column,
		Token:      p.currentToken,
		Expected:   expected,
		Got:        got,
		Context:    context,
		Suggestion: suggestion,
	}

	p.detailedErrors = append(p.detailedErrors, detailedErr)
	// Also add to simple errors for backward compatibility
	p.errors = append(p.errors, message)
}

// getTokenPosition calculates line and column from token position
func (p *Parser) getTokenPosition() (int, int) {
	if p.input == "" {
		return 0, 0
	}

	// Simple line/column calculation - can be improved with lexer position tracking
	lines := strings.Split(p.input[:min(len(p.input), len(p.currentToken.Literal))], "\n")
	line := len(lines)
	column := len(lines[len(lines)-1]) + 1
	return line, column
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (p *Parser) peekError(t token.TokenType) {
	expected := string(t)
	got := string(p.peekToken.Type)
	message := "expected next token to be " + expected + ", got " + got + " instead"

	context := ""
	if p.peekToken.Literal != "" {
		context = "near '" + p.peekToken.Literal + "'"
	}

	suggestion := ""
	if t == token.RPAREN && p.peekToken.Type == token.EOF {
		suggestion = "add missing ')' to close function call or expression"
	} else if t == token.RBRACE && p.peekToken.Type == token.EOF {
		suggestion = "add missing '}' to close block"
	}

	p.addDetailedError(message, expected, got, context, suggestion)
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
	case token.FOR:
		stmt = p.parseForStatement()
	case token.TYPE:
		return p.parseTypeDeclaration()
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
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	name := p.currentToken.Literal
	var typeAnn *string
	if p.peekToken.Type == token.COLON {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		// parse basic and generic type annotations (e.g., Result<int>)
		typeStr := p.currentToken.Literal
		if p.peekToken.Type == token.LT {
			for p.peekToken.Type != token.GT {
				p.nextToken()
				typeStr += p.currentToken.Literal
			}
			// consume '>'
			p.nextToken()
			typeStr += p.currentToken.Literal
		}
		annotation := typeStr
		typeAnn = &annotation
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
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
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
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

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Value: lexer.ProcessStringLiteral(p.currentToken.Literal)}
}
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Value: p.currentToken.Type == token.TRUE}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.UnaryExpression{Operator: tokenToUnaryOperator(p.currentToken.Type)}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func tokenToUnaryOperator(tok token.TokenType) ast.UnaryOperator {
	switch tok {
	case token.BANG:
		return ast.UnaryOpBang
	case token.MINUS:
		return ast.UnaryOpMinus
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

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	message := "no prefix parse function for " + string(t) + " found"
	expected := "expression"
	got := string(t)

	context := ""
	if p.currentToken.Literal != "" {
		context = "at '" + p.currentToken.Literal + "'"
	}

	suggestion := ""
	if t == token.RPAREN {
		suggestion = "unexpected ')' - check for empty function call like 'println()'"
	} else if t == token.EOF {
		suggestion = "unexpected end of file - check for incomplete expression"
	}

	p.addDetailedError(message, expected, got, context, suggestion)
}

func ParseExpression(input string) (ast.Expression, error) {
	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)
	if len(p.errors) > 0 {
		return nil, errors.New(p.errors[0])
	}
	return expr, nil
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	
	var items []ast.ImportItem
	
	// 最初の項目をパース
	p.nextToken()
	if p.currentToken.Type == token.RBRACE {
		// 空のインポート
		if !p.expectPeek(token.FROM) {
			return nil
		}
		if !p.expectPeek(token.STRING) {
			return nil
		}
		module := p.currentToken.Literal
		if len(module) >= 2 && module[0] == '"' && module[len(module)-1] == '"' {
			module = module[1 : len(module)-1]
		}
		return &ast.ImportStatement{Imports: items, Module: module}
	}
	
	for {
		isType := false
		if p.currentToken.Type == token.TYPE {
			isType = true
			if !p.expectPeek(token.IDENT) {
				return nil
			}
		}
		
		if !p.isValidImportIdentifier() {
			return nil
		}
		
		items = append(items, ast.ImportItem{Name: p.currentToken.Literal, IsType: isType})
		
		// 次のトークンをチェック
		if p.peekToken.Type == token.COMMA {
			p.nextToken() // COMMAを消費
			p.nextToken() // 次の項目へ移動
		} else if p.peekToken.Type == token.RBRACE {
			break
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected ',' or '}' in import statement, got %s", p.peekToken.Type))
			return nil
		}
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
	
	module := p.currentToken.Literal
	if len(module) >= 2 && module[0] == '"' && module[len(module)-1] == '"' {
		module = module[1 : len(module)-1]
	}
	return &ast.ImportStatement{Imports: items, Module: module}
}

func (p *Parser) isValidImportIdentifier() bool { return p.currentToken.Type == token.IDENT }

func (p *Parser) parseFunctionDefinition() *ast.FunctionDefinition {
	return p.parseFunctionDefinitionWithVisibility(false)
}

func (p *Parser) parsePublicDeclaration() ast.Statement {
	if p.peekToken.Type != token.FN {
		p.errors = append(p.errors, "pub can only be used with function definitions")
		return nil
	}
	p.nextToken()
	return p.parseFunctionDefinitionWithVisibility(true)
}

func (p *Parser) parseFunctionDefinitionWithVisibility(isPublic bool) *ast.FunctionDefinition {
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	name := p.currentToken.Literal
	// Parse generic type parameters, e.g., <T, U>
	var generics []string
	if p.peekToken.Type == token.LT {
		p.nextToken() // consume '<'
		for {
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			generics = append(generics, p.currentToken.Literal)
			if p.peekToken.Type == token.COMMA {
				p.nextToken() // consume ','
				continue
			}
			break
		}
		if !p.expectPeek(token.GT) {
			return nil
		}
	}
	// Expect '(' for parameters
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	var parameters []ast.Parameter
	if p.peekToken.Type != token.RPAREN {
		p.nextToken()
		for {
			// Check for variadic parameter (...)
			variadic := false
			if p.currentToken.Type == token.DOTDOTDOT {
				variadic = true
				if !p.expectPeek(token.IDENT) {
					return nil
				}
			} else if p.currentToken.Type != token.IDENT {
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
			// Parse parameter type (may include generics like Result<T>)
			paramType := p.currentToken.Literal

			// Check if this is a generic type
			if p.peekToken.Type == token.LT {
				p.nextToken() // consume '<'
				paramType += p.currentToken.Literal

				// Parse everything until we find the matching '>'
				depth := 1
				for depth > 0 && p.peekToken.Type != token.EOF {
					p.nextToken()
					paramType += p.currentToken.Literal
					if p.currentToken.Type == token.LT {
						depth++
					} else if p.currentToken.Type == token.GT {
						depth--
					}
				}
			}

			parameters = append(parameters, ast.Parameter{Name: paramName, Type: paramType, Variadic: variadic})

			// Variadic parameter must be the last one
			if variadic {
				if p.peekToken.Type == token.COMMA {
					message := "variadic parameter must be the last parameter"
					p.addDetailedError(message, "no more parameters", "comma", "after variadic parameter", "move variadic parameter to the end")
					return nil
				}
				break
			}

			if p.peekToken.Type == token.COMMA {
				p.nextToken()
				p.nextToken()
			} else {
				break
			}
		}
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	var returnType *string
	if p.peekToken.Type == token.COLON {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		// Start with the identifier (e.g., "Result" or "int")
		typeStr := p.currentToken.Literal

		// Check if this is a generic type (e.g., Result<T>)
		if p.peekToken.Type == token.LT {
			p.nextToken() // consume '<'
			typeStr += p.currentToken.Literal

			// Parse everything until we find the matching '>'
			depth := 1
			for depth > 0 && p.peekToken.Type != token.EOF {
				p.nextToken()
				typeStr += p.currentToken.Literal
				if p.currentToken.Type == token.LT {
					depth++
				} else if p.currentToken.Type == token.GT {
					depth--
				}
			}
		}

		retType := typeStr
		returnType = &retType
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	bodyBlock := p.parseBlockStatement()
	if bodyBlock == nil {
		return nil
	}
	return &ast.FunctionDefinition{Name: name, Generics: generics, Parameters: parameters, ReturnType: returnType, Body: bodyBlock.Statements, IsPublic: isPublic}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	var value ast.Expression
	if p.peekToken.Type != token.SEMICOLON && p.peekToken.Type != token.EOF && p.peekToken.Type != token.RBRACE {
		p.nextToken()
		value = p.parseExpression(LOWEST)
	}
	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}
	return &ast.ReturnStatement{Value: value}
}

// parseCommaSeparatedExpressions parses a list of comma-separated expressions until an endToken.
// It's called when p.currentToken is the opening delimiter (e.g., LBRACKET, LPAREN for function calls).
func (p *Parser) parseCommaSeparatedExpressions(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// If the next token is the end token, it's an empty list (e.g., "[]" or "()").
	if p.peekToken.Type == end {
		p.nextToken() // Consume the closing end token (e.g., RBRACKET or RPAREN).
		return list
	}

	p.nextToken()                                  // Consume the opening token. Current token is now the first token of the first expression.
	list = append(list, p.parseExpression(LOWEST)) // Parse the first expression.

	// Loop for subsequent comma-separated expressions.
	// After parsing the first element, p.currentToken is the last token of that element.
	// p.peekToken is expected to be a COMMA if there are more elements.
	for p.peekToken.Type == token.COMMA {
		p.nextToken() // Consume the last token of the previously parsed expression. p.currentToken is now this last token.
		// So, if previous was "1", currentToken is "1". peekToken is ",".
		// NO, nextToken advances currentToken to what peekToken was.
		// So, if currentToken was "1" (end of first expr), peekToken was ",".
		// After this first nextToken(), currentToken becomes ",". peekToken is start of next expr.

		p.nextToken()                                  // Consume the COMMA token itself. p.currentToken is now the first token of the next expression.
		list = append(list, p.parseExpression(LOWEST)) // Parse the next expression.
	}

	// Expect and consume the endToken.
	if !p.expectPeek(end) {
		return nil // Error already registered by expectPeek.
	}
	return list
}

// getExpressionPrimitiveType checks if an expression is a known primitive type
// and returns its type as a string (e.g., "INT", "STRING", "FLOAT", "BOOL").
// The second return value is true if it's a recognized primitive, false otherwise.
func getExpressionPrimitiveType(exp ast.Expression) (string, bool) {
	switch exp.(type) {
	case *ast.IntegerLiteral:
		return "INT", true
	case *ast.FloatLiteral:
		return "FLOAT", true
	case *ast.StringLiteral:
		return "STRING", true
	case *ast.BooleanLiteral:
		return "BOOL", true
	default:
		// For this phase, we consider anything else as non-primitive or unknown.
		return fmt.Sprintf("%T", exp), false
	}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{}
	// currentToken is token.LBRACKET when this prefixParseFn is called.
	// parseCommaSeparatedExpressions handles the parsing of elements between LBRACKET and RBRACKET.
	array.Elements = p.parseCommaSeparatedExpressions(token.RBRACKET)

	if array.Elements == nil {
		// This can happen if parseCommaSeparatedExpressions itself encountered an error
		// (e.g., missing RBRACKET) and returned nil. Errors would already be recorded.
		return nil
	}

	if len(array.Elements) > 0 {
		firstElementType, isFirstPrimitive := getExpressionPrimitiveType(array.Elements[0])
		if !isFirstPrimitive {
			msg := fmt.Sprintf("array element type is not a primitive type (int, float, string, bool), got %s for first element", firstElementType)
			p.errors = append(p.errors, msg)
			// Return the array to allow collecting more syntax errors; type errors are semantic.
			// The generator/type-checker will ultimately decide if this partially valid AST is usable.
		}

		// We proceed with type checking against the first element's type only if it was primitive.
		if isFirstPrimitive {
			for i := 1; i < len(array.Elements); i++ {
				element := array.Elements[i]
				elementType, isPrimitive := getExpressionPrimitiveType(element)

				if !isPrimitive {
					msg := fmt.Sprintf("array element type is not a primitive type (int, float, string, bool), got %s at index %d (expected %s)", elementType, i, firstElementType)
					p.errors = append(p.errors, msg)
					continue // Continue to find all non-primitive elements
				}

				if elementType != firstElementType {
					msg := fmt.Sprintf("mismatched types in array literal: expected %s, got %s at index %d", firstElementType, elementType, i)
					p.errors = append(p.errors, msg)
					// Continue to find all mismatches against the first primitive type
				}
			}
		}
		// If the first element was not primitive, we've already logged an error for it.
		// We don't iterate further for mismatches because there's no valid "expected primitive type".
		// However, we could iterate to find *other* non-primitive types if desired.
		// For now, the logic is: if first is primitive, all others must match that primitive.
		// If first is not primitive, that's an error, and further elements are not checked against it.
	}
	return array
}

func (p *Parser) parseMapLiteral() ast.Expression {
	// currentToken is token.LBRACE when this prefixParseFn is called.
	mapLiteral := &ast.MapLiteral{Pairs: make(map[ast.Expression]ast.Expression)}

	// Handle empty map {}
	if p.peekToken.Type == token.RBRACE {
		p.nextToken() // Consume LBRACE
		p.nextToken() // Consume RBRACE
		return mapLiteral
	}

	p.nextToken() // Consume LBRACE, currentToken is now the first token of the first key.

	for p.currentToken.Type != token.RBRACE && p.currentToken.Type != token.EOF {
		// Parse Key
		key := p.parseExpression(LOWEST)
		if key == nil {
			// Error already recorded by parseExpression or its children
			return nil
		}

		// Validate Key Type
		switch key.(type) {
		case *ast.Identifier, *ast.StringLiteral:
			// Valid key type
		default:
			msg := fmt.Sprintf("invalid map key type: expected IDENTIFIER or STRING, got %T", key)
			p.errors = append(p.errors, msg)
			return nil
		}

		if !p.expectPeek(token.COLON) {
			// Error: "expected next token to be COLON, got <actual_token> instead" already added by expectPeek
			return nil
		}
		p.nextToken() // Consume COLON, currentToken is now the first token of the value.

		// Parse Value
		value := p.parseExpression(LOWEST)
		if value == nil {
			// Error already recorded by parseExpression or its children
			return nil
		}
		mapLiteral.Pairs[key] = value

		// After parsing a value, p.currentToken is the last token of the value expression.
		// p.peekToken is what comes AFTER the value expression (e.g., COMMA or RBRACE).
		if p.peekToken.Type == token.COMMA {
			p.nextToken() // Advances p.currentToken to be the COMMA (it was previously value's end).
			// Now p.currentToken is COMMA.

			// Check for trailing comma: e.g. {key: value,}
			if p.peekToken.Type == token.RBRACE {
				p.nextToken() // Consume COMMA, p.currentToken is now RBRACE.
				break         // Exit loop, RBRACE is currentToken.
			}
			p.nextToken() // Consume COMMA, p.currentToken is now the first token of the next key.
		} else if p.peekToken.Type == token.RBRACE {
			p.nextToken() // Advances p.currentToken to be the RBRACE.
			break         // Exit loop, RBRACE is currentToken.
		} else if p.peekToken.Type == token.EOF { // Premature EOF
			msg := "expected ',' or '}' after map value, got EOF"
			p.errors = append(p.errors, msg)
			return nil
		} else { // Unexpected token
			msg := fmt.Sprintf("expected ',' or '}' after map value, got %s instead", p.peekToken.Type)
			p.errors = append(p.errors, msg)
			return nil
		}
	} // End of for loop

	if p.currentToken.Type != token.RBRACE {
		// Only report block close error if no prior errors
		if len(p.detailedErrors) == 0 {
			message := "expected '}' to close block"
			expected := "}"
			got := string(p.currentToken.Type)
			context := ""
			if p.currentToken.Literal != "" {
				context = "found '" + p.currentToken.Literal + "'"
			}
			suggestion := "add missing '}' to close block"
			p.addDetailedError(message, expected, got, context, suggestion)
		}
		return nil
	}
	return mapLiteral
}

func (p *Parser) parseStructLiteral(typeExpr ast.Expression) ast.Expression {
	// currentToken is LBRACE when this (infixParseFn) is called.
	// typeExpr should be an Identifier representing the struct type name
	
	var typeName string
	if ident, ok := typeExpr.(*ast.Identifier); ok {
		typeName = ident.Value
	} else {
		message := "struct literal requires a type name"
		p.addDetailedError(message, "identifier", "expression", "", "use a type name like Result{...}")
		return nil
	}

	structLiteral := &ast.StructLiteral{
		TypeName: typeName,
		Fields:   make(map[string]ast.Expression),
	}

	// Handle empty struct literal TypeName{}
	if p.peekToken.Type == token.RBRACE {
		p.nextToken() // Consume LBRACE
		p.nextToken() // Consume RBRACE
		return structLiteral
	}

	p.nextToken() // Consume LBRACE, currentToken is now the first token of the first field name.

	for p.currentToken.Type != token.RBRACE && p.currentToken.Type != token.EOF {
		// Parse Field Name - must be an identifier
		if p.currentToken.Type != token.IDENT {
			msg := fmt.Sprintf("struct field name must be identifier, got %s", p.currentToken.Type)
			p.errors = append(p.errors, msg)
			return nil
		}
		
		fieldName := p.currentToken.Literal

		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken() // Consume COLON, currentToken is now the first token of the field value.

		// Parse Field Value
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		structLiteral.Fields[fieldName] = value

		// After parsing a value, p.currentToken is the last token of the value expression.
		// p.peekToken is what comes AFTER the value expression (e.g., COMMA or RBRACE).
		if p.peekToken.Type == token.COMMA {
			p.nextToken() // Advances p.currentToken to be the COMMA
			
			// Check for trailing comma: e.g. Result{ok: true,}
			if p.peekToken.Type == token.RBRACE {
				p.nextToken() // Consume COMMA, p.currentToken is now RBRACE.
				break         // Exit loop, RBRACE is currentToken.
			}
			p.nextToken() // Consume COMMA, p.currentToken is now the first token of the next field.
		} else if p.peekToken.Type == token.RBRACE {
			p.nextToken() // Advances p.currentToken to be the RBRACE.
			break         // Exit loop, RBRACE is currentToken.
		} else if p.peekToken.Type == token.EOF {
			msg := "expected ',' or '}' after struct field value, got EOF"
			p.errors = append(p.errors, msg)
			return nil
		} else {
			msg := fmt.Sprintf("expected ',' or '}' after struct field value, got %s instead", p.peekToken.Type)
			p.errors = append(p.errors, msg)
			return nil
		}
	}

	if p.currentToken.Type != token.RBRACE {
		if len(p.detailedErrors) == 0 {
			message := "expected '}' to close struct literal"
			expected := "}"
			got := string(p.currentToken.Type)
			context := ""
			if p.currentToken.Literal != "" {
				context = "found '" + p.currentToken.Literal + "'"
			}
			suggestion := "add missing '}' to close struct literal"
			p.addDetailedError(message, expected, got, context, suggestion)
		}
		return nil
	}
	return structLiteral
}

func (p *Parser) parseFunctionCall(functionExpression ast.Expression) ast.Expression {
	// currentToken is LPAREN when this (infixParseFn) is called.
	var functionName string
	if ident, ok := functionExpression.(*ast.Identifier); ok {
		functionName = ident.Value
	} else {
		// This shouldn't happen in current Zeno language design, but let's handle it gracefully
		message := "function call on non-identifier expression not supported"
		p.addDetailedError(message, "identifier", "expression", "", "use a function name instead of expression")
		return nil
	}

	call := &ast.FunctionCall{Name: functionName}
	call.Arguments = p.parseCommaSeparatedExpressions(token.RPAREN)

	// Handle parsing errors from parseCommaSeparatedExpressions
	if call.Arguments == nil {
		call.Arguments = []ast.Expression{} // Ensure it's not nil
	}

	// Special check for empty function calls to provide better error messages
	if len(call.Arguments) == 0 {
		// Check if this is a function that requires arguments
		if functionName == "println" || functionName == "print" {
			message := "function '" + functionName + "' requires at least one argument"
			p.addDetailedError(message, "at least one argument", "empty call", "in function call", "add an argument like println(\"Hello\")")
			// Return the call anyway to allow parser to continue
		}
	}

	return call
}

// // parseCallArguments was replaced by parseCommaSeparatedExpressions
// func (p *Parser) parseCallArguments() []ast.Expression { ... }

func (p *Parser) parseIfStatement() *ast.IfStatement {
	p.nextToken()
	condition := p.parseExpressionUntil(LOWEST, token.LBRACE)
	if condition == nil {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	thenBlock := p.parseBlockStatement()
	if thenBlock == nil {
		return nil
	}
	var elseIfClauses []ast.ElseIfClause
	var elseBlock *ast.Block
	for p.peekToken.Type == token.ELSE {
		p.nextToken()
		p.nextToken()
		if p.currentToken.Type == token.IF {
			p.nextToken()
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
			elseIfClauses = append(elseIfClauses, ast.ElseIfClause{Condition: elseIfCondition, Block: elseIfBlock})
		} else if p.currentToken.Type == token.LBRACE {
			elseBlock = p.parseBlockStatement()
			if elseBlock == nil {
				return nil
			}
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
		// Only report block close error if no prior errors
		if len(p.detailedErrors) == 0 {
			message := "expected '}' to close block"
			expected := "}"
			got := string(p.currentToken.Type)
			context := ""
			if p.currentToken.Literal != "" {
				context = "found '" + p.currentToken.Literal + "'"
			}
			suggestion := "add missing '}' to close block"
			p.addDetailedError(message, expected, got, context, suggestion)
		}
		return nil
	}
	return block
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	p.nextToken()
	condition := p.parseExpressionUntil(LOWEST, token.LBRACE)
	if condition == nil {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	block := p.parseBlockStatement()
	if block == nil {
		return nil
	}
	return &ast.WhileStatement{Condition: condition, Block: block}
}

// parseForStatement parses 'for <ident> in <expression> { ... }'
func (p *Parser) parseForStatement() *ast.ForStatement {
	// currentToken is FOR
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	varName := p.currentToken.Literal
	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken() // move to iterable expression
	iterable := p.parseExpressionUntil(LOWEST, token.LBRACE)
	if iterable == nil {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	body := p.parseBlockStatement()
	if body == nil {
		return nil
	}
	return &ast.ForStatement{VarName: varName, Iterable: iterable, Body: body}
}

// parseTypeDeclaration parses 'type Name<Generics> = { ... }'
func (p *Parser) parseTypeDeclaration() *ast.TypeDeclaration {
	// currentToken is TYPE
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	name := p.currentToken.Literal
	var generics []string
	if p.peekToken.Type == token.LT {
		// consume '<'
		p.nextToken()
		for p.peekToken.Type != token.GT && p.peekToken.Type != token.EOF {
			p.nextToken()
			if p.currentToken.Type == token.IDENT {
				generics = append(generics, p.currentToken.Literal)
			}
		}
		// consume '>'
		p.nextToken()
	}
	// expect '='
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	// expect '{'
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	// skip until matching '}'
	depth := 1
	for depth > 0 && p.currentToken.Type != token.EOF {
		p.nextToken()
		if p.currentToken.Type == token.LBRACE {
			depth++
		} else if p.currentToken.Type == token.RBRACE {
			depth--
		}
	}
	return &ast.TypeDeclaration{Name: name, Generics: generics, Fields: nil}
}
