package rules

import (
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func getTestConfig() core.Config {
	return core.Config{
		Rules: core.RulesConfig{
			FunctionSize:   core.FunctionSizeConfig{MaxLines: 50, Enabled: true},
			FileSize:       core.FileSizeConfig{MaxLines: 500, Enabled: true},
			Overcommenting: core.OvercommentingConfig{MaxCommentRatio: 0.30, Enabled: true},
		},
	}
}

func TestInlineStyleRule_CheckLine(t *testing.T) {
	config := getTestConfig()
	rule := NewInlineStyleRule(config)

	tests := []struct {
		name     string
		line     string
		hasIssue bool
	}{
		{"inline style object", `<View style={{ flex: 1 }}>`, true},
		{"inline style with spaces", `<View style = {{ flex: 1 }}>`, true},
		{"stylesheet reference", `<View style={styles.container}>`, false},
		{"array of styles", `<View style={[styles.container, styles.active]}>`, false},
		{"no style", `<View>`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.CheckLine(tt.line, 1)
			if tt.hasIssue && result == nil {
				t.Errorf("Expected issue for line: %s", tt.line)
			}
			if !tt.hasIssue && result != nil {
				t.Errorf("Unexpected issue for line: %s", tt.line)
			}
		})
	}
}

func TestAnonymousFunctionInJSXRule_CheckLine(t *testing.T) {
	config := getTestConfig()
	rule := NewAnonymousFunctionInJSXRule(config)

	tests := []struct {
		name     string
		line     string
		hasIssue bool
	}{
		{"arrow function no params", `onPress={() => handlePress()}`, true},
		{"arrow function with param", `onPress={e => handlePress(e)}`, true},
		{"arrow function with params", `onPress={(e, i) => handlePress(e, i)}`, true},
		{"anonymous function", `onPress={function() { handlePress(); }}`, true},
		{"named function reference", `onPress={handlePress}`, false},
		{"bound method", `onPress={this.handlePress}`, false},
		{"renderItem arrow", `renderItem={item => <Text>{item}</Text>}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.CheckLine(tt.line, 1)
			if tt.hasIssue && result == nil {
				t.Errorf("Expected issue for line: %s", tt.line)
			}
			if !tt.hasIssue && result != nil {
				t.Errorf("Unexpected issue for line: %s", tt.line)
			}
		})
	}
}

func TestConsoleLogRule_CheckLine(t *testing.T) {
	config := getTestConfig()
	rule := NewConsoleLogRule(config)

	tests := []struct {
		name     string
		line     string
		hasIssue bool
	}{
		{"console.log", `console.log('debug');`, true},
		{"console.warn", `console.warn('warning');`, true},
		{"console.error", `console.error('error');`, true},
		{"console.info", `console.info('info');`, true},
		{"console.debug", `console.debug('debug');`, true},
		{"commented console", `// console.log('debug');`, false},
		{"block comment", `* console.log('debug');`, false},
		{"no console", `const log = 'hello';`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.CheckLine(tt.line, 1)
			if tt.hasIssue && result == nil {
				t.Errorf("Expected issue for line: %s", tt.line)
			}
			if !tt.hasIssue && result != nil {
				t.Errorf("Unexpected issue for line: %s", tt.line)
			}
		})
	}
}

func TestDeprecatedLifecycleRule_CheckLine(t *testing.T) {
	config := getTestConfig()
	rule := NewDeprecatedLifecycleRule(config)

	tests := []struct {
		name     string
		line     string
		hasIssue bool
	}{
		{"componentWillMount", `componentWillMount() {`, true},
		{"componentWillReceiveProps", `componentWillReceiveProps(nextProps) {`, true},
		{"componentWillUpdate", `componentWillUpdate(nextProps, nextState) {`, true},
		{"UNSAFE_componentWillMount", `UNSAFE_componentWillMount() {`, true},
		{"componentDidMount", `componentDidMount() {`, false},
		{"componentDidUpdate", `componentDidUpdate() {`, false},
		{"render", `render() {`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.CheckLine(tt.line, 1)
			if tt.hasIssue && result == nil {
				t.Errorf("Expected issue for line: %s", tt.line)
			}
			if !tt.hasIssue && result != nil {
				t.Errorf("Unexpected issue for line: %s", tt.line)
			}
		})
	}
}

