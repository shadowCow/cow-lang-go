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
	// Current grammar: Program -> Statement ProgramRest
	if rootNonTerminal.Symbol == "Program" {
		// Program should have two children: Statement and ProgramRest
		if len(rootNonTerminal.Children) != 2 {
			return nil, fmt.Errorf("Program node expected 2 children, got %d", len(rootNonTerminal.Children))
		}

		// Convert the first statement
		stmt, err := convertToStatement(rootNonTerminal.Children[0])
		if err != nil {
			return nil, err
		}
		program.Statements = append(program.Statements, stmt)

		// Extract remaining statements from ProgramRest
		restStmts, err := extractProgramRest(rootNonTerminal.Children[1])
		if err != nil {
			return nil, err
		}
		program.Statements = append(program.Statements, restStmts...)
	}

	return program, nil
}

// extractProgramRest extracts remaining statements from a ProgramRest node.
// ProgramRest: Statement ProgramRest | ε
func extractProgramRest(node parsetree.ParseTree) ([]ast.Statement, error) {
	switch n := node.(type) {
	case *parsetree.EmptyNode:
		// No more statements (epsilon)
		return []ast.Statement{}, nil

	case *parsetree.NonTerminalNode:
		if n.Symbol != "ProgramRest" {
			return nil, fmt.Errorf("expected ProgramRest, got %s", n.Symbol)
		}

		// Could be empty (0 children) or Statement ProgramRest (2 children)
		if len(n.Children) == 0 {
			// Empty - no more statements
			return []ast.Statement{}, nil
		} else if len(n.Children) == 2 {
			// Statement ProgramRest
			// Convert the statement (index 0)
			stmt, err := convertToStatement(n.Children[0])
			if err != nil {
				return nil, err
			}

			// Recursively extract rest
			restStmts, err := extractProgramRest(n.Children[1])
			if err != nil {
				return nil, err
			}

			return append([]ast.Statement{stmt}, restStmts...), nil
		}

		return nil, fmt.Errorf("ProgramRest node expected 0 or 2 children, got %d", len(n.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for ProgramRest: %T", node)
	}
}

// convertToStatement converts a parse tree node to a Cow statement.
// Grammar: Statement -> LetStatement | ExpressionStatement
func convertToStatement(node parsetree.ParseTree) (ast.Statement, error) {
	switch n := node.(type) {
	case *parsetree.NonTerminalNode:
		switch n.Symbol {
		case "Statement":
			// Statement node should have one child: LetStatement or ExpressionStatement
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("Statement node expected 1 child, got %d", len(n.Children))
			}
			return convertToStatement(n.Children[0])

		case "LetStatement":
			// LetStatement: LET IDENTIFIER EQUALS Expression
			if len(n.Children) != 4 {
				return nil, fmt.Errorf("LetStatement node expected 4 children, got %d", len(n.Children))
			}

			// Extract identifier (child 1)
			identifierNode, ok := n.Children[1].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for identifier, got %T", n.Children[1])
			}
			name := identifierNode.Token.Value

			// Extract value expression (child 3)
			valueExpr, err := convertToExpression(n.Children[3])
			if err != nil {
				return nil, err
			}

			// Get LET token for Token field (child 0)
			letNode, ok := n.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for let keyword, got %T", n.Children[0])
			}

			return &ast.LetStatement{
				Token: letNode.Token.Value,
				Name:  name,
				Value: valueExpr,
			}, nil

		case "ExpressionStatement":
			// ExpressionStatement: Expression
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("ExpressionStatement node expected 1 child, got %d", len(n.Children))
			}

			expr, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}

			return &ast.ExpressionStatement{
				Token:      "", // Could extract from expression if needed
				Expression: expr,
			}, nil

		default:
			return nil, fmt.Errorf("unexpected non-terminal in statement context: %s", n.Symbol)
		}

	default:
		return nil, fmt.Errorf("expected non-terminal for statement, got %T", node)
	}
}

// convertToExpression converts a parse tree node to a Cow expression.
// Grammar: Expression -> IDENTIFIER ExpressionRest | Literal
func convertToExpression(node parsetree.ParseTree) (ast.Expression, error) {
	switch n := node.(type) {
	case *parsetree.TerminalNode:
		// Terminal node represents a literal value
		return convertTerminalToExpression(n)

	case *parsetree.NonTerminalNode:
		// Non-terminal node - check what symbol it is
		switch n.Symbol {
		case "Expression":
			// Expression: IDENTIFIER ExpressionRest | Literal
			if len(n.Children) == 2 {
				// IDENTIFIER ExpressionRest
				// Child 0: IDENTIFIER terminal
				// Child 1: ExpressionRest non-terminal
				identifierNode, ok := n.Children[0].(*parsetree.TerminalNode)
				if !ok {
					return nil, fmt.Errorf("expected terminal for identifier, got %T", n.Children[0])
				}
				identifierName := identifierNode.Token.Value

				// Check ExpressionRest to determine if it's a function call or identifier
				expressionRest := n.Children[1]
				return convertIdentifierExpression(identifierName, identifierNode.Token.Value, expressionRest)

			} else if len(n.Children) == 1 {
				// Literal
				return convertToExpression(n.Children[0])
			}
			return nil, fmt.Errorf("Expression node expected 1 or 2 children, got %d", len(n.Children))

		case "Literal":
			// Literal should have one child: a terminal (INT_DECIMAL, INT_HEX, etc.)
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("Literal node expected 1 child, got %d", len(n.Children))
			}
			return convertToExpression(n.Children[0])

		default:
			return nil, fmt.Errorf("unknown non-terminal in expression context: %s", n.Symbol)
		}

	case *parsetree.EmptyNode:
		return nil, fmt.Errorf("unexpected empty node in expression context")

	default:
		return nil, fmt.Errorf("unknown parse tree node type: %T", node)
	}
}

