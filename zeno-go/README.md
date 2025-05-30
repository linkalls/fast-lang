# Zeno Programming Language (Go Implementation)

Zeno is a programming language with TypeScript-inspired import syntax, designed to be simple yet powerful. This Go implementation of the Zeno compiler translates Zeno code to Go with comprehensive error checking and validation.

## Features

- **TypeScript-style Import System**: `import {println} from "std/fmt"` syntax with module support
- **Module System**: User-defined module imports with relative path support (`./`, `../`)
- **Public Function Declarations**: `pub fn` keyword for public function visibility
- **Function Definitions and Calls**: Support for parameters, return types, and return statements
- **Unused Variable Detection**: Compile-time detection of unused variables with helpful error messages
- **Unused Function Detection**: Compile-time detection of unused functions (excludes main and public functions)
- **Import Validation**: Ensures functions are properly imported before use
- **Binary Expressions**: Mathematical operations (+, -, *, /, %) and comparison operators
- **Type Annotations**: Optional type annotations `let x: int = 42`
- **Multilingual Error Messages**: `-jp` flag for Japanese error messages alongside English
- **Variable Declarations**: `let` keyword for variable declarations
- **Enhanced CLI**: `run` and `compile` subcommands with improved error handling
- **Built-in Linter**: Static analysis for code quality and conventions.
- **Floating-Point Literals**: Support for numbers with decimal points (e.g., `3.14`).

## Current Implementation Status

✅ **Completed:**
- Import statement parsing and validation
- User-defined module imports with relative path support
- Public function declarations (`pub fn` keyword)
- Variable declarations (let)
- Function definitions and calls
- Return statements
- Binary expressions (arithmetic, comparison)
- Print statement conversion
- Unused variable detection
- Unused function detection (excludes main and public functions)
- Multilingual error messages (English/Japanese)
- Enhanced CLI with `run` and `compile` subcommands
- Lexical analysis (Lexer)
- AST construction (Parser)
- Go code generation (Generator)
- Standard Library: std/json module (parse, stringify)
- Floating-point literal parsing and generation

🔲 **Planned:**
- Control flow (if/else, while, loop)
- Extended type system
- Expanded standard library

## Language Syntax

### Import Statements
```zeno
import {println, print} from "std/fmt"
import {add, multiply} from "./math_utils"  // User-defined module
```

### Variable Declarations
```zeno
let x = 42           // Variable declaration
let y: int = 100     // With type annotation
let pi = 3.14        // Floating-point number
```

### Function Definitions
```zeno
// Private function (default)
fn helper(x: int): int {
    return x * 2
}

// Public function (accessible from other modules)
pub fn add(a: int, b: int): int {
    return a + b
}

import { println } from "std/fmt" // Assuming println is imported for this example
pub fn greet(name: string) {
    println("Hello, " + name)
}
```

### Module System
```zeno
// math_utils.zeno
pub fn add(a: int, b: int): int {
    return a + b
}

pub fn multiply(a: int, b: int): int {
    return a * b
}

// main.zeno
import {println} from "std/fmt"
import {add, multiply} from "./math_utils"

fn main() {
    let result = add(10, multiply(3, 4))
    println(result)
}
```

### Function Calls
```zeno
let result = add(10, 20)
greet("World")
```

### Main Function
```zeno
import { println } from "std/fmt" // Assuming println is imported

fn main() {
    // Program entry point
    println("Hello, World!")
}
```

### Binary Expressions
```zeno
let sum = 10 + 20
let product = 5 * 6
let comparison = x > y
```

### Printing to Console (using std/fmt)
Printing is handled by functions from the `std/fmt` module. These must be imported before use.
```zeno
print("Hello")       // Requires: import {print} from "std/fmt"
println("World")     // Requires: import {println} from "std/fmt"
```

## Example Program

### Basic Program
```zeno
import {println} from "std/fmt"

let x = 10
let y = 20
let result = x + y
println(result)
```

### Using User-defined Modules
```zeno
// math_utils.zeno
pub fn calculate(a: int, b: int): int {
    return (a + b) * 2
}

// main.zeno  
import {println} from "std/fmt"
import {calculate} from "./math_utils"

fn main() {
    let result = calculate(5, 10)
    println(result)  // Output: 30
}
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
import {println} from "std/fmt"

let x = 10
let unused = 42  // Error: Unused variables found: unused
let y = x + 5
println(y)
```

### Unused Function Detection
```zeno
import {println} from "std/fmt"

fn main() {
    println("Hello")
}

fn unused_helper() {  // Error: Unused functions found: unused_helper
    return 42
}

pub fn public_fn() {  // Public functions are never considered unused
    return "public"
}
```

### Import Validation
```zeno
// Error: println is not imported from std/fmt
let x = 10
println(x)  // Missing import statement
```

## Standard Library

Currently supported modules:

- `std/fmt`: `print`, `println` functions
- `std/io`: `readFile`, `writeFile`, `remove`, `pwd` functions
- `std/json`: JSON parsing (`parse`) and stringification (`stringify`) functions.

### std/io Module Usage

The `std/io` module provides simple and intuitive file I/O operations:

