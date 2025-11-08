package main

import (
	"log"
	"os"

	"github.com/shadowCow/cow-lang-go/lang/in/cli"
)

func main() {
	config := cli.Config{
		Args:   os.Args,
		Output: os.Stdout,
	}

	if err := cli.Run(config); err != nil {
		log.Fatal(err)
	}
}
