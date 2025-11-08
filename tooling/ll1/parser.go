package ll1

import (
	"fmt"

	"github.com/shadowCow/cow-lang-go/tooling/grammar"
	"github.com/shadowCow/cow-lang-go/tooling/lexer"
	"github.com/shadowCow/cow-lang-go/tooling/parsetree"
)

// Parser implements a table-driven LL(1) parser that returns generic parse trees.
type Parser struct {
	table         *ParseTable
	grammar       grammar.SyntacticGrammar
	tokens        []lexer.Token
	pos           int    // Current position in token stream
	trace         bool   // Optional: trace parsing steps for debugging
	filterToken   string // Token type to filter (e.g., "WHITESPACE"), empty for no filtering
}

// NewParser creates a new LL(1) parser.
// filterToken specifies a token type to filter out (e.g., "WHITESPACE"), or empty string for no filtering.
func NewParser(
	table *ParseTable,
	grammar grammar.SyntacticGrammar,
	tokens []lexer.Token,
	filterToken string,
) *Parser {
	// Filter tokens if specified
	filtered := tokens
	if filterToken != "" {
		filtered = make([]lexer.Token, 0, len(tokens))
		for _, tok := range tokens {
			if tok.Type != filterToken {
				filtered = append(filtered, tok)
			}
		}
	}

	return &Parser{
		table:       table,
		grammar:     grammar,
		tokens:      filtered,
		pos:         0,
		trace:       false,
		filterToken: filterToken,
	}
}

// SetTrace enables/disables parse tracing for debugging.
func (p *Parser) SetTrace(enabled bool) {
	p.trace = enabled
}

// Parse parses the token stream and returns a generic parse tree.
func (p *Parser) Parse() (*parsetree.ProgramNode, error) {
	// The stack holds symbols to be processed and their corresponding parse tree nodes
	stack := []stackItem{
		{symbol: symbolEOF, isTerminal: true},
		{symbol: string(p.grammar.StartSymbol), isTerminal: false},
	}

	// Stack for building parse tree nodes
	// As we reduce, we pop children and create parent nodes
	var nodeStack []parsetree.ParseTree

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
					// Success! Build final program node
					if len(nodeStack) == 0 {
						return nil, fmt.Errorf("parse completed but no parse tree was built")
					}
					if len(nodeStack) > 1 {
						return nil, fmt.Errorf("parse completed but multiple trees remain: %d", len(nodeStack))
					}
					return &parsetree.ProgramNode{Root: nodeStack[0]}, nil
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

			// Create terminal parse tree node
			terminalNode := &parsetree.TerminalNode{Token: currentToken}
			nodeStack = append(nodeStack, terminalNode)

			if p.trace {
				fmt.Printf("  Matched terminal: %s\n", terminalNode.String())
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

			// Count how many symbols this production will add
			symbols := p.extractSymbols(production)
			childCount := len(symbols)

			// Mark this position so we know how many children to collect
			// We'll use a marker item to track this
			stack = append(stack, stackItem{
				symbol:     top.symbol,
				isTerminal: false,
				isMarker:   true,
				childCount: childCount,
			})

			// Expand production by pushing its symbols onto stack (in reverse order)
			for i := len(symbols) - 1; i >= 0; i-- {
				stack = append(stack, symbols[i])
			}

			// Handle empty productions
			if childCount == 0 {
				// Pop the marker we just added
				stack = stack[:len(stack)-1]
				// Create an empty node
				emptyNode := &parsetree.EmptyNode{Symbol: nonTerminal}
				nodeStack = append(nodeStack, emptyNode)
				if p.trace {
					fmt.Printf("  Created empty node for %s\n", nonTerminal)
				}
			}
		}

		// Check if we need to reduce (found a marker)
		// Keep reducing while there are markers at the top of the stack
		for len(stack) > 0 && stack[len(stack)-1].isMarker {
			marker := stack[len(stack)-1]
			stack = stack[:len(stack)-1] // Pop marker

			// Collect children from node stack
			if len(nodeStack) < marker.childCount {
				return nil, fmt.Errorf("internal error: not enough nodes to reduce %s (need %d, have %d)",
					marker.symbol, marker.childCount, len(nodeStack))
			}

			// Pop children in reverse order (they were pushed in order)
			children := make([]parsetree.ParseTree, marker.childCount)
			for i := marker.childCount - 1; i >= 0; i-- {
				children[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}

			// Create non-terminal node
			nonTerminalNode := &parsetree.NonTerminalNode{
				Symbol:   grammar.Symbol(marker.symbol),
				Children: children,
			}
			nodeStack = append(nodeStack, nonTerminalNode)

			if p.trace {
				fmt.Printf("  Reduced to %s with %d children\n", marker.symbol, marker.childCount)
			}
		}
	}

	// Should not reach here if grammar is correct
	return nil, fmt.Errorf("parsing incomplete")
}

// stackItem represents an item on the parse stack.
type stackItem struct {
	symbol     string
	isTerminal bool
	isMarker   bool // True if this is a reduction marker
	childCount int  // Number of children to reduce (only for markers)
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
