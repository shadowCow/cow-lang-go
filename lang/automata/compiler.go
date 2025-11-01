package automata

import "github.com/shadowCow/cow-lang-go/lang/grammar"

// CompilePatternToNFA converts a LexicalPattern into an NFA using Thompson's construction.
// Each pattern type is converted into a simple NFA fragment, then combined.
func CompilePatternToNFA(pattern grammar.LexicalPattern) *NFA {
	switch p := pattern.(type) {
	case grammar.Literal:
		return nfaFromLiteral(p)
	case grammar.CharRange:
		return nfaFromCharRange(p)
	case grammar.CharSet:
		return nfaFromCharSet(p)
	case grammar.AnyChar:
		return nfaFromAnyChar(p)
	case grammar.AnyCharExcept:
		return nfaFromAnyCharExcept(p)
	case grammar.LexSequence:
		return nfaFromSequence(p)
	case grammar.LexAlternative:
		return nfaFromAlternative(p)
	case grammar.LexOptional:
		return nfaFromOptional(p)
	case grammar.LexZeroOrMore:
		return nfaFromZeroOrMore(p)
	case grammar.LexOneOrMore:
		return nfaFromOneOrMore(p)
	default:
		// Should never happen if all pattern types are handled
		panic("unknown lexical pattern type")
	}
}

// nfaFromLiteral creates an NFA that matches an exact string.
// For "abc": start --a--> s1 --b--> s2 --c--> accept
func nfaFromLiteral(lit grammar.Literal) *NFA {
	str := string(lit)
	if len(str) == 0 {
		// Empty string: just epsilon transition
		nfa := NewNFA()
		nfa.AddEpsilonTransition(nfa.Start, nfa.Accept)
		return nfa
	}

	nfa := NewNFA()
	current := nfa.Start

	// Create chain of states for each character
	for i, r := range str {
		if i == len(str)-1 {
			// Last character transitions to accept state
			nfa.AddTransition(current, r, nfa.Accept)
		} else {
			// Intermediate state
			next := nfa.AddState()
			nfa.AddTransition(current, r, next)
			current = next
		}
	}

	return nfa
}

// nfaFromCharRange creates an NFA that matches any character in a range.
// start --[from-to]--> accept
func nfaFromCharRange(cr grammar.CharRange) *NFA {
	nfa := NewNFA()

	// Add transition for each character in range
	for r := cr.From; r <= cr.To; r++ {
		nfa.AddTransition(nfa.Start, r, nfa.Accept)
	}

	return nfa
}

// nfaFromCharSet creates an NFA that matches any character in a set.
func nfaFromCharSet(cs grammar.CharSet) *NFA {
	nfa := NewNFA()

	for _, r := range cs {
		nfa.AddTransition(nfa.Start, r, nfa.Accept)
	}

	return nfa
}

// nfaFromAnyChar creates an NFA that matches any single character.
// This is tricky - we'll handle it during DFA construction by making
// the transition function check all possible characters.
func nfaFromAnyChar(ac grammar.AnyChar) *NFA {
	nfa := NewNFA()

	// Add transitions for all printable ASCII and common Unicode ranges
	// In practice, this should cover a reasonable range of characters
	for r := rune(0); r <= rune(127); r++ {
		nfa.AddTransition(nfa.Start, r, nfa.Accept)
	}

	// TODO: Could extend to more Unicode ranges if needed
	return nfa
}

// nfaFromAnyCharExcept creates an NFA that matches any character except those in the set.
func nfaFromAnyCharExcept(ace grammar.AnyCharExcept) *NFA {
	nfa := NewNFA()

	// Create set for quick lookup
	excluded := make(map[rune]bool)
	for _, r := range ace {
		excluded[r] = true
	}

	// Add transitions for all characters except the excluded ones
	for r := rune(0); r <= rune(127); r++ {
		if !excluded[r] {
			nfa.AddTransition(nfa.Start, r, nfa.Accept)
		}
	}

	return nfa
}

