// Package ll1 implements LL(1) parser generation algorithms.
// This includes FIRST/FOLLOW set computation and parse table generation.
package ll1

import (
	"github.com/shadowCow/cow-lang-go/tooling/grammar"
)

// FirstSets holds the FIRST sets for all symbols in a grammar.
// FIRST(X) is the set of terminals that can begin strings derivable from X.
type FirstSets struct {
	// Map from symbol (terminal or non-terminal) to its FIRST set
	sets map[string]map[string]bool
	// Tracks which non-terminals can derive epsilon (empty string)
	nullable map[grammar.Symbol]bool
}

// NewFirstSets creates an empty FirstSets structure.
func NewFirstSets() *FirstSets {
	return &FirstSets{
		sets:     make(map[string]map[string]bool),
		nullable: make(map[grammar.Symbol]bool),
	}
}

// Get returns the FIRST set for a symbol (terminal or non-terminal).
func (fs *FirstSets) Get(symbol string) map[string]bool {
	if set, ok := fs.sets[symbol]; ok {
		return set
	}
	return make(map[string]bool)
}

// IsNullable returns true if a non-terminal can derive epsilon.
func (fs *FirstSets) IsNullable(symbol grammar.Symbol) bool {
	return fs.nullable[symbol]
}

// ComputeFirstSets computes FIRST sets for all symbols in the grammar.
// Returns the FirstSets structure.
func ComputeFirstSets(g grammar.SyntacticGrammar) *FirstSets {
	fs := NewFirstSets()

	// Initialize: For each terminal, FIRST(terminal) = {terminal}
	// We need to discover all terminals by traversing the grammar
	terminals := collectTerminals(g)
	for _, term := range terminals {
		termKey := string(term)
		fs.sets[termKey] = map[string]bool{termKey: true}
	}

	// Iteratively compute FIRST sets for non-terminals until fixpoint
	changed := true
	for changed {
		changed = false

		for symbol, production := range g.Productions {
			symbolKey := string(symbol)
			if fs.sets[symbolKey] == nil {
				fs.sets[symbolKey] = make(map[string]bool)
			}

			oldSize := len(fs.sets[symbolKey])
			oldNullable := fs.nullable[symbol]

			// Compute FIRST for this production and add to symbol's FIRST set
			firstSet, nullable := fs.computeFirstOfProduction(production)
			for term := range firstSet {
				fs.sets[symbolKey][term] = true
			}
			if nullable {
				fs.nullable[symbol] = true
			}

			// Check if anything changed
			if len(fs.sets[symbolKey]) != oldSize || fs.nullable[symbol] != oldNullable {
				changed = true
			}
		}
	}

	return fs
}

// computeFirstOfProduction computes FIRST set for a production rule.
// Returns (first_set, is_nullable).
func (fs *FirstSets) computeFirstOfProduction(prod grammar.ProductionRule) (map[string]bool, bool) {
	result := make(map[string]bool)
	nullable := false

	switch p := prod.(type) {
	case grammar.Terminal:
		// FIRST(terminal) = {terminal}
		termKey := string(p.TokenType)
		result[termKey] = true
		nullable = false

	case grammar.NonTerminal:
		// FIRST(NonTerminal) = FIRST(Symbol)
		symbolKey := string(p.Symbol)
		for term := range fs.Get(symbolKey) {
			result[term] = true
		}
		nullable = fs.IsNullable(p.Symbol)

	case grammar.SynSequence:
		// FIRST(A B C) = FIRST(A) if A not nullable,
		//                else FIRST(A) ∪ FIRST(B) if B not nullable, etc.
		nullable = true
		for _, elem := range p {
			firstElem, nullableElem := fs.computeFirstOfProduction(elem)
			for term := range firstElem {
				result[term] = true
			}
			if !nullableElem {
				nullable = false
				break
			}
		}
		// If all elements are nullable, sequence is nullable

	case grammar.SynAlternative:
		// FIRST(A | B | C) = FIRST(A) ∪ FIRST(B) ∪ FIRST(C)
		nullable = false
		for _, alt := range p {
			firstAlt, nullableAlt := fs.computeFirstOfProduction(alt)
			for term := range firstAlt {
				result[term] = true
			}
			if nullableAlt {
				nullable = true
			}
		}

	case grammar.SynOptional:
		// FIRST(A?) = FIRST(A)
		// Always nullable since it's optional
		firstInner, _ := fs.computeFirstOfProduction(p.Inner)
		for term := range firstInner {
			result[term] = true
		}
		nullable = true

	case grammar.SynZeroOrMore:
		// FIRST(A*) = FIRST(A)
		// Always nullable since it can match zero times
		firstInner, _ := fs.computeFirstOfProduction(p.Inner)
		for term := range firstInner {
			result[term] = true
		}
		nullable = true

	case grammar.SynOneOrMore:
		// FIRST(A+) = FIRST(A)
		// Nullable if A is nullable
		firstInner, nullableInner := fs.computeFirstOfProduction(p.Inner)
		for term := range firstInner {
			result[term] = true
		}
		nullable = nullableInner
	}

	return result, nullable
}

// collectTerminals traverses the grammar and collects all terminal token types.
func collectTerminals(g grammar.SyntacticGrammar) []grammar.TokenType {
	terminals := make(map[grammar.TokenType]bool)

	for _, production := range g.Productions {
		collectTerminalsFromProduction(production, terminals)
	}

	result := make([]grammar.TokenType, 0, len(terminals))
	for term := range terminals {
		result = append(result, term)
	}
	return result
}

// collectTerminalsFromProduction recursively finds terminals in a production.
func collectTerminalsFromProduction(prod grammar.ProductionRule, terminals map[grammar.TokenType]bool) {
	switch p := prod.(type) {
	case grammar.Terminal:
		terminals[p.TokenType] = true
	case grammar.NonTerminal:
		// Nothing to collect from non-terminals
	case grammar.SynSequence:
		for _, elem := range p {
			collectTerminalsFromProduction(elem, terminals)
		}
	case grammar.SynAlternative:
		for _, alt := range p {
			collectTerminalsFromProduction(alt, terminals)
		}
	case grammar.SynOptional:
		collectTerminalsFromProduction(p.Inner, terminals)
	case grammar.SynZeroOrMore:
		collectTerminalsFromProduction(p.Inner, terminals)
	case grammar.SynOneOrMore:
		collectTerminalsFromProduction(p.Inner, terminals)
	}
}
