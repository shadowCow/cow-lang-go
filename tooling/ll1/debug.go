package ll1

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/shadowCow/cow-lang-go/tooling/grammar"
)

// PrintFirstSets prints the FIRST sets in a readable format.
func PrintFirstSets(firstSets *FirstSets, out io.Writer) {
	fmt.Fprintln(out, "FIRST SETS:")
	fmt.Fprintln(out, "===========")

	// Get all symbols and sort for consistent output
	var symbols []string
	for symbol := range firstSets.sets {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	for _, symbol := range symbols {
		firstSet := firstSets.Get(symbol)
		terminals := make([]string, 0, len(firstSet))
		for term := range firstSet {
			terminals = append(terminals, term)
		}
		sort.Strings(terminals)

		nullable := ""
		if sym, ok := grammar.Symbol(symbol), false; ok || true {
			if firstSets.IsNullable(sym) {
				nullable = " [nullable]"
			}
		}

		fmt.Fprintf(out, "  FIRST(%s) = {%s}%s\n", symbol, strings.Join(terminals, ", "), nullable)
	}
	fmt.Fprintln(out, "")
}

// PrintFollowSets prints the FOLLOW sets in a readable format.
func PrintFollowSets(followSets *FollowSets, out io.Writer) {
	fmt.Fprintln(out, "FOLLOW SETS:")
	fmt.Fprintln(out, "============")

	// Get all non-terminals and sort for consistent output
	var symbols []grammar.Symbol
	for symbol := range followSets.sets {
		symbols = append(symbols, symbol)
	}
	sort.Slice(symbols, func(i, j int) bool {
		return string(symbols[i]) < string(symbols[j])
	})

	for _, symbol := range symbols {
		followSet := followSets.Get(symbol)
		terminals := make([]string, 0, len(followSet))
		for term := range followSet {
			terminals = append(terminals, term)
		}
		sort.Strings(terminals)

		fmt.Fprintf(out, "  FOLLOW(%s) = {%s}\n", symbol, strings.Join(terminals, ", "))
	}
	fmt.Fprintln(out, "")
}

// PrintParseTable prints the LL(1) parse table as a grid.
func PrintParseTable(table *ParseTable, out io.Writer) {
	fmt.Fprintln(out, "LL(1) PARSE TABLE:")
	fmt.Fprintln(out, "==================")

	if len(table.nonTerminals) == 0 || len(table.terminals) == 0 {
		fmt.Fprintln(out, "  (empty table)")
		return
	}

	// Sort for consistent output
	nonTerminals := make([]string, len(table.nonTerminals))
	for i, nt := range table.nonTerminals {
		nonTerminals[i] = string(nt)
	}
	sort.Strings(nonTerminals)

	terminals := make([]string, len(table.terminals))
	copy(terminals, table.terminals)
	sort.Strings(terminals)

	// Calculate column widths
	ntColWidth := 10
	for _, nt := range nonTerminals {
		if len(nt) > ntColWidth {
			ntColWidth = len(nt)
		}
	}

	termColWidth := 15
	for _, term := range terminals {
		if len(term) > termColWidth {
			termColWidth = len(term)
		}
	}

	// Print header row
	fmt.Fprintf(out, "  %*s |", ntColWidth, "")
	for _, term := range terminals {
		fmt.Fprintf(out, " %-*s |", termColWidth, term)
	}
	fmt.Fprintln(out, "")

	// Print separator
	fmt.Fprintf(out, "  %s-+", strings.Repeat("-", ntColWidth))
	for range terminals {
		fmt.Fprintf(out, "-%s-+", strings.Repeat("-", termColWidth))
	}
	fmt.Fprintln(out, "")

	// Print table rows
	for _, nt := range nonTerminals {
		fmt.Fprintf(out, "  %-*s |", ntColWidth, nt)
		for _, term := range terminals {
			key := tableKey{grammar.Symbol(nt), term}
			if prod, exists := table.table[key]; exists {
				prodStr := formatProductionShort(prod)
				if len(prodStr) > termColWidth {
					prodStr = prodStr[:termColWidth-2] + ".."
				}
				fmt.Fprintf(out, " %-*s |", termColWidth, prodStr)
			} else {
				fmt.Fprintf(out, " %-*s |", termColWidth, "")
			}
		}
		fmt.Fprintln(out, "")
	}
	fmt.Fprintln(out, "")
}

// formatProductionShort returns a short string representation of a production.
func formatProductionShort(prod grammar.ProductionRule) string {
	switch p := prod.(type) {
	case grammar.Terminal:
		return string(p.TokenType)
	case grammar.NonTerminal:
		return string(p.Symbol)
	case grammar.SynSequence:
		if len(p) == 0 {
			return "ε"
		}
		parts := make([]string, len(p))
		for i, elem := range p {
			parts[i] = formatProductionShort(elem)
		}
		return strings.Join(parts, " ")
	case grammar.SynAlternative:
		parts := make([]string, len(p))
		for i, alt := range p {
			parts[i] = formatProductionShort(alt)
		}
		return strings.Join(parts, "|")
	case grammar.SynOptional:
		return formatProductionShort(p.Inner) + "?"
	case grammar.SynZeroOrMore:
		return formatProductionShort(p.Inner) + "*"
	case grammar.SynOneOrMore:
		return formatProductionShort(p.Inner) + "+"
	default:
		return "?"
	}
}

// PrintGrammar prints the grammar in a readable format.
func PrintGrammar(g grammar.SyntacticGrammar, out io.Writer) {
	fmt.Fprintln(out, "GRAMMAR:")
	fmt.Fprintln(out, "========")
	fmt.Fprintf(out, "Start symbol: %s\n\n", g.StartSymbol)
	fmt.Fprintln(out, "Productions:")

	// Sort non-terminals for consistent output
	var symbols []grammar.Symbol
	for symbol := range g.Productions {
		symbols = append(symbols, symbol)
	}
	sort.Slice(symbols, func(i, j int) bool {
		return string(symbols[i]) < string(symbols[j])
	})

	for _, symbol := range symbols {
		production := g.Productions[symbol]
		fmt.Fprintf(out, "  %s -> %s\n", symbol, formatProduction(production))
	}
	fmt.Fprintln(out, "")
}

// PrintParseTrace prints a trace of the parsing process.
// This is called from the parser when tracing is enabled.
type ParseTracer struct {
	stepNum int
}

// NewParseTracer creates a new parse tracer.
func NewParseTracer() *ParseTracer {
	return &ParseTracer{stepNum: 0}
}

// Step prints a parse step.
func (pt *ParseTracer) Step(stack []string, input string, action string, out io.Writer) {
	pt.stepNum++
	fmt.Fprintf(out, "Step %d:\n", pt.stepNum)
	fmt.Fprintf(out, "  Stack:  [%s]\n", strings.Join(stack, ", "))
	fmt.Fprintf(out, "  Input:  %s\n", input)
	fmt.Fprintf(out, "  Action: %s\n\n", action)
}
