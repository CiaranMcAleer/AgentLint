package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

// InlineStyleRule detects inline styles in JSX which cause performance issues
type InlineStyleRule struct {
	config  core.Config
	pattern *regexp.Regexp
}

func NewInlineStyleRule(config core.Config) *InlineStyleRule {
	return &InlineStyleRule{
		config:  config,
		pattern: regexp.MustCompile(`style\s*=\s*\{\s*\{`),
	}
}

func (r *InlineStyleRule) ID() string                    { return "inline-style" }
func (r *InlineStyleRule) Name() string                  { return "Inline Style" }
func (r *InlineStyleRule) Description() string           { return "Detects inline styles that cause unnecessary re-renders" }
func (r *InlineStyleRule) Category() core.RuleCategory   { return core.CategoryPerformance }
func (r *InlineStyleRule) Severity() core.Severity       { return core.SeverityWarning }

func (r *InlineStyleRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	return nil
}

// CheckLine checks a single line for inline styles
func (r *InlineStyleRule) CheckLine(line string, lineNum int) *core.Result {
	if r.pattern.MatchString(line) {
		return &core.Result{
			RuleID:     r.ID(),
			RuleName:   r.Name(),
			Category:   string(r.Category()),
			Severity:   string(r.Severity()),
			Line:       lineNum,
			Message:    "Inline style object creates new reference on every render",
			Suggestion: "Use StyleSheet.create() to define styles outside the component",
		}
	}
	return nil
}

// AnonymousFunctionInJSXRule detects anonymous functions in JSX props
type AnonymousFunctionInJSXRule struct {
	config   core.Config
	patterns []*regexp.Regexp
}

