package core_test

import (
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func BenchmarkResult_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = core.Result{
			RuleID:     "test-rule",
			RuleName:   "Test Rule",
			Category:   "test",
			Severity:   "warning",
			FilePath:   "/path/to/file.go",
			Line:       42,
			Column:     10,
			Message:    "This is a test message",
			Suggestion: "Consider fixing this issue",
		}
	}
}

func BenchmarkConfig_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = core.Config{
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
}

func BenchmarkRulesConfig_Copy(b *testing.B) {
	original := core.RulesConfig{
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
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy := original
		_ = copy
	}
}

func BenchmarkSeverityConstants(b *testing.B) {
	severities := []core.Severity{
		core.SeverityError,
		core.SeverityWarning,
		core.SeverityInfo,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range severities {
			_ = string(s)
		}
	}
}

func BenchmarkCategoryConstants(b *testing.B) {
	categories := []core.RuleCategory{
		core.CategorySize,
		core.CategoryComments,
		core.CategoryOrphaned,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range categories {
			_ = string(c)
		}
	}
}

func BenchmarkResultSlice_Append(b *testing.B) {
	result := core.Result{
		RuleID:   "test-rule",
		RuleName: "Test Rule",
		Category: "test",
		Severity: "warning",
		FilePath: "/path/to/file.go",
		Line:     42,
		Message:  "This is a test message",
	}

	b.Run("Append10", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := make([]core.Result, 0)
			for j := 0; j < 10; j++ {
				results = append(results, result)
			}
			_ = results
		}
	})

	b.Run("Append100", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := make([]core.Result, 0)
			for j := 0; j < 100; j++ {
				results = append(results, result)
			}
			_ = results
		}
	})

	b.Run("PreallocAppend100", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			results := make([]core.Result, 0, 100)
			for j := 0; j < 100; j++ {
				results = append(results, result)
			}
			_ = results
		}
	})
}
