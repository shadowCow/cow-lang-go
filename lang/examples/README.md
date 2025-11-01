# Cow Language Examples

This directory contains example Cow language programs.

## hello_numbers.cow

A simple program that demonstrates printing various number literals:
- Decimal integers
- Hexadecimal integers (0xFF)
- Binary integers (0b1010)
- Floating-point numbers
- Scientific notation
- Numbers with underscores for readability

## Running Examples

### Using the Runner Package (Easiest)

The simplest way to run these examples is using the `runner` package:

```go
package main

import (
    "log"
    "os"

    "github.com/shadowCow/cow-lang-go/lang/runner"
)

func main() {
    if err := runner.Run("hello_numbers.cow", os.Stdout); err != nil {
        log.Fatal(err)
    }
}
```

### Manual Pipeline (Advanced)

You can also run the complete pipeline manually:
1. Read the source file
2. Compile the lexical grammar using `langdef.GetLexical()`
3. Tokenize using `lexer.NewLexer(dfa, source)`
4. Parse using `parser.NewParser(tokens)`
5. Evaluate using `eval.NewEvaluator(os.Stdout)`

See the integration tests in `langdef/interpreter_integration_test.go` for complete examples.
