package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/linkalls/zeno-lang/generator"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

func main() {
	fmt.Println("=== Zeno to Go Compiler ===")
	
	if len(os.Args) < 2 {
		// No file provided, run demo tests
		runDemoTests()
		return
	}
	
	// File compilation mode
	filename := os.Args[1]
	if !strings.HasSuffix(filename, ".zeno") {
		fmt.Printf("Error: Expected .zeno file, got: %s\n", filename)
		os.Exit(1)
	}
	
	err := compileFile(filename)
	if err != nil {
		fmt.Printf("Compilation failed: %v\n", err)
		os.Exit(1)
	}
}

func runDemoTests() {
	fmt.Println("\nRunning demo tests:")
	
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
		fmt.Printf("❌ Parser errors: %v\n", p.Errors())
		return
	}
	
	// Step 3: Generate Go code
	goCode, err := generator.Generate(program)
	if err != nil {
		fmt.Printf("❌ Generator error: %v\n", err)
		return
	}
	
	// Clean up the output for better display
	goCode = strings.TrimSpace(goCode)
	
	fmt.Printf("Go output: %s\n", goCode)
	fmt.Println("✅ Success!")
}

func compileFile(filename string) error {
	// Read the Zeno source file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	fmt.Printf("Compiling file: %s\n", filename)
	fmt.Printf("Source code:\n%s\n", string(content))
	
	// Parse the Zeno code
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parser errors: %v", p.Errors())
	}
	
	// Generate Go code
	goCode, err := generator.Generate(program)
	if err != nil {
		return fmt.Errorf("generation error: %w", err)
	}
	
	// Output file name (replace .zeno with .go)
	outputFile := strings.TrimSuffix(filename, ".zeno") + ".go"
	
	// Write the generated Go code
	err = os.WriteFile(outputFile, []byte(goCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
	}
	
	fmt.Printf("✅ Successfully compiled to: %s\n", outputFile)
	return nil
}
