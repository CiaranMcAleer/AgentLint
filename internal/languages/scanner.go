package languages

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// MultiScanner scans directories for files of multiple languages
type MultiScanner struct {
	registry   *Registry
	ignoreDirs []string
}

// NewMultiScanner creates a new multi-language file scanner
func NewMultiScanner(registry *Registry) *MultiScanner {
	return &MultiScanner{
		registry: registry,
		ignoreDirs: []string{
			".git",
			"node_modules",
			"vendor",
			".vscode",
			".idea",
			"__pycache__",
			".venv",
			"venv",
			"env",
			".env",
			".tox",
			".eggs",
			"dist",
			"build",
			".pytest_cache",
			".mypy_cache",
			".cache",
		},
	}
}

// Scan scans a directory and returns files grouped by language
func (s *MultiScanner) Scan(ctx context.Context, rootPath string) (map[string][]string, error) {
	filesByLanguage := make(map[string][]string)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if info.IsDir() {
			// Skip ignored directories
			for _, ignoreDir := range s.ignoreDirs {
				if info.Name() == ignoreDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Get file extension
		ext := filepath.Ext(path)
		if ext == "" {
			return nil
		}

		// Find analyzer for this extension
		analyzer, exists := s.registry.GetAnalyzerByExtension(ext)
		if !exists {
			return nil
		}

		// Group file by language
		language := analyzer.Name()
		filesByLanguage[language] = append(filesByLanguage[language], path)

		return nil
	})

	return filesByLanguage, err
}

// AddIgnoreDir adds a directory pattern to ignore during scanning
func (s *MultiScanner) AddIgnoreDir(dir string) {
	s.ignoreDirs = append(s.ignoreDirs, dir)
}

// SetIgnoreDirs sets the list of directories to ignore during scanning
func (s *MultiScanner) SetIgnoreDirs(dirs []string) {
	s.ignoreDirs = dirs
}

// ScanWithFilter scans a directory with a custom filter function
func (s *MultiScanner) ScanWithFilter(ctx context.Context, rootPath string, filter func(path string) bool) (map[string][]string, error) {
	filesByLanguage := make(map[string][]string)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return s.processFileWithFilter(ctx, path, info, filter, filesByLanguage)
	})

	return filesByLanguage, err
}

// processFileWithFilter processes a single file during filtered scanning
func (s *MultiScanner) processFileWithFilter(ctx context.Context, path string, info os.FileInfo, filter func(path string) bool, filesByLanguage map[string][]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if info.IsDir() {
		return s.handleDirectory(info)
	}

	if filter != nil && !filter(path) {
		return nil
	}

	return s.addFileToLanguageMap(path, filesByLanguage)
}

// handleDirectory checks if directory should be skipped
func (s *MultiScanner) handleDirectory(info os.FileInfo) error {
	for _, ignoreDir := range s.ignoreDirs {
		if info.Name() == ignoreDir {
			return filepath.SkipDir
		}
	}
	return nil
}

// addFileToLanguageMap adds a file to the language map if it has a supported extension
func (s *MultiScanner) addFileToLanguageMap(path string, filesByLanguage map[string][]string) error {
	ext := filepath.Ext(path)
	if ext == "" {
		return nil
	}

	analyzer, exists := s.registry.GetAnalyzerByExtension(ext)
	if !exists {
		return nil
	}

	language := analyzer.Name()
	filesByLanguage[language] = append(filesByLanguage[language], path)
	return nil
}

// ScanForLanguage scans a directory for files of a specific language
func (s *MultiScanner) ScanForLanguage(ctx context.Context, rootPath string, language string) ([]string, error) {
	analyzer, exists := s.registry.GetAnalyzer(language)
	if !exists {
		return nil, nil
	}

	extensions := analyzer.SupportedExtensions()
	extSet := make(map[string]bool)
	for _, ext := range extensions {
		extSet[ext] = true
	}

	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			for _, ignoreDir := range s.ignoreDirs {
				if info.Name() == ignoreDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		if extSet[ext] {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// IgnoreTestFiles returns a filter function that ignores test files
func IgnoreTestFiles(language string) func(path string) bool {
	return func(path string) bool {
		base := filepath.Base(path)
		switch language {
		case "go":
			return !strings.HasSuffix(base, "_test.go")
		case "python":
			return !strings.HasPrefix(base, "test_") && !strings.HasSuffix(base, "_test.py")
		default:
			return true
		}
	}
}
