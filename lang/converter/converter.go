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
	// Current grammar: Program -> TopLevelItem TopLevelItemRest
	if rootNonTerminal.Symbol == "Program" {
		// Program should have two children: TopLevelItem and TopLevelItemRest
		if len(rootNonTerminal.Children) != 2 {
			return nil, fmt.Errorf("Program node expected 2 children, got %d", len(rootNonTerminal.Children))
		}

		// Convert the first top-level item
		stmt, err := convertTopLevelItem(rootNonTerminal.Children[0])
		if err != nil {
			return nil, err
		}
		program.Statements = append(program.Statements, stmt)

		// Extract remaining items from TopLevelItemRest
		restStmts, err := extractTopLevelItemRest(rootNonTerminal.Children[1])
		if err != nil {
			return nil, err
		}
		program.Statements = append(program.Statements, restStmts...)
	}

	return program, nil
}

// convertTopLevelItem converts a TopLevelItem to a statement.
// TopLevelItem: FunctionDef | Statement
func convertTopLevelItem(node parsetree.ParseTree) (ast.Statement, error) {
	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for top-level item, got %T", node)
	}

	if nonTerminal.Symbol != "TopLevelItem" {
		return nil, fmt.Errorf("expected TopLevelItem, got %s", nonTerminal.Symbol)
	}

	// TopLevelItem should have one child: FunctionDef, LetStatement, or TopLevelExpression
	if len(nonTerminal.Children) != 1 {
		return nil, fmt.Errorf("TopLevelItem expected 1 child, got %d", len(nonTerminal.Children))
	}

	child := nonTerminal.Children[0]

	// Check if it's TopLevelExpression (needs to be wrapped in ExpressionStatement or converted to IndexAssignment)
	if childNonTerm, ok := child.(*parsetree.NonTerminalNode); ok {
		if childNonTerm.Symbol == "TopLevelExpression" {
			// TopLevelExpression has one child: Assignment
			// Assignment has: LogicalOr AssignmentRest
			assignmentNode := childNonTerm.Children[0]
			if assignNonTerm, ok := assignmentNode.(*parsetree.NonTerminalNode); ok && assignNonTerm.Symbol == "Assignment" {
				if len(assignNonTerm.Children) == 2 {
					// Check AssignmentRest
					assignmentRest := assignNonTerm.Children[1]
					if restNode, ok := assignmentRest.(*parsetree.NonTerminalNode); ok && len(restNode.Children) > 0 {
						// Has assignment: EQUALS Assignment
						// This should be an index assignment: arr[0] = value
						return convertIndexAssignmentFromExpression(assignNonTerm)
					}
				}
			}
			// No assignment, just a regular expression
			expr, err := convertToExpression(childNonTerm.Children[0])
			if err != nil {
				return nil, err
			}
			return &ast.ExpressionStatement{
				Token:      "", // Will be set by evaluator
				Expression: expr,
			}, nil
		}
	}

	return convertToStatement(child)
}

