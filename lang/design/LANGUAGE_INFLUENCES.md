# Language Influences

This language draws inspiration from many existing languages. This document maps which features come from which languages and how they're adapted.

## Type System & Functional Programming

### Haskell
**What we're taking**:
- Algebraic Data Types (ADTs) - sum and product types
- Pattern matching with exhaustiveness checking
- Type classes (potentially, for polymorphism)
- Lazy evaluation for streams (not everything)

**What we're adapting**:
- Syntax is Rust-like, not Haskell-like
- Strictness by default (unlike Haskell's lazy-by-default)
- Bounded collections (Haskell has infinite lists)

**Example mapping**:
```haskell
-- Haskell
data Maybe a = Just a | Nothing

-- Our language
type Option<T> = Some of T | None;
```

### OCaml
**What we're taking**:
- ML-style type definition syntax (`type T = A of X | B of Y`)
- Pattern matching syntax inspiration
- Practical functional programming approach

**What we're adapting**:
- Rust-like braces instead of indentation
- No object system (at least not initially)

### F#
**What we're taking**:
- Balance of functional and practical features
- Computation expressions concept (potentially for effects)

### Rust
**What we're taking**:
- Overall syntax style (braces, semicolons, `fn`, `let`, `match`)
- Enum pattern matching
- Expression-oriented language (blocks return values)
- Type safety without garbage collection (eventually)

**What we're NOT taking**:
- Ownership/borrowing system (we have different distribution semantics)
- Trait system (at least initially)
- Unsafe blocks (want to prove safety differently)

### Scala
**What we're taking**:
- Actor model integration (Akka-style)
- Mixing functional and concurrent features

---

## Concurrency & Distribution

### Erlang / Elixir
**What we're taking**:
- Actor model with isolated state
- Message passing between processes
- OTP patterns (gen_server, gen_statem) → FSTs
- Location transparency (process can be local or remote)
- "Let it crash" philosophy (potentially)

**Example mapping**:
```erlang
%% Erlang gen_server callback
handle_call(get, _From, State) ->
    {reply, State, State};
handle_call({set, Value}, _From, _State) ->
    {reply, ok, Value}.

%% Our FST (concept)
fn transition(state: i32, cmd: Command) -> (i32, Event) {
    match cmd {
        Get => (state, CurrentValue(state)),
        Set(value) => (value, Changed(value))
    }
}
```

**What we're adapting**:
- Stronger typing (Erlang is dynamically typed)
- Explicit FST transition signatures
- Different distribution ownership model

### Pony
**What we're taking**:
- Actor model with type safety
- Reference capabilities concept (adapted to distribution)
- Compile-time concurrency safety

**What we're adapting**:
- Different capability model focused on distribution, not just concurrency
- Our ownership types: `Local`, `RemoteView`, `Replicated`, `Distributed`

### Akka (Scala/Java)
**What we're taking**:
- Typed actors
- Actor supervision (potentially)
- Distributed actor systems

**What we're NOT taking**:
- JVM dependency
- Dynamic features

---

## Event-Driven & State Machines

### Elm
**What we're taking**:
- Architecture pattern: State + Message → (State, Effect)
- Strong typing for state transitions
- Pure update functions

**Our adaptation**:
```elm
-- Elm
update : Msg -> Model -> (Model, Cmd Msg)

-- Our FST
transition : (State, Command) -> (State, Event)
```

**Key difference**: We emit `Event` instead of `Cmd` (effect description). Events are data, not effect instructions.

### XState (JavaScript)
**Conceptual influence**:
- State machines as first-class entities
- Explicit state transitions
- Statecharts for complex FSMs

**What we're NOT taking**:
- JavaScript dynamic typing
- Runtime configuration (we want compile-time checking)

---

## Resource Bounds & Real-Time

### ATS (Applied Type System)
**What we're taking**:
- Dependent types for resource bounds (simplified version)
- Compile-time size tracking
- Linear types for resource management

**What we're NOT taking**:
- Full dependent type complexity
- Complex proof obligations

### Idris
**Conceptual influence**:
- Sized types
- Dependent types for correctness

**What we're adapting**:
- Simpler approach: fixed sizes, not full dependent types
- Pragmatic over theoretically perfect

### Ada / SPARK
**What we're taking**:
- Safety for embedded systems
- Bounded collections
- Compile-time verification goals

### Embedded Systems Languages (C with constraints)
**What we're taking**:
- No unbounded allocations
- Predictable memory usage
- Fixed-size buffers

**What we're improving**:
- Type safety
- Memory safety without manual management

---

## Streaming & Data Processing

### Apache Flink
**What we're taking**:
- Windowing concepts (tumbling, sliding, session windows)
- Bounded views of unbounded streams
- Event time vs processing time

### Kafka Streams
**What we're taking**:
- Stream processing abstractions
- State in stream processing
- Bounded state stores

### Reactive Extensions (Rx)
**Conceptual influence**:
- Observable streams
- Compositional operators

---

## Effect Systems & Clean Architecture

### Koka
**What we're taking**:
- Algebraic effect system concept
- Effect types in function signatures
- Effect handlers

**What we're adapting**:
- Simpler initial version
- Port-based effects rather than full algebraic effects

### Eff
**Conceptual influence**:
- Effect handlers
- Composing effects

### Unison
**What we're taking**:
- Distributed-first design
- Content-addressed code
- Effects tracked in types
- Ability patterns (similar to our ports)

**What we're NOT taking** (initially):
- Content addressing for code
- Distributed codebase

---

## Distribution & Consistency

### CRDTs (research)
**What we're taking**:
- Conflict-free data types
- Eventual consistency primitives
- Well-studied merge semantics

**Types to include**:
- G-Counter, PN-Counter (counters)
- LWW-Register (last-write-wins)
- OR-Set (observed-remove set)
- RGA (replicated growable array)

### Raft / Paxos (distributed consensus)
**What we're taking**:
- Strong consistency when needed
- Leader election patterns
- Consensus primitives

**How we use it**:
- Optional for when strong consistency required
- Most things use CRDTs (eventual consistency)

### Orleans (Microsoft)
**What we're taking**:
- Virtual actors distributed across cluster
- Location transparency
- Actor activation/deactivation

---

## Putting It All Together

This language is essentially:
```
Rust syntax
+ OCaml/Haskell ADTs and pattern matching
+ Erlang actors and distribution
+ Elm architecture
+ Pony reference capabilities (adapted for distribution)
+ Flink/Kafka streaming
+ Koka/Eff effect system (simplified)
+ Embedded systems resource constraints
+ CRDT eventual consistency
= This Language
```

Each influence provides a piece of the puzzle:
- **Functional languages** → Type safety and correctness
- **Actor languages** → Concurrency and distribution
- **Reactive/streaming** → Event-driven processing
- **Effect systems** → Clean architecture enforcement
- **Embedded languages** → Resource bounds and predictability
- **Distributed systems** → Ownership and consistency models

The innovation isn't in any single feature, but in combining these well-understood concepts into a coherent whole optimized for distributed embedded stream processing.
