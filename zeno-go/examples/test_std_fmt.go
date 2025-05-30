package main

import (
	"fmt"
	"os"
)

// Native function helpers
func zenoNativeReadFile(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return ""
	}
	return string(data)
}

func zenoNativeWriteFile(filename string, content string) bool {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
		return false
	}
	return true
}

func zenoNativePrint(args ...interface{}) {
	fmt.Print(args...)
}

func zenoNativePrintln(args ...interface{}) {
	fmt.Println(args...)
}

func Print(value string) {
	zenoNativePrint(value)
}

func Println(value string) {
	zenoNativePrintln(value)
}

func main() {
	Println("Testing println: Line 1")
	Print("Testing print: Part 1, ")
	Print("Part 2 - with explicit newline in string\n")
	var number_as_string = "12345"
	Print("Number as string: ")
	Println(number_as_string)
	var boolean_as_string = "true"
	Print("Boolean as string: ")
	Println(boolean_as_string)
	Println("Test complete.")
}