// extractTopLevelItemRest extracts remaining top-level items.
// TopLevelItemRest: NEWLINE TopLevelItemRest2 | ε
func extractTopLevelItemRest(node parsetree.ParseTree) ([]ast.Statement, error) {
	switch n := node.(type) {
	case *parsetree.EmptyNode:
		return []ast.Statement{}, nil

	case *parsetree.NonTerminalNode:
		if n.Symbol != "TopLevelItemRest" {
			return nil, fmt.Errorf("expected TopLevelItemRest, got %s", n.Symbol)
		}

		if len(n.Children) == 0 {
			return []ast.Statement{}, nil
		} else if len(n.Children) == 2 {
			// NEWLINE TopLevelItemRest2
			return extractTopLevelItemRest2(n.Children[1])
		}
		return nil, fmt.Errorf("TopLevelItemRest expected 0 or 2 children, got %d", len(n.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for TopLevelItemRest: %T", node)
	}
}

// extractTopLevelItemRest2 extracts items from TopLevelItemRest2.
// TopLevelItemRest2: TopLevelItem TopLevelItemRest | ε
func extractTopLevelItemRest2(node parsetree.ParseTree) ([]ast.Statement, error) {
	switch n := node.(type) {
	case *parsetree.EmptyNode:
		return []ast.Statement{}, nil

	case *parsetree.NonTerminalNode:
		if n.Symbol != "TopLevelItemRest2" {
			return nil, fmt.Errorf("expected TopLevelItemRest2, got %s", n.Symbol)
		}

		if len(n.Children) == 0 {
			return []ast.Statement{}, nil
		} else if len(n.Children) == 2 {
			// TopLevelItem TopLevelItemRest
			stmt, err := convertTopLevelItem(n.Children[0])
			if err != nil {
				return nil, err
			}
			restStmts, err := extractTopLevelItemRest(n.Children[1])
			if err != nil {
				return nil, err
			}
			return append([]ast.Statement{stmt}, restStmts...), nil
		}
		return nil, fmt.Errorf("TopLevelItemRest2 expected 0 or 2 children, got %d", len(n.Children))

	default:
		return nil, fmt.Errorf("unexpected node type for TopLevelItemRest2: %T", node)
	}
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

			// Check if the expression is actually an assignment (arr[0] = value)
			// Expression -> Assignment -> LogicalOr AssignmentRest
			exprNode := n.Children[0]
			if exprNonTerm, ok := exprNode.(*parsetree.NonTerminalNode); ok {
				if exprNonTerm.Symbol == "Expression" && len(exprNonTerm.Children) == 1 {
					// Unwrap to get Assignment or FunctionLiteral
					innerNode := exprNonTerm.Children[0]
					if innerNonTerm, ok := innerNode.(*parsetree.NonTerminalNode); ok {
						// Check if it's an Assignment sequence
						if len(innerNonTerm.Children) == 1 {
							// It's a SynSequence wrapping Assignment
							assignmentNode := innerNonTerm.Children[0]
							if assignNonTerm, ok := assignmentNode.(*parsetree.NonTerminalNode); ok && assignNonTerm.Symbol == "Assignment" {
								if len(assignNonTerm.Children) == 2 {
									// Check AssignmentRest
									assignmentRest := assignNonTerm.Children[1]
									if restNode, ok := assignmentRest.(*parsetree.NonTerminalNode); ok && len(restNode.Children) > 0 {
										// Has assignment: EQUALS Assignment
										// This is an index assignment: arr[0] = value
										return convertIndexAssignmentFromExpression(assignNonTerm)
									}
								}
							}
						}
					}
				}
			}

			// Not an assignment, convert as regular expression
			expr, err := convertToExpression(n.Children[0])
			if err != nil {
				return nil, err
			}

			return &ast.ExpressionStatement{
				Token:      "", // Could extract from expression if needed
				Expression: expr,
			}, nil

		case "IndexAssignment":
			// IndexAssignment: IDENTIFIER IndexChain EQUALS Expression
			if len(n.Children) != 4 {
				return nil, fmt.Errorf("IndexAssignment node expected 4 children, got %d", len(n.Children))
			}

			// Extract array name (child 0)
			nameNode, ok := n.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for array name, got %T", n.Children[0])
			}

			// Extract index chain (child 1)
			indices, err := extractIndexChain(n.Children[1])
			if err != nil {
				return nil, err
			}

			// Extract value expression (child 3)
			valueExpr, err := convertToExpression(n.Children[3])
			if err != nil {
				return nil, err
			}

			return &ast.IndexAssignment{
				Token:   nameNode.Token.Value,
				Name:    nameNode.Token.Value,
				Indices: indices,
				Value:   valueExpr,
			}, nil

		case "FunctionDef":
			// FunctionDef: FN IDENTIFIER LPAREN ParameterList RPAREN Block
			if len(n.Children) != 6 {
				return nil, fmt.Errorf("FunctionDef node expected 6 children, got %d", len(n.Children))
			}

			// Extract fn token (child 0)
			fnNode, ok := n.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for fn keyword, got %T", n.Children[0])
			}

			// Extract function name (child 1)
			nameNode, ok := n.Children[1].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for function name, got %T", n.Children[1])
			}

			// Extract parameter list (child 3)
			params, err := extractParameterList(n.Children[3])
			if err != nil {
				return nil, err
			}

			// Extract block (child 5)
			block, err := convertBlock(n.Children[5])
			if err != nil {
				return nil, err
			}

			return &ast.FunctionDef{
				Token:      fnNode.Token.Value,
				Name:       nameNode.Token.Value,
				Parameters: params,
				Body:       block,
			}, nil

		case "ReturnStatement":
			// ReturnStatement: RETURN Expression
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("ReturnStatement node expected 2 children, got %d", len(n.Children))
			}

			// Extract return token (child 0)
			returnNode, ok := n.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for return keyword, got %T", n.Children[0])
			}

			// Extract value expression (child 1)
			valueExpr, err := convertToExpression(n.Children[1])
			if err != nil {
				return nil, err
			}

			return &ast.ReturnStatement{
				Token: returnNode.Token.Value,
				Value: valueExpr,
			}, nil

		case "ForStatement":
			// ForStatement: FOR ForCondition Block
			if len(n.Children) != 3 {
				return nil, fmt.Errorf("ForStatement node expected 3 children, got %d", len(n.Children))
			}

			// Extract for token (child 0)
			forNode, ok := n.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for for keyword, got %T", n.Children[0])
			}

			// Extract condition (child 1) - could be epsilon for infinite loop
			var condition ast.Expression
			forCondition := n.Children[1]
			if condNode, ok := forCondition.(*parsetree.NonTerminalNode); ok {
				if len(condNode.Children) > 0 {
					// Has condition
					var err error
					condition, err = convertToExpression(condNode.Children[0])
					if err != nil {
						return nil, fmt.Errorf("error converting for condition: %v", err)
					}
				}
				// else: epsilon, condition remains nil for infinite loop
			}

			// Extract body block (child 2)
			body, err := convertBlock(n.Children[2])
			if err != nil {
				return nil, fmt.Errorf("error converting for body: %v", err)
			}

			return &ast.ForStatement{
				Token:     forNode.Token.Value,
				Condition: condition,
				Body:      body,
			}, nil

		case "BreakStatement":
			// BreakStatement: BREAK
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("BreakStatement node expected 1 child, got %d", len(n.Children))
			}

			// Extract break token (child 0)
			breakNode, ok := n.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for break keyword, got %T", n.Children[0])
			}

			return &ast.BreakStatement{
				Token: breakNode.Token.Value,
			}, nil

		case "ContinueStatement":
			// ContinueStatement: CONTINUE
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("ContinueStatement node expected 1 child, got %d", len(n.Children))
			}

			// Extract continue token (child 0)
			continueNode, ok := n.Children[0].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected terminal for continue keyword, got %T", n.Children[0])
			}

			return &ast.ContinueStatement{
				Token: continueNode.Token.Value,
			}, nil

		case "Block":
			// Block: LBRACE BlockStatements RBRACE
			block, err := convertBlock(n)
			if err != nil {
				return nil, err
			}
			return block, nil

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
		case "Assignment":
			// Assignment: LogicalOr AssignmentRest
			if len(n.Children) != 2 {
				return nil, fmt.Errorf("Assignment node expected 2 children, got %d", len(n.Children))
			}
			// Check if there's an actual assignment (AssignmentRest is not epsilon)
			assignmentRest := n.Children[1]
			if restNode, ok := assignmentRest.(*parsetree.NonTerminalNode); ok {
				if len(restNode.Children) == 0 {
					// Epsilon - no assignment, just return left side
					return convertToExpression(n.Children[0])
				}
				// Has assignment: EQUALS Assignment
				// This is only valid in statement context, not general expressions
				// For now, return an error - ExpressionStatement will handle it specially
				return nil, fmt.Errorf("assignment is only allowed at statement level, not in expressions")
			}
			return convertToExpression(n.Children[0])

		case "Expression":
			// Expression: LogicalOr | FunctionLiteral
			if len(n.Children) != 1 {
				return nil, fmt.Errorf("Expression node expected 1 child, got %d", len(n.Children))
			}
			// The child is either a sequence containing LogicalOr, or FunctionLiteral
			child := n.Children[0]
			if childNonTerm, ok := child.(*parsetree.NonTerminalNode); ok {
				if childNonTerm.Symbol == "FunctionLiteral" {
					return convertFunctionLiteral(childNonTerm)
				}
				// Otherwise it's a sequence containing LogicalOr - unwrap it
				if len(childNonTerm.Children) == 1 {
					return convertToExpression(childNonTerm.Children[0])
				}
			}
			return convertToExpression(child)

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
		// Could be Literal, ArrayLiteral, or FunctionLiteral
		if firstChild.Symbol == "Literal" {
			if len(node.Children) != 1 {
				return nil, fmt.Errorf("Primary Literal variant expected 1 child, got %d", len(node.Children))
			}
			return convertToExpression(firstChild)
		} else if firstChild.Symbol == "ArrayLiteral" {
			if len(node.Children) != 1 {
				return nil, fmt.Errorf("Primary ArrayLiteral variant expected 1 child, got %d", len(node.Children))
			}
			return convertArrayLiteral(firstChild)
		} else if firstChild.Symbol == "FunctionLiteral" {
			// FunctionLiteral
			if len(node.Children) != 1 {
				return nil, fmt.Errorf("Primary FunctionLiteral variant expected 1 child, got %d", len(node.Children))
			}
			return convertFunctionLiteral(firstChild)
		}
		return nil, fmt.Errorf("unexpected non-terminal in Primary: %s", firstChild.Symbol)

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

		// Check if it's empty (0 children) or one of the 3-child variants
		if len(rest.Children) == 0 {
			// Empty - just an identifier
			return &ast.Identifier{
				Token: token,
				Name:  name,
			}, nil
		}

		// PrimaryRest can have 3 or 4 children depending on the variant
		if len(rest.Children) < 3 || len(rest.Children) > 4 {
			return nil, fmt.Errorf("PrimaryRest node expected 0, 3, or 4 children, got %d", len(rest.Children))
		}

		// Determine which variant by checking first child's token type
		firstToken, ok := rest.Children[0].(*parsetree.TerminalNode)
		if !ok {
			return nil, fmt.Errorf("expected terminal as first child of PrimaryRest, got %T", rest.Children[0])
		}

		baseExpr := &ast.Identifier{Token: token, Name: name}

		switch firstToken.Token.Type {
		case "LPAREN":
			// Function call: LPAREN Arguments RPAREN (3 children)
			arguments, err := extractArguments(rest.Children[1])
			if err != nil {
				return nil, err
			}
			return &ast.FunctionCall{
				Token:     token,
				Name:      name,
				Arguments: arguments,
			}, nil

		case "LBRACKET":
			// Index access: LBRACKET Expression RBRACKET PrimaryRest (4 children)
			if len(rest.Children) != 4 {
				return nil, fmt.Errorf("index access expected 4 children, got %d", len(rest.Children))
			}
			indexExpr, err := convertToExpression(rest.Children[1])
			if err != nil {
				return nil, err
			}
			indexAccess := &ast.IndexAccess{
				Token:  "[",
				Object: baseExpr,
				Index:  indexExpr,
			}
			// Process recursive PrimaryRest for chaining like arr[0][1]
			return processPrimaryRest(indexAccess, rest.Children[3])

		case "DOT":
			// Member access: DOT IDENTIFIER PrimaryRest (3 children)
			if len(rest.Children) != 3 {
				return nil, fmt.Errorf("member access expected 3 children, got %d", len(rest.Children))
			}
			memberToken, ok := rest.Children[1].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected IDENTIFIER after DOT, got %T", rest.Children[1])
			}
			memberAccess := &ast.MemberAccess{
				Token:  ".",
				Object: baseExpr,
				Member: memberToken.Token.Value,
			}
			// Process the nested PrimaryRest to allow chaining like arr.len()
			return processPrimaryRest(memberAccess, rest.Children[2])

		default:
			return nil, fmt.Errorf("unexpected token type in PrimaryRest: %s", firstToken.Token.Type)
		}

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

