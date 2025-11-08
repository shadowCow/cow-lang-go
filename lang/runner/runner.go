// Package runner provides a simple API to execute Cow language programs from files.
package runner

import (
	"fmt"
	"io"
	"os"

	"github.com/shadowCow/cow-lang-go/lang/automata"
	"github.com/shadowCow/cow-lang-go/lang/eval"
	"github.com/shadowCow/cow-lang-go/lang/langdef"
	"github.com/shadowCow/cow-lang-go/lang/lexer"
	"github.com/shadowCow/cow-lang-go/lang/ll1"
)

// Run executes a Cow language program from a file.
// It performs the complete pipeline: read file → lex → parse → evaluate.
// Output from the program (e.g., println statements) is written to the provided io.Writer.
//
// If debug is true, prints grammar information, FIRST/FOLLOW sets, parse table, and parse trace.
//
// Returns an error if any stage fails (file reading, lexing, parsing, or evaluation).
func Run(filePath string, output io.Writer, debug bool) error {
	// Read the source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	// Get the syntactic grammar
	synGrammar := langdef.GetSyntacticGrammar()

	// Debug: Print grammar
	if debug {
		ll1.PrintGrammar(synGrammar)
	}

	// Compute FIRST sets
	firstSets := ll1.ComputeFirstSets(synGrammar)
	if debug {
		ll1.PrintFirstSets(firstSets)
	}

	// Compute FOLLOW sets
	followSets := ll1.ComputeFollowSets(synGrammar, firstSets)
	if debug {
		ll1.PrintFollowSets(followSets)
	}

	// Build LL(1) parse table
	parseTable, err := ll1.BuildParseTable(synGrammar, firstSets, followSets)
	if err != nil {
		return fmt.Errorf("failed to build LL(1) parse table: %w", err)
	}
	if debug {
		ll1.PrintParseTable(parseTable)
	}

	// Compile the lexical grammar to a DFA
	lexGrammar := langdef.GetLexical()
	dfa := automata.CompileLexicalGrammar(lexGrammar)

	// Tokenize the source code
	lex := lexer.NewLexer(dfa, string(source))
	tokens, err := lex.Tokenize()
	if err != nil {
		return fmt.Errorf("lexer error in %q: %w", filePath, err)
	}

	// Parse tokens into an AST using LL(1) parser
	p := ll1.NewParser(parseTable, synGrammar, tokens)
	if debug {
		p.SetTrace(true) // Enable parse tracing in debug mode
	}
	program, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parser error in %q: %w", filePath, err)
	}

	// Evaluate the program
	evaluator := eval.NewEvaluator(output)
	err = evaluator.Eval(program)
	if err != nil {
		return fmt.Errorf("evaluation error in %q: %w", filePath, err)
	}

	return nil
}
