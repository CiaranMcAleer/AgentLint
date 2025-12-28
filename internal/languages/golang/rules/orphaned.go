package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentlint/agentlint/internal/core"
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
func (r *UnusedFunctionRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckUnusedFunctions {
		return nil
	}

	switch n := node.(type) {
	case *UnusedFunctionAnalysis:
		if n.IsUnused && !n.IsExported && !strings.HasSuffix(n.Name, "Test") && !strings.HasPrefix(n.Name, "Example") {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.Line,
				Message:    fmt.Sprintf("Function '%s' is defined but never used", n.Name),
				Suggestion: fmt.Sprintf("Consider removing function '%s' or using it somewhere in the codebase", n.Name),
			}
		}
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

	switch n := node.(type) {
	case *UnusedVariableAnalysis:
		if n.IsUnused {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.Line,
				Message:    fmt.Sprintf("Variable '%s' is declared but never used", n.Name),
				Suggestion: fmt.Sprintf("Consider removing variable '%s' or using it somewhere in the code", n.Name),
			}
		}
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
	case *UnreachableCodeAnalysis:
		return &core.Result{
			RuleID:     r.ID(),
			RuleName:   r.Name(),
			Category:   string(r.Category()),
			Severity:   string(r.Severity()),
			Line:       n.Line,
			Message:    "Code after return statement is unreachable",
			Suggestion: "Remove the unreachable code",
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
	case *DeadImportAnalysis:
		if n.IsUnused {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       n.Line,
				Message:    fmt.Sprintf("Import '%s' is never used", n.Path),
				Suggestion: fmt.Sprintf("Remove the unused import '%s'", n.Path),
			}
		}
	}

	return nil
}

// Analysis result types for orphaned code detection

// UnusedFunctionAnalysis contains analysis results for unused function detection
type UnusedFunctionAnalysis struct {
	Name      string
	IsUnused  bool
	IsExported bool
	Line      int
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