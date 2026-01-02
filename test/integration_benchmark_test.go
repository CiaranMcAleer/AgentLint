package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/config"
	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
	"github.com/CiaranMcAleer/AgentLint/internal/output"
)

func setupIntegrationConfig() core.Config {
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

func createTestProject(tmpDir string, fileCount, linesPerFunc int) error {
	for i := 0; i < fileCount; i++ {
		lines := []string{
			"package main",
			"",
			"import \"fmt\"",
			"",
		}

		for j := 0; j < 3; j++ {
			lines = append(lines, fmt.Sprintf("func function%d_%d() {", i, j))
			for k := 0; k < linesPerFunc; k++ {
				lines = append(lines, fmt.Sprintf("\tx%d := %d", k, k))
			}
			lines = append(lines, "}")
			lines = append(lines, "")
		}

		content := strings.Join(lines, "\n")
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// ================ Integration Benchmarks ================

func BenchmarkIntegration_FullPipeline_5Files(b *testing.B) {
	tmpDir := b.TempDir()
	createTestProject(tmpDir, 5, 20)

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := languages.NewRegistry()
		goAnalyzer := golang.NewAnalyzer(cfg)
		registry.Register(goAnalyzer)

		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(ctx, tmpDir)

		var allResults []core.Result
		for _, file := range files {
			results, _ := goAnalyzer.Analyze(ctx, file, cfg)
			allResults = append(allResults, results...)
		}
		_ = allResults
	}
}

func BenchmarkIntegration_FullPipeline_20Files(b *testing.B) {
	tmpDir := b.TempDir()
	createTestProject(tmpDir, 20, 20)

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := languages.NewRegistry()
		goAnalyzer := golang.NewAnalyzer(cfg)
		registry.Register(goAnalyzer)

		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(ctx, tmpDir)

		parallelAnalyzer := golang.NewParallelAnalyzer(cfg, 4)
		_ = parallelAnalyzer.AnalyzeFiles(ctx, files, cfg)
	}
}

func BenchmarkIntegration_FullPipeline_50Files(b *testing.B) {
	tmpDir := b.TempDir()
	createTestProject(tmpDir, 50, 20)

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := languages.NewRegistry()
		goAnalyzer := golang.NewAnalyzer(cfg)
		registry.Register(goAnalyzer)

		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(ctx, tmpDir)

		parallelAnalyzer := golang.NewParallelAnalyzer(cfg, 4)
		_ = parallelAnalyzer.AnalyzeFiles(ctx, files, cfg)
	}
}

func BenchmarkIntegration_WithOutput_Console(b *testing.B) {
	tmpDir := b.TempDir()
	createTestProject(tmpDir, 10, 60) // Large functions to trigger violations

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		goAnalyzer := golang.NewAnalyzer(cfg)
		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(ctx, tmpDir)

		var allResults []core.Result
		for _, file := range files {
			results, _ := goAnalyzer.Analyze(ctx, file, cfg)
			allResults = append(allResults, results...)
		}

		formatter := output.NewConsoleFormatter(false)
		_ = formatter.Format(allResults)
	}
}

func BenchmarkIntegration_WithOutput_JSON(b *testing.B) {
	tmpDir := b.TempDir()
	createTestProject(tmpDir, 10, 60)

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		goAnalyzer := golang.NewAnalyzer(cfg)
		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(ctx, tmpDir)

		var allResults []core.Result
		for _, file := range files {
			results, _ := goAnalyzer.Analyze(ctx, file, cfg)
			allResults = append(allResults, results...)
		}

		formatter := output.NewJSONFormatter(false)
		_ = formatter.Format(allResults)
	}
}

func BenchmarkIntegration_ConfigLoading(b *testing.B) {
	tmpDir := b.TempDir()
	configFile := filepath.Join(tmpDir, "agentlint.yaml")
	os.WriteFile(configFile, []byte(`rules:
  functionSize:
    enabled: true
    maxLines: 50
  fileSize:
    enabled: true
    maxLines: 500
`), 0644)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := config.NewConfigLoader()
		_, _ = loader.FindConfig(configFile)
	}
}

