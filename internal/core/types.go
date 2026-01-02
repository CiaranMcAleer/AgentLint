package core

import "context"

// Result represents a finding from a rule
type Result struct {
	RuleID     string `json:"rule_id"`
	RuleName   string `json:"rule_name"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	FilePath   string `json:"file_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// RuleCategory defines the category of a rule
type RuleCategory string

const (
	CategorySize     RuleCategory = "size"
	CategoryComments RuleCategory = "comments"
	CategoryOrphaned RuleCategory = "orphaned"
)

// Severity defines the severity level of a result
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Analyzer interface for language-specific implementations
type Analyzer interface {
	Analyze(ctx context.Context, filePath string, config Config) ([]Result, error)
	SupportedExtensions() []string
	Name() string
}

// Rule interface for individual detection rules
type Rule interface {
	ID() string
	Name() string
	Description() string
	Category() RuleCategory
	Severity() Severity
	Check(ctx context.Context, node interface{}, config Config) *Result
}

// Config represents the configuration for AgentLint
type Config struct {
	Rules    RulesConfig    `yaml:"rules"`
	Output   OutputConfig   `yaml:"output"`
	Language LanguageConfig `yaml:"language"`
}

// RulesConfig contains configuration for all rules
type RulesConfig struct {
	FunctionSize   FunctionSizeConfig   `yaml:"functionSize"`
	FileSize       FileSizeConfig       `yaml:"fileSize"`
	Overcommenting OvercommentingConfig `yaml:"overcommenting"`
	OrphanedCode   OrphanedCodeConfig   `yaml:"orphanedCode"`
}

// FunctionSizeConfig contains configuration for function size rules
type FunctionSizeConfig struct {
	Enabled  bool `yaml:"enabled"`
	MaxLines int  `yaml:"maxLines"`
}

// FileSizeConfig contains configuration for file size rules
type FileSizeConfig struct {
	Enabled  bool `yaml:"enabled"`
	MaxLines int  `yaml:"maxLines"`
}

// OvercommentingConfig contains configuration for comment analysis rules
type OvercommentingConfig struct {
	Enabled           bool    `yaml:"enabled"`
	MaxCommentRatio   float64 `yaml:"maxCommentRatio"`
	CheckRedundant    bool    `yaml:"checkRedundant"`
	CheckDocCoverage  bool    `yaml:"checkDocCoverage"`
}

// OrphanedCodeConfig contains configuration for orphaned code detection
type OrphanedCodeConfig struct {
	Enabled              bool `yaml:"enabled"`
	CheckUnusedFunctions bool `yaml:"checkUnusedFunctions"`
	CheckUnusedVariables bool `yaml:"checkUnusedVariables"`
	CheckUnreachableCode bool `yaml:"checkUnreachableCode"`
	CheckDeadImports     bool `yaml:"checkDeadImports"`
}

// OutputConfig contains configuration for output formatting
type OutputConfig struct {
	Format  string `yaml:"format"` // console, json
	Verbose bool   `yaml:"verbose"`
}

// LanguageConfig contains language-specific configuration
type LanguageConfig struct {
	Go          GoConfig          `yaml:"go"`
	Python      PythonConfig      `yaml:"python"`
	ReactNative ReactNativeConfig `yaml:"reactnative"`
}

// GoConfig contains Go-specific configuration
type GoConfig struct {
	IgnoreTests bool `yaml:"ignoreTests"`
}

// PythonConfig contains Python-specific configuration
type PythonConfig struct {
	IgnoreTests bool `yaml:"ignoreTests"`
}

// ReactNativeConfig contains React Native/JavaScript/TypeScript configuration
type ReactNativeConfig struct {
	IgnoreTests bool `yaml:"ignoreTests"`
}