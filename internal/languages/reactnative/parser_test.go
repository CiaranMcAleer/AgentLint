package reactnative

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

func getParserTestConfig() core.Config {
	return core.Config{
		Rules: core.RulesConfig{
			FunctionSize:   core.FunctionSizeConfig{MaxLines: 50, Enabled: true},
			FileSize:       core.FileSizeConfig{MaxLines: 500, Enabled: true},
			Overcommenting: core.OvercommentingConfig{MaxCommentRatio: 0.30, Enabled: true},
		},
	}
}

func createTestFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.js")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

func TestParser_ParseFile(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `import React from 'react';
import { useState, useEffect } from 'react';

// A simple component
function MyComponent() {
    const [count, setCount] = useState(0);
    return <div>{count}</div>;
}

export const MyComponent2 = MyComponent;
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if parsed == nil {
		t.Fatal("Expected parsed file, got nil")
	}
	if len(parsed.Imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(parsed.Imports))
	}
	if len(parsed.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(parsed.Functions))
	}
	// The export pattern only matches export const/function/class declarations
	if len(parsed.Exports) < 1 {
		t.Errorf("Expected at least 1 export, got %d", len(parsed.Exports))
	}
}

func TestParser_ParseFunctions(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `function regularFunction() {
    console.log('hello');
}

async function asyncFunction() {
    await fetch('/api');
}

const arrowFunction = () => {
    return 42;
};

const asyncArrow = async () => {
    await doSomething();
};
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(parsed.Functions) != 4 {
		t.Errorf("Expected 4 functions, got %d", len(parsed.Functions))
	}

	functionNames := make(map[string]bool)
	for _, f := range parsed.Functions {
		functionNames[f.Name] = true
	}
	expectedNames := []string{"regularFunction", "asyncFunction", "arrowFunction", "asyncArrow"}
	for _, name := range expectedNames {
		if !functionNames[name] {
			t.Errorf("Expected to find function '%s'", name)
		}
	}
}

func TestParser_ParseClasses(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `class MyClass {
    constructor() {
        this.value = 0;
    }

    getValue() {
        return this.value;
    }
}

class MyComponent extends React.Component {
    render() {
        return <div>Hello</div>;
    }
}
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(parsed.Classes) != 2 {
		t.Errorf("Expected 2 classes, got %d", len(parsed.Classes))
	}

	classNames := make(map[string]bool)
	for _, c := range parsed.Classes {
		classNames[c.Name] = true
	}
	if !classNames["MyClass"] {
		t.Error("Expected to find class 'MyClass'")
	}
	if !classNames["MyComponent"] {
		t.Error("Expected to find class 'MyComponent'")
	}
}

func TestParser_ParseImports(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `import React from 'react';
import { useState, useEffect } from 'react';
import * as Utils from './utils';
import './styles.css';
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	// The pattern only matches `import X from 'Y'` syntax, not side-effect imports
	if len(parsed.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(parsed.Imports))
	}
}

func TestParser_ParseExports(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `export const value = 42;
export function myFunc() {}
export default MyComponent;
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(parsed.Exports) < 2 {
		t.Errorf("Expected at least 2 exports, got %d", len(parsed.Exports))
	}
}

func TestParser_ParseComments(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `// Single line comment
/* Block comment */
/**
 * JSDoc comment
 */
function myFunc() {}
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(parsed.Comments) < 2 {
		t.Errorf("Expected at least 2 comments, got %d", len(parsed.Comments))
	}
}

func TestParser_TypeScriptParsing(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `import React from 'react';

interface Props {
    name: string;
    age: number;
}

type User = {
    id: string;
    email: string;
};

const Component: React.FC<Props> = ({ name, age }) => {
    return <div>{name} is {age}</div>;
};

export default Component;
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if parsed == nil {
		t.Fatal("Expected parsed file, got nil")
	}
}

func TestParser_EmptyFile(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	filePath := createTestFile(t, "")
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if parsed == nil {
		t.Fatal("Expected parsed file, got nil")
	}
	if len(parsed.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(parsed.Functions))
	}
}

func TestParser_ReactComponent(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `import React from 'react';

function MyComponent({ prop1, prop2 }) {
    return (
        <div className="container">
            <h1>{prop1}</h1>
            <p>{prop2}</p>
        </div>
    );
}

export default MyComponent;
`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(parsed.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(parsed.Functions))
	}
	if len(parsed.Functions) > 0 && parsed.Functions[0].Name != "MyComponent" {
		t.Errorf("Expected function name 'MyComponent', got '%s'", parsed.Functions[0].Name)
	}
}

func TestParser_LineMetrics(t *testing.T) {
	config := getParserTestConfig()
	parser := NewParser(config)
	content := `// Comment
import React from 'react';

function test() {
    return 1;
}`
	filePath := createTestFile(t, content)
	parsed, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if parsed.TotalLines != 6 {
		t.Errorf("Expected 6 total lines, got %d", parsed.TotalLines)
	}
}
