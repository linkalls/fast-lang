# TODO for Zeno Language

This file lists planned features and improvements for the Zeno programming language and its compiler.

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
    - [ ] String manipulation functions.
    - [ ] Math functions.
    - [ ] Basic data structure operations.
- [ ] **Code Generation Optimizations:**
    - [ ] Explore optimizations in the generated Rust code.
- [ ] **Type System Enhancements:**
    - [ ] More robust type checking and inference.
    - [ ] Support for generic types.
    - [ ] Union types (if not fully covered by enums + pattern matching).
- [ ] **Build System Improvements:**
    - [ ] More sophisticated build options for the Zeno compiler itself.
- [ ] **Debugger Support (Long Term):**
    - [ ] Investigate options for debugging Zeno code.
- [ ] **Package Manager (Long Term):**
    - [ ] A package manager for Zeno libraries.

## Documentation
- [ ] **Comprehensive Language Guide:**
    - [ ] Detailed documentation for all language features.
- [ ] **Compiler Internals Documentation:**
    - [ ] Document the lexer, parser, and code generator design.
