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

	// Simple let statement (current grammar requires statements)
	source := `let _ = 42`

	err := os.WriteFile(testFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run the program
	var output bytes.Buffer
	err = Run(testFile, &output, false)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Let statements don't produce output
	// They just assign values to variables
	expected := ""
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestRunWithExampleFile tests running a simple example file.
func TestRunWithExampleFile(t *testing.T) {
	// Path to the simple example file (just a let statement)
	exampleFile := "../examples/simple_literal.cow"

	// Check if file exists
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		t.Skip("Example file not found, skipping test")
	}

	// Run the program
	var output bytes.Buffer
	err := Run(exampleFile, &output, false)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Let statements don't produce output
	expected := ""
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestRunFileNotFound tests error handling when file doesn't exist.
func TestRunFileNotFound(t *testing.T) {
	var output bytes.Buffer
	err := Run("/nonexistent/file.cow", &output, false)

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
	source := "42 @"

	err := os.WriteFile(testFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run the program
	var output bytes.Buffer
	err = Run(testFile, &output, false)

	if err == nil {
		t.Fatal("Expected lexer error, got nil")
	}
}

// TestRunParserError tests error handling for parser errors.
func TestRunParserError(t *testing.T) {
	// Create a temporary test file with parser error
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "parser_error.cow")

	// Standalone identifier is not valid (only let statements or function defs)
	source := "some_identifier"

	err := os.WriteFile(testFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run the program
	var output bytes.Buffer
	err = Run(testFile, &output, false)

	if err == nil {
		t.Fatal("Expected parser error, got nil")
	}
}
