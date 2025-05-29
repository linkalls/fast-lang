package main

import (
	"fmt"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

func main() {
	input := "import {println} from \"std/fmt\";"
	
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	errors := p.Errors()
	if len(errors) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range errors {
			fmt.Printf("  %s\n", err)
		}
	} else {
		fmt.Printf("Parsed successfully: %s\n", program.String())
	}
}
