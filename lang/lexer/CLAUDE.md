# Lexer Package

This package contains data structures and algorithms for lexical analysis - transforming character streams into tokens.

## Current Implementation
- **Lexer** - Basic lexer that uses a DFA to process input strings and determine final states

## Usage
```go
import (
    "github.com/shadowCow/cow-lang-go/lang/lexer"
    "github.com/shadowCow/cow-lang-go/lang/automata"
)

dfa := automata.CreateTestDfa()
lexer := lexer.Lexer{
    dfa: dfa,
    state: dfa.InitialState,
}
finalState := lexer.Lex("input")
```