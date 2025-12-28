package config

import "github.com/agentlint/agentlint/internal/core"

// DefaultConfig returns the default configuration for AgentLint
func DefaultConfig() core.Config {
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
				Enabled:           true,
				MaxCommentRatio:   0.3,
				CheckRedundant:    true,
				CheckDocCoverage:  true,
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
			Format: "console",
			Verbose: false,
		},
		Language: core.LanguageConfig{
			Go: core.GoConfig{
				IgnoreTests: false,
			},
		},
	}
}