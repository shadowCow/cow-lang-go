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
	"github.com/shadowCow/cow-lang-go/lang/parser"
)

// Run executes a Cow language program from a file.
// It performs the complete pipeline: read file → lex → parse → evaluate.
// Output from the program (e.g., println statements) is written to the provided io.Writer.
//
// Returns an error if any stage fails (file reading, lexing, parsing, or evaluation).
func Run(filePath string, output io.Writer) error {
	// Read the source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
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

	// Parse tokens into an AST
	p := parser.NewParser(tokens)
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
