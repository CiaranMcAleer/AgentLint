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
	methods         map[string]map[string]*FunctionInfo // receiver type -> method name -> info
	calls           map[string][]string
	methodCalls     map[string][]string // tracks method calls separately
	funcReferences  map[string]bool     // tracks functions used as references (callbacks, etc.)
	mu              sync.RWMutex
	ignoredPrefixes []string
}

type FunctionInfo struct {
	Name       string
	File       string
	Exported   bool
	IsMain     bool
	IsTest     bool
	IsInit     bool
	IsMethod   bool
	Receiver   string // receiver type name for methods
	Line       int
	Package    string
}

func NewCrossFileAnalyzer() *CrossFileAnalyzer {
	return &CrossFileAnalyzer{
		fset:            token.NewFileSet(),
		functions:       make(map[string]map[string]*FunctionInfo),
		methods:         make(map[string]map[string]*FunctionInfo),
		calls:           make(map[string][]string),
		methodCalls:     make(map[string][]string),
		funcReferences:  make(map[string]bool),
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
	pkgName := a.getPackageName(f)

	a.collectDeclarations(f, filePath, pkgName)
	a.collectCalls(f, filePath)

	return nil
}

// getPackageName extracts the package name from a parsed file
func (a *CrossFileAnalyzer) getPackageName(f *ast.File) string {
	if f.Name != nil {
		return f.Name.Name
	}
	return ""
}

// collectDeclarations collects all function and method declarations from a file
func (a *CrossFileAnalyzer) collectDeclarations(f *ast.File, filePath, pkgName string) {
	ast.Inspect(f, func(n ast.Node) bool {
		if node, ok := n.(*ast.FuncDecl); ok {
			a.registerFunction(node, filePath, pkgName)
		}
		return true
	})
}

// registerFunction registers a function or method declaration
func (a *CrossFileAnalyzer) registerFunction(node *ast.FuncDecl, filePath, pkgName string) {
	receiverType := getReceiverTypeName(node)
	isMethod := receiverType != ""

	funcInfo := &FunctionInfo{
		Name:     node.Name.Name,
		File:     filePath,
		Exported: node.Name.IsExported(),
		IsMain:   node.Name.Name == "main",
		IsTest:   strings.HasPrefix(node.Name.Name, "Test") || strings.HasSuffix(node.Name.Name, "Test"),
		IsInit:   node.Name.Name == "init",
		IsMethod: isMethod,
		Receiver: receiverType,
		Line:     a.fset.Position(node.Pos()).Line,
		Package:  pkgName,
	}

	if isMethod {
		if a.methods[receiverType] == nil {
			a.methods[receiverType] = make(map[string]*FunctionInfo)
		}
		a.methods[receiverType][node.Name.Name] = funcInfo
	} else {
		a.functions[filePath][node.Name.Name] = funcInfo
	}
}

// collectCalls collects all function calls from a file
func (a *CrossFileAnalyzer) collectCalls(f *ast.File, filePath string) {
	ast.Inspect(f, func(n ast.Node) bool {
		if node, ok := n.(*ast.FuncDecl); ok {
			a.collectCallsFromNode(filePath, node.Name.Name, node.Body)
		}
		return true
	})
}

// getReceiverTypeName extracts the receiver type name from a function declaration
func getReceiverTypeName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return ""
	}

	recv := funcDecl.Recv.List[0]
	switch t := recv.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

// collectCallsFromNode traverses a node and records all function/method calls and references
func (a *CrossFileAnalyzer) collectCallsFromNode(filePath, callerName string, node ast.Node) {
	if node == nil {
		return
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.CallExpr:
			a.recordCallExpr(filePath, callerName, expr)

		case *ast.Ident:
			// Check if this identifier is a function reference (not a call)
			// This catches cases like: handler := myFunction
			if expr.Obj != nil && expr.Obj.Kind == ast.Fun {
				a.funcReferences[expr.Name] = true
			}

		case *ast.SelectorExpr:
			// Check for function references via selector (e.g., pkg.Function used as value)
			// We'll be conservative and just record the method name
			// expr.Sel is already *ast.Ident
			a.funcReferences[expr.Sel.Name] = true
		}
		return true
	})
}

// recordCallExpr handles recording of a call expression
func (a *CrossFileAnalyzer) recordCallExpr(filePath, callerName string, call *ast.CallExpr) {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		// Direct function call: functionName()
		a.recordCall(filePath, callerName, fun.Name)

	case *ast.SelectorExpr:
		// Method call: obj.Method() or pkg.Function()
		methodName := fun.Sel.Name
		a.recordMethodCall(filePath, callerName, methodName)

		// Also record as a regular call in case it's a package-level function
		a.recordCall(filePath, callerName, methodName)

	case *ast.FuncLit:
		// Anonymous function - traverse its body too
		a.collectCallsFromNode(filePath, callerName, fun.Body)
	}

	// Also check arguments for function references
	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			// Function passed as argument
			a.funcReferences[ident.Name] = true
		}
	}
}

