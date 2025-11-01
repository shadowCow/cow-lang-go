package langdef

import (
	"testing"
)

// TestGetGrammar verifies that the grammar can be retrieved without panicking.
// Once the grammar is fully defined, this test should be expanded to validate
// the grammar structure.
func TestGetGrammar(t *testing.T) {
	grammar := GetGrammar()

	// Basic sanity checks
	if grammar.Lexical.Tokens == nil {
		t.Error("Lexical tokens should not be nil")
	}

	if grammar.Syntactic.Productions == nil {
		t.Error("Syntactic productions should not be nil")
	}

	// TODO: Add more comprehensive grammar validation tests once grammar is defined
	// - Verify all expected tokens are present
	// - Verify all expected symbols are present
	// - Verify production rules are complete
	// - Test that grammar can successfully parse example programs
}

// TestGetLexical verifies that the lexical grammar can be retrieved.
func TestGetLexical(t *testing.T) {
	lexical := GetLexical()

	if lexical.Tokens == nil {
		t.Error("Tokens should not be nil")
	}

	if len(lexical.Tokens) == 0 {
		t.Log("Warning: No tokens defined yet (expected during scaffolding phase)")
	}

	// TODO: Add tests for specific token patterns once defined
}

// TestGetSyntactic verifies that the syntactic grammar can be retrieved.
func TestGetSyntactic(t *testing.T) {
	syntactic := GetSyntactic()

	if syntactic.Productions == nil {
		t.Error("Productions should not be nil")
	}

	if len(syntactic.Productions) == 0 {
		t.Log("Warning: No productions defined yet (expected during scaffolding phase)")
	}

	// TODO: Add tests for specific production rules once defined
}
