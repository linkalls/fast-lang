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

// Generator manages code generation with scope and import tracking
type Generator struct {
	imports      map[string][]string          // module -> imported identifiers
	declaredVars map[string]bool              // variable name -> declared
	usedVars     map[string]bool              // variable name -> used
	standardLibs map[string]map[string]string // module -> function -> go equivalent
	showJapanese bool                         // whether to show Japanese error messages
}

// NewGenerator creates a new generator instance
func NewGenerator() *Generator {
	g := &Generator{
		imports:      make(map[string][]string),
		declaredVars: make(map[string]bool),
		usedVars:     make(map[string]bool),
		standardLibs: make(map[string]map[string]string),
	}

	// Define standard library mappings
	g.standardLibs["std/fmt"] = map[string]string{
		"print":   "fmt.Print",
		"println": "fmt.Println",
	}

	return g
}

// Generate generates Go code from the AST
func Generate(program *ast.Program) (string, error) {
	return GenerateWithOptions(program, false)
}

// GenerateWithOptions generates Go code from the AST with options
func GenerateWithOptions(program *ast.Program, showJapanese bool) (string, error) {
	g := NewGenerator()
	g.showJapanese = showJapanese
	return g.generateProgram(program)
}

// generateProgram generates Go code for the entire program
func (g *Generator) generateProgram(program *ast.Program) (string, error) {
	var builder strings.Builder

	// First pass: collect imports and declarations
	for _, stmt := range program.Statements {
		if err := g.collectImportsAndDeclarations(stmt); err != nil {
			return "", err
		}
	}

	// Generate package and imports
	builder.WriteString("package main\n\n")
	builder.WriteString("import (\n")
	builder.WriteString("\t\"fmt\"\n")
	builder.WriteString(")\n\n")

	// Separate function definitions from other statements
	var functionDefs []*ast.FunctionDefinition
	var otherStmts []ast.Statement
	var mainFunc *ast.FunctionDefinition

	for _, stmt := range program.Statements {
		if funcDef, ok := stmt.(*ast.FunctionDefinition); ok {
			if funcDef.Name == "main" {
				mainFunc = funcDef
			} else {
				functionDefs = append(functionDefs, funcDef)
			}
		} else if _, ok := stmt.(*ast.ImportStatement); !ok {
			// Skip import statements as they're handled above
			otherStmts = append(otherStmts, stmt)
		}
	}

	// Generate function definitions at top level
	for _, funcDef := range functionDefs {
		if err := g.generateStatement(funcDef, &builder, 0); err != nil {
			return "", err
		}
		builder.WriteString("\n")
	}

	// Generate main function
	if mainFunc != nil {
		// User defined main function
		builder.WriteString("func main() {\n")
		for _, bodyStmt := range mainFunc.Body {
			if err := g.generateStatement(bodyStmt, &builder, 1); err != nil {
				return "", err
			}
		}
		builder.WriteString("}\n")
	} else if len(otherStmts) > 0 {
		// No user-defined main, but there are other statements - wrap them in main
		builder.WriteString("func main() {\n")
		for _, stmt := range otherStmts {
			if err := g.generateStatement(stmt, &builder, 1); err != nil {
				return "", err
			}
		}
		builder.WriteString("}\n")
	} else {
		// No main function and no statements - generate empty main
		builder.WriteString("func main() {\n")
		builder.WriteString("}\n")
	}

	// Check for unused variables
	if err := g.checkUnusedVariables(); err != nil {
		return "", err
	}

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
func (g *Generator) generateStatement(stmt ast.Statement, builder *strings.Builder, indentLevel int) error {
	switch s := stmt.(type) {
	case *ast.ImportStatement:
		// Import statements are handled in the header, skip them here
		return nil

	case *ast.LetDeclaration:
		builder.WriteString(indent(indentLevel))
		builder.WriteString("var ")
		builder.WriteString(s.Name)

		if s.TypeAnn != nil {
			builder.WriteString(" ")
			builder.WriteString(mapType(*s.TypeAnn))
		}

		builder.WriteString(" = ")
		if err := g.generateExpression(s.ValueExpression, builder); err != nil {
			return err
		}
		builder.WriteString("\n")

	case *ast.FunctionDefinition:
		builder.WriteString(indent(indentLevel))
		builder.WriteString("func ")
		builder.WriteString(s.Name)
		builder.WriteString("(")

		// Generate parameters
		for i, param := range s.Parameters {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(param.Name)
			builder.WriteString(" ")
			builder.WriteString(mapType(param.Type))
		}

		builder.WriteString(")")

		// Generate return type if present
		if s.ReturnType != nil {
			builder.WriteString(" ")
			builder.WriteString(mapType(*s.ReturnType))
		}

		builder.WriteString(" {\n")

		// Generate function body
		for _, bodyStmt := range s.Body {
			if err := g.generateStatement(bodyStmt, builder, indentLevel+1); err != nil {
				return err
			}
		}

		builder.WriteString(indent(indentLevel))
		builder.WriteString("}\n")

	case *ast.ReturnStatement:
		builder.WriteString(indent(indentLevel))
		builder.WriteString("return")
		if s.Value != nil {
			builder.WriteString(" ")
			if err := g.generateExpression(s.Value, builder); err != nil {
				return err
			}
		}
		builder.WriteString("\n")

	case *ast.ExpressionStatement:
		builder.WriteString(indent(indentLevel))
		if err := g.generateExpression(s.Expression, builder); err != nil {
			return err
		}
		builder.WriteString("\n")

	case *ast.PrintStatement:
		builder.WriteString(indent(indentLevel))
		if s.Newline {
			// Validate that println is imported
			if err := g.validateImports("println"); err != nil {
				return err
			}
			builder.WriteString("fmt.Println(")
		} else {
			// Validate that print is imported
			if err := g.validateImports("print"); err != nil {
				return err
			}
			builder.WriteString("fmt.Print(")
		}

		for i, arg := range s.Arguments {
			if i > 0 {
				builder.WriteString(", ")
			}
			if err := g.generateExpression(arg, builder); err != nil {
				return err
			}
		}

		builder.WriteString(")\n")

	default:
		return GenerationError{Message: fmt.Sprintf("Unsupported statement type: %T", stmt)}
	}

	return nil
}

// generateExpression generates Go code for an expression
func (g *Generator) generateExpression(expr ast.Expression, builder *strings.Builder) error {
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
		if err := g.generateExpression(e.Left, builder); err != nil {
			return err
		}
		builder.WriteString(" ")
		builder.WriteString(e.Operator.String())
		builder.WriteString(" ")
		if err := g.generateExpression(e.Right, builder); err != nil {
			return err
		}
		builder.WriteString(")")

	case *ast.FunctionCall:
		builder.WriteString(e.Name)
		builder.WriteString("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				builder.WriteString(", ")
			}
			if err := g.generateExpression(arg, builder); err != nil {
				return err
			}
		}
		builder.WriteString(")")

	default:
		return GenerationError{Message: fmt.Sprintf("Unsupported expression type: %T", expr)}
	}

	return nil
}

