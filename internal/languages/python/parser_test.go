package python

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func TestParser_ParsesFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "functions.py")

	content := `def function_one():
    pass

def function_two(a, b):
    return a + b

def _private_function():
    pass
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewParser(core.Config{})
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(parsed.Functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(parsed.Functions))
	}

	foundPrivate := false
	for _, fn := range parsed.Functions {
		if fn.Name == "_private_function" && fn.IsPrivate {
			foundPrivate = true
		}
	}
	if !foundPrivate {
		t.Error("Expected to find private function")
	}
}

func TestParser_ParsesClasses(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "classes.py")

	content := `class MyClass:
    def __init__(self):
        pass

class ChildClass(MyClass):
    pass
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewParser(core.Config{})
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(parsed.Classes) != 2 {
		t.Errorf("Expected 2 classes, got %d", len(parsed.Classes))
	}
}

func TestParser_ParsesImports(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "imports.py")

	content := `import os
import sys
from typing import List, Optional
from collections import defaultdict
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewParser(core.Config{})
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(parsed.Imports) != 4 {
		t.Errorf("Expected 4 imports, got %d", len(parsed.Imports))
	}

	fromImportCount := 0
	for _, imp := range parsed.Imports {
		if imp.IsFrom {
			fromImportCount++
		}
	}
	if fromImportCount != 2 {
		t.Errorf("Expected 2 from imports, got %d", fromImportCount)
	}
}

func TestParser_CountsLines(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "lines.py")

	content := `# This is a comment
def hello():
    # Another comment
    print("Hello")

# Blank line above
x = 1
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewParser(core.Config{})
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if parsed.TotalLines != 7 {
		t.Errorf("Expected 7 total lines, got %d", parsed.TotalLines)
	}

	if parsed.CommentLines < 2 {
		t.Errorf("Expected at least 2 comment lines, got %d", parsed.CommentLines)
	}

	if parsed.BlankLines < 1 {
		t.Errorf("Expected at least 1 blank line, got %d", parsed.BlankLines)
	}
}

func TestParser_CacheWorks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "cached.py")

	content := `def hello():
    pass
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewParser(core.Config{})

	// First parse
	parsed1, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("First ParseFile failed: %v", err)
	}

	// Second parse should return cached result
	parsed2, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("Second ParseFile failed: %v", err)
	}

	// Should be the same pointer (cached)
	if parsed1 != parsed2 {
		t.Error("Expected cached result to be returned")
	}
}

func TestParser_ParsesDecorators(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "decorators.py")

	content := `@decorator
def decorated_function():
    pass

@classmethod
def class_method(cls):
    pass
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewParser(core.Config{})
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(parsed.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(parsed.Functions))
	}

	// Check decorators are captured
	for _, fn := range parsed.Functions {
		if len(fn.Decorators) == 0 {
			t.Errorf("Expected function %s to have decorators", fn.Name)
		}
	}
}

func TestParser_ParsesMethodsInClass(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "class_methods.py")

	content := `class MyClass:
    def __init__(self):
        pass

    def method_one(self):
        pass

    def _private_method(self):
        pass
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewParser(core.Config{})
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	methodCount := 0
	for _, fn := range parsed.Functions {
		if fn.IsMethod && fn.ClassName == "MyClass" {
			methodCount++
		}
	}

	if methodCount != 3 {
		t.Errorf("Expected 3 methods in MyClass, got %d", methodCount)
	}
}
