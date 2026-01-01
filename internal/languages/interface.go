package languages

import (
	"context"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

// Registry holds all available language analyzers
type Registry struct {
	analyzers map[string]core.Analyzer
}

// NewRegistry creates a new language registry
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]core.Analyzer),
	}
}

// Register registers a language analyzer
func (r *Registry) Register(analyzer core.Analyzer) {
	r.analyzers[analyzer.Name()] = analyzer
}

// GetAnalyzer returns an analyzer for the given language
func (r *Registry) GetAnalyzer(language string) (core.Analyzer, bool) {
	analyzer, exists := r.analyzers[language]
	return analyzer, exists
}

// GetAnalyzerByExtension returns an analyzer for the given file extension
func (r *Registry) GetAnalyzerByExtension(extension string) (core.Analyzer, bool) {
	for _, analyzer := range r.analyzers {
		for _, ext := range analyzer.SupportedExtensions() {
			if ext == extension {
				return analyzer, true
			}
		}
	}
	return nil, false
}

// GetAllAnalyzers returns all registered analyzers
func (r *Registry) GetAllAnalyzers() map[string]core.Analyzer {
	result := make(map[string]core.Analyzer)
	for k, v := range r.analyzers {
		result[k] = v
	}
	return result
}

// FileScanner scans directories for files and groups them by language
type FileScanner struct {
	registry *Registry
}

// NewFileScanner creates a new file scanner
func NewFileScanner(registry *Registry) *FileScanner {
	return &FileScanner{
		registry: registry,
	}
}

// Scan scans a directory and returns a map of language to file paths
func (s *FileScanner) Scan(ctx context.Context, rootPath string) (map[string][]string, error) {
	// This will be implemented to walk the directory tree
	// and group files by their language based on extension
	// For now, return an empty map
	return make(map[string][]string), nil
}

// AnalyzerFactory creates language-specific analyzers
type AnalyzerFactory interface {
	CreateAnalyzer(config core.Config) core.Analyzer
	GetLanguageName() string
	GetSupportedExtensions() []string
}
