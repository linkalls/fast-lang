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

func Remove(path string) bool {
	return zenoNativeRemove(path)
}

func Pwd() string {
	return zenoNativeGetCurrentDirectory()
}

func main() {
	Println("=== Testing std/io extended features: pwd and remove ===")
	var currentDir = Pwd()
	Print("Current working directory: ")
	Println(currentDir)
	if (currentDir == "") {
		Println("ERROR: pwd() returned an empty string!")
	} else {
		Println("pwd() test: OK (returned a non-empty path)")
	}
	Println("------------------------------------")
	var testFileForRemove = "test_remove_me.txt"
	var testContent = "This file is for testing the remove() function."
	Println(("Creating file for removal test: " + testFileForRemove))
	WriteFile(testFileForRemove, testContent)
	var contentBeforeRemove = ReadFile(testFileForRemove)
	if (contentBeforeRemove == testContent) {
		Println("File created successfully for remove test.")
	} else {
		Println("ERROR: File creation/read back failed before remove test!")
	}
	Println("------------------------------------")
	Println(("Attempting to remove: " + testFileForRemove))
	var removeSuccess = Remove(testFileForRemove)
	if removeSuccess {
		Println("remove() reported success.")
	} else {
		Println("ERROR: remove() reported failure for existing file!")
	}
	Println("------------------------------------")
	Println(("Attempting to read removed file (should be empty): " + testFileForRemove))
	var contentAfterRemove = ReadFile(testFileForRemove)
	if (contentAfterRemove == "") {
		Println("File successfully removed (readFile returned empty).")
	} else {
		Println("ERROR: File still exists or readFile did not return empty after remove!")
		Print("Content found: ")
		Println(contentAfterRemove)
	}
	Println("------------------------------------")
	var nonExistentFile = "this_file_does_not_exist_for_removal.txt"
	Println(("Attempting to remove non-existent file: " + nonExistentFile))
	var removeNonExistentSuccess = Remove(nonExistentFile)
	var invertedTest = (!removeNonExistentSuccess)
	if invertedTest {
		Println("remove() correctly reported failure for non-existent file.")
	} else {
		Println("ERROR: remove() reported success for non-existent file!")
	}
	Println("------------------------------------")
	Println("=== std/io extended features test completed ===")
}
