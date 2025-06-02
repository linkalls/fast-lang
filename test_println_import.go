package main

import (
	"os"
	"encoding/json"
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

func zenoNativePrintVariadic(args []interface{}) {
	fmt.Print(args...)
}

func zenoNativePrintlnVariadic(args []interface{}) {
	fmt.Println(args...)
}

func zenoNativePrintVariadicWithFirst(first interface{}, rest []interface{}) {
	fmt.Print(first)
	for _, arg := range rest {
		fmt.Print(" ", arg)
	}
}

func zenoNativePrintlnVariadicWithFirst(first interface{}, rest []interface{}) {
	fmt.Print(first)
	for _, arg := range rest {
		fmt.Print(" ", arg)
	}
	fmt.Println()
}

func zenoNativeRemove(path string) bool {
	err := os.Remove(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", path, err)
		return false
	}
	return true
}

func zenoNativeGetCurrentDirectory() string {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		return ""
	}
	return pwd
}

func zenoNativeJsonParse(jsonString string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON string '%s': %v\n", jsonString, err)
		return nil
	}
	return result
}

func zenoNativeJsonStringify(value interface{}) string {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error stringifying to JSON for value '%v': %v\n", value, err)
		return ""
	}
	return string(jsonBytes)
}

func Println(first interface{}, rest ...interface{}) {
	zenoNativePrintlnVariadicWithFirst(first, rest)
}

func main() {
	Println("Hello from imported println!")
}
