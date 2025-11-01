package langdef

import "github.com/shadowCow/cow-lang-go/lang/grammar"

// Non-terminal symbol constants for the Cow language syntactic grammar.
// These represent the structural elements of the language.

const (
	// Top-level program structure
	SYM_PROGRAM grammar.Symbol = "Program"

	// Literal values
	SYM_LITERAL grammar.Symbol = "Literal"
)

// GetSyntacticGrammar returns the syntactic grammar for the Cow language.
// This defines how tokens are organized into language constructs.
//
// Production rules should be organized hierarchically
func GetSyntacticGrammar() grammar.SyntacticGrammar {
	return grammar.SyntacticGrammar{
		StartSymbol: SYM_PROGRAM,
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			// Minimal grammar: a program is a literal (for now)
			// Later this will expand to include functions, types, etc.
			SYM_PROGRAM: grammar.NonTerminal{
				Symbol: SYM_LITERAL,
			},

			// A literal can be any number token
			// Later this will expand to include strings, booleans, etc.
			SYM_LITERAL: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_INT_DECIMAL},
				grammar.Terminal{TokenType: TOKEN_INT_HEX},
				grammar.Terminal{TokenType: TOKEN_INT_BINARY},
				grammar.Terminal{TokenType: TOKEN_FLOAT},
			},
		},
	}
}
