package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/linkalls/zeno-lang/ast"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
	"github.com/linkalls/zeno-lang/types"
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
	declaredFns  map[string]string            // function name -> go function name
	usedFns      map[string]bool              // function name -> used
	userModules  map[string]map[string]string // user module -> function -> go equivalent
	moduleASTs   map[string]*ast.Program      // user module -> parsed AST
	standardLibs map[string]map[string]string // module -> function -> go equivalent
	currentDir   string                       // directory of the current file being compiled
	symbolTable  *types.SymbolTable           // symbol table for type inference
}

// NewGenerator creates a new generator instance
func NewGenerator() *Generator {
	g := &Generator{
		imports:      make(map[string][]string),
		declaredVars: make(map[string]bool),
		usedVars:     make(map[string]bool),
		declaredFns:  make(map[string]string),
		usedFns:      make(map[string]bool),
		userModules:  make(map[string]map[string]string),
		moduleASTs:   make(map[string]*ast.Program),
		standardLibs: make(map[string]map[string]string),
		symbolTable:  types.NewSymbolTable(nil),
	}

	// Define standard library mappings
	g.standardLibs["std/fmt"] = map[string]string{
		// Note: print and println are built-in functions, not imported
	}

	g.standardLibs["std/io"] = map[string]string{
		"readFile":  "readFile",
		"writeFile": "writeFile",
	}

	return g
}

// Generate generates Go code from the AST
func Generate(program *ast.Program) (string, error) {
	return GenerateWithOptions(program)
}

// GenerateWithOptions generates Go code from the AST with options
func GenerateWithOptions(program *ast.Program) (string, error) {
	return GenerateWithFile(program, "")
}

// GenerateWithFile generates Go code from the AST with source file path for module resolution
func GenerateWithFile(program *ast.Program, sourceFile string) (string, error) {
	g := NewGenerator()
	g.currentDir = sourceFile
	return g.generateProgram(program)
}

