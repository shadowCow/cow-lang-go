# LL(1) Parser Generator Implementation

This document describes the LL(1) parser generator implementation that replaces the previous recursive descent parser.

## Overview

The LL(1) parser generator mirrors the lexer's DFA generation approach:
- **Lexer**: Lexical patterns → NFA → DFA → Tokenizer
- **Parser**: Syntactic grammar → FIRST/FOLLOW sets → Parse table → Stack-based parser

## Architecture

### Package: `lang/ll1/`

The LL(1) implementation consists of several components:

#### 1. FIRST Set Computation (`first.go`)

Computes FIRST sets for all grammar symbols:
- `FIRST(terminal)` = `{terminal}`
- `FIRST(NonTerminal)` = computed from its productions
- Handles sequences: `FIRST(A B C)` includes `FIRST(A)`, and if A is nullable, includes `FIRST(B)`, etc.
- Handles alternatives: `FIRST(A | B | C)` = `FIRST(A) ∪ FIRST(B) ∪ FIRST(C)`
- Handles operators: `?`, `*`, `+` with appropriate nullability

**Key types:**
```go
type FirstSets struct {
    sets     map[string]map[string]bool  // Symbol -> FIRST set
    nullable map[grammar.Symbol]bool     // Which non-terminals can derive ε
}
```

#### 2. FOLLOW Set Computation (`follow.go`)

Computes FOLLOW sets for all non-terminals:
- `FOLLOW(StartSymbol)` includes `$` (end-of-input)
- For production `A -> α B β`:
  - `FIRST(β) - {ε}` ⊆ `FOLLOW(B)`
  - If β is nullable: `FOLLOW(A)` ⊆ `FOLLOW(B)`
- Iterates until fixpoint is reached

**Key types:**
```go
type FollowSets struct {
    sets map[grammar.Symbol]map[string]bool  // NonTerminal -> FOLLOW set
}
```

#### 3. Parse Table Generation (`table.go`)

Builds the LL(1) parse table `M[NonTerminal, Terminal] -> Production`:

**Algorithm:**
- For each production `A -> α`:
  - For each terminal `t` in `FIRST(α)`: add entry `M[A, t] = A -> α`
  - If α is nullable, for each terminal `t` in `FOLLOW(A)`: add entry `M[A, t] = A -> α`

**Conflict Detection:**
- If any cell `M[A, t]` would contain multiple different productions, the grammar is **not LL(1)**
- Returns detailed error showing:
  - Which cell has the conflict
  - Which productions conflict
  - Why they conflict

**Key types:**
```go
type ParseTable struct {
    table map[tableKey]grammar.ProductionRule
}

type GrammarNotLL1Error struct {
    Conflicts []Conflict  // Details of all conflicts found
}
```

#### 4. Stack-Based Parser (`parser.go`)

Table-driven parser using an explicit stack:

**Algorithm:**
```
stack = [EOF, StartSymbol]
while stack not empty:
    top = stack.pop()
    lookahead = current token

    if top is terminal:
        match terminal with lookahead
    else:  // top is non-terminal
        production = table[top, lookahead]
        push production symbols onto stack (in reverse order)
```

**Features:**
- Builds AST during parsing
- Provides detailed error messages with line/column information
- Optional trace mode for educational purposes

**Key types:**
```go
type Parser struct {
    table   *ParseTable
    grammar grammar.SyntacticGrammar
    tokens  []lexer.Token
    trace   bool  // Enable parse step tracing
}
```

#### 5. Debug/Visualization (`debug.go`)

Educational features for understanding LL(1) parsing:
- `PrintGrammar()` - Shows grammar productions
- `PrintFirstSets()` - Displays FIRST sets for all symbols
- `PrintFollowSets()` - Displays FOLLOW sets for all non-terminals
- `PrintParseTable()` - Pretty-prints parse table as a grid

Example output:
```
GRAMMAR:
========
Start symbol: Program

Productions:
  Literal -> (INT_DECIMAL | INT_HEX | INT_BINARY | FLOAT)
  Program -> Literal

FIRST SETS:
===========
  FIRST(Literal) = {FLOAT, INT_BINARY, INT_DECIMAL, INT_HEX}
  FIRST(Program) = {FLOAT, INT_BINARY, INT_DECIMAL, INT_HEX}

FOLLOW SETS:
============
  FOLLOW(Literal) = {$}
  FOLLOW(Program) = {$}

LL(1) PARSE TABLE:
==================
             | $               | FLOAT           | INT_BINARY      | INT_DECIMAL     | INT_HEX         |
  -----------+-----------------+-----------------+-----------------+-----------------+-----------------+
  Literal    |                 | FLOAT           | INT_BINARY      | INT_DECIMAL     | INT_HEX         |
  Program    |                 | Literal         | Literal         | Literal         | Literal         |
```

## Integration

### Runner Updates (`lang/runner/runner.go`)

Two functions for running Cow programs:

