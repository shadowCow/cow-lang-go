# Lang Package

This directory contains the core implementation of the Cow programming language, including lexical analysis, parsing, compilation, and runtime components.

## Current Implementation
- **Automata** - Data structures for finite automata used in lexical analysis
- **Lexer** - Lexical analyzer that transforms character streams into tokens

## Directory Structure
```
lang/
├── automata/          # Finite automata implementation
├── lexer/             # Lexical analysis
├── parser/            # Syntax analysis (future)
├── ast/               # Abstract syntax tree (future)
├── compiler/          # Code generation (future)
└── runtime/           # Language runtime (future)
```

## Usage
```go
import "github.com/shadowCow/cow-lang-go/lang/lexer"
import "github.com/shadowCow/cow-lang-go/lang/automata"
```