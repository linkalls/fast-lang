package linter

import (
	"fmt" // For potential error formatting
	"strings" // Added for strings.HasPrefix
	"github.com/linkalls/zeno-lang/ast"
)

// Linter orchestrates the linting process.
type Linter struct {
	rules  []Rule
	issues []Issue
	// Optionally, configuration for rules can be added here later.
}

func NewLinter(rules []Rule) *Linter {
	return &Linter{rules: rules}
}

// Lint analyzes the given AST program and returns a list of issues.
func (l *Linter) Lint(program *ast.Program, filepath string) ([]Issue, error) {
	l.issues = []Issue{} // Reset issues for this run

	visitor := &linterVisitor{
		linter:         l,
		filepath:       filepath,
		program:             program,
		declaredVars:        make(map[string]ast.Node),
		usedVars:            make(map[string]bool),
		declaredFns:         make(map[string]*ast.FunctionDefinition),
		calledFns:           make(map[string]bool),
		importedSymbols:     make(map[string]*ast.ImportStatement),
		usedImportedSymbols: make(map[string]bool),
	}

	if err := Walk(program, visitor); err != nil {
		// If Walk itself returns an error (e.g. from a visitor method), propagate it.
		return l.issues, fmt.Errorf("error during AST walk for file %s: %w", filepath, err)
	}

	// Post-traversal checks for rules that require them
	for _, rule := range l.rules {
		// Check for UnusedVariableRule
		if uvRule, ok := rule.(*UnusedVariableRule); ok {
			postIssues := uvRule.PostCheck(visitor.declaredVars, visitor.usedVars, filepath)
			l.issues = append(l.issues, postIssues...)
		}
		// Check for UnusedFunctionRule
		if ufRule, ok := rule.(*UnusedFunctionRule); ok {
			postIssues := ufRule.PostCheck(visitor.declaredFns, visitor.calledFns, filepath)
			l.issues = append(l.issues, postIssues...)
		}
		// Check for UnusedImportRule
		if uiRule, ok := rule.(*UnusedImportRule); ok {
			postIssues := uiRule.PostCheck(visitor.importedSymbols, visitor.usedImportedSymbols, filepath)
			l.issues = append(l.issues, postIssues...)
		}
		// Example of using an interface for PostCheck if preferred later:
		// if postCheckRule, ok := rule.(interface {
		// 	PostCheck( /* need a generic way or multiple interfaces */ ) []Issue
		// }); ok {
		// 	// ...
		// }
	}

	return l.issues, nil
}

// RegisterRule adds a rule to the linter.
func (l *Linter) RegisterRule(rule Rule) {
	l.rules = append(l.rules, rule)
}

// --- linterVisitor Implementation ---

type linterVisitor struct {
	linter         *Linter
	filepath       string
	program        *ast.Program
	declaredVars        map[string]ast.Node // var name -> declaration node (for position)
	usedVars            map[string]bool     // var name -> true if used
	declaredFns         map[string]*ast.FunctionDefinition // Zeno fn name -> AST Node
	calledFns           map[string]bool     // Zeno fn name -> true if called
	importedSymbols     map[string]*ast.ImportStatement    // Imported symbol name -> its ast.ImportStatement node
	usedImportedSymbols map[string]bool     // Imported symbol name -> true if used
}

func (v *linterVisitor) applyRules(node ast.Node) error {
	for _, rule := range v.linter.rules {
		issues := rule.Check(node, v.program)
		for i := range issues {
			if issues[i].Filepath == "" {
				issues[i].Filepath = v.filepath
			}
			// TODO: Populate Line/Column from AST node if available and not set by rule.
			// This requires AST nodes to have position information.
			// For now, rules are responsible or this needs to be added when AST supports it.
		}
		v.linter.issues = append(v.linter.issues, issues...)
	}
	return nil // Individual rules don't stop the walk; errors from Walk itself would.
}

func (v *linterVisitor) VisitProgram(node *ast.Program) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitImportStatement(node *ast.ImportStatement) error {
	if v.importedSymbols != nil {
		for _, importedName := range node.Imports {
			// For `import {a, b as c} from "mod"`, ast.ImportStatement.Imports
			// currently stores the final name used in the file (e.g. "a", "c").
			// If ast.ImportIdentifier included original name and alias separately,
			// this would need adjustment. Assuming Imports is []string of effective names.
			v.importedSymbols[importedName] = node
		}
	}
	return v.applyRules(node)
}

func (v *linterVisitor) VisitLetDeclaration(node *ast.LetDeclaration) error {
	// Store variable declaration
	if v.declaredVars != nil {
		v.declaredVars[node.Name] = node // Store the node itself for position info later
	}
	// Also apply other rules to this node
	return v.applyRules(node)
}

func (v *linterVisitor) VisitAssignmentStatement(node *ast.AssignmentStatement) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitExpressionStatement(node *ast.ExpressionStatement) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitFunctionDefinition(node *ast.FunctionDefinition) error {
	if v.declaredFns != nil && node.Name != "main" && !node.IsPublic {
		v.declaredFns[node.Name] = node
	}
	return v.applyRules(node)
}

func (v *linterVisitor) VisitReturnStatement(node *ast.ReturnStatement) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitIfStatement(node *ast.IfStatement) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitWhileStatement(node *ast.WhileStatement) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitBlock(node *ast.Block) error {
	return v.applyRules(node)
}

// Expressions
func (v *linterVisitor) VisitIdentifier(node *ast.Identifier) error {
	// Mark variable as used if it's in declaredVars
	// This is a simplified check; context matters (e.g., LHS of assignment is not a "use")
	// For a basic unused variable check, any read-like reference is a use.
	// A more sophisticated check would look at the parent node to determine context.
	if v.usedVars != nil {
		if _, isDeclared := v.declaredVars[node.Value]; isDeclared {
			v.usedVars[node.Value] = true
		}
	}
	// Check if the identifier is an imported symbol
	if v.usedImportedSymbols != nil {
		if _, isImported := v.importedSymbols[node.Value]; isImported {
			v.usedImportedSymbols[node.Value] = true
		}
	}
	return v.applyRules(node)
}

func (v *linterVisitor) VisitIntegerLiteral(node *ast.IntegerLiteral) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitStringLiteral(node *ast.StringLiteral) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitBooleanLiteral(node *ast.BooleanLiteral) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitFunctionCall(node *ast.FunctionCall) error {
	if v.calledFns != nil {
		// Mark Zeno-defined functions as called
		if !strings.HasPrefix(node.Name, "__native_") {
			v.calledFns[node.Name] = true
		}
	}
	// Mark imported functions as used
	if v.usedImportedSymbols != nil {
		if _, isImported := v.importedSymbols[node.Name]; isImported {
			v.usedImportedSymbols[node.Name] = true
		}
	}
	return v.applyRules(node)
}

func (v *linterVisitor) VisitBinaryExpression(node *ast.BinaryExpression) error {
	return v.applyRules(node)
}

func (v *linterVisitor) VisitUnaryExpression(node *ast.UnaryExpression) error {
	return v.applyRules(node)
}
