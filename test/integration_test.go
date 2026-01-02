package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	golang "github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
	"github.com/CiaranMcAleer/AgentLint/internal/output"
	"github.com/CiaranMcAleer/AgentLint/internal/profiling"
)

func TestIntegrationBasic(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	lines := []string{"package main", "", "func largeFunc() {"}
	for i := 0; i < 60; i++ {
		lines = append(lines, fmt.Sprintf("	line%d := %d", i, i))
	}
	lines = append(lines, "}", "", "func main() {", "	largeFunc()", "}")
	content := strings.Join(lines, "\n")
	os.WriteFile(testFile, []byte(content), 0644)

	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
		},
	}

	analyzer := golang.NewAnalyzer(config)

	results, err := analyzer.Analyze(context.Background(), testFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	foundLargeFunc := false
	for _, r := range results {
		if strings.Contains(r.Message, "largeFunc") {
			foundLargeFunc = true
		}
	}

	if !foundLargeFunc {
		t.Error("Expected to find large function 'largeFunc'")
	}
}

func TestIntegrationOvercommenting(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

// This is a comment
// Another comment
// More comments
// Even more comments
// Many comments

func main() {
	_ = 1 + 1
}
`
	os.WriteFile(testFile, []byte(content), 0644)

	config := core.Config{
		Rules: core.RulesConfig{
			Overcommenting: core.OvercommentingConfig{
				Enabled:         true,
				MaxCommentRatio: 0.3,
			},
		},
	}

	analyzer := golang.NewAnalyzer(config)

	results, err := analyzer.Analyze(context.Background(), testFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to find overcommenting issue")
	}
}

func TestIntegrationUnusedFunction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file with an unused function
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package testpkg

func unusedFunc() {
	_ = 1 + 1
}

func main() {
	_ = 2 + 2
}
`
	os.WriteFile(testFile, []byte(content), 0644)

	// Use CrossFileAnalyzer for unused function detection
	// (single-file analysis cannot reliably detect unused functions
	// as they may be called from other files)
	crossFileAnalyzer := golang.NewCrossFileAnalyzer()
	err := crossFileAnalyzer.AnalyzeDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	results := crossFileAnalyzer.FindUnusedFunctions()

	foundUnused := false
	for _, r := range results {
		if strings.Contains(r.Message, "unusedFunc") {
			foundUnused = true
		}
	}

	if !foundUnused {
		t.Error("Expected to find unused function 'unusedFunc'")
	}
}

func TestIntegrationJSONOutput(t *testing.T) {
	results := []core.Result{
		{
			RuleID:     "test-rule",
			RuleName:   "Test Rule",
			Category:   "size",
			Severity:   "warning",
			FilePath:   "test.go",
			Line:       10,
			Message:    "Test message",
			Suggestion: "Test suggestion",
		},
	}

	formatter := output.NewJSONFormatter(false)
	err := formatter.Format(results)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
}

func TestIntegrationParallelAnalysis(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 20; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.go", i))
		lines := []string{"package main", "", fmt.Sprintf("func func%d() {", i)}
		for j := 0; j < 60; j++ {
			lines = append(lines, fmt.Sprintf("	line%d := %d", j, j))
		}
		lines = append(lines, "}")
		content := strings.Join(lines, "\n")
		os.WriteFile(testFile, []byte(content), 0644)
	}

	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
		},
	}

	parallelAnalyzer := golang.NewParallelAnalyzer(config, 4)
	scanner := golang.NewFileScanner()
	files, _ := scanner.Scan(context.Background(), tmpDir)

	start := time.Now()
	results := parallelAnalyzer.AnalyzeFiles(context.Background(), files, config)
	elapsed := time.Since(start)

	if len(results) == 0 {
		t.Error("Expected results from parallel analysis")
	}

	t.Logf("Parallel analysis of %d files took %v", len(files), elapsed)
}

func TestIntegrationCrossFileAnalysis(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	os.WriteFile(mainFile, []byte(`package main

func main() {
	foo()
	bar()
}

func foo() {
	_ = 1
}
`), 0644)

	utilsFile := filepath.Join(tmpDir, "utils.go")
	os.WriteFile(utilsFile, []byte(`package main

func bar() {
	_ = 2
}

func unused() {
	_ = 3
}
`), 0644)

	crossFileAnalyzer := golang.NewCrossFileAnalyzer()
	err := crossFileAnalyzer.AnalyzeDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Cross file analysis failed: %v", err)
	}

	unusedResults := crossFileAnalyzer.FindUnusedFunctions()
	if len(unusedResults) == 0 {
		t.Error("Expected to find unused function 'unused'")
	}
}

