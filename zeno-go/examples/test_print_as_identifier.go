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

func Println(value string) {
	zenoNativePrintln(value)
}

func main() {
	var print = "This is a string assigned to a variable named 'print'."
	Println(print)
	var myPrintlnVar = "This is a string assigned to 'myPrintlnVar'."
	Println(myPrintlnVar)
}
