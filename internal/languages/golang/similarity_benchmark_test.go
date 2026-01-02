package golang_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func BenchmarkNewSimilarityAnalyzer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = golang.NewSimilarityAnalyzer()
	}
}

func BenchmarkSimilarityAnalyzer_AnalyzeDirectory(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		tmpDir := b.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte(`package main
func process() { if x > 0 { for i := 0; i < 10; i++ { _ = i } } }`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte(`package main
func handle() { if y > 0 { for j := 0; j < 10; j++ { _ = j } } }`), 0644)

		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			analyzer := golang.NewSimilarityAnalyzer()
			_, _ = analyzer.AnalyzeDirectory(ctx, tmpDir, 0.8)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		tmpDir := b.TempDir()
		for i := 0; i < 10; i++ {
			content := fmt.Sprintf(`package main
func process%d() {
	if x%d > 0 {
		for i := 0; i < 10; i++ {
			if i > 5 { _ = i }
		}
	}
}`, i, i)
			os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)), []byte(content), 0644)
		}

		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			analyzer := golang.NewSimilarityAnalyzer()
			_, _ = analyzer.AnalyzeDirectory(ctx, tmpDir, 0.8)
		}
	})
}
