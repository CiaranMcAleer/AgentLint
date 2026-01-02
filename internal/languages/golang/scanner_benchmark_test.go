package golang_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func BenchmarkNewFileScanner(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = golang.NewFileScanner()
	}
}

func BenchmarkFileScanner_Scan_Empty(b *testing.B) {
	tmpDir := b.TempDir()
	scanner := golang.NewFileScanner()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.Scan(ctx, tmpDir)
	}
}

// createFlatGoFiles creates n Go files in tmpDir
func createFlatGoFiles(tmpDir string, n int) {
	for i := 0; i < n; i++ {
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)),
			[]byte(fmt.Sprintf("package main\nfunc f%d() {}", i)), 0644)
	}
}

// createNestedGoFiles creates a nested directory structure with Go files
func createNestedGoFiles(tmpDir string, dirs, filesPerDir int) {
	for i := 0; i < dirs; i++ {
		subDir := filepath.Join(tmpDir, fmt.Sprintf("pkg%d", i))
		os.MkdirAll(subDir, 0755)
		for j := 0; j < filesPerDir; j++ {
			os.WriteFile(filepath.Join(subDir, fmt.Sprintf("file%d.go", j)),
				[]byte(fmt.Sprintf("package pkg%d\nfunc f%d() {}", i, j)), 0644)
		}
	}
}

// runScanBenchmark runs the file scanner benchmark with the given setup
func runScanBenchmark(b *testing.B, setup func(string)) {
	tmpDir := b.TempDir()
	setup(tmpDir)
	scanner := golang.NewFileScanner()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.Scan(ctx, tmpDir)
	}
}

func BenchmarkFileScanner_Scan(b *testing.B) {
	b.Run("10Files", func(b *testing.B) {
		runScanBenchmark(b, func(dir string) { createFlatGoFiles(dir, 10) })
	})
	b.Run("100Files", func(b *testing.B) {
		runScanBenchmark(b, func(dir string) { createFlatGoFiles(dir, 100) })
	})
	b.Run("Nested", func(b *testing.B) {
		runScanBenchmark(b, func(dir string) { createNestedGoFiles(dir, 5, 10) })
	})
}