// extractParameterList extracts parameter names from a ParameterList parse tree node.
// ParameterList: ε | IDENTIFIER ParameterRest
func extractParameterList(node parsetree.ParseTree) ([]string, error) {
	// Handle epsilon production
	if _, ok := node.(*parsetree.EmptyNode); ok {
		return []string{}, nil
	}

	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for parameter list, got %T", node)
	}

	if nonTerminal.Symbol != "ParameterList" {
		return nil, fmt.Errorf("expected ParameterList node, got %s", nonTerminal.Symbol)
	}

	// Check if empty (epsilon)
	if len(nonTerminal.Children) == 0 {
		return []string{}, nil
	}

	// ParameterList: IDENTIFIER ParameterRest
	if len(nonTerminal.Children) != 2 {
		return nil, fmt.Errorf("ParameterList node expected 0 or 2 children, got %d", len(nonTerminal.Children))
	}

	// Extract first parameter (child 0)
	firstParam, ok := nonTerminal.Children[0].(*parsetree.TerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected terminal for first parameter, got %T", nonTerminal.Children[0])
	}

	params := []string{firstParam.Token.Value}

	// Extract remaining parameters from ParameterRest
	restParams, err := extractParameterRest(nonTerminal.Children[1])
	if err != nil {
		return nil, err
	}

	return append(params, restParams...), nil
}

