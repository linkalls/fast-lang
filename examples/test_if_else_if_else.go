package main

import (
	"os"
	"fmt"
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

func testIfs(a int, b int) {
	Println("Testing with a and b (values not directly printed in this line)")
	if (a > b) {
		Println("  Result: a is greater than b")
	} else if (a < b) {
		Println("  Result: a is less than b")
	} else {
		Println("  Result: a is equal to b")
	}
	if (a == 10) {
		Println("  Condition: a is 10")
	}
	if (b == 20) {
		Println("  Condition: b is 20, then block only")
	} else {
		Println("  Condition: b is not 20, else block only")
	}
	if (a > 0) {
		if (b > 0) {
			Println("  Flow: both a and b are positive")
		} else if (b == 0) {
			Println("  Flow: a is positive, b is zero")
		} else {
			Println("  Flow: a is positive, b is negative")
		}
	} else if (a == 0) {
		Println("  Flow: a is zero")
	} else {
		if (b < 0) {
			Println("  Flow: both a and b are negative")
		} else {
			Println("  Flow: a is negative, b is not negative (zero or positive)")
		}
	}
	Println("---")
}

func main() {
	testIfs(10, 5)
	testIfs(5, 10)
	testIfs(7, 7)
	testIfs(10, 20)
	testIfs(0, 0)
	testIfs((-5), (-10))
	testIfs((-5), 5)
	testIfs(10, 0)
}
