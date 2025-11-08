package langdef

import (
	"fmt"
	"sort"
	"testing"

	"github.com/shadowCow/cow-lang-go/tooling/automata"
	"github.com/shadowCow/cow-lang-go/tooling/grammar"
	"github.com/shadowCow/cow-lang-go/tooling/ll1"
)

// TestSimplePattern tests compiling a very simple pattern.
func TestSimplePattern(t *testing.T) {
	// Create a simple grammar with just one token
	lexGrammar := grammar.LexicalGrammar{
		Tokens: []grammar.TokenDefinition{
			{
				Name:     "DIGIT",
				Pattern:  grammar.CharRange{From: '0', To: '9'},
				Priority: 1,
			},
		},
	}

	dfa := automata.CompileLexicalGrammar(lexGrammar)

	fmt.Printf("DFA Initial State: %s\n", dfa.InitialState)
	fmt.Printf("DFA States: %d\n", len(dfa.States))
	fmt.Printf("DFA Accepting States: %d\n", len(dfa.AcceptingStates))

	// Print all states and their transitions
	for stateName, state := range dfa.States {
		fmt.Printf("\nState %s:\n", stateName)
		for char, next := range state.Transitions {
			fmt.Printf("  '%c' -> %s\n", char, next)
		}
		if len(state.Transitions) == 0 {
			fmt.Printf("  (no transitions)\n")
		}
	}

	// Print accepting states
	fmt.Printf("\nAccepting States:\n")
	for stateName, acceptInfo := range dfa.AcceptingStates {
		fmt.Printf("  %s: token=%s, priority=%d\n", stateName, acceptInfo.TokenType, acceptInfo.Priority)
	}

	if len(dfa.States) == 0 {
		t.Fatal("DFA has no states")
	}

	if len(dfa.AcceptingStates) == 0 {
		t.Fatal("DFA has no accepting states")
	}
}

func TestPrintFirstAndFollowSets(t *testing.T) {
	g := GetSyntacticGrammar()

	firstSets := ll1.ComputeFirstSets(g)
	followSets := ll1.ComputeFollowSets(g, firstSets)

	fmt.Println("\n=== FIRST SETS ===")
	for _, symbol := range []string{"Factor", "FactorRest", "Expression", "Arguments"} {
		first := firstSets.Get(symbol)
		keys := make([]string, 0, len(first))
		for k := range first {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Printf("FIRST(%s) = %v\n", symbol, keys)
		fmt.Printf("  Nullable: %v\n", firstSets.IsNullable(grammar.Symbol(symbol)))
	}

	fmt.Println("\n=== FOLLOW SETS ===")
	for _, symbol := range []string{"Factor", "FactorRest", "Expression", "Arguments"} {
		follow := followSets.Get(grammar.Symbol(symbol))
		keys := make([]string, 0, len(follow))
		for k := range follow {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Printf("FOLLOW(%s) = %v\n", symbol, keys)
	}
}