// generateProgram generates Go code for the entire program
func (g *Generator) generateProgram(program *ast.Program) (string, error) {
	var builder strings.Builder

	// Validate function types first
	if err := g.validateFunctionTypes(program); err != nil {
		return "", err
	}

	// First pass: collect imports and declarations
	for _, stmt := range program.Statements {
		if err := g.collectImportsAndDeclarations(stmt); err != nil {
			return "", err
		}
	}

	// Generate package and imports
	builder.WriteString("package main\n\n")
	builder.WriteString("import (\n")

	// Add required imports based on used standard library functions
	requiredImports := make(map[string]bool)

	// Check which standard library modules are imported
	for module := range g.imports {
		if strings.HasPrefix(module, "std/") {
			if module == "std/fmt" {
				requiredImports["fmt"] = true
			} else if module == "std/io" {
				requiredImports["os"] = true
			}
		}
	}

	// Always include fmt for basic functionality
	requiredImports["fmt"] = true

	// Generate import statements
	for imp := range requiredImports {
		builder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
	}

	builder.WriteString(")\n\n")

	// Generate std/io helper functions if needed
	if _, hasStdIo := g.imports["std/io"]; hasStdIo {
		g.generateStdIoHelpers(&builder)
	}

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

	// Generate imported user module functions first
	for modulePath, moduleAST := range g.moduleASTs {
		if importedFuncs, exists := g.imports[modulePath]; exists {
			for _, stmt := range moduleAST.Statements {
				if funcDef, ok := stmt.(*ast.FunctionDefinition); ok && funcDef.IsPublic {
					// Check if this function is actually imported
					for _, importedFunc := range importedFuncs {
						if importedFunc == funcDef.Name {
							if err := g.generateStatement(funcDef, &builder, 0); err != nil {
								return "", err
							}
							builder.WriteString("\n")
							break
						}
					}
				}
			}
		}
	}

	// Generate function definitions at top level
	for _, funcDef := range functionDefs {
		if err := g.generateStatement(funcDef, &builder, 0); err != nil {
			return "", err
		}
		builder.WriteString("\n")
	}

	// Always generate main function
	builder.WriteString("func main() {\n")
	if mainFunc != nil {
		// User defined main function - use its body
		for _, bodyStmt := range mainFunc.Body {
			if err := g.generateStatement(bodyStmt, &builder, 1); err != nil {
				return "", err
			}
		}
	} else if len(otherStmts) > 0 {
		// No user-defined main, but there are other statements - wrap them in main
		for _, stmt := range otherStmts {
			if err := g.generateStatement(stmt, &builder, 1); err != nil {
				return "", err
			}
		}
	}
	// Always close main function
	builder.WriteString("}\n")

	// Check for unused variables
	if err := g.checkUnusedVariables(); err != nil {
		return "", err
	}

	// Check for unused functions
	if err := g.checkUnusedFunctions(); err != nil {
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
		return "int"
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
		// Register variable type in symbol table
		var varType types.Type
		if s.TypeAnn != nil {
			// Use explicit type annotation
			varType = g.mapASTTypeToType(*s.TypeAnn)
		} else {
			// Infer type from value expression
			varType = g.inferType(s.ValueExpression)
		}
		g.registerVariableWithType(s.Name, varType)

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

	case *ast.AssignmentStatement:
		// Mark variable as used during assignment
		g.usedVars[s.Name] = true
		// Mark variables used in the assignment value
		g.markVariableUsage(s.Value)

		builder.WriteString(indent(indentLevel))
		builder.WriteString(s.Name)
		builder.WriteString(" = ")
		if err := g.generateExpression(s.Value, builder); err != nil {
			return err
		}
		builder.WriteString("\n")

	case *ast.FunctionDefinition:
		builder.WriteString(indent(indentLevel))
		builder.WriteString("func ")

		// Handle function visibility - in Go, public functions start with uppercase
		functionName := s.Name
		if s.IsPublic {
			// Make first letter uppercase for public functions
			if len(functionName) > 0 {
				functionName = strings.ToUpper(string(functionName[0])) + functionName[1:]
			}
		} else {
			// Make first letter lowercase for private functions (unless it's main)
			if functionName != "main" && len(functionName) > 0 {
				functionName = strings.ToLower(string(functionName[0])) + functionName[1:]
			}
		}

		builder.WriteString(functionName)
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

	case *ast.IfStatement:
		builder.WriteString(indent(indentLevel))
		builder.WriteString("if ")
		if err := g.generateCondition(s.Condition, builder); err != nil {
			return err
		}
		builder.WriteString(" ")
		if err := g.generateBlock(s.ThenBlock, builder, indentLevel); err != nil {
			return err
		}

		// Generate else if clauses
		for _, elseIf := range s.ElseIfClauses {
			builder.WriteString(" else if ")
			if err := g.generateCondition(elseIf.Condition, builder); err != nil {
				return err
			}
			builder.WriteString(" ")
			if err := g.generateBlock(elseIf.Block, builder, indentLevel); err != nil {
				return err
			}
		}

		// Generate else clause if present
		if s.ElseBlock != nil {
			builder.WriteString(" else ")
			if err := g.generateBlock(s.ElseBlock, builder, indentLevel); err != nil {
				return err
			}
		}

		builder.WriteString("\n")

	case *ast.WhileStatement:
		builder.WriteString(indent(indentLevel))
		builder.WriteString("for ")
		if err := g.generateCondition(s.Condition, builder); err != nil {
			return err
		}
		builder.WriteString(" ")
		if err := g.generateBlock(s.Block, builder, indentLevel); err != nil {
			return err
		}
		builder.WriteString("\n")

	default:
		return GenerationError{Message: fmt.Sprintf("Unsupported statement type: %T", stmt)}
	}

	return nil
}

