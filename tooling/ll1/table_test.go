package ll1

import (
	"strings"
	"testing"

	"github.com/shadowCow/cow-lang-go/tooling/grammar"
)

// TestLL1ConflictDetection verifies that the parser generator detects LL(1) conflicts.
func TestLL1ConflictDetection(t *testing.T) {
	// Create a non-LL(1) grammar with a conflict
	// S -> A | B
	// A -> a x
	// B -> a y
	// This has a conflict because both A and B start with 'a'
	conflictingGrammar := grammar.SyntacticGrammar{
		StartSymbol: "S",
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			"S": grammar.SynAlternative{
				grammar.NonTerminal{Symbol: "A"},
				grammar.NonTerminal{Symbol: "B"},
			},
			"A": grammar.SynSequence{
				grammar.Terminal{TokenType: "a"},
				grammar.Terminal{TokenType: "x"},
			},
			"B": grammar.SynSequence{
				grammar.Terminal{TokenType: "a"},
				grammar.Terminal{TokenType: "y"},
			},
		},
	}

	// Compute FIRST and FOLLOW sets
	firstSets := ComputeFirstSets(conflictingGrammar)
	followSets := ComputeFollowSets(conflictingGrammar, firstSets)

	// Try to build parse table - should fail with conflict
	_, err := BuildParseTable(conflictingGrammar, firstSets, followSets)

	if err == nil {
		t.Fatal("Expected LL(1) conflict error, but got none")
	}

	// Check that the error is a GrammarNotLL1Error
	if _, ok := err.(*GrammarNotLL1Error); !ok {
		t.Fatalf("Expected GrammarNotLL1Error, got %T: %v", err, err)
	}

	// Verify the error message mentions the conflict
	errMsg := err.Error()
	if !strings.Contains(errMsg, "conflict") {
		t.Errorf("Error message should mention 'conflict', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "[S, a]") {
		t.Errorf("Error message should mention the conflicting cell [S, a], got: %s", errMsg)
	}
}

// TestLL1ValidGrammar verifies that a valid LL(1) grammar is accepted.
func TestLL1ValidGrammar(t *testing.T) {
	// Create a simple LL(1) grammar
	// S -> a | b
	validGrammar := grammar.SyntacticGrammar{
		StartSymbol: "S",
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			"S": grammar.SynAlternative{
				grammar.Terminal{TokenType: "a"},
				grammar.Terminal{TokenType: "b"},
			},
		},
	}

	// Compute FIRST and FOLLOW sets
	firstSets := ComputeFirstSets(validGrammar)
	followSets := ComputeFollowSets(validGrammar, firstSets)

	// Build parse table - should succeed
	table, err := BuildParseTable(validGrammar, firstSets, followSets)

	if err != nil {
		t.Fatalf("Expected no error for valid LL(1) grammar, got: %v", err)
	}

	if table == nil {
		t.Fatal("Expected parse table, got nil")
	}

	// Verify table entries
	if table.Get("S", "a") == nil {
		t.Error("Expected production for [S, a]")
	}
	if table.Get("S", "b") == nil {
		t.Error("Expected production for [S, b]")
	}
}

// TestFirstSetsComputation verifies FIRST set computation.
func TestFirstSetsComputation(t *testing.T) {
	// Grammar:
	// S -> A B
	// A -> a | ε
	// B -> b
	g := grammar.SyntacticGrammar{
		StartSymbol: "S",
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			"S": grammar.SynSequence{
				grammar.NonTerminal{Symbol: "A"},
				grammar.NonTerminal{Symbol: "B"},
			},
			"A": grammar.SynAlternative{
				grammar.Terminal{TokenType: "a"},
				grammar.SynOptional{Inner: grammar.Terminal{TokenType: "a"}}, // Represents epsilon
			},
			"B": grammar.Terminal{TokenType: "b"},
		},
	}

	firstSets := ComputeFirstSets(g)

	// Check FIRST(B) = {b}
	firstB := firstSets.Get("B")
	if !firstB["b"] {
		t.Error("FIRST(B) should contain 'b'")
	}
	if len(firstB) != 1 {
		t.Errorf("FIRST(B) should have exactly 1 element, got %d", len(firstB))
	}

	// Check FIRST(A) = {a}
	firstA := firstSets.Get("A")
	if !firstA["a"] {
		t.Error("FIRST(A) should contain 'a'")
	}

	// Check FIRST(S) = {a, b} (because A can be epsilon)
	firstS := firstSets.Get("S")
	if !firstS["a"] {
		t.Error("FIRST(S) should contain 'a'")
	}
	if !firstS["b"] {
		t.Error("FIRST(S) should contain 'b' (because A is nullable)")
	}
}

// TestFollowSetsComputation verifies FOLLOW set computation.
func TestFollowSetsComputation(t *testing.T) {
	// Grammar:
	// S -> A B
	// A -> a
	// B -> b
	g := grammar.SyntacticGrammar{
		StartSymbol: "S",
		Productions: map[grammar.Symbol]grammar.ProductionRule{
			"S": grammar.SynSequence{
				grammar.NonTerminal{Symbol: "A"},
				grammar.NonTerminal{Symbol: "B"},
			},
			"A": grammar.Terminal{TokenType: "a"},
			"B": grammar.Terminal{TokenType: "b"},
		},
	}

	firstSets := ComputeFirstSets(g)
	followSets := ComputeFollowSets(g, firstSets)

	// Check FOLLOW(S) = {$}
	followS := followSets.Get("S")
	if !followS["$"] {
		t.Error("FOLLOW(S) should contain '$' (end of input)")
	}

	// Check FOLLOW(A) = {b} (A is followed by B in the production S -> A B)
	followA := followSets.Get("A")
	if !followA["b"] {
		t.Error("FOLLOW(A) should contain 'b'")
	}

	// Check FOLLOW(B) = {$} (B is at the end of S -> A B)
	followB := followSets.Get("B")
	if !followB["$"] {
		t.Error("FOLLOW(B) should contain '$'")
	}
}
