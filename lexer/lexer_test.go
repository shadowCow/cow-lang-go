package lexer

import (
	"testing"
	"github.com/shadowCow/cow-lang-go/automata"
)

func TestLexer(t *testing.T) {
	cases := []struct {
        input string
        expectedState string
    }{
        {"42", automata.StateTwoName},
        {"4get", automata.StateThreeName},
		{"forget", automata.StateThreeName},
    }

	dfa := automata.CreateTestDfa()

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