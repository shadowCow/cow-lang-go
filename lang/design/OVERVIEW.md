# Language Design Overview

## Vision

A programming language designed for **distributed embedded stream processors** that combines:
- Type safety and functional programming
- Event-driven architecture with actor-like concurrency
- Explicit resource bounds for embedded/real-time systems
- Distribution-aware semantics with flexible consistency models

## Target Use Cases

1. **Distributed Systems / Microservices** - Services communicating across network boundaries
2. **Embedded / Resource-Constrained** - Limited memory with compile-time guarantees
3. **Real-Time Data Processing** - Stream processing and event handling with timing constraints

## Core Concepts

### 1. Algebraic Data Types (ADTs)
Union types and pattern matching similar to Haskell, OCaml, and Rust.

**Example concept:**
```
type Option<T> = Some of T | None
type Result<T, E> = Ok of T | Err of E
```

### 2. Event-Driven Architecture
Programs respond to events rather than executing imperative sequences. Natural fit for stream processing and reactive systems.

### 3. Finite State Transducers (FSTs) as Primitives
The fundamental unit of computation is an FST with signature:
```
(State, Command) -> (State, Event)
```

Similar to Elm architecture but:
- **State**: Internal state (ADT)
- **Command**: Input message triggering transition
- **Event**: Output message emitted after transition

FSTs combine:
- Actor-like isolation (each FST has private state)
- Event-driven processing (react to commands, emit events)
- Pure logic (transition function is pure)

### 4. Ports & Adapters / Clean Architecture
The language enforces separation between:
- **Pure business logic** (FST transition functions)
- **Effectful operations** (I/O, network, external systems)

Ports define boundaries; adapters implement integration with the outside world.

### 5. Distributed-First with Ownership Semantics
Data has explicit ownership and distribution characteristics:
- `Local<T>` - Local-only data
- `RemoteView<T>` - Local view of remotely-owned data
- `Replicated<T>` - Locally-owned data replicated to remote nodes
- `Distributed<T>` - Data with independent local/remote changes requiring sync

### 6. Bounded Collections, Unbounded Streams
**No unbounded in-memory collections** - prevents memory exhaustion in embedded/real-time systems.

- Collections are fixed-size (compile-time or runtime-parameterized)
- Streams can be unbounded but are consumed incrementally
- Windowing operations make bounded views of streams

**Example concepts:**
```
[i32; 10]              // Compile-time sized array
Vec<i32, max=100>      // Runtime-bounded vector
Stream<Event>          // Unbounded stream
```

## Design Principles

1. **Make the implicit explicit** - Resource usage, effects, distribution, ownership
2. **Fail at compile time, not runtime** - Leverage type system for correctness
3. **Composable abstractions** - FSTs compose via typed channels
4. **Predictable performance** - Bounded memory, deterministic behavior
5. **Distribution-aware** - Don't hide network boundaries; make them first-class

## Why These Features Together?

The constraints reinforce each other:
- **Real-time + embedded** → bounded memory, deterministic behavior
- **Distributed + typed channels** → explicit communication paths, easier reasoning
- **FSTs + pure transitions** → testable, composable, portable across network
- **Application-defined consistency** → flexibility for different guarantees per use case

This isn't just another general-purpose language - it's optimized for a specific problem domain where all these concerns matter simultaneously.
