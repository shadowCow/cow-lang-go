package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shadowCow/cow-lang-go/lang/runner"
)

func main() {
	// Check for exactly one command line argument
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.cow>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Run the Cow program
	if err := runner.Run(filePath, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
