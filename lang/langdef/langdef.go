// Package langdef defines the grammar for the Cow programming language.
//
// The Cow language is designed for distributed embedded stream processors,
// combining type safety, functional programming, event-driven architecture,
// and explicit resource bounds.
//
// This package uses the grammar framework (github.com/shadowCow/cow-lang-go/tooling/grammar)
// to define both lexical and syntactic rules for the language.
//
// Current Status: Phase 1 Planning
// The scaffolding is in place, but the actual grammar rules have not yet been defined.
// See lang/design/ directory for language design documentation.
package langdef

import "github.com/shadowCow/cow-lang-go/tooling/grammar"

// Grammar represents the complete grammar definition for the Cow language,
// including both lexical (tokenization) and syntactic (parsing) rules.
type Grammar struct {
	Lexical   grammar.LexicalGrammar
	Syntactic grammar.SyntacticGrammar
}

// GetGrammar returns the complete grammar definition for the Cow language.
//
// This is the main entry point for accessing the language grammar.
// Parser implementations should use this function to obtain the rules
// needed to parse Cow source code.
//
// TODO: Once the grammar is defined, this will return a complete,
// working grammar for Phase 1 of the language.
func GetGrammar() Grammar {
	return Grammar{
		Lexical:   GetLexicalGrammar(),
		Syntactic: GetSyntacticGrammar(),
	}
}

// GetLexical is a convenience function that returns just the lexical grammar.
func GetLexical() grammar.LexicalGrammar {
	return GetLexicalGrammar()
}

// GetSyntactic is a convenience function that returns just the syntactic grammar.
func GetSyntactic() grammar.SyntacticGrammar {
	return GetSyntacticGrammar()
}
