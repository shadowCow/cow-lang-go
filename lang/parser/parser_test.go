package parser

import (
	"testing"

	"github.com/shadowCow/cow-lang-go/lang/ast"
	"github.com/shadowCow/cow-lang-go/lang/lexer"
)

// TestParseIntLiteral tests parsing integer literals.
func TestParseIntLiteral(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []lexer.Token
		expected int64
	}{
		{
			name: "decimal integer",
			tokens: []lexer.Token{
				{Type: "INT_DECIMAL", Value: "42", Line: 1, Column: 1},
			},
			expected: 42,
		},
		{
			name: "hex integer",
			tokens: []lexer.Token{
				{Type: "INT_HEX", Value: "0xFF", Line: 1, Column: 1},
			},
			expected: 255,
		},
		{
			name: "binary integer",
			tokens: []lexer.Token{
				{Type: "INT_BINARY", Value: "0b1010", Line: 1, Column: 1},
			},
			expected: 10,
		},
		{
			name: "integer with underscores",
			tokens: []lexer.Token{
				{Type: "INT_DECIMAL", Value: "1_000_000", Line: 1, Column: 1},
			},
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.tokens)
			program, err := p.Parse()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
			}

			intLit, ok := stmt.Expression.(*ast.IntLiteral)
			if !ok {
				t.Fatalf("Expected IntLiteral, got %T", stmt.Expression)
			}

			if intLit.Value != tt.expected {
				t.Errorf("Expected value %d, got %d", tt.expected, intLit.Value)
			}
		})
	}
}

// TestParseFloatLiteral tests parsing float literals.
func TestParseFloatLiteral(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []lexer.Token
		expected float64
	}{
		{
			name: "simple float",
			tokens: []lexer.Token{
				{Type: "FLOAT", Value: "3.14", Line: 1, Column: 1},
			},
			expected: 3.14,
		},
		{
			name: "scientific notation",
			tokens: []lexer.Token{
				{Type: "FLOAT", Value: "1.5e2", Line: 1, Column: 1},
			},
			expected: 150.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.tokens)
			program, err := p.Parse()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
			}

			floatLit, ok := stmt.Expression.(*ast.FloatLiteral)
			if !ok {
				t.Fatalf("Expected FloatLiteral, got %T", stmt.Expression)
			}

			if floatLit.Value != tt.expected {
				t.Errorf("Expected value %f, got %f", tt.expected, floatLit.Value)
			}
		})
	}
}

// TestParseFunctionCall tests parsing function calls.
func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []lexer.Token
		funcName string
		numArgs  int
	}{
		{
			name: "println with one argument",
			tokens: []lexer.Token{
				{Type: "IDENTIFIER", Value: "println", Line: 1, Column: 1},
				{Type: "LPAREN", Value: "(", Line: 1, Column: 8},
				{Type: "INT_DECIMAL", Value: "42", Line: 1, Column: 9},
				{Type: "RPAREN", Value: ")", Line: 1, Column: 11},
			},
			funcName: "println",
			numArgs:  1,
		},
		{
			name: "println with multiple arguments",
			tokens: []lexer.Token{
				{Type: "IDENTIFIER", Value: "println", Line: 1, Column: 1},
				{Type: "LPAREN", Value: "(", Line: 1, Column: 8},
				{Type: "INT_DECIMAL", Value: "42", Line: 1, Column: 9},
				{Type: "COMMA", Value: ",", Line: 1, Column: 11},
				{Type: "FLOAT", Value: "3.14", Line: 1, Column: 13},
				{Type: "RPAREN", Value: ")", Line: 1, Column: 17},
			},
			funcName: "println",
			numArgs:  2,
		},
		{
			name: "println with no arguments",
			tokens: []lexer.Token{
				{Type: "IDENTIFIER", Value: "println", Line: 1, Column: 1},
				{Type: "LPAREN", Value: "(", Line: 1, Column: 8},
				{Type: "RPAREN", Value: ")", Line: 1, Column: 9},
			},
			funcName: "println",
			numArgs:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.tokens)
			program, err := p.Parse()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
			}

			call, ok := stmt.Expression.(*ast.FunctionCall)
			if !ok {
				t.Fatalf("Expected FunctionCall, got %T", stmt.Expression)
			}

			if call.Name != tt.funcName {
				t.Errorf("Expected function name %q, got %q", tt.funcName, call.Name)
			}

			if len(call.Arguments) != tt.numArgs {
				t.Errorf("Expected %d arguments, got %d", tt.numArgs, len(call.Arguments))
			}
		})
	}
}

// TestParseMultipleStatements tests parsing multiple statements.
func TestParseMultipleStatements(t *testing.T) {
	tokens := []lexer.Token{
		{Type: "IDENTIFIER", Value: "println", Line: 1, Column: 1},
		{Type: "LPAREN", Value: "(", Line: 1, Column: 8},
		{Type: "INT_DECIMAL", Value: "42", Line: 1, Column: 9},
		{Type: "RPAREN", Value: ")", Line: 1, Column: 11},
		{Type: "IDENTIFIER", Value: "println", Line: 2, Column: 1},
		{Type: "LPAREN", Value: "(", Line: 2, Column: 8},
		{Type: "FLOAT", Value: "3.14", Line: 2, Column: 9},
		{Type: "RPAREN", Value: ")", Line: 2, Column: 13},
	}

	p := NewParser(tokens)
	program, err := p.Parse()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	// Check first statement
	stmt1, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected first statement to be ExpressionStatement, got %T", program.Statements[0])
	}

	call1, ok := stmt1.Expression.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("Expected first expression to be FunctionCall, got %T", stmt1.Expression)
	}

	if call1.Name != "println" {
		t.Errorf("Expected function name 'println', got %q", call1.Name)
	}

	// Check second statement
	stmt2, ok := program.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected second statement to be ExpressionStatement, got %T", program.Statements[1])
	}

	call2, ok := stmt2.Expression.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("Expected second expression to be FunctionCall, got %T", stmt2.Expression)
	}

	if call2.Name != "println" {
		t.Errorf("Expected function name 'println', got %q", call2.Name)
	}
}

// TestParserErrors tests error handling in the parser.
func TestParserErrors(t *testing.T) {
	tests := []struct {
		name   string
		tokens []lexer.Token
	}{
		{
			name: "missing closing paren",
			tokens: []lexer.Token{
				{Type: "IDENTIFIER", Value: "println", Line: 1, Column: 1},
				{Type: "LPAREN", Value: "(", Line: 1, Column: 8},
				{Type: "INT_DECIMAL", Value: "42", Line: 1, Column: 9},
			},
		},
		{
			name: "missing opening paren",
			tokens: []lexer.Token{
				{Type: "IDENTIFIER", Value: "println", Line: 1, Column: 1},
				{Type: "INT_DECIMAL", Value: "42", Line: 1, Column: 9},
				{Type: "RPAREN", Value: ")", Line: 1, Column: 11},
			},
		},
		{
			name: "unexpected token",
			tokens: []lexer.Token{
				{Type: "LPAREN", Value: "(", Line: 1, Column: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.tokens)
			_, err := p.Parse()

			if err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}
