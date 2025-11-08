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
// ProgramRest: NEWLINE ProgramRest2 | ε
func extractProgramRest(node parsetree.ParseTree) ([]ast.Statement, error) {
	switch n := node.(type) {
	case *parsetree.EmptyNode:
		// No more statements (epsilon)
		return []ast.Statement{}, nil

	case *parsetree.NonTerminalNode:
		if n.Symbol != "ProgramRest" {
			return nil, fmt.Errorf("expected ProgramRest, got %s", n.Symbol)
		}

		// Could be empty (0 children) or NEWLINE ProgramRest2 (2 children)
		if len(n.Children) == 0 {
			// Empty - no more statements
			return []ast.Statement{}, nil
		} else if len(n.Children) == 2 {
			// NEWLINE ProgramRest2
			// Skip the NEWLINE token (index 0) and process ProgramRest2 (index 1)
			return extractProgramRest2(n.Children[1])
		}

		return nil, fmt.Errorf("ProgramRest node expected 0 or 2 children, got %d", len(n.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for ProgramRest: %T", node)
	}
}

// extractProgramRest2 extracts remaining statements from a ProgramRest2 node.
// ProgramRest2: Statement ProgramRest | ε
func extractProgramRest2(node parsetree.ParseTree) ([]ast.Statement, error) {
	switch n := node.(type) {
	case *parsetree.EmptyNode:
		// No more statements (epsilon - trailing newline case)
		return []ast.Statement{}, nil

	case *parsetree.NonTerminalNode:
		if n.Symbol != "ProgramRest2" {
			return nil, fmt.Errorf("expected ProgramRest2, got %s", n.Symbol)
		}

		// Could be empty (0 children) or Statement ProgramRest (2 children)
		if len(n.Children) == 0 {
			// Empty - trailing newline
			return []ast.Statement{}, nil
		} else if len(n.Children) == 2 {
			// Statement ProgramRest
			// Convert the statement (index 0)
			stmt, err := convertToStatement(n.Children[0])
			if err != nil {
				return nil, err
			}

			// Recursively extract rest from ProgramRest (index 1)
			restStmts, err := extractProgramRest(n.Children[1])
			if err != nil {
				return nil, err
			}

			return append([]ast.Statement{stmt}, restStmts...), nil
		}

		return nil, fmt.Errorf("ProgramRest2 node expected 0 or 2 children, got %d", len(n.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for ProgramRest2: %T", node)
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
// Grammar: Expression -> Term AddRest
func convertToExpression(node parsetree.ParseTree) (ast.Expression, error) {
	switch n := node.(type) {
	case *parsetree.TerminalNode:
		// Terminal node represents a literal value
		return convertTerminalToExpression(n)

	case *parsetree.NonTerminalNode:
		// Non-terminal node - check what symbol it is
		switch n.Symbol {
		case "Expression":
			// Expression: LogicalOr
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("Expression node expected 1 child, got %d", len(n.Children))
			}
			return convertToExpression(n.Children[0])

		case "LogicalOr":
			// LogicalOr: LogicalAnd LogicalOrRest
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("LogicalOr node expected 2 children, got %d", len(n.Children))
			}
			left, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}
			return processLogicalOrRest(left, n.Children[1])

		case "LogicalAnd":
			// LogicalAnd: Equality LogicalAndRest
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("LogicalAnd node expected 2 children, got %d", len(n.Children))
			}
			left, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}
			return processLogicalAndRest(left, n.Children[1])

		case "Equality":
			// Equality: Comparison EqualityRest
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("Equality node expected 2 children, got %d", len(n.Children))
			}
			left, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}
			return processEqualityRest(left, n.Children[1])

		case "Comparison":
			// Comparison: Arithmetic ComparisonRest
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("Comparison node expected 2 children, got %d", len(n.Children))
			}
			left, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}
			return processComparisonRest(left, n.Children[1])

		case "Arithmetic":
			// Arithmetic: Term AddRest
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("Arithmetic node expected 2 children, got %d", len(n.Children))
			}
			left, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}
			return processAddRest(left, n.Children[1])

		case "Term":
			// Term: Unary MulRest
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("Term node expected 2 children, got %d", len(n.Children))
			}
			left, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}
			return processMulRest(left, n.Children[1])

		case "Unary":
			// Unary: UnaryOp Unary | Primary
			return convertUnary(n)

		case "Primary":
			// Primary: IDENTIFIER PrimaryRest | Literal | LPAREN Expression RPAREN
			return convertPrimary(n)

		case "Literal":
			// Literal should have one child: a terminal
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

