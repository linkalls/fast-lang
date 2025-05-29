package generator

import (
	"strings"
	"testing"

	"github.com/linkalls/zeno-lang/ast"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

// Helper function to run generator tests
func runGeneratorTest(t *testing.T, inputZeno string, expectedGoSubstrings []string) string {
	l := lexer.New(inputZeno)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors for input:\n%s\nErrors: %v", inputZeno, p.Errors())
	}

	goCode, err := Generate(program)
	if err != nil {
		t.Fatalf("Generator error for input:\n%s\nError: %v", inputZeno, err)
	}

	t.Logf("\n--- Zeno Input:\n%s\n--- Generated Go Output:\n%s\n---", inputZeno, goCode)

	for _, sub := range expectedGoSubstrings {
		if !strings.Contains(goCode, sub) {
			t.Errorf("Generated code does not contain expected substring: '%s'.\nFull code:\n%s", sub, goCode)
		}
	}

	// Basic check for balanced braces
	braceCount := 0
	for _, char := range goCode {
		if char == '{' {
			braceCount++
		} else if char == '}' {
			braceCount--
		}
	}
	if braceCount != 0 {
		t.Errorf("Unbalanced braces in generated code for input:\n%s", inputZeno)
	}

	return goCode
}

func TestGenerateLetDeclarations(t *testing.T) {
	runGeneratorTest(t, "let x = 10;", []string{
		"package main",
		"import (",
		"\"fmt\"",
		"func main() {",
		"var x = 10",
		"}",
	})
}

func TestGenerateStringLiterals(t *testing.T) {
	runGeneratorTest(t, `let s = "Hello World";`, []string{
		`var s = "Hello World"`,
	})
}

func TestGenerateStructure(t *testing.T) {
	code := runGeneratorTest(t, "let x = 42;", []string{
		"package main",
		"func main() {",
		"var x = 42",
		"}",
	})

	// Check the overall structure
	lines := strings.Split(code, "\n")
	if len(lines) < 6 {
		t.Errorf("Generated code should have at least 6 lines, got %d", len(lines))
	}

	// Check that it starts with package declaration
	if !strings.HasPrefix(lines[0], "package main") {
		t.Errorf("Code should start with 'package main', got: %s", lines[0])
	}

	// Check that it has proper imports
	hasImport := false
	for _, line := range lines {
		if strings.Contains(line, "import") {
			hasImport = true
			break
		}
	}
	if !hasImport {
		t.Error("Generated code should have import statement")
	}
}

// Test direct AST node generation (using only currently supported types)
func TestGenerateIntegerLiteral(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetDeclaration{
				Name:            "x",
				Mutable:         false,
				ValueExpression: &ast.IntegerLiteral{Value: 42},
			},
		},
	}

	code, err := Generate(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	if !strings.Contains(code, "var x = 42") {
		t.Errorf("Generated code should contain 'var x = 42', got:\n%s", code)
	}
}

func TestGenerateStringEscaping(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetDeclaration{
				Name:            "s",
				Mutable:         false,
				ValueExpression: &ast.StringLiteral{Value: "Hello\nWorld\t\"Quoted\""},
			},
		},
	}

	code, err := Generate(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// The string should be properly escaped
	if !strings.Contains(code, `"Hello\nWorld\t\"Quoted\""`) {
		t.Errorf("Generated code should properly escape strings, got:\n%s", code)
	}
}

func TestGenerateTypeAnnotations(t *testing.T) {
	typeAnn := "int"
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetDeclaration{
				Name:            "x",
				TypeAnn:         &typeAnn,
				Mutable:         false,
				ValueExpression: &ast.IntegerLiteral{Value: 42},
			},
		},
	}

	code, err := Generate(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	if !strings.Contains(code, "var x int64 = 42") {
		t.Errorf("Generated code should contain type annotation, got:\n%s", code)
	}
}

func TestGenerateIdentifier(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.Identifier{Value: "someVariable"},
			},
		},
	}

	code, err := Generate(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	if !strings.Contains(code, "someVariable") {
		t.Errorf("Generated code should contain identifier, got:\n%s", code)
	}
}

// Test simple binary operations
func TestGenerateBinaryExpressionBasic(t *testing.T) {
	// This test uses direct AST construction since parser doesn't support binary expressions yet
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetDeclaration{
				Name:    "result",
				Mutable: false,
				ValueExpression: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: ast.BinaryOpPlus,
					Right:    &ast.IntegerLiteral{Value: 3},
				},
			},
		},
	}

	code, err := Generate(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	if !strings.Contains(code, "var result = (5 + 3)") {
		t.Errorf("Generated code should contain binary expression, got:\n%s", code)
	}
}
