package golang

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

type SimilarityAnalyzer struct {
	fset       *token.FileSet
	funcSigs   map[string][]string
	funcBodies map[string]string
	mu         sync.RWMutex
}

func NewSimilarityAnalyzer() *SimilarityAnalyzer {
	return &SimilarityAnalyzer{
		fset:       token.NewFileSet(),
		funcSigs:   make(map[string][]string),
		funcBodies: make(map[string]string),
	}
}

func (a *SimilarityAnalyzer) AnalyzeDirectory(ctx context.Context, dirPath string, threshold float64) ([]core.Result, error) {
	var results []core.Result

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDirForSimilarity(info.Name()) {
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

	if err != nil {
		return nil, err
	}

	similarities := a.findSimilarFunctions(threshold)
	for _, sim := range similarities {
		results = append(results, core.Result{
			RuleID:     "code-similarity",
			RuleName:   "Code Similarity",
			Category:   "complexity",
			Severity:   "info",
			FilePath:   sim.File1,
			Line:       sim.Line1,
			Message:    sim.Message,
			Suggestion: sim.Suggestion,
		})
	}

	return results, nil
}

func shouldSkipDirForSimilarity(name string) bool {
	skipDirs := []string{".git", "node_modules", "vendor", ".vscode", ".idea"}
	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}
	return false
}

func (a *SimilarityAnalyzer) analyzeFile(filePath string) error {
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

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body == nil {
				return true
			}

			funcName := node.Name.Name
			if isIgnoredFunctionName(funcName) {
				return true
			}

			signature := a.getFunctionSignature(node)
			body := a.getNormalizedBody(node.Body)

			key := filePath + ":" + funcName
			a.funcSigs[key] = signature
			a.funcBodies[key] = body
		}
		return true
	})

	return nil
}

func isIgnoredFunctionName(name string) bool {
	ignored := []string{"init", "main", "Test", "Benchmark", "Example"}
	for _, prefix := range ignored {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (a *SimilarityAnalyzer) getFunctionSignature(node *ast.FuncDecl) []string {
	var sig []string

	if node.Recv != nil && len(node.Recv.List) > 0 {
		sig = append(sig, "receiver")
	}

	if node.Type.Params != nil {
		for range node.Type.Params.List {
			sig = append(sig, "param")
		}
	}

	if node.Type.Results != nil {
		for range node.Type.Results.List {
			sig = append(sig, "return")
		}
	}

	return sig
}

func (a *SimilarityAnalyzer) getNormalizedBody(body *ast.BlockStmt) string {
	var tokens []string

	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl:
			return false
		case *ast.IfStmt:
			tokens = append(tokens, "IF")
		case *ast.ForStmt:
			tokens = append(tokens, "FOR")
		case *ast.RangeStmt:
			tokens = append(tokens, "RANGE")
		case *ast.SwitchStmt:
			tokens = append(tokens, "SWITCH")
		case *ast.TypeSwitchStmt:
			tokens = append(tokens, "TYPE_SWITCH")
		case *ast.SelectStmt:
			tokens = append(tokens, "SELECT")
		case *ast.ReturnStmt:
			tokens = append(tokens, "RETURN")
		case *ast.BranchStmt:
			tokens = append(tokens, "BREAK")
		case *ast.AssignStmt:
			tokens = append(tokens, "ASSIGN")
		case *ast.CallExpr:
			tokens = append(tokens, "CALL")
		case *ast.BinaryExpr:
			tokens = append(tokens, "EXPR")
		case *ast.UnaryExpr:
			tokens = append(tokens, "UNARY")
		}
		return true
	})

	return strings.Join(tokens, " ")
}

type Similarity struct {
	File1      string
	Line1      int
	File2      string
	Line2      int
	Similarity float64
	Message    string
	Suggestion string
}

func (a *SimilarityAnalyzer) findSimilarFunctions(threshold float64) []Similarity {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var similarities []Similarity

	keys := make([]string, 0, len(a.funcBodies))
	for k := range a.funcBodies {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			key1, key2 := keys[i], keys[j]

			sim := a.calculateSimilarity(key1, key2)
			if sim >= threshold {
				similarities = append(similarities, Similarity{
					File1:      key1,
					File2:      key2,
					Similarity: sim,
					Message:    "Similar code patterns detected",
					Suggestion: "Consider extracting common logic into a shared function",
				})
			}
		}
	}

	return similarities
}

func (a *SimilarityAnalyzer) calculateSimilarity(key1, key2 string) float64 {
	body1 := a.funcBodies[key1]
	body2 := a.funcBodies[key2]

	if len(body1) == 0 || len(body2) == 0 {
		return 0
	}

	tokens1 := strings.Fields(body1)
	tokens2 := strings.Fields(body2)

	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0
	}

	matchCount := 0
	for _, t1 := range tokens1 {
		for _, t2 := range tokens2 {
			if t1 == t2 {
				matchCount++
				break
			}
		}
	}

	smaller := len(tokens1)
	if len(tokens2) < smaller {
		smaller = len(tokens2)
	}

	if smaller == 0 {
		return 0
	}

	return float64(matchCount) / float64(smaller)
}
