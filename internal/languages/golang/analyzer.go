package golang

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang/rules"
)

// Analyzer implements the core.Analyzer interface for Go
type Analyzer struct {
	parser *Parser
	rules  []core.Rule
}

// NewAnalyzer creates a new Go analyzer
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

// Analyze analyzes a Go file and returns results
func (a *Analyzer) Analyze(ctx context.Context, filePath string, config core.Config) ([]core.Result, error) {
	file, fset, err := a.parser.ParseFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	fileMetrics, err := a.parser.CalculateMetrics(ctx, filePath, file)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate metrics for file %s: %w", filePath, err)
	}

	// Pre-allocate results slice with estimated capacity
	results := make([]core.Result, 0, 8)
	results = a.applyFileRules(ctx, results, fileMetrics, config)
	results = a.applyFunctionRules(ctx, results, file, fset, filePath, config)

	return results, nil
}

// applyFileRules applies file-level rules and returns accumulated results
func (a *Analyzer) applyFileRules(ctx context.Context, results []core.Result, metrics *rules.FileMetrics, config core.Config) []core.Result {
	for _, rule := range a.rules {
		if !isRuleEnabled(rule, config) || isFunctionRule(rule) {
			continue
		}
		if result := rule.Check(ctx, metrics, config); result != nil {
			if result.FilePath == "" {
				result.FilePath = metrics.Path
			}
			results = append(results, *result)
		}
	}
	return results
}

// applyFunctionRules applies function-level rules to each function in the file
func (a *Analyzer) applyFunctionRules(ctx context.Context, results []core.Result, file *ast.File, fset *token.FileSet, filePath string, config core.Config) []core.Result {
	for _, rule := range a.rules {
		if !isRuleEnabled(rule, config) || !isFunctionRule(rule) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			funcMetrics, err := a.parser.CalculateFunctionMetrics(ctx, funcDecl, fset, file)
			if err != nil {
				return false
			}
			if result := rule.Check(ctx, funcMetrics, config); result != nil {
				if result.FilePath == "" {
					result.FilePath = filePath
				}
				results = append(results, *result)
			}
			return true
		})
	}
	return results
}

// SupportedExtensions returns the file extensions supported by this analyzer
func (a *Analyzer) SupportedExtensions() []string {
	return []string{".go"}
}

// Name returns the name of this analyzer
func (a *Analyzer) Name() string {
	return "go"
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

// FileScanner scans directories for Go files
type FileScanner struct {
	ignoreDirs []string
}

// NewFileScanner creates a new Go file scanner
func NewFileScanner() *FileScanner {
	return &FileScanner{
		ignoreDirs: []string{
			".git",
			"node_modules",
			"vendor",
			".vscode",
			".idea",
		},
	}
}

// Scan scans a directory for Go files
func (s *FileScanner) Scan(ctx context.Context, rootPath string) ([]string, error) {
	var goFiles []string

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

		// Check if it's a Go file
		if strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	return goFiles, err
}

// ScanForRegistry scans a directory and groups files by language
func (s *FileScanner) ScanForRegistry(ctx context.Context, rootPath string, registry *languages.Registry) (map[string][]string, error) {
	filesByLanguage := make(map[string][]string)

	goFiles, err := s.Scan(ctx, rootPath)
	if err != nil {
		return nil, err
	}

	// Group files by language using the registry
	for _, file := range goFiles {
		ext := filepath.Ext(file)
		if analyzer, exists := registry.GetAnalyzerByExtension(ext); exists {
			language := analyzer.Name()
			filesByLanguage[language] = append(filesByLanguage[language], file)
		}
	}

	return filesByLanguage, nil
}
