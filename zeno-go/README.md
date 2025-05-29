# Zeno Programming Language (Go Implementation)

Zeno is a programming language with TypeScript-inspired import syntax, designed to be simple yet powerful. This Go implementation of the Zeno compiler translates Zeno code to Go with comprehensive error checking and validation.

## Features

- **TypeScript-style Import System**: `import {println} from "std/fmt"` syntax
- **Unused Variable Detection**: Compile-time detection of unused variables with helpful error messages
- **Import Validation**: Ensures functions are properly imported before use
- **Binary Expressions**: Mathematical operations (+, -, *, /, %) and comparison operators
- **Type Annotations**: Optional type annotations `let x: int = 42;`
- **Multilingual Error Messages**: `-jp` flag for Japanese error messages alongside English
- **Variable Declarations**: `let` keyword for variable declarations

## Current Implementation Status

✅ **Completed:**
- Import statement parsing and validation
- Variable declarations (let)
- Binary expressions (arithmetic, comparison)
- Print statement conversion
- Unused variable detection
- Multilingual error messages (English/Japanese)
- Lexical analysis (Lexer)
- AST construction (Parser)
- Go code generation (Generator)

🔲 **Planned:**
- Function definitions and calls
- Control flow (if/else, while, loop)
- Mutable variables (mut)
- Extended type system
- Expanded standard library

## Language Syntax

### Import Statements
```zeno
import {println, print} from "std/fmt";
```

### Variable Declarations
```zeno
let x = 42;           // Variable declaration
let y: int = 100;     // With type annotation
```

### Binary Expressions
```zeno
let sum = 10 + 20;
let product = 5 * 6;
let comparison = x > y;
```

### Print Statements
```zeno
print("Hello");       // Requires: import {print} from "std/fmt";
println("World");     // Requires: import {println} from "std/fmt";
```

## Example Program

```zeno
import {println} from "std/fmt";

let x = 10;
let y = 20;
let result = x + y;
println(result);
```

Generated Go code:

```go
package main

import (
	"fmt"
)

func main() {
	var x = 10
	var y = 20
	var result = (x + y)
	fmt.Println(result)
}
```

## Error Detection

### Unused Variable Detection
```zeno
import {println} from "std/fmt";

let x = 10;
let unused = 42;  // Error: Unused variables found: unused
let y = x + 5;
println(y);
```

### Import Validation
```zeno
// Error: println is not imported from std/fmt
let x = 10;
println(x);  // Missing import statement
```

## Standard Library

Currently supported modules:

- `std/fmt`: `print`, `println` functions

## Using the Zeno Compiler

### Building the Compiler

```bash
cd zeno-go
go build ./cmd/zeno-compiler
```

### Basic Usage

```bash
# Compile a Zeno file
./zeno-compiler example.zeno

# Show Japanese error messages as well
./zeno-compiler -jp example.zeno

# Run internal tests (when no file is specified)
./zeno-compiler
```

### Test Files

The project includes several test files:

- `test_simple.zeno` - Basic functionality test
- `test_import.zeno` - Import statement test
- `test_unused.zeno` - Unused variable detection test
- `test_no_import.zeno` - Missing import error test

## Development Tools

Debug tools are also included:

- `debug_lexer.go` - For testing lexer functionality
- `debug_parser.go` - For testing parser functionality

## Contributing

Contributions are welcome! Please see `TODO.md` for areas where you can help.

### Development Process

1. Clone the project
2. Build compiler with `go build ./cmd/zeno-compiler`
3. Test with provided test files
4. Implement new features or improvements

### Bug Reports

Please report bugs and feature requests through GitHub Issues.
