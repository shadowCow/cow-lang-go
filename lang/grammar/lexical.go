package grammar

// TokenType represents a category of tokens in the lexical grammar.
type TokenType string

// LexicalGrammar defines how characters are transformed into tokens.
type LexicalGrammar struct {
	Tokens []TokenDefinition
}

// TokenDefinition defines a single token type and how to recognize it.
type TokenDefinition struct {
	Name     TokenType
	Pattern  LexicalPattern
	Priority int // Higher priority wins when multiple patterns match
}

// LexicalPattern is a marker interface for all lexical pattern types.
type LexicalPattern interface {
	IsLexicalPattern()
}

// Literal matches an exact sequence of characters.
type Literal string

func (Literal) IsLexicalPattern() {}

// CharSet matches any single character from the provided set.
type CharSet []rune

func (CharSet) IsLexicalPattern() {}

// CharRange matches any single character within the inclusive range.
type CharRange struct {
	From rune
	To   rune
}

func (CharRange) IsLexicalPattern() {}

// AnyChar matches any single character.
type AnyChar struct{}

func (AnyChar) IsLexicalPattern() {}

// AnyCharExcept matches any single character except those in the set.
type AnyCharExcept []rune

func (AnyCharExcept) IsLexicalPattern() {}

// LexSequence matches a series of patterns in order.
type LexSequence []LexicalPattern

func (LexSequence) IsLexicalPattern() {}

// LexAlternative matches one of several patterns.
type LexAlternative []LexicalPattern

func (LexAlternative) IsLexicalPattern() {}

// LexOptional matches zero or one occurrence of the pattern.
type LexOptional struct {
	Inner LexicalPattern
}

func (LexOptional) IsLexicalPattern() {}

// LexZeroOrMore matches zero or more repetitions of the pattern.
type LexZeroOrMore struct {
	Inner LexicalPattern
}

func (LexZeroOrMore) IsLexicalPattern() {}

// LexOneOrMore matches one or more repetitions of the pattern.
type LexOneOrMore struct {
	Inner LexicalPattern
}

func (LexOneOrMore) IsLexicalPattern() {}
