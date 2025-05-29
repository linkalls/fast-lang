package parser

import (
	"testing"

	"github.com/linkalls/zeno-lang/ast"
	"github.com/linkalls/zeno-lang/lexer"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
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

func TestMutStatements(t *testing.T) {
	input := `
mut x = 5;
mut y: int = 10;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d",
			len(program.Statements))
	}

	// Test first statement (mut x = 5;)
	stmt := program.Statements[0]
	letStmt, ok := stmt.(*ast.LetDeclaration)
	if !ok {
		t.Fatalf("stmt not *ast.LetDeclaration. got=%T", stmt)
	}
	if letStmt.Name != "x" {
		t.Errorf("letStmt.Name not 'x'. got=%s", letStmt.Name)
	}
	if !letStmt.Mutable {
		t.Errorf("letStmt.Mutable not true. got=%t", letStmt.Mutable)
	}

	// Test second statement (mut y: int = 10;)
	stmt2 := program.Statements[1]
	letStmt2, ok := stmt2.(*ast.LetDeclaration)
	if !ok {
		t.Fatalf("stmt not *ast.LetDeclaration. got=%T", stmt2)
	}
	if letStmt2.Name != "y" {
		t.Errorf("letStmt2.Name not 'y'. got=%s", letStmt2.Name)
	}
	if !letStmt2.Mutable {
		t.Errorf("letStmt2.Mutable not true. got=%t", letStmt2.Mutable)
	}
	if letStmt2.TypeAnn == nil {
		t.Errorf("letStmt2.TypeAnn is nil")
	} else if *letStmt2.TypeAnn != "int" {
		t.Errorf("letStmt2.TypeAnn not 'int'. got=%s", *letStmt2.TypeAnn)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

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
