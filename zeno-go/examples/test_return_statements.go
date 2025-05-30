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

func testEarlyVoidReturn() {
	Println("T1.1: Before early void return")
	return
}

func testVoidReturnInIf(condition bool) {
	var condStr = ""
	if condition {
		condStr = "true"
	} else {
		condStr = "false"
	}
	Println(("T1.2: Testing with condition = " + condStr))
	if condition {
		Println("T1.2: Void return from if (true)")
		return
	}
	Println("T1.2: After if (condition was false or returned)")
}

func testVoidReturnInElse(condition bool) {
	var condStr = ""
	if condition {
		condStr = "true"
	} else {
		condStr = "false"
	}
	Println(("T1.3: Testing with condition = " + condStr))
	if condition {
		Println("T1.3: Inside if block (condition true)")
	} else {
		Println("T1.3: Void return from else (condition false)")
		return
	}
	Println("T1.3: After if/else (should only print if condition was true)")
}

func testVoidReturnNested(outer bool, inner bool) {
	var outerStr = ""
	if outer {
		outerStr = "true"
	} else {
		outerStr = "false"
	}
	var innerStr = ""
	if inner {
		innerStr = "true"
	} else {
		innerStr = "false"
	}
	Println(((("T1.4: Testing with outer=" + outerStr) + ", inner=") + innerStr))
	if outer {
		if inner {
			Println("T1.4: Void return from nested if")
			return
		}
		Println("T1.4: After inner if (inner was false)")
	} else {
		Println("T1.4: Outer was false")
		return
	}
	Println("T1.4: End of testVoidReturnNested function body")
}

func testValReturnEnd() int {
	var x = 10
	Println("T2.1: Preparing to return x * 2")
	return (x * 2)
}

func testValReturnEarly(early bool) string {
	var earlyStr = ""
	if early {
		earlyStr = "true"
	} else {
		earlyStr = "false"
	}
	Println(("T2.2: Testing with early = " + earlyStr))
	if early {
		Println("T2.2: Returning early with string")
		return "Returned early"
	}
	Println("T2.2: Returning normally with string")
	return "Returned normally"
}

func testValReturnFromIf(condition bool) int {
	var condStr = ""
	if condition {
		condStr = "true"
	} else {
		condStr = "false"
	}
	Println(("T2.3: Testing with condition = " + condStr))
	if condition {
		Println("T2.3: Returning 100 from if block")
		return 100
	}
	Println("T2.3: Returning 0 as default")
	return 0
}

func testValReturnFromIfElse(condition bool) string {
	var condStr = ""
	if condition {
		condStr = "true"
	} else {
		condStr = "false"
	}
	Println(("T2.4: Testing with condition = " + condStr))
	if condition {
		Println("T2.4: Returning from if block")
		return "From if"
	} else {
		Println("T2.4: Returning from else block")
		return "From else"
	}
}

func main() {
	Println("=== Testing Return Statements ===")
	Println("\n--- Testing Void Returns (T1.x) ---")
	testEarlyVoidReturn()
	Println("---")
	testVoidReturnInIf(true)
	Println("---")
	testVoidReturnInIf(false)
	Println("---")
	testVoidReturnInElse(true)
	Println("---")
	testVoidReturnInElse(false)
	Println("---")
	testVoidReturnNested(true, true)
	Println("---")
	testVoidReturnNested(true, false)
	Println("---")
	testVoidReturnNested(false, true)
	Println("---")
	testVoidReturnNested(false, false)
	Println("\n--- Testing Value Returns (T2.x) ---")
	var res2_1 = testValReturnEnd()
	Print("T2.1 Result: ")
	if (res2_1 == 20) {
		Println("20 (Correct)")
	} else {
		Println("Not 20 (Incorrect)")
	}
	Println("---")
	var res2_2_early = testValReturnEarly(true)
	Print("T2.2 Early Result: ")
	Println(res2_2_early)
	Println("---")
	var res2_2_normal = testValReturnEarly(false)
	Print("T2.2 Normal Result: ")
	Println(res2_2_normal)
	Println("---")
	var res2_3_if = testValReturnFromIf(true)
	Print("T2.3 If Result: ")
	if (res2_3_if == 100) {
		Println("100 (Correct)")
	} else {
		Println("Not 100 (Incorrect)")
	}
	Println("---")
	var res2_3_else = testValReturnFromIf(false)
	Print("T2.3 Else Result: ")
	if (res2_3_else == 0) {
		Println("0 (Correct)")
	} else {
		Println("Not 0 (Incorrect)")
	}
	Println("---")
	var res2_4_if = testValReturnFromIfElse(true)
	Print("T2.4 If/Else (true) Result: ")
	Println(res2_4_if)
	Println("---")
	var res2_4_else = testValReturnFromIfElse(false)
	Print("T2.4 If/Else (false) Result: ")
	Println(res2_4_else)
	Println("\n=== Return Statement Test Completed ===")
}