```zeno
import { println } from "std/fmt"
import { readFile, writeFile } from "std/io"

fn main() {
    // Write content to a file
    let content = "Hello, Zeno!\nThis is a test file."
    writeFile("example.txt", content)
    println("File written successfully!")
    
    // Read content from a file
    let fileContent = readFile("example.txt")
    println("File content:")
    println(fileContent)
    
    // Write structured data
    let jsonData = "{\"name\": \"Zeno\", \"version\": \"1.0\"}"
    writeFile("config.json", jsonData)
    
    let configData = readFile("config.json")
    println("Config: ", configData)
}
```

#### std/io Functions

- `writeFile(filename: string, content: string)`: Writes content to a file with automatic error handling
- `readFile(filename: string): string`: Reads file content and returns it as a string, returns empty string on error
- `remove(filename: string): bool`: Removes the specified file or empty directory. Returns `true` on success, `false` on failure.
- `pwd(): string`: Returns the current working directory as an absolute path. Returns an empty string on failure.

### std/json Module Usage

The `std/json` module provides functions to parse JSON strings into Zeno data structures and stringify Zeno data structures into JSON strings.

```zeno
import { println, print } from "std/fmt"
import { parse, stringify } from "std/json"

fn main() {
    let jsonString = "{\"name\": \"Zeno\", \"version\": 0.2, \"active\": true}"
    println("Original JSON string: " + jsonString)

    let parsedData = parse(jsonString)
    // At present, 'parsedData' is of type 'any'. Interacting with its structure
    // (e.g., accessing map keys or array elements) will depend on future
    // Zeno language features for type inspection and manipulation of 'any'.

    let reStringified = stringify(parsedData)
    print("Re-stringified JSON: ")
    println(reStringified)

    let zenoData = "a simple string" // Example of a Zeno primitive
    let jsonFromZeno = stringify(zenoData)
    print("JSON from Zeno string 'a simple string': ")
    println(jsonFromZeno) // Expected: "\"a simple string\""
    
    let invalidJson = "{\"key\": value_not_string}" // Note: value_not_string needs to be a Zeno string for this to be a valid Zeno line
    let parsedError = parse(invalidJson)
    print("Result of parsing invalid JSON: ")
    println(stringify(parsedError)) // Expected: "null"
}
```

#### std/json Functions

- `parse(jsonString: string): any`: Parses a JSON string. Returns the parsed data as type `any` (representing a Zeno string, number, boolean, list, or map). Returns Zeno's `nil` equivalent (which stringifies to JSON `null`) on parsing error.
- `stringify(value: any): string`: Converts a Zeno value (of type `any`, expected to be composed of primitives, lists, or maps) into a JSON string. Returns an empty string `""` on stringification error.

## Using the Zeno Compiler

### Building the Compiler

```bash
cd zeno-go
go build ./cmd/zeno
```

### Enhanced CLI Usage

```bash
# Run a Zeno file (compile and execute)
./zeno run example.zeno

# Compile a Zeno file to Go (output to stdout)
./zeno compile example.zeno

# Compile a Zeno file to binary (output to file)
./zeno build example.zeno

# Show Japanese error messages as well
./zeno run -jp example.zeno
./zeno compile -jp example.zeno

# Show help
./zeno --help
./zeno run --help
./zeno compile --help
```

## Linting Zeno Code

Zeno includes a built-in linter to help you identify potential issues and enforce coding conventions in your Zeno source files.

### Usage

You can run the linter using the `lint` subcommand:

-   **Lint a single file:**
    ```bash
    ./zeno lint path/to/yourfile.zeno
    ```
-   **Lint all `.zeno` and `.zn` files in a directory (recursively):**
    ```bash
    ./zeno lint path/to/your_directory
    ```

The linter will print any issues found to the console in the format:
`filepath:line:column: [rule-name] message`

If any linting issues are found, the command will exit with a status code of 1. Otherwise, it will exit with 0.

### Supported Rules (Initial Set)

The linter currently checks for the following:

1.  **`unused-variable`**: Detects variables declared with `let` that are not used. (Rule L1)
2.  **`unused-function`**: Detects non-public functions (`fn`) that are defined but not used (excludes `main` function). (Rule L2)
3.  **`function-naming-convention`**: Ensures private functions (`fn`) are `lowerCamelCase` and public functions (`pub fn`) are `UpperCamelCase`. (Rule L3)
4.  **`variable-naming-convention`**: Ensures variables declared with `let` are in `lowerCamelCase` (ignores `_` identifier). (Rule L4)
5.  **`unused-import`**: Detects symbols imported from modules that are not used in the current file. (Rule L5)

*(Note: Line and column numbers in issue reports are currently placeholders (0:0) and will be improved with future parser enhancements to include positional information in AST nodes.)*

*(Future enhancements may include a configuration file to customize enabled rules and their parameters.)*

### Example Files

The project includes comprehensive example files in the `examples/` directory:

- `test_pub_functions.zeno` - Public function declarations test
- `test_unused_functions.zeno` - Unused function detection test
- `math_utils.zeno` - User-defined module with public functions
- `test_module_import.zeno` - Module import system test
- `test_simple.zeno` - Basic functionality test
- `test_import.zeno` - Standard library import test
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
2. Build compiler with `go build ./cmd/zeno`
3. Test with provided test files
4. Implement new features or improvements

### Bug Reports

Please report bugs and feature requests through GitHub Issues.
