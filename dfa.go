package main

type Dfa struct {
	InitialState string
	States map[string]DfaState
}

type DfaState struct {
	Name string
	Transitions map[rune]string
	DefaultTransition string
}

func (d Dfa) NextState(currentState string, input rune) string {
	transition, exists := d.States[currentState].Transitions[input]

	if !exists {
		transition = d.States[currentState].DefaultTransition
	}

	return transition
}