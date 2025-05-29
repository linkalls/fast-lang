package main

import (
	"fmt"
)

func main() {
	var x = 5
	var y = (x > 3)
	fmt.Println(y)
	if y {
		fmt.Println("g")
	}
}
