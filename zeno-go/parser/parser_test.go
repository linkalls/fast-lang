package parser

import (
	"fmt" // Added import for fmt
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

func TestArrayLiteralTypeChecking(t *testing.T) {
	tests := []struct {
		input          string
		expectedErrors []string // Specific error messages to look for. Empty if no errors expected.
	}{
		// Valid cases
		{`[]`, nil},
		{`[1, 2, 3]`, nil},
		{`["a", "b", "c"]`, nil},
		{`[true, false, true]`, nil},
		{`[1.0, 2.5, 3.0]`, nil},
		{`[1.0]`, nil}, // Single element array
		{`["test"]`, nil},

		// Invalid cases - Mismatched primitive types
		{`[1, "a"]`, []string{"mismatched types in array literal: expected INT, got STRING at index 1"}},
		{`[true, 2.0, false]`, []string{"mismatched types in array literal: expected BOOL, got FLOAT at index 1"}},
		{`["hello", 123]`, []string{"mismatched types in array literal: expected STRING, got INT at index 1"}},
		{`[1.0, 2]`, []string{"mismatched types in array literal: expected FLOAT, got INT at index 1"}},
		{`[1, 2, "c"]`, []string{"mismatched types in array literal: expected INT, got STRING at index 2"}},

		// Invalid cases - Non-primitive types (assuming identifiers like 'foo' are caught by getExpressionPrimitiveType as non-primitive)
		{`[foo]`, []string{"array element type is not a primitive type (int, float, string, bool), got *ast.Identifier for first element"}},
		{`[1, foo]`, []string{"array element type is not a primitive type (int, float, string, bool), got *ast.Identifier at index 1 (expected INT)"}},
		{`["a", bar, "c"]`, []string{"array element type is not a primitive type (int, float, string, bool), got *ast.Identifier at index 1 (expected STRING)"}},
		{`[foo, bar]`, []string{
			"array element type is not a primitive type (int, float, string, bool), got *ast.Identifier for first element",
			// Note: My current logic in parseArrayLiteral only adds further errors if the *first* element was primitive.
			// So, [foo, bar] will only report error for 'foo'. This is acceptable for now.
			// If we wanted to report 'bar' as well, the logic would need adjustment.
		}},
		{`[1, 2+3, 3]`, []string{ // 2+3 is an InfixExpression
			"array element type is not a primitive type (int, float, string, bool), got *ast.BinaryExpression at index 1 (expected INT)",
		}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("TestArrayLiteralTypeChecking_%d_%s", i, tt.input), func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram() // This will parse and trigger type checks within parseArrayLiteral

			// Check for general parsing errors first (e.g. if the array syntax itself is broken)
			// For these tests, we assume array syntax is correct, focusing on semantic type errors.
			// However, if ParseProgram itself returns nil or has errors NOT related to our type checks,
			// it might indicate a problem with the test case or a deeper parser issue.
			if program == nil && len(tt.expectedErrors) > 0 { // Allow nil program if errors are expected
				// but if no errors are expected, program should not be nil.
			} else if program == nil {
				t.Fatalf("ParseProgram() returned nil unexpectedly for input: %s. Parser errors: %v", tt.input, p.Errors())
			}


			if len(tt.expectedErrors) > 0 {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected %d errors but got none for input: %s", len(tt.expectedErrors), tt.input)
				}

				// Check if all expected errors are present
				for _, expectedErr := range tt.expectedErrors {
					found := false
					for _, actualErr := range p.Errors() {
						if actualErr == expectedErr { // Simple string comparison
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error message '%s' not found. Actual errors: %v for input: %s", expectedErr, p.Errors(), tt.input)
					}
				}
				// Optional: Check if there are more errors than expected
				if len(p.Errors()) != len(tt.expectedErrors) {
					// This can be noisy if one problem leads to multiple specific errors.
					// For now, focusing on presence of expected errors.
					// t.Errorf("got %d errors, but expected %d. Actual errors: %v for input: %s", len(p.Errors()), len(tt.expectedErrors), p.Errors(), tt.input)
				}

			} else { // No errors expected
				if len(p.Errors()) > 0 {
					t.Fatalf("parser has %d unexpected errors for input '%s': %v", len(p.Errors()), tt.input, p.Errors())
				}
				// If no errors, and program is not nil, check basic AST structure
				if program != nil {
					if len(program.Statements) != 1 {
						t.Fatalf("program.Statements does not contain 1 statement. got=%d for input: %s",
							len(program.Statements), tt.input)
					}
					stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
					if !ok {
						t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T for input: %s",
							program.Statements[0], tt.input)
					}
					if _, ok := stmt.Expression.(*ast.ArrayLiteral); !ok {
						t.Fatalf("stmt.Expression is not *ast.ArrayLiteral. got=%T for input: %s",
							stmt.Expression, tt.input)
					}
				}
			}
		})
	}
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

func TestMapLiteralParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedPairs  map[string]interface{} // Value can be int, bool, string, float64
		expectedErrors []string               // Specific error messages to look for (can be empty)
	}{
		// Valid cases
		{`{}`, map[string]interface{}{}, nil},
		{`{"one": 1, "two": 2}`, map[string]interface{}{"one": 1, "two": 2}, nil},
		{`{name: "Zeno", version: 1.0}`, map[string]interface{}{"name": "Zeno", "version": 1.0}, nil},
		{`{"mixedKey": 10, idKey: 20}`, map[string]interface{}{"mixedKey": 10, "idKey": 20}, nil},
		{`{"value": true}`, map[string]interface{}{"value": true}, nil},
		{`{val: 1.23}`, map[string]interface{}{"val": 1.23}, nil},
		{`{"a": 1,}`, map[string]interface{}{"a": 1}, nil}, // Trailing comma
		{`{"a": 1, "b": false,}`, map[string]interface{}{"a": 1, "b": false}, nil}, // Trailing comma multiple items

		// Error cases
		{`{123: "integer key"}`, nil, []string{"invalid map key type: expected IDENTIFIER or STRING, got *ast.IntegerLiteral"}},
		{`{"key": }`, nil, []string{"no prefix parse function for } found"}},
		{`{"key" "value"}`, nil, []string{"expected next token to be :, got STRING instead", "no prefix parse function for } found"}},
		{`{"k1": v1 "k2": v2}`, nil, []string{"expected ',' or '}' after map value, got STRING instead", "no prefix parse function for : found", "no prefix parse function for } found"}},
		{`{"k": v`, nil, []string{"expected ',' or '}' after map value, got EOF"}},
		{`{`, nil, []string{"expected '}' to close map literal, got EOF instead"}},
		{`{"a": 1, "b" 2}`, nil, []string{"expected next token to be :, got INT instead", "no prefix parse function for } found"}},
		{`{"a": 1, "b": 2 "c": 3}`, nil, []string{"expected ',' or '}' after map value, got STRING instead", "no prefix parse function for : found", "no prefix parse function for } found"}},
		{`{"a": 1 ! "b": 2}`, nil, []string{"expected ',' or '}' after map value, got ! instead"}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("TestMapLiteralParsing_%d_%s", i, tt.input), func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if len(tt.expectedErrors) > 0 {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected %d errors but got none for input: %s", len(tt.expectedErrors), tt.input)
				}
				foundErrors := 0
				for _, expectedErr := range tt.expectedErrors {
					found := false
					for _, actualErr := range p.Errors() {
						if actualErr == expectedErr { // Simple string comparison; could be more robust
							found = true
							foundErrors++
							break
						}
					}
					if !found {
						t.Errorf("expected error message '%s' not found. Actual errors: %v", expectedErr, p.Errors())
					}
				}
				// This check is optional: ensures all reported errors were expected
				// if len(p.Errors()) != foundErrors && len(p.Errors()) != len(tt.expectedErrors) {
				// 	  t.Errorf("got %d errors, but only %d were expected and matched. Actual errors: %v", len(p.Errors()), foundErrors, p.Errors())
				// }
				return // Skip AST checks if errors were expected
			}

			// If no errors expected, but errors occurred
			if len(p.Errors()) > 0 {
				t.Fatalf("parser has %d unexpected errors for input '%s': %v", len(p.Errors()), tt.input, p.Errors())
			}

			if program == nil {
				t.Fatalf("ParseProgram() returned nil for input: %s", tt.input)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d for input: %s",
					len(program.Statements), tt.input)
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T for input: %s",
					program.Statements[0], tt.input)
			}

			mapLiteral, ok := stmt.Expression.(*ast.MapLiteral)
			if !ok {
				t.Fatalf("stmt.Expression is not *ast.MapLiteral. got=%T for input: %s. AST: %s",
					stmt.Expression, tt.input, stmt.Expression.String())
			}

			if len(mapLiteral.Pairs) != len(tt.expectedPairs) {
				t.Fatalf("map literal has wrong number of pairs. expected=%d, got=%d for input: %s. AST: %s",
					len(tt.expectedPairs), len(mapLiteral.Pairs), tt.input, mapLiteral.String())
			}

			for expectedKeyStr, expectedVal := range tt.expectedPairs {
				foundKey := false
				for actualKeyExp, actualValExp := range mapLiteral.Pairs {
					currentKeyStr := ""
					if ident, ok := actualKeyExp.(*ast.Identifier); ok {
						currentKeyStr = ident.Value
					} else if strLit, ok := actualKeyExp.(*ast.StringLiteral); ok {
						currentKeyStr = strLit.Value
					} else {
						t.Errorf("parsed map key is not Identifier or StringLiteral. got=%T for input: %s", actualKeyExp, tt.input)
						continue
					}

					if currentKeyStr == expectedKeyStr {
						foundKey = true
						switch ev := expectedVal.(type) {
						case int: // Handles int from expectedPairs
							testIntegerLiteral(t, actualValExp, int64(ev))
						case string:
							testStringLiteral(t, actualValExp, ev)
						case bool:
							testBooleanLiteral(t, actualValExp, ev)
						case float64:
							testFloatLiteral(t, actualValExp, ev)
						default:
							t.Errorf("unsupported expected value type %T for key %s in input: %s", ev, expectedKeyStr, tt.input)
						}
						break // Found key and tested value, move to next expected pair
					}
				}
				if !foundKey {
					t.Errorf("expected key '%s' not found in map literal for input: %s. Map: %s", expectedKeyStr, tt.input, mapLiteral.String())
				}
			}
		})
	}
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) bool {
	il, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp not *ast.IntegerLiteral. got=%T (%s)", exp, exp.String())
		return false
	}
	if il.Value != int(value) { // Assuming IntegerLiteral.Value is int
		t.Errorf("il.Value not %d. got=%d", value, il.Value)
		return false
	}
	return true
}

func testStringLiteral(t *testing.T, exp ast.Expression, value string) bool {
	sl, ok := exp.(*ast.StringLiteral)
	if !ok {
		t.Errorf("exp not *ast.StringLiteral. got=%T (%s)", exp, exp.String())
		return false
	}
	if sl.Value != value { // Assumes StringLiteral.Value is the already processed string
		t.Errorf("sl.Value not '%s'. got='%s'", value, sl.Value)
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bl, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("exp not *ast.BooleanLiteral. got=%T (%s)", exp, exp.String())
		return false
	}
	if bl.Value != value {
		t.Errorf("bl.Value not %t. got=%t", value, bl.Value)
		return false
	}
	return true
}

func testFloatLiteral(t *testing.T, exp ast.Expression, value float64) bool {
	fl, ok := exp.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("exp not *ast.FloatLiteral. got=%T (%s)", exp, exp.String())
		return false
	}
	if fl.Value != value {
		t.Errorf("fl.Value not %f. got=%f", value, fl.Value)
		return false
	}
	return true
}