// nfaFromSequence creates an NFA for a sequence of patterns.
// Concatenates NFAs: A.accept --ε--> B.start
func nfaFromSequence(seq grammar.LexSequence) *NFA {
	if len(seq) == 0 {
		// Empty sequence
		nfa := NewNFA()
		nfa.AddEpsilonTransition(nfa.Start, nfa.Accept)
		return nfa
	}

	// Start with first pattern
	result := CompilePatternToNFA(seq[0])

	// Concatenate remaining patterns
	for _, pattern := range seq[1:] {
		next := CompilePatternToNFA(pattern)

		// Renumber next NFA to avoid state ID conflicts
		offset := len(result.States)
		next.RenumberStates(offset)

		// Merge states
		for id, state := range next.States {
			result.States[id] = state
		}

		// Connect result.accept to next.start with epsilon
		result.AddEpsilonTransition(result.Accept, next.Start)

		// New accept state is next's accept
		result.Accept = next.Accept
	}

	return result
}

// nfaFromAlternative creates an NFA for alternative patterns.
// Creates parallel paths: start --ε--> A --ε--> accept
//                               --ε--> B --ε--> accept
func nfaFromAlternative(alt grammar.LexAlternative) *NFA {
	if len(alt) == 0 {
		// Empty alternative - should not happen
		nfa := NewNFA()
		nfa.AddEpsilonTransition(nfa.Start, nfa.Accept)
		return nfa
	}

	nfa := NewNFA()

	// Compile each alternative
	for _, pattern := range alt {
		altNFA := CompilePatternToNFA(pattern)

		// Renumber to avoid conflicts
		offset := len(nfa.States)
		altNFA.RenumberStates(offset)

		// Merge states
		for id, state := range altNFA.States {
			nfa.States[id] = state
		}

		// Connect start to alternative start
		nfa.AddEpsilonTransition(nfa.Start, altNFA.Start)

		// Connect alternative accept to accept
		nfa.AddEpsilonTransition(altNFA.Accept, nfa.Accept)
	}

	return nfa
}

// nfaFromOptional creates an NFA for optional pattern (A?).
// start --ε--> A --ε--> accept
//       --ε-----------> accept
func nfaFromOptional(opt grammar.LexOptional) *NFA {
	inner := CompilePatternToNFA(opt.Inner)

	nfa := NewNFA()

	// Renumber inner NFA
	offset := len(nfa.States)
	inner.RenumberStates(offset)

	// Merge states
	for id, state := range inner.States {
		nfa.States[id] = state
	}

	// Epsilon from start to inner start
	nfa.AddEpsilonTransition(nfa.Start, inner.Start)

	// Epsilon from inner accept to accept
	nfa.AddEpsilonTransition(inner.Accept, nfa.Accept)

	// Epsilon from start to accept (bypass)
	nfa.AddEpsilonTransition(nfa.Start, nfa.Accept)

	return nfa
}

// nfaFromZeroOrMore creates an NFA for zero-or-more pattern (A*).
// start --ε--> A --ε--> accept
//       --ε-----------> accept
//              ^--ε--┘
func nfaFromZeroOrMore(zom grammar.LexZeroOrMore) *NFA {
	inner := CompilePatternToNFA(zom.Inner)

	nfa := NewNFA()

	// Renumber inner NFA
	offset := len(nfa.States)
	inner.RenumberStates(offset)

	// Merge states
	for id, state := range inner.States {
		nfa.States[id] = state
	}

	// Epsilon from start to inner start
	nfa.AddEpsilonTransition(nfa.Start, inner.Start)

	// Epsilon from inner accept to accept
	nfa.AddEpsilonTransition(inner.Accept, nfa.Accept)

	// Epsilon from start to accept (zero iterations)
	nfa.AddEpsilonTransition(nfa.Start, nfa.Accept)

	// Epsilon from inner accept back to inner start (loop)
	nfa.AddEpsilonTransition(inner.Accept, inner.Start)

	return nfa
}

