package languages_test

import (
	"context"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
)

func BenchmarkNewRegistry(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = languages.NewRegistry()
	}
}

func BenchmarkRegistry_Register(b *testing.B) {
	config := core.Config{}
	analyzer := golang.NewAnalyzer(config)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := languages.NewRegistry()
		registry.Register(analyzer)
	}
}

func BenchmarkRegistry_GetAnalyzer(b *testing.B) {
	registry := languages.NewRegistry()
	config := core.Config{}
	analyzer := golang.NewAnalyzer(config)
	registry.Register(analyzer)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetAnalyzer("go")
	}
}

func BenchmarkRegistry_GetAnalyzerByExtension(b *testing.B) {
	registry := languages.NewRegistry()
	config := core.Config{}
	analyzer := golang.NewAnalyzer(config)
	registry.Register(analyzer)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetAnalyzerByExtension(".go")
	}
}

func BenchmarkRegistry_GetAllAnalyzers(b *testing.B) {
	registry := languages.NewRegistry()
	config := core.Config{}
	analyzer := golang.NewAnalyzer(config)
	registry.Register(analyzer)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetAllAnalyzers()
	}
}

func BenchmarkNewFileScanner(b *testing.B) {
	registry := languages.NewRegistry()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = languages.NewFileScanner(registry)
	}
}

func BenchmarkFileScanner_Scan(b *testing.B) {
	registry := languages.NewRegistry()
	config := core.Config{}
	analyzer := golang.NewAnalyzer(config)
	registry.Register(analyzer)

	scanner := languages.NewFileScanner(registry)
	ctx := context.Background()
	tmpDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scanner.Scan(ctx, tmpDir)
	}
}
