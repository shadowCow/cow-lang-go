package automata

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shadowCow/cow-lang-go/lang/grammar"
)

// NFAToDFA converts an NFA to a DFA using subset construction.
func NFAToDFA(nfa *NFA) Dfa {
	// Compute epsilon closure of start state
	startClosure := epsilonClosure(nfa, map[int]bool{nfa.Start: true})

	// DFA states are sets of NFA states
	// We'll use a string representation of the set as the DFA state name
	dfa := Dfa{
		InitialState: stateSetToString(startClosure),
		States:       make(map[string]DfaState),
	}

	// Queue of DFA states to process (as NFA state sets)
	queue := []map[int]bool{startClosure}
	processed := make(map[string]bool)

	for len(queue) > 0 {
		// Pop from queue
		currentSet := queue[0]
		queue = queue[1:]

		currentName := stateSetToString(currentSet)
		if processed[currentName] {
			continue
		}
		processed[currentName] = true

		// Find all possible transitions from this set
		transitions := make(map[rune]string)
		symbolsMap := make(map[rune]map[int]bool)

		// Collect all possible symbols and their target states
		for stateID := range currentSet {
			state := nfa.States[stateID]
			for symbol, targets := range state.Transitions {
				if symbolsMap[symbol] == nil {
					symbolsMap[symbol] = make(map[int]bool)
				}
				for target := range targets {
					symbolsMap[symbol][target] = true
				}
			}
		}

		// For each symbol, compute epsilon closure of target states
		for symbol, targets := range symbolsMap {
			closure := epsilonClosure(nfa, targets)
			nextName := stateSetToString(closure)
			transitions[symbol] = nextName

			// Add to queue if not processed
			if !processed[nextName] {
				queue = append(queue, closure)
			}
		}

		// Create DFA state
		dfa.States[currentName] = DfaState{
			Name:              currentName,
			Transitions:       transitions,
			DefaultTransition: "", // Dead state - no transition
		}
	}

	return dfa
}

// NFAToDFAWithTokens converts an NFA with token information to a DFA.
// Accepting states in the DFA remember which token they matched.
func NFAToDFAWithTokens(nfa *NFA) DfaWithTokens {
	// Compute epsilon closure of start state
	startClosure := epsilonClosure(nfa, map[int]bool{nfa.Start: true})

	// DFA states are sets of NFA states
	dfa := DfaWithTokens{
		InitialState:    stateSetToString(startClosure),
		States:          make(map[string]DfaStateWithToken),
		AcceptingStates: make(map[string]AcceptingState),
	}

	// Queue of DFA states to process
	queue := []map[int]bool{startClosure}
	processed := make(map[string]bool)

	for len(queue) > 0 {
		// Pop from queue
		currentSet := queue[0]
		queue = queue[1:]

		currentName := stateSetToString(currentSet)
		if processed[currentName] {
			continue
		}
		processed[currentName] = true

		// Check if this DFA state contains any NFA accept states
		var tokenType grammar.TokenType
		maxPriority := -1
		isAccepting := false

		for stateID := range currentSet {
			if acceptInfo, ok := nfa.AcceptStates[stateID]; ok {
				isAccepting = true
				if acceptInfo.Priority > maxPriority {
					maxPriority = acceptInfo.Priority
					tokenType = acceptInfo.TokenType
				}
			}
		}

		// Find all possible transitions
		transitions := make(map[rune]string)
		symbolsMap := make(map[rune]map[int]bool)

		for stateID := range currentSet {
			state := nfa.States[stateID]
			for symbol, targets := range state.Transitions {
				if symbolsMap[symbol] == nil {
					symbolsMap[symbol] = make(map[int]bool)
				}
				for target := range targets {
					symbolsMap[symbol][target] = true
				}
			}
		}

		// For each symbol, compute epsilon closure
		for symbol, targets := range symbolsMap {
			closure := epsilonClosure(nfa, targets)
			nextName := stateSetToString(closure)
			transitions[symbol] = nextName

			if !processed[nextName] {
				queue = append(queue, closure)
			}
		}

		// Create DFA state
		dfa.States[currentName] = DfaStateWithToken{
			Name:              currentName,
			Transitions:       transitions,
			DefaultTransition: "",
		}

		// Mark as accepting if needed
		if isAccepting {
			dfa.AcceptingStates[currentName] = AcceptingState{
				TokenType: tokenType,
				Priority:  maxPriority,
			}
		}
	}

	return dfa
}

// epsilonClosure computes the epsilon closure of a set of NFA states.
// This is all states reachable by following zero or more epsilon transitions.
func epsilonClosure(nfa *NFA, states map[int]bool) map[int]bool {
	closure := make(map[int]bool)
	stack := make([]int, 0, len(states))

	// Initialize with input states
	for state := range states {
		closure[state] = true
		stack = append(stack, state)
	}

	// Follow epsilon transitions
	for len(stack) > 0 {
		// Pop from stack
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Add all epsilon-reachable states
		for target := range nfa.States[current].Epsilon {
			if !closure[target] {
				closure[target] = true
				stack = append(stack, target)
			}
		}
	}

	return closure
}

// stateSetToString converts a set of NFA state IDs to a canonical string representation.
// This is used as the DFA state name.
func stateSetToString(states map[int]bool) string {
	if len(states) == 0 {
		return "∅" // Empty set
	}

	// Sort state IDs for canonical representation
	ids := make([]int, 0, len(states))
	for id := range states {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	// Build string
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// DfaWithTokens is a DFA that tracks which tokens are accepted by which states.
type DfaWithTokens struct {
	InitialState    string
	States          map[string]DfaStateWithToken
	AcceptingStates map[string]AcceptingState
}

// DfaStateWithToken is a DFA state that can have associated token information.
type DfaStateWithToken struct {
	Name              string
	Transitions       map[rune]string
	DefaultTransition string
}

// AcceptingState tracks token information for accepting states.
type AcceptingState struct {
	TokenType grammar.TokenType
	Priority  int
}

// NextState returns the next state given current state and input rune.
func (d *DfaWithTokens) NextState(currentState string, input rune) string {
	state, exists := d.States[currentState]
	if !exists {
		return ""
	}

	next, exists := state.Transitions[input]
	if exists {
		return next
	}

	return state.DefaultTransition
}

// IsAccepting returns true if the state is an accepting state.
func (d *DfaWithTokens) IsAccepting(state string) bool {
	_, ok := d.AcceptingStates[state]
	return ok
}

// GetTokenType returns the token type for an accepting state.
func (d *DfaWithTokens) GetTokenType(state string) grammar.TokenType {
	if acc, ok := d.AcceptingStates[state]; ok {
		return acc.TokenType
	}
	return ""
}
