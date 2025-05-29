package main

import (
	"fmt"
	"strings"

	"github.com/linkalls/zeno-lang/generator"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

func main() {
	fmt.Println("Testing Zeno binary expression support...")

	zenoCode := "let result = 5 + 3;"
	fmt.Printf("Input: %s\n", zenoCode)

	// Parse
	l := lexer.New(zenoCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Errors: %v\n", p.Errors())
		return
	}

	// Generate
	goCode, err := generator.Generate(program)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Output:\n%s\n", strings.TrimSpace(goCode))
}
