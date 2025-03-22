package main

type Lexer struct {
	dfa Dfa
	state string
}

func (l *Lexer) Lex(input string) string {
	for _, r := range input {
        l.state = l.dfa.NextState(l.state, r)
    }

	return l.state
}