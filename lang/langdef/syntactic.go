package langdef

import "github.com/shadowCow/cow-lang-go/tooling/grammar"

// Non-terminal symbol constants for the Cow language syntactic grammar.
// These represent the structural elements of the language.

const (
	// Top-level program structure
	SYM_PROGRAM grammar.Symbol = "Program"

	// Expressions
	SYM_EXPRESSION grammar.Symbol = "Expression"

	// Function calls
	SYM_FUNCTION_CALL grammar.Symbol = "FunctionCall"
	SYM_ARGUMENTS     grammar.Symbol = "Arguments"
	SYM_ARG_LIST      grammar.Symbol = "ArgumentList"
	SYM_ARG_REST      grammar.Symbol = "ArgumentRest"

	// Literal values
	SYM_LITERAL grammar.Symbol = "Literal"
)

// GetSyntacticGrammar returns the syntactic grammar for the Cow language.
// This defines how tokens are organized into language constructs.
//
// Grammar (LL(1) - left-factored):
//   Program -> Expression
//   Expression -> FunctionCall | Literal
//   FunctionCall -> IDENTIFIER LPAREN Arguments RPAREN
//   Arguments -> ε | ArgumentList
//   ArgumentList -> Expression ArgumentRest
//   ArgumentRest -> COMMA Expression ArgumentRest | ε
//   Literal -> INT_DECIMAL | INT_HEX | INT_BINARY | FLOAT
func GetSyntacticGrammar() grammar.SyntacticGrammar {
	return grammar.SyntacticGrammar{
		StartSymbol: SYM_PROGRAM,
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			// Program is a single expression
			SYM_PROGRAM: grammar.NonTerminal{
				Symbol: SYM_EXPRESSION,
			},

			// Expression can be a function call or a literal
			SYM_EXPRESSION: grammar.SynAlternative{
				grammar.NonTerminal{Symbol: SYM_FUNCTION_CALL},
				grammar.NonTerminal{Symbol: SYM_LITERAL},
			},

			// FunctionCall: IDENTIFIER LPAREN Arguments RPAREN
			SYM_FUNCTION_CALL: grammar.SynSequence{
				grammar.Terminal{TokenType: TOKEN_IDENTIFIER},
				grammar.Terminal{TokenType: TOKEN_LPAREN},
				grammar.NonTerminal{Symbol: SYM_ARGUMENTS},
				grammar.Terminal{TokenType: TOKEN_RPAREN},
			},

			// Arguments: optional argument list
			SYM_ARGUMENTS: grammar.SynAlternative{
				grammar.NonTerminal{Symbol: SYM_ARG_LIST},
				// Empty - represents no arguments (epsilon)
				grammar.SynSequence{}, // empty sequence = epsilon
			},

			// ArgumentList: Expression ArgumentRest
			// This is left-factored to be LL(1)
			SYM_ARG_LIST: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_EXPRESSION},
				grammar.NonTerminal{Symbol: SYM_ARG_REST},
			},

			// ArgumentRest: COMMA Expression ArgumentRest | ε
			// Handles the tail of the argument list
			SYM_ARG_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_COMMA},
					grammar.NonTerminal{Symbol: SYM_EXPRESSION},
					grammar.NonTerminal{Symbol: SYM_ARG_REST},
				},
				// Empty - no more arguments (epsilon)
				grammar.SynSequence{}, // empty sequence = epsilon
			},

			// A literal can be any number token
			SYM_LITERAL: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_INT_DECIMAL},
				grammar.Terminal{TokenType: TOKEN_INT_HEX},
				grammar.Terminal{TokenType: TOKEN_INT_BINARY},
				grammar.Terminal{TokenType: TOKEN_FLOAT},
			},
		},
	}
}
