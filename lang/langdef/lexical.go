package langdef

import "github.com/shadowCow/cow-lang-go/lang/grammar"

// Token type constants for the Cow language.
// These will be expanded as the language grammar is defined.

const (
	// Number literals
	TOKEN_INT_DECIMAL grammar.TokenType = "INT_DECIMAL" // 42, 1_000_000
	TOKEN_INT_HEX     grammar.TokenType = "INT_HEX"     // 0xFF, 0x1A_3B
	TOKEN_INT_BINARY  grammar.TokenType = "INT_BINARY"  // 0b1010, 0b1111_0000
	TOKEN_FLOAT       grammar.TokenType = "FLOAT"       // 3.14, 1.5e10, 2e-5

	// Identifiers
	TOKEN_IDENTIFIER grammar.TokenType = "IDENTIFIER" // function names, variable names

	// Punctuation
	TOKEN_LPAREN grammar.TokenType = "LPAREN" // (
	TOKEN_RPAREN grammar.TokenType = "RPAREN" // )
	TOKEN_COMMA  grammar.TokenType = "COMMA"  // ,

	// Whitespace (to be skipped)
	TOKEN_WHITESPACE grammar.TokenType = "WHITESPACE" // spaces, tabs, newlines

	// TODO: Add remaining tokens for Phase 1
	// - Keywords: TOKEN_KEYWORD_FN, TOKEN_KEYWORD_LET, TOKEN_KEYWORD_MATCH, etc.
	// - Operators: TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, etc.
	// - More punctuation: TOKEN_LBRACE, TOKEN_RBRACE, TOKEN_SEMICOLON, etc.
	// - String literals: TOKEN_STRING
	// - Boolean literals: TOKEN_TRUE, TOKEN_FALSE
)

// GetLexicalGrammar returns the lexical grammar for the Cow language.
// This defines how the source text is tokenized.
func GetLexicalGrammar() grammar.LexicalGrammar {
	// Helper patterns for number literals
	digit := grammar.CharRange{From: '0', To: '9'}
	digitOrUnderscore := grammar.LexAlternative{
		digit,
		grammar.Literal("_"),
	}

	hexDigit := grammar.LexAlternative{
		digit,
		grammar.CharRange{From: 'a', To: 'f'},
		grammar.CharRange{From: 'A', To: 'F'},
	}
	hexDigitOrUnderscore := grammar.LexAlternative{
		hexDigit,
		grammar.Literal("_"),
	}

	binaryDigit := grammar.LexAlternative{
		grammar.Literal("0"),
		grammar.Literal("1"),
	}
	binaryDigitOrUnderscore := grammar.LexAlternative{
		binaryDigit,
		grammar.Literal("_"),
	}

	// Integer part: one or more digits (with optional underscores)
	integerPart := grammar.LexSequence{
		digit,
		grammar.LexZeroOrMore{Inner: digitOrUnderscore},
	}

	// Exponent: [eE] [+-]? digits
	exponent := grammar.LexSequence{
		grammar.LexAlternative{
			grammar.Literal("e"),
			grammar.Literal("E"),
		},
		grammar.LexOptional{
			Inner: grammar.LexAlternative{
				grammar.Literal("+"),
				grammar.Literal("-"),
			},
		},
		grammar.LexSequence{
			digit,
			grammar.LexZeroOrMore{Inner: digitOrUnderscore},
		},
	}

	// Helper patterns for identifiers
	letter := grammar.LexAlternative{
		grammar.CharRange{From: 'a', To: 'z'},
		grammar.CharRange{From: 'A', To: 'Z'},
	}
	letterOrDigit := grammar.LexAlternative{
		letter,
		digit,
		grammar.Literal("_"),
	}

	// Whitespace characters
	whitespaceChar := grammar.LexAlternative{
		grammar.Literal(" "),
		grammar.Literal("\t"),
		grammar.Literal("\n"),
		grammar.Literal("\r"),
	}

	return grammar.LexicalGrammar{
		Tokens: []grammar.TokenDefinition{
			// Identifiers: must start with letter or underscore, followed by letters/digits/underscores
			// Higher priority to match before being confused with number literals
			{
				Name: TOKEN_IDENTIFIER,
				Pattern: grammar.LexSequence{
					grammar.LexAlternative{
						letter,
						grammar.Literal("_"),
					},
					grammar.LexZeroOrMore{Inner: letterOrDigit},
				},
				Priority: 4,
			},

			// Number literals
			// Note: Hex and binary have higher priority than decimal since they start with '0'
			// Float has higher priority than decimal int to match decimal points

			// Hexadecimal integers: 0xFF, 0x1A_3B
			{
				Name: TOKEN_INT_HEX,
				Pattern: grammar.LexSequence{
					grammar.Literal("0x"),
					grammar.LexOneOrMore{Inner: hexDigitOrUnderscore},
				},
				Priority: 3,
			},

			// Binary integers: 0b1010, 0b1111_0000
			{
				Name: TOKEN_INT_BINARY,
				Pattern: grammar.LexSequence{
					grammar.Literal("0b"),
					grammar.LexOneOrMore{Inner: binaryDigitOrUnderscore},
				},
				Priority: 3,
			},

			// Float literals: 3.14, 1.5e10, 2e-5, 3.14e-8
			{
				Name: TOKEN_FLOAT,
				Pattern: grammar.LexAlternative{
					// Form 1: integer '.' integer [exponent]
					grammar.LexSequence{
						integerPart,
						grammar.Literal("."),
						integerPart,
						grammar.LexOptional{Inner: exponent},
					},
					// Form 2: integer exponent (no decimal point)
					grammar.LexSequence{
						integerPart,
						exponent,
					},
				},
				Priority: 2,
			},

			// Decimal integers: 42, 1_000_000
			{
				Name:     TOKEN_INT_DECIMAL,
				Pattern:  integerPart,
				Priority: 1,
			},

			// Punctuation - single character tokens
			{
				Name:     TOKEN_LPAREN,
				Pattern:  grammar.Literal("("),
				Priority: 1,
			},
			{
				Name:     TOKEN_RPAREN,
				Pattern:  grammar.Literal(")"),
				Priority: 1,
			},
			{
				Name:     TOKEN_COMMA,
				Pattern:  grammar.Literal(","),
				Priority: 1,
			},

			// Whitespace - one or more whitespace characters
			{
				Name:     TOKEN_WHITESPACE,
				Pattern:  grammar.LexOneOrMore{Inner: whitespaceChar},
				Priority: 1,
			},
		},
	}
}
