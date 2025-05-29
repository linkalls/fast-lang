package main

import (
	"fmt"
)

func main() {
		var x = 10
	var y = 20
	var result = (x + y)
	fmt.Println(result)

	// Print all variables to avoid unused variable warnings
	fmt.Printf("x = %v\n", x)
	fmt.Printf("y = %v\n", y)
	fmt.Printf("result = %v\n", result)
}
