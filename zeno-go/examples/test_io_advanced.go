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

func ReadFile(path string) string {
	return zenoNativeReadFile(path)
}

func WriteFile(path string, content string) bool {
	return zenoNativeWriteFile(path, content)
}

func main() {
	Println("=== Zeno std/io Module Test ===")
	Println("Test 1: Writing a simple text file...")
	var simpleContent = "Hello from Zeno!"
	WriteFile("simple.txt", simpleContent)
	Println("✓ simple.txt created")
	Println("Test 2: Writing a configuration file...")
	var configContent = "# Zeno Configuration\nname=MyApp\nversion=1.0\ndebug=true"
	WriteFile("config.txt", configContent)
	Println("✓ config.txt created")
	Println("Test 3: Reading files...")
	Print("simple.txt content: ")
	var simpleRead = ReadFile("simple.txt")
	Println(simpleRead)
	Println("config.txt content:")
	var configRead = ReadFile("config.txt")
	Println(configRead)
	Println("Test 4: Writing structured data...")
	var jsonData = "{\"name\": \"Zeno\", \"type\": \"programming-language\"}"
	WriteFile("data.json", jsonData)
	Println("✓ data.json created")
	var jsonRead = ReadFile("data.json")
	Print("data.json content: ")
	Println(jsonRead)
	Println("=== All tests completed! ===")
}
