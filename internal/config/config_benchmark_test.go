package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/config"
	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func BenchmarkDefaultConfig(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.DefaultConfig()
	}
}

func BenchmarkNewConfigLoader(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.NewConfigLoader()
	}
}

func BenchmarkConfigHierarchy_New(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.NewConfigHierarchy()
	}
}

func BenchmarkConfigHierarchy_Merge(b *testing.B) {
	hierarchy := config.NewConfigHierarchy()
	hierarchy.SetDefaults(config.DefaultConfig())

	globalConfig := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 100,
			},
		},
	}
	hierarchy.SetGlobal(globalConfig)

	projectConfig := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 75,
			},
		},
	}
	hierarchy.SetProject(projectConfig)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hierarchy.Merge()
	}
}

func BenchmarkConfigLoader_FindConfig(b *testing.B) {
	tmpDir := b.TempDir()
	configFile := filepath.Join(tmpDir, "agentlint.yaml")
	os.WriteFile(configFile, []byte("rules:\n  functionSize:\n    enabled: true\n"), 0644)

	loader := config.NewConfigLoader()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.FindConfig(configFile)
	}
}

func BenchmarkAgentLintError_Error(b *testing.B) {
	err := config.NewConfigError(config.ErrCodeConfigNotFound, "test error", "/path/to/file", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkAgentLintError_ErrorWithWrapped(b *testing.B) {
	wrappedErr := os.ErrNotExist
	err := config.NewConfigError(config.ErrCodeConfigNotFound, "test error", "/path/to/file", wrappedErr)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkNewFileError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.NewFileError(config.ErrCodeFileParse, "test error", "/path/to/file", 42, nil)
	}
}

func BenchmarkConfigHierarchy_SetOperations(b *testing.B) {
	cfg := config.DefaultConfig()

	b.Run("SetDefaults", func(b *testing.B) {
		hierarchy := config.NewConfigHierarchy()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			hierarchy.SetDefaults(cfg)
		}
	})

	b.Run("SetGlobal", func(b *testing.B) {
		hierarchy := config.NewConfigHierarchy()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			hierarchy.SetGlobal(cfg)
		}
	})

	b.Run("SetProject", func(b *testing.B) {
		hierarchy := config.NewConfigHierarchy()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			hierarchy.SetProject(cfg)
		}
	})

	b.Run("SetCLI", func(b *testing.B) {
		hierarchy := config.NewConfigHierarchy()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			hierarchy.SetCLI(cfg)
		}
	})
}
