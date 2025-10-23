# Automata Package

This package contains data structures and algorithms for various types of automata used in language processing.

## Current Implementation
- **Deterministic Finite Automaton (DFA)** - Basic state machine for pattern recognition

## Usage
```go
import "github.com/shadowCow/cow-lang-go/automata"

dfa := automata.CreateTestDfa()
nextState := dfa.NextState(currentState, inputRune)
```