// generateExpression generates Go code for an expression
func (g *Generator) generateExpression(expr ast.Expression, builder *strings.Builder) error {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		builder.WriteString(strconv.Itoa(e.Value))

	case *ast.StringLiteral:
		// Escape the string for Go
		escaped := strconv.Quote(e.Value)
		builder.WriteString(escaped)

	case *ast.BooleanLiteral:
		if e.Value {
			builder.WriteString("true")
		} else {
			builder.WriteString("false")
		}

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
		// Validate that the function is properly imported if it's a standard library function
		if err := g.validateImports(e.Name); err != nil {
			return err
		}

		// First check if this is a function defined in the current file
		functionName := e.Name
		if goFuncName, exists := g.declaredFns[functionName]; exists {
			functionName = goFuncName
		} else {
			// Check if this is a function from a user-defined module
			for module, functions := range g.userModules {
				if goFuncName, exists := functions[functionName]; exists {
					// Check if this function is imported from this module
					if importedFuncs, imported := g.imports[module]; imported {
						for _, importedFunc := range importedFuncs {
							if importedFunc == functionName {
								functionName = goFuncName // Use the Go-style function name
								break
							}
						}
					}
					break
				}
			}

			// Check if this is a standard library function
			for module, functions := range g.standardLibs {
				if goFuncName, exists := functions[functionName]; exists {
					// Check if this function is imported from this module
					if importedFuncs, imported := g.imports[module]; imported {
						for _, importedFunc := range importedFuncs {
							if importedFunc == functionName {
								functionName = goFuncName // Use the Go standard library mapping
								break
							}
						}
					}
					break
				}
			}
		}

		builder.WriteString(functionName)
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

// generateCondition generates Go code for a condition expression, using type information for smart boolean conversion
func (g *Generator) generateCondition(expr ast.Expression, builder *strings.Builder) error {
	switch e := expr.(type) {
	case *ast.BooleanLiteral:
		// Already boolean, generate as-is
		return g.generateExpression(expr, builder)
	case *ast.BinaryExpression:
		// Binary expressions like comparisons are already boolean
		return g.generateExpression(expr, builder)
	case *ast.IntegerLiteral:
		// Convert integer literals to boolean: value != 0
		builder.WriteString("(")
		if err := g.generateExpression(expr, builder); err != nil {
			return err
		}
		builder.WriteString(" != 0)")
		return nil
	case *ast.Identifier:
		// Use type information to generate appropriate boolean conversion
		varType := g.getVariableType(e.Value)

		// Debug output
		fmt.Printf("DEBUG: Variable %s has type %v\n", e.Value, varType)

		switch varType {
		case types.BoolType:
			// Boolean variables can be used directly
			return g.generateExpression(expr, builder)
		case types.IntType:
			// Integer variables need != 0 conversion
			builder.WriteString("(")
			if err := g.generateExpression(expr, builder); err != nil {
				return err
			}
			builder.WriteString(" != 0)")
			return nil
		case types.StringType:
			// String variables need != "" conversion
			builder.WriteString("(")
			if err := g.generateExpression(expr, builder); err != nil {
				return err
			}
			builder.WriteString(" != \"\")")
			return nil
		case types.FloatType:
			// Float variables need != 0.0 conversion
			builder.WriteString("(")
			if err := g.generateExpression(expr, builder); err != nil {
				return err
			}
			builder.WriteString(" != 0.0)")
			return nil
		default:
			// Unknown type - generate as-is and let Go handle it
			return g.generateExpression(expr, builder)
		}
	default:
		// For other expressions, generate as-is first
		// Let Go's type system handle the boolean conversion
		return g.generateExpression(expr, builder)
	}
}

// generateBlock generates Go code for a block of statements
func (g *Generator) generateBlock(block *ast.Block, builder *strings.Builder, indentLevel int) error {
	if block == nil {
		builder.WriteString("{}\n")
		return nil
	}

	builder.WriteString("{\n")
	for _, stmt := range block.Statements {
		if err := g.generateStatement(stmt, builder, indentLevel+1); err != nil {
			return err
		}
	}
	builder.WriteString(indent(indentLevel))
	builder.WriteString("}")
	return nil
}

// collectImportsAndDeclarations performs first pass to collect imports and variable declarations
func (g *Generator) collectImportsAndDeclarations(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.ImportStatement:
		g.imports[s.Module] = s.Imports
		// If it's a user module (starts with "./" or "../"), process it
		if strings.HasPrefix(s.Module, "./") || strings.HasPrefix(s.Module, "../") {
			if err := g.processUserModule(s.Module, s.Imports); err != nil {
				return err
			}
		}
	case *ast.LetDeclaration:
		g.declaredVars[s.Name] = true
		// Register variable type in symbol table
		var varType types.Type
		if s.TypeAnn != nil {
			// Use explicit type annotation
			varType = g.mapASTTypeToType(*s.TypeAnn)
		} else {
			// Infer type from value expression
			varType = g.inferType(s.ValueExpression)
		}
		g.registerVariableWithType(s.Name, varType)
		// Also check for variable usage in the value expression
		if s.ValueExpression != nil {
			g.markVariableUsage(s.ValueExpression)
		}
	case *ast.AssignmentStatement:
		// Mark variable as used during assignment
		g.usedVars[s.Name] = true
		// Mark variables used in the assignment value
		g.markVariableUsage(s.Value)
	case *ast.FunctionDefinition:
		// Track function declaration with Go naming convention
		goFuncName := s.Name
		if s.IsPublic {
			// Public functions start with uppercase
			goFuncName = strings.ToUpper(s.Name[:1]) + s.Name[1:]
		}
		g.declaredFns[s.Name] = goFuncName
		// Process function body
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
	case *ast.IfStatement:
		// Mark variables used in if condition
		g.markVariableUsage(s.Condition)
		// Process then block
		if s.ThenBlock != nil {
			g.markBlockUsage(s.ThenBlock)
		}
		// Process else if clauses
		for _, elseIf := range s.ElseIfClauses {
			g.markVariableUsage(elseIf.Condition)
			if elseIf.Block != nil {
				g.markBlockUsage(elseIf.Block)
			}
		}
		// Process else block
		if s.ElseBlock != nil {
			g.markBlockUsage(s.ElseBlock)
		}
	case *ast.WhileStatement:
		// Mark variables used in while condition
		g.markVariableUsage(s.Condition)
		// Process while block
		if s.Block != nil {
			g.markBlockUsage(s.Block)
		}
	}
	return nil
}