// extractParameterRest extracts remaining parameters from ParameterRest node.
// ParameterRest: COMMA IDENTIFIER ParameterRest | ε
func extractParameterRest(node parsetree.ParseTree) ([]string, error) {
	// Handle epsilon production
	if _, ok := node.(*parsetree.EmptyNode); ok {
		return []string{}, nil
	}

	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for parameter rest, got %T", node)
	}

	if nonTerminal.Symbol != "ParameterRest" {
		return nil, fmt.Errorf("expected ParameterRest node, got %s", nonTerminal.Symbol)
	}

	// Check if empty (epsilon)
	if len(nonTerminal.Children) == 0 {
		return []string{}, nil
	}

	// ParameterRest: COMMA IDENTIFIER ParameterRest
	if len(nonTerminal.Children) != 3 {
		return nil, fmt.Errorf("ParameterRest node expected 0 or 3 children, got %d", len(nonTerminal.Children))
	}

	// Extract parameter (child 1)
	param, ok := nonTerminal.Children[1].(*parsetree.TerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected terminal for parameter, got %T", nonTerminal.Children[1])
	}

	params := []string{param.Token.Value}

	// Recursively extract remaining parameters
	restParams, err := extractParameterRest(nonTerminal.Children[2])
	if err != nil {
		return nil, err
	}

	return append(params, restParams...), nil
}

