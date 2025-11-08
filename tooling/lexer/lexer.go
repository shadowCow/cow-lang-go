package lexer

import (
	"fmt"
	"unicode/utf8"

	"github.com/shadowCow/cow-lang-go/tooling/automata"
)

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type   string // Token type (e.g., "INT_DECIMAL", "FLOAT")
	Value  string // Actual text matched
	Line   int    // Line number (1-indexed)
	Column int    // Column number (1-indexed)
	Offset int    // Byte offset in source (0-indexed)
}

// Lexer tokenizes source code using a compiled DFA.
type Lexer struct {
	dfa    automata.DfaWithTokens
	source string
	offset int    // Current position in source
	line   int    // Current line (1-indexed)
	column int    // Current column (1-indexed)
}

// NewLexer creates a new lexer with a compiled DFA.
func NewLexer(dfa automata.DfaWithTokens, source string) *Lexer {
	return &Lexer{
		dfa:    dfa,
		source: source,
		offset: 0,
		line:   1,
		column: 1,
	}
}

// Tokenize returns all tokens from the source.
func (l *Lexer) Tokenize() ([]Token, error) {
	tokens := make([]Token, 0)

	for l.offset < len(l.source) {
		token, err := l.nextToken()
		if err != nil {
			return tokens, err
		}
		if token != nil {
			tokens = append(tokens, *token)
		}
	}

	return tokens, nil
}

// nextToken returns the next token using longest-match tokenization.
func (l *Lexer) nextToken() (*Token, error) {
	if l.offset >= len(l.source) {
		return nil, nil
	}

	startOffset := l.offset
	startLine := l.line
	startColumn := l.column

	state := l.dfa.InitialState
	lastAcceptState := ""
	lastAcceptOffset := -1
	lastAcceptLine := l.line
	lastAcceptColumn := l.column

	// Try to match as long as possible (longest match)
	for l.offset < len(l.source) {
		// Decode the next UTF-8 rune
		r, size := utf8.DecodeRuneInString(l.source[l.offset:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 sequence
			break
		}

		nextState := l.dfa.NextState(state, r)

		if nextState == "" {
			// No valid transition
			break
		}

		// Move to next state
		state = nextState
		l.offset += size // Advance by the actual number of bytes for this rune

		// Track position
		if r == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}

		// Check if this is an accepting state
		if l.dfa.IsAccepting(state) {
			lastAcceptState = state
			lastAcceptOffset = l.offset
			lastAcceptLine = l.line
			lastAcceptColumn = l.column
		}
	}

	// If we found an accepting state, create token
	if lastAcceptOffset > startOffset {
		tokenType := l.dfa.GetTokenType(lastAcceptState)
		value := l.source[startOffset:lastAcceptOffset]

		// Reset to end of last accept
		l.offset = lastAcceptOffset
		l.line = lastAcceptLine
		l.column = lastAcceptColumn

		return &Token{
			Type:   string(tokenType), // Convert grammar.TokenType to string
			Value:  value,
			Line:   startLine,
			Column: startColumn,
			Offset: startOffset,
		}, nil
	}

	// No token matched - error
	// Decode the rune at the error position for a better error message
	r, _ := utf8.DecodeRuneInString(l.source[startOffset:])
	if r == utf8.RuneError {
		return nil, fmt.Errorf("invalid UTF-8 sequence at line %d, column %d",
			startLine, startColumn)
	}
	return nil, fmt.Errorf("unexpected character at line %d, column %d: %q",
		startLine, startColumn, r)
}