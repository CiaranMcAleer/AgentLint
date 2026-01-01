package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

type AgentLintError struct {
	Code    ErrorCode
	Message string
	Path    string
	Line    int
	Err     error
}

type ErrorCode string

const (
	ErrCodeConfigNotFound   ErrorCode = "E001"
	ErrCodeConfigParse      ErrorCode = "E002"
	ErrCodeConfigValidation ErrorCode = "E003"
	ErrCodeFileNotFound     ErrorCode = "E004"
	ErrCodeFileParse        ErrorCode = "E005"
	ErrCodeAnalysis         ErrorCode = "E006"
	ErrCodeOutput           ErrorCode = "E007"
)

func (e *AgentLintError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (path=%s, line=%d): %v", e.Code, e.Message, e.Path, e.Line, e.Err)
	}
	return fmt.Sprintf("%s: %s (path=%s, line=%d)", e.Code, e.Message, e.Path, e.Line)
}

func (e *AgentLintError) Unwrap() error {
	return e.Err
}

func NewConfigError(code ErrorCode, message string, path string, err error) *AgentLintError {
	return &AgentLintError{
		Code:    code,
		Message: message,
		Path:    path,
		Err:     err,
	}
}

func NewFileError(code ErrorCode, message string, path string, line int, err error) *AgentLintError {
	return &AgentLintError{
		Code:    code,
		Message: message,
		Path:    path,
		Line:    line,
		Err:     err,
	}
}

type ConfigLoader struct {
	globalConfigPaths []string
}

func NewConfigLoader() *ConfigLoader {
	homeDir := os.Getenv("HOME")
	return &ConfigLoader{
		globalConfigPaths: []string{
			"/etc/agentlint.yaml",
			"/etc/agentlint.yml",
			homeDir + "/.agentlint.yaml",
			homeDir + "/.agentlint.yml",
			os.Getenv("AGENTLINT_CONFIG"),
		},
	}
}

func (c *ConfigLoader) FindConfig(path string) (string, error) {
	if path != "" {
		if filepath.IsAbs(path) {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		} else {
			absPath, err := filepath.Abs(path)
			if err == nil {
				if _, err := os.Stat(absPath); err == nil {
					return absPath, nil
				}
			}
		}
		return "", NewConfigError(ErrCodeConfigNotFound, "configuration file not found", path, nil)
	}

	for _, configPath := range c.globalConfigPaths {
		if configPath == "" {
			continue
		}
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", NewConfigError(ErrCodeConfigNotFound, "no configuration file found", "", nil)
}

func (c *ConfigLoader) LoadConfig(path string) (core.Config, error) {
	configPath, err := c.FindConfig(path)
	if err != nil {
		return core.Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return core.Config{}, NewConfigError(ErrCodeConfigNotFound, "failed to read config file", configPath, err)
	}

	var config core.Config
	if err := parseConfig(data, &config); err != nil {
		return core.Config{}, NewConfigError(ErrCodeConfigParse, "failed to parse config", configPath, err)
	}

	return config, nil
}

func parseConfig(data []byte, config *core.Config) error {
	return nil
}

type ConfigHierarchy struct {
	defaults core.Config
	global   core.Config
	project  core.Config
	cli      core.Config
}

func NewConfigHierarchy() *ConfigHierarchy {
	return &ConfigHierarchy{
		defaults: DefaultConfig(),
	}
}

func (h *ConfigHierarchy) Merge() core.Config {
	config := h.defaults

	if h.global.Rules.FunctionSize.Enabled {
		config.Rules.FunctionSize.Enabled = h.global.Rules.FunctionSize.Enabled
	}
	if h.global.Rules.FunctionSize.MaxLines > 0 {
		config.Rules.FunctionSize.MaxLines = h.global.Rules.FunctionSize.MaxLines
	}

	if h.project.Rules.FunctionSize.Enabled {
		config.Rules.FunctionSize.Enabled = h.project.Rules.FunctionSize.Enabled
	}
	if h.project.Rules.FunctionSize.MaxLines > 0 {
		config.Rules.FunctionSize.MaxLines = h.project.Rules.FunctionSize.MaxLines
	}

	if h.cli.Rules.FunctionSize.Enabled {
		config.Rules.FunctionSize.Enabled = h.cli.Rules.FunctionSize.Enabled
	}
	if h.cli.Rules.FunctionSize.MaxLines > 0 {
		config.Rules.FunctionSize.MaxLines = h.cli.Rules.FunctionSize.MaxLines
	}

	return config
}

func (h *ConfigHierarchy) SetGlobal(config core.Config) {
	h.global = config
}

func (h *ConfigHierarchy) SetProject(config core.Config) {
	h.project = config
}

func (h *ConfigHierarchy) SetCLI(config core.Config) {
	h.cli = config
}

func (h *ConfigHierarchy) SetDefaults(config core.Config) {
	h.defaults = config
}

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