// collectImportsAndDeclarations performs first pass to collect imports and variable declarations
func (g *Generator) collectImportsAndDeclarations(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.ImportStatement:
		g.imports[s.Module] = s.Imports
	case *ast.LetDeclaration:
		g.declaredVars[s.Name] = true
		// Also check for variable usage in the value expression
		if s.ValueExpression != nil {
			g.markVariableUsage(s.ValueExpression)
		}
	case *ast.FunctionDefinition:
		// Functions don't need variable tracking for now
		// But we might want to track function bodies in the future
		for _, bodyStmt := range s.Body {
			g.collectImportsAndDeclarations(bodyStmt)
		}
	case *ast.ReturnStatement:
		if s.Value != nil {
			g.markVariableUsage(s.Value)
		}
	case *ast.ExpressionStatement:
		// Check for variable usage in expressions
		g.markVariableUsage(s.Expression)
	case *ast.PrintStatement:
		// Mark variables used in print statements
		for _, arg := range s.Arguments {
			g.markVariableUsage(arg)
		}
	}
	return nil
}

// markVariableUsage marks variables as used when referenced in expressions
func (g *Generator) markVariableUsage(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		g.usedVars[e.Value] = true
	case *ast.BinaryExpression:
		g.markVariableUsage(e.Left)
		g.markVariableUsage(e.Right)
	case *ast.UnaryExpression:
		g.markVariableUsage(e.Right)
	case *ast.FunctionCall:
		// Mark variables used in function arguments
		for _, arg := range e.Arguments {
			g.markVariableUsage(arg)
		}
	}
}

// checkUnusedVariables returns an error if there are unused variables
func (g *Generator) checkUnusedVariables() error {
	var unusedVars []string
	for varName := range g.declaredVars {
		if !g.usedVars[varName] {
			unusedVars = append(unusedVars, varName)
		}
	}

	if len(unusedVars) > 0 {
		msg := fmt.Sprintf("Unused variables found: %s", strings.Join(unusedVars, ", "))
		if g.showJapanese {
			msg += fmt.Sprintf(" / 未使用の変数があります: %s", strings.Join(unusedVars, ", "))
		}
		return GenerationError{Message: msg}
	}

	return nil
}

// validateImports checks that all used functions are properly imported
func (g *Generator) validateImports(functionName string) error {
	// Check if function is a standard library function
	for module, functions := range g.standardLibs {
		if _, exists := functions[functionName]; exists {
			// Check if this function is imported from the correct module
			if importedFuncs, imported := g.imports[module]; imported {
				for _, importedFunc := range importedFuncs {
					if importedFunc == functionName {
						return nil // Function is properly imported
					}
				}
			}
			msg := fmt.Sprintf("Function '%s' is not imported from '%s'", functionName, module)
			if g.showJapanese {
				msg += fmt.Sprintf(" / 関数 '%s' は '%s' からimportされていません", functionName, module)
			}
			return GenerationError{Message: msg}
		}
	}
	return nil
}