// markVariableUsage marks variables as used when referenced in expressions
func (g *Generator) markVariableUsage(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		g.usedVars[e.Value] = true
	case *ast.BooleanLiteral:
		// Boolean literals don't reference variables, nothing to do
	case *ast.IntegerLiteral:
		// Integer literals don't reference variables, nothing to do
	case *ast.StringLiteral:
		// String literals don't reference variables, nothing to do
	case *ast.BinaryExpression:
		g.markVariableUsage(e.Left)
		g.markVariableUsage(e.Right)
	case *ast.UnaryExpression:
		g.markVariableUsage(e.Right)
	case *ast.FunctionCall:
		// Mark function as used
		g.usedFns[e.Name] = true
		// Mark variables used in function arguments
		for _, arg := range e.Arguments {
			g.markVariableUsage(arg)
		}
	}
}

// markBlockUsage marks variables as used when referenced in a block of statements
func (g *Generator) markBlockUsage(block *ast.Block) {
	if block == nil {
		return
	}
	for _, stmt := range block.Statements {
		g.collectImportsAndDeclarations(stmt)
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
		return GenerationError{Message: msg}
	}

	return nil
}

// checkUnusedFunctions returns an error if there are unused functions
func (g *Generator) checkUnusedFunctions() error {
	var unusedFns []string
	for fnName := range g.declaredFns {
		// Skip main function as it's the entry point
		if fnName == "main" {
			continue
		}
		// Skip public functions as they might be used externally
		if g.isPublicFunction(fnName) {
			continue
		}
		if !g.usedFns[fnName] {
			unusedFns = append(unusedFns, fnName)
		}
	}

	if len(unusedFns) > 0 {
		msg := fmt.Sprintf("Unused functions found: %s", strings.Join(unusedFns, ", "))
		return GenerationError{Message: msg}
	}

	return nil
}

// isPublicFunction checks if a function name indicates it's public (Go function name starts with uppercase)
func (g *Generator) isPublicFunction(fnName string) bool {
	if goFuncName, exists := g.declaredFns[fnName]; exists {
		if len(goFuncName) == 0 {
			return false
		}
		return goFuncName[0] >= 'A' && goFuncName[0] <= 'Z'
	}
	return false
}

