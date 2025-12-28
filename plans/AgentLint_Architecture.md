# AgentLint Architecture Plan

## Overview
AgentLint is a CLI-based linter written in Go, designed to detect LLM-generated code bad smells. The MVP will focus on Go projects with a flexible architecture to support multiple languages in the future.

## Core Architecture

### 1. Project Structure
```
agentlint/
├── cmd/
│   └── agentlint/
│       └── main.go              # CLI entry point
├── internal/
│   ├── core/                    # Core interfaces and types
│   │   ├── analyzer.go          # Main analyzer interface
│   │   ├── rule.go              # Rule interface
│   │   └── result.go            # Result types
│   ├── languages/               # Language-specific implementations
│   │   ├── go/                  # Go language support
│   │   │   ├── parser.go        # Go AST parser
│   │   │   ├── analyzer.go      # Go-specific analyzer
│   │   │   └── rules/           # Go-specific rules
│   │   └── interface.go         # Language interface definition
│   ├── rules/                   # Common rule implementations
│   │   ├── size.go              # Large function/file detection
│   │   ├── comments.go          # Overcommenting detection
│   │   └── orphaned.go          # Orphaned code detection
│   ├── config/                  # Configuration management
│   │   └── config.go
│   └── output/                  # Output formatters
│       ├── console.go
│       └── json.go
├── pkg/
│   └── api/                     # Public API
├── test/                        # Test files
├── configs/                     # Default configurations
└── docs/                        # Documentation
```

### 2. Core Components

#### 2.1 Language-Agnostic Framework
```go
// Analyzer interface for language-specific implementations
type Analyzer interface {
    Analyze(filePath string, config Config) ([]Result, error)
    SupportedExtensions() []string
}

// Rule interface for individual detection rules
type Rule interface {
    Name() string
    Description() string
    Check(node interface{}, config Config) *Result
}
```

#### 2.2 Go-Specific Implementation
- Utilize Go's `go/parser` and `go/ast` packages for AST parsing
- Implement Go-specific analyzers that understand Go constructs
- Create rules tailored to Go code patterns

#### 2.3 Rule Engine
- Pluggable rule system allowing easy addition of new rules
- Rule categories: Size, Comments, Orphaned Code
- Configurable thresholds for each rule

### 3. Detection Rules (MVP)

#### 3.1 Large Function/File Detection
- **Function Size**: Default threshold of 50 lines (configurable)
- **File Size**: Default threshold of 500 lines (configurable)
- **Complexity**: Cyclomatic complexity threshold (future enhancement)

#### 3.2 Overcommenting Detection
- **Comment-to-Code Ratio**: Default threshold of 30% (configurable)
- **Redundant Comments**: Detect comments that restate obvious code
- **Missing Documentation**: Flag exported functions without comments

#### 3.3 Orphaned Code Detection
- **Unused Functions**: Functions defined but never called
- **Unused Variables**: Variables declared but never used
- **Unreachable Code**: Code that can never be executed
- **Dead Imports**: Import statements that aren't used

### 4. Configuration System
```yaml
# agentlint.yaml example
rules:
  functionSize:
    enabled: true
    maxLines: 50
  fileSize:
    enabled: true
    maxLines: 500
  overcommenting:
    enabled: true
    maxCommentRatio: 0.3
  orphanedCode:
    enabled: true
    checkUnusedFunctions: true
    checkUnusedVariables: true
    checkUnreachableCode: true

output:
  format: "console" # console, json
  verbose: false
```

### 5. CLI Interface
```bash
# Basic usage
agentlint ./path/to/go/project

# With configuration
agentlint -config agentlint.yaml ./path/to/go/project

# Output options
agentlint -format json -output report.json ./path/to/go/project
```

### 6. Output Formats
- **Console**: Human-readable output with file paths, line numbers, and descriptions
- **JSON**: Machine-readable format for CI/CD integration

## Extensibility for Future Languages

### 1. Language Interface
Each new language will implement the `Analyzer` interface:
```go
type LanguageAnalyzer struct {
    // Language-specific implementation
}

func (l *LanguageAnalyzer) Analyze(filePath string, config Config) ([]Result, error) {
    // Parse language-specific AST
    // Apply rules
    // Return results
}
```

### 2. Plugin System (Future Enhancement)
- Dynamic loading of language modules
- Registry pattern for language analyzers
- Configuration-driven rule selection

## Implementation Phases

### Phase 1: Core Framework
1. Set up project structure
2. Define core interfaces
3. Implement basic CLI
4. Create configuration system

### Phase 2: Go Implementation
1. Implement Go AST parser
2. Create Go-specific analyzer
3. Implement core detection rules
4. Add output formatters

### Phase 3: Testing & Documentation
1. Comprehensive test suite
2. Documentation and examples
3. Performance optimization

### Phase 4: Future Expansion
1. Design language plugin system
2. Implement second language (Python/JavaScript)
3. Advanced rules and auto-fixing capabilities

## Technical Considerations

### Performance
- Parallel file processing using goroutines
- Efficient AST traversal
- Incremental analysis for large projects

### Error Handling
- Graceful handling of malformed code
- Clear error messages for configuration issues
- Partial analysis when some files fail

### Dependencies
- Minimal external dependencies
- Leverage Go's standard library where possible
- Consider using established AST libraries for future languages

This architecture provides a solid foundation for the MVP while ensuring flexibility for future expansion to support multiple languages and advanced features.