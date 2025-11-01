package lexer

import (
	"testing"

	"github.com/shadowCow/cow-lang-go/lang/automata"
	"github.com/shadowCow/cow-lang-go/lang/grammar"
)

// TestLexerBasic tests the basic tokenization functionality.
func TestLexerBasic(t *testing.T) {
	// Create a simple grammar
	lexGrammar := grammar.LexicalGrammar{
		Tokens: []grammar.TokenDefinition{
			{
				Name:     "DIGIT",
				Pattern:  grammar.CharRange{From: '0', To: '9'},
				Priority: 1,
			},
			{
				Name:     "LETTER",
				Pattern:  grammar.CharRange{From: 'a', To: 'z'},
				Priority: 1,
			},
		},
	}

	dfa := automata.CompileLexicalGrammar(lexGrammar)

	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single digit",
			input: "5",
			expected: []Token{
				{Type: "DIGIT", Value: "5", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "single letter",
			input: "a",
			expected: []Token{
				{Type: "LETTER", Value: "a", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "digit and letter",
			input: "5a",
			expected: []Token{
				{Type: "DIGIT", Value: "5", Line: 1, Column: 1, Offset: 0},
				{Type: "LETTER", Value: "a", Line: 1, Column: 2, Offset: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(dfa, tt.input)
			tokens, err := lex.Tokenize()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, expected := range tt.expected {
				actual := tokens[i]
				if actual.Type != expected.Type {
					t.Errorf("Token %d: expected type %q, got %q", i, expected.Type, actual.Type)
				}
				if actual.Value != expected.Value {
					t.Errorf("Token %d: expected value %q, got %q", i, expected.Value, actual.Value)
				}
			}
		})
	}
}

// TestLexerError tests error handling for unexpected characters.
func TestLexerError(t *testing.T) {
	// Create grammar that only accepts digits
	lexGrammar := grammar.LexicalGrammar{
		Tokens: []grammar.TokenDefinition{
			{
				Name:     "DIGIT",
				Pattern:  grammar.CharRange{From: '0', To: '9'},
				Priority: 1,
			},
		},
	}

	dfa := automata.CompileLexicalGrammar(lexGrammar)
	lex := NewLexer(dfa, "5x")

	tokens, err := lex.Tokenize()

	if err == nil {
		t.Fatal("Expected error for unexpected character, got nil")
	}

	// Should have lexed the first token before hitting error
	if len(tokens) != 1 {
		t.Errorf("Expected 1 token before error, got %d", len(tokens))
	}
}