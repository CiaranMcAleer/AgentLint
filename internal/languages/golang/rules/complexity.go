package rules

import (
	"context"
	"fmt"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

type ParameterCountRule struct {
	config core.Config
}

func NewParameterCountRule(config core.Config) *ParameterCountRule {
	return &ParameterCountRule{
		config: config,
	}
}

func (r *ParameterCountRule) ID() string {
	return "parameter-count"
}

func (r *ParameterCountRule) Name() string {
	return "High Parameter Count"
}

func (r *ParameterCountRule) Description() string {
	return "Detects functions with too many parameters"
}

func (r *ParameterCountRule) Category() core.RuleCategory {
	return core.CategorySize
}

func (r *ParameterCountRule) Severity() core.Severity {
	return core.SeverityWarning
}

func (r *ParameterCountRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	maxParams := 5

	switch n := node.(type) {
	case *FunctionMetrics:
		if n.ParameterCount > maxParams {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.Position.Line,
				Message:    fmt.Sprintf("Function '%s' has too many parameters (%d, max %d)", n.Name, n.ParameterCount, maxParams),
				Suggestion: fmt.Sprintf("Consider grouping parameters into a struct or breaking down function '%s'", n.Name),
			}
		}
	}

	return nil
}

type NestingDepthRule struct {
	config core.Config
}

func NewNestingDepthRule(config core.Config) *NestingDepthRule {
	return &NestingDepthRule{
		config: config,
	}
}

func (r *NestingDepthRule) ID() string {
	return "nesting-depth"
}

func (r *NestingDepthRule) Name() string {
	return "Excessive Nesting Depth"
}

func (r *NestingDepthRule) Description() string {
	return "Detects functions with excessive nesting depth"
}

func (r *NestingDepthRule) Category() core.RuleCategory {
	return core.CategorySize
}

func (r *NestingDepthRule) Severity() core.Severity {
	return core.SeverityWarning
}

func (r *NestingDepthRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	maxDepth := 4

	switch n := node.(type) {
	case *FunctionMetrics:
		if n.NestingDepth > maxDepth {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.Position.Line,
				Message:    fmt.Sprintf("Function '%s' has excessive nesting depth (%d, max %d)", n.Name, n.NestingDepth, maxDepth),
				Suggestion: fmt.Sprintf("Consider flattening the control flow in function '%s' or extracting nested logic", n.Name),
			}
		}
	}

	return nil
}

type CommentQualityRule struct {
	config core.Config
}

func NewCommentQualityRule(config core.Config) *CommentQualityRule {
	return &CommentQualityRule{
		config: config,
	}
}

func (r *CommentQualityRule) ID() string {
	return "comment-quality"
}

func (r *CommentQualityRule) Name() string {
	return "Low Comment Quality"
}

func (r *CommentQualityRule) Description() string {
	return "Detects low-quality comments that don't add value"
}

func (r *CommentQualityRule) Category() core.RuleCategory {
	return core.CategoryComments
}

func (r *CommentQualityRule) Severity() core.Severity {
	return core.SeverityInfo
}

func (r *CommentQualityRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	switch n := node.(type) {
	case *CommentGroup:
		commentText := n.Text

		if isLowQualityComment(commentText) {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.Position.Line,
				Message:    fmt.Sprintf("Low-quality comment detected: %q", truncate(commentText, 50)),
				Suggestion: "Consider improving this comment to explain 'why' rather than 'what'",
			}
		}
	}

	return nil
}

func isLowQualityComment(comment string) bool {
	lowQualityPatterns := []string{
		"todo",
		"fixme",
		"xxx",
		"hack",
		"bug",
		"this is broken",
		"temporary",
	}

	lowerComment := comment
	for _, pattern := range lowQualityPatterns {
		if len(lowerComment) < 200 && containsPattern(lowerComment, pattern) {
			return true
		}
	}

	return false
}

func containsPattern(s, pattern string) bool {
	return len(s) >= len(pattern) && (s == pattern || len(s) > len(pattern) && (s[:len(pattern)] == pattern || containsSubstring(s, pattern)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type ComplexityThresholdRule struct {
	config core.Config
}

func NewComplexityThresholdRule(config core.Config) *ComplexityThresholdRule {
	return &ComplexityThresholdRule{
		config: config,
	}
}

func (r *ComplexityThresholdRule) ID() string {
	return "complexity-threshold"
}

func (r *ComplexityThresholdRule) Name() string {
	return "High Cyclomatic Complexity"
}

func (r *ComplexityThresholdRule) Description() string {
	return "Detects functions with excessive cyclomatic complexity"
}

func (r *ComplexityThresholdRule) Category() core.RuleCategory {
	return core.CategorySize
}

func (r *ComplexityThresholdRule) Severity() core.Severity {
	return core.SeverityWarning
}

func (r *ComplexityThresholdRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	maxComplexity := 10

	switch n := node.(type) {
	case *FunctionMetrics:
		if n.CyclomaticComplexity > maxComplexity {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.Position.Line,
				Message:    fmt.Sprintf("Function '%s' has high cyclomatic complexity (%d, max %d)", n.Name, n.CyclomaticComplexity, maxComplexity),
				Suggestion: fmt.Sprintf("Consider simplifying function '%s' by extracting logic or using early returns", n.Name),
			}
		}
	}

	return nil
}
