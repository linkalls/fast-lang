package linter

import "github.com/linkalls/zeno-lang/ast" // Assuming ast package path

// Rule defines the interface for a linting rule.
type Rule interface {
	Name() string                                     // Returns the unique name of the rule.
	Description() string                              // Describes what the rule checks for.
	Check(node ast.Node, program *ast.Program) []Issue // Performs the check on the given AST node.
	                                                  // `program` provides context of the whole program if needed.
}
