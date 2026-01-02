package golang_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func BenchmarkNewASTCache(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = golang.NewASTCache(5 * time.Minute)
	}
}

func BenchmarkASTCache_GetMiss(b *testing.B) {
	cache := golang.NewASTCache(5 * time.Minute)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = cache.Get("/nonexistent/file.go")
	}
}

func BenchmarkASTCache_GetHit(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte("package main\nfunc main() {}"), 0644)

	config := benchmarkConfig()
	parser := golang.NewParser(config)
	ctx := context.Background()

	file, fset, _ := parser.ParseFile(ctx, testFile)
	cache := golang.NewASTCache(5 * time.Minute)
	cache.Set(testFile, file, fset)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = cache.Get(testFile)
	}
}

func BenchmarkASTCache_Set(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte("package main\nfunc main() {}"), 0644)

	config := benchmarkConfig()
	parser := golang.NewParser(config)
	ctx := context.Background()

	file, fset, _ := parser.ParseFile(ctx, testFile)
	cache := golang.NewASTCache(5 * time.Minute)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(testFile, file, fset)
	}
}

func BenchmarkASTCache_Operations(b *testing.B) {
	b.Run("Invalidate", func(b *testing.B) {
		cache := golang.NewASTCache(5 * time.Minute)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache.Invalidate("/path/to/file.go")
		}
	})

	b.Run("InvalidateAll", func(b *testing.B) {
		cache := golang.NewASTCache(5 * time.Minute)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache.InvalidateAll()
		}
	})

	b.Run("Size", func(b *testing.B) {
		cache := golang.NewASTCache(5 * time.Minute)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = cache.Size()
		}
	})

	b.Run("Stats", func(b *testing.B) {
		cache := golang.NewASTCache(5 * time.Minute)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = cache.Stats()
		}
	})
}