// convertBlock converts a Block parse tree node to an AST Block.
// Block: LBRACE BlockStatements RBRACE
func convertBlock(node parsetree.ParseTree) (*ast.Block, error) {
	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for block, got %T", node)
	}

	if nonTerminal.Symbol != "Block" {
		return nil, fmt.Errorf("expected Block node, got %s", nonTerminal.Symbol)
	}

	// Block: LBRACE BlockStatements RBRACE
	if len(nonTerminal.Children) != 3 {
		return nil, fmt.Errorf("Block node expected 3 children, got %d", len(nonTerminal.Children))
	}

	// Extract LBRACE token (child 0)
	lbrace, ok := nonTerminal.Children[0].(*parsetree.TerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected terminal for {, got %T", nonTerminal.Children[0])
	}

	// Extract statements from BlockStatements (child 1)
	statements, err := extractBlockStatements(nonTerminal.Children[1])
	if err != nil {
		return nil, err
	}

	return &ast.Block{
		Token:      lbrace.Token.Value,
		Statements: statements,
	}, nil
}

// extractBlockStatements extracts statements from BlockStatements node.
// BlockStatements: NEWLINE BlockStatements | Statement BlockStmtRest | ε
func extractBlockStatements(node parsetree.ParseTree) ([]ast.Statement, error) {
	// Handle epsilon production
	if _, ok := node.(*parsetree.EmptyNode); ok {
		return []ast.Statement{}, nil
	}

	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for block statements, got %T", node)
	}

	if nonTerminal.Symbol != "BlockStatements" {
		return nil, fmt.Errorf("expected BlockStatements node, got %s", nonTerminal.Symbol)
	}

	// Check if empty (epsilon)
	if len(nonTerminal.Children) == 0 {
		return []ast.Statement{}, nil
	}

	// Could be: NEWLINE BlockStatements (1 or 2 children) or Statement BlockStmtRest (2 children)
	if len(nonTerminal.Children) == 1 {
		// Must be NEWLINE BlockStatements where BlockStatements is epsilon
		// Just skip the newline
		return []ast.Statement{}, nil
	}

	if len(nonTerminal.Children) != 2 {
		return nil, fmt.Errorf("BlockStatements node expected 0, 1, or 2 children, got %d", len(nonTerminal.Children))
	}

	// Check first child to determine which alternative
	if firstChild, ok := nonTerminal.Children[0].(*parsetree.TerminalNode); ok {
		// NEWLINE BlockStatements
		if firstChild.Token.Type == "NEWLINE" {
			// Skip newline and recursively extract from BlockStatements
			return extractBlockStatements(nonTerminal.Children[1])
		}
	}

	// Statement BlockStmtRest
	// Convert first statement
	stmt, err := convertToStatement(nonTerminal.Children[0])
	if err != nil {
		return nil, err
	}

	// Extract remaining statements from BlockStmtRest
	restStmts, err := extractBlockStmtRest(nonTerminal.Children[1])
	if err != nil {
		return nil, err
	}

	return append([]ast.Statement{stmt}, restStmts...), nil
}

// extractBlockStmtRest extracts remaining statements from BlockStmtRest node.
// BlockStmtRest: NEWLINE BlockStatements | ε
func extractBlockStmtRest(node parsetree.ParseTree) ([]ast.Statement, error) {
	// Handle epsilon production
	if _, ok := node.(*parsetree.EmptyNode); ok {
		return []ast.Statement{}, nil
	}

	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for block stmt rest, got %T", node)
	}

	if nonTerminal.Symbol != "BlockStmtRest" {
		return nil, fmt.Errorf("expected BlockStmtRest node, got %s", nonTerminal.Symbol)
	}

	// Check if empty (epsilon)
	if len(nonTerminal.Children) == 0 {
		return []ast.Statement{}, nil
	}

	// BlockStmtRest: NEWLINE BlockStatements
	if len(nonTerminal.Children) != 2 {
		return nil, fmt.Errorf("BlockStmtRest node expected 0 or 2 children, got %d", len(nonTerminal.Children))
	}

	// Recursively extract statements from BlockStatements
	return extractBlockStatements(nonTerminal.Children[1])
}

