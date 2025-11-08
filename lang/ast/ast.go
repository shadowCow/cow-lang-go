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

// LetStatement represents a variable declaration with initialization.
// Syntax: let <name> = <value>
type LetStatement struct {
	Token string     // The 'let' token
	Name  string     // The variable name
	Value Expression // The initialization expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token }

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

// BoolLiteral represents a boolean literal (true or false).
type BoolLiteral struct {
	Token string // The token text ("true" or "false")
	Value bool   // The boolean value
}

func (bl *BoolLiteral) expressionNode()      {}
func (bl *BoolLiteral) TokenLiteral() string { return bl.Token }

// StringLiteral represents a string literal.
// For regular strings ("..."), escape sequences are processed.
// For raw strings (`...`), the value is taken as-is.
type StringLiteral struct {
	Token string // The token text (e.g., "hello", `world`)
	Value string // The processed string value
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token }

// FunctionCall represents a function call expression.
type FunctionCall struct {
	Token     string       // The function name token
	Name      string       // The function name (e.g., "println")
	Arguments []Expression // The function arguments
}

func (fc *FunctionCall) expressionNode()      {}
func (fc *FunctionCall) TokenLiteral() string { return fc.Token }

// Identifier represents a variable reference in an expression.
type Identifier struct {
	Token string // The identifier token
	Name  string // The variable name
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token }

// BinaryExpression represents a binary operation.
// Handles arithmetic (+, -, *, /, %), comparison (<, >, <=, >=),
// equality (==, !=), and logical (&&, ||) operators.
type BinaryExpression struct {
	Token    string     // The operator token
	Left     Expression // The left operand
	Operator string     // The operator
	Right    Expression // The right operand
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token }

// UnaryExpression represents a unary operation (e.g., !true, -5).
type UnaryExpression struct {
	Token    string     // The operator token
	Operator string     // The operator (!, -)
	Operand  Expression // The operand
}

func (ue *UnaryExpression) expressionNode()      {}
func (ue *UnaryExpression) TokenLiteral() string { return ue.Token }
