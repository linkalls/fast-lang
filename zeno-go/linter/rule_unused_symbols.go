package linter

import (
	"fmt"
	"github.com/linkalls/zeno-lang/ast"
)

// UnusedVariableRule (L1)
// Detects variables declared with 'let' that are not used.
type UnusedVariableRule struct{}

func (r *UnusedVariableRule) Name() string {
	return "unused-variable"
}

func (r *UnusedVariableRule) Description() string {
	return "Detects variables declared with 'let' that are not used."
}

// Check for UnusedVariableRule is a no-op during individual node traversal.
// The main logic is performed in PostCheck after all nodes have been visited.
func (r *UnusedVariableRule) Check(node ast.Node, program *ast.Program) []Issue {
	return nil // No issues are reported on a per-node basis for this rule.
}

// PostCheck is called by the Linter after the AST traversal is complete.
// It uses the collected declaration and usage information to find unused variables.
func (r *UnusedVariableRule) PostCheck(
	declaredVars map[string]ast.Node, // var name -> declaration node
	usedVars map[string]bool,         // var name -> true if used
	filepath string,
) []Issue {
	issues := []Issue{}

	for varName, declNode := range declaredVars {
		if varName == "_" { // The underscore variable is conventionally ignored for unused checks
			continue
		}
		if !usedVars[varName] {
			// TODO: Extract Line/Column from declNode once AST nodes support position info.
			// For now, using placeholder 0.
			line := 0
			column := 0
			// TODO: Extract Line/Column from declNode once AST nodes support position info.
            // The following is to "use" declNode to prevent compiler errors about unused variables
            // until proper position extraction from AST nodes is implemented.
            if declNode != nil {
                _ = declNode // Explicitly mark as used for now
            }
			// if declNodeWithPos, ok := declNode.(ast.PositionalNode); ok { // Hypothetical interface
			// 	line = declNodeWithPos.Line()
			// 	column = declNodeWithPos.Column()
			// }

			issues = append(issues, Issue{
				Filepath: filepath,
				Line:     line,   // Placeholder
				Column:   column, // Placeholder
				RuleName: r.Name(),
				Message:  fmt.Sprintf("Variable '%s' is declared but not used.", varName),
			})
		}
	}
	return issues
}

// --- UnusedImportRule (L5) ---

// UnusedImportRule detects symbols imported from modules that are not used.
type UnusedImportRule struct{}

func (r *UnusedImportRule) Name() string {
	return "unused-import"
}

func (r *UnusedImportRule) Description() string {
	return "Detects symbols imported from modules that are not used in the current file."
}

// Check for UnusedImportRule is a no-op during individual node traversal.
func (r *UnusedImportRule) Check(node ast.Node, program *ast.Program) []Issue {
	return nil
}

// PostCheck is called by the Linter after the AST traversal is complete.
func (r *UnusedImportRule) PostCheck(
	importedSymbols map[string]*ast.ImportStatement, // Imported symbol name -> its ast.ImportStatement node
	usedImportedSymbols map[string]bool,             // Imported symbol name -> true if used
	filepath string,
) []Issue {
	issues := []Issue{}

	for symbolName, importStmtNode := range importedSymbols {
		if !usedImportedSymbols[symbolName] {
			line := 0
			column := 0
			// TODO: Extract Line/Column from importStmtNode or even more precisely
			// from the specific symbol within the import list if AST supports it.
			// if importStmtNodeWithPos, ok := importStmtNode.(ast.PositionalNode); ok {
			// 	line = importStmtNodeWithPos.Line()
			// 	column = importStmtNodeWithPos.Column()
			// }

			issues = append(issues, Issue{
				Filepath: filepath,
				Line:     line,   // Placeholder: Line of the import statement
				Column:   column, // Placeholder
				RuleName: r.Name(),
				Message:  fmt.Sprintf("Imported symbol '%s' from module '%s' is not used.", symbolName, importStmtNode.Module),
			})
		}
	}
	return issues
}

// --- UnusedFunctionRule (L2) ---

// UnusedFunctionRule detects non-public functions that are defined but not used.
type UnusedFunctionRule struct{}

func (r *UnusedFunctionRule) Name() string {
	return "unused-function"
}

func (r *UnusedFunctionRule) Description() string {
	return "Detects non-public functions ('fn') that are defined but not used. Excludes 'main'."
}

// Check for UnusedFunctionRule is a no-op during individual node traversal.
func (r *UnusedFunctionRule) Check(node ast.Node, program *ast.Program) []Issue {
	return nil
}

// PostCheck is called by the Linter after the AST traversal is complete.
func (r *UnusedFunctionRule) PostCheck(
	declaredFns map[string]*ast.FunctionDefinition, // Zeno fn name -> AST Node
	calledFns map[string]bool,                     // Zeno fn name -> true if called
	filepath string,
) []Issue {
	issues := []Issue{}
	for fnName, fnDefNode := range declaredFns {
		// The logic to only add non-public, non-main functions to declaredFns
		// will be in the visitor's VisitFunctionDefinition.
		// Here we assume declaredFns contains only the functions we care about (non-public, non-main).
		if !calledFns[fnName] {
			line := 0
			column := 0
			// TODO: Extract Line/Column from fnDefNode once AST nodes support position info.
            // The following is to "use" fnDefNode to prevent compiler errors about unused variables
            // until proper position extraction from AST nodes is implemented.
            if fnDefNode != nil {
                _ = fnDefNode // Explicitly mark as used for now
            }
			// if fnDefNodeWithPos, ok := fnDefNode.(ast.PositionalNode); ok { // Hypothetical interface fully removed
			// 	 line = fnDefNodeWithPos.Line()
			// 	 column = fnDefNodeWithPos.Column()
			// }

			issues = append(issues, Issue{
				Filepath: filepath,
				Line:     line,   // Placeholder
				Column:   column, // Placeholder
				RuleName: r.Name(),
				Message:  fmt.Sprintf("Function '%s' is defined but not used.", fnName),
			})
		}
	}
	return issues
}

// Note: We might need a way to identify rules that need a PostCheck call.
// This could be done via a type assertion in Linter.Lint or by adding an optional interface.
// For example:
// type PostTraversalRule interface {
//     Rule
//     PostCheck( /* ... params ... */ ) []Issue
// }