// convertFunctionLiteral converts a FunctionLiteral parse tree node to an AST FunctionLiteral.
// FunctionLiteral: FN LPAREN ParameterList RPAREN Block
func convertFunctionLiteral(node *parsetree.NonTerminalNode) (ast.Expression, error) {
	if node.Symbol != "FunctionLiteral" {
		return nil, fmt.Errorf("expected FunctionLiteral node, got %s", node.Symbol)
	}

	// FunctionLiteral: FN LPAREN ParameterList RPAREN Block
	if len(node.Children) != 5 {
		return nil, fmt.Errorf("FunctionLiteral node expected 5 children, got %d", len(node.Children))
	}

	// Extract fn token (child 0)
	fnNode, ok := node.Children[0].(*parsetree.TerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected terminal for fn keyword, got %T", node.Children[0])
	}

	// Extract parameter list (child 2)
	params, err := extractParameterList(node.Children[2])
	if err != nil {
		return nil, err
	}

	// Extract block (child 4)
	block, err := convertBlock(node.Children[4])
	if err != nil {
		return nil, err
	}

	return &ast.FunctionLiteral{
		Token:      fnNode.Token.Value,
		Parameters: params,
		Body:       block,
	}, nil
}

// processPrimaryRest processes a PrimaryRest node and builds index/member access expressions.
// This handles chaining like arr[0].len() or obj.field[0]
func processPrimaryRest(base ast.Expression, primaryRest parsetree.ParseTree) (ast.Expression, error) {
	switch rest := primaryRest.(type) {
	case *parsetree.EmptyNode:
		// No more chaining - return base expression
		return base, nil

	case *parsetree.NonTerminalNode:
		if rest.Symbol != "PrimaryRest" {
			return nil, fmt.Errorf("expected PrimaryRest, got %s", rest.Symbol)
		}

		if len(rest.Children) == 0 {
			// Epsilon - no more chaining
			return base, nil
		}

		// PrimaryRest can have 3 or 4 children depending on the variant
		if len(rest.Children) < 3 || len(rest.Children) > 4 {
			return nil, fmt.Errorf("PrimaryRest expected 0, 3, or 4 children, got %d", len(rest.Children))
		}

		// Determine which variant by checking first child
		firstToken, ok := rest.Children[0].(*parsetree.TerminalNode)
		if !ok {
			return nil, fmt.Errorf("expected terminal as first child of PrimaryRest, got %T", rest.Children[0])
		}

		var nextBase ast.Expression
		var recursiveRest parsetree.ParseTree

		switch firstToken.Token.Type {
		case "LPAREN":
			// Function call: LPAREN Arguments RPAREN (3 children, no recursion)
			arguments, err := extractArguments(rest.Children[1])
			if err != nil {
				return nil, err
			}
			// Extract the function name from the base expression
			var funcName string
			if memberAccess, ok := base.(*ast.MemberAccess); ok {
				funcName = memberAccess.Member
				// Pass the MemberAccess itself as the first argument
				// It will be evaluated to an ArrayMethod if it's an array method call
				nextBase = &ast.FunctionCall{
					Token:     funcName,
					Name:      funcName,
					Arguments: append([]ast.Expression{memberAccess}, arguments...),
				}
			} else {
				return nil, fmt.Errorf("function call syntax only supported after member access (e.g., arr.len())")
			}
			return nextBase, nil // No recursion for function calls

		case "LBRACKET":
			// Index access: LBRACKET Expression RBRACKET PrimaryRest (4 children)
			if len(rest.Children) != 4 {
				return nil, fmt.Errorf("index access expected 4 children, got %d", len(rest.Children))
			}
			indexExpr, err := convertToExpression(rest.Children[1])
			if err != nil {
				return nil, err
			}
			nextBase = &ast.IndexAccess{
				Token:  "[",
				Object: base,
				Index:  indexExpr,
			}
			recursiveRest = rest.Children[3] // Child 3 is the recursive PrimaryRest

		case "DOT":
			// Member access: DOT IDENTIFIER PrimaryRest (3 children)
			if len(rest.Children) != 3 {
				return nil, fmt.Errorf("member access expected 3 children, got %d", len(rest.Children))
			}
			memberToken, ok := rest.Children[1].(*parsetree.TerminalNode)
			if !ok {
				return nil, fmt.Errorf("expected IDENTIFIER after DOT, got %T", rest.Children[1])
			}
			nextBase = &ast.MemberAccess{
				Token:  ".",
				Object: base,
				Member: memberToken.Token.Value,
			}
			recursiveRest = rest.Children[2] // Child 2 is the recursive PrimaryRest

		default:
			return nil, fmt.Errorf("unexpected token type in PrimaryRest: %s", firstToken.Token.Type)
		}

		// Recursively process remaining PrimaryRest for chaining
		return processPrimaryRest(nextBase, recursiveRest)

	default:
		return nil, fmt.Errorf("unexpected node type for PrimaryRest: %T", primaryRest)
	}
}

