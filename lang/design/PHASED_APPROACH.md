# Phased Implementation Approach

This language has ambitious goals. To make progress incrementally, we'll build in phases, where each phase adds capabilities on top of the previous foundation.

## Phase 1: Core Functional Language

**Goal**: Get basic functional programming with ADTs and pattern matching working.

### Features
- **Primitive types**: `i32`, `i64`, `f32`, `f64`, `bool`, `string`, `unit` (`()`)
- **Algebraic Data Types**:
  - Sum types (unions): `type Option<T> = Some of T | None`
  - Product types: Records/structs, tuples
- **Functions**: First-class, closures, recursion
- **Pattern Matching**: Exhaustiveness checking, nested patterns
- **Let bindings**: Immutable by default
- **Basic expressions**:
  - Arithmetic: `+`, `-`, `*`, `/`, `%`
  - Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
  - Logical: `&&`, `||`, `!`
  - If/else expressions
  - Function calls

### Example Programs
```
type Option<T> = Some of T | None;

fn map<A, B>(opt: Option<A>, f: fn(A) -> B) -> Option<B> {
    match opt {
        Some(x) => Some(f(x)),
        None => None
    }
}

fn factorial(n: i32) -> i32 {
    if n <= 1 {
        1
    } else {
        n * factorial(n - 1)
    }
}
```

### Deliverables
- Complete lexical and syntactic grammar
- Parser producing AST
- Type checker for basic type system
- Interpreter or simple code generator

### Reference Languages
OCaml, Haskell, F#, Rust (enums + pattern matching)

---

## Phase 2: Bounded Collections & Streams

**Goal**: Memory-safe collections with explicit bounds.

### Features
- **Compile-time sized arrays**: `[T; N]` where N is compile-time constant
- **Runtime-bounded vectors**: `Vec<T, max=N>` or `BoundedVec<T>(capacity: usize)`
- **Unbounded streams**: `Stream<T>` - can be infinite but processed incrementally
- **Windowing operations**:
  - Tumbling windows
  - Sliding windows
  - Session windows
- **Collection operations**: map, filter, fold, take, drop
- **Type system tracks bounds**: Prevent buffer overflows at compile time

### Example Concepts
```
let arr: [i32; 10] = [0; 10];  // Array of 10 zeros

let vec = BoundedVec::new(100);  // Capacity 100, runtime parameter
vec.push(42);  // Returns Result - can fail if full

fn process_stream(stream: Stream<Event>) -> Stream<Result> {
    stream
        .window_tumbling(duration: 5s)
        .map(|events| aggregate(events))
}
```

### Deliverables
- Collection type definitions in type system
- Bounds checking (static where possible)
- Stream operators
- Error handling for bounded collection overflow

### Reference Languages
- Rust: arrays with const generics
- ATS, Idris: dependent types for sizes
- Flink, Kafka Streams: windowing operations
- Embedded systems languages: fixed-size buffers

---

## Phase 3: FST Primitives

**Goal**: Event-driven state machines as first-class language constructs.

### Features
- **FST definition syntax** (TBD - deferred decision)
- **State types**: Use ADTs from Phase 1
- **Command/Event types**: Message types using ADTs
- **Transition function**: `(State, Command) -> (State, Event)`
  - Pure function - no side effects
  - Pattern matching on state and command
- **Typed channels**: Point-to-point communication between FSTs
- **FST lifecycle**:
  - Spawn: Create new FST instance
  - Send: Send command to FST
  - Receive: Pattern match on events
  - Terminate: FST lifecycle end

### Example Concept (syntax TBD)
```
type CounterState = Count of i32;
type CounterCommand = Increment | Decrement | Get;
type CounterEvent = Changed of i32 | CurrentValue of i32;

// Transition function
fn counter_transition(state: CounterState, cmd: CounterCommand) -> (CounterState, CounterEvent) {
    let Count(n) = state;
    match cmd {
        Increment => (Count(n + 1), Changed(n + 1)),
        Decrement => (Count(n - 1), Changed(n - 1)),
        Get => (state, CurrentValue(n))
    }
}

// Usage
let counter = spawn_fst(counter_transition, Count(0));
send(counter, Increment);
match receive(counter) {
    Changed(n) => println("Counter: {}", n)
}
```

