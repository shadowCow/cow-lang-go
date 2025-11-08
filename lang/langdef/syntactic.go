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

	// Logical OR (lowest precedence)
	SYM_LOGICAL_OR      grammar.Symbol = "LogicalOr"
	SYM_LOGICAL_OR_REST grammar.Symbol = "LogicalOrRest"

	// Logical AND
	SYM_LOGICAL_AND      grammar.Symbol = "LogicalAnd"
	SYM_LOGICAL_AND_REST grammar.Symbol = "LogicalAndRest"

	// Equality
	SYM_EQUALITY      grammar.Symbol = "Equality"
	SYM_EQUALITY_REST grammar.Symbol = "EqualityRest"
	SYM_EQUALITY_OP   grammar.Symbol = "EqualityOp"

	// Comparison
	SYM_COMPARISON      grammar.Symbol = "Comparison"
	SYM_COMPARISON_REST grammar.Symbol = "ComparisonRest"
	SYM_COMPARISON_OP   grammar.Symbol = "ComparisonOp"

	// Arithmetic (addition/subtraction)
	SYM_ARITHMETIC grammar.Symbol = "Arithmetic"
	SYM_ADD_REST   grammar.Symbol = "AddRest"
	SYM_ADD_OP     grammar.Symbol = "AddOp"

	// Term (multiplication/division/modulo)
	SYM_TERM     grammar.Symbol = "Term"
	SYM_MUL_REST grammar.Symbol = "MulRest"
	SYM_MUL_OP   grammar.Symbol = "MulOp"

	// Unary
	SYM_UNARY    grammar.Symbol = "Unary"
	SYM_UNARY_OP grammar.Symbol = "UnaryOp"

	// Primary (highest precedence)
	SYM_PRIMARY      grammar.Symbol = "Primary"
	SYM_PRIMARY_REST grammar.Symbol = "PrimaryRest"

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
//   Expression -> LogicalOr
//   LogicalOr -> LogicalAnd LogicalOrRest
//   LogicalOrRest -> OR LogicalAnd LogicalOrRest | ε
//   LogicalAnd -> Equality LogicalAndRest
//   LogicalAndRest -> AND Equality LogicalAndRest | ε
//   Equality -> Comparison EqualityRest
//   EqualityRest -> EqualityOp Comparison EqualityRest | ε
//   EqualityOp -> EQUAL_EQUAL | NOT_EQUAL
//   Comparison -> Arithmetic ComparisonRest
//   ComparisonRest -> ComparisonOp Arithmetic ComparisonRest | ε
//   ComparisonOp -> LESS_THAN | LESS_EQUAL | GREATER_THAN | GREATER_EQUAL
//   Arithmetic -> Term AddRest
//   AddRest -> AddOp Term AddRest | ε
//   AddOp -> PLUS | MINUS
//   Term -> Unary MulRest
//   MulRest -> MulOp Unary MulRest | ε
//   MulOp -> MULTIPLY | DIVIDE | MODULO
//   Unary -> UnaryOp Unary | Primary
//   UnaryOp -> NOT | MINUS
//   Primary -> IDENTIFIER PrimaryRest | Literal | LPAREN Expression RPAREN
//   PrimaryRest -> LPAREN Arguments RPAREN | ε
//   Arguments -> ε | ArgumentList
//   ArgumentList -> Expression ArgumentRest
//   ArgumentRest -> COMMA Expression ArgumentRest | ε
//   Literal -> INT_DECIMAL | INT_HEX | INT_BINARY | FLOAT | TRUE | FALSE
//
// Note: The grammar enforces operator precedence (lowest to highest):
//   1. || (logical or)
//   2. && (logical and)
//   3. == != (equality)
//   4. < <= > >= (comparison)
//   5. + - (addition, subtraction)
//   6. * / % (multiplication, division, modulo)
//   7. ! - (unary not, unary minus)
//   8. literals, identifiers, function calls, parentheses
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

			// Expression: LogicalOr
			SYM_EXPRESSION: grammar.NonTerminal{Symbol: SYM_LOGICAL_OR},

			// LogicalOr: LogicalAnd LogicalOrRest
			SYM_LOGICAL_OR: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_LOGICAL_AND},
				grammar.NonTerminal{Symbol: SYM_LOGICAL_OR_REST},
			},

			// LogicalOrRest: OR LogicalAnd LogicalOrRest | ε
			SYM_LOGICAL_OR_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_OR},
					grammar.NonTerminal{Symbol: SYM_LOGICAL_AND},
					grammar.NonTerminal{Symbol: SYM_LOGICAL_OR_REST},
				},
				grammar.SynSequence{}, // epsilon
			},

			// LogicalAnd: Equality LogicalAndRest
			SYM_LOGICAL_AND: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_EQUALITY},
				grammar.NonTerminal{Symbol: SYM_LOGICAL_AND_REST},
			},

			// LogicalAndRest: AND Equality LogicalAndRest | ε
			SYM_LOGICAL_AND_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_AND},
					grammar.NonTerminal{Symbol: SYM_EQUALITY},
					grammar.NonTerminal{Symbol: SYM_LOGICAL_AND_REST},
				},
				grammar.SynSequence{}, // epsilon
			},

			// Equality: Comparison EqualityRest
			SYM_EQUALITY: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_COMPARISON},
				grammar.NonTerminal{Symbol: SYM_EQUALITY_REST},
			},

			// EqualityRest: EqualityOp Comparison EqualityRest | ε
			SYM_EQUALITY_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_EQUALITY_OP},
					grammar.NonTerminal{Symbol: SYM_COMPARISON},
					grammar.NonTerminal{Symbol: SYM_EQUALITY_REST},
				},
				grammar.SynSequence{}, // epsilon
			},

			// EqualityOp: EQUAL_EQUAL | NOT_EQUAL
			SYM_EQUALITY_OP: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_EQUAL_EQUAL},
				grammar.Terminal{TokenType: TOKEN_NOT_EQUAL},
			},

			// Comparison: Arithmetic ComparisonRest
			SYM_COMPARISON: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_ARITHMETIC},
				grammar.NonTerminal{Symbol: SYM_COMPARISON_REST},
			},

			// ComparisonRest: ComparisonOp Arithmetic ComparisonRest | ε
			SYM_COMPARISON_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_COMPARISON_OP},
					grammar.NonTerminal{Symbol: SYM_ARITHMETIC},
					grammar.NonTerminal{Symbol: SYM_COMPARISON_REST},
				},
				grammar.SynSequence{}, // epsilon
			},

			// ComparisonOp: LESS_THAN | LESS_EQUAL | GREATER_THAN | GREATER_EQUAL
			SYM_COMPARISON_OP: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_LESS_THAN},
				grammar.Terminal{TokenType: TOKEN_LESS_EQUAL},
				grammar.Terminal{TokenType: TOKEN_GREATER_THAN},
				grammar.Terminal{TokenType: TOKEN_GREATER_EQUAL},
			},

			// Arithmetic: Term AddRest
			SYM_ARITHMETIC: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_TERM},
				grammar.NonTerminal{Symbol: SYM_ADD_REST},
			},

			// AddRest: AddOp Term AddRest | ε
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

			// Term: Unary MulRest
			SYM_TERM: grammar.SynSequence{
				grammar.NonTerminal{Symbol: SYM_UNARY},
				grammar.NonTerminal{Symbol: SYM_MUL_REST},
			},

			// MulRest: MulOp Unary MulRest | ε
			SYM_MUL_REST: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_MUL_OP},
					grammar.NonTerminal{Symbol: SYM_UNARY},
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

			// Unary: UnaryOp Unary | Primary
			SYM_UNARY: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.NonTerminal{Symbol: SYM_UNARY_OP},
					grammar.NonTerminal{Symbol: SYM_UNARY},
				},
				grammar.NonTerminal{Symbol: SYM_PRIMARY},
			},

			// UnaryOp: NOT | MINUS
			SYM_UNARY_OP: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_NOT},
				grammar.Terminal{TokenType: TOKEN_MINUS},
			},

			// Primary: IDENTIFIER PrimaryRest | Literal | LPAREN Expression RPAREN
			SYM_PRIMARY: grammar.SynAlternative{
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_IDENTIFIER},
					grammar.NonTerminal{Symbol: SYM_PRIMARY_REST},
				},
				grammar.NonTerminal{Symbol: SYM_LITERAL},
				grammar.SynSequence{
					grammar.Terminal{TokenType: TOKEN_LPAREN},
					grammar.NonTerminal{Symbol: SYM_EXPRESSION},
					grammar.Terminal{TokenType: TOKEN_RPAREN},
				},
			},

			// PrimaryRest: LPAREN Arguments RPAREN | ε
			SYM_PRIMARY_REST: grammar.SynAlternative{
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

			// A literal can be a number or boolean
			SYM_LITERAL: grammar.SynAlternative{
				grammar.Terminal{TokenType: TOKEN_INT_DECIMAL},
				grammar.Terminal{TokenType: TOKEN_INT_HEX},
				grammar.Terminal{TokenType: TOKEN_INT_BINARY},
				grammar.Terminal{TokenType: TOKEN_FLOAT},
				grammar.Terminal{TokenType: TOKEN_TRUE},
				grammar.Terminal{TokenType: TOKEN_FALSE},
			},
		},
	}
}
