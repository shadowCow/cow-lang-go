# Language Tooling Library

A generic, reusable library for building programming language implementations in Go. This library provides the foundational components for lexical analysis, parsing, and grammar processing.

## Overview

This tooling library was extracted from the Cow programming language implementation to create a reusable foundation for language development. It provides:

- **Grammar Definitions** - Abstract representations of lexical and syntactic grammars
- **Automata Theory** - NFA and DFA construction and conversion algorithms
- **Lexical Analysis** - Table-driven tokenization using compiled DFAs
- **LL(1) Parsing** - Automatic parser generation from context-free grammars
- **Parse Trees** - Generic tree structures for representing parsed input

## Architecture

```
Grammar Definition → Compilation → Runtime Engine → Output
```

### Lexical Pipeline
```
Lexical Patterns → NFA (Thompson) → DFA (Subset Construction) → Tokenizer → Tokens
```

### Syntactic Pipeline
```
Syntactic Grammar → FIRST/FOLLOW Sets → Parse Table → LL(1) Parser → Parse Tree
```

## Packages

### `grammar/`
Defines abstract data structures for representing grammars.

**Lexical Patterns:**
- `Literal` - Match exact text: `"if"`, `"while"`
- `CharRange` - Match character range: `[a-z]`, `[0-9]`
- `CharSet` - Match character set: `[abc]`
- `AnyChar` - Match any character: `.`
- `AnyCharExcept` - Match anything except: `[^abc]`
- `LexSequence` - Match sequence: `A B C`
- `LexAlternative` - Match alternatives: `A | B | C`
- `LexOptional` - Match zero or one: `A?`
- `LexZeroOrMore` - Match zero or more: `A*`
- `LexOneOrMore` - Match one or more: `A+`

**Syntactic Rules:**
- `Terminal` - Reference to a token type
- `NonTerminal` - Reference to another production rule
- `SynSequence` - Ordered sequence of rules
- `SynAlternative` - Choice between rules
- `SynOptional`, `SynZeroOrMore`, `SynOneOrMore` - Repetition operators

### `automata/`
Implements finite automata for pattern matching.

**Key Components:**
- `NFA` - Non-deterministic Finite Automaton with epsilon transitions
- `DFA` - Deterministic Finite Automaton
- `CompilePatternToNFA` - Thompson's construction algorithm
- `NFAToDFAWithTokens` - Subset construction with token priority

**Usage:**
```go
pattern := grammar.LexSequence{
    grammar.Literal{Text: "if"},
    grammar.CharSet{Chars: []rune{' ', '\t'}},
}
nfa := automata.CompilePatternToNFA(pattern)
```

### `lexer/`
Table-driven lexical analyzer.

**Features:**
- Longest-match tokenization
- UTF-8 support
- Position tracking (line, column, offset)
- Error reporting with location information

**Usage:**
```go
// Define tokens
lexGrammar := grammar.LexicalGrammar{
    Tokens: []grammar.TokenDefinition{
        {TokenType: "KEYWORD_IF", Pattern: grammar.Literal{Text: "if"}, Priority: 2},
        {TokenType: "IDENTIFIER", Pattern: grammar.CharRange{From: 'a', To: 'z'}, Priority: 1},
    },
}

// Compile to DFA
dfa := automata.CompileLexicalGrammar(lexGrammar)

// Tokenize
lex := lexer.NewLexer(dfa, sourceCode)
tokens, err := lex.Tokenize()
```

### `ll1/`
LL(1) parser generation and execution.

**Components:**
- `first.go` - Compute FIRST sets for grammar symbols
- `follow.go` - Compute FOLLOW sets for non-terminals
- `table.go` - Generate LL(1) parse tables with conflict detection
- `parser.go` - Table-driven parser that produces parse trees
- `debug.go` - Visualization utilities for grammar analysis

**Features:**
- Automatic conflict detection
- Detailed error messages for non-LL(1) grammars
- Parse tracing for debugging
- Pretty-printing of FIRST/FOLLOW sets and parse tables

**Usage:**
```go
// Define grammar
synGrammar := grammar.SyntacticGrammar{
    StartSymbol: "Program",
    Productions: map[grammar.Symbol]grammar.ProductionRule{
        "Program": grammar.NonTerminal{Symbol: "Statement"},
        "Statement": grammar.SynAlternative{
            grammar.Terminal{TokenType: "NUMBER"},
            grammar.Terminal{TokenType: "STRING"},
        },
    },
}

// Build parser
firstSets := ll1.ComputeFirstSets(synGrammar)
followSets := ll1.ComputeFollowSets(synGrammar, firstSets)
parseTable, err := ll1.BuildParseTable(synGrammar, firstSets, followSets)
if err != nil {
    // Grammar is not LL(1) - error contains conflict details
}

// Parse
parser := ll1.NewParser(parseTable, synGrammar, tokens, "WHITESPACE")
parseTree, err := parser.Parse()
```

### `parsetree/`
Generic parse tree structures.

**Node Types:**
- `TerminalNode` - Leaf node containing a matched token
- `NonTerminalNode` - Interior node with child nodes
- `ProgramNode` - Root of the parse tree
- `EmptyNode` - Represents epsilon productions

