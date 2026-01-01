package golang_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func setupTestConfig() core.Config {
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
	}
}

func BenchmarkAnalyzerSingleFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func largeFunction() {
	for i := 0; i < 100; i++ {
		if i > 50 {
			for j := 0; j < 50; j++ {
				if j > 25 {
					_ = i + j
				}
			}
		}
	}
}

func smallFunction() {
	_ = 1 + 1
}
`
	os.WriteFile(testFile, []byte(content), 0644)

	config := setupTestConfig()
	analyzer := golang.NewAnalyzer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := analyzer.Analyze(ctx, testFile, config)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkAnalyzerMultipleFiles(b *testing.B) {
	tmpDir := b.TempDir()
	numFiles := 50

	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.go", i))
		content := fmt.Sprintf(`package main

func function%d() {
	for i := 0; i < 100; i++ {
		if i > 50 {
			for j := 0; j < 50; j++ {
				if j > 25 {
					_ = i + j
				}
			}
		}
	}
}

func helper%d() {
	_ = 1 + 1
}
`, i, i)
		os.WriteFile(testFile, []byte(content), 0644)
	}

	scanner := golang.NewFileScanner()
	files, err := scanner.Scan(context.Background(), tmpDir)
	if err != nil {
		b.Fatalf("Failed to scan files: %v", err)
	}

	config := setupTestConfig()
	analyzer := golang.NewAnalyzer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		for _, file := range files {
			_, err := analyzer.Analyze(ctx, file, config)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	}
}

func BenchmarkRuleLargeFunction(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func veryLargeFunction() {
	line1 := 1
	line2 := 2
	line3 := 3
	line4 := 4
	line5 := 5
	line6 := 6
	line7 := 7
	line8 := 8
	line9 := 9
	line10 := 10
	line11 := 11
	line12 := 12
	line13 := 13
	line14 := 14
	line15 := 15
	line16 := 16
	line17 := 17
	line18 := 18
	line19 := 19
	line20 := 20
	line21 := 21
	line22 := 22
	line23 := 23
	line24 := 24
	line25 := 25
	line26 := 26
	line27 := 27
	line28 := 28
	line29 := 29
	line30 := 30
	line31 := 31
	line32 := 32
	line33 := 33
	line34 := 34
	line35 := 35
	line36 := 36
	line37 := 37
	line38 := 38
	line39 := 39
	line40 := 40
	line41 := 41
	line42 := 42
	line43 := 43
	line44 := 44
	line45 := 45
	line46 := 46
	line47 := 47
	line48 := 48
	line49 := 49
	line50 := 50
	line51 := 51
}
`
	os.WriteFile(testFile, []byte(content), 0644)

	config := setupTestConfig()
	analyzer := golang.NewAnalyzer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		results, err := analyzer.Analyze(ctx, testFile, config)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
		_ = results
	}
}

func BenchmarkRuleOvercommenting(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

// This is a comment
// Another comment
// Yet another comment
// Adding more comments
// Even more comments
// Comments everywhere
// More comments here
// Keep commenting
// Comments galore
// Final comment

// This function does something
func doSomething() {
	// Set x to 1
	x := 1
	// Increment x
	x = x + 1
	// Return x
	return x
}
`
	os.WriteFile(testFile, []byte(content), 0644)

	config := setupTestConfig()
	analyzer := golang.NewAnalyzer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		results, err := analyzer.Analyze(ctx, testFile, config)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
		_ = results
	}
}

func BenchmarkRuleUnusedFunction(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

// unusedFunction is never called
func unusedFunction() {
	return
}

// anotherUnused is also never called
func anotherUnused() {
	return
}

// usedFunction is called from main
func usedFunction() {
	return
}

func main() {
	usedFunction()
}
`
	os.WriteFile(testFile, []byte(content), 0644)

	config := setupTestConfig()
	analyzer := golang.NewAnalyzer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		results, err := analyzer.Analyze(ctx, testFile, config)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
		_ = results
	}
}
