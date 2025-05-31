package linter

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/linkalls/zeno-lang/ast"
)

// --- Helper Functions for Case Checking ---

// isLowerCamelCase checks if s is valid lowerCamelCase.
// e.g., myVariable, anotherOne.
// Simple initial check: starts with lowercase, no underscores/hyphens.
func isLowerCamelCase(s string) bool {
	if len(s) == 0 {
		return false // Empty string is not valid lowerCamelCase
	}
	firstChar := rune(s[0])
	if !(unicode.IsLower(firstChar) && unicode.IsLetter(firstChar)) {
		return false
	}
	// Allow digits after the first character, but no underscores or hyphens.
	// More sophisticated checks could allow sequences of uppercase letters if not first.
	return !strings.Contains(s, "_") && !strings.Contains(s, "-")
}

// isUpperCamelCase checks if s is valid UpperCamelCase.
// e.g., MyFunction, AnotherOne.
// Simple initial check: starts with uppercase, no underscores/hyphens.
func isUpperCamelCase(s string) bool {
	if len(s) == 0 {
		return false // Empty string is not valid UpperCamelCase
	}
	firstChar := rune(s[0])
	if !(unicode.IsUpper(firstChar) && unicode.IsLetter(firstChar)) {
		return false
	}
	// Allow digits after the first character, but no underscores or hyphens.
	return !strings.Contains(s, "_") && !strings.Contains(s, "-")
}

// --- Rule Implementations ---

// FunctionNameRule (L3)
// Ensures 'fn' are lowerCamelCase and 'pub fn' are UpperCamelCase.
type FunctionNameRule struct{}

func (r *FunctionNameRule) Name() string {
	return "function-naming-convention"
}

func (r *FunctionNameRule) Description() string {
	return "Ensures private functions ('fn') are lowerCamelCase and public functions ('pub fn') are UpperCamelCase."
}

func (r *FunctionNameRule) Check(node ast.Node, program *ast.Program) []Issue {
	issues := []Issue{}
	fnDef, ok := node.(*ast.FunctionDefinition)
	if !ok {
		return issues // Not a function definition, skip
	}

	// Skip "main" function from this rule, as it's a special case.
	if fnDef.Name == "main" {
		return issues
	}

	if fnDef.IsPublic {
		if !isUpperCamelCase(fnDef.Name) {
			issues = append(issues, Issue{
				// Filepath will be set by the Linter's visitor
				Line:     0, // Placeholder - AST nodes need line/col info
				Column:   0, // Placeholder
				RuleName: r.Name(),
				Message:  fmt.Sprintf("Public function '%s' should be in UpperCamelCase (e.g., MyFunction).", fnDef.Name),
			})
		}
	} else { // Private function
		if !isLowerCamelCase(fnDef.Name) {
			issues = append(issues, Issue{
				Line:     0, // Placeholder
				Column:   0, // Placeholder
				RuleName: r.Name(),
				Message:  fmt.Sprintf("Private function '%s' should be in lowerCamelCase (e.g., myFunction).", fnDef.Name),
			})
		}
	}
	return issues
}

// VariableNameRule (L4)
// Ensures 'let' declared variables are lowerCamelCase.
type VariableNameRule struct{}

func (r *VariableNameRule) Name() string {
	return "variable-naming-convention"
}

func (r *VariableNameRule) Description() string {
	return "Ensures 'let' declared variables are in lowerCamelCase."
}

func (r *VariableNameRule) Check(node ast.Node, program *ast.Program) []Issue {
	issues := []Issue{}
	letDecl, ok := node.(*ast.LetDeclaration)
	if !ok {
		return issues // Not a let declaration, skip
	}

	if letDecl.Name == "_" { // Conventionally ignored variables
		return issues
	}

	if !isLowerCamelCase(letDecl.Name) {
		issues = append(issues, Issue{
			Line:     0, // Placeholder
			Column:   0, // Placeholder
			RuleName: r.Name(),
			Message:  fmt.Sprintf("Variable '%s' should be in lowerCamelCase (e.g., myVariable).", letDecl.Name),
		})
	}
	return issues
}
