package python

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/python/rules"
)

// Analyzer implements the core.Analyzer interface for Python
type Analyzer struct {
	parser *Parser
	rules  []core.Rule
}

// NewAnalyzer creates a new Python analyzer
func NewAnalyzer(config core.Config) *Analyzer {
	parser := NewParser(config)

	// Initialize rules
	rulesList := []core.Rule{
		rules.NewLargeFunctionRule(config),
		rules.NewLargeFileRule(config),
		rules.NewOvercommentingRule(config),
		rules.NewUnusedFunctionRule(config),
		rules.NewUnusedVariableRule(config),
		rules.NewUnreachableCodeRule(config),
		rules.NewDeadImportRule(config),
	}

	return &Analyzer{
		parser: parser,
		rules:  rulesList,
	}
}

// Analyze analyzes a Python file and returns results
func (a *Analyzer) Analyze(ctx context.Context, filePath string, config core.Config) ([]core.Result, error) {
	parsed, err := a.parser.ParseFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	fileMetrics := a.parser.CalculateFileMetrics(ctx, filePath, parsed)
	functionMetrics := a.parser.CalculateFunctionMetrics(ctx, parsed)

	// Pre-allocate results slice with estimated capacity
	results := make([]core.Result, 0, 8)
	results = a.applyFileRules(ctx, results, fileMetrics, filePath, config)
	results = a.applyFunctionRules(ctx, results, functionMetrics, filePath, config)

	return results, nil
}

// applyFileRules applies file-level rules and returns accumulated results
func (a *Analyzer) applyFileRules(ctx context.Context, results []core.Result, metrics *rules.FileMetrics, filePath string, config core.Config) []core.Result {
	for _, rule := range a.rules {
		if !isRuleEnabled(rule, config) || isFunctionRule(rule) {
			continue
		}
		if result := rule.Check(ctx, metrics, config); result != nil {
			result.FilePath = filePath
			results = append(results, *result)
		}
	}
	return results
}

// applyFunctionRules applies function-level rules to each function in the file
func (a *Analyzer) applyFunctionRules(ctx context.Context, results []core.Result, functionMetrics []*rules.FunctionMetrics, filePath string, config core.Config) []core.Result {
	for _, rule := range a.rules {
		if !isRuleEnabled(rule, config) || !isFunctionRule(rule) {
			continue
		}
		for _, funcMetrics := range functionMetrics {
			if result := rule.Check(ctx, funcMetrics, config); result != nil {
				if result.FilePath == "" {
					result.FilePath = filePath
				}
				results = append(results, *result)
			}
		}
	}
	return results
}

// SupportedExtensions returns the file extensions supported by this analyzer
func (a *Analyzer) SupportedExtensions() []string {
	return []string{".py", ".pyw"}
}

// Name returns the name of this analyzer
func (a *Analyzer) Name() string {
	return "python"
}

// isRuleEnabled checks if a rule is enabled in the configuration
func isRuleEnabled(rule core.Rule, config core.Config) bool {
	switch rule.Category() {
	case core.CategorySize:
		if strings.Contains(rule.ID(), "function") {
			return config.Rules.FunctionSize.Enabled
		}
		if strings.Contains(rule.ID(), "file") {
			return config.Rules.FileSize.Enabled
		}
	case core.CategoryComments:
		return config.Rules.Overcommenting.Enabled
	case core.CategoryOrphaned:
		return config.Rules.OrphanedCode.Enabled
	}
	return true
}

// isFunctionRule checks if a rule applies to functions
func isFunctionRule(rule core.Rule) bool {
	return strings.Contains(rule.ID(), "function") ||
		strings.Contains(rule.ID(), "unused") ||
		strings.Contains(rule.ID(), "unreachable")
}

// FileScanner scans directories for Python files
type FileScanner struct {
	ignoreDirs []string
}

// NewFileScanner creates a new Python file scanner
func NewFileScanner() *FileScanner {
	return &FileScanner{
		ignoreDirs: []string{
			".git",
			"node_modules",
			"vendor",
			".vscode",
			".idea",
			"__pycache__",
			".venv",
			"venv",
			"env",
			".env",
			".tox",
			".eggs",
			"*.egg-info",
			"dist",
			"build",
			".pytest_cache",
			".mypy_cache",
		},
	}
}

// Scan scans a directory for Python files
func (s *FileScanner) Scan(ctx context.Context, rootPath string) ([]string, error) {
	var pythonFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip ignored directories
			for _, ignoreDir := range s.ignoreDirs {
				if info.Name() == ignoreDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if it's a Python file
		if strings.HasSuffix(path, ".py") || strings.HasSuffix(path, ".pyw") {
			pythonFiles = append(pythonFiles, path)
		}

		return nil
	})

	return pythonFiles, err
}
