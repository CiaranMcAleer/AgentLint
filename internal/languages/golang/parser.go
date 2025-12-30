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

	"github.com/agentlint/agentlint/internal/core"
	"github.com/agentlint/agentlint/internal/languages/golang/rules"
)

// Parser handles parsing Go source code into ASTs
type Parser struct {
	fset   *token.FileSet
	config core.Config
}

// NewParser creates a new Go parser
func NewParser(config core.Config) *Parser {
	return &Parser{
		fset:   token.NewFileSet(),
		config: config,
	}
}

// ParseFile parses a Go source file into an AST
func (p *Parser) ParseFile(ctx context.Context, filePath string) (*ast.File, *token.FileSet, error) {
	// Check if file should be ignored
	if p.shouldIgnoreFile(filePath) {
		return nil, nil, fmt.Errorf("file ignored: %s", filePath)
	}

	// Read the file
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse the file
	file, err := parser.ParseFile(p.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	return file, p.fset, nil
}

// shouldIgnoreFile checks if a file should be ignored based on configuration
func (p *Parser) shouldIgnoreFile(filePath string) bool {
	// Ignore test files if configured
	if p.config.Language.Go.IgnoreTests {
		base := filepath.Base(filePath)
		if strings.HasSuffix(base, "_test.go") {
			return true
		}
	}

	// Ignore files with _ in front (like _generated.go)
	base := filepath.Base(filePath)
	if strings.HasPrefix(base, "_") {
		return true
	}

	return false
}

// CalculateMetrics calculates various metrics for a Go file
func (p *Parser) CalculateMetrics(ctx context.Context, filePath string, file *ast.File) (*rules.FileMetrics, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	lines := strings.Split(string(src), "\n")

	lineCounts := countLineTypes(lines)
	astCounts := countASTElements(file)

	commentRatio := 0.0
	if lineCounts.code > 0 {
		commentRatio = float64(lineCounts.comment) / float64(lineCounts.code)
	}

	return &rules.FileMetrics{
		Path:          filePath,
		TotalLines:    lineCounts.total,
		CodeLines:     lineCounts.code,
		CommentLines:  lineCounts.comment,
		BlankLines:    lineCounts.blank,
		CommentRatio:  commentRatio,
		FunctionCount: astCounts.functions,
		ImportCount:   astCounts.imports,
		ExportedCount: astCounts.exported,
	}, nil
}

type lineCounts struct {
	total   int
	code    int
	comment int
	blank   int
}

func countLineTypes(lines []string) lineCounts {
	var counts lineCounts
	counts.total = len(lines)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			counts.blank++
		} else if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			counts.comment++
		} else {
			counts.code++
		}
	}
	return counts
}

type astCounts struct {
	functions int
	imports   int
	exported  int
}

func countASTElements(file *ast.File) astCounts {
	var counts astCounts

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			counts.functions++
			if node.Name.IsExported() {
				counts.exported++
			}
		case *ast.ImportSpec:
			counts.imports++
		}
		return true
	})

	return counts
}

// CalculateFunctionMetrics calculates metrics for a function declaration
func (p *Parser) CalculateFunctionMetrics(ctx context.Context, funcDecl *ast.FuncDecl, fset *token.FileSet) (*rules.FunctionMetrics, error) {
	start := fset.Position(funcDecl.Pos())
	end := fset.Position(funcDecl.End())

	lineCount := end.Line - start.Line + 1

	return &rules.FunctionMetrics{
		Name:                 funcDecl.Name.Name,
		Receiver:             getReceiverName(funcDecl),
		Exported:             funcDecl.Name.IsExported(),
		LineCount:            lineCount,
		ParameterCount:       countParams(funcDecl),
		ReturnCount:          countReturns(funcDecl),
		CyclomaticComplexity: p.calculateCyclomaticComplexity(funcDecl),
		Position:             start,
	}, nil
}

func getReceiverName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func countParams(funcDecl *ast.FuncDecl) int {
	if funcDecl.Type.Params != nil {
		return len(funcDecl.Type.Params.List)
	}
	return 0
}

func countReturns(funcDecl *ast.FuncDecl) int {
	if funcDecl.Type.Results != nil {
		return len(funcDecl.Type.Results.List)
	}
	return 0
}

// calculateCyclomaticComplexity calculates the cyclomatic complexity of a function
func (p *Parser) calculateCyclomaticComplexity(funcDecl *ast.FuncDecl) int {
	complexity := 1

	ast.Inspect(funcDecl, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause:
			complexity++
		case *ast.SwitchStmt:
			complexity += 1 + countSwitchCases(n.(*ast.SwitchStmt))
		case *ast.SelectStmt:
			complexity += 1 + countSelectCases(n.(*ast.SelectStmt))
		case *ast.BinaryExpr:
			if n.(*ast.BinaryExpr).Op == token.LAND || n.(*ast.BinaryExpr).Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}

func countSwitchCases(switchStmt *ast.SwitchStmt) int {
	if switchStmt.Body == nil {
		return 0
	}
	count := 0
	for _, stmt := range switchStmt.Body.List {
		if _, ok := stmt.(*ast.CaseClause); ok {
			count++
		}
	}
	return count
}

func countSelectCases(selectStmt *ast.SelectStmt) int {
	if selectStmt.Body == nil {
		return 0
	}
	count := 0
	for _, stmt := range selectStmt.Body.List {
		if commClause, ok := stmt.(*ast.CommClause); ok && commClause.Comm != nil {
			count++
		}
	}
	return count
}
