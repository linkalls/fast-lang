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

func ReadFile(path string) string {
	return zenoNativeReadFile(path)
}

func WriteFile(path string, content string) bool {
	return zenoNativeWriteFile(path, content)
}

func Println(value string) {
	zenoNativePrintln(value)
}

func main() {
	var testFileName = "test_io_output.txt"
	var testContent = "Hello from Zeno std/io test!"
	Println("Attempting to write to file...")
	var success = WriteFile(testFileName, testContent)
	if success {
		Println("writeFile reported success.")
	} else {
		Println("writeFile reported failure.")
	}
	Println("Attempting to read from file...")
	var readContent = ReadFile(testFileName)
	Println(readContent)
	Println("Attempting to read non-existent file...")
	var nonExistent = ReadFile("this_file_should_not_exist.txt")
	Println(nonExistent)
}
