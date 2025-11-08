package grammar

import "testing"

func TestLexicalPatternConstruction(t *testing.T) {
	t.Run("Literal pattern", func(t *testing.T) {
		pattern := Literal("if")
		if pattern != "if" {
			t.Errorf("expected 'if', got %v", pattern)
		}
	})

	t.Run("CharRange pattern", func(t *testing.T) {
		pattern := CharRange{From: 'a', To: 'z'}
		if pattern.From != 'a' || pattern.To != 'z' {
			t.Errorf("expected range a-z, got %c-%c", pattern.From, pattern.To)
		}
	})

	t.Run("CharSet pattern", func(t *testing.T) {
		pattern := CharSet{'a', 'b', 'c'}
		if len(pattern) != 3 {
			t.Errorf("expected 3 characters, got %d", len(pattern))
		}
	})

	t.Run("LexSequence pattern", func(t *testing.T) {
		pattern := LexSequence{
			Literal("h"),
			Literal("i"),
		}
		if len(pattern) != 2 {
			t.Errorf("expected 2 elements, got %d", len(pattern))
		}
	})

	t.Run("LexAlternative pattern", func(t *testing.T) {
		pattern := LexAlternative{
			Literal("if"),
			Literal("else"),
		}
		if len(pattern) != 2 {
			t.Errorf("expected 2 alternatives, got %d", len(pattern))
		}
	})

	t.Run("LexOptional pattern", func(t *testing.T) {
		pattern := LexOptional{Inner: Literal("?")}
		if pattern.Inner == nil {
			t.Error("expected inner pattern, got nil")
		}
	})

	t.Run("LexZeroOrMore pattern", func(t *testing.T) {
		pattern := LexZeroOrMore{Inner: CharRange{From: '0', To: '9'}}
		if pattern.Inner == nil {
			t.Error("expected inner pattern, got nil")
		}
	})

	t.Run("LexOneOrMore pattern", func(t *testing.T) {
		pattern := LexOneOrMore{Inner: CharRange{From: '0', To: '9'}}
		if pattern.Inner == nil {
			t.Error("expected inner pattern, got nil")
		}
	})
}

func TestTokenDefinition(t *testing.T) {
	t.Run("Create token definition", func(t *testing.T) {
		token := TokenDefinition{
			Name: TOKEN_NUMBER,
			Priority: 1,
		}

		if token.Name != TOKEN_NUMBER {
			t.Errorf("expected TOKEN_NUMBER, got %v", token.Name)
		}
		if token.Priority != 1 {
			t.Errorf("expected priority 1, got %d", token.Priority)
		}
	})
}

func TestLexicalGrammar(t *testing.T) {
	t.Run("Create lexical grammar", func(t *testing.T) {
		grammar := LexicalGrammar{
			Tokens: []TokenDefinition{
				{
					Name: TOKEN_NUMBER,
					Pattern: LexOneOrMore{
						Inner: CharRange{From: '0', To: '9'},
					},
					Priority: 1,
				},
				{
					Name:     TOKEN_PLUS,
					Pattern:  Literal("+"),
					Priority: 2,
				},
			},
		}

		if len(grammar.Tokens) != 2 {
			t.Errorf("expected 2 tokens, got %d", len(grammar.Tokens))
		}
	})
}

func TestExampleLexicalGrammar(t *testing.T) {
	t.Run("Example grammar has tokens", func(t *testing.T) {
		grammar := ExampleLexicalGrammar()

		if len(grammar.Tokens) == 0 {
			t.Error("expected tokens in example grammar, got none")
		}
	})

	t.Run("Example grammar has number token", func(t *testing.T) {
		grammar := ExampleLexicalGrammar()

		foundNumber := false
		for _, token := range grammar.Tokens {
			if token.Name == TOKEN_NUMBER {
				foundNumber = true
				break
			}
		}

		if !foundNumber {
			t.Error("expected NUMBER token in example grammar")
		}
	})

	t.Run("Example grammar has identifier token", func(t *testing.T) {
		grammar := ExampleLexicalGrammar()

		foundIdentifier := false
		for _, token := range grammar.Tokens {
			if token.Name == TOKEN_IDENTIFIER {
				foundIdentifier = true
				break
			}
		}

		if !foundIdentifier {
			t.Error("expected IDENTIFIER token in example grammar")
		}
	})
}