1. **`Run(filePath, output)`** - Normal execution
   - Compiles grammar to parse table once
   - Parses input using LL(1) parser
   - Evaluates AST

2. **`RunWithDebug(filePath, output)`** - Educational mode
   - Prints grammar
   - Prints FIRST/FOLLOW sets
   - Prints parse table
   - Enables parse tracing
   - Shows each stack operation

### CLI Updates (`lang/cmd/cow-lang/main.go`)

Added `--debug` flag:
```bash
# Normal execution
./bin/cow-lang program.cow

# Debug mode (shows LL(1) internals)
./bin/cow-lang --debug program.cow
```

## Testing

### Comprehensive Test Suite (`lang/ll1/table_test.go`)

Tests cover:

1. **LL(1) Conflict Detection**
   - Creates a non-LL(1) grammar with conflicts
   - Verifies that conflict detection works
   - Checks error message quality

2. **Valid LL(1) Grammar**
   - Creates a valid LL(1) grammar
   - Verifies parse table is generated successfully
   - Checks table entries are correct

3. **FIRST Set Computation**
   - Tests sequences with nullable elements
   - Verifies epsilon handling

4. **FOLLOW Set Computation**
   - Tests FOLLOW propagation through productions
   - Verifies EOF marker handling

### Integration Tests

All existing tests updated to work with LL(1) parser:
- `lang/runner/runner_test.go` - Updated for simple grammar
- `lang/langdef/interpreter_integration_test.go` - Removed (old parser tests)

## Comparison: DFA vs PDA

| Aspect | Lexer (DFA) | Parser (LL(1)) |
|--------|-------------|----------------|
| Input | Lexical patterns (regex) | Context-free grammar |
| Build Phase | Pattern → NFA → DFA | Grammar → FIRST/FOLLOW → Parse Table |
| Runtime | Simulate DFA on character stream | Stack machine with parse table |
| Output | Tokens | AST |
| Recognizes | Regular languages | Context-free languages (LL(1) subset) |

## Key Differences from Recursive Descent

### Recursive Descent (Previous)
- ✅ Easy to write and understand
- ✅ Excellent error messages
- ✅ Can handle any grammar
- ❌ Manual implementation for each rule
- ❌ Grammar changes require code changes

### LL(1) Table-Driven (New)
- ✅ Automatically generated from grammar
- ✅ Grammar changes just regenerate table
- ✅ Theoretically elegant
- ✅ Educational - shows parser internals
- ❌ Only works for LL(1) grammars
- ⚠️ Error messages can be less intuitive

## Current Grammar

The current grammar is intentionally simple to demonstrate the LL(1) approach:

```
Program -> Literal
Literal -> INT_DECIMAL | INT_HEX | INT_BINARY | FLOAT
```

This grammar is **LL(1)** because:
- No conflicts: each alternative starts with a different terminal
- No ambiguity
- FIRST sets are disjoint for all alternatives

## Future Extensions

To support more complex language features, extend the grammar in `lang/langdef/syntactic.go`:

### Example: Adding Function Calls
```go
SYM_STATEMENT: grammar.NonTerminal{Symbol: SYM_EXPRESSION},
SYM_EXPRESSION: grammar.SynAlternative{
    grammar.NonTerminal{Symbol: SYM_CALL},
    grammar.NonTerminal{Symbol: SYM_LITERAL},
},
SYM_CALL: grammar.SynSequence{
    grammar.Terminal{TokenType: TOKEN_IDENTIFIER},
    grammar.Terminal{TokenType: TOKEN_LPAREN},
    grammar.NonTerminal{Symbol: SYM_ARGUMENTS},
    grammar.Terminal{TokenType: TOKEN_RPAREN},
},
```

The LL(1) parser generator will automatically:
1. Compute FIRST/FOLLOW sets for new productions
2. Build the parse table
3. Detect any LL(1) conflicts
4. Generate detailed error messages if the grammar isn't LL(1)

### Handling Non-LL(1) Grammars

If a grammar isn't LL(1), you have options:

1. **Refactor the grammar** - Left-factor or eliminate left recursion
2. **Use different parser** - Switch to LR(1), LALR, or GLR
3. **Add precedence rules** - Resolve conflicts with priorities

## Educational Value

This implementation demonstrates:
- The relationship between automata theory and parsing
- How FIRST/FOLLOW sets enable predictive parsing
- How parse tables work
- The limitations of LL(1) grammars
- The parallel between lexical and syntactic analysis

## Performance Characteristics

- **Table Generation**: O(n³) worst case (n = grammar size)
- **Parsing**: O(n) where n = input length
- **Space**: O(|NonTerminals| × |Terminals|) for parse table

The parse table is generated once and can be cached, making parsing very efficient.

## Conclusion

The LL(1) parser generator successfully replicates the lexer's approach of compiling a declarative grammar specification into an efficient runtime engine. This provides a clean, maintainable foundation for language development while serving as an educational tool for understanding parsing theory.
