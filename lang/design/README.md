# Language Design Documentation

This directory contains comprehensive design documentation for the language.

## Reading Guide

**Start here**: [OVERVIEW.md](OVERVIEW.md)
- High-level vision and core concepts
- Why these features together
- Target use cases

**Then read**: [PHASED_APPROACH.md](PHASED_APPROACH.md)
- Implementation roadmap (5 phases)
- What to build and in what order
- Deliverables for each phase

**For syntax**: [SYNTAX_DECISIONS.md](SYNTAX_DECISIONS.md)
- Lexical elements (keywords, operators, literals)
- Syntax for each feature
- Example programs
- Deferred decisions

**For context**: [LANGUAGE_INFLUENCES.md](LANGUAGE_INFLUENCES.md)
- What we're borrowing from existing languages
- How concepts map to this language
- What we're adapting and why

**For specifics**: [TECHNICAL_DETAILS.md](TECHNICAL_DETAILS.md)
- Communication model (typed channels)
- Collection bounds (compile-time vs runtime)
- Consistency model (application-defined)
- Memory model, serialization, error handling

## Current Status

**Phase**: Planning
**Next Step**: Implement Phase 1 (Core Functional Language)

The grammar framework is in place (`lang/grammar/` package). Next task is to define the actual language grammar for Phase 1 features and begin parser implementation.

## Quick Reference

### Core Concepts
1. **ADTs** - Algebraic data types with pattern matching
2. **FSTs** - Finite State Transducers as primitives: `(State, Command) -> (State, Event)`
3. **Event-Driven** - Reactive architecture
4. **Ports/Adapters** - Enforced clean architecture
5. **Distributed** - Ownership semantics: `Local`, `RemoteView`, `Replicated`, `Distributed`
6. **Bounded** - No unbounded collections, only bounded collections and streams

### Syntax Style
- **Overall**: Rust-like (braces, semicolons, `fn`, `let`, `match`)
- **ADTs**: ML-style (`type Option<T> = Some of T | None`)
- **FSTs**: TBD (deferred until after Phase 1)

### Language Equation
```
Rust syntax
+ OCaml/Haskell ADTs
+ Erlang actors
+ Elm architecture
+ Pony capabilities (adapted)
+ Flink streaming
+ Effect system
+ Embedded constraints
+ CRDT consistency
= This Language
```

## Phase 1 Example

```rust
type Option<T> = Some of T | None;

fn map<A, B>(opt: Option<A>, f: fn(A) -> B) -> Option<B> {
    match opt {
        Some(x) => Some(f(x)),
        None => None
    }
}

fn main() -> () {
    let x = Some(42);
    let y = map(x, fn(n) -> n * 2);

    match y {
        Some(result) => println("Result: {}", result),
        None => println("No value")
    }
}
```

## Contributing to Design

When adding to or updating design docs:
1. Keep docs in sync with implementation decisions
2. Document the "why" not just the "what"
3. Include examples
4. Note trade-offs and alternatives considered
5. Update this README if adding new files
