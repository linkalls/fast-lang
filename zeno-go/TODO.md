# TODO for Zeno Language (Go Implementation)

This file lists planned features and improvements for the Zeno programming language Go implementation and its compiler.

## Language Features
- [ ] **Full Function Definitions:**
    - [ ] Syntax for defining functions with parameters and return types (`fn name(param: type) -> type { ... }`).
    - [ ] Parsing function definitions.
    - [ ] Code generation for function definitions and calls.
    - [ ] Return statements.
- [ ] **Modules/Namespacing:**
    - [ ] Simple module system for code organization.
    - [ ] Import/export functionality.
- [ ] **Basic Data Structures:**
    - [ ] Arrays (fixed-size or dynamic).
    - [ ] Structs or record types.
- [ ] **Concurrency Primitives (Inspired by Go):**
    - [ ] Lightweight "goroutine-like" tasks.
    - [ ] Channels for communication between tasks.
- [ ] **Interfaces/Traits (Inspired by TypeScript/Rust):**
    - [ ] Define contracts for types.
- [ ] **Pattern Matching (Basic):**
    - [ ] Simple `match` expressions.

## Compiler/Tooling Enhancements
- [ ] **Enhanced Error Reporting:**
    - [ ] Display source code snippets with error locations.
    - [ ] More descriptive and helpful error messages.
    - [ ] Suggest fixes for common errors.
- [ ] **Standard Library:**
    - [ ] Basic I/O operations beyond `print/println`.
    - [ ] File system operations.
    - [ ] String manipulation utilities.
- [ ] **Performance Optimizations:**
    - [ ] Better memory management in generated Go code.
    - [ ] Dead code elimination.
    - [ ] Constant folding and other compile-time optimizations.
- [ ] **Language Server Protocol (LSP):**
    - [ ] Syntax highlighting support.
    - [ ] Auto-completion.
    - [ ] Error diagnostics in real-time.
- [ ] **Package Manager/Build System:**
    - [ ] Dependency management.
    - [ ] Project templates and scaffolding.

## Code Quality and Testing
- [ ] **Comprehensive Test Coverage:**
    - [ ] Unit tests for lexer, parser, and generator.
    - [ ] Integration tests for the full compilation pipeline.
    - [ ] Test cases for error conditions and edge cases.
- [ ] **Benchmarking:**
    - [ ] Performance benchmarks for compiler components.
    - [ ] Comparison with other language implementations.
- [ ] **Documentation:**
    - [ ] Detailed API documentation for all modules.
    - [ ] Language specification document.
    - [ ] Tutorial and examples.

## Infrastructure
- [ ] **CI/CD Pipeline:**
    - [ ] Automated testing on multiple platforms.
    - [ ] Automated releases and distribution.
- [ ] **Cross-platform Support:**
    - [ ] Ensure compatibility with Windows, macOS, and Linux.
    - [ ] Platform-specific optimizations where necessary.

## Long-term Goals
- [ ] **Self-hosting:**
    - [ ] Implement the Zeno compiler in Zeno itself.
- [ ] **Alternative Backends:**
    - [ ] Support for compiling to other targets (WASM, native code).
- [ ] **IDE Integration:**
    - [ ] VS Code extension.
    - [ ] Integration with other popular editors.

## Migration from Rust Implementation
- [x] **Core Lexer:** Port token definitions and lexical analysis.
- [x] **Core Parser:** Port AST definitions and parsing logic.
- [x] **Core Generator:** Port code generation to Go instead of Rust.
- [x] **CLI Interface:** Port command-line argument parsing and main driver.
- [ ] **Test Suite Migration:** Adapt existing Rust tests to Go.
- [ ] **Example Programs:** Ensure all examples work with Go implementation.
