package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentlint/agentlint/internal/core"
	golang "github.com/agentlint/agentlint/internal/languages/golang"
)

func TestGoAnalyzer(t *testing.T) {
	// Create a temporary directory with test Go files
	tempDir, err := os.MkdirTemp("", "agentlint-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test Go files
	testFiles := map[string]string{
		"large_function.go": `
package main

import "fmt"

// This function is too large
func largeFunction() {
	fmt.Println("Line 1")
	fmt.Println("Line 2")
	fmt.Println("Line 3")
	fmt.Println("Line 4")
	fmt.Println("Line 5")
	fmt.Println("Line 6")
	fmt.Println("Line 7")
	fmt.Println("Line 8")
	fmt.Println("Line 9")
	fmt.Println("Line 10")
	fmt.Println("Line 11")
	fmt.Println("Line 12")
	fmt.Println("Line 13")
	fmt.Println("Line 14")
	fmt.Println("Line 15")
	fmt.Println("Line 16")
	fmt.Println("Line 17")
	fmt.Println("Line 18")
	fmt.Println("Line 19")
	fmt.Println("Line 20")
	fmt.Println("Line 21")
	fmt.Println("Line 22")
	fmt.Println("Line 23")
	fmt.Println("Line 24")
	fmt.Println("Line 25")
	fmt.Println("Line 26")
	fmt.Println("Line 27")
	fmt.Println("Line 28")
	fmt.Println("Line 29")
	fmt.Println("Line 30")
	fmt.Println("Line 31")
	fmt.Println("Line 32")
	fmt.Println("Line 33")
	fmt.Println("Line 34")
	fmt.Println("Line 35")
	fmt.Println("Line 36")
	fmt.Println("Line 37")
	fmt.Println("Line 38")
	fmt.Println("Line 39")
	fmt.Println("Line 40")
	fmt.Println("Line 41")
	fmt.Println("Line 42")
	fmt.Println("Line 43")
	fmt.Println("Line 44")
	fmt.Println("Line 45")
	fmt.Println("Line 46")
	fmt.Println("Line 47")
	fmt.Println("Line 48")
	fmt.Println("Line 49")
	fmt.Println("Line 50")
	fmt.Println("Line 51")
}
`,
		"unused_function.go": `
package main

import "fmt"

// This function is never used
func unusedFunction() {
	fmt.Println("This function is never called")
}

func main() {
	fmt.Println("Hello, world!")
}
`,
		"overcommented.go": `
package main

import "fmt"

// This is a comment
func main() {
	// Print hello
	fmt.Println("Hello, world!") // Print hello world
	
	// Declare a variable
	x := 42 // Set x to 42
	
	// Print x
	fmt.Println(x) // Print the value of x
}
`,
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Create configuration with low thresholds for testing
	cfg := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 30,
			},
			FileSize: core.FileSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
			Overcommenting: core.OvercommentingConfig{
				Enabled:          true,
				MaxCommentRatio:  0.2,
				CheckRedundant:   true,
				CheckDocCoverage: true,
			},
			OrphanedCode: core.OrphanedCodeConfig{
				Enabled:              true,
				CheckUnusedFunctions: true,
				CheckUnusedVariables: true,
				CheckUnreachableCode: true,
				CheckDeadImports:     true,
			},
		},
		Output: core.OutputConfig{
			Format:  "console",
			Verbose: false,
		},
		Language: core.LanguageConfig{
			Go: core.GoConfig{
				IgnoreTests: false,
			},
		},
	}

	// Create Go analyzer
	analyzer := golang.NewAnalyzer(cfg)

	// Test each file
	ctx := context.Background()
	for filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(tempDir, filename)
			results, err := analyzer.Analyze(ctx, filePath, cfg)
			if err != nil {
				t.Fatalf("Failed to analyze file %s: %v", filename, err)
			}

			// Check that we found some issues
			if len(results) == 0 {
				t.Errorf("Expected to find issues in %s, but found none", filename)
			}

			// Print results for debugging
			t.Logf("Found %d issues in %s:", len(results), filename)
			for _, result := range results {
				t.Logf("  %s: %s", result.RuleName, result.Message)
			}
		})
	}
}

func TestFileScanner(t *testing.T) {
	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "agentlint-scan-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"file1.go",
		"file2.go",
		"not_go.txt",
		"subdir/file3.go",
		"subdir/file4.go",
	}

	for _, filename := range testFiles {
		dir := filepath.Dir(filename)
		if dir != "." {
			err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
			if err != nil {
				t.Fatalf("Failed to create subdir: %v", err)
			}
		}

		err := os.WriteFile(filepath.Join(tempDir, filename), []byte("package main"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Create scanner
	scanner := golang.NewFileScanner()

	// Scan for files
	ctx := context.Background()
	files, err := scanner.Scan(ctx, tempDir)
	if err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}

	// Check that we found the expected Go files
	expectedFiles := 4 // file1.go, file2.go, subdir/file3.go, subdir/file4.go
	if len(files) != expectedFiles {
		t.Errorf("Expected to find %d Go files, but found %d", expectedFiles, len(files))
	}

	// Check that all found files are .go files
	for _, file := range files {
		if filepath.Ext(file) != ".go" {
			t.Errorf("Found non-Go file: %s", file)
		}
	}
}
