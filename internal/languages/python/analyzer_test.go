package python

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func TestAnalyzer_SupportedExtensions(t *testing.T) {
	config := core.Config{}
	analyzer := NewAnalyzer(config)

	extensions := analyzer.SupportedExtensions()
	if len(extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(extensions))
	}

	expectedExtensions := map[string]bool{".py": true, ".pyw": true}
	for _, ext := range extensions {
		if !expectedExtensions[ext] {
			t.Errorf("Unexpected extension: %s", ext)
		}
	}
}

func TestAnalyzer_Name(t *testing.T) {
	config := core.Config{}
	analyzer := NewAnalyzer(config)

	if analyzer.Name() != "python" {
		t.Errorf("Expected name 'python', got '%s'", analyzer.Name())
	}
}

func TestAnalyzer_LargeFunctionRule(t *testing.T) {
	// Create a temporary Python file with a large function
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large_func.py")

	// Create a function with 60 lines (exceeds default 50)
	content := "def large_function():\n"
	for i := 0; i < 60; i++ {
		content += "    x = " + string(rune('a'+i%26)) + "\n"
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
		},
	}

	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), filePath, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should find the large function
	found := false
	for _, result := range results {
		if result.RuleID == "large-function" {
			found = true
			if result.Line != 1 {
				t.Errorf("Expected line 1, got %d", result.Line)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find large-function rule violation")
	}
}

func TestAnalyzer_LargeFileRule(t *testing.T) {
	// Create a temporary Python file with many lines
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large_file.py")

	// Create a file with 600 lines (exceeds default 500)
	content := ""
	for i := 0; i < 600; i++ {
		content += "x = " + string(rune('a'+i%26)) + "\n"
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config := core.Config{
		Rules: core.RulesConfig{
			FileSize: core.FileSizeConfig{
				Enabled:  true,
				MaxLines: 500,
			},
		},
	}

	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), filePath, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should find the large file
	found := false
	for _, result := range results {
		if result.RuleID == "large-file" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find large-file rule violation")
	}
}

func TestAnalyzer_OvercommentingRule(t *testing.T) {
	// Create a temporary Python file with excessive comments
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "overcommented.py")

	content := `# Comment 1
# Comment 2
# Comment 3
# Comment 4
# Comment 5
x = 1
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config := core.Config{
		Rules: core.RulesConfig{
			Overcommenting: core.OvercommentingConfig{
				Enabled:         true,
				MaxCommentRatio: 0.3,
			},
		},
	}

	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), filePath, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should find overcommenting
	found := false
	for _, result := range results {
		if result.RuleID == "overcommenting" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find overcommenting rule violation")
	}
}

func TestAnalyzer_NoFalsePositivesForSmallFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "small.py")

	content := `def hello():
    print("Hello, World!")

def main():
    hello()

if __name__ == "__main__":
    main()
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
			FileSize: core.FileSizeConfig{
				Enabled:  true,
				MaxLines: 500,
			},
			Overcommenting: core.OvercommentingConfig{
				Enabled:         true,
				MaxCommentRatio: 0.3,
			},
		},
	}

	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), filePath, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(results) > 0 {
		t.Errorf("Expected no issues for small clean file, got %d", len(results))
		for _, r := range results {
			t.Logf("  - %s: %s", r.RuleID, r.Message)
		}
	}
}

func TestAnalyzer_DisabledRules(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large_func.py")

	// Create a large function
	content := "def large_function():\n"
	for i := 0; i < 60; i++ {
		content += "    x = " + string(rune('a'+i%26)) + "\n"
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Disable the function size rule
	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  false,
				MaxLines: 50,
			},
		},
	}

	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), filePath, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should NOT find the large function since rule is disabled
	for _, result := range results {
		if result.RuleID == "large-function" {
			t.Error("Should not find large-function violation when rule is disabled")
		}
	}
}

func TestAnalyzer_MethodDetection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "class_methods.py")

	// Create a class with a large method
	content := `class MyClass:
    def large_method(self):
`
	for i := 0; i < 60; i++ {
		content += "        x = " + string(rune('a'+i%26)) + "\n"
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  true,
				MaxLines: 50,
			},
		},
	}

	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), filePath, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should find the large method
	found := false
	for _, result := range results {
		if result.RuleID == "large-function" {
			found = true
			if result.Message == "" {
				t.Error("Expected non-empty message")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find large-function rule violation for method")
	}
}

func TestAllRulesHaveRequiredMethods(t *testing.T) {
	config := core.Config{}
	analyzer := NewAnalyzer(config)

	if len(analyzer.rules) == 0 {
		t.Error("Analyzer should have rules")
	}

	for _, rule := range analyzer.rules {
		if rule.ID() == "" {
			t.Errorf("Rule %s has empty ID", rule.Name())
		}
		if rule.Name() == "" {
			t.Error("Rule has empty Name")
		}
		if rule.Description() == "" {
			t.Errorf("Rule %s has empty Description", rule.ID())
		}
		if rule.Category() == "" {
			t.Errorf("Rule %s has empty Category", rule.ID())
		}
		if rule.Severity() == "" {
			t.Errorf("Rule %s has empty Severity", rule.ID())
		}
	}
}

func TestAnalyzer_HasAllExpectedRules(t *testing.T) {
	config := core.Config{}
	analyzer := NewAnalyzer(config)

	expectedRules := map[string]bool{
		"large-function":   false,
		"large-file":       false,
		"overcommenting":   false,
		"unused-function":  false,
		"unused-variable":  false,
		"unreachable-code": false,
		"dead-import":      false,
	}

	for _, rule := range analyzer.rules {
		if _, exists := expectedRules[rule.ID()]; exists {
			expectedRules[rule.ID()] = true
		}
	}

	for ruleID, found := range expectedRules {
		if !found {
			t.Errorf("Expected rule %s not found in analyzer", ruleID)
		}
	}
}
