# TODO for Zeno Language (Go Implementation)

This file lists planned features and improvements for the Zeno programming language Go implementation and its compiler.

## Recently Completed ✅

- [x] **TypeScript-style Import System:** `import {println} from "std/fmt"` syntax
- [x] **Import Statement Parsing:** Full parsing and validation of import statements
- [x] **Unused Variable Detection:** Compile-time detection with helpful error messages
- [x] **Import Validation:** Ensures functions are imported before use
- [x] **Multilingual Error Messages:** Japanese error messages with `-jp` flag
- [x] **Binary Expressions:** Mathematical and comparison operators
- [x] **Type Annotations:** Optional type annotations support
- [x] **Generator Restructure:** Struct-based generator with scope tracking
- [x] **Variable Usage Analysis:** Two-pass compilation for proper validation
    - [x] **Print/Println as Functions:** Changed print/println from keywords to regular functions provided by the `std/fmt` module. This involved removing `PRINT`/`PRINTLN` tokens, `ast.PrintStatement`, and related parser/generator logic. All print operations now require importing from `std/fmt`.
- [x] **std/io Module:** File I/O operations with readFile/writeFile functions
- [x] **String Escape Sequences:** Proper handling of \n, \t, \", \\ in string literals
- [x] **Public Function Declarations:** `pub fn` keyword for function visibility
- [x] **User-defined Module System:** Relative path imports (`./`, `../`) with proper path resolution
- [x] **Unused Function Detection:** Compile-time detection (excludes main and public functions)
- [x] **Enhanced CLI Interface:** `run` and `compile` subcommands with improved error handling
- [x] **Type System Foundation:** Complete type system with symbol table and type inference
- [x] **Type-based Conditional Generation:** Smart boolean conversion based on variable types
  - [x] Bool types: Generate as-is (`if y {`)
  - [x] Int types: Generate with `!= 0` conversion (`if (x != 0) {`)
  - [x] String types: Generate with `!= ""` conversion (`if (z != "") {`)
  - [x] Float types: Generate with `!= 0.0` conversion
- [x] **Consistent Main Function Generation:** Always generate main function wrapper for all programs
- [x] **std/io Module Enhancement:** Added `remove` and `pwd` functions.
- [x] **Parser Bug Fix: `return` statement handling:** Resolved issues related to token consumption/state recovery after `return` statements in various block contexts, ensuring robust parsing. (Verified via `test_return_statements.zeno` and `test_std_io_extended.zeno`)

## High Priority 🚀

## High Priority 🚀

- [x] **Function Definitions and Calls:**
    - [x] Syntax for defining functions with parameters and return types
    - [x] Parsing function definitions and calls
    - [x] Code generation for user-defined functions
    - [x] Return statements and return type validation
    - [x] Public function declarations with `pub` keyword
    - [x] Function visibility in generated Go code

- [x] **Type System Implementation:**
    - [x] Basic type system with symbol table
    - [x] Type inference from expressions  
    - [x] Variable type registration during let declarations
    - [x] Type-aware conditional code generation

- [ ] **Control Flow Statements:**
    - [x] if statements with type-aware boolean conversion
    - [ ] if/else if/else statements (full implementation)
    - [ ] while loops (Basic implementation might exist, needs verification for completeness)
    - [ ] loop statements with break/continue (Needs verification for completeness)
    - [ ] for loops (Needs design and implementation)
    - [ ] if/else expressions (Allow `if/else` to return values and be used in expressions, e.g., `let x = if cond {1} else {2}`)

- [ ] **Assignment and Mutation:**
    - [ ] Assignment statements (`x = value`) - No semicolons required
    - [ ] Type checking for assignments (Ensuring type compatibility on re-assignment)

**Language Design Note:** Zeno follows a semicolon-free syntax design similar to languages like Python and Ruby. All statements are terminated by newlines rather than semicolons for cleaner, more readable code. The language uses a unified `let`-only variable declaration system for simplicity and consistency.

## Medium Priority 📋