func TestIntegrationSimilarityDetection(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	os.WriteFile(file1, []byte(`package main

func processX() {
	if a > 0 {
		for i := 0; i < 10; i++ {
			if i > 5 {
				_ = i
			}
		}
	}
}
`), 0644)

	file2 := filepath.Join(tmpDir, "file2.go")
	os.WriteFile(file2, []byte(`package main

func processY() {
	if b > 0 {
		for j := 0; j < 10; j++ {
			if j > 5 {
				_ = j
			}
		}
	}
}
`), 0644)

	similarityAnalyzer := golang.NewSimilarityAnalyzer()
	results, err := similarityAnalyzer.AnalyzeDirectory(context.Background(), tmpDir, 0.7)
	if err != nil {
		t.Fatalf("Similarity analysis failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to find similar code patterns")
	}
}

func TestIntegrationProfiling(t *testing.T) {
	stats := profiling.GetStats()
	if stats.NumCPU == 0 {
		t.Error("Expected NumCPU to be > 0")
	}
	if stats.NumGoroutine == 0 {
		t.Error("Expected NumGoroutine to be > 0")
	}
}

func TestIntegrationConfigHierarchy(t *testing.T) {
	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 100,
			},
		},
	}

	analyzer := golang.NewAnalyzer(config)
	if analyzer == nil {
		t.Error("Failed to create analyzer with custom config")
	}
}

func TestIntegrationLargeScale(t *testing.T) {
	tmpDir := t.TempDir()

	numFiles := 100
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.go", i))
		content := fmt.Sprintf(`package main

const Size%d = 100

func process%d() {
	for i := 0; i < Size%d; i++ {
		_ = i * 2
	}
}
`, i, i, i)
		os.WriteFile(testFile, []byte(content), 0644)
	}

	scanner := golang.NewFileScanner()
	files, _ := scanner.Scan(context.Background(), tmpDir)

	if len(files) != numFiles {
		t.Errorf("Expected %d files, got %d", numFiles, len(files))
	}

	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 10,
			},
		},
	}

	start := time.Now()
	parallelAnalyzer := golang.NewParallelAnalyzer(config, 0)
	results := parallelAnalyzer.AnalyzeFiles(context.Background(), files, config)
	elapsed := time.Since(start)

	t.Logf("Analyzed %d files in %v, found %d issues", len(files), elapsed, len(results))

	if elapsed > 30*time.Second {
		t.Logf("Analysis took longer than expected: %v", elapsed)
	}
}

func TestIntegrationCLIIntegration(t *testing.T) {
	t.Skip("CLI integration test requires binary to be built and configured properly")
}

func BenchmarkIntegrationLarge(b *testing.B) {
	tmpDir := b.TempDir()

	numFiles := 200
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.go", i))
		content := fmt.Sprintf(`package main

func function%d() {
	for i := 0; i < 50; i++ {
		if i > 25 {
			for j := 0; j < 25; j++ {
				if j > 12 {
					_ = i + j
				}
			}
		}
	}
}
`, i)
		os.WriteFile(testFile, []byte(content), 0644)
	}

	scanner := golang.NewFileScanner()
	files, _ := scanner.Scan(context.Background(), tmpDir)

	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parallelAnalyzer := golang.NewParallelAnalyzer(config, 0)
		parallelAnalyzer.AnalyzeFiles(context.Background(), files, config)
	}
}

func BenchmarkIntegrationCrossFile(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 50; i++ {
		mainFile := filepath.Join(tmpDir, fmt.Sprintf("main%d.go", i))
		os.WriteFile(mainFile, []byte(fmt.Sprintf(`package main

func main%d() {
	func%d_a()
	func%d_b()
}

func func%d_a() {
	_ = 1
}

func func%d_b() {
	_ = 2
}

func unused%d() {
	_ = 3
}
`, i, i, i, i, i, i)), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crossFileAnalyzer := golang.NewCrossFileAnalyzer()
		crossFileAnalyzer.AnalyzeDirectory(context.Background(), tmpDir)
	}
}
