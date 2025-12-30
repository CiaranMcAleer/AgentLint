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
	totalLines := len(lines)

	var codeLines, commentLines, blankLines int
	var functionCount, importCount, exportedCount int

	// Count line types
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			blankLines++
		} else if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			commentLines++
		} else {
			codeLines++
		}
	}

	// Count AST elements
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			functionCount++
			if node.Name.IsExported() {
				exportedCount++
			}
		case *ast.ImportSpec:
			importCount++
		}
		return true
	})

	commentRatio := 0.0
	if codeLines > 0 {
		commentRatio = float64(commentLines) / float64(codeLines)
	}

	return &rules.FileMetrics{
		Path:          filePath,
		TotalLines:    totalLines,
		CodeLines:     codeLines,
		CommentLines:  commentLines,
		BlankLines:    blankLines,
		CommentRatio:  commentRatio,
		FunctionCount: functionCount,
		ImportCount:   importCount,
		ExportedCount: exportedCount,
	}, nil
}

// CalculateFunctionMetrics calculates metrics for a function declaration
func (p *Parser) CalculateFunctionMetrics(ctx context.Context, funcDecl *ast.FuncDecl, fset *token.FileSet) (*rules.FunctionMetrics, error) {
	// Get function position
	start := fset.Position(funcDecl.Pos())
	end := fset.Position(funcDecl.End())

	lineCount := end.Line - start.Line + 1

	// Count parameters
	paramCount := 0
	if funcDecl.Type.Params != nil {
		paramCount = len(funcDecl.Type.Params.List)
	}

	// Count return values
	returnCount := 0
	if funcDecl.Type.Results != nil {
		returnCount = len(funcDecl.Type.Results.List)
	}

	// Calculate cyclomatic complexity
	complexity := p.calculateCyclomaticComplexity(funcDecl)

	// Check if exported
	exported := funcDecl.Name.IsExported()

	// Get receiver name if method
	receiver := ""
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
			receiver = ident.Name
		}
	}

	return &rules.FunctionMetrics{
		Name:                 funcDecl.Name.Name,
		Receiver:             receiver,
		Exported:             exported,
		LineCount:            lineCount,
		ParameterCount:       paramCount,
		ReturnCount:          returnCount,
		CyclomaticComplexity: complexity,
		Position:             start,
	}, nil
}

// calculateCyclomaticComplexity calculates the cyclomatic complexity of a function
func (p *Parser) calculateCyclomaticComplexity(funcDecl *ast.FuncDecl) int {
	complexity := 1 // Base complexity

	ast.Inspect(funcDecl, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause:
			complexity++
		case *ast.SwitchStmt:
			complexity++
			// Count case clauses
			if node.Body != nil {
				for _, stmt := range node.Body.List {
					if _, ok := stmt.(*ast.CaseClause); ok {
						complexity++
					}
				}
			}
		case *ast.SelectStmt:
			complexity++
			// Count case clauses
			if node.Body != nil {
				for _, stmt := range node.Body.List {
					if commClause, ok := stmt.(*ast.CommClause); ok {
						if commClause.Comm != nil {
							complexity++
						}
					}
				}
			}
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}
