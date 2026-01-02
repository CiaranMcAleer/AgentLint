package reactnative

import (
	"os"
	"sync"
	"time"
)

// ParsedFile represents a parsed JavaScript/TypeScript file
type ParsedFile struct {
	Lines        []string
	Functions    []FunctionDef
	Classes      []ClassDef
	Components   []ComponentDef
	Imports      []ImportStmt
	Exports      []ExportStmt
	Comments     []Comment
	Variables    []VariableDef
	TotalLines   int
	CodeLines    int
	CommentLines int
	BlankLines   int
}

// FunctionDef represents a function definition
type FunctionDef struct {
	Name       string
	StartLine  int
	EndLine    int
	Parameters []string
	IsAsync    bool
	IsArrow    bool
	IsExported bool
	IsMethod   bool
	ClassName  string
	Indent     int
}

// ClassDef represents a class definition
type ClassDef struct {
	Name       string
	StartLine  int
	EndLine    int
	Extends    string
	IsExported bool
	Methods    []FunctionDef
}

// ComponentDef represents a React component
type ComponentDef struct {
	Name        string
	StartLine   int
	EndLine     int
	IsClass     bool
	IsFunctional bool
	IsExported  bool
	HasHooks    bool
}

// ImportStmt represents an import statement
type ImportStmt struct {
	Module     string
	Names      []string
	IsDefault  bool
	IsNamed    bool
	Line       int
	IsUsed     bool
}

// ExportStmt represents an export statement
type ExportStmt struct {
	Name      string
	IsDefault bool
	Line      int
}

// Comment represents a comment
type Comment struct {
	Text       string
	Line       int
	IsInline   bool
	IsBlock    bool
	IsJSDoc    bool
}

// VariableDef represents a variable definition
type VariableDef struct {
	Name     string
	Line     int
	Kind     string // const, let, var
	IsExported bool
	IsUsed   bool
}

type cachedFile struct {
	parsed   *ParsedFile
	modTime  time.Time
	filePath string
}

// Cache holds cached parsed files
type Cache struct {
	cache  map[string]*cachedFile
	mu     sync.RWMutex
	maxAge time.Duration
}

func NewCache(maxAge time.Duration) *Cache {
	if maxAge == 0 {
		maxAge = 5 * time.Minute
	}
	return &Cache{
		cache:  make(map[string]*cachedFile),
		maxAge: maxAge,
	}
}

func (c *Cache) Get(filePath string) (*ParsedFile, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.cache[filePath]
	if !exists {
		return nil, false
	}

	if time.Since(cached.modTime) > c.maxAge {
		delete(c.cache, filePath)
		return nil, false
	}

	return cached.parsed, true
}

func (c *Cache) Set(filePath string, parsed *ParsedFile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stat, err := os.Stat(filePath)
	if err != nil {
		return
	}

	c.cache[filePath] = &cachedFile{
		parsed:   parsed,
		modTime:  stat.ModTime(),
		filePath: filePath,
	}
}
