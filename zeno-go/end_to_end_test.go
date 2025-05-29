package main

import (
	"fmt"
	"strings"

	"github.com/linkalls/zeno-lang/generator"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

func main() {
	fmt.Println("=== Zeno to Go Compiler End-to-End Test ===")

	// Test case 1: Simple binary expression
	testZenoToGo("let result = 5 + 3;", "Binary expression test")

	// Test case 2: Multiple operations
	testZenoToGo("let calculation = 10 * 2 + 5;", "Multiple operations test")

	// Test case 3: With type annotation
	testZenoToGo("let value: int = 42;", "Type annotation test")

	// Test case 4: Mutable variable
	testZenoToGo("mut counter = 0;", "Mutable variable test")

	// Test case 5: String literal
	testZenoToGo("let message = \"Hello, World!\";", "String literal test")
}

func testZenoToGo(zenoCode string, testName string) {
	fmt.Printf("\n--- %s ---\n", testName)
	fmt.Printf("Zeno input: %s\n", zenoCode)

	// Step 1: Lexical analysis
	l := lexer.New(zenoCode)

	// Step 2: Parse to AST
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}

	// Step 3: Generate Go code
	g := generator.New()
	goCode := g.GenerateProgram(program)

	// Clean up the output for better display
	goCode = strings.TrimSpace(goCode)

	fmt.Printf("Go output: %s\n", goCode)
	fmt.Println("✅ Success!")
}
