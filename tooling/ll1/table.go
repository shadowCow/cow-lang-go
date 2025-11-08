package ll1

import (
	"fmt"
	"strings"

	"github.com/shadowCow/cow-lang-go/tooling/grammar"
)

// ParseTable represents an LL(1) parse table.
// M[NonTerminal, Terminal] -> ProductionRule
type ParseTable struct {
	// Map from (non-terminal, terminal) to production rule
	table map[tableKey]grammar.ProductionRule
	// Track all non-terminals and terminals for visualization
	nonTerminals []grammar.Symbol
	terminals    []string
}

// tableKey is a composite key for the parse table.
type tableKey struct {
	nonTerminal grammar.Symbol
	terminal    string
}

// NewParseTable creates an empty parse table.
func NewParseTable() *ParseTable {
	return &ParseTable{
		table:        make(map[tableKey]grammar.ProductionRule),
		nonTerminals: []grammar.Symbol{},
		terminals:    []string{},
	}
}

// Get returns the production to use for a given (non-terminal, lookahead) pair.
// Returns nil if no production is defined.
func (pt *ParseTable) Get(nonTerminal grammar.Symbol, lookahead string) grammar.ProductionRule {
	key := tableKey{nonTerminal, lookahead}
	return pt.table[key]
}

// Conflict represents an LL(1) conflict in the grammar.
type Conflict struct {
	NonTerminal  grammar.Symbol
	Lookahead    string
	Productions  []grammar.ProductionRule
	Reason       string
}

// Error returns a formatted error message for the conflict.
func (c *Conflict) Error() string {
	return fmt.Sprintf("LL(1) conflict at [%s, %s]: %s\n  Multiple productions possible:\n%s",
		c.NonTerminal, c.Lookahead, c.Reason, c.formatProductions())
}

// formatProductions formats the conflicting productions for display.
func (c *Conflict) formatProductions() string {
	var lines []string
	for i, prod := range c.Productions {
		lines = append(lines, fmt.Sprintf("    %d. %s -> %s", i+1, c.NonTerminal, formatProduction(prod)))
	}
	return strings.Join(lines, "\n")
}

// BuildParseTable constructs an LL(1) parse table from a grammar.
// Returns an error if the grammar is not LL(1) (i.e., has conflicts).
func BuildParseTable(g grammar.SyntacticGrammar, firstSets *FirstSets, followSets *FollowSets) (*ParseTable, error) {
	pt := NewParseTable()

	// Collect all non-terminals and terminals for later visualization
	for symbol := range g.Productions {
		pt.nonTerminals = append(pt.nonTerminals, symbol)
	}
	terminalMap := make(map[string]bool)
	for _, term := range collectTerminals(g) {
		terminalMap[string(term)] = true
	}
	terminalMap[EndOfInputMarker] = true // Add EOF marker
	for term := range terminalMap {
		pt.terminals = append(pt.terminals, term)
	}

	// Track conflicts
	var conflicts []Conflict

	// For each production A -> α
	for nonTerminal, production := range g.Productions {
		// Build table entries for this production
		newConflicts := pt.addProductionToTable(nonTerminal, production, firstSets, followSets)
		conflicts = append(conflicts, newConflicts...)
	}

	// If there are conflicts, return error with details
	if len(conflicts) > 0 {
		return nil, &GrammarNotLL1Error{Conflicts: conflicts}
	}

	return pt, nil
}

// GrammarNotLL1Error indicates the grammar has LL(1) conflicts.
type GrammarNotLL1Error struct {
	Conflicts []Conflict
}

// Error implements the error interface.
func (e *GrammarNotLL1Error) Error() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Grammar is not LL(1): found %d conflict(s)", len(e.Conflicts)))
	for i, conflict := range e.Conflicts {
		lines = append(lines, fmt.Sprintf("\nConflict %d:", i+1))
		lines = append(lines, "  "+strings.ReplaceAll(conflict.Error(), "\n", "\n  "))
	}
	return strings.Join(lines, "\n")
}

