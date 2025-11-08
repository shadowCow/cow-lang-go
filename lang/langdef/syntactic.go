package langdef

import "github.com/shadowCow/cow-lang-go/tooling/grammar"

// Non-terminal symbol constants for the Cow language syntactic grammar.
// These represent the structural elements of the language.

const (
	// Top-level program structure
	SYM_PROGRAM       grammar.Symbol = "Program"
	SYM_PROGRAM_REST  grammar.Symbol = "ProgramRest"
	SYM_PROGRAM_REST2 grammar.Symbol = "ProgramRest2"

	// Statements
	SYM_STATEMENT            grammar.Symbol = "Statement"
	SYM_LET_STATEMENT        grammar.Symbol = "LetStatement"
	SYM_EXPRESSION_STATEMENT grammar.Symbol = "ExpressionStatement"

	// Expressions (with operator precedence)
	SYM_EXPRESSION grammar.Symbol = "Expression"
	SYM_ADD_REST   grammar.Symbol = "AddRest"
	SYM_TERM       grammar.Symbol = "Term"
	SYM_MUL_REST   grammar.Symbol = "MulRest"
	SYM_FACTOR     grammar.Symbol = "Factor"
	SYM_FACTOR_REST grammar.Symbol = "FactorRest"

	// Operators
	SYM_ADD_OP grammar.Symbol = "AddOp"
	SYM_MUL_OP grammar.Symbol = "MulOp"

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
// Grammar (LL(1) - left-factored with operator precedence):
//   Program -> Statement ProgramRest
//   ProgramRest -> NEWLINE ProgramRest2 | ε
//   ProgramRest2 -> Statement ProgramRest | ε
//   Statement -> LetStatement | ExpressionStatement
//   LetStatement -> LET IDENTIFIER EQUALS Expression
//   ExpressionStatement -> Expression
//
//   Expression -> Term AddRest
//   AddRest -> AddOp Term AddRest | ε
//   AddOp -> PLUS | MINUS
//
//   Term -> Factor MulRest
//   MulRest -> MulOp Factor MulRest | ε
//   MulOp -> MULTIPLY | DIVIDE | MODULO
//
//   Factor -> IDENTIFIER FactorRest | Literal | LPAREN Expression RPAREN
//   FactorRest -> LPAREN Arguments RPAREN | ε
//   Arguments -> ε | ArgumentList
//   ArgumentList -> Expression ArgumentRest
//   ArgumentRest -> COMMA Expression ArgumentRest | ε
//   Literal -> INT_DECIMAL | INT_HEX | INT_BINARY | FLOAT
//
// Note: The grammar enforces operator precedence:
//   - Lowest:  + - (addition, subtraction)
//   - Higher:  * / % (multiplication, division, modulo)
//   - Highest: literals, identifiers, function calls, parentheses
func GetSyntacticGrammar() grammar.SyntacticGrammar {
	return grammar.SyntacticGrammar{
		StartSymbol: SYM_PROGRAM,
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			// Program is a sequence of statements
			SYM_PROGRAM: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_STATEMENT},
				grammar.NonTerminal{Symbol: SYM_PROGRAM_REST},
			},

			// ProgramRest: NEWLINE ProgramRest2 | ε
			// Handles newline-separated statements, allowing trailing newlines
			SYM_PROGRAM_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_NEWLINE},
					grammar.NonTerminal{Symbol: SYM_PROGRAM_REST2},
				},
				// Empty - no more statements (epsilon)
				grammar.SynSequence{}, // empty sequence = epsilon
			},

			// ProgramRest2: Statement ProgramRest | ε
			// Continues the statement sequence or allows trailing newline
			SYM_PROGRAM_REST2: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_STATEMENT},
					grammar.NonTerminal{Symbol: SYM_PROGRAM_REST},
				},
				// Empty - allows trailing newline (epsilon)
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

			// Expression: Term AddRest
			// Handles addition and subtraction (lowest precedence)
			SYM_EXPRESSION: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_TERM},
				grammar.NonTerminal{Symbol: SYM_ADD_REST},
			},

			// AddRest: AddOp Term AddRest | ε
			// Right-recursive to handle left-associativity during evaluation
			SYM_ADD_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_ADD_OP},
					grammar.NonTerminal{Symbol: SYM_TERM},
					grammar.NonTerminal{Symbol: SYM_ADD_REST},
				},
				grammar.SynSequence{}, // epsilon
			},

			// AddOp: PLUS | MINUS
			SYM_ADD_OP: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_PLUS},
				grammar.Terminal{TokenType: TOKEN_MINUS},
			},

			// Term: Factor MulRest
			// Handles multiplication, division, and modulo (higher precedence)
			SYM_TERM: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_FACTOR},
				grammar.NonTerminal{Symbol: SYM_MUL_REST},
			},

			// MulRest: MulOp Factor MulRest | ε
			SYM_MUL_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_MUL_OP},
					grammar.NonTerminal{Symbol: SYM_FACTOR},
					grammar.NonTerminal{Symbol: SYM_MUL_REST},
				},
				grammar.SynSequence{}, // epsilon
			},

			// MulOp: MULTIPLY | DIVIDE | MODULO
			SYM_MUL_OP: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_MULTIPLY},
				grammar.Terminal{TokenType: TOKEN_DIVIDE},
				grammar.Terminal{TokenType: TOKEN_MODULO},
			},

			// Factor: IDENTIFIER FactorRest | Literal | LPAREN Expression RPAREN
			// Handles highest precedence items: atoms and parenthesized expressions
			SYM_FACTOR: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_IDENTIFIER},
					grammar.NonTerminal{Symbol: SYM_FACTOR_REST},
				},
				grammar.NonTerminal{Symbol: SYM_LITERAL},
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_LPAREN},
					grammar.NonTerminal{Symbol: SYM_EXPRESSION},
					grammar.Terminal{TokenType: TOKEN_RPAREN},
				},
			},

			// FactorRest: LPAREN Arguments RPAREN | ε
			// Distinguishes between function calls and identifiers
			SYM_FACTOR_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_LPAREN},
					grammar.NonTerminal{Symbol: SYM_ARGUMENTS},
					grammar.Terminal{TokenType: TOKEN_RPAREN},
				},
				grammar.SynSequence{}, // epsilon
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
