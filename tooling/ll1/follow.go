package ll1

import (
	"github.com/shadowCow/cow-lang-go/tooling/grammar"
)

// FollowSets holds the FOLLOW sets for all non-terminals in a grammar.
// FOLLOW(X) is the set of terminals that can appear immediately after X in a derivation.
type FollowSets struct {
	// Map from non-terminal symbol to its FOLLOW set
	sets map[grammar.Symbol]map[string]bool
}

// NewFollowSets creates an empty FollowSets structure.
func NewFollowSets() *FollowSets {
	return &FollowSets{
		sets: make(map[grammar.Symbol]map[string]bool),
	}
}

// Get returns the FOLLOW set for a non-terminal symbol.
func (fs *FollowSets) Get(symbol grammar.Symbol) map[string]bool {
	if set, ok := fs.sets[symbol]; ok {
		return set
	}
	return make(map[string]bool)
}

// EndOfInputMarker is the special symbol representing end of input.
const EndOfInputMarker = "$"

// ComputeFollowSets computes FOLLOW sets for all non-terminals in the grammar.
// Requires FIRST sets to be computed first.
func ComputeFollowSets(g grammar.SyntacticGrammar, firstSets *FirstSets) *FollowSets {
	fs := NewFollowSets()

	// Initialize FOLLOW sets for all non-terminals
	for symbol := range g.Productions {
		fs.sets[symbol] = make(map[string]bool)
	}

	// Rule 1: FOLLOW(StartSymbol) includes EOF ($)
	fs.sets[g.StartSymbol][EndOfInputMarker] = true

	// Iterate until fixpoint
	changed := true
	for changed {
		changed = false

		for symbol, production := range g.Productions {
			if fs.addFollowsFromProduction(symbol, production, g, firstSets) {
				changed = true
			}
		}
	}

	return fs
}

// addFollowsFromProduction adds FOLLOW set entries based on a production rule.
// Returns true if any FOLLOW sets were modified.
func (fs *FollowSets) addFollowsFromProduction(
	leftSide grammar.Symbol,
	production grammar.ProductionRule,
	g grammar.SyntacticGrammar,
	firstSets *FirstSets,
) bool {
	changed := false

	switch p := production.(type) {
	case grammar.Terminal:
		// Terminals don't contribute to FOLLOW sets
		return false

	case grammar.NonTerminal:
		// A -> B means FOLLOW(A) ⊆ FOLLOW(B)
		changed = fs.addToFollow(p.Symbol, fs.Get(leftSide))

	case grammar.SynSequence:
		// For A -> α B β:
		// - FIRST(β) - {ε} is added to FOLLOW(B)
		// - If β is nullable, FOLLOW(A) is added to FOLLOW(B)
		for i, elem := range p {
			// Find all non-terminals in this element
			nonterminals := collectNonTerminalsFromProduction(elem)

			// Compute what can follow: FIRST of everything after this element
			following := p[i+1:]
			firstOfFollowing, nullableFollowing := computeFirstOfSequence(following, firstSets)

			// Add FIRST(following) to FOLLOW of each non-terminal in elem
			for _, nt := range nonterminals {
				if fs.addToFollow(nt, firstOfFollowing) {
					changed = true
				}

				// If following is nullable, add FOLLOW(leftSide) to FOLLOW(nt)
				if nullableFollowing {
					if fs.addToFollow(nt, fs.Get(leftSide)) {
						changed = true
					}
				}
			}
		}

	case grammar.SynAlternative:
		// For A -> B | C | D, process each alternative
		for _, alt := range p {
			if fs.addFollowsFromProduction(leftSide, alt, g, firstSets) {
				changed = true
			}
		}

	case grammar.SynOptional:
		// A -> B? means FOLLOW(A) ⊆ FOLLOW(B)
		nonterminals := collectNonTerminalsFromProduction(p.Inner)
		for _, nt := range nonterminals {
			if fs.addToFollow(nt, fs.Get(leftSide)) {
				changed = true
			}
		}

	case grammar.SynZeroOrMore:
		// A -> B* means FOLLOW(A) ⊆ FOLLOW(B)
		// Also B can be followed by FIRST(B) due to repetition
		nonterminals := collectNonTerminalsFromProduction(p.Inner)
		firstOfInner, _ := firstSets.computeFirstOfProduction(p.Inner)
		for _, nt := range nonterminals {
			if fs.addToFollow(nt, fs.Get(leftSide)) {
				changed = true
			}
			if fs.addToFollow(nt, firstOfInner) {
				changed = true
			}
		}

	case grammar.SynOneOrMore:
		// A -> B+ means FOLLOW(A) ⊆ FOLLOW(B)
		// Also B can be followed by FIRST(B) due to repetition
		nonterminals := collectNonTerminalsFromProduction(p.Inner)
		firstOfInner, _ := firstSets.computeFirstOfProduction(p.Inner)
		for _, nt := range nonterminals {
			if fs.addToFollow(nt, fs.Get(leftSide)) {
				changed = true
			}
			if fs.addToFollow(nt, firstOfInner) {
				changed = true
			}
		}
	}

	return changed
}

// addToFollow adds terminals from 'toAdd' to the FOLLOW set of 'symbol'.
// Returns true if the FOLLOW set was modified.
func (fs *FollowSets) addToFollow(symbol grammar.Symbol, toAdd map[string]bool) bool {
	if fs.sets[symbol] == nil {
		fs.sets[symbol] = make(map[string]bool)
	}

	oldSize := len(fs.sets[symbol])
	for term := range toAdd {
		fs.sets[symbol][term] = true
	}
	return len(fs.sets[symbol]) != oldSize
}

// collectNonTerminalsFromProduction recursively finds all non-terminals in a production.
func collectNonTerminalsFromProduction(prod grammar.ProductionRule) []grammar.Symbol {
	var result []grammar.Symbol

	switch p := prod.(type) {
	case grammar.Terminal:
		// No non-terminals
	case grammar.NonTerminal:
		result = append(result, p.Symbol)
	case grammar.SynSequence:
		for _, elem := range p {
			result = append(result, collectNonTerminalsFromProduction(elem)...)
		}
	case grammar.SynAlternative:
		for _, alt := range p {
			result = append(result, collectNonTerminalsFromProduction(alt)...)
		}
	case grammar.SynOptional:
		result = append(result, collectNonTerminalsFromProduction(p.Inner)...)
	case grammar.SynZeroOrMore:
		result = append(result, collectNonTerminalsFromProduction(p.Inner)...)
	case grammar.SynOneOrMore:
		result = append(result, collectNonTerminalsFromProduction(p.Inner)...)
	}

	return result
}

// computeFirstOfSequence computes the FIRST set for a sequence of production rules.
// Returns (first_set, is_nullable).
func computeFirstOfSequence(seq []grammar.ProductionRule, firstSets *FirstSets) (map[string]bool, bool) {
	result := make(map[string]bool)
	nullable := true

	for _, elem := range seq {
		firstElem, nullableElem := firstSets.computeFirstOfProduction(elem)
		for term := range firstElem {
			result[term] = true
		}
		if !nullableElem {
			nullable = false
			break
		}
	}

	// If sequence is empty, it's nullable
	if len(seq) == 0 {
		nullable = true
	}

	return result, nullable
}
