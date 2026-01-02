package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	golang "github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

// generateLargeFunc generates a Go file with a function of the given line count
func generateLargeFunc(lines int) string {
	var sb strings.Builder
	sb.WriteString("package main\n\nimport \"fmt\"\n\n// This function is too large\nfunc largeFunction() {\n")
	for i := 1; i <= lines; i++ {
		fmt.Fprintf(&sb, "\tfmt.Println(\"Line %d\")\n", i)
	}
	sb.WriteString("}\n")
	return sb.String()
}

// getTestFilesForAnalyzer returns test file content for analyzer tests
func getTestFilesForAnalyzer() map[string]string {
	return map[string]string{
		"large_function.go": generateLargeFunc(51),
		"another_large.go": generateLargeFunc(35), // Will exceed MaxLines: 30
		"overcommented.go": `package main

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
}

// getTestConfigForAnalyzer returns a test config with low thresholds
func getTestConfigForAnalyzer() core.Config {
	return core.Config{
		Rules: core.RulesConfig{
			FunctionSize:   core.FunctionSizeConfig{Enabled: true, MaxLines: 30},
			FileSize:       core.FileSizeConfig{Enabled: true, MaxLines: 50},
			Overcommenting: core.OvercommentingConfig{Enabled: true, MaxCommentRatio: 0.2, CheckRedundant: true, CheckDocCoverage: true},
			OrphanedCode:   core.OrphanedCodeConfig{Enabled: true, CheckUnusedFunctions: true, CheckUnusedVariables: true, CheckUnreachableCode: true, CheckDeadImports: true},
		},
		Output:   core.OutputConfig{Format: "console", Verbose: false},
		Language: core.LanguageConfig{Go: core.GoConfig{IgnoreTests: false}},
	}
}

func TestGoAnalyzer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agentlint-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := getTestFilesForAnalyzer()
	for filename, content := range testFiles {
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	cfg := getTestConfigForAnalyzer()
	analyzer := golang.NewAnalyzer(cfg)
	ctx := context.Background()

	for filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(tempDir, filename)
			results, err := analyzer.Analyze(ctx, filePath, cfg)
			if err != nil {
				t.Fatalf("Failed to analyze file %s: %v", filename, err)
			}
			if len(results) == 0 {
				t.Errorf("Expected to find issues in %s, but found none", filename)
			}
			t.Logf("Found %d issues in %s:", len(results), filename)
			for _, result := range results {
				t.Logf("  %s: %s", result.RuleName, result.Message)
			}
		})
	}
}

// createTestFiles creates test files in the given directory
func createTestFiles(t *testing.T, tempDir string, files []string) {
	for _, filename := range files {
		dir := filepath.Dir(filename)
		if dir != "." {
			if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
				t.Fatalf("Failed to create subdir: %v", err)
			}
		}
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte("package main"), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}
}

func TestFileScanner(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agentlint-scan-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := []string{"file1.go", "file2.go", "not_go.txt", "subdir/file3.go", "subdir/file4.go"}
	createTestFiles(t, tempDir, testFiles)

	scanner := golang.NewFileScanner()
	ctx := context.Background()
	files, err := scanner.Scan(ctx, tempDir)
	if err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}

	expectedFiles := 4
	if len(files) != expectedFiles {
		t.Errorf("Expected to find %d Go files, but found %d", expectedFiles, len(files))
	}

	for _, file := range files {
		if filepath.Ext(file) != ".go" {
			t.Errorf("Found non-Go file: %s", file)
		}
	}
}
