package reactnative

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/reactnative/rules"
)

// Analyzer implements the core.Analyzer interface for React Native (JS/TS/JSX/TSX)
type Analyzer struct {
	parser     *Parser
	rules      []core.Rule
	lineRules  []rules.LineCheckRule
}

// NewAnalyzer creates a new React Native analyzer
func NewAnalyzer(config core.Config) *Analyzer {
	parser := NewParser(config)

	rulesList := []core.Rule{
		rules.NewLargeFunctionRule(config),
		rules.NewLargeFileRule(config),
		rules.NewOvercommentingRule(config),
		rules.NewUnusedFunctionRule(config),
		rules.NewUnusedVariableRule(config),
		rules.NewUnreachableCodeRule(config),
		rules.NewDeadImportRule(config),
	}

	lineRulesList := []rules.LineCheckRule{
		rules.NewInlineStyleRule(config),
		rules.NewAnonymousFunctionInJSXRule(config),
		rules.NewConsoleLogRule(config),
		rules.NewDeprecatedLifecycleRule(config),
		rules.NewHardcodedDimensionRule(config),
		rules.NewDirectStateMutationRule(config),
	}

	return &Analyzer{
		parser:    parser,
		rules:     rulesList,
		lineRules: lineRulesList,
	}
}

// Analyze analyzes a React Native file and returns results
func (a *Analyzer) Analyze(ctx context.Context, filePath string, config core.Config) ([]core.Result, error) {
	parsed, err := a.parser.ParseFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	fileMetrics := a.parser.CalculateFileMetrics(ctx, filePath, parsed)
	functionMetrics := a.parser.CalculateFunctionMetrics(ctx, parsed)

	results := make([]core.Result, 0, 16)
	results = a.applyFileRules(ctx, results, fileMetrics, filePath, config)
	results = a.applyFunctionRules(ctx, results, functionMetrics, filePath, config)
	results = a.applyLineRules(ctx, results, parsed, filePath, config)

	return results, nil
}

func (a *Analyzer) applyLineRules(ctx context.Context, results []core.Result, parsed *ParsedFile, filePath string, config core.Config) []core.Result {
	for lineNum, line := range parsed.Lines {
		for _, rule := range a.lineRules {
			if result := rule.CheckLine(line, lineNum+1); result != nil {
				result.FilePath = filePath
				results = append(results, *result)
			}
		}
	}
	return results
}

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
	return []string{".js", ".jsx", ".ts", ".tsx"}
}

// Name returns the name of this analyzer
func (a *Analyzer) Name() string {
	return "reactnative"
}

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

func isFunctionRule(rule core.Rule) bool {
	return strings.Contains(rule.ID(), "function") ||
		strings.Contains(rule.ID(), "unused") ||
		strings.Contains(rule.ID(), "unreachable")
}

// FileScanner scans directories for React Native files
type FileScanner struct {
	ignoreDirs []string
}

func NewFileScanner() *FileScanner {
	return &FileScanner{
		ignoreDirs: []string{
			".git",
			"node_modules",
			"vendor",
			".vscode",
			".idea",
			"dist",
			"build",
			".next",
			"coverage",
			".expo",
			"android",
			"ios",
		},
	}
}

func (s *FileScanner) Scan(ctx context.Context, rootPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, ignoreDir := range s.ignoreDirs {
				if info.Name() == ignoreDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
