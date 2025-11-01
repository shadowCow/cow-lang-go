// Package parser implements a parser for the Cow language.
// It converts a stream of tokens into an Abstract Syntax Tree (AST).
package parser

import (
	"fmt"
	"strconv"

	"github.com/shadowCow/cow-lang-go/lang/ast"
	"github.com/shadowCow/cow-lang-go/lang/lexer"
)

// Parser holds the state during parsing.
type Parser struct {
	tokens   []lexer.Token
	position int // Current position in tokens
}

// NewParser creates a new parser from a token stream.
// Whitespace tokens are filtered out before parsing.
func NewParser(tokens []lexer.Token) *Parser {
	// Filter out whitespace tokens
	filtered := make([]lexer.Token, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Type != "WHITESPACE" {
			filtered = append(filtered, tok)
		}
	}

	return &Parser{
		tokens:   filtered,
		position: 0,
	}
}

// Parse parses the token stream and returns an AST.
func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		program.Statements = append(program.Statements, stmt)
	}

	return program, nil
}

// parseStatement parses a single statement.
// For now, we only support expression statements.
func (p *Parser) parseStatement() (ast.Statement, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionStatement{
		Token:      p.previous().Value,
		Expression: expr,
	}, nil
}

// parseExpression parses an expression.
// For now, expressions are either:
// - Function calls (e.g., println(42))
// - Primary expressions (number literals)
func (p *Parser) parseExpression() (ast.Expression, error) {
	return p.parseCallExpression()
}

// parseCallExpression parses function call expressions.
// If it's not a function call, it falls back to parsePrimary.
func (p *Parser) parseCallExpression() (ast.Expression, error) {
	// Check if current token looks like an identifier (function name)
	// For now, we only have built-in functions, so we'll check for specific names
	if !p.isAtEnd() && p.peek().Type == "IDENTIFIER" {
		nameToken := p.advance()
		name := nameToken.Value

		// Expect opening parenthesis
		if p.isAtEnd() || p.peek().Value != "(" {
			return nil, fmt.Errorf("expected '(' after function name at line %d, column %d",
				nameToken.Line, nameToken.Column)
		}
		p.advance() // consume '('

		// Parse arguments
		var arguments []ast.Expression

		// If not immediately closing paren, parse arguments
		if !p.isAtEnd() && p.peek().Value != ")" {
			for {
				arg, err := p.parsePrimary()
				if err != nil {
					return nil, err
				}
				arguments = append(arguments, arg)

				// Check for comma (more arguments) or closing paren
				if p.isAtEnd() {
					return nil, fmt.Errorf("expected ')' or ',' after argument")
				}

				if p.peek().Value == ")" {
					break
				}

				if p.peek().Value != "," {
					return nil, fmt.Errorf("expected ')' or ',' after argument, got %q", p.peek().Value)
				}
				p.advance() // consume ','
			}
		}

		// Expect closing parenthesis
		if p.isAtEnd() || p.peek().Value != ")" {
			return nil, fmt.Errorf("expected ')' after function arguments")
		}
		p.advance() // consume ')'

		return &ast.FunctionCall{
			Token:     nameToken.Value,
			Name:      name,
			Arguments: arguments,
		}, nil
	}

	// Not a function call, parse as primary expression
	return p.parsePrimary()
}

// parsePrimary parses primary expressions (number literals).
func (p *Parser) parsePrimary() (ast.Expression, error) {
	if p.isAtEnd() {
		return nil, fmt.Errorf("unexpected end of input")
	}

	token := p.advance()

	switch token.Type {
	case "INT_DECIMAL", "INT_HEX", "INT_BINARY":
		value, err := parseIntLiteral(token)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer at line %d, column %d: %v",
				token.Line, token.Column, err)
		}
		return &ast.IntLiteral{
			Token: token.Value,
			Value: value,
		}, nil

	case "FLOAT":
		value, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float at line %d, column %d: %v",
				token.Line, token.Column, err)
		}
		return &ast.FloatLiteral{
			Token: token.Value,
			Value: value,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected token %q (type %s) at line %d, column %d",
			token.Value, token.Type, token.Line, token.Column)
	}
}

// parseIntLiteral parses an integer literal token value.
// Handles decimal, hexadecimal, and binary formats.
func parseIntLiteral(token lexer.Token) (int64, error) {
	// Remove underscores (used for readability in literals)
	value := removeUnderscores(token.Value)

	switch token.Type {
	case "INT_DECIMAL":
		return strconv.ParseInt(value, 10, 64)
	case "INT_HEX":
		// Remove "0x" prefix
		if len(value) < 3 {
			return 0, fmt.Errorf("invalid hex literal: %s", token.Value)
		}
		return strconv.ParseInt(value[2:], 16, 64)
	case "INT_BINARY":
		// Remove "0b" prefix
		if len(value) < 3 {
			return 0, fmt.Errorf("invalid binary literal: %s", token.Value)
		}
		return strconv.ParseInt(value[2:], 2, 64)
	default:
		return 0, fmt.Errorf("unknown integer token type: %s", token.Type)
	}
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

// Helper methods for token navigation

func (p *Parser) peek() lexer.Token {
	return p.tokens[p.position]
}

func (p *Parser) previous() lexer.Token {
	return p.tokens[p.position-1]
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.position++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.position >= len(p.tokens)
}
