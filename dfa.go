package main

type Dfa struct {
	State string
	States map[string]DfaState
}

type DfaState struct {
	Name string
	Transitions map[rune]string
	DefaultTransition string
}

func (d Dfa) NextState(input rune) string {
	transition, exists := d.States[d.State].Transitions[input]

	if !exists {
		transition = d.States[d.State].DefaultTransition
	}

	return transition
}