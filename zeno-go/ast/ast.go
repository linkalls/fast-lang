package ast

import "fmt"

// Node represents any node in the AST
type Node interface {
	String() string
}

// Statement represents all statement nodes
type Statement interface {
	Node
	statementNode()
}

// Expression represents all expression nodes
type Expression interface {
	Node
	expressionNode()
}

// Program represents the root node of the AST
type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	result := ""
	for _, stmt := range p.Statements {
		result += stmt.String()
	}
	return result
}

// BinaryOperator represents binary operators
type BinaryOperator int

const (
	BinaryOpPlus BinaryOperator = iota
	BinaryOpMinus
	BinaryOpMultiply
	BinaryOpDivide
	BinaryOpModulo
	BinaryOpEq
	BinaryOpNotEq
	BinaryOpLt
	BinaryOpLte
	BinaryOpGt
	BinaryOpGte
	BinaryOpAnd
	BinaryOpOr
)

func (op BinaryOperator) String() string {
	switch op {
	case BinaryOpPlus:
		return "+"
	case BinaryOpMinus:
		return "-"
	case BinaryOpMultiply:
		return "*"
	case BinaryOpDivide:
		return "/"
	case BinaryOpModulo:
		return "%"
	case BinaryOpEq:
		return "=="
	case BinaryOpNotEq:
		return "!="
	case BinaryOpLt:
		return "<"
	case BinaryOpLte:
		return "<="
	case BinaryOpGt:
		return ">"
	case BinaryOpGte:
		return ">="
	case BinaryOpAnd:
		return "&&"
	case BinaryOpOr:
		return "||"
	default:
		return "UNKNOWN"
	}
}

// UnaryOperator represents unary operators
type UnaryOperator int

const (
	UnaryOpMinus UnaryOperator = iota
	UnaryOpBang
)

func (op UnaryOperator) String() string {
	switch op {
	case UnaryOpMinus:
		return "-"
	case UnaryOpBang:
		return "!"
	default:
		return "UNKNOWN"
	}
}

// LetDeclaration represents let declarations
type LetDeclaration struct {
	Name            string
	TypeAnn         *string // Optional type annotation
	ValueExpression Expression
}

func (ld *LetDeclaration) statementNode() {}
func (ld *LetDeclaration) String() string {
	result := "let " + ld.Name
	if ld.TypeAnn != nil {
		result += ": " + *ld.TypeAnn
	}
	result += " = " + ld.ValueExpression.String()
	return result
}

// AssignmentStatement represents assignment statements (x = value)
type AssignmentStatement struct {
	Name  string     // Variable name being assigned to
	Value Expression // Value being assigned
}

func (as *AssignmentStatement) statementNode() {}
func (as *AssignmentStatement) String() string {
	return as.Name + " = " + as.Value.String()
}

// ExpressionStatement represents expression statements
type ExpressionStatement struct {
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) String() string {
	return es.Expression.String()
}

// IntegerLiteral represents integer literals
type IntegerLiteral struct {
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", il.Value)
}

// StringLiteral represents string literals
type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) String() string {
	return "\"" + sl.Value + "\""
}

// BooleanLiteral represents boolean literals
type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}
func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "true"
	}
	return "false"
}

// Identifier represents identifiers
type Identifier struct {
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string {
	return i.Value
}

// BinaryExpression represents binary expressions
type BinaryExpression struct {
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) String() string {
	return "(" + be.Left.String() + " " + be.Operator.String() + " " + be.Right.String() + ")"
}

// UnaryExpression represents unary expressions
type UnaryExpression struct {
	Operator UnaryOperator
	Right    Expression
}

func (ue *UnaryExpression) expressionNode() {}
func (ue *UnaryExpression) String() string {
	return "(" + ue.Operator.String() + ue.Right.String() + ")"
}

// ImportStatement represents import statements
type ImportStatement struct {
	Imports []string // List of imported identifiers
	Module  string   // Module name to import from
}

func (is *ImportStatement) statementNode() {}
func (is *ImportStatement) String() string {
	result := "import {"
	for i, imp := range is.Imports {
		if i > 0 {
			result += ", "
		}
		result += imp
	}
	result += "} from \"" + is.Module + "\""
	return result
}

// PrintStatement represents print and println statements
type PrintStatement struct {
	Arguments []Expression // Arguments to print
	Newline   bool         // true for println, false for print
}

func (ps *PrintStatement) statementNode() {}
func (ps *PrintStatement) String() string {
	var result string
	if ps.Newline {
		result = "println("
	} else {
		result = "print("
	}
	for i, arg := range ps.Arguments {
		if i > 0 {
			result += ", "
		}
		result += arg.String()
	}
	result += ")"
	return result
}

// Parameter represents a function parameter
type Parameter struct {
	Name string
	Type string
}

func (p *Parameter) String() string {
	return p.Name + ": " + p.Type
}

// FunctionDefinition represents function definitions
type FunctionDefinition struct {
	Name       string
	Parameters []Parameter
	ReturnType *string // Optional return type
	Body       []Statement
	IsPublic   bool // Whether the function is public (pub fn)
}

func (fd *FunctionDefinition) statementNode() {}
func (fd *FunctionDefinition) String() string {
	result := "fn " + fd.Name + "("
	for i, param := range fd.Parameters {
		if i > 0 {
			result += ", "
		}
		result += param.String()
	}
	result += ")"
	if fd.ReturnType != nil {
		result += ": " + *fd.ReturnType
	}
	result += " {\n"
	for _, stmt := range fd.Body {
		result += "  " + stmt.String() + "\n"
	}
	result += "}"
	return result
}

// FunctionCall represents function calls
type FunctionCall struct {
	Name      string
	Arguments []Expression
}

func (fc *FunctionCall) expressionNode() {}
func (fc *FunctionCall) String() string {
	result := fc.Name + "("
	for i, arg := range fc.Arguments {
		if i > 0 {
			result += ", "
		}
		result += arg.String()
	}
	result += ")"
	return result
}

// Block represents a block of statements
type Block struct {
	Statements []Statement
}

func (b *Block) String() string {
	result := "{\n"
	for _, stmt := range b.Statements {
		result += "  " + stmt.String() + "\n"
	}
	result += "}"
	return result
}

// IfStatement represents if/else if/else statements
type IfStatement struct {
	Condition     Expression
	ThenBlock     *Block
	ElseIfClauses []ElseIfClause
	ElseBlock     *Block // Optional else block
}

func (ifs *IfStatement) statementNode() {}
func (ifs *IfStatement) String() string {
	result := "if " + ifs.Condition.String() + " " + ifs.ThenBlock.String()

	for _, elseif := range ifs.ElseIfClauses {
		result += " else if " + elseif.Condition.String() + " " + elseif.Block.String()
	}

	if ifs.ElseBlock != nil {
		result += " else " + ifs.ElseBlock.String()
	}

	return result
}

// ElseIfClause represents an else if clause
type ElseIfClause struct {
	Condition Expression
	Block     *Block
}

// ReturnStatement represents return statements
type ReturnStatement struct {
	Value Expression // Optional return value
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) String() string {
	if rs.Value != nil {
		return "return " + rs.Value.String() + ""
	}
	return "return"
}

// WhileStatement represents while loops
type WhileStatement struct {
	Condition Expression
	Block     *Block
}

func (ws *WhileStatement) statementNode() {}
func (ws *WhileStatement) String() string {
	return "while " + ws.Condition.String() + " " + ws.Block.String()
}