func NewAnonymousFunctionInJSXRule(config core.Config) *AnonymousFunctionInJSXRule {
	return &AnonymousFunctionInJSXRule{
		config: config,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?:on\w+|render\w*)\s*=\s*\{\s*\(\s*\)\s*=>`),
			regexp.MustCompile(`(?:on\w+|render\w*)\s*=\s*\{\s*\w+\s*=>`),
			regexp.MustCompile(`(?:on\w+|render\w*)\s*=\s*\{\s*\([^)]*\)\s*=>`),
			regexp.MustCompile(`(?:on\w+|render\w*)\s*=\s*\{\s*function\s*\(`),
		},
	}
}

func (r *AnonymousFunctionInJSXRule) ID() string                    { return "anonymous-function-jsx" }
func (r *AnonymousFunctionInJSXRule) Name() string                  { return "Anonymous Function in JSX" }
func (r *AnonymousFunctionInJSXRule) Description() string           { return "Detects anonymous functions in JSX props that cause re-renders" }
func (r *AnonymousFunctionInJSXRule) Category() core.RuleCategory   { return core.CategoryPerformance }
func (r *AnonymousFunctionInJSXRule) Severity() core.Severity       { return core.SeverityWarning }

func (r *AnonymousFunctionInJSXRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	return nil
}

// CheckLine checks a single line for anonymous functions in JSX
func (r *AnonymousFunctionInJSXRule) CheckLine(line string, lineNum int) *core.Result {
	for _, pattern := range r.patterns {
		if pattern.MatchString(line) {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       lineNum,
				Message:    "Anonymous function in JSX creates new reference on every render",
				Suggestion: "Extract to a named function or use useCallback hook",
			}
		}
	}
	return nil
}

// ConsoleLogRule detects console.log statements left in code
type ConsoleLogRule struct {
	config  core.Config
	pattern *regexp.Regexp
}

func NewConsoleLogRule(config core.Config) *ConsoleLogRule {
	return &ConsoleLogRule{
		config:  config,
		pattern: regexp.MustCompile(`console\.(log|warn|error|info|debug|trace)\s*\(`),
	}
}

func (r *ConsoleLogRule) ID() string                    { return "console-log" }
func (r *ConsoleLogRule) Name() string                  { return "Console Log" }
func (r *ConsoleLogRule) Description() string           { return "Detects console.log statements that should be removed in production" }
func (r *ConsoleLogRule) Category() core.RuleCategory   { return core.CategoryPerformance }
func (r *ConsoleLogRule) Severity() core.Severity       { return core.SeverityInfo }

func (r *ConsoleLogRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	return nil
}

// CheckLine checks a single line for console.log statements
func (r *ConsoleLogRule) CheckLine(line string, lineNum int) *core.Result {
	// Skip if line is commented out
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
		return nil
	}

	if r.pattern.MatchString(line) {
		return &core.Result{
			RuleID:     r.ID(),
			RuleName:   r.Name(),
			Category:   string(r.Category()),
			Severity:   string(r.Severity()),
			Line:       lineNum,
			Message:    "Console statement should be removed before production",
			Suggestion: "Remove console statement or use a logging library with log levels",
		}
	}
	return nil
}

// DeprecatedLifecycleRule detects deprecated React lifecycle methods
type DeprecatedLifecycleRule struct {
	config             core.Config
	deprecatedMethods  map[string]string
}

func NewDeprecatedLifecycleRule(config core.Config) *DeprecatedLifecycleRule {
	return &DeprecatedLifecycleRule{
		config: config,
		deprecatedMethods: map[string]string{
			"componentWillMount":        "Use componentDidMount or useEffect hook instead",
			"componentWillReceiveProps": "Use getDerivedStateFromProps or useEffect hook instead",
			"componentWillUpdate":       "Use getSnapshotBeforeUpdate or useEffect hook instead",
			"UNSAFE_componentWillMount": "Use componentDidMount or useEffect hook instead",
			"UNSAFE_componentWillReceiveProps": "Use getDerivedStateFromProps or useEffect hook instead",
			"UNSAFE_componentWillUpdate": "Use getSnapshotBeforeUpdate or useEffect hook instead",
		},
	}
}

func (r *DeprecatedLifecycleRule) ID() string                    { return "deprecated-lifecycle" }
func (r *DeprecatedLifecycleRule) Name() string                  { return "Deprecated Lifecycle Method" }
func (r *DeprecatedLifecycleRule) Description() string           { return "Detects deprecated React lifecycle methods" }
func (r *DeprecatedLifecycleRule) Category() core.RuleCategory   { return core.CategoryDeprecated }
func (r *DeprecatedLifecycleRule) Severity() core.Severity       { return core.SeverityWarning }

func (r *DeprecatedLifecycleRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	return nil
}

// CheckLine checks a single line for deprecated lifecycle methods
func (r *DeprecatedLifecycleRule) CheckLine(line string, lineNum int) *core.Result {
	for method, suggestion := range r.deprecatedMethods {
		pattern := regexp.MustCompile(fmt.Sprintf(`\b%s\s*\(`, method))
		if pattern.MatchString(line) {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       lineNum,
				Message:    fmt.Sprintf("Deprecated lifecycle method '%s' detected", method),
				Suggestion: suggestion,
			}
		}
	}
	return nil
}

// MissingKeyPropRule detects .map() calls without key props
type MissingKeyPropRule struct {
	config     core.Config
	mapPattern *regexp.Regexp
}

func NewMissingKeyPropRule(config core.Config) *MissingKeyPropRule {
	return &MissingKeyPropRule{
		config:     config,
		mapPattern: regexp.MustCompile(`\.map\s*\(\s*(?:\([^)]*\)|[\w]+)\s*=>`),
	}
}

func (r *MissingKeyPropRule) ID() string                    { return "missing-key-prop" }
func (r *MissingKeyPropRule) Name() string                  { return "Missing Key Prop" }
func (r *MissingKeyPropRule) Description() string           { return "Detects .map() rendering without key props" }
func (r *MissingKeyPropRule) Category() core.RuleCategory   { return core.CategoryPerformance }
func (r *MissingKeyPropRule) Severity() core.Severity       { return core.SeverityWarning }

func (r *MissingKeyPropRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	return nil
}

// HardcodedDimensionRule detects hardcoded pixel values
type HardcodedDimensionRule struct {
	config  core.Config
	pattern *regexp.Regexp
}

func NewHardcodedDimensionRule(config core.Config) *HardcodedDimensionRule {
	return &HardcodedDimensionRule{
		config:  config,
		pattern: regexp.MustCompile(`(?:width|height|margin\w*|padding\w*|top|bottom|left|right|fontSize)\s*:\s*\d{3,}`),
	}
}

func (r *HardcodedDimensionRule) ID() string                    { return "hardcoded-dimension" }
func (r *HardcodedDimensionRule) Name() string                  { return "Hardcoded Dimension" }
func (r *HardcodedDimensionRule) Description() string           { return "Detects large hardcoded dimension values that may not be responsive" }
func (r *HardcodedDimensionRule) Category() core.RuleCategory   { return core.CategoryStyle }
func (r *HardcodedDimensionRule) Severity() core.Severity       { return core.SeverityInfo }

func (r *HardcodedDimensionRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	return nil
}

// CheckLine checks a single line for hardcoded dimensions
func (r *HardcodedDimensionRule) CheckLine(line string, lineNum int) *core.Result {
	if r.pattern.MatchString(line) {
		return &core.Result{
			RuleID:     r.ID(),
			RuleName:   r.Name(),
			Category:   string(r.Category()),
			Severity:   string(r.Severity()),
			Line:       lineNum,
			Message:    "Large hardcoded dimension value may not be responsive across devices",
			Suggestion: "Consider using Dimensions API, percentage values, or flex layout",
		}
	}
	return nil
}

// DirectStateMutationRule detects direct state mutations
type DirectStateMutationRule struct {
	config   core.Config
	patterns []*regexp.Regexp
}

func NewDirectStateMutationRule(config core.Config) *DirectStateMutationRule {
	return &DirectStateMutationRule{
		config: config,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`this\.state\.\w+\s*=`),
			regexp.MustCompile(`this\.state\.\w+\.push\(`),
			regexp.MustCompile(`this\.state\.\w+\.pop\(`),
			regexp.MustCompile(`this\.state\.\w+\.splice\(`),
			regexp.MustCompile(`this\.state\.\w+\.shift\(`),
			regexp.MustCompile(`this\.state\.\w+\.unshift\(`),
		},
	}
}

func (r *DirectStateMutationRule) ID() string                    { return "direct-state-mutation" }
func (r *DirectStateMutationRule) Name() string                  { return "Direct State Mutation" }
func (r *DirectStateMutationRule) Description() string           { return "Detects direct mutations of React state" }
func (r *DirectStateMutationRule) Category() core.RuleCategory   { return core.CategoryBug }
func (r *DirectStateMutationRule) Severity() core.Severity       { return core.SeverityError }

func (r *DirectStateMutationRule) Check(ctx context.Context, node interface{}, config core.Config) *core.Result {
	return nil
}

// CheckLine checks a single line for direct state mutations
func (r *DirectStateMutationRule) CheckLine(line string, lineNum int) *core.Result {
	for _, pattern := range r.patterns {
		if pattern.MatchString(line) {
			return &core.Result{
				RuleID:     r.ID(),
				RuleName:   r.Name(),
				Category:   string(r.Category()),
				Severity:   string(r.Severity()),
				Line:       lineNum,
				Message:    "Direct state mutation detected - state should be immutable",
				Suggestion: "Use setState() or the state setter function from useState hook",
			}
		}
	}
	return nil
}

// LineCheckRule interface for rules that check individual lines
type LineCheckRule interface {
	core.Rule
	CheckLine(line string, lineNum int) *core.Result
}
