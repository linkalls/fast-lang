# Zeno Programming Language (Go Implementation)

Zeno is a statically-typed programming language with a syntax inspired by Go and TypeScript, designed to be simple yet powerful. This Go implementation of the Zeno compiler currently compiles Zeno code to Go.

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
- **Compilation Target:** Generates Go code.

## Current Status
- Lexer: Implemented.
- Parser: Implemented, supports optional semicolons.
- Code Generator: Implemented, generates Go code from the AST.
- Compiler Driver: Implemented.
- Project uses Go 1.21+.

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
println("Hello, " + name + "!") // String concatenation
```

## Using the Zeno Compiler (CLI)

To use the Zeno compiler, you first need to build it from source.

### Building the Compiler
1.  Navigate to the root directory of the Zeno project (the `zeno-go/` directory).
2.  Run the following Go command:
    ```bash
    go build -o zeno ./cmd/zeno-compiler
    ```
3.  The compiler executable will be located at `./zeno`.

### Command-Line Interface
The basic command structure is:
```bash
./zeno <SOURCE_FILE.zeno> [OPTIONS]
```

**Common Operations & Examples:**

1.  **Compile a Zeno file to view the generated Go code:**
    This will create `output.go` (or `<SOURCE_FILE>.go` by default if `-o` is not used) with the translated Go code.
    ```bash
    ./zeno examples/hello.zeno --output-go-file output.go --keep-go
    # Or to use the default .go output name (e.g., examples/hello.go):
    ./zeno examples/hello.zeno --keep-go
    ```

2.  **Compile a Zeno file directly to an executable:**
    This generates the Go code, compiles it using `go build`, and creates an executable (e.g., `my_program`).
    ```bash
    ./zeno examples/variables.zeno --compile --output-executable-file my_program
    ```
    If `--output-executable-file` is omitted, the executable will have the same name as the source file (without extension, e.g., `examples/variables`).

3.  **Compile and immediately run a Zeno file:**
    This is useful for quick testing. The intermediate Go file is deleted by default unless `--keep-go` is specified.
    ```bash
    ./zeno examples/controlflow.zeno --compile --run
    ```

**Important Flags:**
-   `<source_file>`: (Required) Path to the Zeno source file (e.g., `examples/hello.zeno`).
-   `--output-go-file <path>`, `-o <path>`: Specifies the output file for the generated Go code.
-   `--output-executable-file <path>`, `-O <path>`: Specifies the output file name for the compiled executable.
-   `--compile`, `-c`: Compiles the generated Go code to an executable using `go build`.
-   `--run`, `-r`: Runs the compiled executable. Requires `--compile`.
-   `--keep-go`: Prevents the deletion of the intermediate `.go` file after compilation.
-   `--help`: Displays help information about the CLI arguments.

## Contributing
Contributions are welcome! Please see `TODO.md` for areas where you can help.
(Further contribution guidelines to be added).
