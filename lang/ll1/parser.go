package ll1

import (
	"fmt"

	"github.com/shadowCow/cow-lang-go/lang/ast"
	"github.com/shadowCow/cow-lang-go/lang/grammar"
	"github.com/shadowCow/cow-lang-go/lang/lexer"
)

// Parser implements a table-driven LL(1) parser.
type Parser struct {
	table   *ParseTable
	grammar grammar.SyntacticGrammar
	tokens  []lexer.Token
	pos     int // Current position in token stream
	// Optional: trace parsing steps for debugging
	trace bool
}

// NewParser creates a new LL(1) parser.
func NewParser(
	table *ParseTable,
	grammar grammar.SyntacticGrammar,
	tokens []lexer.Token,
) *Parser {
	// Filter out whitespace tokens
	filtered := make([]lexer.Token, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Type != "WHITESPACE" {
			filtered = append(filtered, tok)
		}
	}

	return &Parser{
		table:   table,
		grammar: grammar,
		tokens:  filtered,
		pos:     0,
		trace:   false,
	}
}

// SetTrace enables/disables parse tracing for debugging.
func (p *Parser) SetTrace(enabled bool) {
	p.trace = enabled
}

// Parse parses the token stream and returns an AST.
func (p *Parser) Parse() (*ast.Program, error) {
	// The stack holds symbols to be processed
	// We push the start symbol initially
	stack := []stackItem{
		{symbol: symbolEOF, isTerminal: true},
		{symbol: string(p.grammar.StartSymbol), isTerminal: false},
	}

	// Track AST nodes as we parse
	// For the simple grammar, we'll build a Program with one statement
	var programStatements []ast.Statement

	for len(stack) > 0 {
		// Pop top of stack
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Get current lookahead
		lookahead := p.currentToken()

		if p.trace {
			fmt.Printf("Stack top: %s, Lookahead: %s\n", top.symbol, lookahead)
		}

		if top.isTerminal {
			// Top is a terminal - match it with input
			if top.symbol == symbolEOF {
				// Expect end of input
				if p.pos >= len(p.tokens) {
					// Success!
					break
				}
				return nil, fmt.Errorf("unexpected token %q at line %d, column %d (expected end of input)",
					p.tokens[p.pos].Value, p.tokens[p.pos].Line, p.tokens[p.pos].Column)
			}

			// Match terminal with current token
			if p.pos >= len(p.tokens) {
				return nil, fmt.Errorf("unexpected end of input (expected %s)", top.symbol)
			}

			currentToken := p.tokens[p.pos]
			if currentToken.Type != top.symbol {
				return nil, fmt.Errorf("unexpected token %q (type %s) at line %d, column %d (expected %s)",
					currentToken.Value, currentToken.Type, currentToken.Line, currentToken.Column, top.symbol)
			}

			// Build AST node for this terminal
			astNode, err := p.buildASTForTerminal(currentToken)
			if err != nil {
				return nil, err
			}
			if astNode != nil {
				// Wrap in expression statement and add to program
				programStatements = append(programStatements, &ast.ExpressionStatement{
					Token:      currentToken.Value,
					Expression: astNode,
				})
			}

			// Match successful, advance input
			p.pos++

		} else {
			// Top is a non-terminal - look up production in table
			nonTerminal := grammar.Symbol(top.symbol)
			production := p.table.Get(nonTerminal, lookahead)

			if production == nil {
				// No production found - syntax error
				if p.pos >= len(p.tokens) {
					return nil, fmt.Errorf("unexpected end of input while parsing %s", nonTerminal)
				}
				token := p.tokens[p.pos]
				return nil, fmt.Errorf("unexpected token %q (type %s) at line %d, column %d while parsing %s",
					token.Value, token.Type, token.Line, token.Column, nonTerminal)
			}

			if p.trace {
				fmt.Printf("  Expanding %s -> %s\n", nonTerminal, formatProduction(production))
			}

			// Expand production by pushing its symbols onto stack (in reverse order)
			symbols := p.extractSymbols(production)
			for i := len(symbols) - 1; i >= 0; i-- {
				stack = append(stack, symbols[i])
			}
		}
	}

	// Return the constructed program
	return &ast.Program{
		Statements: programStatements,
	}, nil
}

