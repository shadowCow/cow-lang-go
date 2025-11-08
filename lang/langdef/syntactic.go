package langdef

import "github.com/shadowCow/cow-lang-go/tooling/grammar"

// Non-terminal symbol constants for the Cow language syntactic grammar.
// These represent the structural elements of the language.

const (
	// Top-level program structure
	SYM_PROGRAM      grammar.Symbol = "Program"
	SYM_PROGRAM_REST grammar.Symbol = "ProgramRest"

	// Statements
	SYM_STATEMENT            grammar.Symbol = "Statement"
	SYM_LET_STATEMENT        grammar.Symbol = "LetStatement"
	SYM_EXPRESSION_STATEMENT grammar.Symbol = "ExpressionStatement"

	// Expressions
	SYM_EXPRESSION      grammar.Symbol = "Expression"
	SYM_EXPRESSION_REST grammar.Symbol = "ExpressionRest"

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
//   Program -> Statement ProgramRest
//   ProgramRest -> Statement ProgramRest | ε
//   Statement -> LetStatement | ExpressionStatement
//   LetStatement -> LET IDENTIFIER EQUALS Expression
//   ExpressionStatement -> Expression
//   Expression -> IDENTIFIER ExpressionRest | Literal
//   ExpressionRest -> LPAREN Arguments RPAREN | ε
//   Arguments -> ε | ArgumentList
//   ArgumentList -> Expression ArgumentRest
//   ArgumentRest -> COMMA Expression ArgumentRest | ε
//   Literal -> INT_DECIMAL | INT_HEX | INT_BINARY | FLOAT
//
// Note: IDENTIFIER ExpressionRest is left-factored to handle both:
//   - FunctionCall: IDENTIFIER LPAREN Arguments RPAREN (when ExpressionRest has LPAREN)
//   - Identifier: IDENTIFIER (when ExpressionRest is ε)
func GetSyntacticGrammar() grammar.SyntacticGrammar {
	return grammar.SyntacticGrammar{
		StartSymbol: SYM_PROGRAM,
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			// Program is a sequence of statements
			SYM_PROGRAM: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_STATEMENT},
				grammar.NonTerminal{Symbol: SYM_PROGRAM_REST},
			},

			// ProgramRest: Statement ProgramRest | ε
			// Handles multiple statements in a program
			SYM_PROGRAM_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_STATEMENT},
					grammar.NonTerminal{Symbol: SYM_PROGRAM_REST},
				},
				// Empty - no more statements (epsilon)
				grammar.SynSequence{}, // empty sequence = epsilon
			},

			// Statement can be a let statement or an expression statement
			SYM_STATEMENT: grammar.SynAlternative{
				grammar.NonTerminal{Symbol: SYM_LET_STATEMENT},
				grammar.NonTerminal{Symbol: SYM_EXPRESSION_STATEMENT},
			},

			// LetStatement: LET IDENTIFIER EQUALS Expression
			SYM_LET_STATEMENT: grammar.SynSequence{
				grammar.Terminal{TokenType: TOKEN_LET},
				grammar.Terminal{TokenType: TOKEN_IDENTIFIER},
				grammar.Terminal{TokenType: TOKEN_EQUALS},
				grammar.NonTerminal{Symbol: SYM_EXPRESSION},
			},

			// ExpressionStatement: Expression
			SYM_EXPRESSION_STATEMENT: grammar.NonTerminal{
				Symbol: SYM_EXPRESSION,
			},

			// Expression: IDENTIFIER ExpressionRest | Literal
			// This is left-factored to distinguish between function calls and identifiers
			SYM_EXPRESSION: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_IDENTIFIER},
					grammar.NonTerminal{Symbol: SYM_EXPRESSION_REST},
				},
				grammar.NonTerminal{Symbol: SYM_LITERAL},
			},

			// ExpressionRest: LPAREN Arguments RPAREN | ε
			// If LPAREN, it's a function call; if ε, it's just an identifier
			SYM_EXPRESSION_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_LPAREN},
					grammar.NonTerminal{Symbol: SYM_ARGUMENTS},
					grammar.Terminal{TokenType: TOKEN_RPAREN},
				},
				// Empty - just an identifier (epsilon)
				grammar.SynSequence{}, // empty sequence = epsilon
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