// convertPrimary converts a Primary node to an expression.
// Primary: IDENTIFIER PrimaryRest | Literal | LPAREN Expression RPAREN
func convertPrimary(node *parsetree.NonTerminalNode) (ast.Expression, error) {
	if node.Symbol != "Primary" {
		return nil, fmt.Errorf("expected Primary, got %s", node.Symbol)
	}

	if len(node.Children) == 0 {
		return nil, fmt.Errorf("Primary node has no children")
	}

	// Check first child to determine which alternative
	switch firstChild := node.Children[0].(type) {
	case *parsetree.TerminalNode:
		// Could be IDENTIFIER or LPAREN
		if firstChild.Token.Type == "IDENTIFIER" {
			// IDENTIFIER PrimaryRest
			if len(node.Children) != 2 {
				return nil, fmt.Errorf("Primary IDENTIFIER variant expected 2 children, got %d", len(node.Children))
			}
			name := firstChild.Token.Value
			token := firstChild.Token.Value
			return convertIdentifierPrimary(name, token, node.Children[1])
		} else if firstChild.Token.Type == "LPAREN" {
			// LPAREN Expression RPAREN
			if len(node.Children) != 3 {
				return nil, fmt.Errorf("Primary LPAREN variant expected 3 children, got %d", len(node.Children))
			}
			return convertToExpression(node.Children[1])
		}
		return nil, fmt.Errorf("unexpected terminal in Primary: %s", firstChild.Token.Type)

	case *parsetree.NonTerminalNode:
		// Literal
		if len(node.Children) != 1 {
			return nil, fmt.Errorf("Primary Literal variant expected 1 child, got %d", len(node.Children))
		}
		return convertToExpression(firstChild)

	default:
		return nil, fmt.Errorf("unexpected first child type in Primary: %T", firstChild)
	}
}

// convertUnary converts a Unary node to an expression.
// Unary: UnaryOp Unary | Primary
func convertUnary(node *parsetree.NonTerminalNode) (ast.Expression, error) {
	if node.Symbol != "Unary" {
		return nil, fmt.Errorf("expected Unary, got %s", node.Symbol)
	}

	if len(node.Children) == 0 {
		return nil, fmt.Errorf("Unary node has no children")
	}

	// Check if first child is UnaryOp or Primary
	switch firstChild := node.Children[0].(type) {
	case *parsetree.NonTerminalNode:
		if firstChild.Symbol == "UnaryOp" {
			// UnaryOp Unary
			if len(node.Children) != 2 {
				return nil, fmt.Errorf("Unary UnaryOp variant expected 2 children, got %d", len(node.Children))
			}

			// Extract the operator
			if len(firstChild.Children) != 1 {
				return nil, fmt.Errorf("UnaryOp expected 1 child, got %d", len(firstChild.Children))
			}
			opToken, ok := firstChild.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("UnaryOp child expected to be terminal, got %T", firstChild.Children[0])
			}
			operator := opToken.Token.Type

			// Convert the operand recursively
			operand, err := convertToExpression(node.Children[1])
			if err != nil {
				return nil, err
			}

			return &ast.UnaryExpression{
				Token:    opToken.Token.Value,
				Operator: operator,
				Operand:  operand,
			}, nil
		} else {
			// Primary
			if len(node.Children) != 1 {
				return nil, fmt.Errorf("Unary Primary variant expected 1 child, got %d", len(node.Children))
			}
			return convertToExpression(firstChild)
		}

	default:
		return nil, fmt.Errorf("unexpected first child type in Unary: %T", firstChild)
	}
}

