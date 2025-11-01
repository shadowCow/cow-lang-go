package langdef

import (
	"testing"

	"github.com/shadowCow/cow-lang-go/lang/automata"
	"github.com/shadowCow/cow-lang-go/lang/lexer"
)

// TestLexNumberLiterals tests lexing number literals from the Cow language grammar.
func TestLexNumberLiterals(t *testing.T) {
	// Compile the lexical grammar
	lexGrammar := GetLexical()
	dfa := automata.CompileLexicalGrammar(lexGrammar)

	tests := []struct {
		name     string
		input    string
		expected []lexer.Token
	}{
		{
			name:  "decimal integer",
			input: "42",
			expected: []lexer.Token{
				{Type: "INT_DECIMAL", Value: "42", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "decimal with underscores",
			input: "1_000_000",
			expected: []lexer.Token{
				{Type: "INT_DECIMAL", Value: "1_000_000", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "hexadecimal",
			input: "0xFF",
			expected: []lexer.Token{
				{Type: "INT_HEX", Value: "0xFF", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "binary",
			input: "0b1010",
			expected: []lexer.Token{
				{Type: "INT_BINARY", Value: "0b1010", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "float",
			input: "3.14",
			expected: []lexer.Token{
				{Type: "FLOAT", Value: "3.14", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "float with exponent",
			input: "1.5e10",
			expected: []lexer.Token{
				{Type: "FLOAT", Value: "1.5e10", Line: 1, Column: 1, Offset: 0},
			},
		},
		{
			name:  "scientific notation",
			input: "2e-5",
			expected: []lexer.Token{
				{Type: "FLOAT", Value: "2e-5", Line: 1, Column: 1, Offset: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := lexer.NewLexer(dfa, tt.input)
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
				if actual.Line != expected.Line {
					t.Errorf("Token %d: expected line %d, got %d", i, expected.Line, actual.Line)
				}
				if actual.Column != expected.Column {
					t.Errorf("Token %d: expected column %d, got %d", i, expected.Column, actual.Column)
				}
			}
		})
	}
}

// TestPriorityHandling tests that hex and binary take priority over decimal for inputs starting with 0.
func TestPriorityHandling(t *testing.T) {
	lexGrammar := GetLexical()
	dfa := automata.CompileLexicalGrammar(lexGrammar)

	tests := []struct {
		name         string
		input        string
		expectedType string
	}{
		{
			name:         "0x prefix should be hex not decimal",
			input:        "0x10",
			expectedType: "INT_HEX",
		},
		{
			name:         "0b prefix should be binary not decimal",
			input:        "0b10",
			expectedType: "INT_BINARY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := lexer.NewLexer(dfa, tt.input)
			tokens, err := lex.Tokenize()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != 1 {
				t.Fatalf("Expected 1 token, got %d", len(tokens))
			}

			if tokens[0].Type != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, tokens[0].Type)
			}
		})
	}
}
