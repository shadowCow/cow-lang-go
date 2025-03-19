package main

import (
	"testing"
)

func TestTransition(t *testing.T) {
    cases := []struct {
        input rune
        expectedState string
    }{
        {'1', stateTwoName},
        {'2', stateTwoName},
        {'a', stateThreeName},
        {'b', stateThreeName},
    }

    for _, tc := range cases {
        dfa := createTestDfa()

        nextState := dfa.NextState(tc.input)

        if nextState != tc.expectedState {
            t.Errorf("dfa.NextState(%c) got %s; want %s", tc.input, nextState, tc.expectedState)
        }
    }
}


const (
    stateOneName = "start"
    stateTwoName = "numeric"
    stateThreeName = "wordic"
)

func createTestDfa() Dfa {

    stateOne := DfaState{
        Name: stateOneName,
        Transitions: map[rune]string{
            '0': stateTwoName,
            '1': stateTwoName,
            '2': stateTwoName,
            '3': stateTwoName,
            '4': stateTwoName,
            '5': stateTwoName,
            '6': stateTwoName,
            '7': stateTwoName,
            '8': stateTwoName,
            '9': stateTwoName,
        },
        DefaultTransition: stateThreeName,
    }
    stateTwo := DfaState{
        Name: stateTwoName,
        Transitions: map[rune]string{
            '0': stateTwoName,
            '1': stateTwoName,
            '2': stateTwoName,
            '3': stateTwoName,
            '4': stateTwoName,
            '5': stateTwoName,
            '6': stateTwoName,
            '7': stateTwoName,
            '8': stateTwoName,
            '9': stateTwoName,
        },
        DefaultTransition: stateThreeName,
    }
    stateThree := DfaState{
        Name: stateThreeName,
        Transitions: map[rune]string{},
        DefaultTransition: stateThreeName,
    }
    states := map[string]DfaState{}
    states[stateOne.Name] = stateOne;
    states[stateTwo.Name] = stateTwo;
    states[stateThree.Name] = stateThree;

    testDfa := Dfa{
        State: stateOne.Name,
        States: states,
    }

    return testDfa;
}