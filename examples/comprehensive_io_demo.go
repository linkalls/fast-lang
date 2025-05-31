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

func ReadFile(path string) string {
	return zenoNativeReadFile(path)
}

func WriteFile(path string, content string) bool {
	return zenoNativeWriteFile(path, content)
}

func Print(value string) {
	zenoNativePrint(value)
}

func Println(value string) {
	zenoNativePrintln(value)
}

func main() {
	Println("🚀 Zeno std/io Module - Comprehensive Demo")
	Println("==========================================")
	Println("📝 Basic File Operations:")
	var greeting = "Hello, World from Zeno!"
	WriteFile("hello.txt", greeting)
	var readGreeting = ReadFile("hello.txt")
	Print("📖 Read: ")
	Println(readGreeting)
	Println("⚙️  Configuration File Example:")
	var config = "# Application Configuration\napp_name=ZenoApp\nport=8080\ndebug=true\nversion=1.0.0"
	WriteFile("app.conf", config)
	var configContent = ReadFile("app.conf")
	Println("📋 Configuration file contents:")
	Println(configContent)
	Println("💾 Data Serialization Example:")
	var userData = "{\"id\": 1, \"name\": \"Alice\", \"email\": \"alice@example.com\"}"
	WriteFile("user.json", userData)
	var userJson = ReadFile("user.json")
	Print("👤 User data: ")
	Println(userJson)
	Println("📄 Multi-line Content Example:")
	var multiLine = "Line 1: Introduction\nLine 2: Features\nLine 3: Usage\nLine 4: Conclusion"
	WriteFile("document.txt", multiLine)
	var document = ReadFile("document.txt")
	Println("📚 Document contents:")
	Println(document)
	Println("🚫 Error Handling Demo:")
	var nonExistent = ReadFile("does_not_exist.txt")
	Print("📄 Non-existent file read result: '")
	Print(nonExistent)
	Println("'")
	Println("✅ Gracefully handled non-existent file (returned empty string)")
	Println("✨ Demo completed successfully!")
	Println("Created files: hello.txt, app.conf, user.json, document.txt")
}
