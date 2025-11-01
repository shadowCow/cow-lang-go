package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestRun tests running a Cow program from a file.
func TestRun(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.cow")

	source := `println(42)
println(3.14)
println(0xFF)`

	err := os.WriteFile(testFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run the program
	var output bytes.Buffer
	err = Run(testFile, &output)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Check output
	expected := "42\n3.14\n255\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestRunWithExampleFile tests running the example file.
func TestRunWithExampleFile(t *testing.T) {
	// Path to the example file
	exampleFile := "../examples/hello_numbers.cow"

	// Check if file exists
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		t.Skip("Example file not found, skipping test")
	}

	// Run the program
	var output bytes.Buffer
	err := Run(exampleFile, &output)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Expected output based on hello_numbers.cow
	expected := "42\n255\n10\n3.14\n150\n1000000\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestRunFileNotFound tests error handling when file doesn't exist.
func TestRunFileNotFound(t *testing.T) {
	var output bytes.Buffer
	err := Run("/nonexistent/file.cow", &output)

	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

// TestRunLexerError tests error handling for lexer errors.
func TestRunLexerError(t *testing.T) {
	// Create a temporary test file with invalid syntax
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.cow")

	// Character that's not in our grammar
	source := "println(42) @"

	err := os.WriteFile(testFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run the program
	var output bytes.Buffer
	err = Run(testFile, &output)

	if err == nil {
		t.Fatal("Expected lexer error, got nil")
	}
}

// TestRunParserError tests error handling for parser errors.
func TestRunParserError(t *testing.T) {
	// Create a temporary test file with parser error (missing closing paren)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "parser_error.cow")

	source := "println(42"

	err := os.WriteFile(testFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run the program
	var output bytes.Buffer
	err = Run(testFile, &output)

	if err == nil {
		t.Fatal("Expected parser error, got nil")
	}
}