// validateImports checks that all used functions are properly imported
func (g *Generator) validateImports(functionName string) error {
	// Built-in functions don't need to be imported
	builtinFunctions := map[string]bool{
		"print":   true,
		"println": true,
	}

	if builtinFunctions[functionName] {
		return nil
	}

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
			return GenerationError{Message: msg}
		}
	}

	// Check if function is from a user-defined module
	for module, functions := range g.userModules {
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
			return GenerationError{Message: msg}
		}
	}

	return nil
}

// processUserModule processes a user-defined module and extracts public functions
func (g *Generator) processUserModule(modulePath string, importedFunctions []string) error {
	// Convert relative path to absolute path
	var zenoFilePath string
	if strings.HasSuffix(modulePath, ".zeno") {
		zenoFilePath = modulePath
	} else {
		zenoFilePath = modulePath + ".zeno"
	}

	// If it's a relative path and we have a current directory, resolve it
	if (strings.HasPrefix(zenoFilePath, "./") || strings.HasPrefix(zenoFilePath, "../")) && g.currentDir != "" {
		// Get the directory of the current source file
		baseDir := filepath.Dir(g.currentDir)
		// Resolve the relative path
		zenoFilePath = filepath.Join(baseDir, zenoFilePath)
	}

	// Read the module file
	content, err := os.ReadFile(zenoFilePath)
	if err != nil {
		msg := fmt.Sprintf("Failed to read module file '%s': %v", zenoFilePath, err)
		return GenerationError{Message: msg}
	}

	// Parse the module
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		msg := fmt.Sprintf("Parse errors in module '%s': %v", zenoFilePath, p.Errors())
		return GenerationError{Message: msg}
	}

	// Extract public functions from the module
	publicFunctions := make(map[string]string)
	for _, stmt := range program.Statements {
		if funcDef, ok := stmt.(*ast.FunctionDefinition); ok && funcDef.IsPublic {
			// Convert function name to Go convention (uppercase first letter)
			goFuncName := funcDef.Name
			if len(goFuncName) > 0 {
				goFuncName = strings.ToUpper(string(goFuncName[0])) + goFuncName[1:]
			}
			publicFunctions[funcDef.Name] = goFuncName
		}
	}

	// Validate that all imported functions exist and are public
	for _, importedFunc := range importedFunctions {
		if _, exists := publicFunctions[importedFunc]; !exists {
			msg := fmt.Sprintf("Function '%s' is not exported from module '%s'", importedFunc, modulePath)
			return GenerationError{Message: msg}
		}
	}

	// Store the module mapping and AST
	g.userModules[modulePath] = publicFunctions
	g.moduleASTs[modulePath] = program
	return nil
}

// generateStdIoHelpers generates helper functions for std/io module
func (g *Generator) generateStdIoHelpers(builder *strings.Builder) {
	// Generate readFile helper function
	builder.WriteString("// std/io helper functions\n")
	builder.WriteString("func readFile(filename string) string {\n")
	builder.WriteString("\tdata, err := os.ReadFile(filename)\n")
	builder.WriteString("\tif err != nil {\n")
	builder.WriteString("\t\tfmt.Printf(\"Error reading file %s: %v\\n\", filename, err)\n")
	builder.WriteString("\t\treturn \"\"\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\treturn string(data)\n")
	builder.WriteString("}\n\n")

	// Generate writeFile helper function
	builder.WriteString("func writeFile(filename string, content string) {\n")
	builder.WriteString("\terr := os.WriteFile(filename, []byte(content), 0644)\n")
	builder.WriteString("\tif err != nil {\n")
	builder.WriteString("\t\tfmt.Printf(\"Error writing file %s: %v\\n\", filename, err)\n")
	builder.WriteString("\t}\n")
	builder.WriteString("}\n\n")
}

