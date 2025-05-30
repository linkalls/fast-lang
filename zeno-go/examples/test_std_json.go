package main

import (
	"fmt"
	"os"
	"encoding/json"
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

// zenoNativeJsonParse parses a JSON string and returns the result as interface{}.
// Returns nil if parsing fails.
func zenoNativeJsonParse(jsonString string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON string '%s': %v\n", jsonString, err)
		return nil
	}
	return result
}

// zenoNativeJsonStringify converts an interface{} to a JSON string.
// Returns an empty string if stringification fails.
func zenoNativeJsonStringify(value interface{}) string {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error stringifying to JSON for value '%v': %v\n", value, err)
		return ""
	}
	return string(jsonBytes)
}

func Print(value string) {
	zenoNativePrint(value)
}

func Println(value string) {
	zenoNativePrintln(value)
}

func Parse(jsonString string) interface{} {
	return zenoNativeJsonParse(jsonString)
}

func Stringify(value interface{}) string {
	return zenoNativeJsonStringify(value)
}

func main() {
	Println("=== Testing std/json module ===")
	Println("\n--- Test Case 1: Object Round Trip ---")
	var jsonObjectString = "{\"name\": \"Zeno\", \"version\": 0.1, \"isAwesome\": true, \"features\": [\"typed\", \"simple\"], \"details\": null}"
	Print("Original Object JSON: ")
	Println(jsonObjectString)
	var parsedObject = Parse(jsonObjectString)
	var stringifiedObject = Stringify(parsedObject)
	Print("Stringified Object JSON: ")
	Println(stringifiedObject)
	if ((stringifiedObject == "") && (jsonObjectString != "{}")) {
		Println("ERROR: Stringify returned empty for a non-empty parsed object!")
	} else if ((stringifiedObject == "null") && (jsonObjectString != "null")) {
		Println("ERROR: Stringify returned JSON null for a non-null parsed object!")
	} else {
		Println("Object Round Trip: Appears OK (stringified is not unexpectedly empty or null)")
	}
	Println("\n--- Test Case 2: Array Round Trip ---")
	var jsonArrayString = "[10, \"hello\", false, null, {\"key\": \"value\"}]"
	Print("Original Array JSON: ")
	Println(jsonArrayString)
	var parsedArray = Parse(jsonArrayString)
	var stringifiedArray = Stringify(parsedArray)
	Print("Stringified Array JSON: ")
	Println(stringifiedArray)
	if ((stringifiedArray == "") && (jsonArrayString != "[]")) {
		Println("ERROR: Stringify returned empty for a non-empty parsed array!")
	} else if ((stringifiedArray == "null") && (jsonArrayString != "null")) {
		Println("ERROR: Stringify returned JSON null for a non-null parsed array!")
	} else {
		Println("Array Round Trip: Appears OK (stringified is not unexpectedly empty or null)")
	}
	Println("\n--- Test Case 3: Stringify Primitives ---")
	Print("Stringify string \"zeno\": ")
	Println(Stringify("zeno"))
	Print("Stringify int 123: ")
	Println(Stringify(123))
	Print("Stringify float 3.14: ")
	Println(Stringify(3.14))
	Print("Stringify bool true: ")
	Println(Stringify(true))
	Println("\n--- Test Case 4: Parse Invalid JSON ---")
	var invalidJson = "{\"name\": \"Zeno\", "
	Print("Parsing invalid JSON: '")
	Print(invalidJson)
	Println("'")
	var parsedInvalid = Parse(invalidJson)
	var stringifiedNull = Stringify(parsedInvalid)
	Print("Stringified result of invalid parse: ")
	Println(stringifiedNull)
	if (stringifiedNull == "null") {
		Println("Parse Invalid JSON: OK (resulted in JSON null when stringified)")
	} else {
		Println(("ERROR: Invalid JSON parse did not result in null when stringified! Got: " + stringifiedNull))
	}
	Println("\n--- Test Case 5: Parse Valid JSON null ---")
	var nullJson = "null"
	Print("Parsing JSON: '")
	Print(nullJson)
	Println("'")
	var parsedValidNull = Parse(nullJson)
	var stringifiedValidNull = Stringify(parsedValidNull)
	Print("Stringified result of valid null parse: ")
	Println(stringifiedValidNull)
	if (stringifiedValidNull == "null") {
		Println("Parse Valid JSON null: OK")
	} else {
		Println(("ERROR: Valid JSON null parse did not result in null when stringified! Got: " + stringifiedValidNull))
	}
	Println("\n=== std/json module test completed ===")
}
