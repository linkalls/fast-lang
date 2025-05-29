package main

import (
	"fmt"
)

func main() {
	var x = 5
	if (x > 3) {
		fmt.Println("x is greater than 3")
	} else if (x == 3) {
		fmt.Println("x equals 3")
	} else {
		fmt.Println("x is less than 3")
	}
}
