package main

import (
	"fmt"
)

func privateAdd(a int64, b int64) int64 {
	return (a + b)
}

func PublicMultiply(x int64, y int64) int64 {
	return (x * y)
}

func greet(name string) {
	fmt.Println((("Hello, " + name) + "!"))
}

func PublicGreet(name string) {
	fmt.Println((("Public greeting: " + name) + "!"))
}

func main() {
	var sum int64 = privateAdd(3, 4)
	var product int64 = PublicMultiply(5, 6)
	fmt.Println("Sum: ", sum)
	fmt.Println("Product: ", product)
	greet("Private User")
	PublicGreet("Public User")
}
