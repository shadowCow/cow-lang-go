package langdef

import (
	"fmt"
	"testing"

	"github.com/shadowCow/cow-lang-go/tooling/automata"
	"github.com/shadowCow/cow-lang-go/tooling/grammar"
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
