// Package cli provides the command-line interface adapter for the Cow language.
// This package handles argument parsing and delegates to the runner for execution.
package cli

import (
	"fmt"
	"io"

	"github.com/shadowCow/cow-lang-go/lang/runner"
)

// Config holds the configuration for the CLI.
type Config struct {
	Args   []string  // Command-line arguments (including program name)
	Output io.Writer // Output stream for program output
}

// Run executes the CLI with the given configuration.
// It parses the arguments, validates them, and delegates to the runner.
func Run(config Config) error {
	// Parse arguments
	debug := false
	var filePath string

	// Skip program name (first argument)
	args := config.Args[1:]

	// Parse flags and arguments
	for len(args) > 0 {
		arg := args[0]
		if arg == "--debug" {
			debug = true
			args = args[1:]
		} else {
			filePath = arg
			args = args[1:]
			break
		}
	}

	// Validate that a file path was provided
	if filePath == "" {
		return fmt.Errorf("usage: cow-lang [--debug] <file.cow>")
	}

	// Execute the file using the runner
	return runner.Run(filePath, config.Output, debug)
}
