# Zeno Programming Language

Zeno is a statically-typed programming language with a syntax inspired by Go and TypeScript, designed to be simple yet powerful. The compiler for Zeno is implemented in Rust (Edition 2024) and currently compiles Zeno code to Rust.

## Features
- **Go/TypeScript Inspired Syntax:** Aims for readability and modern development practices.
- **Static Typing:** With type inference for conciseness.
    - `let name = value;` (immutable, inferred)
    - `let name: type = value;` (immutable, explicit)
    - `mut name = value;` (mutable, inferred)
    - `mut name: type = value;` (mutable, explicit)
- **Optional Semicolons:** Semicolons at the end of statements are optional.
- **Basic Types:** `int`, `float`, `bool`, `string`.
- **Control Flow:** `if/else if/else`, `loop`, `while`, `for`.
- **Output:** `print()` and `println()` functions.
- **Comments:** `// single-line` and `/* multi-line */`.
- **Compilation Target:** Generates Rust code.

## Current Status
- Lexer: Implemented.
- Parser: Implemented, supports optional semicolons.
- Code Generator: Implemented, generates Rust code from the AST.
- Compiler Driver: In progress.
- Project uses Rust Edition 2024.

## Language Syntax Overview

**Variable Declarations:**
```zeno
// Immutable, type inferred
let message = "Hello, Zeno!"
let count = 100

// Mutable, explicit type
mut temperature: float = 25.5
mut is_active: bool = true

is_active = false
```

**Control Flow:**
```zeno
if count > 50 {
    println("Count is greater than 50")
} else if count == 50 {
    println("Count is exactly 50")
} else {
    println("Count is less than 50")
}

loop {
    println("Looping...")
    break // Exit the loop
}

mut i = 0
while i < 3 {
    print(i)
    i = i + 1
} // Output: 012

for let j = 0; j < 3; j = j + 1 {
    print(j)
} // Output: 012
```

**Output:**
```zeno
print("This prints on one line. ")
println("This prints on a new line.")
let name = "Zeno"
println("Hello, " + name + "!") // String concatenation (assuming + works for strings)
```

## Building (Placeholder)
Detailed build instructions will be added once the compiler driver (`src/main.rs`) is complete.
To build the compiler project itself (the Rust project):
```bash
# (Navigate to the zeno directory if not already there)
cargo build
```

## Contributing
Contributions are welcome! Please see `TODO.md` for areas where you can help.
(Further contribution guidelines to be added).
```
