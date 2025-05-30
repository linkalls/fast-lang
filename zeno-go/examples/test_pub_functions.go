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

// zenoNativeRemove attempts to remove the file or empty directory.
// Returns true on success, false on failure.
func zenoNativeRemove(path string) bool {
	err := os.Remove(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", path, err)
		return false
	}
	return true
}

// zenoNativeGetCurrentDirectory returns the current working directory path.
// Returns an empty string on failure.
func zenoNativeGetCurrentDirectory() string {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		return ""
	}
	return pwd
}

func Println(value string) {
	zenoNativePrintln(value)
}

func privateAdd(a int, b int) int {
	return (a + b)
}

func PublicMultiply(x int, y int) int {
	return (x * y)
}

func greet(name string) {
	Println((("Hello, " + name) + "!"))
}

func PublicGreet(name string) {
	Println((("Public greeting: " + name) + "!"))
}

func main() {
	privateAdd(3, 4)
	PublicMultiply(5, 6)
	Println("Sum and product calculated by private/public functions (not displayed).")
	greet("Private User")
	PublicGreet("Public User")
}
