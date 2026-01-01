package golang

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func TestParallelAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 10; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.go", i))
		lines := []string{"package main", "", fmt.Sprintf("func function%d() {", i)}
		for j := 0; j < 60; j++ {
			lines = append(lines, fmt.Sprintf("	line%d := %d", j, j))
		}
		lines = append(lines, "}", "", fmt.Sprintf("func helper%d() {", i), "	_ = 1 + 1", "}")
		content := strings.Join(lines, "\n")
		os.WriteFile(testFile, []byte(content), 0644)
	}

	config := setupTestConfigForParallel()
	analyzer := NewParallelAnalyzer(config, 4)

	files, err := NewFileScanner().Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}

	results := analyzer.AnalyzeFiles(context.Background(), files, config)

	if len(results) == 0 {
		t.Error("Expected results from parallel analysis")
	}

	if analyzer.WorkerCount() != 4 {
		t.Errorf("Expected 4 workers, got %d", analyzer.WorkerCount())
	}
}

func setupTestConfigForParallel() core.Config {
	return core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
			FileSize: core.FileSizeConfig{
				Enabled:  true,
				MaxLines: 500,
			},
			Overcommenting: core.OvercommentingConfig{
				Enabled:          true,
				MaxCommentRatio:  0.3,
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
}

func TestASTCache(t *testing.T) {
	cache := NewASTCache(0)

	if cache.Size() != 0 {
		t.Errorf("Expected empty cache, got size %d", cache.Size())
	}

	stats := cache.Stats()
	if stats.Entries != 0 {
		t.Errorf("Expected 0 entries in stats, got %d", stats.Entries)
	}

	cache.InvalidateAll()
	if cache.Size() != 0 {
		t.Errorf("Expected empty cache after invalidation, got size %d", cache.Size())
	}
}

func TestCrossFileAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	os.WriteFile(mainFile, []byte(`package main

func main() {
	foo()
	bar()
}

func foo() {
	_ = 1 + 1
}
`), 0644)

	utilsFile := filepath.Join(tmpDir, "utils.go")
	os.WriteFile(utilsFile, []byte(`package main

func bar() {
	_ = 2 + 2
}

func unused() {
	_ = 3 + 3
}
`), 0644)

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	unusedResults := analyzer.FindUnusedFunctions()
	if len(unusedResults) == 0 {
		t.Error("Expected to find unused function 'unused'")
	}

	callGraph := analyzer.GetCallGraph()
	if len(callGraph) == 0 {
		t.Error("Expected call graph to have entries")
	}
}

func TestSimilarityAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	os.WriteFile(file1, []byte(`package main

func processData() {
	if x > 0 {
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

func handleData() {
	if y > 0 {
		for j := 0; j < 10; j++ {
			if j > 5 {
				_ = j
			}
		}
	}
}
`), 0644)

	analyzer := NewSimilarityAnalyzer()
	results, err := analyzer.AnalyzeDirectory(context.Background(), tmpDir, 0.8)
	if err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to find similar code patterns")
	}
}

func BenchmarkLargeAnalysis(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 100; i++ {
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

	scanner := NewFileScanner()
	files, _ := scanner.Scan(context.Background(), tmpDir)

	config := setupTestConfigForParallel()
	analyzer := NewParallelAnalyzer(config, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeFiles(context.Background(), files, config)
	}
}
