package automata

import (
	"testing"

	"github.com/shadowCow/cow-lang-go/tooling/grammar"
)

// TestCompileLiteral tests compiling a simple literal pattern to NFA.
func TestCompileLiteral(t *testing.T) {
	pattern := grammar.Literal("abc")
	nfa := CompilePatternToNFA(pattern)

	if nfa == nil {
		t.Fatal("NFA should not be nil")
	}

	if nfa.Start == nfa.Accept {
		t.Error("Start and accept states should be different")
	}

	if len(nfa.States) == 0 {
		t.Error("NFA should have states")
	}
}

// TestCompileCharRange tests compiling a character range.
func TestCompileCharRange(t *testing.T) {
	pattern := grammar.CharRange{From: 'a', To: 'z'}
	nfa := CompilePatternToNFA(pattern)

	if nfa == nil {
		t.Fatal("NFA should not be nil")
	}

	// Should have transitions for each character in range
	startState := nfa.States[nfa.Start]
	if len(startState.Transitions) == 0 {
		t.Error("Should have transitions from start state")
	}
}

// TestCompileSequence tests compiling a sequence of patterns.
func TestCompileSequence(t *testing.T) {
	pattern := grammar.LexSequence{
		grammar.Literal("a"),
		grammar.Literal("b"),
		grammar.Literal("c"),
	}
	nfa := CompilePatternToNFA(pattern)

	if nfa == nil {
		t.Fatal("NFA should not be nil")
	}

	// Should have multiple states for the sequence
	if len(nfa.States) < 4 { // At least start, intermediate states, and accept
		t.Errorf("Expected at least 4 states, got %d", len(nfa.States))
	}
}

// TestCompileAlternative tests compiling alternative patterns.
func TestCompileAlternative(t *testing.T) {
	pattern := grammar.LexAlternative{
		grammar.Literal("a"),
		grammar.Literal("b"),
	}
	nfa := CompilePatternToNFA(pattern)

	if nfa == nil {
		t.Fatal("NFA should not be nil")
	}

	// Should have epsilon transitions from start to alternatives
	startState := nfa.States[nfa.Start]
	if len(startState.Epsilon) < 2 {
		t.Errorf("Expected at least 2 epsilon transitions from start, got %d", len(startState.Epsilon))
	}
}

// TestNFAToDFA tests converting an NFA to DFA.
func TestNFAToDFA(t *testing.T) {
	// Create simple NFA for literal "ab"
	pattern := grammar.Literal("ab")
	nfa := CompilePatternToNFA(pattern)

	// Mark as accepting with token info
	nfa.AcceptStates[nfa.Accept] = AcceptInfo{
		TokenType: "TEST_TOKEN",
		Priority:  1,
	}

	// Convert to DFA
	dfa := NFAToDFAWithTokens(nfa)

	if dfa.InitialState == "" {
		t.Error("DFA should have an initial state")
	}

	if len(dfa.States) == 0 {
		t.Error("DFA should have states")
	}

	if len(dfa.AcceptingStates) == 0 {
		t.Error("DFA should have at least one accepting state")
	}
}

// TestCompileLexicalGrammar tests compiling a complete lexical grammar.
func TestCompileLexicalGrammar(t *testing.T) {
	lexGrammar := grammar.LexicalGrammar{
		Tokens: []grammar.TokenDefinition{
			{
				Name:     "NUMBER",
				Pattern:  grammar.LexOneOrMore{Inner: grammar.CharRange{From: '0', To: '9'}},
				Priority: 1,
			},
			{
				Name:     "PLUS",
				Pattern:  grammar.Literal("+"),
				Priority: 2,
			},
		},
	}

	dfa := CompileLexicalGrammar(lexGrammar)

	if dfa.InitialState == "" {
		t.Error("Compiled DFA should have an initial state")
	}

	if len(dfa.States) == 0 {
		t.Error("Compiled DFA should have states")
	}

	if len(dfa.AcceptingStates) == 0 {
		t.Error("Compiled DFA should have accepting states for tokens")
	}
}