// stackItem represents an item on the parse stack.
type stackItem struct {
	symbol     string
	isTerminal bool
}

const symbolEOF = "$"

// currentToken returns the lookahead token type.
func (p *Parser) currentToken() string {
	if p.pos >= len(p.tokens) {
		return symbolEOF
	}
	return p.tokens[p.pos].Type
}

// extractSymbols extracts the symbols from a production to push onto the stack.
func (p *Parser) extractSymbols(prod grammar.ProductionRule) []stackItem {
	switch production := prod.(type) {
	case grammar.Terminal:
		return []stackItem{{symbol: string(production.TokenType), isTerminal: true}}

	case grammar.NonTerminal:
		return []stackItem{{symbol: string(production.Symbol), isTerminal: false}}

	case grammar.SynSequence:
		var symbols []stackItem
		for _, elem := range production {
			symbols = append(symbols, p.extractSymbols(elem)...)
		}
		return symbols

	case grammar.SynAlternative:
		// For alternatives, the table should have selected which one to use
		// We shouldn't encounter a raw SynAlternative here during parsing
		// This is a logic error in table construction
		panic("encountered SynAlternative during parsing - table construction bug")

	case grammar.SynOptional:
		// Optional productions are already handled by the table
		// If we're expanding an optional, we chose to expand it
		return p.extractSymbols(production.Inner)

	case grammar.SynZeroOrMore:
		// Zero-or-more productions are already handled by the table
		// This is complex - we may need to expand multiple times
		// For now, handle as single expansion
		return p.extractSymbols(production.Inner)

	case grammar.SynOneOrMore:
		// One-or-more productions are already handled by the table
		return p.extractSymbols(production.Inner)

	default:
		panic(fmt.Sprintf("unknown production type: %T", prod))
	}
}

// buildASTForTerminal creates an AST node for a matched terminal.
func (p *Parser) buildASTForTerminal(token lexer.Token) (ast.Expression, error) {
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
		value, err := parseFloatLiteral(token)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float at line %d, column %d: %v",
				token.Line, token.Column, err)
		}
		return &ast.FloatLiteral{
			Token: token.Value,
			Value: value,
		}, nil

	default:
		// For other terminals (keywords, operators, etc.), we may not build AST nodes
		return nil, nil
	}
}

// parseIntLiteral parses an integer literal token value.
// Handles decimal, hexadecimal, and binary formats.
func parseIntLiteral(token lexer.Token) (int64, error) {
	// Remove underscores (used for readability in literals)
	value := removeUnderscores(token.Value)

	switch token.Type {
	case "INT_DECIMAL":
		var result int64
		_, err := fmt.Sscanf(value, "%d", &result)
		return result, err
	case "INT_HEX":
		// Remove "0x" prefix
		if len(value) < 3 {
			return 0, fmt.Errorf("invalid hex literal: %s", token.Value)
		}
		var result int64
		_, err := fmt.Sscanf(value[2:], "%x", &result)
		return result, err
	case "INT_BINARY":
		// Remove "0b" prefix and parse binary
		if len(value) < 3 {
			return 0, fmt.Errorf("invalid binary literal: %s", token.Value)
		}
		var result int64
		for _, ch := range value[2:] {
			result = result * 2
			if ch == '1' {
				result++
			} else if ch != '0' {
				return 0, fmt.Errorf("invalid binary digit: %c", ch)
			}
		}
		return result, nil
	default:
		return 0, fmt.Errorf("unknown integer token type: %s", token.Type)
	}
}

// parseFloatLiteral parses a float literal token value.
func parseFloatLiteral(token lexer.Token) (float64, error) {
	value := removeUnderscores(token.Value)
	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	return result, err
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
