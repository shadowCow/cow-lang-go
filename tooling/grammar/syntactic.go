package grammar

// Symbol represents a non-terminal symbol in the syntactic grammar.
type Symbol string

// SyntacticGrammar defines how tokens are transformed into a parse tree.
type SyntacticGrammar struct {
	Productions map[Symbol]ProductionRule
	StartSymbol Symbol
}

// ProductionRule is a marker interface for all production rule types.
type ProductionRule interface {
	IsProductionRule()
}

// Terminal references a token type from the lexical grammar.
type Terminal struct {
	TokenType TokenType
}

func (Terminal) IsProductionRule() {}

// NonTerminal references another production rule by its symbol.
type NonTerminal struct {
	Symbol Symbol
}

func (NonTerminal) IsProductionRule() {}

// SynSequence matches a series of production rules in order.
type SynSequence []ProductionRule

func (SynSequence) IsProductionRule() {}

// SynAlternative matches one of several production rules.
type SynAlternative []ProductionRule

func (SynAlternative) IsProductionRule() {}

// SynOptional matches zero or one occurrence of the production rule.
type SynOptional struct {
	Inner ProductionRule
}

func (SynOptional) IsProductionRule() {}

// SynZeroOrMore matches zero or more repetitions of the production rule.
type SynZeroOrMore struct {
	Inner ProductionRule
}

func (SynZeroOrMore) IsProductionRule() {}

// SynOneOrMore matches one or more repetitions of the production rule.
type SynOneOrMore struct {
	Inner ProductionRule
}

func (SynOneOrMore) IsProductionRule() {}
