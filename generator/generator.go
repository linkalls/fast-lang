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

// snakeToCamel converts snake_case to UpperCamelCase.
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// GenerationError represents errors during code generation
type GenerationError struct {
	Message string
}

func (e GenerationError) Error() string {
	return "Generation Error: " + e.Message
}

// Generator manages code generation with scope and import tracking
type Generator struct {
	imports      map[string][]string
	declaredVars map[string]bool
	usedVars     map[string]bool
	declaredFns  map[string]string
	usedFns      map[string]bool
	importTypes  map[string][]string // 型インポートの追跡
	userModules  map[string]map[string]string
	moduleASTs   map[string]*ast.Program
	standardLibs map[string]map[string]string
	currentDir   string
	symbolTable  *types.SymbolTable
	program      *ast.Program
}

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
		importTypes:  make(map[string][]string),
	}
	return g
}

func Generate(program *ast.Program) (string, error) {
	return GenerateWithOptions(program)
}

func GenerateWithOptions(program *ast.Program) (string, error) {
	return GenerateWithFile(program, "")
}

func GenerateWithFile(program *ast.Program, sourceFile string) (string, error) {
	g := NewGenerator()
	g.currentDir = sourceFile
	g.program = program
	return g.generateProgram(program)
}

func (g *Generator) generateProgram(program *ast.Program) (string, error) {
	var builder strings.Builder
	if err := g.validateFunctionTypes(program); err != nil {
		return "", err
	}
	for _, stmt := range program.Statements {
		if err := g.collectImportsAndDeclarations(stmt); err != nil {
			return "", err
		}
	}
	builder.WriteString("package main\n\n")
	builder.WriteString("import (\n")
	requiredImports := make(map[string]bool)
	for module := range g.imports {
		if strings.HasPrefix(module, "std/") {
			if module == "std/fmt" {
				requiredImports["fmt"] = true
			} else if module == "std/io" {
				requiredImports["os"] = true
			}
		}
	}
	requiredImports["fmt"] = true
	requiredImports["os"] = true
	requiredImports["encoding/json"] = true
	for imp := range requiredImports {
		builder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
	}
	builder.WriteString(")\n\n")
	// Generate Go generic type alias for Zeno 'Result<T>'
	for _, stmt := range program.Statements {
		if tdecl, ok := stmt.(*ast.TypeDeclaration); ok && tdecl.Name == "Result" && len(tdecl.Generics) == 1 {
			gen := tdecl.Generics[0]
			builder.WriteString(fmt.Sprintf("type Result[%s any] struct {\n", gen))
			builder.WriteString("\tOk bool\n")
			builder.WriteString(fmt.Sprintf("\tValue %s\n", gen))
			builder.WriteString("\tError string\n")
			builder.WriteString("}\n\n")
			break
		}
	}

	// Generate type definitions for imported types
	for modulePath, typeNames := range g.importTypes {
		if moduleAST, exists := g.moduleASTs[modulePath]; exists {
			for _, stmt := range moduleAST.Statements {
				if typeDecl, ok := stmt.(*ast.TypeDeclaration); ok {
					for _, typeName := range typeNames {
						if typeDecl.Name == typeName {
							// Generate a simple type alias for now
							builder.WriteString(fmt.Sprintf("type %s map[string]interface{}\n\n", typeName))
						}
					}
				}
			}
		}
	}
	g.generateNativeFunctionHelpers(&builder)
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
			otherStmts = append(otherStmts, stmt)
		}
	}
	for modulePath, moduleAST := range g.moduleASTs {
		if importedFuncs, exists := g.imports[modulePath]; exists {
			for _, stmt := range moduleAST.Statements {
				if funcDef, ok := stmt.(*ast.FunctionDefinition); ok && funcDef.IsPublic {
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
	for _, funcDef := range functionDefs {
		if err := g.generateStatement(funcDef, &builder, 0); err != nil {
			return "", err
		}
		builder.WriteString("\n")
	}
	builder.WriteString("func main() {\n")
	if mainFunc != nil {
		for _, bodyStmt := range mainFunc.Body {
			if err := g.generateStatement(bodyStmt, &builder, 1); err != nil {
				return "", err
			}
		}
	} else if len(otherStmts) > 0 {
		for _, stmt := range otherStmts {
			if err := g.generateStatement(stmt, &builder, 1); err != nil {
				return "", err
			}
		}
	}
	builder.WriteString("}\n")
	if err := g.checkUnusedVariables(); err != nil {
		return "", err
	}
	if err := g.checkUnusedFunctions(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func indent(level int) string { return strings.Repeat("\t", level) }

// getGoTypeForZenoPrimitiveType converts a Zeno primitive type to its Go equivalent string.
func getGoTypeForZenoPrimitiveType(zenoType types.Type) string {
	switch zenoType {
	case types.IntType:
		return "int"
	case types.FloatType:
		return "float64"
	case types.StringType:
		return "string"
	case types.BoolType:
		return "bool"
	default:
		return "interface{}" // Default for non-primitive or unknown types
	}
}

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
	case "any":
		return "interface{}"
	case "void":
		return ""
	default:
		return zenoType
	}
}

func (g *Generator) generateStatement(stmt ast.Statement, builder *strings.Builder, indentLevel int) error {
	switch s := stmt.(type) {
	case *ast.TypeDeclaration:
		// skip type declarations
		return nil
	case *ast.ImportStatement:
		return nil
	case *ast.LetDeclaration:
		var varType types.Type
		if s.TypeAnn != nil {
			varType = g.mapASTTypeToType(*s.TypeAnn)
		} else {
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
		g.usedVars[s.Name] = true
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
		functionName := s.Name
		if s.IsPublic {
			if len(functionName) > 0 {
				functionName = strings.ToUpper(string(functionName[0])) + functionName[1:]
			}
		} else {
			if functionName != "main" && len(functionName) > 0 {
				functionName = strings.ToLower(string(functionName[0])) + functionName[1:]
			}
		}
		builder.WriteString(functionName)
		// Generic type parameters
		if len(s.Generics) > 0 {
			builder.WriteString("[")
			for i, gen := range s.Generics {
				if i > 0 {
					builder.WriteString(", ")
				}
				builder.WriteString(gen)
				builder.WriteString(" any")
			}
			builder.WriteString("]")
		}
		builder.WriteString("(")
		for i, param := range s.Parameters {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(param.Name)
			builder.WriteString(" ")
			if param.Variadic {
				builder.WriteString("...")
			}
			builder.WriteString(mapType(param.Type))
		}
		builder.WriteString(")")
		if s.ReturnType != nil {
			builder.WriteString(" ")
			builder.WriteString(mapType(*s.ReturnType))
		}
		builder.WriteString(" {\n")
		originalSymbolTable := g.symbolTable
		g.symbolTable = types.NewSymbolTable(originalSymbolTable)
		for _, param := range s.Parameters {
			paramType := g.mapASTTypeToType(param.Type)
			g.symbolTable.Define(param.Name, paramType)
		}
		for _, bodyStmt := range s.Body {
			if err := g.generateStatement(bodyStmt, builder, indentLevel+1); err != nil {
				g.symbolTable = originalSymbolTable
				return err
			}
		}
		g.symbolTable = originalSymbolTable
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
	case *ast.ForStatement:
		// s は *ast.ForStatement 型としてバインドされるので、そのまま利用
		builder.WriteString(indent(indentLevel))
		// Zeno の for-in を Go の range ループに変換
		builder.WriteString("for _, " + s.VarName + " := range ")
		if err := g.generateExpression(s.Iterable, builder); err != nil {
			return err
		}
		builder.WriteString(" ")
		if err := g.generateBlock(s.Body, builder, indentLevel); err != nil {
			return err
		}
		builder.WriteString("\n")
	default:
		return GenerationError{Message: fmt.Sprintf("Unsupported statement type: %T", stmt)}
	}
	return nil
}

func (g *Generator) generateExpression(expr ast.Expression, builder *strings.Builder) error {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		builder.WriteString(strconv.Itoa(e.Value))
	case *ast.StringLiteral:
		escaped := strconv.Quote(e.Value)
		builder.WriteString(escaped)
	case *ast.FloatLiteral:
		builder.WriteString(strconv.FormatFloat(e.Value, 'f', -1, 64))
	case *ast.BooleanLiteral:
		if e.Value {
			builder.WriteString("true")
		} else {
			builder.WriteString("false")
		}
	case *ast.Identifier:
		builder.WriteString(e.Value)

	case *ast.MemberExpression:
		// Generate map or struct field access as index into map
		if err := g.generateExpression(e.Object, builder); err != nil {
			return err
		}
		builder.WriteString("[")
		builder.WriteString(strconv.Quote(e.Property))
		builder.WriteString("]")

	case *ast.ArrayLiteral:
		if len(e.Elements) == 0 {
			builder.WriteString("[]interface{}{}") // Default for empty array
		} else {
			// Determine element type based on the first element, as parser ensures homogeneity for primitives.
			// For generator, if parser passed it, we trust homogeneity for primitives.
			// If not a primitive type according to parser, it would be an error there,
			// or it's a more complex type not yet handled for specific Go typing.
			firstElementZenoType := g.inferType(e.Elements[0])
			goElementType := getGoTypeForZenoPrimitiveType(firstElementZenoType)

			// If the parser guarantees homogeneity for primitive types, we use that type.
			// If the elements are not primitives (e.g. identifiers, function calls),
			// getGoTypeForZenoPrimitiveType will return "interface{}", leading to []interface{}.
			// This is a safe default if specific typing isn't possible here.
			builder.WriteString(fmt.Sprintf("[]%s{", goElementType))
			for i, elem := range e.Elements {
				if i > 0 {
					builder.WriteString(", ")
				}
				if err := g.generateExpression(elem, builder); err != nil {
					return err
				}
			}
			builder.WriteString("}")
		}
	case *ast.MapLiteral:
		builder.WriteString("map[string]interface{}{")
		count := 0
		for keyExpr, valueExpr := range e.Pairs {
			if count > 0 {
				builder.WriteString(", ")
			}
			// Process key
			var keyString string
			switch k := keyExpr.(type) {
			case *ast.Identifier:
				keyString = k.Value
			case *ast.StringLiteral:
				keyString = k.Value
			default:
				// Should not happen if parser validation is correct
				return GenerationError{Message: fmt.Sprintf("unsupported map key type: %T", k)}
			}
			builder.WriteString(fmt.Sprintf("\"%s\": ", keyString))

			// Process value
			if err := g.generateExpression(valueExpr, builder); err != nil {
				return err
			}
			count++
		}
		builder.WriteString("}")
	case *ast.UnaryExpression:
		builder.WriteString("(")
		builder.WriteString(e.Operator.String())
		if err := g.generateExpression(e.Right, builder); err != nil {
			return err
		}
		builder.WriteString(")")
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
		// Check if function is imported first, before special-casing
		var functionName string
		if goName, exists := g.declaredFns[e.Name]; exists {
			functionName = goName
		} else {
			// Special-case Zeno print and println only if not imported
			if e.Name == "println" {
				builder.WriteString("fmt.Println(")
				for i, arg := range e.Arguments {
					if i > 0 {
						builder.WriteString(", ")
					}
					if err := g.generateExpression(arg, builder); err != nil {
						return err
					}
				}
				builder.WriteString(")")
				return nil
			}
			if e.Name == "print" {
				builder.WriteString("fmt.Print(")
				for i, arg := range e.Arguments {
					if i > 0 {
						builder.WriteString(", ")
					}
					if err := g.generateExpression(arg, builder); err != nil {
						return err
					}
				}
				builder.WriteString(")")
				return nil
			}
			functionName = e.Name
		}
		if err := g.validateImports(e.Name); err != nil {
			return err
		}
		builder.WriteString(functionName)
		builder.WriteString("(")
		// generate arguments
		for i, arg := range e.Arguments {
			if i > 0 {
				builder.WriteString(", ")
			}
			if err := g.generateExpression(arg, builder); err != nil {
				return err
			}
		}
		builder.WriteString(")")
	case *ast.StructLiteral:
		// Generate struct literal as map[string]interface{}
		builder.WriteString("map[string]interface{}{")
		count := 0
		for fieldName, valueExpr := range e.Fields {
			if count > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("\"%s\": ", fieldName))
			if err := g.generateExpression(valueExpr, builder); err != nil {
				return err
			}
			count++
		}
		builder.WriteString("}")
	default:
		return GenerationError{Message: fmt.Sprintf("Unsupported expression type: %T", expr)}
	}
	return nil
}

func (g *Generator) generateCondition(expr ast.Expression, builder *strings.Builder) error {
	// ... (content remains the same as fetched in Turn 61) ...
	switch e := expr.(type) {
	case *ast.BooleanLiteral:
		return g.generateExpression(expr, builder)
	case *ast.BinaryExpression:
		return g.generateExpression(expr, builder)
	case *ast.IntegerLiteral:
		builder.WriteString("(")
		if err := g.generateExpression(expr, builder); err != nil {
			return err
		}
		builder.WriteString(" != 0)")
		return nil
	case *ast.Identifier:
		varType := g.getVariableType(e.Value)
		// fmt.Printf("DEBUG: Variable %s has type %v\n", e.Value, varType)
		switch varType {
		case types.BoolType:
			return g.generateExpression(expr, builder)
		case types.IntType:
			builder.WriteString("(")
			if err := g.generateExpression(expr, builder); err != nil {
				return err
			}
			builder.WriteString(" != 0)")
			return nil
		case types.StringType:
			builder.WriteString("(")
			if err := g.generateExpression(expr, builder); err != nil {
				return err
			}
			builder.WriteString(" != \"\")")
			return nil
		case types.FloatType:
			builder.WriteString("(")
			if err := g.generateExpression(expr, builder); err != nil {
				return err
			}
			builder.WriteString(" != 0.0)")
			return nil
		default:
			return g.generateExpression(expr, builder)
		}
	default:
		return g.generateExpression(expr, builder)
	}
}

func (g *Generator) generateBlock(block *ast.Block, builder *strings.Builder, indentLevel int) error {
	// ... (content remains the same as fetched in Turn 61) ...
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

func (g *Generator) collectImportsAndDeclarations(stmt ast.Statement) error {
	// ... (content remains the same as fetched in Turn 61) ...
	switch s := stmt.(type) {
	case *ast.ImportStatement:
		// 関数インポートと型インポートを分離
		var names []string
		var typeNames []string
		for _, imp := range s.Imports {
			if imp.IsType {
				typeNames = append(typeNames, imp.Name)
			} else {
				names = append(names, imp.Name)
			}
		}
		g.imports[s.Module] = names
		if len(typeNames) > 0 {
			g.importTypes[s.Module] = typeNames
		}
		if strings.HasPrefix(s.Module, "std/") {
			if err := g.processStdModule(s.Module, names, typeNames); err != nil {
				return err
			}
		} else if strings.HasPrefix(s.Module, "./") || strings.HasPrefix(s.Module, "../") {
			if err := g.processUserModule(s.Module, names); err != nil {
				return err
			}
		}
	case *ast.LetDeclaration:
		g.declaredVars[s.Name] = true
		var varType types.Type
		if s.TypeAnn != nil {
			varType = g.mapASTTypeToType(*s.TypeAnn)
		} else {
			varType = g.inferType(s.ValueExpression)
		}
		g.registerVariableWithType(s.Name, varType)
		if s.ValueExpression != nil {
			g.markVariableUsage(s.ValueExpression)
		}
	case *ast.AssignmentStatement:
		g.usedVars[s.Name] = true
		g.markVariableUsage(s.Value)
	case *ast.FunctionDefinition:
		goFuncName := s.Name
		if s.IsPublic {
			if len(goFuncName) > 0 {
				goFuncName = strings.ToUpper(s.Name[:1]) + s.Name[1:]
			}
		}
		g.declaredFns[s.Name] = goFuncName
		for _, bodyStmt := range s.Body {
			g.collectImportsAndDeclarations(bodyStmt)
		}
	case *ast.ReturnStatement:
		if s.Value != nil {
			g.markVariableUsage(s.Value)
		}
	case *ast.ExpressionStatement:
		g.markVariableUsage(s.Expression)
	case *ast.IfStatement:
		g.markVariableUsage(s.Condition)
		if s.ThenBlock != nil {
			g.markBlockUsage(s.ThenBlock)
		}
		for _, elseIf := range s.ElseIfClauses {
			g.markVariableUsage(elseIf.Condition)
			if elseIf.Block != nil {
				g.markBlockUsage(elseIf.Block)
			}
		}
		if s.ElseBlock != nil {
			g.markBlockUsage(s.ElseBlock)
		}
	case *ast.WhileStatement:
		g.markVariableUsage(s.Condition)
		if s.Block != nil {
			g.markBlockUsage(s.Block)
		}
	}
	return nil
}

func (g *Generator) markVariableUsage(expr ast.Expression) {
	// ... (content remains the same as fetched in Turn 61) ...
	switch e := expr.(type) {
	case *ast.Identifier:
		g.usedVars[e.Value] = true
	case *ast.BooleanLiteral, *ast.IntegerLiteral, *ast.StringLiteral:
		// No action needed
	case *ast.BinaryExpression:
		g.markVariableUsage(e.Left)
		g.markVariableUsage(e.Right)
	case *ast.UnaryExpression:
		g.markVariableUsage(e.Right)
	case *ast.FunctionCall:
		g.usedFns[e.Name] = true
		for _, arg := range e.Arguments {
			g.markVariableUsage(arg)
		}
	case *ast.MemberExpression:
		// Mark the object variable as used
		g.markVariableUsage(e.Object)
	}
}

func (g *Generator) markBlockUsage(block *ast.Block) {
	// ... (content remains the same as fetched in Turn 61) ...
	if block == nil {
		return
	}
	for _, stmt := range block.Statements {
		g.collectImportsAndDeclarations(stmt)
	}
}

func (g *Generator) checkUnusedVariables() error {
	// ... (content remains the same as fetched in Turn 61) ...
	var unusedVars []string
	for varName := range g.declaredVars {
		if !g.usedVars[varName] {
			unusedVars = append(unusedVars, varName)
		}
	}
	if len(unusedVars) > 0 {
		return GenerationError{Message: fmt.Sprintf("Unused variables found: %s", strings.Join(unusedVars, ", "))}
	}
	return nil
}

func (g *Generator) checkUnusedFunctions() error {
	// ... (content remains the same as fetched in Turn 61) ...
	var unusedFns []string
	for fnName := range g.declaredFns {
		if fnName == "main" {
			continue
		}
		if g.isPublicFunction(fnName) {
			continue
		}
		if !g.usedFns[fnName] {
			unusedFns = append(unusedFns, fnName)
		}
	}
	if len(unusedFns) > 0 {
		return GenerationError{Message: fmt.Sprintf("Unused functions found: %s", strings.Join(unusedFns, ", "))}
	}
	return nil
}

func (g *Generator) isPublicFunction(fnName string) bool {
	// ... (content remains the same as fetched in Turn 61) ...
	if goFuncName, exists := g.declaredFns[fnName]; exists {
		if len(goFuncName) == 0 {
			return false
		}
		return goFuncName[0] >= 'A' && goFuncName[0] <= 'Z'
	}
	return false
}

func (g *Generator) validateImports(functionName string) error {
	// ... (content remains the same as fetched in Turn 61) ...
	builtinFunctions := map[string]bool{}
	if builtinFunctions[functionName] {
		return nil
	}
	for module, functions := range g.standardLibs {
		if _, exists := functions[functionName]; exists {
			if importedFuncs, imported := g.imports[module]; imported {
				for _, importedFunc := range importedFuncs {
					if importedFunc == functionName {
						return nil
					}
				}
			}
			return GenerationError{Message: fmt.Sprintf("Function '%s' is not imported from '%s'", functionName, module)}
		}
	}
	for module, functions := range g.userModules {
		if _, exists := functions[functionName]; exists {
			if importedFuncs, imported := g.imports[module]; imported {
				for _, importedFunc := range importedFuncs {
					if importedFunc == functionName {
						return nil
					}
				}
			}
			return GenerationError{Message: fmt.Sprintf("Function '%s' is not imported from '%s'", functionName, module)}
		}
	}
	return nil
}

func (g *Generator) processUserModule(modulePath string, importedFunctions []string) error {
	// ... (content remains the same as fetched in Turn 61) ...
	var zenoFilePath string
	if strings.HasSuffix(modulePath, ".zeno") {
		zenoFilePath = modulePath
	} else {
		zenoFilePath = modulePath + ".zeno"
	}
	if (strings.HasPrefix(zenoFilePath, "./") || strings.HasPrefix(zenoFilePath, "../")) && g.currentDir != "" {
		baseDir := filepath.Dir(g.currentDir)
		zenoFilePath = filepath.Join(baseDir, zenoFilePath)
	}
	content, err := os.ReadFile(zenoFilePath)
	if err != nil {
		return GenerationError{Message: fmt.Sprintf("Failed to read module file '%s': %v", zenoFilePath, err)}
	}
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return GenerationError{Message: fmt.Sprintf("Parse errors in module '%s': %v", zenoFilePath, p.Errors())}
	}
	publicFunctions := make(map[string]string)
	for _, stmt := range program.Statements {
		if funcDef, ok := stmt.(*ast.FunctionDefinition); ok && funcDef.IsPublic {
			goFuncName := funcDef.Name
			if len(goFuncName) > 0 {
				goFuncName = strings.ToUpper(string(goFuncName[0])) + goFuncName[1:]
			}
			publicFunctions[funcDef.Name] = goFuncName
		}
	}
	for _, importedFunc := range importedFunctions {
		if _, exists := publicFunctions[importedFunc]; !exists {
			return GenerationError{Message: fmt.Sprintf("Function '%s' is not exported from module '%s'", importedFunc, modulePath)}
		}
		// Add imported function to declaredFns for proper name resolution
		g.declaredFns[importedFunc] = publicFunctions[importedFunc]
	}
	g.userModules[modulePath] = publicFunctions
	g.moduleASTs[modulePath] = program
	return nil
}

func (g *Generator) processStdModule(modulePath string, importedFunctions []string, importedTypes []string) error {
	// ... (content remains the same as fetched in Turn 61) ...
	moduleShortName := strings.TrimPrefix(modulePath, "std/")
	zenoFilePath := filepath.Join("std", moduleShortName+".zeno")
	content, err := os.ReadFile(zenoFilePath)
	if err != nil {
		return GenerationError{Message: fmt.Sprintf("Failed to read module file '%s': %v", zenoFilePath, err)}
	}
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return GenerationError{Message: fmt.Sprintf("Parse errors in module '%s': %v", zenoFilePath, p.Errors())}
	}
	publicFunctions := make(map[string]string)
	publicTypes := make(map[string]string)

	for _, stmt := range program.Statements {
		if funcDef, ok := stmt.(*ast.FunctionDefinition); ok && funcDef.IsPublic {
			goFuncName := funcDef.Name
			if len(goFuncName) > 0 {
				goFuncName = strings.ToUpper(string(goFuncName[0])) + goFuncName[1:]
			}
			publicFunctions[funcDef.Name] = goFuncName
		} else if typeDef, ok := stmt.(*ast.TypeDeclaration); ok {
			// Handle type declarations - assume all types in std modules are public
			publicTypes[typeDef.Name] = typeDef.Name
		}
	}

	for _, importedFunc := range importedFunctions {
		if _, exists := publicFunctions[importedFunc]; !exists {
			return GenerationError{Message: fmt.Sprintf("Function '%s' is not exported from module '%s'", importedFunc, modulePath)}
		}
		// Add imported function to declaredFns for proper name resolution
		g.declaredFns[importedFunc] = publicFunctions[importedFunc]
	}

	// Auto-import dependent types for std/result functions
	if modulePath == "std/result" && len(importedFunctions) > 0 {
		// If importing functions from std/result, auto-import Result type
		resultTypeAlreadyImported := false
		for _, typeName := range importedTypes {
			if typeName == "Result" {
				resultTypeAlreadyImported = true
				break
			}
		}
		if !resultTypeAlreadyImported {
			// Auto-add Result type to importTypes
			if g.importTypes[modulePath] == nil {
				g.importTypes[modulePath] = []string{}
			}
			g.importTypes[modulePath] = append(g.importTypes[modulePath], "Result")
		}
	}

	for _, importedType := range importedTypes {
		if _, exists := publicTypes[importedType]; !exists {
			return GenerationError{Message: fmt.Sprintf("Type '%s' is not exported from module '%s'", importedType, modulePath)}
		}
	}
	g.standardLibs[modulePath] = publicFunctions
	g.moduleASTs[modulePath] = program
	return nil
}

func (g *Generator) generateNativeFunctionHelpers(builder *strings.Builder) {
	// ... (content remains the same as fetched in Turn 61, including JSON helpers) ...
	builder.WriteString("// Native function helpers\n")
	builder.WriteString("func zenoNativeReadFile(filename string) string {\n\tdata, err := os.ReadFile(filename)\n\tif err != nil {\n\t\tfmt.Printf(\"Error reading file %s: %v\\n\", filename, err)\n\t\treturn \"\"\n\t}\n\treturn string(data)\n}\n\n")
	builder.WriteString("func zenoNativeWriteFile(filename string, content string) bool {\n\terr := os.WriteFile(filename, []byte(content), 0644)\n\tif err != nil {\n\t\tfmt.Printf(\"Error writing file %s: %v\\n\", filename, err)\n\t\treturn false\n\t}\n\treturn true\n}\n\n")
	builder.WriteString("func zenoNativePrint(args ...interface{}) {\n\tfmt.Print(args...)\n}\n\n")
	builder.WriteString("func zenoNativePrintln(args ...interface{}) {\n\tfmt.Println(args...)\n}\n\n")

	// Variadic versions that handle slices of any type
	builder.WriteString("func zenoNativePrintVariadic(args []interface{}) {\n\tfmt.Print(args...)\n}\n\n")
	builder.WriteString("func zenoNativePrintlnVariadic(args []interface{}) {\n\tfmt.Println(args...)\n}\n\n")

	// Variadic versions that require at least one argument
	builder.WriteString("func zenoNativePrintVariadicWithFirst(first interface{}, rest []interface{}) {\n\tfmt.Print(first)\n\tfor _, arg := range rest {\n\t\tfmt.Print(\" \", arg)\n\t}\n}\n\n")
	builder.WriteString("func zenoNativePrintlnVariadicWithFirst(first interface{}, rest []interface{}) {\n\tfmt.Print(first)\n\tfor _, arg := range rest {\n\t\tfmt.Print(\" \", arg)\n\t}\n\tfmt.Println()\n}\n\n")
	builder.WriteString("func zenoNativeRemove(path string) bool {\n\terr := os.Remove(path)\n\tif err != nil {\n\t\tfmt.Fprintf(os.Stderr, \"Error removing %s: %v\\n\", path, err)\n\t\treturn false\n\t}\n\treturn true\n}\n\n")
	builder.WriteString("func zenoNativeGetCurrentDirectory() string {\n\tpwd, err := os.Getwd()\n\tif err != nil {\n\t\tfmt.Fprintf(os.Stderr, \"Error getting current directory: %v\\n\", err)\n\t\treturn \"\"\n\t}\n\treturn pwd\n}\n\n")
	builder.WriteString("func zenoNativeJsonParse(jsonString string) interface{} {\n\tvar result interface{}\n\terr := json.Unmarshal([]byte(jsonString), &result)\n\tif err != nil {\n\t\tfmt.Fprintf(os.Stderr, \"Error parsing JSON string '%s': %v\\n\", jsonString, err)\n\t\treturn nil\n\t}\n\treturn result\n}\n\n")
	builder.WriteString("func zenoNativeJsonStringify(value interface{}) string {\n\tjsonBytes, err := json.Marshal(value)\n\tif err != nil {\n\t\tfmt.Fprintf(os.Stderr, \"Error stringifying to JSON for value '%v': %v\\n\", value, err)\n\t\treturn \"\"\n\t}\n\treturn string(jsonBytes)\n}\n\n")
}

func (g *Generator) inferType(expr ast.Expression) types.Type {
	switch e := expr.(type) {
	case *ast.BooleanLiteral:
		return types.BoolType
	case *ast.IntegerLiteral:
		return types.IntType
	case *ast.StringLiteral:
		return types.StringType
	case *ast.FloatLiteral:
		return types.FloatType
	case *ast.ArrayLiteral: // Added
		return types.AnyType // Placeholder for now
	case *ast.Identifier:
		if symbol, ok := g.symbolTable.Resolve(e.Value); ok {
			return symbol.Type
		}
		return types.IntType
	case *ast.BinaryExpression:
		switch e.Operator {
		case ast.BinaryOpEq, ast.BinaryOpNotEq, ast.BinaryOpLt, ast.BinaryOpLte, ast.BinaryOpGt, ast.BinaryOpGte:
			return types.BoolType
		case ast.BinaryOpPlus, ast.BinaryOpMinus, ast.BinaryOpMultiply, ast.BinaryOpDivide, ast.BinaryOpModulo:
			leftType := g.inferType(e.Left)
			rightType := g.inferType(e.Right)
			if leftType == types.FloatType || rightType == types.FloatType {
				return types.FloatType
			}
			return types.IntType
		case ast.BinaryOpAnd, ast.BinaryOpOr:
			return types.BoolType
		}
	case *ast.FunctionCall:
		var funcDef *ast.FunctionDefinition
		if g.program != nil {
			for _, stmt := range g.program.Statements {
				if def, ok := stmt.(*ast.FunctionDefinition); ok && def.Name == e.Name {
					funcDef = def
					break
				}
			}
		}
		if funcDef == nil {
			for modulePath, moduleAST := range g.moduleASTs {
				isImportedFromThisModule := false
				if importedFuncs, exists := g.imports[modulePath]; exists {
					for _, importedFnName := range importedFuncs {
						if importedFnName == e.Name {
							isImportedFromThisModule = true
							break
						}
					}
				}
				if isImportedFromThisModule {
					for _, stmt := range moduleAST.Statements {
						if def, ok := stmt.(*ast.FunctionDefinition); ok && def.Name == e.Name {
							if def.IsPublic {
								funcDef = def
								break
							}
						}
					}
				}
				if funcDef != nil {
					break
				}
			}
		}
		if funcDef != nil && funcDef.ReturnType != nil {
			return g.mapASTTypeToType(*funcDef.ReturnType)
		}
		// fmt.Printf("WARN: Could not accurately determine return type for function call '%s'. Defaulting to IntType.\n", e.Name)
		return types.IntType
	case *ast.UnaryExpression:
		switch e.Operator {
		case ast.UnaryOpBang:
			return types.BoolType
		case ast.UnaryOpMinus:
			return g.inferType(e.Right)
		default:
			// warning: suppressed inference fallback log
			return types.IntType
		}
	}
	return types.IntType
}

func (g *Generator) registerVariableWithType(name string, varType types.Type) {
	g.symbolTable.Define(name, varType)
}

func (g *Generator) getVariableType(name string) types.Type {
	if symbol, ok := g.symbolTable.Resolve(name); ok {
		// debug: suppress output
		return symbol.Type
	}
	// debug: suppress output
	return types.IntType
}

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
		return types.IntType
	}
}

