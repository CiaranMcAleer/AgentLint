package rules

import (
	"context"
	"fmt"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

// FunctionMetrics contains metrics about a JavaScript/TypeScript function
type FunctionMetrics struct {
	Name       string
	IsMethod   bool
	ClassName  string
	IsAsync    bool
	IsArrow    bool
	IsExported bool
	LineCount  int
	StartLine  int
}

// FileMetrics contains metrics about a JavaScript/TypeScript file
type FileMetrics struct {
	Path           string
	TotalLines     int
	CodeLines      int
	CommentLines   int
	BlankLines     int
	CommentRatio   float64
	FunctionCount  int
	ImportCount    int
	ClassCount     int
	ComponentCount int
}

// LargeFunctionRule detects functions that are too large
type LargeFunctionRule struct {
	config core.Config
}

func NewLargeFunctionRule(config core.Config) *LargeFunctionRule {
	return &LargeFunctionRule{config: config}
}

func (r *LargeFunctionRule) ID() string          { return "large-function" }
func (r *LargeFunctionRule) Name() string        { return "Large Function" }
func (r *LargeFunctionRule) Description() string { return "Detects functions that exceed the maximum number of lines" }
func (r *LargeFunctionRule) Category() core.RuleCategory { return core.CategorySize }
func (r *LargeFunctionRule) Severity() core.Severity     { return core.SeverityWarning }

func (r *LargeFunctionRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	maxLines := config.Rules.FunctionSize.MaxLines

	switch n := node.(type) {
	case *FunctionMetrics:
		if n.LineCount > maxLines {
			funcType := "Function"
			if n.IsArrow {
				funcType = "Arrow function"
			}
			if n.IsMethod {
				funcType = "Method"
			}
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.StartLine,
				Message:    fmt.Sprintf("%s '%s' is too large (%d lines, max %d)", funcType, n.Name, n.LineCount, maxLines),
				Suggestion: fmt.Sprintf("Consider breaking down %s '%s' into smaller functions", funcType, n.Name),
			}
		}
	}
	return nil
}

// LargeFileRule detects files that are too large
type LargeFileRule struct {
	config core.Config
}

func NewLargeFileRule(config core.Config) *LargeFileRule {
	return &LargeFileRule{config: config}
}

func (r *LargeFileRule) ID() string          { return "large-file" }
func (r *LargeFileRule) Name() string        { return "Large File" }
func (r *LargeFileRule) Description() string { return "Detects files that exceed the maximum number of lines" }
func (r *LargeFileRule) Category() core.RuleCategory { return core.CategorySize }
func (r *LargeFileRule) Severity() core.Severity     { return core.SeverityWarning }

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
				Suggestion: "Consider splitting this file into multiple smaller modules",
			}
		}
	}
	return nil
}