func (a *CrossFileAnalyzer) recordCall(filePath, caller, callee string) {
	key := filePath + ":" + caller
	a.calls[key] = append(a.calls[key], callee)
}

func (a *CrossFileAnalyzer) recordMethodCall(filePath, caller, methodName string) {
	key := filePath + ":" + caller
	a.methodCalls[key] = append(a.methodCalls[key], methodName)
}

func (a *CrossFileAnalyzer) FindUnusedFunctions() []core.Result {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := a.findUnusedRegularFunctions()
	results = append(results, a.findUnusedMethods()...)
	return results
}

// findUnusedRegularFunctions finds unused regular (non-method) functions
func (a *CrossFileAnalyzer) findUnusedRegularFunctions() []core.Result {
	var results []core.Result
	for filePath, funcs := range a.functions {
		for name, funcInfo := range funcs {
			if a.isIgnoredFunction(funcInfo) || a.isCalled(funcInfo) {
				continue
			}
			results = append(results, a.buildUnusedFunctionResult(filePath, name, funcInfo))
		}
	}
	return results
}

// findUnusedMethods finds unused methods
func (a *CrossFileAnalyzer) findUnusedMethods() []core.Result {
	var results []core.Result
	for _, methods := range a.methods {
		for name, funcInfo := range methods {
			if a.isIgnoredFunction(funcInfo) || a.isMethodCalled(funcInfo) {
				continue
			}
			results = append(results, a.buildUnusedMethodResult(name, funcInfo))
		}
	}
	return results
}

// buildUnusedFunctionResult creates a result for an unused function
func (a *CrossFileAnalyzer) buildUnusedFunctionResult(filePath, name string, funcInfo *FunctionInfo) core.Result {
	return core.Result{
		RuleID:     "cross-file-unused-function",
		RuleName:   "Cross-File Unused Function",
		Category:   "orphaned",
		Severity:   "warning",
		FilePath:   filePath,
		Line:       funcInfo.Line,
		Message:    fmt.Sprintf("Function '%s' is not called anywhere in the project", name),
		Suggestion: "Review if this function is needed or if it should be exported/called",
	}
}

// buildUnusedMethodResult creates a result for an unused method
func (a *CrossFileAnalyzer) buildUnusedMethodResult(name string, funcInfo *FunctionInfo) core.Result {
	return core.Result{
		RuleID:     "cross-file-unused-method",
		RuleName:   "Cross-File Unused Method",
		Category:   "orphaned",
		Severity:   "warning",
		FilePath:   funcInfo.File,
		Line:       funcInfo.Line,
		Message:    fmt.Sprintf("Method '%s' on receiver '%s' is not called anywhere in the project", name, funcInfo.Receiver),
		Suggestion: "Review if this method is needed or if it implements an interface",
	}
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

	// Exported functions may be called from external packages,
	// so we can't determine if they're unused from internal analysis alone
	if funcInfo.Exported {
		return true
	}

	return false
}

func (a *CrossFileAnalyzer) isCalled(funcInfo *FunctionInfo) bool {
	if funcInfo.IsMain || funcInfo.IsInit || funcInfo.IsTest {
		return true
	}

	// Check if function is used as a reference (callback, assigned to variable, etc.)
	if a.funcReferences[funcInfo.Name] {
		return true
	}

	// Check direct function calls
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

	// Also check calls from methods
	for _, methods := range a.methods {
		for _, caller := range methods {
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

// isMethodCalled checks if a method is called anywhere in the project
func (a *CrossFileAnalyzer) isMethodCalled(funcInfo *FunctionInfo) bool {
	if funcInfo.IsMain || funcInfo.IsInit || funcInfo.IsTest {
		return true
	}

	// Check if method is used as a reference
	if a.funcReferences[funcInfo.Name] {
		return true
	}

	// Check method calls from functions
	for _, funcs := range a.functions {
		for _, caller := range funcs {
			callerKey := caller.File + ":" + caller.Name
			for _, callee := range a.methodCalls[callerKey] {
				if callee == funcInfo.Name {
					return true
				}
			}
			// Also check regular calls (methods can be called directly in some contexts)
			for _, callee := range a.calls[callerKey] {
				if callee == funcInfo.Name {
					return true
				}
			}
		}
	}

	// Check method calls from other methods
	for _, methods := range a.methods {
		for _, caller := range methods {
			callerKey := caller.File + ":" + caller.Name
			for _, callee := range a.methodCalls[callerKey] {
				if callee == funcInfo.Name {
					return true
				}
			}
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
