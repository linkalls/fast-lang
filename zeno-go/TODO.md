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
- [x] **Standard Library Definition:** Basic std/fmt module with print/println

## High Priority 🚀

- [ ] **Function Definitions and Calls:**
    - [ ] Syntax for defining functions with parameters and return types
    - [ ] Parsing function definitions and calls
    - [ ] Code generation for user-defined functions
    - [ ] Return statements and return type validation

- [ ] **Control Flow Statements:**
    - [ ] if/else if/else statements
    - [ ] while loops
    - [ ] loop statements with break/continue
    - [ ] for loops

- [ ] **Mutable Variables:**
    - [ ] `mut` keyword for mutable variable declarations
    - [ ] Assignment statements for mutable variables
    - [ ] Mutation validation and error checking

## Medium Priority 📋

- [ ] **Extended Standard Library:**
    - [ ] std/io module for file operations
    - [ ] std/string module for string manipulation
    - [ ] std/math module for mathematical functions
    - [ ] std/collections module for data structures

- [ ] **Enhanced Type System:**
    - [ ] Type inference improvements
    - [ ] Better type error messages
    - [ ] Optional types and null safety
    - [ ] Generic types (basic implementation)

- [ ] **Basic Data Structures:**
    - [ ] Arrays (fixed-size or dynamic)
    - [ ] Structs or record types
    - [ ] Basic collections (lists, maps)

- [ ] **Comments Support:**
    - [ ] Single-line comments (`//`)
    - [ ] Multi-line comments (`/* */`)
    - [ ] Documentation comments

## Low Priority 🔮

- [ ] **Advanced Language Features:**
    - [ ] Pattern matching with `match` expressions
    - [ ] Interfaces/traits for type contracts
    - [ ] Module system and namespacing
    - [ ] Concurrency primitives (goroutine-like)
    - [ ] Channels for communication

- [ ] **Performance Optimizations:**
    - [ ] Dead code elimination
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
- [ ] **Test Suite Migration:** Adapt existing Rust tests to Go
- [ ] **Example Programs:** Ensure all examples work with Go implementation
- [ ] **Feature Parity:** Complete remaining features from Rust version
