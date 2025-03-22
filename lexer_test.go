package main

import (
	"testing"
)

func TestLexer(t *testing.T) {
	cases := []struct {
        input string
        expectedState string
    }{
        {"42", StateTwoName},
        {"4get", StateThreeName},
		{"forget", StateThreeName},
    }

	dfa := createTestDfa()

    for _, tc := range cases {	
        lexer := Lexer{
			dfa: dfa,
			state: dfa.InitialState, 
		}

        finalState := lexer.Lex(tc.input)

        if finalState != tc.expectedState {
            t.Errorf("lexer.Lex(%s) got %s; want %s", tc.input, finalState, tc.expectedState)
        }
    }
}