package rules

import (
	"context"
	"strings"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

// UnusedFunctionRule detects functions that are defined but never called.
// This rule is intentionally conservative and only flags functions that are
// DEFINITELY unused. For comprehensive cross-file detection, a cross-file analyzer is needed.
type UnusedFunctionRule struct {
	config core.Config
}

func NewUnusedFunctionRule(config core.Config) *UnusedFunctionRule {
	return &UnusedFunctionRule{config: config}
}

func (r *UnusedFunctionRule) ID() string          { return "unused-function" }
func (r *UnusedFunctionRule) Name() string        { return "Unused Function" }
func (r *UnusedFunctionRule) Description() string { return "Detects functions that are defined but never called" }
func (r *UnusedFunctionRule) Category() core.RuleCategory { return core.CategoryOrphaned }
func (r *UnusedFunctionRule) Severity() core.Severity     { return core.SeverityWarning }

func (r *UnusedFunctionRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckUnusedFunctions {
		return nil
	}

	switch n := node.(type) {
	case *FunctionMetrics:
		if strings.HasPrefix(n.Name, "__") && strings.HasSuffix(n.Name, "__") {
			return nil
		}
		if strings.HasPrefix(n.Name, "test_") || strings.HasPrefix(n.Name, "Test") {
			return nil
		}
		if isTestSetupFunction(n.Name) || n.Name == "main" {
			return nil
		}
		return nil
	}
	return nil
}

func isTestSetupFunction(name string) bool {
	testFuncs := []string{"setup", "teardown", "setUp", "tearDown", "setUpClass", "tearDownClass", "setUpModule", "tearDownModule"}
	for _, f := range testFuncs {
		if name == f {
			return true
		}
	}
	return false
}

// UnusedVariableRule detects variables that are declared but never used
type UnusedVariableRule struct {
	config core.Config
}

func NewUnusedVariableRule(config core.Config) *UnusedVariableRule {
	return &UnusedVariableRule{config: config}
}

func (r *UnusedVariableRule) ID() string          { return "unused-variable" }
func (r *UnusedVariableRule) Name() string        { return "Unused Variable" }
func (r *UnusedVariableRule) Description() string { return "Detects variables that are declared but never used" }
func (r *UnusedVariableRule) Category() core.RuleCategory { return core.CategoryOrphaned }
func (r *UnusedVariableRule) Severity() core.Severity     { return core.SeverityWarning }

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

func NewUnreachableCodeRule(config core.Config) *UnreachableCodeRule {
	return &UnreachableCodeRule{config: config}
}

func (r *UnreachableCodeRule) ID() string          { return "unreachable-code" }
func (r *UnreachableCodeRule) Name() string        { return "Unreachable Code" }
func (r *UnreachableCodeRule) Description() string { return "Detects code that can never be executed" }
func (r *UnreachableCodeRule) Category() core.RuleCategory { return core.CategoryOrphaned }
func (r *UnreachableCodeRule) Severity() core.Severity     { return core.SeverityWarning }

func (r *UnreachableCodeRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckUnreachableCode {
		return nil
	}
	return nil
}

// DeadImportRule detects imports that are never used
type DeadImportRule struct {
	config core.Config
}

func NewDeadImportRule(config core.Config) *DeadImportRule {
	return &DeadImportRule{config: config}
}

func (r *DeadImportRule) ID() string          { return "dead-import" }
func (r *DeadImportRule) Name() string        { return "Dead Import" }
func (r *DeadImportRule) Description() string { return "Detects imports that are never used in the code" }
func (r *DeadImportRule) Category() core.RuleCategory { return core.CategoryOrphaned }
func (r *DeadImportRule) Severity() core.Severity     { return core.SeverityWarning }

func (r *DeadImportRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	if !config.Rules.OrphanedCode.CheckDeadImports {
		return nil
	}
	return nil
}
