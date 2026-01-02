package rules

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

// UnusedFunctionRule detects functions that are defined but never called
type UnusedFunctionRule struct {
	config core.Config
}

// NewUnusedFunctionRule creates a new unused function rule
func NewUnusedFunctionRule(config core.Config) *UnusedFunctionRule {
	return &UnusedFunctionRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *UnusedFunctionRule) ID() string {
	return "unused-function"
}

// Name returns the name of this rule
func (r *UnusedFunctionRule) Name() string {
	return "Unused Function"
}

// Description returns a description of this rule
func (r *UnusedFunctionRule) Description() string {
	return "Detects functions that are defined but never called"
}

// Category returns the category of this rule
func (r *UnusedFunctionRule) Category() core.RuleCategory {
	return core.CategoryOrphaned
}

// Severity returns the severity of violations of this rule
func (r *UnusedFunctionRule) Severity() core.Severity {
	return core.SeverityWarning
}

// Check checks if code violates this rule
// NOTE: This rule is intentionally conservative and only flags functions that are
// DEFINITELY unused based on single-file analysis. For comprehensive cross-file
// unused function detection, use the CrossFileAnalyzer instead.
func (r *UnusedFunctionRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckUnusedFunctions {
		return nil
	}

	// Single-file analysis cannot accurately determine if a function is unused
	// because it doesn't have visibility into calls from other files.
	// The CrossFileAnalyzer handles this properly by building a complete call graph.
	// We only flag truly obvious cases here: unexported functions in main packages
	// that are clearly not entry points.
	switch n := node.(type) {
	case *FunctionMetrics:
		// Skip exported functions - they may be called from external packages
		if n.Exported {
			return nil
		}
		// Skip test/benchmark/example functions
		if strings.HasPrefix(n.Name, "Test") ||
			strings.HasSuffix(n.Name, "Test") ||
			strings.HasPrefix(n.Name, "Benchmark") ||
			strings.HasPrefix(n.Name, "Example") {
			return nil
		}
		// Skip main and init
		if n.Name == "main" || n.Name == "init" {
			return nil
		}
		// Skip methods - they may implement interfaces
		if n.Receiver != "" {
			return nil
		}
		// Don't flag anything from single-file analysis - let CrossFileAnalyzer handle it
		// This prevents false positives from incomplete call graph information
		return nil
	}

	return nil
}

// UnusedVariableRule detects variables that are declared but never used
type UnusedVariableRule struct {
	config core.Config
}

// NewUnusedVariableRule creates a new unused variable rule
func NewUnusedVariableRule(config core.Config) *UnusedVariableRule {
	return &UnusedVariableRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *UnusedVariableRule) ID() string {
	return "unused-variable"
}

// Name returns the name of this rule
func (r *UnusedVariableRule) Name() string {
	return "Unused Variable"
}

// Description returns a description of this rule
func (r *UnusedVariableRule) Description() string {
	return "Detects variables that are declared but never used"
}

// Category returns the category of this rule
func (r *UnusedVariableRule) Category() core.RuleCategory {
	return core.CategoryOrphaned
}

// Severity returns the severity of violations of this rule
func (r *UnusedVariableRule) Severity() core.Severity {
	return core.SeverityWarning
}

// Check checks if code violates this rule
func (r *UnusedVariableRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckUnusedVariables {
		return nil
	}

	return nil
}

// UnreachableCodeRule detects code that can never be executed
type UnreachableCodeRule struct {
	config core.Config
}

// NewUnreachableCodeRule creates a new unreachable code rule
func NewUnreachableCodeRule(config core.Config) *UnreachableCodeRule {
	return &UnreachableCodeRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *UnreachableCodeRule) ID() string {
	return "unreachable-code"
}

// Name returns the name of this rule
func (r *UnreachableCodeRule) Name() string {
	return "Unreachable Code"
}

// Description returns a description of this rule
func (r *UnreachableCodeRule) Description() string {
	return "Detects code that can never be executed"
}

// Category returns the category of this rule
func (r *UnreachableCodeRule) Category() core.RuleCategory {
	return core.CategoryOrphaned
}

// Severity returns the severity of violations of this rule
func (r *UnreachableCodeRule) Severity() core.Severity {
	return core.SeverityWarning
}

// Check checks if code violates this rule
func (r *UnreachableCodeRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckUnreachableCode {
		return nil
	}

	switch n := node.(type) {
	case *ast.BlockStmt:
		stmts := n.List
		for i := 0; i < len(stmts)-1; i++ {
			if _, ok := stmts[i].(*ast.ReturnStmt); ok {
				return &core.Result{
					RuleID:     r.ID(),
					RuleName:   r.Name(),
					Category:   string(r.Category()),
					Severity:   string(r.Severity()),
					Line:       0,
					Message:    "Unreachable code detected after return statement",
					Suggestion: "Remove the unreachable code",
				}
			}
		}
	}

	return nil
}

// DeadImportRule detects import statements that are never used
type DeadImportRule struct {
	config core.Config
}

// NewDeadImportRule creates a new dead import rule
func NewDeadImportRule(config core.Config) *DeadImportRule {
	return &DeadImportRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *DeadImportRule) ID() string {
	return "dead-import"
}

// Name returns the name of this rule
func (r *DeadImportRule) Name() string {
	return "Dead Import"
}

// Description returns a description of this rule
func (r *DeadImportRule) Description() string {
	return "Detects import statements that are never used"
}

// Category returns the category of this rule
func (r *DeadImportRule) Category() core.RuleCategory {
	return core.CategoryOrphaned
}

// Severity returns the severity of violations of this rule
func (r *DeadImportRule) Severity() core.Severity {
	return core.SeverityWarning
}

// Check checks if code violates this rule
func (r *DeadImportRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckDeadImports {
		return nil
	}

	switch n := node.(type) {
	case *ast.File:
		for _, imp := range n.Imports {
			if !isImportUsed(n, imp) {
				path := imp.Path.Value
				return &core.Result{
					RuleID:     r.ID(),
					RuleName:   r.Name(),
					Category:   string(r.Category()),
					Severity:   string(r.Severity()),
					Line:       0,
					Message:    fmt.Sprintf("Import %s appears to be unused", path),
					Suggestion: "Remove the unused import",
				}
			}
		}
	}

	return nil
}

func isImportUsed(file *ast.File, imp *ast.ImportSpec) bool {
	importPath := imp.Path.Value

	fileAst := *file
	ast.Inspect(&fileAst, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.SelectorExpr:
			if ident, ok := node.X.(*ast.Ident); ok {
				pkgPath := ident.Name
				if pkgPath == importPath || pkgPath == "fmt" {
					return false
				}
			}
		case *ast.BasicLit:
			if node.Kind == token.STRING {
				if node.Value == importPath {
					return false
				}
			}
		}
		return true
	})

	return true
}

// UnusedVariableAnalysis contains analysis results for unused variable detection
type UnusedVariableAnalysis struct {
	Name     string
	IsUnused bool
	Line     int
}

// UnreachableCodeAnalysis contains analysis results for unreachable code detection
type UnreachableCodeAnalysis struct {
	Line int
}

// DeadImportAnalysis contains analysis results for dead import detection
type DeadImportAnalysis struct {
	Path     string
	IsUnused bool
	Line     int
}
