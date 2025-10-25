# Syntax Decisions

This document captures syntax choices for the language.

## Overall Style: Rust-like Structure + ML-style Types

**Rationale**:
- Rust-like structure (braces, semicolons, keywords) is familiar and scales well
- ML-style ADT syntax is elegant and concise for algebraic types
- This hybrid gives us structured code with expressive type definitions

## Lexical Elements

### Keywords
```
fn      - Function definition
let     - Variable binding
type    - Type definition
match   - Pattern matching
if      - Conditional
else    - Conditional alternative
return  - Early return from function
true    - Boolean literal
false   - Boolean literal
of      - ADT variant separator (ML-style)
```

More keywords will be added for FSTs, ports, effects in later phases.

### Identifiers
- **Values/functions**: Start with lowercase or underscore
  - `my_function`, `value`, `_private`
- **Types/constructors**: Start with uppercase
  - `Option`, `Result`, `MyType`, `Some`, `None`

### Literals
```
42              // Integer
3.14            // Float
true, false     // Boolean
"hello"         // String
()              // Unit
```

### Operators
```
// Arithmetic
+  -  *  /  %

// Comparison
==  !=  <  >  <=  >=

// Logical
&&  ||  !

// Assignment
=

// Type/Pattern
|   ->  =>  :  of

// Member access (future)
.
```

### Delimiters
```
{ }     // Blocks
( )     // Grouping, function calls, tuples
[ ]     // Arrays, indexing (future)
,       // Separator
;       // Statement terminator
:       // Type annotation
```

### Comments
```
// Line comment

/*
   Block comment
   (potentially)
*/
```

## Syntax by Feature

### Type Definitions (ML-style)

**Sum types (enums/tagged unions)**:
```
type Option<T> = Some of T | None;

type Result<T, E> = Ok of T | Err of E;

type Color = Red | Green | Blue;

type Tree<T> =
    | Leaf of T
    | Node of (Tree<T>, T, Tree<T>);
```

**Product types (structs/records)** - TBD, options:
```
// Option A: ML-style records
type Point = { x: f32, y: f32 };

// Option B: Rust-style structs
struct Point {
    x: f32,
    y: f32
}
```

**Type aliases**:
```
type UserId = i32;
```

### Function Definitions (Rust-like)

```
fn add(x: i32, y: i32) -> i32 {
    x + y
}

fn factorial(n: i32) -> i32 {
    if n <= 1 {
        1
    } else {
        n * factorial(n - 1)
    }
}

fn apply<A, B>(x: A, f: fn(A) -> B) -> B {
    f(x)
}
```

### Let Bindings

```
let x = 42;
let y: i32 = 100;
let f = fn(x) -> x + 1;

// Destructuring (future)
let (a, b) = (1, 2);
let Point { x, y } = point;
```

### Match Expressions (Rust-like)

```
match option {
    Some(x) => x,
    None => 0
}

match result {
    Ok(value) => println("Success: {}", value),
    Err(e) => println("Error: {}", e)
}

// Nested patterns
match tree {
    Leaf(x) => x,
    Node(left, value, right) => {
        let l = sum(left);
        let r = sum(right);
        l + value + r
    }
}
```

### If/Else Expressions

```
if x > 0 {
    "positive"
} else if x < 0 {
    "negative"
} else {
    "zero"
}

// If is an expression
let sign = if x >= 0 { 1 } else { -1 };
```

### Expressions

```
// Arithmetic
2 + 3 * 4

// Comparison
x == y && z > 10

// Function calls
factorial(5)
map(option, fn(x) -> x * 2)

// Blocks are expressions (last expression is the value)
let result = {
    let x = expensive_computation();
    let y = transform(x);
    y + 1  // No semicolon - this is the block's value
};
```

### Type Annotations

```
fn process(x: i32) -> Option<String> { ... }

let values: [i32; 10] = [0; 10];

let f: fn(i32) -> i32 = factorial;
```

## Precedence and Associativity

From highest to lowest precedence:

1. Function calls, member access (future): `f(x)`, `obj.field`
2. Unary operators: `!`, `-` (negation)
3. Multiplicative: `*`, `/`, `%` (left-associative)
4. Additive: `+`, `-` (left-associative)
5. Comparison: `<`, `>`, `<=`, `>=`
6. Equality: `==`, `!=`
7. Logical AND: `&&` (left-associative)
8. Logical OR: `||` (left-associative)
9. Assignment: `=` (right-associative, future for mutation)

Parentheses override precedence as usual.

## Deferred Decisions

### FST Syntax
Not yet decided. Options include:
- Keyword-based declaration (`fst Counter { ... }`)
- Type-based (FST is just a type, transition is a function)
- Module-based (FSTs are modules with required structure)

Will be determined after Phase 1 implementation provides experience with the language feel.

### Product Types (Structs/Records)
Haven't decided between ML-style `{ }` records vs Rust-style `struct` keyword.

### Mutation
Phase 1 is immutable only. If mutation is added later, syntax TBD:
- Rust-style `mut` keyword?
- ML-style `ref` cells?
- Something else?

### Module System
Not yet designed. Will need it eventually for organizing code.

### Effects/Ports Syntax
Deferred to Phase 4.

### Distribution Syntax
Deferred to Phase 5.

## Example Phase 1 Program

Putting it all together:

```
// Type definitions
type Option<T> = Some of T | None;

type List<T> =
    | Nil
    | Cons of (T, List<T>);

// Helper function
fn is_empty<T>(list: List<T>) -> bool {
    match list {
        Nil => true,
        Cons(_, _) => false
    }
}

// Map over list
fn map<A, B>(list: List<A>, f: fn(A) -> B) -> List<B> {
    match list {
        Nil => Nil,
        Cons(head, tail) => {
            let new_head = f(head);
            let new_tail = map(tail, f);
            Cons(new_head, new_tail)
        }
    }
}

// Length of list
fn length<T>(list: List<T>) -> i32 {
    match list {
        Nil => 0,
        Cons(_, tail) => 1 + length(tail)
    }
}

// Main entry point
fn main() -> () {
    let numbers = Cons(1, Cons(2, Cons(3, Nil)));
    let doubled = map(numbers, fn(x) -> x * 2);
    let len = length(doubled);

    println("Length: {}", len);  // Assuming println is built-in
}
```

This gives a taste of what Phase 1 programs will look like.
