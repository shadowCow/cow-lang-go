package automata

import "github.com/shadowCow/cow-lang-go/tooling/grammar"

// NFA represents a non-deterministic finite automaton.
// NFAs can have multiple transitions for the same input symbol,
// and can have epsilon (ε) transitions that don't consume input.
type NFA struct {
	Start  int
	Accept int
	States map[int]*NFAState
	// AcceptStates maps state IDs to their token information
	AcceptStates map[int]AcceptInfo
}

// NFAState represents a state in an NFA.
type NFAState struct {
	ID int
	// Transitions maps input runes to sets of next states
	Transitions map[rune]map[int]bool
	// Epsilon transitions don't consume input
	Epsilon map[int]bool
}

// AcceptInfo stores token information for an accepting state.
type AcceptInfo struct {
	TokenType grammar.TokenType
	Priority  int
}

// NewNFA creates a new NFA with a start and accept state.
func NewNFA() *NFA {
	nfa := &NFA{
		Start:        0,
		Accept:       1,
		States:       make(map[int]*NFAState),
		AcceptStates: make(map[int]AcceptInfo),
	}
	nfa.States[0] = &NFAState{
		ID:          0,
		Transitions: make(map[rune]map[int]bool),
		Epsilon:     make(map[int]bool),
	}
	nfa.States[1] = &NFAState{
		ID:          1,
		Transitions: make(map[rune]map[int]bool),
		Epsilon:     make(map[int]bool),
	}
	return nfa
}

// AddState adds a new state to the NFA and returns its ID.
func (nfa *NFA) AddState() int {
	id := len(nfa.States)
	nfa.States[id] = &NFAState{
		ID:          id,
		Transitions: make(map[rune]map[int]bool),
		Epsilon:     make(map[int]bool),
	}
	return id
}

// AddTransition adds a transition from one state to another on input rune.
func (nfa *NFA) AddTransition(from int, input rune, to int) {
	if nfa.States[from].Transitions[input] == nil {
		nfa.States[from].Transitions[input] = make(map[int]bool)
	}
	nfa.States[from].Transitions[input][to] = true
}

// AddEpsilonTransition adds an epsilon transition from one state to another.
func (nfa *NFA) AddEpsilonTransition(from, to int) {
	nfa.States[from].Epsilon[to] = true
}

// RenumberStates renumbers all states starting from offset.
// Returns the new start and accept state IDs.
func (nfa *NFA) RenumberStates(offset int) (newStart, newAccept int) {
	newStates := make(map[int]*NFAState)
	mapping := make(map[int]int)

	// Create mapping from old IDs to new IDs
	for oldID := range nfa.States {
		mapping[oldID] = oldID + offset
	}

	// Create new states with updated IDs
	for oldID, state := range nfa.States {
		newID := mapping[oldID]
		newState := &NFAState{
			ID:          newID,
			Transitions: make(map[rune]map[int]bool),
			Epsilon:     make(map[int]bool),
		}

		// Update transitions
		for r, targets := range state.Transitions {
			newState.Transitions[r] = make(map[int]bool)
			for targetID := range targets {
				newState.Transitions[r][mapping[targetID]] = true
			}
		}

		// Update epsilon transitions
		for targetID := range state.Epsilon {
			newState.Epsilon[mapping[targetID]] = true
		}

		newStates[newID] = newState
	}

	nfa.States = newStates
	nfa.Start = mapping[nfa.Start]
	nfa.Accept = mapping[nfa.Accept]

	// Update accept states
	newAcceptStates := make(map[int]AcceptInfo)
	for oldID, info := range nfa.AcceptStates {
		newAcceptStates[mapping[oldID]] = info
	}
	nfa.AcceptStates = newAcceptStates

	return nfa.Start, nfa.Accept
}

// Copy creates a deep copy of the NFA.
func (nfa *NFA) Copy() *NFA {
	copy := &NFA{
		Start:        nfa.Start,
		Accept:       nfa.Accept,
		States:       make(map[int]*NFAState),
		AcceptStates: make(map[int]AcceptInfo),
	}

	for id, state := range nfa.States {
		newState := &NFAState{
			ID:          state.ID,
			Transitions: make(map[rune]map[int]bool),
			Epsilon:     make(map[int]bool),
		}

		// Copy transitions
		for r, targets := range state.Transitions {
			newState.Transitions[r] = make(map[int]bool)
			for target := range targets {
				newState.Transitions[r][target] = true
			}
		}

		// Copy epsilon transitions
		for target := range state.Epsilon {
			newState.Epsilon[target] = true
		}

		copy.States[id] = newState
	}

	// Copy accept states
	for id, info := range nfa.AcceptStates {
		copy.AcceptStates[id] = info
	}

	return copy
}
