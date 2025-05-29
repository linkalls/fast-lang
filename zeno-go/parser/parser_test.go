package parser

import (
	"testing"

	"github.com/linkalls/zeno-lang/ast"
	"github.com/linkalls/zeno-lang/lexer"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5
let y = 10
let foobar = 838383
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func TestAssignmentStatements(t *testing.T) {
	input := `
let x = 5
x = 10
let y = 1
y = 2
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d",
			len(program.Statements))
	}

	// Test first statement (let x = 5)
	stmt := program.Statements[0]
	letStmt, ok := stmt.(*ast.LetDeclaration)
	if !ok {
		t.Fatalf("stmt not *ast.LetDeclaration. got=%T", stmt)
	}
	if letStmt.Name != "x" {
		t.Errorf("letStmt.Name not 'x'. got=%s", letStmt.Name)
	}

	// Test second statement (x = 10)
	stmt2 := program.Statements[1]
	assignStmt, ok := stmt2.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("stmt not *ast.AssignmentStatement. got=%T", stmt2)
	}
	if assignStmt.Name != "x" {
		t.Errorf("assignStmt.Name not 'x'. got=%s", assignStmt.Name)
	}

	// Test third statement (let y = 1)
	stmt3 := program.Statements[2]
	letStmt2, ok := stmt3.(*ast.LetDeclaration)
	if !ok {
		t.Fatalf("stmt not *ast.LetDeclaration. got=%T", stmt3)
	}
	if letStmt2.Name != "y" {
		t.Errorf("letStmt2.Name not 'y'. got=%s", letStmt2.Name)
	}

	// Test fourth statement (y = 2)
	stmt4 := program.Statements[3]
	assignStmt2, ok := stmt4.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("stmt not *ast.AssignmentStatement. got=%T", stmt4)
	}
	if assignStmt2.Name != "y" {
		t.Errorf("assignStmt2.Name not 'y'. got=%s", assignStmt2.Name)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	letStmt, ok := s.(*ast.LetDeclaration)
	if !ok {
		t.Errorf("s not *ast.LetDeclaration. got=%T", s)
		return false
	}

	if letStmt.Name != name {
		t.Errorf("letStmt.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
