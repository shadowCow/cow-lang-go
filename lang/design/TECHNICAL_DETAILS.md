# Technical Details

This document captures specific technical decisions about the language design.

## Communication Model: Typed Channels (Point-to-Point)

**Decision**: FSTs communicate via typed channels, not pub/sub event buses.

### Rationale
- **Type safety**: Channel endpoints have types - FST A knows exactly what FST B expects
- **Explicit dependencies**: Communication graph is clear in code
- **Easier reasoning**: Point-to-point is simpler to understand than pub/sub
- **Resource bounds**: Know maximum number of connections
- **Distribution**: Easier to route messages across network with explicit endpoints

### Implications
```
// Conceptual example
type Counter = fst<CounterCommand, CounterEvent>;

let counter: Counter = spawn_fst(...);

// Type-safe send - compiler knows what commands Counter accepts
send(counter, Increment);  // OK
send(counter, "invalid");   // Compile error

// Type-safe receive - compiler knows what events Counter emits
match receive(counter) {
    Changed(n) => ...,
    CurrentValue(n) => ...
}
```

### Trade-offs
**Pros**:
- Strong typing end-to-end
- Clear communication patterns
- Easier to trace message flow
- Better for distributed tracing/debugging

**Cons**:
- More coupling than pub/sub
- Harder to add new consumers (must wire up channels)
- Pattern matching on message types still needed

**Potential future addition**: Pub/sub as a library built on top of channels, not a core primitive.

---

## Collection Bounds: Both Compile-Time and Runtime

**Decision**: Support both compile-time sized arrays and runtime-parameterized bounded collections.

### Compile-Time Sized Arrays
```
let buffer: [u8; 1024] = [0; 1024];
```

**Use cases**:
- Known fixed sizes (protocol headers, buffers)
- Embedded systems with static allocation
- Maximum performance (no bounds checking needed)

**Implementation**: Part of type system, size is a const parameter.

### Runtime-Parameterized Bounded Collections
```
let capacity = read_config();  // Runtime value
let queue = BoundedQueue::new(capacity);
```

**Use cases**:
- Configuration-driven sizing
- Different limits for different deployments
- Adaptive sizing within bounds

**Implementation**:
- Capacity is a runtime value
- Operations return `Result` - can fail if full
- Size is tracked, but not part of type
- May use const generics with max bound: `BoundedVec<T, const MAX: usize>`

### Trade-offs

| Aspect | Compile-Time | Runtime-Bounded |
|--------|-------------|-----------------|
| Safety | Perfect - overflow impossible | Good - overflow returns error |
| Flexibility | Low - size in code | High - size from config |
| Performance | Best - no checks | Good - single check |
| Memory | Fixed allocation | Fixed allocation |
| Use case | Protocol buffers | Configurable limits |

### Design Principle
**Use the most restrictive option that works**:
1. If size is known at compile time → use `[T; N]`
2. If size varies but has a maximum → use `BoundedVec<T, MAX>`
3. For truly unbounded data → use `Stream<T>` with windowing

---

## Consistency Model: Application-Defined

**Decision**: Language provides primitives for different consistency models; programmer chooses per data structure.

### Rationale
Different parts of an application have different consistency needs:
- **Eventually consistent**: Analytics, metrics, caches - use CRDTs
- **Strongly consistent**: Financial transactions, user authentication - use consensus
- **Local only**: Temporary computations - no distribution needed

No single model fits all use cases.

### Provided Primitives

#### 1. Eventually Consistent (CRDTs)
```
type Counter = Distributed<GCounter>;  // Grow-only counter
type UserSet = Distributed<ORSet<UserId>>;  // Observed-remove set

let counter = GCounter::new();
counter.increment(node_id, 5);  // Local increment
// Automatically merged with other replicas
```

**CRDT types to provide**:
- Counters: G-Counter, PN-Counter
- Registers: LWW-Register, MV-Register
- Sets: G-Set, OR-Set
- Maps: OR-Map, LWW-Map
- Sequences: RGA, WOOT

