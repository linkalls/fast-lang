package ast

import (
	"bytes"
	"fmt"
	"strings"
	// Placeholder for token package, using ast.Token for now
	// import "github.com/linkalls/zeno-lang/token"
)

// Token is a placeholder for actual token definitions from a lexer/token package.
// It's defined here to avoid circular dependencies until the token package is created.
type Token struct {
	Type    string // e.g., "LET", "IDENT", "INT"
	Literal string // e.g., "let", "myVar", "5"
}

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string // Returns the literal value of the token this node is associated with
	String() string       // For debugging and testing, a string representation of the node
}

// Statement is a subtype of Node representing a statement.
type Statement interface {
	Node
	statementNode() // Marker method for statement types
}

// Expression is a subtype of Node representing an expression.
type Expression interface {
	Node
	expressionNode() // Marker method for expression types
}

// --- Program Node ---

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// --- Literals / Basic Expressions ---

type Identifier struct {
	Token Token // Using ast.Token placeholder
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token Token // Using ast.Token placeholder
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token Token // Using ast.Token placeholder
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token Token // Using ast.Token placeholder
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

type BooleanLiteral struct {
	Token Token // Using ast.Token placeholder
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// --- Statements ---

type LetStatement struct {
	Token   Token // The 'let' or 'mut' token
	Name    *Identifier
	TypeAnn string // Optional type annotation string
	Value   Expression
	Mutable bool
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	if ls.Mutable {
		out.WriteString("mut ")
	} else {
		out.WriteString("let ")
	}
	out.WriteString(ls.Name.String())
	if ls.TypeAnn != "" {
		out.WriteString(": ")
		out.WriteString(ls.TypeAnn)
	}
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type ExpressionStatement struct {
	Token      Token // The first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";" // Add semicolon for statement context in debug
	}
	return ""
}

type BlockStatement struct {
	Token      Token // The '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{\n")
	for _, s := range bs.Statements {
		out.WriteString("\t" + s.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
}

type IfStatement struct {
	Token       Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement // For 'else { block }'. For 'else if', this would be nil and chain would be handled by parser.
	                           // Or, Alternative could be an 'Expression' which is an IfExpression, or a Block.
	                           // For simplicity, let's assume Alternative is a Block for 'else' or nil.
	                           // 'else if' can be modeled by nesting IfStatement in Alternative's Block.
	                           // Simpler: Alternative can be a Statement (IfStatement or BlockStatement)
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if (") // Keep parens for condition in String() for clarity
	out.WriteString(is.Condition.String())
	out.WriteString(") ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

type LoopStatement struct {
	Token Token // The 'loop' token
	Body  *BlockStatement
}

func (ls *LoopStatement) statementNode()       {}
func (ls *LoopStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LoopStatement) String() string {
	return "loop " + ls.Body.String()
}

type WhileStatement struct {
	Token     Token // The 'while' token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while (") // Keep parens for condition in String() for clarity
	out.WriteString(ws.Condition.String())
	out.WriteString(") ")
	out.WriteString(ws.Body.String())
	return out.String()
}

type ForStatement struct {
	Token       Token // The 'for' token
	Initializer Statement   // Optional: e.g., LetStatement or ExpressionStatement (for assignment)
	Condition   Expression  // Optional
	Increment   Expression  // Optional
	Body        *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for (")
	if fs.Initializer != nil {
		// Remove trailing semicolon for String() representation inside for
		s := fs.Initializer.String()
		out.WriteString(strings.TrimRight(s, ";"))
	}
	out.WriteString("; ")
	if fs.Condition != nil {
		out.WriteString(fs.Condition.String())
	}
	out.WriteString("; ")
	if fs.Increment != nil {
		out.WriteString(fs.Increment.String())
	}
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

type AssignmentStatement struct {
	Token Token // The Identifier token (name of variable)
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type PrintStatement struct {
	Token    Token // The 'print' or 'println' token
	Argument Expression
	Newline  bool
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PrintStatement) String() string {
	var out bytes.Buffer
	if ps.Newline {
		out.WriteString("println(")
	} else {
		out.WriteString("print(")
	}
	if ps.Argument != nil {
		out.WriteString(ps.Argument.String())
	}
	out.WriteString(");")
	return out.String()
}

type BreakStatement struct {
	Token Token // The 'break' token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break;" }

type ContinueStatement struct {
	Token Token // The 'continue' token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue;" }

type ReturnStatement struct {
	Token       Token // The 'return' token
	ReturnValue Expression // Optional, can be nil
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString("return")
	if rs.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// --- Expressions (beyond literals) ---

type PrefixExpression struct {
	Token    Token // The prefix token, e.g., "!", "-"
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

type InfixExpression struct {
	Token    Token // The operator token, e.g., "+"
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

type CallExpression struct {
	Token     Token      // The '(' token or function identifier token
	Function  Expression // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}
