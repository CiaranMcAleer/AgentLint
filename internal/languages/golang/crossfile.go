package golang

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

type CrossFileAnalyzer struct {
	fset            *token.FileSet
	functions       map[string]map[string]*FunctionInfo
	calls           map[string][]string
	mu              sync.RWMutex
	ignoredPrefixes []string
}

type FunctionInfo struct {
	Name     string
	File     string
	Exported bool
	IsMain   bool
	IsTest   bool
	IsInit   bool
	Line     int
}

func NewCrossFileAnalyzer() *CrossFileAnalyzer {
	return &CrossFileAnalyzer{
		fset:            token.NewFileSet(),
		functions:       make(map[string]map[string]*FunctionInfo),
		calls:           make(map[string][]string),
		ignoredPrefixes: []string{"Benchmark", "Example", "Test"},
	}
}

func (a *CrossFileAnalyzer) AnalyzeDirectory(ctx context.Context, dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		if err := a.analyzeFile(path); err != nil {
			return err
		}

		return nil
	})
}

func shouldSkipDir(name string) bool {
	skipDirs := []string{".git", "node_modules", "vendor", ".vscode", ".idea"}
	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return false
}

func (a *CrossFileAnalyzer) analyzeFile(filePath string) error {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	f, err := parser.ParseFile(a.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.functions[filePath] = make(map[string]*FunctionInfo)

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			funcInfo := &FunctionInfo{
				Name:     node.Name.Name,
				File:     filePath,
				Exported: node.Name.IsExported(),
				IsMain:   node.Name.Name == "main",
				IsTest:   strings.HasSuffix(node.Name.Name, "Test"),
				IsInit:   node.Name.Name == "init",
				Line:     a.fset.Position(node.Pos()).Line,
			}
			a.functions[filePath][node.Name.Name] = funcInfo

			ast.Inspect(node.Body, func(n ast.Node) bool {
				switch call := n.(type) {
				case *ast.CallExpr:
					if ident, ok := call.Fun.(*ast.Ident); ok {
						a.recordCall(filePath, node.Name.Name, ident.Name)
					}
				}
				return true
			})
		}
		return true
	})

	return nil
}

func (a *CrossFileAnalyzer) recordCall(filePath, caller, callee string) {
	key := filePath + ":" + caller
	a.calls[key] = append(a.calls[key], callee)
}

func (a *CrossFileAnalyzer) FindUnusedFunctions() []core.Result {
	var results []core.Result

	a.mu.RLock()
	defer a.mu.RUnlock()

	for filePath, funcs := range a.functions {
		for name, funcInfo := range funcs {
			if a.isIgnoredFunction(funcInfo) {
				continue
			}

			if !a.isCalled(funcInfo) {
				results = append(results, core.Result{
					RuleID:     "cross-file-unused-function",
					RuleName:   "Cross-File Unused Function",
					Category:   "orphaned",
					Severity:   "warning",
					FilePath:   filePath,
					Line:       funcInfo.Line,
					Message:    fmt.Sprintf("Function '%s' is not called anywhere in the project", name),
					Suggestion: "Review if this function is needed or if it should be exported/called",
				})
			}
		}
	}

	return results
}

func (a *CrossFileAnalyzer) isIgnoredFunction(funcInfo *FunctionInfo) bool {
	if funcInfo.IsMain || funcInfo.IsInit {
		return true
	}

	if funcInfo.IsTest {
		return true
	}

	for _, prefix := range a.ignoredPrefixes {
		if strings.HasPrefix(funcInfo.Name, prefix) {
			return true
		}
	}

	if funcInfo.Exported {
		return false
	}

	return false
}

func (a *CrossFileAnalyzer) isCalled(funcInfo *FunctionInfo) bool {
	if funcInfo.IsMain || funcInfo.IsInit || funcInfo.IsTest {
		return true
	}

	for _, funcs := range a.functions {
		for _, caller := range funcs {
			callerKey := caller.File + ":" + caller.Name
			for _, callee := range a.calls[callerKey] {
				if callee == funcInfo.Name {
					return true
				}
			}
		}
	}

	return false
}

func (a *CrossFileAnalyzer) GetCallGraph() map[string][]string {
	callGraph := make(map[string][]string)

	a.mu.RLock()
	defer a.mu.RUnlock()

	for key, calls := range a.calls {
		callGraph[key] = calls
	}

	return callGraph
}
