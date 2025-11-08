// Package converter transforms generic parse trees into Cow-specific AST nodes.
// This is the bridge between the generic tooling and the Cow language implementation.
package converter

import (
	"fmt"
	"strconv"

	"github.com/shadowCow/cow-lang-go/lang/ast"
	"github.com/shadowCow/cow-lang-go/tooling/parsetree"
)

// ParseTreeToAST converts a generic parse tree to a Cow-specific AST.
func ParseTreeToAST(tree *parsetree.ProgramNode) (*ast.Program, error) {
	if tree == nil {
		return nil, fmt.Errorf("nil parse tree")
	}

	// The root should be a non-terminal node representing the program
	rootNonTerminal, ok := tree.Root.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal root, got %T", tree.Root)
	}

	// Convert the root to a program
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	// Process the root based on the Cow grammar
	// Current grammar: Program -> Literal
	if rootNonTerminal.Symbol == "Program" {
		// Program should have one child: Literal
		if len(rootNonTerminal.Children) != 1 {
			return nil, fmt.Errorf("Program node expected 1 child, got %d", len(rootNonTerminal.Children))
		}

		// Convert the child to a statement
		stmt, err := convertToStatement(rootNonTerminal.Children[0])
		if err != nil {
			return nil, err
		}

		program.Statements = append(program.Statements, stmt)
	}

	return program, nil
}

// convertToStatement converts a parse tree node to a Cow statement.
func convertToStatement(node parsetree.ParseTree) (ast.Statement, error) {
	// For now, all statements are expression statements containing literals
	expr, err := convertToExpression(node)
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionStatement{
		Token:      "", // We could extract this from the expression if needed
		Expression: expr,
	}, nil
}

// convertToExpression converts a parse tree node to a Cow expression.
func convertToExpression(node parsetree.ParseTree) (ast.Expression, error) {
	switch n := node.(type) {
	case *parsetree.TerminalNode:
		// Terminal node represents a literal value
		return convertTerminalToExpression(n)

	case *parsetree.NonTerminalNode:
		// Non-terminal node - check what symbol it is
		if n.Symbol == "Literal" {
			// Literal should have one child: a terminal (INT_DECIMAL, INT_HEX, etc.)
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("Literal node expected 1 child, got %d", len(n.Children))
			}
			return convertToExpression(n.Children[0])
		}
		return nil, fmt.Errorf("unknown non-terminal in expression context: %s", n.Symbol)

	case *parsetree.EmptyNode:
		return nil, fmt.Errorf("unexpected empty node in expression context")

	default:
		return nil, fmt.Errorf("unknown parse tree node type: %T", node)
	}
}

// convertTerminalToExpression converts a terminal node to a Cow expression.
// This is where Cow-specific token interpretation happens (INT_DECIMAL, INT_HEX, etc.)
func convertTerminalToExpression(node *parsetree.TerminalNode) (ast.Expression, error) {
	token := node.Token

	switch token.Type {
	case "INT_DECIMAL", "INT_HEX", "INT_BINARY":
		value, err := parseIntLiteral(token.Type, token.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer at line %d, column %d: %w",
				token.Line, token.Column, err)
		}
		return &ast.IntLiteral{
			Token: token.Value,
			Value: value,
		}, nil

	case "FLOAT":
		value, err := parseFloatLiteral(token.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float at line %d, column %d: %w",
				token.Line, token.Column, err)
		}
		return &ast.FloatLiteral{
			Token: token.Value,
			Value: value,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected terminal token type in expression: %s", token.Type)
	}
}

// parseIntLiteral parses an integer literal token value.
// Handles decimal, hexadecimal, and binary formats.
// This is Cow-specific parsing logic.
func parseIntLiteral(tokenType, value string) (int64, error) {
	// Remove underscores (used for readability in literals)
	value = removeUnderscores(value)

	switch tokenType {
	case "INT_DECIMAL":
		return strconv.ParseInt(value, 10, 64)
	case "INT_HEX":
		// Remove "0x" prefix
		if len(value) < 3 {
			return 0, fmt.Errorf("invalid hex literal: %s", value)
		}
		return strconv.ParseInt(value[2:], 16, 64)
	case "INT_BINARY":
		// Remove "0b" prefix
		if len(value) < 3 {
			return 0, fmt.Errorf("invalid binary literal: %s", value)
		}
		return strconv.ParseInt(value[2:], 2, 64)
	default:
		return 0, fmt.Errorf("unknown integer token type: %s", tokenType)
	}
}

// parseFloatLiteral parses a float literal token value.
// This is Cow-specific parsing logic.
func parseFloatLiteral(value string) (float64, error) {
	value = removeUnderscores(value)
	return strconv.ParseFloat(value, 64)
}

// removeUnderscores removes all underscore characters from a string.
func removeUnderscores(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '_' {
			result = append(result, s[i])
		}
	}
	return string(result)
}