#### 2. Strong Consistency (Consensus)
```
type AccountBalance = Consistent<i64>;

let balance = ConsistentRegister::new(100);
match balance.compare_and_swap(100, 50) {
    Ok(_) => println("Withdrawal succeeded"),
    Err(current) => println("Balance changed, now {}", current)
}
```

**Primitives**:
- Linearizable registers
- Distributed locks
- Leader election
- Consensus-based replication (Raft)

#### 3. Local Only
```
type Cache = Local<HashMap<Key, Value>>;

let cache: Local<Cache> = Local::new(HashMap::new());
// This data never leaves this node
```

#### 4. Remote View (Read-Only)
```
type Config = RemoteView<AppConfig>;

let config: RemoteView<AppConfig> = fetch_from_config_server();
// Can read, cannot modify
// Updates come from authoritative source
```

#### 5. Replicated (Local Write, Remote Read)
```
type Settings = Replicated<UserSettings>;

let settings: Replicated<UserSettings> = Replicated::new(...);
settings.update(...);  // Write locally
// Automatically propagated to replicas
// Only this node can write, others read-only
```

### Choosing Consistency Level

**Decision matrix**:

| Need | Use | Example |
|------|-----|---------|
| Exact count, ordering matters | Strong consistency | Account balance |
| Approximate count okay | CRDT Counter | Page views |
| Add/remove items, order doesn't matter | CRDT Set | User tags |
| Last write wins | LWW Register | User profile |
| Concurrent edits must be preserved | MV Register or custom merge | Collaborative editing |
| No distribution needed | Local | Temporary cache |
| Read from remote, never write | RemoteView | Configuration |
| Write locally, replicate read-only | Replicated | User preferences |

### Implementation Strategy

**Type system tracks consistency**:
```
type Distribution<T> =
    | Local of T
    | RemoteView of T
    | Replicated of T
    | Distributed of CRDT<T>
    | Consistent of Consensus<T>;
```

**Serialization boundary**: Only `Replicated`, `Distributed`, and `Consistent` can cross network.

**Merge functions**: For custom CRDTs, programmer provides merge:
```
fn merge(a: MyType, b: MyType) -> MyType {
    // Application-specific merge logic
    // Must be commutative, associative, idempotent
}
```

**Compiler checks**: Ensure merge functions have correct properties (when possible).

---

## Memory Model for Embedded/Real-Time

**Decision**: Predictable memory usage with compile-time or initialization-time allocation.

### No Garbage Collection
- Deterministic performance required for real-time
- Predictable memory usage for embedded
- No GC pauses

### Allocation Strategy

#### Static Allocation (Compile-Time)
```
static BUFFER: [u8; 4096] = [0; 4096];
```

#### Arena Allocation (Initialization-Time)
```
let arena = Arena::with_capacity(1024 * 1024);  // 1MB arena
let obj1 = arena.alloc(MyStruct { ... });
let obj2 = arena.alloc(AnotherStruct { ... });
// All deallocated when arena is dropped
```

#### Region-Based Memory
```
fn process_event() {
    let region = Region::new(4096);  // Temporary region
    let temp_data = region.alloc(...);
    // Use temp_data
    // region automatically freed at end of scope
}
```

### Ownership and Lifetimes

**Inspired by Rust, but adapted**:
- Ownership tracking for safety
- No borrowing (simplifies distributed semantics)
- Move semantics by default
- Clone when needed (explicit)

**Different from Rust**:
- No `&` borrows (complicates distribution)
- Channels transfer ownership (like Go)
- Simpler model, less powerful locally, better for distribution

### FST Memory Model

Each FST has:
- **State**: Owned by FST, bounded size
- **Channel buffers**: Bounded capacity
- **Temporary allocations**: Per-transition region

