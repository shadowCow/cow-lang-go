package grammar

import "testing"

func TestProductionRuleConstruction(t *testing.T) {
	t.Run("Terminal rule", func(t *testing.T) {
		rule := Terminal{TokenType: TOKEN_NUMBER}
		if rule.TokenType != TOKEN_NUMBER {
			t.Errorf("expected TOKEN_NUMBER, got %v", rule.TokenType)
		}
	})

	t.Run("NonTerminal rule", func(t *testing.T) {
		rule := NonTerminal{Symbol: SYM_EXPRESSION}
		if rule.Symbol != SYM_EXPRESSION {
			t.Errorf("expected SYM_EXPRESSION, got %v", rule.Symbol)
		}
	})

	t.Run("SynSequence rule", func(t *testing.T) {
		rule := SynSequence{
			Terminal{TokenType: TOKEN_IDENTIFIER},
			Terminal{TokenType: TOKEN_EQUALS},
			NonTerminal{Symbol: SYM_EXPRESSION},
		}
		if len(rule) != 3 {
			t.Errorf("expected 3 elements, got %d", len(rule))
		}
	})

	t.Run("SynAlternative rule", func(t *testing.T) {
		rule := SynAlternative{
			Terminal{TokenType: TOKEN_PLUS},
			Terminal{TokenType: TOKEN_MINUS},
		}
		if len(rule) != 2 {
			t.Errorf("expected 2 alternatives, got %d", len(rule))
		}
	})

	t.Run("SynOptional rule", func(t *testing.T) {
		rule := SynOptional{Inner: Terminal{TokenType: TOKEN_SEMICOLON}}
		if rule.Inner == nil {
			t.Error("expected inner rule, got nil")
		}
	})

	t.Run("SynZeroOrMore rule", func(t *testing.T) {
		rule := SynZeroOrMore{Inner: NonTerminal{Symbol: SYM_STATEMENT}}
		if rule.Inner == nil {
			t.Error("expected inner rule, got nil")
		}
	})

	t.Run("SynOneOrMore rule", func(t *testing.T) {
		rule := SynOneOrMore{Inner: NonTerminal{Symbol: SYM_STATEMENT}}
		if rule.Inner == nil {
			t.Error("expected inner rule, got nil")
		}
	})
}

func TestSyntacticGrammar(t *testing.T) {
	t.Run("Create syntactic grammar", func(t *testing.T) {
		grammar := SyntacticGrammar{
			StartSymbol: SYM_PROGRAM,
			Productions: map[Symbol]ProductionRule{
				SYM_PROGRAM: SynZeroOrMore{
					Inner: NonTerminal{Symbol: SYM_STATEMENT},
				},
				SYM_STATEMENT: SynSequence{
					Terminal{TokenType: TOKEN_IDENTIFIER},
					Terminal{TokenType: TOKEN_SEMICOLON},
				},
			},
		}

		if grammar.StartSymbol != SYM_PROGRAM {
			t.Errorf("expected SYM_PROGRAM start symbol, got %v", grammar.StartSymbol)
		}
		if len(grammar.Productions) != 2 {
			t.Errorf("expected 2 productions, got %d", len(grammar.Productions))
		}
	})

	t.Run("Productions are accessible by symbol", func(t *testing.T) {
		grammar := SyntacticGrammar{
			StartSymbol: SYM_EXPRESSION,
			Productions: map[Symbol]ProductionRule{
				SYM_EXPRESSION: SynAlternative{
					Terminal{TokenType: TOKEN_NUMBER},
					Terminal{TokenType: TOKEN_IDENTIFIER},
				},
			},
		}

		rule, exists := grammar.Productions[SYM_EXPRESSION]
		if !exists {
			t.Error("expected SYM_EXPRESSION production to exist")
		}
		if rule == nil {
			t.Error("expected non-nil production rule")
		}
	})
}

func TestExampleSyntacticGrammar(t *testing.T) {
	t.Run("Example grammar has start symbol", func(t *testing.T) {
		grammar := ExampleSyntacticGrammar()

		if grammar.StartSymbol == "" {
			t.Error("expected non-empty start symbol")
		}
	})

	t.Run("Example grammar has productions", func(t *testing.T) {
		grammar := ExampleSyntacticGrammar()

		if len(grammar.Productions) == 0 {
			t.Error("expected productions in example grammar, got none")
		}
	})

	t.Run("Start symbol exists in productions", func(t *testing.T) {
		grammar := ExampleSyntacticGrammar()

		_, exists := grammar.Productions[grammar.StartSymbol]
		if !exists {
			t.Errorf("start symbol %q not found in productions", grammar.StartSymbol)
		}
	})

	t.Run("Example grammar has expression production", func(t *testing.T) {
		grammar := ExampleSyntacticGrammar()

		_, exists := grammar.Productions[SYM_EXPRESSION]
		if !exists {
			t.Error("expected SYM_EXPRESSION production in example grammar")
		}
	})

	t.Run("Example grammar has factor production", func(t *testing.T) {
		grammar := ExampleSyntacticGrammar()

		_, exists := grammar.Productions[SYM_FACTOR]
		if !exists {
			t.Error("expected SYM_FACTOR production in example grammar")
		}
	})
}
