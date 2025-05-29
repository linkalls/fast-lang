package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/linkalls/zeno-lang/ast"
)

// GenerationError represents errors during code generation
type GenerationError struct {
	Message string
}

func (e GenerationError) Error() string {
	return "Generation Error: " + e.Message
}

// Generate generates Go code from the AST
func Generate(program *ast.Program) (string, error) {
	var builder strings.Builder

	// Generate package and imports
	builder.WriteString("package main\n\n")
	builder.WriteString("import (\n")
	builder.WriteString("\t\"fmt\"\n")
	builder.WriteString(")\n\n")
	builder.WriteString("func main() {\n")

	// Generate statements
	for _, stmt := range program.Statements {
		if err := generateStatement(stmt, &builder, 1); err != nil {
			return "", err
		}
	}

	builder.WriteString("}\n")
	return builder.String(), nil
}

// Helper function for indentation
func indent(level int) string {
	return strings.Repeat("\t", level)
}

// Map Zeno type strings to Go type strings
func mapType(zenoType string) string {
	switch zenoType {
	case "int":
		return "int64"
	case "float":
		return "float64"
	case "bool":
		return "bool"
	case "string":
		return "string"
	default:
		// If not a known type, assume it's already a valid Go type
		return zenoType
	}
}

// generateStatement generates Go code for a statement
func generateStatement(stmt ast.Statement, builder *strings.Builder, indentLevel int) error {
	builder.WriteString(indent(indentLevel))

	switch s := stmt.(type) {
	case *ast.LetDeclaration:
		builder.WriteString("var ")
		builder.WriteString(s.Name)
		if s.TypeAnn != nil {
			builder.WriteString(" ")
			builder.WriteString(mapType(*s.TypeAnn))
		}
		builder.WriteString(" = ")
		if err := generateExpression(s.ValueExpression, builder); err != nil {
			return err
		}
		builder.WriteString("\n")

	case *ast.ExpressionStatement:
		if err := generateExpression(s.Expression, builder); err != nil {
			return err
		}
		builder.WriteString("\n")

	default:
		return GenerationError{Message: fmt.Sprintf("Unsupported statement type: %T", stmt)}
	}

	return nil
}

// generateExpression generates Go code for an expression
func generateExpression(expr ast.Expression, builder *strings.Builder) error {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		builder.WriteString(strconv.FormatInt(e.Value, 10))

	case *ast.StringLiteral:
		// Escape the string for Go
		escaped := strconv.Quote(e.Value)
		builder.WriteString(escaped)

	case *ast.Identifier:
		builder.WriteString(e.Value)

	case *ast.BinaryExpression:
		builder.WriteString("(")
		if err := generateExpression(e.Left, builder); err != nil {
			return err
		}
		builder.WriteString(" ")
		builder.WriteString(e.Operator.String())
		builder.WriteString(" ")
		if err := generateExpression(e.Right, builder); err != nil {
			return err
		}
		builder.WriteString(")")

	default:
		return GenerationError{Message: fmt.Sprintf("Unsupported expression type: %T", expr)}
	}

	return nil
}
