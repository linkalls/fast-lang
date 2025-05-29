package main

import (
	"fmt"
)

func main() {
	var sum = (10 + 5)
	var product = (4 * 3)
	var complex = ((2 * 3) + 1)
	var counter = 0
	var message = "Binary expressions work!"

	// Print all variables to avoid unused variable warnings
	fmt.Printf("sum = %v\n", sum)
	fmt.Printf("product = %v\n", product)
	fmt.Printf("complex = %v\n", complex)
	fmt.Printf("counter = %v\n", counter)
	fmt.Printf("message = %v\n", message)
}