// convertArrayLiteral converts an ArrayLiteral parse tree node to an AST ArrayLiteral.
// ArrayLiteral: LBRACKET ElementList RBRACKET | LBRACKET RBRACKET
func convertArrayLiteral(node *parsetree.NonTerminalNode) (ast.Expression, error) {
	if node.Symbol != "ArrayLiteral" {
		return nil, fmt.Errorf("expected ArrayLiteral node, got %s", node.Symbol)
	}

	// New grammar: ArrayLiteral: LBRACKET ArrayContent RBRACKET
	// Always has 3 children
	if len(node.Children) != 3 {
		return nil, fmt.Errorf("ArrayLiteral expected 3 children, got %d", len(node.Children))
	}

	// Check ArrayContent (Children[1])
	// ArrayContent: ElementList | ε
	arrayContent := node.Children[1]

	// Check if it's epsilon (empty)
	if _, isEmpty := arrayContent.(*parsetree.EmptyNode); isEmpty {
		// Empty array
		return &ast.ArrayLiteral{
			Token:    "[",
			Elements: []ast.Expression{},
		}, nil
	}

	// Non-empty array: extract elements from ArrayContent
	// ArrayContent should be a NonTerminalNode containing ElementList
	contentNode, ok := arrayContent.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected ArrayContent to be non-terminal or empty, got %T", arrayContent)
	}

	// ArrayContent has one child: ElementList
	if len(contentNode.Children) != 1 {
		return nil, fmt.Errorf("ArrayContent expected 1 child, got %d", len(contentNode.Children))
	}

	elements, err := extractElementList(contentNode.Children[0])
	if err != nil {
		return nil, err
	}

	return &ast.ArrayLiteral{
		Token:    "[",
		Elements: elements,
	}, nil
}

// extractElementList extracts elements from an ElementList parse tree node.
// ElementList: Expression ElementRest
func extractElementList(node parsetree.ParseTree) ([]ast.Expression, error) {
	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for element list, got %T", node)
	}

	if nonTerminal.Symbol != "ElementList" {
		return nil, fmt.Errorf("expected ElementList node, got %s", nonTerminal.Symbol)
	}

	if len(nonTerminal.Children) != 2 {
		return nil, fmt.Errorf("ElementList expected 2 children, got %d", len(nonTerminal.Children))
	}

	// Extract first element (child 0)
	firstElement, err := convertToExpression(nonTerminal.Children[0])
	if err != nil {
		return nil, err
	}

	// Extract remaining elements from ElementRest (child 1)
	restElements, err := extractElementRest(nonTerminal.Children[1])
	if err != nil {
		return nil, err
	}

	return append([]ast.Expression{firstElement}, restElements...), nil
}

// extractElementRest extracts remaining elements from an ElementRest parse tree node.
// ElementRest: COMMA Expression ElementRest | ε
func extractElementRest(node parsetree.ParseTree) ([]ast.Expression, error) {
	// Handle epsilon production
	if _, ok := node.(*parsetree.EmptyNode); ok {
		return []ast.Expression{}, nil
	}

	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for element rest, got %T", node)
	}

	if nonTerminal.Symbol != "ElementRest" {
		return nil, fmt.Errorf("expected ElementRest node, got %s", nonTerminal.Symbol)
	}

	// Check if empty (epsilon)
	if len(nonTerminal.Children) == 0 {
		return []ast.Expression{}, nil
	}

	// ElementRest: COMMA Expression ElementRest
	if len(nonTerminal.Children) != 3 {
		return nil, fmt.Errorf("ElementRest expected 0 or 3 children, got %d", len(nonTerminal.Children))
	}

	// Extract element (child 1)
	element, err := convertToExpression(nonTerminal.Children[1])
	if err != nil {
		return nil, err
	}

	// Recursively extract remaining elements
	restElements, err := extractElementRest(nonTerminal.Children[2])
	if err != nil {
		return nil, err
	}

	return append([]ast.Expression{element}, restElements...), nil
}

