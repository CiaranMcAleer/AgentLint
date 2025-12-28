package rules

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/agentlint/agentlint/internal/core"
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
		
		// Check for redundant comments
		if config.Rules.Overcommenting.CheckRedundant {
			// This would require additional analysis of comment content
			// For now, we'll just check the ratio
		}
		
		// Check for missing documentation on exported functions
		if config.Rules.Overcommenting.CheckDocCoverage {
			// This would be checked in a separate rule or with additional context
		}
	}

	return nil
}

// RedundantCommentRule detects redundant comments
type RedundantCommentRule struct {
	config core.Config
}

// NewRedundantCommentRule creates a new redundant comment rule
func NewRedundantCommentRule(config core.Config) *RedundantCommentRule {
	return &RedundantCommentRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *RedundantCommentRule) ID() string {
	return "redundant-comment"
}

// Name returns the name of this rule
func (r *RedundantCommentRule) Name() string {
	return "Redundant Comment"
}

// Description returns a description of this rule
func (r *RedundantCommentRule) Description() string {
	return "Detects comments that simply restate what the code does"
}

// Category returns the category of this rule
func (r *RedundantCommentRule) Category() core.RuleCategory {
	return core.CategoryComments
}

// Severity returns the severity of violations of this rule
func (r *RedundantCommentRule) Severity() core.Severity {
	return core.SeverityInfo
}

// Check checks if code violates this rule
func (r *RedundantCommentRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.Overcommenting.CheckRedundant {
		return nil
	}

	switch n := node.(type) {
	case *CommentGroup:
		commentText := strings.TrimSpace(n.Text)
		
		// Check for common redundant patterns
		redundantPatterns := []string{
			"increment i",
			"decrement i",
			"return true",
			"return false",
			"check if",
			"loop through",
			"initialize variable",
		}
		
		for _, pattern := range redundantPatterns {
			if strings.Contains(strings.ToLower(commentText), pattern) {
				return &core.Result{
					RuleID:     r.ID(),
					RuleName:   r.Name(),
					Category:   string(r.Category()),
					Severity:   string(r.Severity()),
					Line:       n.Position.Line,
					Message:    fmt.Sprintf("Comment appears to be redundant: %q", commentText),
					Suggestion: "Consider removing this redundant comment or making it more meaningful",
				}
			}
		}
	}

	return nil
}

// MissingDocumentationRule detects missing documentation on exported functions
type MissingDocumentationRule struct {
	config core.Config
}

// NewMissingDocumentationRule creates a new missing documentation rule
func NewMissingDocumentationRule(config core.Config) *MissingDocumentationRule {
	return &MissingDocumentationRule{
		config: config,
	}
}

// ID returns the unique identifier for this rule
func (r *MissingDocumentationRule) ID() string {
	return "missing-documentation"
}

// Name returns the name of this rule
func (r *MissingDocumentationRule) Name() string {
	return "Missing Documentation"
}

// Description returns a description of this rule
func (r *MissingDocumentationRule) Description() string {
	return "Detects exported functions without documentation"
}

// Category returns the category of this rule
func (r *MissingDocumentationRule) Category() core.RuleCategory {
	return core.CategoryComments
}

// Severity returns the severity of violations of this rule
func (r *MissingDocumentationRule) Severity() core.Severity {
	return core.SeverityInfo
}

// Check checks if code violates this rule
func (r *MissingDocumentationRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.Overcommenting.CheckDocCoverage {
		return nil
	}

	switch n := node.(type) {
	case *ast.FuncDecl:
		// Only check exported functions
		if !n.Name.IsExported() {
			return nil
		}
		
		// Check if function has documentation
		if n.Doc == nil || n.Doc.Text() == "" {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       0, // Will be set by caller
				Message:    fmt.Sprintf("Exported function '%s' is missing documentation", n.Name.Name),
				Suggestion: fmt.Sprintf("Add a comment documenting the purpose and behavior of '%s'", n.Name.Name),
			}
		}
	}

	return nil
}

// CommentGroup represents a comment group with position information
type CommentGroup struct {
	Text     string
	Position token.Position
}