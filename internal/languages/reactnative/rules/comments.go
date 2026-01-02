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

func NewOvercommentingRule(config core.Config) *OvercommentingRule {
	return &OvercommentingRule{config: config}
}

func (r *OvercommentingRule) ID() string          { return "overcommenting" }
func (r *OvercommentingRule) Name() string        { return "Overcommenting" }
func (r *OvercommentingRule) Description() string { return "Detects code with excessive comments" }
func (r *OvercommentingRule) Category() core.RuleCategory { return core.CategoryComments }
func (r *OvercommentingRule) Severity() core.Severity     { return core.SeverityInfo }

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