### Deliverables
- FST type system and semantics
- Channel types for communication
- FST runtime (single-node first)
- Testing framework for FSTs

### Reference Languages
- Erlang/Elixir: OTP gen_server, gen_statem
- Elm: Update function in Elm Architecture
- Pony: Actors with behaviors
- Akka: Actor model

---

## Phase 4: Ports & Effect System

**Goal**: Enforce clean architecture - separate pure logic from effectful operations.

### Features
- **Port definitions**: Declare boundaries to external world
  - Input ports (receive data from outside)
  - Output ports (send data to outside)
- **Effect types**: Track which functions perform effects
- **Pure FST guarantee**: Transition functions cannot perform I/O
- **Adapter implementations**: Connect ports to actual I/O
- **Dependency injection**: Ports are injected, making FSTs testable

### Example Concept
```
// Port definition
port HttpPort {
    fn request(url: String) -> Response;
}

port DatabasePort {
    fn query(sql: String) -> Vec<Row, max=1000>;
}

// FST uses ports but doesn't implement them
fst MyService {
    ports: {
        http: HttpPort,
        db: DatabasePort
    },

    transition: fn(state, command) -> (state, event) {
        // Can reference ports but execution is deferred
        // Actual I/O happens outside FST
    }
}

// Adapter provides real implementation
adapter HttpAdapter implements HttpPort {
    fn request(url: String) -> Response {
        // Actual HTTP call here
    }
}
```

### Deliverables
- Port syntax and type system
- Effect tracking in type system
- Adapter mechanism
- Testing with mock adapters

### Reference Languages
- Elm: Ports for JS interop
- Koka, Eff: Algebraic effect systems
- Ur/Web: Effect isolation
- Hexagonal architecture patterns in various languages

---

## Phase 5: Distribution & Ownership

**Goal**: Distributed-first semantics with explicit ownership and consistency.

### Features
- **Ownership type modifiers**:
  - `Local<T>`: Data exists only on this node
  - `RemoteView<T>`: Read-only view of remote data
  - `Replicated<T>`: Locally-owned, replicated to other nodes
  - `Distributed<T>`: Can be modified independently, requires reconciliation
- **Network-transparent FST addressing**: Send to FST regardless of location
- **Consistency primitives**:
  - CRDT types (LWW-Register, G-Counter, PN-Counter, OR-Set, etc.)
  - Consensus helpers (for strong consistency when needed)
  - Application-defined merge functions
- **Serialization boundaries**: Types must be serializable to cross network
- **Location awareness**: Query where data/FSTs live
- **Migration**: Move FSTs between nodes

### Example Concepts
```
type Counter = Distributed<GCounter>;  // CRDT counter

let local_state: Local<MyState> = ...;  // Never leaves this node
let remote_view: RemoteView<OtherState> = fetch_remote(...);  // Can read, not write
let replicated: Replicated<Config> = ...;  // Written locally, replicated

// FST can live anywhere
let remote_fst: FstAddr<MyFst> = spawn_on(node_id, ...);
send(remote_fst, command);  // Works regardless of location
```

### Deliverables
- Ownership type system
- Serialization framework
- Distribution runtime
- CRDT library
- Consensus protocol integration (Raft or similar)
- Network layer

### Reference Languages
- Unison: Content-addressed, distributed-first
- Pony: Reference capabilities for safety
- Distributed Erlang: Location transparency
- CRDT libraries: Automerge, Yjs
- Orleans: Virtual actors distributed across cluster

---

## Sequencing Rationale

1. **Phase 1 before 2**: Need basic type system before adding collection constraints
2. **Phase 2 before 3**: FSTs need collections for state management
3. **Phase 3 before 4**: Need FSTs to exist before adding effect system
4. **Phase 4 before 5**: Need clear effect boundaries before distribution (network I/O is an effect)
5. **Phase 5 is complex**: Builds on everything - types, FSTs, effects, serialization

Each phase delivers value independently while building toward the full vision.
