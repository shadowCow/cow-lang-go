// Package ast defines the Abstract Syntax Tree node types for the Cow language.
package ast

// Node is the base interface for all AST nodes.
type Node interface {
	// TokenLiteral returns the literal value of the token that produced this node.
	// Useful for debugging and error messages.
	TokenLiteral() string
}

// Statement represents a statement in the program.
// Statements do not produce values (or produce unit/void).
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression in the program.
// Expressions produce values.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST.
// It contains a list of statements that make up the program.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// ExpressionStatement wraps an expression as a statement.
// Used for expressions that are evaluated for their side effects.
type ExpressionStatement struct {
	Token      string     // The first token of the expression
	Expression Expression // The expression being evaluated
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token }

// IntLiteral represents an integer literal.
type IntLiteral struct {
	Token string // The token text (e.g., "42", "0xFF")
	Value int64  // The parsed integer value
}

func (il *IntLiteral) expressionNode()      {}
func (il *IntLiteral) TokenLiteral() string { return il.Token }

// FloatLiteral represents a floating-point literal.
type FloatLiteral struct {
	Token string  // The token text (e.g., "3.14", "1.5e10")
	Value float64 // The parsed float value
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token }

// FunctionCall represents a function call expression.
type FunctionCall struct {
	Token     string       // The function name token
	Name      string       // The function name (e.g., "println")
	Arguments []Expression // The function arguments
}

func (fc *FunctionCall) expressionNode()      {}
func (fc *FunctionCall) TokenLiteral() string { return fc.Token }
