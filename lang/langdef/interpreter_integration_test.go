package langdef

import (
	"bytes"
	"testing"

	"github.com/shadowCow/cow-lang-go/lang/automata"
	"github.com/shadowCow/cow-lang-go/lang/eval"
	"github.com/shadowCow/cow-lang-go/lang/lexer"
	"github.com/shadowCow/cow-lang-go/lang/parser"
)

// TestPrintlnNumbers tests the complete pipeline: lex → parse → eval
// for simple programs that print number literals.
func TestPrintlnNumbers(t *testing.T) {
	// Compile the lexical grammar
	lexGrammar := GetLexical()
	dfa := automata.CompileLexicalGrammar(lexGrammar)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "print single decimal integer",
			input:    "println(42)",
			expected: "42\n",
		},
		{
			name:     "print hex integer",
			input:    "println(0xFF)",
			expected: "255\n",
		},
		{
			name:     "print binary integer",
			input:    "println(0b1010)",
			expected: "10\n",
		},
		{
			name:     "print float",
			input:    "println(3.14)",
			expected: "3.14\n",
		},
		{
			name:     "print scientific notation",
			input:    "println(1.5e2)",
			expected: "150\n",
		},
		{
			name:     "print multiple numbers",
			input:    "println(42, 3.14)",
			expected: "42\n3.14\n",
		},
		{
			name:     "print with underscores",
			input:    "println(1_000_000)",
			expected: "1000000\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Lex
			lex := lexer.NewLexer(dfa, tt.input)
			tokens, err := lex.Tokenize()
			if err != nil {
				t.Fatalf("Lexer error: %v", err)
			}

			// Parse
			p := parser.NewParser(tokens)
			program, err := p.Parse()
			if err != nil {
				t.Fatalf("Parser error: %v", err)
			}

			// Eval
			var output bytes.Buffer
			evaluator := eval.NewEvaluator(&output)
			err = evaluator.Eval(program)
			if err != nil {
				t.Fatalf("Evaluator error: %v", err)
			}

			// Check output
			if output.String() != tt.expected {
				t.Errorf("Expected output %q, got %q", tt.expected, output.String())
			}
		})
	}
}

// TestMultipleStatements tests programs with multiple println statements.
func TestMultipleStatements(t *testing.T) {
	lexGrammar := GetLexical()
	dfa := automata.CompileLexicalGrammar(lexGrammar)

	input := `println(42)
println(3.14)
println(0xFF)`

	expected := "42\n3.14\n255\n"

	// Lex
	lex := lexer.NewLexer(dfa, input)
	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	// Parse
	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}

	// Eval
	var output bytes.Buffer
	evaluator := eval.NewEvaluator(&output)
	err = evaluator.Eval(program)
	if err != nil {
		t.Fatalf("Evaluator error: %v", err)
	}

	// Check output
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}
