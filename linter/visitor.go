package linter

import (
	"fmt" // For potential error formatting in Walk later if needed

	"github.com/linkalls/zeno-lang/ast"
)

// Visitor interface for traversing the AST.
type Visitor interface {
	VisitProgram(node *ast.Program) error
	VisitImportStatement(node *ast.ImportStatement) error
	VisitLetDeclaration(node *ast.LetDeclaration) error
	VisitAssignmentStatement(node *ast.AssignmentStatement) error
	VisitExpressionStatement(node *ast.ExpressionStatement) error
	VisitFunctionDefinition(node *ast.FunctionDefinition) error
	VisitReturnStatement(node *ast.ReturnStatement) error
	VisitIfStatement(node *ast.IfStatement) error
	VisitWhileStatement(node *ast.WhileStatement) error
	VisitBlock(node *ast.Block) error

	// Expressions
	VisitIdentifier(node *ast.Identifier) error
	VisitIntegerLiteral(node *ast.IntegerLiteral) error
	VisitStringLiteral(node *ast.StringLiteral) error
	VisitBooleanLiteral(node *ast.BooleanLiteral) error
	VisitFunctionCall(node *ast.FunctionCall) error
	VisitBinaryExpression(node *ast.BinaryExpression) error
	VisitUnaryExpression(node *ast.UnaryExpression) error
	VisitArrayLiteral(node *ast.ArrayLiteral) error   // Added
	VisitMapLiteral(node *ast.MapLiteral) error       // Added
	VisitStructLiteral(node *ast.StructLiteral) error // Added
	// Note: ast.Parameter is not typically visited standalone by this kind of walker,
	// it's part of FunctionDefinition. Similarly for ElseIfClause.
}

