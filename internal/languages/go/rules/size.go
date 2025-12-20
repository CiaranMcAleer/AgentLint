package rules

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/agentlint/agentlint/internal/core"
)

// LargeFunctionRule detects functions that are too large
type LargeFunctionRule struct {
	config core.Config
}

// NewLargeFunctionRule creates a new large function rule
func NewLargeFunctionRule(config core.Config) *LargeFunctionRule {
	return &LargeFunctionRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *LargeFunctionRule) ID() string {
	return "large-function"
}

// Name returns the name of this rule
func (r *LargeFunctionRule) Name() string {
	return "Large Function"
}

// Description returns a description of this rule
func (r *LargeFunctionRule) Description() string {
	return "Detects functions that exceed the maximum number of lines"
}

// Category returns the category of this rule
func (r *LargeFunctionRule) Category() core.RuleCategory {
	return core.CategorySize
}

// Severity returns the severity of violations of this rule
func (r *LargeFunctionRule) Severity() core.Severity {
	return core.SeverityWarning
}

// Check checks if a function violates this rule
func (r *LargeFunctionRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	maxLines := config.Rules.FunctionSize.MaxLines

	switch n := node.(type) {
	case *FunctionMetrics:
		if n.LineCount > maxLines {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       0, // Will be set by caller
				Message:    fmt.Sprintf("Function '%s' is too large (%d lines, max %d)", n.Name, n.LineCount, maxLines),
				Suggestion: fmt.Sprintf("Consider breaking down function '%s' into smaller functions", n.Name),
			}
		}
	case *ast.FuncDecl:
		// This case is handled by the analyzer using FunctionMetrics
		// We keep this for completeness but it won't be used in the main flow
	}

	return nil
}

// LargeFileRule detects files that are too large
type LargeFileRule struct {
	config core.Config
}

// NewLargeFileRule creates a new large file rule
func NewLargeFileRule(config core.Config) *LargeFileRule {
	return &LargeFileRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *LargeFileRule) ID() string {
	return "large-file"
}

// Name returns the name of this rule
func (r *LargeFileRule) Name() string {
	return "Large File"
}

// Description returns a description of this rule
func (r *LargeFileRule) Description() string {
	return "Detects files that exceed the maximum number of lines"
}

// Category returns the category of this rule
func (r *LargeFileRule) Category() core.RuleCategory {
	return core.CategorySize
}

// Severity returns the severity of violations of this rule
func (r *LargeFileRule) Severity() core.Severity {
	return core.SeverityWarning
}

// Check checks if a file violates this rule
func (r *LargeFileRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	maxLines := config.Rules.FileSize.MaxLines

	switch n := node.(type) {
	case *FileMetrics:
		if n.TotalLines > maxLines {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       1,
				Message:    fmt.Sprintf("File is too large (%d lines, max %d)", n.TotalLines, maxLines),
				Suggestion: "Consider splitting this file into multiple smaller files",
			}
		}
	case *ast.File:
		// This case is handled by the analyzer using FileMetrics
		// We keep this for completeness but it won't be used in the main flow
	}

	return nil
}

// FunctionMetrics contains metrics about a Go function
type FunctionMetrics struct {
	Name                string
	Receiver            string
	Exported            bool
	LineCount           int
	ParameterCount      int
	ReturnCount         int
	CyclomaticComplexity int
	Position            token.Position
}

// FileMetrics contains metrics about a Go file
type FileMetrics struct {
	Path          string
	TotalLines    int
	CodeLines     int
	CommentLines  int
	BlankLines    int
	CommentRatio  float64
	FunctionCount int
	ImportCount   int
	ExportedCount int
}