// addProductionToTable adds entries to the parse table for a production.
// Returns any conflicts detected.
func (pt *ParseTable) addProductionToTable(
	nonTerminal grammar.Symbol,
	production grammar.ProductionRule,
	firstSets *FirstSets,
	followSets *FollowSets,
) []Conflict {
	var conflicts []Conflict

	switch p := production.(type) {
	case grammar.Terminal:
		// A -> a: add to M[A, a]
		terminal := string(p.TokenType)
		conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)

	case grammar.NonTerminal:
		// A -> B: add to M[A, t] for all t in FIRST(B)
		firstB := firstSets.Get(string(p.Symbol))
		for terminal := range firstB {
			conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
		}
		// If B is nullable, add to M[A, t] for all t in FOLLOW(A)
		if firstSets.IsNullable(p.Symbol) {
			for terminal := range followSets.Get(nonTerminal) {
				conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
			}
		}

	case grammar.SynSequence:
		// A -> α: add to M[A, t] for all t in FIRST(α)
		firstSeq, nullableSeq := firstSets.computeFirstOfProduction(production)
		for terminal := range firstSeq {
			conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
		}
		// If α is nullable, add to M[A, t] for all t in FOLLOW(A)
		if nullableSeq {
			for terminal := range followSets.Get(nonTerminal) {
				conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
			}
		}

	case grammar.SynAlternative:
		// A -> B | C | D: process each alternative
		// This is where we commonly find conflicts!
		for _, alt := range p {
			conflicts = append(conflicts, pt.addProductionToTable(nonTerminal, alt, firstSets, followSets)...)
		}

	case grammar.SynOptional:
		// A -> B?: add to M[A, t] for all t in FIRST(B)
		firstInner, _ := firstSets.computeFirstOfProduction(p.Inner)
		for terminal := range firstInner {
			conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
		}
		// Since optional is always nullable, add to M[A, t] for all t in FOLLOW(A)
		for terminal := range followSets.Get(nonTerminal) {
			conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
		}

	case grammar.SynZeroOrMore:
		// A -> B*: similar to optional
		firstInner, _ := firstSets.computeFirstOfProduction(p.Inner)
		for terminal := range firstInner {
			conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
		}
		for terminal := range followSets.Get(nonTerminal) {
			conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
		}

	case grammar.SynOneOrMore:
		// A -> B+: add to M[A, t] for all t in FIRST(B)
		firstInner, nullableInner := firstSets.computeFirstOfProduction(p.Inner)
		for terminal := range firstInner {
			conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
		}
		// If B is nullable, add to M[A, t] for all t in FOLLOW(A)
		if nullableInner {
			for terminal := range followSets.Get(nonTerminal) {
				conflicts = append(conflicts, pt.addEntry(nonTerminal, terminal, production)...)
			}
		}
	}

	return conflicts
}

// addEntry adds an entry to the parse table and detects conflicts.
// Returns a conflict if the cell is already occupied by a different production.
func (pt *ParseTable) addEntry(nonTerminal grammar.Symbol, terminal string, production grammar.ProductionRule) []Conflict {
	key := tableKey{nonTerminal, terminal}

	// Check if cell is already occupied
	if existing, exists := pt.table[key]; exists {
		// Check if it's the same production (not a conflict)
		if !sameProduction(existing, production) {
			return []Conflict{{
				NonTerminal: nonTerminal,
				Lookahead:   terminal,
				Productions: []grammar.ProductionRule{existing, production},
				Reason:      "Multiple different productions can start with this lookahead",
			}}
		}
		// Same production, no conflict
		return nil
	}

	// Add to table
	pt.table[key] = production
	return nil
}

// sameProduction checks if two productions are the same.
// This is a simple equality check - in a more sophisticated system,
// we might want structural equality.
func sameProduction(a, b grammar.ProductionRule) bool {
	return fmt.Sprintf("%#v", a) == fmt.Sprintf("%#v", b)
}

// formatProduction returns a string representation of a production for display.
func formatProduction(prod grammar.ProductionRule) string {
	switch p := prod.(type) {
	case grammar.Terminal:
		return string(p.TokenType)
	case grammar.NonTerminal:
		return string(p.Symbol)
	case grammar.SynSequence:
		parts := make([]string, len(p))
		for i, elem := range p {
			parts[i] = formatProduction(elem)
		}
		return strings.Join(parts, " ")
	case grammar.SynAlternative:
		parts := make([]string, len(p))
		for i, alt := range p {
			parts[i] = formatProduction(alt)
		}
		return "(" + strings.Join(parts, " | ") + ")"
	case grammar.SynOptional:
		return formatProduction(p.Inner) + "?"
	case grammar.SynZeroOrMore:
		return formatProduction(p.Inner) + "*"
	case grammar.SynOneOrMore:
		return formatProduction(p.Inner) + "+"
	default:
		return "?"
	}
}
