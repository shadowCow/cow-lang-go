// Package parsetree defines generic parse tree structures returned by parsers.
// Parse trees mirror the grammatical structure of the input and can be
// converted to language-specific AST representations.
package parsetree

import (
	"fmt"
	"strings"

	"github.com/shadowCow/cow-lang-go/tooling/grammar"
	"github.com/shadowCow/cow-lang-go/tooling/lexer"
)

// ParseTree is the interface for all parse tree nodes.
type ParseTree interface {
	// NodeType returns a string describing the type of this node
	NodeType() string
	// String returns a string representation of the tree (for debugging)
	String() string
}

// TerminalNode represents a leaf node in the parse tree (a matched token).
type TerminalNode struct {
	Token lexer.Token
}

// NodeType returns "Terminal"
func (t *TerminalNode) NodeType() string {
	return "Terminal"
}

// String returns a string representation of the terminal
func (t *TerminalNode) String() string {
	return fmt.Sprintf("Terminal{%s:%q}", t.Token.Type, t.Token.Value)
}

// NonTerminalNode represents an interior node in the parse tree.
// It corresponds to a non-terminal symbol in the grammar and contains
// the children that were matched during parsing.
type NonTerminalNode struct {
	Symbol   grammar.Symbol  // The non-terminal symbol this node represents
	Children []ParseTree     // The child nodes (may be terminals or non-terminals)
}

// NodeType returns "NonTerminal"
func (n *NonTerminalNode) NodeType() string {
	return "NonTerminal"
}

// String returns a string representation of the non-terminal and its children
func (n *NonTerminalNode) String() string {
	if len(n.Children) == 0 {
		return fmt.Sprintf("NonTerminal{%s}", n.Symbol)
	}

	childStrs := make([]string, len(n.Children))
	for i, child := range n.Children {
		childStrs[i] = child.String()
	}
	return fmt.Sprintf("NonTerminal{%s: [%s]}", n.Symbol, strings.Join(childStrs, ", "))
}

// ProgramNode represents the root of a parse tree.
// It contains the top-level parse tree representing the entire program.
type ProgramNode struct {
	Root ParseTree  // The root of the parse tree (usually a NonTerminalNode)
}

// NodeType returns "Program"
func (p *ProgramNode) NodeType() string {
	return "Program"
}

// String returns a string representation of the program
func (p *ProgramNode) String() string {
	return fmt.Sprintf("Program{%s}", p.Root.String())
}

// EmptyNode represents an empty/epsilon production in the parse tree.
// This can occur when a production derives epsilon (empty string).
type EmptyNode struct {
	Symbol grammar.Symbol  // The symbol that derived epsilon
}

// NodeType returns "Empty"
func (e *EmptyNode) NodeType() string {
	return "Empty"
}

// String returns a string representation
func (e *EmptyNode) String() string {
	return fmt.Sprintf("Empty{%s}", e.Symbol)
}
