package golang_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func benchmarkConfig() core.Config {
	return core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{Enabled: true, MaxLines: 50},
			FileSize:     core.FileSizeConfig{Enabled: true, MaxLines: 500},
			Overcommenting: core.OvercommentingConfig{
				Enabled: true, MaxCommentRatio: 0.3, CheckRedundant: true, CheckDocCoverage: true,
			},
			OrphanedCode: core.OrphanedCodeConfig{
				Enabled: true, CheckUnusedFunctions: true, CheckUnusedVariables: true,
				CheckUnreachableCode: true, CheckDeadImports: true,
			},
		},
		Output:   core.OutputConfig{Format: "console", Verbose: false},
		Language: core.LanguageConfig{Go: core.GoConfig{IgnoreTests: false}},
	}
}

func BenchmarkNewParser(b *testing.B) {
	config := benchmarkConfig()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = golang.NewParser(config)
	}
}

func BenchmarkParser_ParseFile_Small(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "small.go")
	os.WriteFile(testFile, []byte("package main\nfunc main() { println(1) }"), 0644)

	config := benchmarkConfig()
	parser := golang.NewParser(config)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.ParseFile(ctx, testFile)
	}
}

func BenchmarkParser_ParseFile_Medium(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "medium.go")

	var sb strings.Builder
	sb.WriteString("package main\n\n")
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&sb, "func function%d() {\n", i)
		for j := 0; j < 10; j++ {
			fmt.Fprintf(&sb, "\tx%d := %d\n", j, j)
		}
		sb.WriteString("}\n\n")
	}
	os.WriteFile(testFile, []byte(sb.String()), 0644)

	config := benchmarkConfig()
	parser := golang.NewParser(config)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.ParseFile(ctx, testFile)
	}
}

func BenchmarkParser_ParseFile_Large(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "large.go")

	var sb strings.Builder
	sb.WriteString("package main\n\n")
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&sb, "func function%d() {\n", i)
		for j := 0; j < 20; j++ {
			fmt.Fprintf(&sb, "\tx%d := %d\n", j, j)
		}
		sb.WriteString("}\n\n")
	}
	os.WriteFile(testFile, []byte(sb.String()), 0644)

	config := benchmarkConfig()
	parser := golang.NewParser(config)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.ParseFile(ctx, testFile)
	}
}