// convertIdentifierPrimary converts an identifier with PrimaryRest.
// If PrimaryRest is empty, it's an Identifier.
// If PrimaryRest has LPAREN, it's a FunctionCall.
func convertIdentifierPrimary(name, token string, primaryRest parsetree.ParseTree) (ast.Expression, error) {
	switch rest := primaryRest.(type) {
	case *parsetree.EmptyNode:
		// PrimaryRest is ε, so this is just an identifier
		return &ast.Identifier{
			Token: token,
			Name:  name,
		}, nil

	case *parsetree.NonTerminalNode:
		if rest.Symbol != "PrimaryRest" {
			return nil, fmt.Errorf("expected PrimaryRest, got %s", rest.Symbol)
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
		return nil, fmt.Errorf("PrimaryRest node expected 0 or 3 children, got %d", len(rest.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for PrimaryRest: %T", primaryRest)
	}
}

// processLogicalOrRest processes a LogicalOrRest node and builds left-associative binary expressions.
// LogicalOrRest: OR LogicalAnd LogicalOrRest | ε
func processLogicalOrRest(left ast.Expression, rest parsetree.ParseTree) (ast.Expression, error) {
	return processBinaryOpRest(left, rest, "LogicalOrRest", "OR")
}

// processLogicalAndRest processes a LogicalAndRest node and builds left-associative binary expressions.
// LogicalAndRest: AND Equality LogicalAndRest | ε
func processLogicalAndRest(left ast.Expression, rest parsetree.ParseTree) (ast.Expression, error) {
	return processBinaryOpRest(left, rest, "LogicalAndRest", "AND")
}

// processEqualityRest processes an EqualityRest node and builds left-associative binary expressions.
// EqualityRest: EqualityOp Comparison EqualityRest | ε
func processEqualityRest(left ast.Expression, rest parsetree.ParseTree) (ast.Expression, error) {
	return processBinaryOpRest(left, rest, "EqualityRest", "")
}

// processComparisonRest processes a ComparisonRest node and builds left-associative binary expressions.
// ComparisonRest: ComparisonOp Arithmetic ComparisonRest | ε
func processComparisonRest(left ast.Expression, rest parsetree.ParseTree) (ast.Expression, error) {
	return processBinaryOpRest(left, rest, "ComparisonRest", "")
}

// processBinaryOpRest is a generic function for processing "Rest" nodes in the grammar.
// It handles patterns like: Op Operand Rest | ε
func processBinaryOpRest(left ast.Expression, rest parsetree.ParseTree, expectedSymbol, fixedOp string) (ast.Expression, error) {
	switch r := rest.(type) {
	case *parsetree.EmptyNode:
		return left, nil

	case *parsetree.NonTerminalNode:
		if string(r.Symbol) != expectedSymbol {
			return nil, fmt.Errorf("expected %s, got %s", expectedSymbol, r.Symbol)
		}

		if len(r.Children) == 0 {
			return left, nil
		}

		if len(r.Children) != 3 {
			return nil, fmt.Errorf("%s node expected 0 or 3 children, got %d", expectedSymbol, len(r.Children))
		}

		// Extract operator (child 0)
		var operator string
		if fixedOp != "" {
			operator = fixedOp
		} else {
			var err error
			operator, err = extractOperator(r.Children[0])
			if err != nil {
				return nil, err
			}
		}

		// Convert right operand (child 1)
		right, err := convertToExpression(r.Children[1])
		if err != nil {
			return nil, err
		}

		// Build binary expression
		binaryExpr := &ast.BinaryExpression{
			Token:    operator,
			Left:     left,
			Operator: operator,
			Right:    right,
		}

		// Process remaining rest (child 2) - builds left-associativity
		return processBinaryOpRest(binaryExpr, r.Children[2], expectedSymbol, fixedOp)

	default:
		return nil, fmt.Errorf("unexpected node type for %s: %T", expectedSymbol, rest)
	}
}

// processAddRest processes an AddRest node and builds left-associative binary expressions.
// AddRest: AddOp Term AddRest | ε
func processAddRest(left ast.Expression, addRest parsetree.ParseTree) (ast.Expression, error) {
	switch rest := addRest.(type) {
	case *parsetree.EmptyNode:
		// No more operations - return the left expression as-is
		return left, nil

	case *parsetree.NonTerminalNode:
		if rest.Symbol != "AddRest" {
			return nil, fmt.Errorf("expected AddRest, got %s", rest.Symbol)
		}

		if len(rest.Children) == 0 {
			// Epsilon - no more operations
			return left, nil
		}

		if len(rest.Children) != 3 {
			return nil, fmt.Errorf("AddRest node expected 0 or 3 children, got %d", len(rest.Children))
		}

		// AddOp Term AddRest
		// Child 0: AddOp
		// Child 1: Term
		// Child 2: AddRest

		// Extract operator
		operator, err := extractOperator(rest.Children[0])
		if err != nil {
			return nil, err
		}

		// Convert right term
		rightTerm, err := convertToExpression(rest.Children[1])
		if err != nil {
			return nil, err
		}

		// Build binary expression: left op rightTerm
		binaryExpr := &ast.BinaryExpression{
			Token:    operator,
			Left:     left,
			Operator: operator,
			Right:    rightTerm,
		}

		// Process remaining AddRest (builds left-associativity)
		return processAddRest(binaryExpr, rest.Children[2])

	default:
		return nil, fmt.Errorf("unexpected node type for AddRest: %T", addRest)
	}
}

// processMulRest processes a MulRest node and builds left-associative binary expressions.
// MulRest: MulOp Factor MulRest | ε
func processMulRest(left ast.Expression, mulRest parsetree.ParseTree) (ast.Expression, error) {
	switch rest := mulRest.(type) {
	case *parsetree.EmptyNode:
		// No more operations - return the left expression as-is
		return left, nil

	case *parsetree.NonTerminalNode:
		if rest.Symbol != "MulRest" {
			return nil, fmt.Errorf("expected MulRest, got %s", rest.Symbol)
		}

		if len(rest.Children) == 0 {
			// Epsilon - no more operations
			return left, nil
		}

		if len(rest.Children) != 3 {
			return nil, fmt.Errorf("MulRest node expected 0 or 3 children, got %d", len(rest.Children))
		}

		// MulOp Factor MulRest
		// Child 0: MulOp
		// Child 1: Factor
		// Child 2: MulRest

		// Extract operator
		operator, err := extractOperator(rest.Children[0])
		if err != nil {
			return nil, err
		}

		// Convert right factor
		rightFactor, err := convertToExpression(rest.Children[1])
		if err != nil {
			return nil, err
		}

		// Build binary expression: left op rightFactor
		binaryExpr := &ast.BinaryExpression{
			Token:    operator,
			Left:     left,
			Operator: operator,
			Right:    rightFactor,
		}

		// Process remaining MulRest (builds left-associativity)
		return processMulRest(binaryExpr, rest.Children[2])

	default:
		return nil, fmt.Errorf("unexpected node type for MulRest: %T", mulRest)
	}
}

// extractOperator extracts the operator string from an AddOp or MulOp node.
func extractOperator(opNode parsetree.ParseTree) (string, error) {
	switch n := opNode.(type) {
	case *parsetree.NonTerminalNode:
		// AddOp or MulOp should have one child: a terminal operator token
		if len(n.Children) != 1 {
			return "", fmt.Errorf("operator node expected 1 child, got %d", len(n.Children))
		}
		terminal, ok := n.Children[0].(*parsetree.TerminalNode)
		if !ok {
			return "", fmt.Errorf("expected terminal for operator, got %T", n.Children[0])
		}
		return terminal.Token.Value, nil

	case *parsetree.TerminalNode:
		return n.Token.Value, nil

	default:
		return "", fmt.Errorf("unexpected node type for operator: %T", opNode)
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

	case "TRUE":
		return &ast.BoolLiteral{
			Token: token.Value,
			Value: true,
		}, nil

	case "FALSE":
		return &ast.BoolLiteral{
			Token: token.Value,
			Value: false,
		}, nil

	case "STRING":
		value, err := parseStringLiteral(token.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse string at line %d, column %d: %w",
				token.Line, token.Column, err)
		}
		return &ast.StringLiteral{
			Token: token.Value,
			Value: value,
		}, nil

	case "RAW_STRING":
		value, err := parseRawStringLiteral(token.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse raw string at line %d, column %d: %w",
				token.Line, token.Column, err)
		}
		return &ast.StringLiteral{
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

// parseStringLiteral parses a regular string literal, processing escape sequences.
// The input value includes the surrounding quotes (e.g., "hello\nworld").
// Returns the string content with escape sequences resolved.
func parseStringLiteral(value string) (string, error) {
	// Remove surrounding quotes
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", fmt.Errorf("invalid string literal: %s", value)
	}
	content := value[1 : len(value)-1]

	// Process escape sequences
	result := make([]byte, 0, len(content))
	for i := 0; i < len(content); i++ {
		if content[i] == '\\' && i+1 < len(content) {
			// Process escape sequence
			switch content[i+1] {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			default:
				return "", fmt.Errorf("invalid escape sequence: \\%c", content[i+1])
			}
			i++ // Skip the next character
		} else {
			result = append(result, content[i])
		}
	}
	return string(result), nil
}

// parseRawStringLiteral parses a raw string literal (backtick-delimited).
// The input value includes the surrounding backticks (e.g., `hello\nworld`).
// Returns the string content as-is, without processing escape sequences.
func parseRawStringLiteral(value string) (string, error) {
	// Remove surrounding backticks
	if len(value) < 2 || value[0] != '`' || value[len(value)-1] != '`' {
		return "", fmt.Errorf("invalid raw string literal: %s", value)
	}
	return value[1 : len(value)-1], nil
}
