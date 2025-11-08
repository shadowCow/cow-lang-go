package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLIWithExamples(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "hello_numbers",
			args:     []string{"cow-lang", "../../examples/hello_numbers.cow"},
			expected: "42\n255\n10\n3.14\n150\n1000000\n",
		},
		{
			name:     "hello_println",
			args:     []string{"cow-lang", "../../examples/hello_println.cow"},
			expected: "42\n",
		},
		{
			name:     "variables",
			args:     []string{"cow-lang", "../../examples/variables.cow"},
			expected: "42\n255\n3.14\n",
		},
		{
			name:     "variables_comprehensive",
			args:     []string{"cow-lang", "../../examples/variables_comprehensive.cow"},
			expected: "42\n100\n42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			config := Config{
				Args:   tt.args,
				Output: &output,
			}

			err := Run(config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			actual := output.String()
			if actual != tt.expected {
				t.Errorf("output mismatch:\nexpected:\n%s\nactual:\n%s", tt.expected, actual)
			}
		})
	}
}

func TestCLIWithDebugFlag(t *testing.T) {
	var output bytes.Buffer
	config := Config{
		Args:   []string{"cow-lang", "--debug", "../../examples/hello_println.cow"},
		Output: &output,
	}

	err := Run(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actual := output.String()

	// With debug flag, we expect debug output plus the program output
	// Verify some debug output is present (should contain GRAMMAR or FIRST SETS)
	if !strings.Contains(actual, "GRAMMAR:") && !strings.Contains(actual, "FIRST SETS:") {
		t.Errorf("expected debug output to be present, got: %q", actual)
	}

	// Just verify that we get the expected program output at the end
	if !strings.HasSuffix(actual, "42\n") {
		// Show last 20 characters or entire string if shorter
		suffix := actual
		if len(actual) > 20 {
			suffix = actual[len(actual)-20:]
		}
		t.Errorf("expected output to end with '42\\n', got: %q", suffix)
	}
}

func TestCLIMissingFile(t *testing.T) {
	var output bytes.Buffer
	config := Config{
		Args:   []string{"cow-lang"},
		Output: &output,
	}

	err := Run(config)
	if err == nil {
		t.Fatal("expected error for missing file argument")
	}

	expectedError := "usage: cow-lang [--debug] <file.cow>"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestCLIFileNotFound(t *testing.T) {
	var output bytes.Buffer
	config := Config{
		Args:   []string{"cow-lang", "nonexistent.cow"},
		Output: &output,
	}

	err := Run(config)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// Error message should mention the file
	if !strings.Contains(err.Error(), "nonexistent.cow") {
		t.Errorf("expected error to mention file name, got: %v", err)
	}
}