func (g *Generator) validateFunctionTypes(program *ast.Program) error {
	// ... (content remains the same as fetched in Turn 61) ...
	for _, stmt := range program.Statements {
		if funcDef, ok := stmt.(*ast.FunctionDefinition); ok {
			if funcDef.Name == "main" {
				continue
			}
			for _, param := range funcDef.Parameters {
				if param.Type == "" {
					return GenerationError{Message: fmt.Sprintf("Function '%s': parameter '%s' must have an explicit type", funcDef.Name, param.Name)}
				}
			}
			if g.hasValueReturnStatement(funcDef.Body) && funcDef.ReturnType == nil {
				return GenerationError{Message: fmt.Sprintf("Function '%s' contains return statements with values but has no explicit return type", funcDef.Name)}
			}
		}
	}
	return nil
}

func (g *Generator) hasValueReturnStatement(statements []ast.Statement) bool {
	// ... (content remains the same as fetched in Turn 61) ...
	for _, stmt := range statements {
		if rs, ok := stmt.(*ast.ReturnStatement); ok && rs.Value != nil {
			return true
		}
		if ifStmt, ok := stmt.(*ast.IfStatement); ok {
			if ifStmt.ThenBlock != nil && g.hasValueReturnStatement(ifStmt.ThenBlock.Statements) {
				return true
			}
			for _, elseIfClause := range ifStmt.ElseIfClauses {
				if elseIfClause.Block != nil && g.hasValueReturnStatement(elseIfClause.Block.Statements) {
					return true
				}
			}
			if ifStmt.ElseBlock != nil && g.hasValueReturnStatement(ifStmt.ElseBlock.Statements) {
				return true
			}
		}
		if whileStmt, ok := stmt.(*ast.WhileStatement); ok {
			if whileStmt.Block != nil && g.hasValueReturnStatement(whileStmt.Block.Statements) {
				return true
			}
		}
	}
	return false
}