**Structure:**
```go
type ParseTree interface {
    NodeType() string
    String() string
}
```

Parse trees mirror the grammatical structure and can be converted to language-specific ASTs.

## Example: Building a Simple Language

Here's a complete example of building a calculator language:

```go
package main

import (
    "fmt"
    "github.com/shadowCow/cow-lang-go/tooling/grammar"
    "github.com/shadowCow/cow-lang-go/tooling/automata"
    "github.com/shadowCow/cow-lang-go/tooling/lexer"
    "github.com/shadowCow/cow-lang-go/tooling/ll1"
)

func main() {
    // 1. Define lexical grammar
    lexGrammar := grammar.LexicalGrammar{
        Tokens: []grammar.TokenDefinition{
            {
                TokenType: "NUMBER",
                Pattern: grammar.LexOneOrMore{
                    Inner: grammar.CharRange{From: '0', To: '9'},
                },
                Priority: 1,
            },
            {
                TokenType: "PLUS",
                Pattern: grammar.Literal{Text: "+"},
                Priority: 1,
            },
            {
                TokenType: "WHITESPACE",
                Pattern: grammar.CharSet{Chars: []rune{' ', '\t', '\n'}},
                Priority: 1,
            },
        },
    }

    // 2. Define syntactic grammar
    synGrammar := grammar.SyntacticGrammar{
        StartSymbol: "Expr",
        Productions: map[grammar.Symbol]grammar.ProductionRule{
            "Expr": grammar.SynSequence{
                grammar.Terminal{TokenType: "NUMBER"},
                grammar.Terminal{TokenType: "PLUS"},
                grammar.Terminal{TokenType: "NUMBER"},
            },
        },
    }

    // 3. Compile lexer
    dfa := automata.CompileLexicalGrammar(lexGrammar)
    lex := lexer.NewLexer(dfa, "5 + 3")
    tokens, _ := lex.Tokenize()

    // 4. Build parser
    firstSets := ll1.ComputeFirstSets(synGrammar)
    followSets := ll1.ComputeFollowSets(synGrammar, firstSets)
    parseTable, _ := ll1.BuildParseTable(synGrammar, firstSets, followSets)

    // 5. Parse
    parser := ll1.NewParser(parseTable, synGrammar, tokens, "WHITESPACE")
    parseTree, _ := parser.Parse()

    fmt.Println(parseTree.String())
    // Output: Program{NonTerminal{Expr: [Terminal{NUMBER:"5"}, Terminal{PLUS:"+"}, Terminal{NUMBER:"3"}]}}
}
```

## Debugging Tools

The library includes visualization tools for grammar analysis:

```go
// Print grammar
ll1.PrintGrammar(synGrammar)

// Print FIRST sets
ll1.PrintFirstSets(firstSets)

// Print FOLLOW sets
ll1.PrintFollowSets(followSets)

// Print parse table
ll1.PrintParseTable(parseTable)

// Enable parse tracing
parser.SetTrace(true)
```

## Design Principles

### Separation of Concerns
- **Grammar definition** is declarative and language-agnostic
- **Compilation** happens once (DFA/parse table generation)
- **Runtime engines** (lexer/parser) are generic and reusable

### Language Independence
- No language-specific logic in the tooling
- Parse trees are generic structures
- Language implementations convert parse trees to their own AST

### Composability
- Each component can be used independently
- Clear interfaces between components
- Easy to extend or replace individual parts

## LL(1) Grammar Requirements

For a grammar to be LL(1):

1. **No ambiguity** - Each production must be uniquely identifiable by lookahead
2. **Disjoint FIRST sets** - For `A -> α | β`, FIRST(α) ∩ FIRST(β) = ∅
3. **Nullable handling** - If α can derive ε, then FIRST(β) ∩ FOLLOW(A) = ∅

The parse table builder will detect and report conflicts with detailed explanations.

## Error Handling

The library provides detailed error messages:

**Lexer errors:**
- Character position (line, column, offset)
- Unexpected character details

**Parser errors:**
- Token position
- Expected vs actual token
- Context (which non-terminal was being parsed)

**Grammar errors:**
- Conflict location in parse table
- Which productions conflict
- Why the grammar isn't LL(1)

## Performance

- **Lexer**: O(n) where n is input length (DFA simulation)
- **Parser**: O(n) where n is token count (table-driven)
- **Grammar compilation**: O(n³) worst case for parse table generation

Parse tables and DFAs can be cached/precompiled for production use.

## Testing

Run all tests:
```bash
go test ./...
```

The test suite includes:
- Automata construction and conversion
- Lexical pattern matching
- FIRST/FOLLOW set computation
- Parse table generation with conflict detection
- End-to-end parsing scenarios

## License

Part of the cow-lang-go project.

## Contributing

This library was extracted from a specific language implementation. Contributions that increase generality and reusability are welcome.

## References

- **Dragon Book**: "Compilers: Principles, Techniques, and Tools" by Aho, Sethi, Ullman
- **Thompson's Construction**: Original NFA construction algorithm
- **Subset Construction**: NFA to DFA conversion algorithm
- **LL(1) Parsing**: Top-down predictive parsing technique