// nfaFromOneOrMore creates an NFA for one-or-more pattern (A+).
// Like A* but without the bypass epsilon from start to accept.
func nfaFromOneOrMore(oom grammar.LexOneOrMore) *NFA {
	inner := CompilePatternToNFA(oom.Inner)

	nfa := NewNFA()

	// Renumber inner NFA
	offset := len(nfa.States)
	inner.RenumberStates(offset)

	// Merge states
	for id, state := range inner.States {
		nfa.States[id] = state
	}

	// Epsilon from start to inner start
	nfa.AddEpsilonTransition(nfa.Start, inner.Start)

	// Epsilon from inner accept to accept
	nfa.AddEpsilonTransition(inner.Accept, nfa.Accept)

	// Epsilon from inner accept back to inner start (loop)
	nfa.AddEpsilonTransition(inner.Accept, inner.Start)

	// NOTE: No bypass epsilon from start to accept (requires at least one iteration)

	return nfa
}

// CompileLexicalGrammar compiles a lexical grammar into a DFA.
// All token patterns are combined into a single NFA, then converted to DFA.
func CompileLexicalGrammar(lexGrammar grammar.LexicalGrammar) DfaWithTokens {
	if len(lexGrammar.Tokens) == 0 {
		// Empty grammar
		return DfaWithTokens{
			InitialState:    "start",
			States:          make(map[string]DfaStateWithToken),
			AcceptingStates: make(map[string]AcceptingState),
		}
	}

	// Compile each token pattern to NFA
	nfas := make([]*NFA, 0, len(lexGrammar.Tokens))

	for _, tokenDef := range lexGrammar.Tokens {
		nfa := CompilePatternToNFA(tokenDef.Pattern)
		// Mark the accept state with token information
		nfa.AcceptStates[nfa.Accept] = AcceptInfo{
			TokenType: tokenDef.Name,
			Priority:  tokenDef.Priority,
		}
		nfas = append(nfas, nfa)
	}

	// Combine all NFAs with alternation
	// Create a new start state with epsilon transitions to each NFA
	combined := combineNFAs(nfas)

	// Convert combined NFA to DFA
	dfa := NFAToDFAWithTokens(combined)

	return dfa
}

// combineNFAs combines multiple NFAs into a single NFA using alternation.
// Creates new start and accept states, with epsilon transitions to/from each NFA.
func combineNFAs(nfas []*NFA) *NFA {
	if len(nfas) == 0 {
		return NewNFA()
	}

	// Create result NFA with new start state
	result := NewNFA()
	offset := len(result.States)

	// We'll need to track multiple accept states (one per token)
	// and their priorities. The DFA construction will handle choosing
	// the highest priority token when multiple match.

	for _, nfa := range nfas {
		// Copy and renumber this NFA
		nfaCopy := nfa.Copy()
		nfaCopy.RenumberStates(offset)

		// Merge states
		for id, state := range nfaCopy.States {
			result.States[id] = state
		}

		// Merge accept states with their token info
		for id, acceptInfo := range nfaCopy.AcceptStates {
			result.AcceptStates[id] = acceptInfo
		}

		// Add epsilon from result start to this NFA's start
		result.AddEpsilonTransition(result.Start, nfaCopy.Start)

		// NOTE: We keep each NFA's accept state separate
		// The DFA construction will create DFA states that may include
		// multiple NFA accept states, and will choose the highest priority

		offset = len(result.States)
	}

	return result
}

// CompileTokenDefinition is a convenience function to compile a single token definition.
func CompileTokenDefinition(tokenDef grammar.TokenDefinition) *NFA {
	nfa := CompilePatternToNFA(tokenDef.Pattern)
	nfa.AcceptStates[nfa.Accept] = AcceptInfo{
		TokenType: tokenDef.Name,
		Priority:  tokenDef.Priority,
	}
	return nfa
}