**Memory bounds are known**:
```
fst Counter {
    state: CounterState,  // Size: known
    command_buffer: BoundedQueue<Command, 100>,  // Size: 100 * sizeof(Command)
    event_buffer: BoundedQueue<Event, 100>,      // Size: 100 * sizeof(Event)
}

// Total memory: sizeof(CounterState) + 100 * sizeof(Command) + 100 * sizeof(Event)
```

### Streams and Unbounded Data

**Streams are processed incrementally**:
```
fn process(stream: Stream<Event>) -> Stream<Result> {
    stream
        .window_tumbling(size: 1000)  // Process 1000 at a time
        .map(|batch| {
            // batch is bounded: Vec<Event, 1000>
            aggregate(batch)
        })
}
```

**Memory used**: Window size only, not entire stream.

---

## Serialization and Network Boundaries

### Serialization Requirements

Types that cross network boundaries must be serializable:
```
type NetworkSafe = Serializable;

// Automatically derivable for simple types
type UserId = i64;  // Serializable

type User = {
    id: UserId,
    name: String,
    age: i32
};  // Serializable (all fields are)

type LocalCache<T> = Local<HashMap<i64, T>>;  // NOT serializable (Local)
```

### Binary Format

**Use case driven**:
- **Internal FST communication**: Efficient binary (MessagePack, Cap'n Proto, or custom)
- **External APIs**: JSON, Protobuf, or application-defined
- **Persistence**: Same as internal communication

**Versioning**: Schema evolution for distributed systems.

### Network Transparency

FSTs can be addressed regardless of location:
```
let fst_addr: FstAddr<MyFst> = ...;  // Could be local or remote
send(fst_addr, command);  // Works either way
```

**Implementation**:
- Local: Direct memory transfer
- Remote: Serialize, send over network, deserialize
- Application doesn't know/care (except for error handling)

---

## Error Handling

### Result Types
```
type Result<T, E> = Ok of T | Err of E;
```

**Use for**:
- Operations that can fail predictably
- Bounded collections overflow
- Network operations
- File I/O (via ports)

### Panic (Abort)
**Use for**:
- Unrecoverable errors
- Programmer mistakes (assertions)
- Contract violations

**In distributed context**: Panic crashes FST, not entire system (supervision can restart).

### Effect Types (Future)
Track effects in type signatures:
```
fn read_file(path: String) -> Result<String, IoError> [IO]
//                                                    ^^^^
//                                                    Effect annotation
```

---

## Performance Characteristics

### Goals
1. **Predictable**: No GC pauses, bounded memory
2. **Efficient**: Zero-cost abstractions where possible
3. **Real-time capable**: Bounded latency for operations
4. **Distributed**: Network overhead is explicit

### FST Performance
- **Transition function**: Pure computation, fast
- **Message send**: Bounded time (channel buffer + optional serialization)
- **Message receive**: Bounded time (pattern matching)

### Collection Performance
- Array access: O(1)
- Bounded vector push: O(1) amortized, O(n) worst case (realloc)
- Stream windowing: O(window size)

### Network Performance
- Local FST: Memory speed
- Remote FST: Network latency + serialization
- No hidden network calls

---

## Testing Strategy

### Unit Testing
- Pure functions easy to test
- FST transitions are pure → easy to test
- Property-based testing for CRDTs

### Integration Testing
- Mock ports for testing FSTs in isolation
- Simulate network for distributed testing

### Property Testing
- CRDT properties (commutativity, etc.)
- Bounded collection invariants
- Type safety properties

---

## Open Questions

1. **FFI**: How to call C/Rust/other languages?
2. **Compilation target**: Native binary? WASM? Both?
3. **Runtime**: Minimal runtime or full actor runtime?
4. **Supervision**: Erlang-style supervision trees?
5. **Hot code loading**: Like Erlang? Or static deploys?
6. **Observability**: Built-in tracing? Metrics? Logging?

These will be resolved as implementation progresses.
