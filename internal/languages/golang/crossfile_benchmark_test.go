package golang_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func BenchmarkNewCrossFileAnalyzer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = golang.NewCrossFileAnalyzer()
	}
}

func BenchmarkCrossFileAnalyzer_AnalyzeDirectory(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		tmpDir := b.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() { foo() }
func foo() { bar() }
func bar() {}`), 0644)

		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			analyzer := golang.NewCrossFileAnalyzer()
			_ = analyzer.AnalyzeDirectory(ctx, tmpDir)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		tmpDir := b.TempDir()
		for i := 0; i < 10; i++ {
			content := fmt.Sprintf(`package main
func function%d() { helper%d() }
func helper%d() {}`, i, i, i)
			os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)), []byte(content), 0644)
		}

		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			analyzer := golang.NewCrossFileAnalyzer()
			_ = analyzer.AnalyzeDirectory(ctx, tmpDir)
		}
	})
}

func BenchmarkCrossFileAnalyzer_FindUnusedFunctions(b *testing.B) {
	tmpDir := b.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() { used() }
func used() {}
func unused() {}`), 0644)

	analyzer := golang.NewCrossFileAnalyzer()
	ctx := context.Background()
	analyzer.AnalyzeDirectory(ctx, tmpDir)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.FindUnusedFunctions()
	}
}

func BenchmarkCrossFileAnalyzer_GetCallGraph(b *testing.B) {
	tmpDir := b.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() { a() }
func a() { b() }
func b() { c() }
func c() {}`), 0644)

	analyzer := golang.NewCrossFileAnalyzer()
	ctx := context.Background()
	analyzer.AnalyzeDirectory(ctx, tmpDir)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.GetCallGraph()
	}
}
