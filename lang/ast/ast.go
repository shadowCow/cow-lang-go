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

// FunctionDef represents a named function definition statement.
// Syntax: fn name(params) { body }
type FunctionDef struct {
	Token      string   // The 'fn' token
	Name       string   // The function name
	Parameters []string // Parameter names
	Body       *Block   // Function body
}

func (fd *FunctionDef) statementNode()       {}
func (fd *FunctionDef) TokenLiteral() string { return fd.Token }

// Block represents a block of statements enclosed in braces.
// Used for function bodies and other block contexts.
type Block struct {
	Token      string      // The '{' token
	Statements []Statement // Statements in the block
}

func (b *Block) statementNode()       {}
func (b *Block) TokenLiteral() string { return b.Token }

// ReturnStatement represents a return statement in a function.
// Syntax: return expression
type ReturnStatement struct {
	Token string     // The 'return' token
	Value Expression // The value to return
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token }

// ForStatement represents a for loop.
// Syntax: for { body } (infinite) or for condition { body } (while-style)
type ForStatement struct {
	Token     string      // The 'for' token
	Condition Expression  // Optional condition (nil for infinite loops)
	Body      *Block      // Loop body
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token }

// BreakStatement represents a break statement to exit a loop.
// Syntax: break
type BreakStatement struct {
	Token string // The 'break' token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token }

// ContinueStatement represents a continue statement to skip to next iteration.
// Syntax: continue
type ContinueStatement struct {
	Token string // The 'continue' token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token }

// FunctionLiteral represents an anonymous function expression.
// Syntax: fn(params) { body }
// Enables first-class functions (assignable to variables, passable as arguments).
type FunctionLiteral struct {
	Token      string   // The 'fn' token
	Parameters []string // Parameter names
	Body       *Block   // Function body
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token }

// ArrayLiteral represents an array literal expression.
// Syntax: [elem1, elem2, ...] or []
type ArrayLiteral struct {
	Token    string       // The '[' token
	Elements []Expression // The array elements
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token }

// IndexAccess represents array/collection indexing.
// Syntax: arr[index]
// The Object will typically be an Identifier or another IndexAccess (for multi-dimensional arrays).
type IndexAccess struct {
	Token  string     // The '[' token
	Object Expression // The array/object being indexed
	Index  Expression // The index expression
}

func (ia *IndexAccess) expressionNode()      {}
func (ia *IndexAccess) TokenLiteral() string { return ia.Token }

// MemberAccess represents accessing a member/method of an object.
// Syntax: obj.member
// Used for array methods like arr.len(), arr.push(item), arr.pop()
type MemberAccess struct {
	Token  string     // The '.' token
	Object Expression // The object being accessed
	Member string     // The member name
}

func (ma *MemberAccess) expressionNode()      {}
func (ma *MemberAccess) TokenLiteral() string { return ma.Token }

// IndexAssignment represents assignment to an array index.
// Syntax: arr[index] = value or arr[i][j] = value
type IndexAssignment struct {
	Token   string       // The identifier token
	Name    string       // The array variable name
	Indices []Expression // The index expressions (one for arr[0], multiple for arr[i][j])
	Value   Expression   // The value to assign
}

func (ia *IndexAssignment) statementNode()       {}
func (ia *IndexAssignment) TokenLiteral() string { return ia.Token }
