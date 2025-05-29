package main

import (
	"fmt"
	"strings"

	"github.com/linkalls/zeno-lang/generator"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

func main() {
	fmt.Println("=== Zeno Binary Expression Tests ===")

	testCases := []struct {
		name string
		code string
	}{
		{"Simple addition", "let result = 5 + 3;"},
		{"Multiplication with precedence", "let calc = 2 * 3 + 4;"},
		{"Complex expression", "let complex = 10 + 2 * 5 - 1;"},
		{"Mutable variable", "mut counter = 1 + 1;"},
		{"Type annotation", "let value: int = 42 * 2;"},
		{"String assignment", "let message = \"Hello\";"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)
		fmt.Printf("Zeno: %s\n", tc.code)

		// Parse
		l := lexer.New(tc.code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Printf("❌ Parser errors: %v\n", p.Errors())
			continue
		}

		// Generate
		goCode, err := generator.Generate(program)
		if err != nil {
			fmt.Printf("❌ Generator error: %v\n", err)
			continue
		}

		// Extract just the variable declaration line
		lines := strings.Split(goCode, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "var ") || strings.HasPrefix(line, "mut ") {
				fmt.Printf("Go: %s\n", line)
				break
			}
		}
		fmt.Println("✅ Success")
	}
}
