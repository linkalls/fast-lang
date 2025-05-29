package main

import (
	"fmt"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

func main() {
	input := "let y = x > 3;"
	l := lexer.New(input)
	p := parser.New(l)
	
	fmt.Println("Parsing:", input)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("  ", err)
		}
	} else {
		fmt.Println("Program parsed successfully:")
		fmt.Println(program.String())
	}
}
