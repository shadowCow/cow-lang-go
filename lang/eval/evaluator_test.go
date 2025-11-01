package eval

import (
	"bytes"
	"testing"

	"github.com/shadowCow/cow-lang-go/lang/ast"
)

// TestEvalIntLiteral tests evaluating integer literals.
func TestEvalIntLiteral(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{"positive integer", 42},
		{"zero", 0},
		{"large integer", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token: "42",
						Expression: &ast.IntLiteral{
							Token: "42",
							Value: tt.value,
						},
					},
				},
			}

			var output bytes.Buffer
			evaluator := NewEvaluator(&output)
			err := evaluator.Eval(program)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// TestEvalFloatLiteral tests evaluating float literals.
func TestEvalFloatLiteral(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"simple float", 3.14},
		{"zero", 0.0},
		{"scientific notation", 1.5e2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token: "3.14",
						Expression: &ast.FloatLiteral{
							Token: "3.14",
							Value: tt.value,
						},
					},
				},
			}

			var output bytes.Buffer
			evaluator := NewEvaluator(&output)
			err := evaluator.Eval(program)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// TestEvalPrintln tests the println built-in function.
func TestEvalPrintln(t *testing.T) {
	tests := []struct {
		name     string
		args     []ast.Expression
		expected string
	}{
		{
			name: "print single integer",
			args: []ast.Expression{
				&ast.IntLiteral{Token: "42", Value: 42},
			},
			expected: "42\n",
		},
		{
			name: "print single float",
			args: []ast.Expression{
				&ast.FloatLiteral{Token: "3.14", Value: 3.14},
			},
			expected: "3.14\n",
		},
		{
			name: "print multiple values",
			args: []ast.Expression{
				&ast.IntLiteral{Token: "42", Value: 42},
				&ast.FloatLiteral{Token: "3.14", Value: 3.14},
			},
			expected: "42\n3.14\n",
		},
		{
			name:     "print no values",
			args:     []ast.Expression{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token: "println",
						Expression: &ast.FunctionCall{
							Token:     "println",
							Name:      "println",
							Arguments: tt.args,
						},
					},
				},
			}

			var output bytes.Buffer
			evaluator := NewEvaluator(&output)
			err := evaluator.Eval(program)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if output.String() != tt.expected {
				t.Errorf("Expected output %q, got %q", tt.expected, output.String())
			}
		})
	}
}

// TestEvalMultipleStatements tests evaluating multiple statements.
func TestEvalMultipleStatements(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Token: "println",
				Expression: &ast.FunctionCall{
					Token: "println",
					Name:  "println",
					Arguments: []ast.Expression{
						&ast.IntLiteral{Token: "42", Value: 42},
					},
				},
			},
			&ast.ExpressionStatement{
				Token: "println",
				Expression: &ast.FunctionCall{
					Token: "println",
					Name:  "println",
					Arguments: []ast.Expression{
						&ast.FloatLiteral{Token: "3.14", Value: 3.14},
					},
				},
			},
		},
	}

	var output bytes.Buffer
	evaluator := NewEvaluator(&output)
	err := evaluator.Eval(program)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "42\n3.14\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestEvalErrors tests error handling in the evaluator.
func TestEvalErrors(t *testing.T) {
	tests := []struct {
		name    string
		program *ast.Program
	}{
		{
			name: "unknown function",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token: "unknown",
						Expression: &ast.FunctionCall{
							Token:     "unknown",
							Name:      "unknown",
							Arguments: []ast.Expression{},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			evaluator := NewEvaluator(&output)
			err := evaluator.Eval(tt.program)

			if err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}

// TestPrintlnFormatting tests that println formats numbers correctly.
func TestPrintlnFormatting(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"integer", int64(42), "42\n"},
		{"zero", int64(0), "0\n"},
		{"negative integer", int64(-42), "-42\n"},
		{"float", float64(3.14), "3.14\n"},
		{"float no decimal", float64(42.0), "42\n"}, // %g formats 42.0 as "42"
		{"scientific notation", float64(1.5e10), "1.5e+10\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expr ast.Expression

			switch v := tt.value.(type) {
			case int64:
				expr = &ast.IntLiteral{Token: "n", Value: v}
			case float64:
				expr = &ast.FloatLiteral{Token: "n", Value: v}
			}

			program := &ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token: "println",
						Expression: &ast.FunctionCall{
							Token:     "println",
							Name:      "println",
							Arguments: []ast.Expression{expr},
						},
					},
				},
			}

			var output bytes.Buffer
			evaluator := NewEvaluator(&output)
			err := evaluator.Eval(program)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if output.String() != tt.expected {
				t.Errorf("Expected output %q, got %q", tt.expected, output.String())
			}
		})
	}
}