// inferType infers the type of an expression
func (g *Generator) inferType(expr ast.Expression) types.Type {
	switch e := expr.(type) {
	case *ast.BooleanLiteral:
		return types.BoolType
	case *ast.IntegerLiteral:
		return types.IntType
	case *ast.StringLiteral:
		return types.StringType
	case *ast.Identifier:
		// Look up the identifier in the symbol table
		if symbol, ok := g.symbolTable.Resolve(e.Value); ok {
			return symbol.Type
		}
		// Default to int if not found (for compatibility)
		return types.IntType
	case *ast.BinaryExpression:
		// Comparison operators return boolean
		switch e.Operator {
		case ast.BinaryOpEq, ast.BinaryOpNotEq, ast.BinaryOpLt, ast.BinaryOpLte, ast.BinaryOpGt, ast.BinaryOpGte:
			return types.BoolType
		case ast.BinaryOpPlus, ast.BinaryOpMinus, ast.BinaryOpMultiply, ast.BinaryOpDivide, ast.BinaryOpModulo:
			// Arithmetic operations: determine result type based on operands
			leftType := g.inferType(e.Left)
			rightType := g.inferType(e.Right)
			// If either operand is float, result is float
			if leftType == types.FloatType || rightType == types.FloatType {
				return types.FloatType
			}
			return types.IntType
		case ast.BinaryOpAnd, ast.BinaryOpOr:
			return types.BoolType
		}
	case *ast.FunctionCall:
		// For function calls, we'd need to track function return types
		// For now, default to int
		return types.IntType
	}
	// Default fallback
	return types.IntType
}

// registerVariableWithType registers a variable with a specific type in the symbol table
func (g *Generator) registerVariableWithType(name string, varType types.Type) {
	g.symbolTable.Define(name, varType)
}

// getVariableType gets the type of a variable from the symbol table
func (g *Generator) getVariableType(name string) types.Type {
	if symbol, ok := g.symbolTable.Resolve(name); ok {
		fmt.Printf("DEBUG: Found variable %s with type %v in symbol table\n", name, symbol.Type)
		return symbol.Type
	}
	// Default to int if not found
	fmt.Printf("DEBUG: Variable %s not found in symbol table, defaulting to IntType\n", name)
	return types.IntType
}

// mapASTTypeToType converts AST type annotations to type system types
func (g *Generator) mapASTTypeToType(astType string) types.Type {
	switch astType {
	case "bool":
		return types.BoolType
	case "int":
		return types.IntType
	case "string":
		return types.StringType
	case "float":
		return types.FloatType
	default:
		// Default to int for unknown types
		return types.IntType
	}
}

// validateFunctionTypes validates that non-main functions have explicit return types and parameter types
func (g *Generator) validateFunctionTypes(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if funcDef, ok := stmt.(*ast.FunctionDefinition); ok {
			// Skip main function - it doesn't need explicit return type
			if funcDef.Name == "main" {
				continue
			}

			// Check that all parameters have explicit types
			for _, param := range funcDef.Parameters {
				if param.Type == "" {
					return GenerationError{Message: fmt.Sprintf("Function '%s': parameter '%s' must have an explicit type", funcDef.Name, param.Name)}
				}
			}

			// Only require explicit return type if the function has return statements
			if g.hasReturnStatement(funcDef.Body) && funcDef.ReturnType == nil {
				return GenerationError{Message: fmt.Sprintf("Function '%s' contains return statements but has no explicit return type", funcDef.Name)}
			}
		}
	}
	return nil
}

// hasReturnStatement checks if a function body contains any return statements
func (g *Generator) hasReturnStatement(statements []ast.Statement) bool {
	for _, stmt := range statements {
		if _, ok := stmt.(*ast.ReturnStatement); ok {
			return true
		}
		// Check for return statements in nested blocks (if, while, etc.)
		if ifStmt, ok := stmt.(*ast.IfStatement); ok {
			if g.hasReturnStatement(ifStmt.ThenBlock.Statements) {
				return true
			}
			for _, elseIfClause := range ifStmt.ElseIfClauses {
				if g.hasReturnStatement(elseIfClause.Block.Statements) {
					return true
				}
			}
			if ifStmt.ElseBlock != nil && g.hasReturnStatement(ifStmt.ElseBlock.Statements) {
				return true
			}
		}
		if whileStmt, ok := stmt.(*ast.WhileStatement); ok {
			if g.hasReturnStatement(whileStmt.Block.Statements) {
				return true
			}
		}
	}
	return false
}
