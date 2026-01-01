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
	"time"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang/rules"
)

type cachedFile struct {
	file     *ast.File
	fset     *token.FileSet
	modTime  time.Time
	filePath string
}

type ASTCache struct {
	cache  map[string]*cachedFile
	mu     sync.RWMutex
	maxAge time.Duration
}

func NewASTCache(maxAge time.Duration) *ASTCache {
	if maxAge == 0 {
		maxAge = 5 * time.Minute
	}
	return &ASTCache{
		cache:  make(map[string]*cachedFile),
		maxAge: maxAge,
	}
}

func (c *ASTCache) Get(filePath string) (*ast.File, *token.FileSet, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.cache[filePath]
	if !exists {
		return nil, nil, false
	}

	if time.Since(cached.modTime) > c.maxAge {
		delete(c.cache, filePath)
		return nil, nil, false
	}

	return cached.file, cached.fset, true
}

func (c *ASTCache) Set(filePath string, file *ast.File, fset *token.FileSet) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stat, err := os.Stat(filePath)
	if err != nil {
		return
	}

	c.cache[filePath] = &cachedFile{
		file:     file,
		fset:     fset,
		modTime:  stat.ModTime(),
		filePath: filePath,
	}
}

func (c *ASTCache) Invalidate(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, filePath)
}

func (c *ASTCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*cachedFile)
}

func (c *ASTCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

func (c *ASTCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		Entries: len(c.cache),
	}

	for _, cached := range c.cache {
		age := time.Since(cached.modTime)
		if age > stats.MaxAge {
			stats.MaxAge = age
		}
		if age < stats.MinAge {
			stats.MinAge = age
		}
		stats.TotalAge += age
	}

	if stats.Entries > 0 {
		stats.AvgAge = stats.TotalAge / time.Duration(stats.Entries)
	}

	return stats
}

type CacheStats struct {
	Entries  int
	MaxAge   time.Duration
	MinAge   time.Duration
	AvgAge   time.Duration
	TotalAge time.Duration
}

type Parser struct {
	fset   *token.FileSet
	config core.Config
	cache  *ASTCache
}

func NewParser(config core.Config) *Parser {
	return &Parser{
		fset:   token.NewFileSet(),
		config: config,
		cache:  NewASTCache(0),
	}
}

func (p *Parser) SetCache(cache *ASTCache) {
	p.cache = cache
}

func (p *Parser) ParseFile(ctx context.Context, filePath string) (*ast.File, *token.FileSet, error) {
	if p.cache != nil {
		if file, fset, ok := p.cache.Get(filePath); ok {
			return file, fset, nil
		}
	}

	if p.shouldIgnoreFile(filePath) {
		return nil, nil, fmt.Errorf("file ignored: %s", filePath)
	}

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	file, err := parser.ParseFile(p.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	if p.cache != nil {
		p.cache.Set(filePath, file, p.fset)
	}

	return file, p.fset, nil
}

func (p *Parser) shouldIgnoreFile(filePath string) bool {
	if p.config.Language.Go.IgnoreTests {
		base := filepath.Base(filePath)
		if strings.HasSuffix(base, "_test.go") {
			return true
		}
	}

	base := filepath.Base(filePath)
	if strings.HasPrefix(base, "_") {
		return true
	}

	return false
}

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

func (p *Parser) CalculateFunctionMetrics(ctx context.Context, funcDecl *ast.FuncDecl, fset *token.FileSet, file *ast.File) (*rules.FunctionMetrics, error) {
	start := fset.Position(funcDecl.Pos())
	end := fset.Position(funcDecl.End())

	lineCount := end.Line - start.Line + 1

	isMainPackage := file.Name.Name == "main"

	return &rules.FunctionMetrics{
		Name:                 funcDecl.Name.Name,
		Receiver:             getReceiverName(funcDecl),
		Exported:             funcDecl.Name.IsExported(),
		IsMainPackage:        isMainPackage,
		LineCount:            lineCount,
		ParameterCount:       countParams(funcDecl),
		ReturnCount:          countReturns(funcDecl),
		CyclomaticComplexity: p.calculateCyclomaticComplexity(funcDecl),
		NestingDepth:         calculateNestingDepth(funcDecl),
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

func calculateNestingDepth(funcDecl *ast.FuncDecl) int {
	if funcDecl.Body == nil {
		return 0
	}

	maxDepth := 0
	currentDepth := 0

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt:
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		}
		return true
	})

	return maxDepth
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