func TestHardcodedDimensionRule_CheckLine(t *testing.T) {
	config := getTestConfig()
	rule := NewHardcodedDimensionRule(config)

	tests := []struct {
		name     string
		line     string
		hasIssue bool
	}{
		{"large width", `width: 400,`, true},
		{"large height", `height: 800,`, true},
		{"large margin", `marginTop: 150,`, true},
		{"large padding", `paddingLeft: 200,`, true},
		{"large fontSize", `fontSize: 100,`, true},
		{"small width", `width: 50,`, false},
		{"percentage", `width: '100%',`, false},
		{"flex", `flex: 1,`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.CheckLine(tt.line, 1)
			if tt.hasIssue && result == nil {
				t.Errorf("Expected issue for line: %s", tt.line)
			}
			if !tt.hasIssue && result != nil {
				t.Errorf("Unexpected issue for line: %s", tt.line)
			}
		})
	}
}

func TestDirectStateMutationRule_CheckLine(t *testing.T) {
	config := getTestConfig()
	rule := NewDirectStateMutationRule(config)

	tests := []struct {
		name     string
		line     string
		hasIssue bool
	}{
		{"direct assignment", `this.state.count = 5;`, true},
		{"array push", `this.state.items.push(item);`, true},
		{"array pop", `this.state.items.pop();`, true},
		{"array splice", `this.state.items.splice(0, 1);`, true},
		{"setState call", `this.setState({ count: 5 });`, false},
		{"local variable", `const items = [];`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.CheckLine(tt.line, 1)
			if tt.hasIssue && result == nil {
				t.Errorf("Expected issue for line: %s", tt.line)
			}
			if !tt.hasIssue && result != nil {
				t.Errorf("Unexpected issue for line: %s", tt.line)
			}
		})
	}
}

func TestInlineStyleRule_ID(t *testing.T) {
	config := getTestConfig()
	rule := NewInlineStyleRule(config)
	if rule.ID() != "inline-style" {
		t.Errorf("Expected ID 'inline-style', got '%s'", rule.ID())
	}
}

func TestAnonymousFunctionInJSXRule_ID(t *testing.T) {
	config := getTestConfig()
	rule := NewAnonymousFunctionInJSXRule(config)
	if rule.ID() != "anonymous-function-jsx" {
		t.Errorf("Expected ID 'anonymous-function-jsx', got '%s'", rule.ID())
	}
}

func TestConsoleLogRule_ID(t *testing.T) {
	config := getTestConfig()
	rule := NewConsoleLogRule(config)
	if rule.ID() != "console-log" {
		t.Errorf("Expected ID 'console-log', got '%s'", rule.ID())
	}
}

func TestDeprecatedLifecycleRule_ID(t *testing.T) {
	config := getTestConfig()
	rule := NewDeprecatedLifecycleRule(config)
	if rule.ID() != "deprecated-lifecycle" {
		t.Errorf("Expected ID 'deprecated-lifecycle', got '%s'", rule.ID())
	}
}

func TestHardcodedDimensionRule_ID(t *testing.T) {
	config := getTestConfig()
	rule := NewHardcodedDimensionRule(config)
	if rule.ID() != "hardcoded-dimension" {
		t.Errorf("Expected ID 'hardcoded-dimension', got '%s'", rule.ID())
	}
}

func TestDirectStateMutationRule_ID(t *testing.T) {
	config := getTestConfig()
	rule := NewDirectStateMutationRule(config)
	if rule.ID() != "direct-state-mutation" {
		t.Errorf("Expected ID 'direct-state-mutation', got '%s'", rule.ID())
	}
}