// Walk traverses an AST node and its children using the provided visitor.
func Walk(node ast.Node, visitor Visitor) error {
	if node == nil {
		return nil
	}
	var err error

	switch n := node.(type) {
	case *ast.Program:
		if err = visitor.VisitProgram(n); err != nil {
			return err
		}
		for _, stmt := range n.Statements {
			if err = Walk(stmt, visitor); err != nil {
				return fmt.Errorf("in program statement: %w", err)
			}
		}
	case *ast.ImportStatement:
		err = visitor.VisitImportStatement(n)
	case *ast.LetDeclaration:
		if err = visitor.VisitLetDeclaration(n); err != nil {
			return err
		}
		if n.ValueExpression != nil {
			if err = Walk(n.ValueExpression, visitor); err != nil {
				return fmt.Errorf("in let declaration value: %w", err)
			}
		}
	case *ast.AssignmentStatement:
		if err = visitor.VisitAssignmentStatement(n); err != nil {
			return err
		}
		if n.Value != nil {
			if err = Walk(n.Value, visitor); err != nil {
				return fmt.Errorf("in assignment statement value: %w", err)
			}
		}
	case *ast.ExpressionStatement:
		if err = visitor.VisitExpressionStatement(n); err != nil {
			return err
		}
		if n.Expression != nil {
			if err = Walk(n.Expression, visitor); err != nil {
				return fmt.Errorf("in expression statement: %w", err)
			}
		}
	case *ast.FunctionDefinition:
		if err = visitor.VisitFunctionDefinition(n); err != nil {
			return err
		}
		// Parameters are part of FunctionDefinition, not walked as separate nodes here typically
		// Their identifiers might be visited if expressions default to visiting identifiers
		for _, stmt := range n.Body { // Body is []ast.Statement
			if err = Walk(stmt, visitor); err != nil {
				return fmt.Errorf("in function body: %w", err)
			}
		}
	case *ast.ReturnStatement:
		if err = visitor.VisitReturnStatement(n); err != nil {
			return err
		}
		if n.Value != nil {
			if err = Walk(n.Value, visitor); err != nil {
				return fmt.Errorf("in return statement value: %w", err)
			}
		}
	case *ast.IfStatement:
		if err = visitor.VisitIfStatement(n); err != nil {
			return err
		}
		if err = Walk(n.Condition, visitor); err != nil {
			return fmt.Errorf("in if condition: %w", err)
		}
		if n.ThenBlock != nil {
			if err = Walk(n.ThenBlock, visitor); err != nil {
				return fmt.Errorf("in if then block: %w", err)
			}
		}
		for i := range n.ElseIfClauses { // Iterate by index to pass pointer if ElseIfClause itself is a Node
			// Assuming ElseIfClause is not an ast.Node itself, but its components are.
			// If ElseIfClause were an ast.Node, it would be: Walk(&n.ElseIfClauses[i], visitor)
			if err = Walk(n.ElseIfClauses[i].Condition, visitor); err != nil {
				return fmt.Errorf("in else if clause condition: %w", err)
			}
			if n.ElseIfClauses[i].Block != nil {
				if err = Walk(n.ElseIfClauses[i].Block, visitor); err != nil {
					return fmt.Errorf("in else if clause block: %w", err)
				}
			}
		}
		if n.ElseBlock != nil {
			if err = Walk(n.ElseBlock, visitor); err != nil {
				return fmt.Errorf("in if else block: %w", err)
			}
		}
	case *ast.WhileStatement:
		if err = visitor.VisitWhileStatement(n); err != nil {
			return err
		}
		if err = Walk(n.Condition, visitor); err != nil {
			return fmt.Errorf("in while condition: %w", err)
		}
		if n.Block != nil {
			if err = Walk(n.Block, visitor); err != nil {
				return fmt.Errorf("in while block: %w", err)
			}
		}
	case *ast.Block:
		if err = visitor.VisitBlock(n); err != nil {
			return err
		}
		for _, stmt := range n.Statements {
			if err = Walk(stmt, visitor); err != nil {
				return fmt.Errorf("in block statement: %w", err)
			}
		}

	// Expressions
	case *ast.Identifier:
		err = visitor.VisitIdentifier(n)
	case *ast.IntegerLiteral:
		err = visitor.VisitIntegerLiteral(n)
	case *ast.StringLiteral:
		err = visitor.VisitStringLiteral(n)
	case *ast.BooleanLiteral:
		err = visitor.VisitBooleanLiteral(n)
	case *ast.FunctionCall:
		if err = visitor.VisitFunctionCall(n); err != nil {
			return err
		}
		for _, arg := range n.Arguments {
			if err = Walk(arg, visitor); err != nil {
				return fmt.Errorf("in function call argument: %w", err)
			}
		}
	case *ast.BinaryExpression:
		if err = visitor.VisitBinaryExpression(n); err != nil {
			return err
		}
		if err = Walk(n.Left, visitor); err != nil {
			return fmt.Errorf("in binary expression left: %w", err)
		}
		if err = Walk(n.Right, visitor); err != nil {
			return fmt.Errorf("in binary expression right: %w", err)
		}
	case *ast.UnaryExpression:
		if err = visitor.VisitUnaryExpression(n); err != nil {
			return err
		}
		if err = Walk(n.Right, visitor); err != nil {
			return fmt.Errorf("in unary expression right: %w", err)
		}
	case *ast.ArrayLiteral:
		// The visitor's VisitArrayLiteral method is responsible for walking children (elements)
		// and applying rules.
		err = visitor.VisitArrayLiteral(n)
	case *ast.MapLiteral:
		// The visitor's VisitMapLiteral method is responsible for walking children (keys/values)
		// and applying rules.
		err = visitor.VisitMapLiteral(n)
	case *ast.StructLiteral:
		if err = visitor.VisitStructLiteral(n); err != nil {
			return err
		}
		// Walk through struct literal field values
		for _, fieldValue := range n.Fields {
			if err = Walk(fieldValue, visitor); err != nil {
				return fmt.Errorf("in struct literal field value: %w", err)
			}
		}
	default:
		// This case should ideally not be hit if all ast.Node types are covered.
		// It implies a new AST node was added but not handled in Walk.
		// For expression-only nodes that don't have children to walk further,
		// their visit method is called and then err is returned.
		// fmt.Printf("Warning: Unhandled AST node type in Walk: %T\n", n)
	}
	return err
}
