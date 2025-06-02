package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/linkalls/zeno-lang/generator"
	"github.com/linkalls/zeno-lang/lexer"
	"github.com/linkalls/zeno-lang/linter"
	"github.com/linkalls/zeno-lang/parser"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zeno",
	Short: "Zeno Language Compiler and Tools",
	Long:  `Zeno is a programming language. This CLI provides tools to compile, run, build, and lint Zeno source files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior if no subcommand is given, or print help
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		// Backward compatibility: if first arg is a .zeno file, try to run it
		if strings.HasSuffix(args[0], ".zeno") {
			fmt.Println("Executing default action (run) for .zeno file.")
			runCmd.Run(cmd, args)
		} else {
			cmd.Help()
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run <filename.zeno>",
	Short: "Compile and run a Zeno file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("=== Zeno Run Command ===\n")
		if err := runFile(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Run failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var compileCmd = &cobra.Command{
	Use:   "compile <filename.zeno>",
	Short: "Compile a Zeno file to Go",
	Long:  `Compiles a Zeno source file (.zeno) into a Go source file (.go) in the same directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("=== Zeno Compile Command ===\n")
		if err := compileFile(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var buildCmd = &cobra.Command{
	Use:   "build <filename.zeno>",
	Short: "Compile a Zeno file to an executable",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("=== Zeno Build Command ===\n")
		if err := buildExecutable(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var lintCmd = &cobra.Command{
	Use:   "lint [filepath or directory]",
	Short: "Lints Zeno source files for potential issues.",
	Long: `Lints Zeno source files (.zeno) for potential issues, including naming conventions,
unused variables, unused functions, and unused imports.
You can specify one or more file paths or directories.
If a directory is specified, it will be walked recursively for .zeno files.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("=== Zeno Lint Command ===\n")
		var allIssues []linter.Issue
		hasErrors := false

		for _, pathArg := range args {
			info, err := os.Stat(pathArg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", pathArg, err)
				hasErrors = true
				continue
			}

			var filesToLint []string
			if info.IsDir() {
				err := filepath.WalkDir(pathArg, func(currentPath string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if !d.IsDir() && (strings.HasSuffix(currentPath, ".zeno") || strings.HasSuffix(currentPath, ".zn")) {
						filesToLint = append(filesToLint, currentPath)
					}
					return nil
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", pathArg, err)
					hasErrors = true
					continue
				}
			} else {
				if strings.HasSuffix(pathArg, ".zeno") || strings.HasSuffix(pathArg, ".zn") {
					filesToLint = append(filesToLint, pathArg)
				} else {
					fmt.Fprintf(os.Stderr, "Skipping non-Zeno file: %s\n", pathArg)
					continue
				}
			}

			for _, filePath := range filesToLint {
				fmt.Printf("Linting file: %s\n", filePath)
				content, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
					hasErrors = true
					continue
				}

				l := lexer.New(string(content))
				p := parser.NewWithInput(l, filePath, string(content))
				program := p.ParseProgram()

				if len(p.Errors()) > 0 {
					fmt.Fprintf(os.Stderr, "Parser errors in %s:\n\n", filePath)
					// Display detailed errors if available
					detailedErrors := p.DetailedErrors()
					if len(detailedErrors) > 0 {
						for _, err := range detailedErrors {
							fmt.Fprintf(os.Stderr, "%s", err.String())
							fmt.Fprintf(os.Stderr, "\n")
						}
					} else {
						// Fallback to simple errors
						for _, msg := range p.Errors() {
							fmt.Fprintf(os.Stderr, "  - %s\n", msg)
						}
					}
					hasErrors = true
					continue
				}

				absFilePath, _ := filepath.Abs(filePath)

				// Initialize linter and register rules
				rules := []linter.Rule{
					&linter.UnusedVariableRule{},
					&linter.UnusedFunctionRule{},
					&linter.FunctionNameRule{},
					&linter.VariableNameRule{},
					&linter.UnusedImportRule{},
				}
				zenoFrameworkLinter := linter.NewLinter(rules)

				issues, err := zenoFrameworkLinter.Lint(program, absFilePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Linter error in %s: %v\n", filePath, err)
					hasErrors = true
				}

				if len(issues) > 0 {
					allIssues = append(allIssues, issues...)
				}
			}
		}

		if len(allIssues) > 0 {
			fmt.Printf("\nFound %d linting issue(s):\n", len(allIssues))
			for _, issue := range allIssues {
				// Use 1 if line/col is 0 from placeholder
				line := issue.Line
				if line == 0 {
					line = 1
				}
				col := issue.Column
				if col == 0 {
					col = 1
				}
				fmt.Printf("%s:%d:%d: [%s] %s\n", issue.Filepath, line, col, issue.RuleName, issue.Message)
			}
			hasErrors = true // Ensure exit code reflects issues found
		} else {
			fmt.Println("No linting issues found.")
		}

		if hasErrors {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(lintCmd)
	// Potentially add flags here, e.g., for -jp (Japanese error messages) if Cobra handles them globally
}

func main() {
	// The old main logic is now handled by Cobra commands.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

// --- Existing helper functions (compileFile, runFile, buildExecutable) ---
// These are kept as they are called by the new Cobra commands.

func compileFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	// fmt.Printf("Compiling file: %s\n", filename) // Cobra command will print this
	// fmt.Printf("Source code:\n%s\n", string(content)) // Too verbose for default compile

	l := lexer.New(string(content))
	p := parser.NewWithInput(l, filename, string(content))
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors in %s:\n\n", filename)
		detailedErrors := p.DetailedErrors()
		if len(detailedErrors) > 0 {
			for _, err := range detailedErrors {
				fmt.Fprintf(os.Stderr, "%s", err.String())
				fmt.Fprintf(os.Stderr, "\n")
			}
		} else {
			for _, msg := range p.Errors() {
				fmt.Fprintf(os.Stderr, "  - %s\n", msg)
			}
		}
		return fmt.Errorf("parser errors found")
	}

	goCode, err := generator.GenerateWithFile(program, filename)
	if err != nil {
		return fmt.Errorf("generation error: %w", err)
	}

	outputFile := strings.TrimSuffix(filename, ".zeno") + ".go"
	if strings.HasSuffix(filename, ".zn") { // also handle .zn
		outputFile = strings.TrimSuffix(filename, ".zn") + ".go"
	}

	err = os.WriteFile(outputFile, []byte(goCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
	}

	fmt.Printf("✅ Successfully compiled %s to: %s\n", filename, outputFile)
	return nil
}

func runFile(filename string) error {
	if !strings.HasSuffix(filename, ".zeno") && !strings.HasSuffix(filename, ".zn") {
		return fmt.Errorf("expected .zeno or .zn file, got: %s", filename)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	// fmt.Printf("Running file: %s\n", filename) // Cobra command will print this

	l := lexer.New(string(content))
	p := parser.NewWithInput(l, filename, string(content))
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors in %s:\n\n", filename)
		detailedErrors := p.DetailedErrors()
		if len(detailedErrors) > 0 {
			for _, err := range detailedErrors {
				fmt.Fprintf(os.Stderr, "%s", err.String())
				fmt.Fprintf(os.Stderr, "\n")
			}
		} else {
			for _, msg := range p.Errors() {
				fmt.Fprintf(os.Stderr, "  - %s\n", msg)
			}
		}
		return fmt.Errorf("parser errors found")
	}

	// fmt.Printf("Generating Go code...\n") // Too verbose
	goCode, err := generator.GenerateWithFile(program, filename)
	if err != nil {
		// fmt.Printf("Generation error details: %v\n", err) // Too verbose
		return fmt.Errorf("generation error: %w", err)
	}

	tempDir := os.TempDir()
	// Ensure generated Go file does not end with _test.go to allow go run
	baseName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	tempGoFile := filepath.Join(tempDir, baseName+"_zeno_run.go")

	err = os.WriteFile(tempGoFile, []byte(goCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tempGoFile, err)
	}
	defer os.Remove(tempGoFile)
	// fmt.Printf("Generated temporary Go file: %s\n", tempGoFile)

	cmd := exec.Command("go", "run", tempGoFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("\n--- Program Output ---")
	err = cmd.Run()
	fmt.Println("--- End Output ---")
	if err != nil {
		// fmt.Printf("Go command failed: %v\n", err) // Error is usually printed by cmd.Stderr
		return fmt.Errorf("failed to run Go program: %w", err)
	}
	return nil
}

func buildExecutable(filename string) error {
	if !strings.HasSuffix(filename, ".zeno") && !strings.HasSuffix(filename, ".zn") {
		return fmt.Errorf("expected .zeno or .zn file, got: %s", filename)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	// fmt.Printf("Building executable from: %s\n", filename) // Cobra command handles this

	l := lexer.New(string(content))
	p := parser.NewWithInput(l, filename, string(content))
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors in %s:\n\n", filename)
		detailedErrors := p.DetailedErrors()
		if len(detailedErrors) > 0 {
			for _, err := range detailedErrors {
				fmt.Fprintf(os.Stderr, "%s", err.String())
				fmt.Fprintf(os.Stderr, "\n")
			}
		} else {
			for _, msg := range p.Errors() {
				fmt.Fprintf(os.Stderr, "  - %s\n", msg)
			}
		}
		return fmt.Errorf("parser errors found")
	}

	// fmt.Printf("Generating Go code...\n")
	goCode, err := generator.GenerateWithFile(program, filename)
	if err != nil {
		return fmt.Errorf("generation error: %w", err)
	}

	baseName := strings.TrimSuffix(filename, ".zeno")
	if strings.HasSuffix(filename, ".zn") {
		baseName = strings.TrimSuffix(filename, ".zn")
	}

	// Create a temporary directory for the build process
	buildDir, err := os.MkdirTemp("", "zeno_build_*")
	if err != nil {
		return fmt.Errorf("failed to create temporary build directory: %w", err)
	}
	defer os.RemoveAll(buildDir) // Clean up the temporary directory

	goFile := filepath.Join(buildDir, filepath.Base(baseName)+".go")
	executableName := filepath.Base(baseName) // Executable in current dir, not temp

	err = os.WriteFile(goFile, []byte(goCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write Go file %s: %w", goFile, err)
	}
	// fmt.Printf("Generated Go file: %s\n", goFile)

	cmd := exec.Command("go", "build", "-o", executableName, goFile)
	cmd.Stdout = os.Stdout // Show build output/errors directly
	cmd.Stderr = os.Stderr
	// fmt.Printf("Building executable: %s\n", executableName)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to build executable: %w", err)
	}

	fmt.Printf("✅ Successfully built executable: %s\n", executableName)
	fmt.Printf("   You can run it with: ./%s\n", executableName)
	return nil
}
