package main

import (
"fmt"
"github.com/linkalls/zeno-lang/lexer"
"github.com/linkalls/zeno-lang/token"
)

func main() {
input := "import {println} from \"std/fmt\";"

l := lexer.New(input)

for {
tok := l.NextToken()
fmt.Printf("Type: %s, Literal: %s\n", tok.Type, tok.Literal)
if tok.Type == token.EOF {
break
}
}
}
