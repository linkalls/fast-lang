package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/linkalls/zeno-lang/generator"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/parser"
)

var (
	showJapanese = flag.Bool("jp", false, "Show error messages in Japanese as well")
)

func main() {
	flag.Parse()

	fmt.Println("=== Zeno Language Compiler ===")

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: No command or file specified")
		fmt.Println("Usage:")
		fmt.Println("  zeno run <filename.zeno>      # Compile and run")
		fmt.Println("  zeno compile <filename.zeno>  # Compile to Go file")
		fmt.Println("  zeno build <filename.zeno>    # Compile to executable")
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "run":
		if len(args) < 2 {
			fmt.Println("Error: 'run' command requires a filename")
			fmt.Println("Usage: zeno run <filename.zeno>")
			os.Exit(1)
		}
		fmt.Printf("=== Zeno Run Command ===\n")
		err := runFile(args[1])
		if err != nil {
			fmt.Printf("Run failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run failed: %v\n", err)
			os.Exit(1)
		}
	case "compile":
		if len(args) < 2 {
			fmt.Println("Error: 'compile' command requires a filename")
			fmt.Println("Usage: zeno compile <filename.zeno>")
			os.Exit(1)
		}
		err := compileFile(args[1])
		if err != nil {
			fmt.Printf("Compilation failed: %v\n", err)
			os.Exit(1)
		}
	case "build":
		if len(args) < 2 {
			fmt.Println("Error: 'build' command requires a filename")
			fmt.Println("Usage: zeno build <filename.zeno>")
			os.Exit(1)
		}
		err := buildExecutable(args[1])
		if err != nil {
			fmt.Printf("Build failed: %v\n", err)
			os.Exit(1)
		}
	default:
		// Backward compatibility: if first arg is a .zeno file, compile it
		if strings.HasSuffix(command, ".zeno") {
			err := compileFile(command)
			if err != nil {
				fmt.Printf("Compilation failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Error: Unknown command '%s'\n", command)
			fmt.Println("Usage:")
			fmt.Println("  zeno run <filename.zeno>      # Compile and run")
			fmt.Println("  zeno compile <filename.zeno>  # Compile to Go file")
			fmt.Println("  zeno build <filename.zeno>    # Compile to executable")
			os.Exit(1)
		}
	}
}

func compileFile(filename string) error {
	// Read the Zeno source file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	fmt.Printf("Compiling file: %s\n", filename)
	fmt.Printf("Source code:\n%s\n", string(content))

	// Parse the Zeno code
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parser errors: %v", p.Errors())
	}

	// Generate Go code
	goCode, err := generator.GenerateWithFile(program, filename, *showJapanese)
	if err != nil {
		return fmt.Errorf("generation error: %w", err)
	}

	// Output file name (replace .zeno with .go)
	outputFile := strings.TrimSuffix(filename, ".zeno") + ".go"

	// Write the generated Go code
	err = os.WriteFile(outputFile, []byte(goCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
	}

	fmt.Printf("✅ Successfully compiled to: %s\n", outputFile)
	return nil
}

func runFile(filename string) error {
	if !strings.HasSuffix(filename, ".zeno") {
		return fmt.Errorf("expected .zeno file, got: %s", filename)
	}

	// Read the Zeno source file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	fmt.Printf("Running file: %s\n", filename)

	// Parse the Zeno code
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors found:\n")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("parser errors: %v", p.Errors())
	}

	// Generate Go code
	fmt.Printf("Generating Go code...\n")
	goCode, err := generator.GenerateWithFile(program, filename, *showJapanese)
	if err != nil {
		fmt.Printf("Generation error details: %v\n", err)
		return fmt.Errorf("generation error: %w", err)
	}

	// Create temporary Go file
	tempDir := os.TempDir()
	baseName := strings.TrimSuffix(filepath.Base(filename), ".zeno")
	tempGoFile := filepath.Join(tempDir, baseName+".go")

	// Write the generated Go code to temporary file
	err = os.WriteFile(tempGoFile, []byte(goCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tempGoFile, err)
	}

	// Clean up temporary file when done
	defer os.Remove(tempGoFile)

	fmt.Printf("Generated temporary Go file: %s\n", tempGoFile)

	// Run the Go file
	cmd := exec.Command("go", "run", tempGoFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("\n--- Program Output ---")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Go command failed: %v\n", err)
		return fmt.Errorf("failed to run Go program: %w", err)
	}

	fmt.Println("--- End Output ---")
	return nil
}

func buildExecutable(filename string) error {
	if !strings.HasSuffix(filename, ".zeno") {
		return fmt.Errorf("expected .zeno file, got: %s", filename)
	}

	// Read the Zeno source file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	fmt.Printf("Building executable from: %s\n", filename)

	// Parse the Zeno code
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors found:\n")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("parser errors: %v", p.Errors())
	}

	// Generate Go code
	fmt.Printf("Generating Go code...\n")
	goCode, err := generator.GenerateWithFile(program, filename, *showJapanese)
	if err != nil {
		fmt.Printf("Generation error details: %v\n", err)
		return fmt.Errorf("generation error: %w", err)
	}

	// Output file names
	baseName := strings.TrimSuffix(filename, ".zeno")
	goFile := baseName + ".go"
	executableName := baseName

	// Write the generated Go code
	err = os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write Go file %s: %w", goFile, err)
	}

	fmt.Printf("Generated Go file: %s\n", goFile)

	// Build the executable using go build
	cmd := exec.Command("go", "build", "-o", executableName, goFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Building executable: %s\n", executableName)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Go build failed: %v\n", err)
		return fmt.Errorf("failed to build executable: %w", err)
	}

	fmt.Printf("✅ Successfully built executable: %s\n", executableName)
	fmt.Printf("   You can run it with: ./%s\n", executableName)
	
	// Clean up the temporary Go file
	err = os.Remove(goFile)
	if err != nil {
		fmt.Printf("Warning: Failed to remove temporary Go file %s: %v\n", goFile, err)
	}

	return nil
}
