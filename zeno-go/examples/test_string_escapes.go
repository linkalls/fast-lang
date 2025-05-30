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

func ReadFile(path string) string {
	return zenoNativeReadFile(path)
}

func WriteFile(path string, content string) bool {
	return zenoNativeWriteFile(path, content)
}

func main() {
	Println("Testing string escape sequences:")
	var multiLineText = "Line 1\nLine 2\nLine 3"
	var tabbedText = "Column1\tColumn2\tColumn3"
	var quotedText = "He said \"Hello, World!\""
	var pathText = "C:\\Users\\zeno\\file.txt"
	WriteFile("multiline.txt", multiLineText)
	WriteFile("tabbed.txt", tabbedText)
	WriteFile("quoted.txt", quotedText)
	WriteFile("path.txt", pathText)
	Println("Multi-line content:")
	var readMulti = ReadFile("multiline.txt")
	Println(readMulti)
	Println("Tabbed content:")
	var readTabbed = ReadFile("tabbed.txt")
	Println(readTabbed)
	Println("Quoted content:")
	var readQuoted = ReadFile("quoted.txt")
	Println(readQuoted)
	Println("Path content:")
	var readPath = ReadFile("path.txt")
	Println(readPath)
}
