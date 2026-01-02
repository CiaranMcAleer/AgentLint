package golang_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func BenchmarkNewAnalyzer(b *testing.B) {
	config := benchmarkConfig()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = golang.NewAnalyzer(config)
	}
}

func BenchmarkAnalyzer_SupportedExtensions(b *testing.B) {
	config := benchmarkConfig()
	analyzer := golang.NewAnalyzer(config)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = analyzer.SupportedExtensions()
	}
}

func BenchmarkAnalyzer_Name(b *testing.B) {
	config := benchmarkConfig()
	analyzer := golang.NewAnalyzer(config)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = analyzer.Name()
	}
}

func BenchmarkAnalyzer_Analyze_ComplexFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "complex.go")

	var sb strings.Builder
	sb.WriteString("package main\n\n")
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&sb, "func function%d(a, b, c int) (int, error) {\n", i)
		for j := 0; j < 30; j++ {
			fmt.Fprintf(&sb, "\tif a > %d { _ = a + b }\n", j)
		}
		sb.WriteString("\treturn 0, nil\n}\n\n")
	}
	os.WriteFile(testFile, []byte(sb.String()), 0644)

	config := benchmarkConfig()
	analyzer := golang.NewAnalyzer(config)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.Analyze(ctx, testFile, config)
	}
}

func BenchmarkNewParallelAnalyzer(b *testing.B) {
	config := benchmarkConfig()

	b.Run("AutoWorkers", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = golang.NewParallelAnalyzer(config, 0)
		}
	})

	b.Run("4Workers", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = golang.NewParallelAnalyzer(config, 4)
		}
	})
}

func BenchmarkParallelAnalyzer_AnalyzeFiles(b *testing.B) {
	b.Run("10Files", func(b *testing.B) {
		tmpDir := b.TempDir()
		for i := 0; i < 10; i++ {
			os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)),
				[]byte(fmt.Sprintf("package main\nfunc function%d() { for i := 0; i < 50; i++ { _ = i } }", i)), 0644)
		}

		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(context.Background(), tmpDir)
		config := benchmarkConfig()
		analyzer := golang.NewParallelAnalyzer(config, 4)
		ctx := context.Background()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = analyzer.AnalyzeFiles(ctx, files, config)
		}
	})

	b.Run("50Files", func(b *testing.B) {
		tmpDir := b.TempDir()
		for i := 0; i < 50; i++ {
			os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)),
				[]byte(fmt.Sprintf("package main\nfunc function%d() { for i := 0; i < 50; i++ { if i > 25 { _ = i } } }", i)), 0644)
		}

		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(context.Background(), tmpDir)
		config := benchmarkConfig()
		analyzer := golang.NewParallelAnalyzer(config, 4)
		ctx := context.Background()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = analyzer.AnalyzeFiles(ctx, files, config)
		}
	})
}

func BenchmarkParallelAnalyzer_WorkerCount(b *testing.B) {
	config := benchmarkConfig()
	analyzer := golang.NewParallelAnalyzer(config, 4)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = analyzer.WorkerCount()
	}
}