// extractIndexChain extracts index expressions from an IndexChain parse tree node.
// IndexChain: LBRACKET Expression RBRACKET IndexChain | LBRACKET Expression RBRACKET
func extractIndexChain(node parsetree.ParseTree) ([]ast.Expression, error) {
	nonTerminal, ok := node.(*parsetree.NonTerminalNode)
	if !ok {
		return nil, fmt.Errorf("expected non-terminal for index chain, got %T", node)
	}

	if nonTerminal.Symbol != "IndexChain" {
		return nil, fmt.Errorf("expected IndexChain node, got %s", nonTerminal.Symbol)
	}

	// IndexChain always has a LBRACKET Expression RBRACKET
	// Followed optionally by another IndexChain (4 children) or nothing (3 children)
	if len(nonTerminal.Children) == 3 {
		// Single index: LBRACKET Expression RBRACKET
		index, err := convertToExpression(nonTerminal.Children[1])
		if err != nil {
			return nil, err
		}
		return []ast.Expression{index}, nil
	}

	if len(nonTerminal.Children) != 4 {
		return nil, fmt.Errorf("IndexChain expected 3 or 4 children, got %d", len(nonTerminal.Children))
	}

	// Multiple indices: LBRACKET Expression RBRACKET IndexChain
	firstIndex, err := convertToExpression(nonTerminal.Children[1])
	if err != nil {
		return nil, err
	}

	// Recursively extract remaining indices
	restIndices, err := extractIndexChain(nonTerminal.Children[3])
	if err != nil {
		return nil, err
	}

	return append([]ast.Expression{firstIndex}, restIndices...), nil
}

// convertIndexAssignmentFromExpression converts an Assignment parse tree node to an IndexAssignment AST node.
// Assignment: LogicalOr AssignmentRest
// AssignmentRest: EQUALS Assignment
// The left side (LogicalOr) must evaluate to an IndexAccess expression.
func convertIndexAssignmentFromExpression(assignmentNode *parsetree.NonTerminalNode) (ast.Statement, error) {
	if len(assignmentNode.Children) != 2 {
		return nil, fmt.Errorf("Assignment expected 2 children, got %d", len(assignmentNode.Children))
	}

	// Parse left side to get the index access expression
	leftExpr, err := convertToExpression(assignmentNode.Children[0])
	if err != nil {
		return nil, fmt.Errorf("error converting left side of assignment: %v", err)
	}

	// Extract array name and indices from the left expression
	// It should be an IndexAccess or nested IndexAccess
	arrName, indices, err := extractIndexAssignmentParts(leftExpr)
	if err != nil {
		return nil, fmt.Errorf("left side of assignment must be an array index access: %v", err)
	}

	// Parse AssignmentRest: EQUALS Assignment
	assignmentRest := assignmentNode.Children[1]
	restNode, ok := assignmentRest.(*parsetree.NonTerminalNode)
	if !ok || len(restNode.Children) != 2 {
		return nil, fmt.Errorf("invalid assignment rest")
	}

	// Parse right side (the value to assign)
	valueExpr, err := convertToExpression(restNode.Children[1]) // Children[1] is the Assignment
	if err != nil {
		return nil, fmt.Errorf("error converting right side of assignment: %v", err)
	}

	return &ast.IndexAssignment{
		Token:   arrName,
		Name:    arrName,
		Indices: indices,
		Value:   valueExpr,
	}, nil
}

// extractIndexAssignmentParts extracts the array name and index expressions from an expression.
// For example, from arr[0] it returns ("arr", [0])
// From matrix[i][j] it returns ("matrix", [i, j])
func extractIndexAssignmentParts(expr ast.Expression) (string, []ast.Expression, error) {
	switch e := expr.(type) {
	case *ast.IndexAccess:
		// Base case or nested index access
		// Check if the object is an Identifier or another IndexAccess
		if ident, ok := e.Object.(*ast.Identifier); ok {
			// Base case: arr[i]
			return ident.Name, []ast.Expression{e.Index}, nil
		} else if _, ok := e.Object.(*ast.IndexAccess); ok {
			// Nested: arr[i][j]
			name, indices, err := extractIndexAssignmentParts(e.Object)
			if err != nil {
				return "", nil, err
			}
			return name, append(indices, e.Index), nil
		}
		return "", nil, fmt.Errorf("index access object must be identifier or index access, got %T", e.Object)

	default:
		return "", nil, fmt.Errorf("assignment target must be an array index access, got %T", expr)
	}
}