- [ ] **Extended Standard Library:**
    - [x] std/io module for file operations (readFile/writeFile, remove, pwd functions)
    - [ ] std/string module for string manipulation
    - [ ] std/math module for mathematical functions
    - [ ] std/collections module for data structures
    - [ ] **(High Priority)** std/json: JSON parsing (e.g., `parse`) and stringification (e.g., `stringify`).
    - [ ] **(High Priority)** std/httpserver: HTTP server framework (Hono-style API: routing, request/response handling).
    - [ ] **(Medium Priority)** std/time: Time-related functionalities (e.g., getting current time, formatting, parsing).
    - [ ] **(Medium Priority)** std/http: HTTP client functionalities (e.g., making GET, POST requests).

- [ ] **Enhanced Type System:**
    - [x] Type inference improvements (basic type inference from expressions, function call return types, unary expressions)
    - [x] Symbol table implementation for type tracking
    - [x] Type-based code generation for boolean contexts
    - [ ] Better type error messages
    - [ ] Optional types and null safety
    - [ ] Generic types (basic implementation)

# Removed Parser Bug item from here as it's now resolved.

- [ ] **Basic Data Structures:**
    - [ ] Arrays (fixed-size or dynamic)
    - [ ] Structs or record types
    - [ ] Basic collections (lists, maps)

- [ ] **Comments Support:**
    - [ ] Single-line comments (`//`)
    - [ ] Multi-line comments (`/* */`)
    - [ ] Documentation comments

## Low Priority 🔮

- [x] **Advanced Language Features:**
    - [x] Module system and namespacing (user-defined modules)
    - [ ] Pattern matching with `match` expressions
    - [ ] Interfaces/traits for type contracts
    - [ ] Concurrency primitives (goroutine-like)
    - [ ] Channels for communication

- [ ] **Performance Optimizations:**
    - [x] Dead code elimination (unused function detection)
    - [ ] Constant folding
    - [ ] Better memory management in generated Go code
    - [ ] Compile-time optimizations

## Development and Tooling 🛠️

- [ ] **Enhanced Error Reporting:**
    - [ ] Display source code snippets with error locations
    - [ ] Suggest fixes for common errors
    - [ ] Better error message formatting

- [ ] **Testing Infrastructure:**
    - [ ] Comprehensive unit tests for all components
    - [ ] Integration tests for full compilation pipeline
    - [ ] Test cases for error conditions and edge cases
    - [ ] Performance benchmarks

- [ ] **Development Tools:**
    - [ ] Language Server Protocol (LSP) implementation
    - [ ] VS Code extension
    - [ ] Syntax highlighting support
    - [ ] Auto-completion and diagnostics

- [ ] **Build System:**
    - [ ] Package manager for dependencies
    - [ ] Project templates and scaffolding
    - [ ] Build configuration system

## Infrastructure 🏗️

- [ ] **CI/CD Pipeline:**
    - [ ] Automated testing on multiple platforms
    - [ ] Automated releases and distribution
    - [ ] Code quality checks

- [ ] **Cross-platform Support:**
    - [ ] Windows, macOS, and Linux compatibility
    - [ ] Platform-specific optimizations

- [ ] **Documentation:**
    - [ ] Complete API documentation
    - [ ] Language specification document
    - [ ] Comprehensive tutorials and examples

## Long-term Vision 🎯

- [ ] **Self-hosting:**
    - [ ] Implement the Zeno compiler in Zeno itself

- [ ] **Alternative Backends:**
    - [ ] WebAssembly (WASM) target
    - [ ] Native code generation
    - [ ] LLVM backend

- [ ] **Advanced IDE Integration:**
    - [ ] IntelliJ/GoLand plugin
    - [ ] Vim/Neovim support
    - [ ] Emacs mode

## Migration Progress from Rust Implementation 🔄

- [x] **Core Lexer:** Token definitions and lexical analysis
- [x] **Core Parser:** AST definitions and parsing logic  
- [x] **Core Generator:** Code generation (Go target instead of Rust)
- [x] **CLI Interface:** Command-line argument parsing and main driver
- [x] **Import System:** TypeScript-style import syntax
- [x] **Error Handling:** Multilingual error messages
- [x] **Type System:** Symbol table and type inference implementation
- [x] **Code Generation:** Smart boolean conversion based on types
- [x] **Main Function Generation:** Consistent main function wrapper
- [ ] **Test Suite Migration:** Adapt existing Rust tests to Go
- [ ] **Example Programs:** Ensure all examples work with Go implementation
- [ ] **Feature Parity:** Complete remaining features from Rust version
