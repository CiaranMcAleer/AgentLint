package rules

import (
	"context"
	"fmt"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

// OvercommentingRule detects overcommented code
type OvercommentingRule struct {
	config core.Config
}

// NewOvercommentingRule creates a new overcommenting rule
func NewOvercommentingRule(config core.Config) *OvercommentingRule {
	return &OvercommentingRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *OvercommentingRule) ID() string {
	return "overcommenting"
}

// Name returns the name of this rule
func (r *OvercommentingRule) Name() string {
	return "Overcommenting"
}

// Description returns a description of this rule
func (r *OvercommentingRule) Description() string {
	return "Detects code with excessive comments"
}

// Category returns the category of this rule
func (r *OvercommentingRule) Category() core.RuleCategory {
	return core.CategoryComments
}

// Severity returns the severity of violations of this rule
func (r *OvercommentingRule) Severity() core.Severity {
	return core.SeverityInfo
}

// Check checks if code violates this rule
func (r *OvercommentingRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	maxRatio := config.Rules.Overcommenting.MaxCommentRatio

	switch n := node.(type) {
	case *FileMetrics:
		if n.CommentRatio > maxRatio {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       1,
				Message:    fmt.Sprintf("File has too many comments (ratio: %.2f, max: %.2f)", n.CommentRatio, maxRatio),
				Suggestion: "Consider reducing comments or ensuring they add meaningful information",
			}
		}
	}

	return nil
}
