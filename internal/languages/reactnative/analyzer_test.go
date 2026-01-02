package reactnative

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func getTestConfig() core.Config {
	return core.Config{
		Rules: core.RulesConfig{
			FunctionSize:   core.FunctionSizeConfig{MaxLines: 50, Enabled: true},
			FileSize:       core.FileSizeConfig{MaxLines: 500, Enabled: true},
			Overcommenting: core.OvercommentingConfig{MaxCommentRatio: 0.30, Enabled: true},
			OrphanedCode:   core.OrphanedCodeConfig{CheckUnusedFunctions: true},
		},
	}
}

func TestAnalyzer_Name(t *testing.T) {
	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	if analyzer.Name() != "reactnative" {
		t.Errorf("Expected name 'reactnative', got '%s'", analyzer.Name())
	}
}

func TestAnalyzer_SupportedExtensions(t *testing.T) {
	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	extensions := analyzer.SupportedExtensions()
	expected := []string{".js", ".jsx", ".ts", ".tsx"}
	if len(extensions) != len(expected) {
		t.Errorf("Expected %d extensions, got %d", len(expected), len(extensions))
	}
	for i, ext := range expected {
		if extensions[i] != ext {
			t.Errorf("Expected extension '%s', got '%s'", ext, extensions[i])
		}
	}
}

func TestAnalyzer_Analyze(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "test.js")
	content := `// Test file
import React from 'react';
import { useState } from 'react';

function SmallComponent() {
    return <div>Hello</div>;
}

export default SmallComponent;
`
	if err := os.WriteFile(jsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), jsFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if results == nil {
		t.Fatal("Expected results, got nil")
	}
}

func TestAnalyzer_LargeFunctionDetection(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "large.js")
	lines := []string{
		"function largeFunction() {",
	}
	for i := 0; i < 60; i++ {
		lines = append(lines, "    console.log('line');")
	}
	lines = append(lines, "}")
	content := strings.Join(lines, "\n")

	if err := os.WriteFile(jsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), jsFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	found := false
	for _, result := range results {
		if result.RuleID == "large-function" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find large-function violation")
	}
}

func TestAnalyzer_LargeFileDetection(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "huge.js")
	lines := []string{}
	for i := 0; i < 600; i++ {
		lines = append(lines, "console.log('line');")
	}
	content := strings.Join(lines, "\n")

	if err := os.WriteFile(jsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), jsFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	found := false
	for _, result := range results {
		if result.RuleID == "large-file" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find large-file violation")
	}
}

func TestAnalyzer_OvercommentingDetection(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "comments.js")
	content := `// Comment 1
// Comment 2
// Comment 3
// Comment 4
// Comment 5
console.log('code');
/* More comments */
/* Even more */
`
	if err := os.WriteFile(jsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), jsFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	found := false
	for _, result := range results {
		if result.RuleID == "overcommenting" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find overcommenting violation")
	}
}

func TestAnalyzer_TypeScriptFiles(t *testing.T) {
	tmpDir := t.TempDir()
	tsxFile := filepath.Join(tmpDir, "Component.tsx")
	content := `import React from 'react';

interface Props {
    name: string;
}

const MyComponent: React.FC<Props> = ({ name }) => {
    return <div>Hello, {name}</div>;
};

export default MyComponent;
`
	if err := os.WriteFile(tsxFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), tsxFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if results == nil {
		t.Fatal("Expected results, got nil")
	}
}

func TestAnalyzer_ArrowFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "arrows.js")
	lines := []string{
		"const largeArrow = () => {",
	}
	for i := 0; i < 60; i++ {
		lines = append(lines, "    console.log('line');")
	}
	lines = append(lines, "};")
	content := strings.Join(lines, "\n")

	if err := os.WriteFile(jsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := getTestConfig()
	analyzer := NewAnalyzer(config)
	results, err := analyzer.Analyze(context.Background(), jsFile, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	found := false
	for _, result := range results {
		if result.RuleID == "large-function" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find large-function violation for arrow function")
	}
}
