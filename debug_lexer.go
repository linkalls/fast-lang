package main

import (
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/token"
)

func main() {
	input := `import {type Result} from "std/result"`
	l := lexer.New(input)

	// Disabled debug loop
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
	}
}
