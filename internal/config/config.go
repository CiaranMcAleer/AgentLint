package config

import (
	"os"
	"path/filepath"

	"github.com/agentlint/agentlint/internal/core"
	"gopkg.in/yaml.v3"
)

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

// LoadConfig loads configuration from file, falling back to defaults
func LoadConfig(configPath string) (core.Config, error) {
	config := DefaultConfig()

	if configPath == "" {
		// Try to find config file in current directory
		if _, err := os.Stat("agentlint.yaml"); err == nil {
			configPath = "agentlint.yaml"
		} else if _, err := os.Stat("agentlint.yml"); err == nil {
			configPath = "agentlint.yml"
		} else if _, err := os.Stat(".agentlint.yaml"); err == nil {
			configPath = ".agentlint.yaml"
		} else if _, err := os.Stat(".agentlint.yml"); err == nil {
			configPath = ".agentlint.yml"
		} else {
			// No config file found, use defaults
			return config, nil
		}
	}

	// Make path absolute if it's relative
	if !filepath.IsAbs(configPath) {
		abs, err := filepath.Abs(configPath)
		if err != nil {
			return config, err
		}
		configPath = abs
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config core.Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}