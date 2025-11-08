package grammar

// Example token types for a simple arithmetic expression language
const (
	TOKEN_NUMBER     TokenType = "NUMBER"
	TOKEN_IDENTIFIER TokenType = "IDENTIFIER"
	TOKEN_PLUS       TokenType = "PLUS"
	TOKEN_MINUS      TokenType = "MINUS"
	TOKEN_STAR       TokenType = "STAR"
	TOKEN_SLASH      TokenType = "SLASH"
	TOKEN_LPAREN     TokenType = "LPAREN"
	TOKEN_RPAREN     TokenType = "RPAREN"
	TOKEN_EQUALS     TokenType = "EQUALS"
	TOKEN_SEMICOLON  TokenType = "SEMICOLON"
	TOKEN_WHITESPACE TokenType = "WHITESPACE"
)

// Example symbol constants for syntactic grammar
const (
	SYM_PROGRAM    Symbol = "Program"
	SYM_STATEMENT  Symbol = "Statement"
	SYM_ASSIGNMENT Symbol = "Assignment"
	SYM_EXPRESSION Symbol = "Expression"
	SYM_TERM       Symbol = "Term"
	SYM_FACTOR     Symbol = "Factor"
)

// ExampleLexicalGrammar returns a sample lexical grammar for arithmetic expressions.
func ExampleLexicalGrammar() LexicalGrammar {
	return LexicalGrammar{
		Tokens: []TokenDefinition{
			// Whitespace (spaces, tabs, newlines)
			{
				Name: TOKEN_WHITESPACE,
				Pattern: LexOneOrMore{
					Inner: LexAlternative{
						Literal(" "),
						Literal("\t"),
						Literal("\n"),
						Literal("\r"),
					},
				},
				Priority: 0,
			},
			// Numbers: one or more digits
			{
				Name: TOKEN_NUMBER,
				Pattern: LexOneOrMore{
					Inner: CharRange{From: '0', To: '9'},
				},
				Priority: 1,
			},
			// Identifiers: letter or underscore, followed by letters, digits, or underscores
			{
				Name: TOKEN_IDENTIFIER,
				Pattern: LexSequence{
					LexAlternative{
						CharRange{From: 'a', To: 'z'},
						CharRange{From: 'A', To: 'Z'},
						Literal("_"),
					},
					LexZeroOrMore{
						Inner: LexAlternative{
							CharRange{From: 'a', To: 'z'},
							CharRange{From: 'A', To: 'Z'},
							CharRange{From: '0', To: '9'},
							Literal("_"),
						},
					},
				},
				Priority: 1,
			},
			// Operators and punctuation
			{
				Name:     TOKEN_PLUS,
				Pattern:  Literal("+"),
				Priority: 2,
			},
			{
				Name:     TOKEN_MINUS,
				Pattern:  Literal("-"),
				Priority: 2,
			},
			{
				Name:     TOKEN_STAR,
				Pattern:  Literal("*"),
				Priority: 2,
			},
			{
				Name:     TOKEN_SLASH,
				Pattern:  Literal("/"),
				Priority: 2,
			},
			{
				Name:     TOKEN_LPAREN,
				Pattern:  Literal("("),
				Priority: 2,
			},
			{
				Name:     TOKEN_RPAREN,
				Pattern:  Literal(")"),
				Priority: 2,
			},
			{
				Name:     TOKEN_EQUALS,
				Pattern:  Literal("="),
				Priority: 2,
			},
			{
				Name:     TOKEN_SEMICOLON,
				Pattern:  Literal(";"),
				Priority: 2,
			},
		},
	}
}

// ExampleSyntacticGrammar returns a sample syntactic grammar for arithmetic expressions.
// Grammar:
//   Program    -> Statement*
//   Statement  -> Assignment ;
//   Assignment -> IDENTIFIER = Expression
//   Expression -> Term ((+ | -) Term)*
//   Term       -> Factor ((* | /) Factor)*
//   Factor     -> NUMBER | IDENTIFIER | ( Expression )
func ExampleSyntacticGrammar() SyntacticGrammar {
	return SyntacticGrammar{
		StartSymbol: SYM_PROGRAM,
		Productions: map[Symbol]ProductionRule{
			SYM_PROGRAM: SynZeroOrMore{
				Inner: NonTerminal{Symbol: SYM_STATEMENT},
			},
			SYM_STATEMENT: SynSequence{
				NonTerminal{Symbol: SYM_ASSIGNMENT},
				Terminal{TokenType: TOKEN_SEMICOLON},
			},
			SYM_ASSIGNMENT: SynSequence{
				Terminal{TokenType: TOKEN_IDENTIFIER},
				Terminal{TokenType: TOKEN_EQUALS},
				NonTerminal{Symbol: SYM_EXPRESSION},
			},
			SYM_EXPRESSION: SynSequence{
				NonTerminal{Symbol: SYM_TERM},
				SynZeroOrMore{
					Inner: SynSequence{
						SynAlternative{
							Terminal{TokenType: TOKEN_PLUS},
							Terminal{TokenType: TOKEN_MINUS},
						},
						NonTerminal{Symbol: SYM_TERM},
					},
				},
			},
			SYM_TERM: SynSequence{
				NonTerminal{Symbol: SYM_FACTOR},
				SynZeroOrMore{
					Inner: SynSequence{
						SynAlternative{
							Terminal{TokenType: TOKEN_STAR},
							Terminal{TokenType: TOKEN_SLASH},
						},
						NonTerminal{Symbol: SYM_FACTOR},
					},
				},
			},
			SYM_FACTOR: SynAlternative{
				Terminal{TokenType: TOKEN_NUMBER},
				Terminal{TokenType: TOKEN_IDENTIFIER},
				SynSequence{
					Terminal{TokenType: TOKEN_LPAREN},
					NonTerminal{Symbol: SYM_EXPRESSION},
					Terminal{TokenType: TOKEN_RPAREN},
				},
			},
		},
	}
}
