# Lang Module

This module contains the complete implementation of the Cow programming language, including lexical analysis, parsing, evaluation, and a command-line interpreter.

## Current Implementation

- **Automata** - Finite automata (NFA/DFA) for lexical analysis
- **Grammar** - Grammar definition framework
- **Lexer** - Lexical analyzer that transforms character streams into tokens
- **Parser** - Syntax analyzer that builds an AST from tokens
- **AST** - Abstract Syntax Tree node definitions
- **Evaluator** - Tree-walking interpreter for executing programs
- **Runner** - High-level API for running Cow programs
- **Command** - CLI tool for executing .cow files

## Directory Structure

Following standard Go project layout conventions:

```
lang/
├── cmd/
│   └── cow-lang/      # Command-line interpreter executable
│       └── main.go
├── ast/               # Abstract syntax tree node types
├── automata/          # Finite automata (NFA/DFA) implementation
├── eval/              # Interpreter/evaluator
├── grammar/           # Grammar definition framework
├── langdef/           # Cow language grammar definition
├── lexer/             # Lexical analysis
├── parser/            # Syntax analysis
├── runner/            # High-level API for running programs
├── examples/          # Example Cow programs
└── bin/               # Built binaries (gitignored)
```

All packages are importable by external Go modules except those in `internal/` (none currently).

## Package dependency directions

Certain package dependency rules should be followed to avoid import cycles.

`automata` depends on `grammar`
`langdef` depends on `grammar`
`lexer` depends on `automata`

## Building and Running

### Build the CLI tool

```bash
go build -o bin/cow-lang ./cmd/cow-lang
```

### Run a Cow program

```bash
./bin/cow-lang examples/hello_numbers.cow
```

## Using as a Library

```go
import (
    "os"
    "github.com/shadowCow/cow-lang-go/lang/runner"
)

// Simple API - run a .cow file
runner.Run("program.cow", os.Stdout)
```

Or use individual packages:

```go
import (
    "github.com/shadowCow/cow-lang-go/lang/automata"
    "github.com/shadowCow/cow-lang-go/lang/langdef"
    "github.com/shadowCow/cow-lang-go/lang/lexer"
    "github.com/shadowCow/cow-lang-go/lang/parser"
    "github.com/shadowCow/cow-lang-go/lang/eval"
)
```
