package python

import (
	"os"
	"sync"
	"time"
)

// ParsedFile represents a parsed Python file
type ParsedFile struct {
	Lines        []string
	Functions    []FunctionDef
	Classes      []ClassDef
	Imports      []ImportStmt
	Comments     []Comment
	Docstrings   []Docstring
	Variables    []VariableDef
	TotalLines   int
	CodeLines    int
	CommentLines int
	BlankLines   int
}

// FunctionDef represents a Python function definition
type FunctionDef struct {
	Name       string
	StartLine  int
	EndLine    int
	Parameters []string
	Decorators []string
	IsMethod   bool
	IsPrivate  bool
	ClassName  string
	Indent     int
}

// ClassDef represents a Python class definition
type ClassDef struct {
	Name       string
	StartLine  int
	EndLine    int
	Bases      []string
	Decorators []string
	Methods    []FunctionDef
}

// ImportStmt represents a Python import statement
type ImportStmt struct {
	Module string
	Names  []string
	IsFrom bool
	Line   int
	IsUsed bool
}

// Comment represents a Python comment
type Comment struct {
	Text     string
	Line     int
	IsInline bool
}

// Docstring represents a Python docstring
type Docstring struct {
	Text      string
	StartLine int
	EndLine   int
	Owner     string
}

// VariableDef represents a Python variable definition
type VariableDef struct {
	Name     string
	Line     int
	IsGlobal bool
	IsUsed   bool
}

// cachedFile represents a cached parsed file
type cachedFile struct {
	parsed   *ParsedFile
	modTime  time.Time
	filePath string
}

// Cache holds cached parsed files with time-based expiration
type Cache struct {
	cache  map[string]*cachedFile
	mu     sync.RWMutex
	maxAge time.Duration
}

// NewCache creates a new cache with the specified max age
func NewCache(maxAge time.Duration) *Cache {
	if maxAge == 0 {
		maxAge = 5 * time.Minute
	}
	return &Cache{
		cache:  make(map[string]*cachedFile),
		maxAge: maxAge,
	}
}

// Get retrieves a cached parsed file if it exists and hasn't expired
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

// Set stores a parsed file in the cache
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