func BenchmarkIntegration_ConfigHierarchy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hierarchy := config.NewConfigHierarchy()
		hierarchy.SetDefaults(config.DefaultConfig())
		hierarchy.SetGlobal(core.Config{
			Rules: core.RulesConfig{
				FunctionSize: core.FunctionSizeConfig{
					Enabled:  true,
					MaxLines: 100,
				},
			},
		})
		hierarchy.SetProject(core.Config{
			Rules: core.RulesConfig{
				FunctionSize: core.FunctionSizeConfig{
					Enabled:  true,
					MaxLines: 75,
				},
			},
		})
		_ = hierarchy.Merge()
	}
}

func BenchmarkIntegration_CrossFileAnalysis(b *testing.B) {
	tmpDir := b.TempDir()

	// Create interconnected files
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main
func main() {
	process()
	handle()
}
`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "process.go"), []byte(`package main
func process() {
	helper()
}
func helper() {}
`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "handle.go"), []byte(`package main
func handle() {
	utility()
}
func utility() {}
func unused() {}
`), 0644)

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer := golang.NewCrossFileAnalyzer()
		_ = analyzer.AnalyzeDirectory(ctx, tmpDir)
		_ = analyzer.FindUnusedFunctions()
	}
}

func BenchmarkIntegration_SimilarityDetection(b *testing.B) {
	tmpDir := b.TempDir()

	// Create files with similar patterns
	for i := 0; i < 5; i++ {
		content := fmt.Sprintf(`package main
func process%d() {
	if x > 0 {
		for i := 0; i < 10; i++ {
			if i > 5 {
				_ = i
			}
		}
	}
}
`, i)
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)), []byte(content), 0644)
	}

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer := golang.NewSimilarityAnalyzer()
		_, _ = analyzer.AnalyzeDirectory(ctx, tmpDir, 0.8)
	}
}

// ================ Parallel vs Sequential Comparison ================

func benchSequential(b *testing.B, files []string, cfg core.Config, ctx context.Context) {
	goAnalyzer := golang.NewAnalyzer(cfg)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var allResults []core.Result
		for _, file := range files {
			results, _ := goAnalyzer.Analyze(ctx, file, cfg)
			allResults = append(allResults, results...)
		}
		_ = allResults
	}
}

func benchParallel(b *testing.B, files []string, cfg core.Config, ctx context.Context, workers int) {
	parallelAnalyzer := golang.NewParallelAnalyzer(cfg, workers)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parallelAnalyzer.AnalyzeFiles(ctx, files, cfg)
	}
}

func BenchmarkIntegration_Sequential_vs_Parallel(b *testing.B) {
	tmpDir := b.TempDir()
	createTestProject(tmpDir, 30, 30)

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	scanner := golang.NewFileScanner()
	files, _ := scanner.Scan(ctx, tmpDir)

	b.Run("Sequential", func(b *testing.B) { benchSequential(b, files, cfg, ctx) })
	b.Run("Parallel_2Workers", func(b *testing.B) { benchParallel(b, files, cfg, ctx, 2) })
	b.Run("Parallel_4Workers", func(b *testing.B) { benchParallel(b, files, cfg, ctx, 4) })
	b.Run("Parallel_8Workers", func(b *testing.B) { benchParallel(b, files, cfg, ctx, 8) })
}

// ================ Memory Pressure Benchmarks ================

func BenchmarkIntegration_MemoryPressure_ManySmallFiles(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 100; i++ {
		content := fmt.Sprintf("package main\nfunc f%d() {}\n", i)
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i)), []byte(content), 0644)
	}

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(ctx, tmpDir)

		parallelAnalyzer := golang.NewParallelAnalyzer(cfg, 4)
		_ = parallelAnalyzer.AnalyzeFiles(ctx, files, cfg)
	}
}

func BenchmarkIntegration_MemoryPressure_FewLargeFiles(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 5; i++ {
		lines := []string{"package main", ""}
		for j := 0; j < 200; j++ {
			lines = append(lines, fmt.Sprintf("func f%d_%d() {", i, j))
			for k := 0; k < 10; k++ {
				lines = append(lines, fmt.Sprintf("\tx%d := %d", k, k))
			}
			lines = append(lines, "}")
		}
		content := strings.Join(lines, "\n")
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("large%d.go", i)), []byte(content), 0644)
	}

	cfg := setupIntegrationConfig()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := golang.NewFileScanner()
		files, _ := scanner.Scan(ctx, tmpDir)

		parallelAnalyzer := golang.NewParallelAnalyzer(cfg, 4)
		_ = parallelAnalyzer.AnalyzeFiles(ctx, files, cfg)
	}
}
