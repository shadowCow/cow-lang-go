package lexer

import (
	"testing"

	"github.com/shadowCow/cow-lang-go/lang/automata"
	"github.com/shadowCow/cow-lang-go/lang/grammar"
)

// TestUnicodeSupport tests that the lexer correctly handles Unicode characters.
func TestUnicodeSupport(t *testing.T) {
	// Create a grammar that accepts any character
	lexGrammar := grammar.LexicalGrammar{
		Tokens: []grammar.TokenDefinition{
			{
				Name: "CHAR",
				// Accept a wide range of Unicode characters
				Pattern: grammar.LexAlternative{
					grammar.CharRange{From: 'a', To: 'z'},
					grammar.CharRange{From: 'A', To: 'Z'},
					grammar.CharRange{From: '0', To: '9'},
					grammar.CharRange{From: 0x4E00, To: 0x9FFF}, // CJK Unified Ideographs
					grammar.CharRange{From: 0x0400, To: 0x04FF}, // Cyrillic
					grammar.CharRange{From: 0x1F600, To: 0x1F64F}, // Emoticons
				},
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
			name:  "Chinese characters",
			input: "中文",
			expected: []Token{
				{Type: "CHAR", Value: "中", Line: 1, Column: 1, Offset: 0},
				{Type: "CHAR", Value: "文", Line: 1, Column: 2, Offset: 3}, // UTF-8: 3 bytes per char
			},
		},
		{
			name:  "Japanese characters",
			input: "日本",
			expected: []Token{
				{Type: "CHAR", Value: "日", Line: 1, Column: 1, Offset: 0},
				{Type: "CHAR", Value: "本", Line: 1, Column: 2, Offset: 3},
			},
		},
		{
			name:  "Cyrillic characters",
			input: "Привет",
			expected: []Token{
				{Type: "CHAR", Value: "П", Line: 1, Column: 1, Offset: 0},
				{Type: "CHAR", Value: "р", Line: 1, Column: 2, Offset: 2}, // UTF-8: 2 bytes per char
				{Type: "CHAR", Value: "и", Line: 1, Column: 3, Offset: 4},
				{Type: "CHAR", Value: "в", Line: 1, Column: 4, Offset: 6},
				{Type: "CHAR", Value: "е", Line: 1, Column: 5, Offset: 8},
				{Type: "CHAR", Value: "т", Line: 1, Column: 6, Offset: 10},
			},
		},
		{
			name:  "Emoji",
			input: "😀😁",
			expected: []Token{
				{Type: "CHAR", Value: "😀", Line: 1, Column: 1, Offset: 0},
				{Type: "CHAR", Value: "😁", Line: 1, Column: 2, Offset: 4}, // UTF-8: 4 bytes per emoji
			},
		},
		{
			name:  "Mixed ASCII and Unicode",
			input: "Hello世界",
			expected: []Token{
				{Type: "CHAR", Value: "H", Line: 1, Column: 1, Offset: 0},
				{Type: "CHAR", Value: "e", Line: 1, Column: 2, Offset: 1},
				{Type: "CHAR", Value: "l", Line: 1, Column: 3, Offset: 2},
				{Type: "CHAR", Value: "l", Line: 1, Column: 4, Offset: 3},
				{Type: "CHAR", Value: "o", Line: 1, Column: 5, Offset: 4},
				{Type: "CHAR", Value: "世", Line: 1, Column: 6, Offset: 5},
				{Type: "CHAR", Value: "界", Line: 1, Column: 7, Offset: 8},
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
				if actual.Line != expected.Line {
					t.Errorf("Token %d: expected line %d, got %d", i, expected.Line, actual.Line)
				}
				if actual.Column != expected.Column {
					t.Errorf("Token %d: expected column %d, got %d", i, expected.Column, actual.Column)
				}
				if actual.Offset != expected.Offset {
					t.Errorf("Token %d: expected offset %d, got %d (for value %q)",
						i, expected.Offset, actual.Offset, actual.Value)
				}
			}
		})
	}
}

// TestUnicodeColumnTracking verifies that columns are counted correctly for Unicode.
func TestUnicodeColumnTracking(t *testing.T) {
	// Create grammar that accepts digits
	lexGrammar := grammar.LexicalGrammar{
		Tokens: []grammar.TokenDefinition{
			{
				Name:     "DIGIT",
				Pattern:  grammar.CharRange{From: '0', To: '9'},
				Priority: 1,
			},
			{
				Name:     "CJK",
				Pattern:  grammar.CharRange{From: 0x4E00, To: 0x9FFF},
				Priority: 1,
			},
		},
	}

	dfa := automata.CompileLexicalGrammar(lexGrammar)

	// Mixed width characters: "1" (1 byte), "中" (3 bytes), "2" (1 byte)
	// Columns should be 1, 2, 3 (not affected by byte width)
	lex := NewLexer(dfa, "1中2")
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []struct {
		value  string
		column int
		offset int
	}{
		{"1", 1, 0},  // 1 byte at offset 0
		{"中", 2, 1}, // 3 bytes at offset 1
		{"2", 3, 4},  // 1 byte at offset 4
	}

	for i, exp := range expected {
		if tokens[i].Value != exp.value {
			t.Errorf("Token %d: expected value %q, got %q", i, exp.value, tokens[i].Value)
		}
		if tokens[i].Column != exp.column {
			t.Errorf("Token %d: expected column %d, got %d", i, exp.column, tokens[i].Column)
		}
		if tokens[i].Offset != exp.offset {
			t.Errorf("Token %d: expected offset %d, got %d", i, exp.offset, tokens[i].Offset)
		}
	}
}
