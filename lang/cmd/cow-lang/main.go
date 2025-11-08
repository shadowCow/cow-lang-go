package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shadowCow/cow-lang-go/lang/runner"
)

func main() {
	// Parse command line arguments
	debug := false
	var filePath string

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--debug] <file.cow>\n", os.Args[0])
		os.Exit(1)
	}

	// Check for --debug flag
	argIdx := 1
	if os.Args[argIdx] == "--debug" {
		debug = true
		argIdx++
	}

	if argIdx >= len(os.Args) {
		fmt.Fprintf(os.Stderr, "Usage: %s [--debug] <file.cow>\n", os.Args[0])
		os.Exit(1)
	}

	filePath = os.Args[argIdx]

	// Run the Cow program
	if err := runner.Run(filePath, os.Stdout, debug); err != nil {
		log.Fatal(err)
	}
}
