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

## Using the Zeno Compiler (CLI)

To use the Zeno compiler, you first need to build it from source.

### Building the Compiler
1.  Navigate to the root directory of the Zeno project (the `zeno/` directory).
2.  Run the following Cargo command:
    ```bash
    cargo build --release
    ```
3.  The compiler executable will be located at `target/release/zeno` (or `target\release\zeno.exe` on Windows).

### Command-Line Interface
The basic command structure is:
```bash
./target/release/zeno <SOURCE_FILE.zeno> [OPTIONS]
```

**Common Operations & Examples:**

1.  **Compile a Zeno file to view the generated Rust code:**
    This will create `output.rs` (or `<SOURCE_FILE>.rs` by default if `-o` is not used) with the translated Rust code.
    ```bash
    ./target/release/zeno examples/hello.zeno --output-rust-file output.rs --keep-rs
    # Or to use the default .rs output name (e.g., examples/hello.rs):
    ./target/release/zeno examples/hello.zeno --keep-rs
    ```

2.  **Compile a Zeno file directly to an executable:**
    This generates the Rust code, compiles it using `rustc`, and creates an executable (e.g., `my_program`).
    ```bash
    ./target/release/zeno examples/variables.zeno --compile --output-executable-file my_program
    ```
    If `--output-executable-file` is omitted, the executable will have the same name as the source file (without extension, e.g., `examples/variables`).

3.  **Compile and immediately run a Zeno file:**
    This is useful for quick testing. The intermediate Rust file is deleted by default unless `--keep-rs` is specified.
    ```bash
    ./target/release/zeno examples/controlflow.zeno --compile --run
    ```

**Important Flags:**
-   `<source_file>`: (Required) Path to the Zeno source file (e.g., `examples/hello.zeno`).
-   `--output-rust-file <path>`, `-o <path>`: Specifies the output file for the generated Rust code.
-   `--output-executable-file <path>`, `-O <path>`: Specifies the output file name for the compiled executable.
-   `--compile`, `-c`: Compiles the generated Rust code to an executable using `rustc`.
-   `--run`, `-r`: Runs the compiled executable. Requires `--compile`.
-   `--keep-rs`: Prevents the deletion of the intermediate `.rs` file after compilation.
-   `--help`: Displays help information about the CLI arguments.

## Contributing
Contributions are welcome! Please see `TODO.md` for areas where you can help.
(Further contribution guidelines to be added).
```