// convertIdentifierExpression converts an identifier with ExpressionRest.
// If ExpressionRest is empty, it's an Identifier.
// If ExpressionRest has LPAREN, it's a FunctionCall.
func convertIdentifierExpression(name, token string, expressionRest parsetree.ParseTree) (ast.Expression, error) {
	switch rest := expressionRest.(type) {
	case *parsetree.EmptyNode:
		// ExpressionRest is ε, so this is just an identifier
		return &ast.Identifier{
			Token: token,
			Name:  name,
		}, nil

	case *parsetree.NonTerminalNode:
		if rest.Symbol != "ExpressionRest" {
			return nil, fmt.Errorf("expected ExpressionRest, got %s", rest.Symbol)
		}

		// Check if it's empty (0 children) or LPAREN Arguments RPAREN (3 children)
		if len(rest.Children) == 0 {
			// Empty - just an identifier
			return &ast.Identifier{
				Token: token,
				Name:  name,
			}, nil
		} else if len(rest.Children) == 3 {
			// LPAREN Arguments RPAREN - function call
			// Extract arguments from child 1 (Arguments)
			arguments, err := extractArguments(rest.Children[1])
			if err != nil {
				return nil, err
			}

			return &ast.FunctionCall{
				Token:     token,
				Name:      name,
				Arguments: arguments,
			}, nil
		}
		return nil, fmt.Errorf("ExpressionRest node expected 0 or 3 children, got %d", len(rest.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for ExpressionRest: %T", expressionRest)
	}
}

// extractArguments extracts function arguments from an Arguments parse tree node.
// Returns a slice of expressions representing the arguments.
func extractArguments(node parsetree.ParseTree) ([]ast.Expression, error) {
	switch n := node.(type) {
	case *parsetree.EmptyNode:
		// No arguments (epsilon production)
		return []ast.Expression{}, nil

	case *parsetree.NonTerminalNode:
		if n.Symbol == "Arguments" {
			// Arguments node should have one child: ArgumentList or empty
			if len(n.Children) == 0 {
				// Empty arguments
				return []ast.Expression{}, nil
			}
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("Arguments node expected 0 or 1 child, got %d", len(n.Children))
			}
			return extractArguments(n.Children[0])
		} else if n.Symbol == "ArgumentList" {
			// ArgumentList is either:
			// - Single expression: Expression
			// - Multiple: Expression COMMA ArgumentList
			return extractArgumentList(n)
		}
		return nil, fmt.Errorf("unexpected non-terminal in arguments: %s", n.Symbol)

	default:
		return nil, fmt.Errorf("unexpected node type in arguments: %T", node)
	}
}

// extractArgumentList recursively extracts arguments from an ArgumentList node.
// ArgumentList: Expression ArgumentRest
func extractArgumentList(node *parsetree.NonTerminalNode) ([]ast.Expression, error) {
	if node.Symbol != "ArgumentList" {
		return nil, fmt.Errorf("expected ArgumentList, got %s", node.Symbol)
	}

	// ArgumentList: Expression ArgumentRest (2 children)
	if len(node.Children) != 2 {
		return nil, fmt.Errorf("ArgumentList node expected 2 children, got %d", len(node.Children))
	}

	// Extract first expression
	firstExpr, err := convertToExpression(node.Children[0])
	if err != nil {
		return nil, err
	}

	// Extract rest from ArgumentRest
	restExprs, err := extractArgumentRest(node.Children[1])
	if err != nil {
		return nil, err
	}

	// Combine
	return append([]ast.Expression{firstExpr}, restExprs...), nil
}

// extractArgumentRest extracts remaining arguments from an ArgumentRest node.
// ArgumentRest: COMMA Expression ArgumentRest | ε
func extractArgumentRest(node parsetree.ParseTree) ([]ast.Expression, error) {
	switch n := node.(type) {
	case *parsetree.EmptyNode:
		// No more arguments (epsilon)
		return []ast.Expression{}, nil

	case *parsetree.NonTerminalNode:
		if n.Symbol != "ArgumentRest" {
			return nil, fmt.Errorf("expected ArgumentRest, got %s", n.Symbol)
		}

		// Could be empty (0 children) or COMMA Expression ArgumentRest (3 children)
		if len(n.Children) == 0 {
			// Empty - no more arguments
			return []ast.Expression{}, nil
		} else if len(n.Children) == 3 {
			// COMMA Expression ArgumentRest
			// Extract the expression (skip COMMA at index 0)
			expr, err := convertToExpression(n.Children[1])
			if err != nil {
				return nil, err
			}

			// Recursively extract rest
			restExprs, err := extractArgumentRest(n.Children[2])
			if err != nil {
				return nil, err
			}

			return append([]ast.Expression{expr}, restExprs...), nil
		}

		return nil, fmt.Errorf("ArgumentRest node expected 0 or 3 children, got %d", len(n.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for ArgumentRest: %T", node)
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
