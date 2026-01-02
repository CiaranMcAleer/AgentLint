package rules_test

import (
	"context"
	"go/token"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang/rules"
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
	}
}

// ================ Size Rules ================

func BenchmarkNewLargeFunctionRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewLargeFunctionRule(config)
	}
}

func BenchmarkLargeFunctionRule_Check_NoViolation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewLargeFunctionRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:      "smallFunc",
		LineCount: 30,
		Position:  token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkLargeFunctionRule_Check_Violation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewLargeFunctionRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:      "largeFunc",
		LineCount: 100,
		Position:  token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkNewLargeFileRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewLargeFileRule(config)
	}
}

func BenchmarkLargeFileRule_Check_NoViolation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewLargeFileRule(config)
	ctx := context.Background()

	metrics := &rules.FileMetrics{
		Path:       "/path/to/file.go",
		TotalLines: 300,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkLargeFileRule_Check_Violation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewLargeFileRule(config)
	ctx := context.Background()

	metrics := &rules.FileMetrics{
		Path:       "/path/to/file.go",
		TotalLines: 1000,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

// ================ Comment Rules ================

func BenchmarkNewOvercommentingRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewOvercommentingRule(config)
	}
}

func BenchmarkOvercommentingRule_Check_NoViolation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewOvercommentingRule(config)
	ctx := context.Background()

	metrics := &rules.FileMetrics{
		Path:         "/path/to/file.go",
		TotalLines:   100,
		CodeLines:    80,
		CommentLines: 10,
		CommentRatio: 0.125,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkOvercommentingRule_Check_Violation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewOvercommentingRule(config)
	ctx := context.Background()

	metrics := &rules.FileMetrics{
		Path:         "/path/to/file.go",
		TotalLines:   100,
		CodeLines:    50,
		CommentLines: 40,
		CommentRatio: 0.8,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkNewRedundantCommentRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewRedundantCommentRule(config)
	}
}

// ================ Orphaned Code Rules ================

func BenchmarkNewUnusedFunctionRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewUnusedFunctionRule(config)
	}
}

func BenchmarkUnusedFunctionRule_Check_Exported(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewUnusedFunctionRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:     "ExportedFunc",
		Exported: true,
		Position: token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkUnusedFunctionRule_Check_Unexported(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewUnusedFunctionRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:     "unexportedFunc",
		Exported: false,
		Position: token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkNewUnusedVariableRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewUnusedVariableRule(config)
	}
}

func BenchmarkNewUnreachableCodeRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewUnreachableCodeRule(config)
	}
}

func BenchmarkNewDeadImportRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewDeadImportRule(config)
	}
}

// ================ Complexity Rules ================

func BenchmarkNewParameterCountRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewParameterCountRule(config)
	}
}

func BenchmarkParameterCountRule_Check_NoViolation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewParameterCountRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:           "smallParamFunc",
		ParameterCount: 3,
		Position:       token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkParameterCountRule_Check_Violation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewParameterCountRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:           "manyParamFunc",
		ParameterCount: 10,
		Position:       token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkNewNestingDepthRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewNestingDepthRule(config)
	}
}

func BenchmarkNestingDepthRule_Check_NoViolation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewNestingDepthRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:         "shallowFunc",
		NestingDepth: 2,
		Position:     token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkNestingDepthRule_Check_Violation(b *testing.B) {
	config := setupTestConfig()
	rule := rules.NewNestingDepthRule(config)
	ctx := context.Background()

	metrics := &rules.FunctionMetrics{
		Name:         "deepFunc",
		NestingDepth: 8,
		Position:     token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.Check(ctx, metrics, config)
	}
}

func BenchmarkNewCommentQualityRule(b *testing.B) {
	config := setupTestConfig()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.NewCommentQualityRule(config)
	}
}

// ================ Rule Interface Methods ================

func BenchmarkRule_InterfaceMethods(b *testing.B) {
	config := setupTestConfig()
	rulesList := []core.Rule{
		rules.NewLargeFunctionRule(config),
		rules.NewLargeFileRule(config),
		rules.NewOvercommentingRule(config),
		rules.NewUnusedFunctionRule(config),
		rules.NewParameterCountRule(config),
		rules.NewNestingDepthRule(config),
	}

	methods := []struct {
		name string
		fn   func(core.Rule)
	}{
		{"ID", func(r core.Rule) { _ = r.ID() }},
		{"Name", func(r core.Rule) { _ = r.Name() }},
		{"Description", func(r core.Rule) { _ = r.Description() }},
		{"Category", func(r core.Rule) { _ = r.Category() }},
		{"Severity", func(r core.Rule) { _ = r.Severity() }},
	}

	for _, m := range methods {
		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for _, r := range rulesList {
					m.fn(r)
				}
			}
		})
	}
}

// ================ Batch Rule Checking ================

func BenchmarkAllRules_Check(b *testing.B) {
	config := setupTestConfig()
	ctx := context.Background()

	rulesList := []core.Rule{
		rules.NewLargeFunctionRule(config),
		rules.NewUnusedFunctionRule(config),
		rules.NewParameterCountRule(config),
		rules.NewNestingDepthRule(config),
	}

	metrics := &rules.FunctionMetrics{
		Name:                 "testFunc",
		Exported:             false,
		LineCount:            75,
		ParameterCount:       7,
		NestingDepth:         5,
		CyclomaticComplexity: 15,
		Position:             token.Position{Line: 1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range rulesList {
			_ = r.Check(ctx, metrics, config)
		}
	}
}
