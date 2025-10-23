package automata

import (
	"testing"
)

func TestTransition(t *testing.T) {
    cases := []struct {
        input rune
        expectedState string
    }{
        {'1', StateTwoName},
        {'2', StateTwoName},
        {'a', StateThreeName},
        {'b', StateThreeName},
    }

    for _, tc := range cases {
        dfa := CreateTestDfa()

        nextState := dfa.NextState(dfa.InitialState, tc.input)

        if nextState != tc.expectedState {
            t.Errorf("dfa.NextState(%c) got %s; want %s", tc.input, nextState, tc.expectedState)
        }
    }
}
