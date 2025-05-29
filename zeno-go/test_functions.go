package main

import (
	"fmt"
)

func add(a int64, b int64) int64 {
	return (a + b)
}

func greet(name string) {
	fmt.Println((("Hello, " + name) + "!"))
}

func main() {
	var result int64 = add(5, 3)
	fmt.Println("Result: ", result)
	greet("Zeno")
}
