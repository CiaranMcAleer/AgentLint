# AgentLint: A Static Analysis Tool for Detecting LLM-Generated Code Anti-Patterns

## Abstract

AgentLint is a command-line static analysis tool designed to identify code quality issues commonly introduced by large language model code generators. The tool implements a modular architecture supporting multiple detection rules organized into three categories: size violations, documentation anti-patterns, and orphaned code artifacts. This document describes the tool's design, configuration options, and detection capabilities.

## 1. Introduction

The proliferation of large language models for code generation has introduced a new category of code quality concerns. Generated code frequently exhibits patterns that, while syntactically correct, violate established software engineering practices. These patterns include excessively large functions, excessive commenting that reduces signal-to-noise ratio, and the retention of unused code artifacts.

AgentLint addresses these concerns through automated static analysis. The tool processes Go source code files and reports violations of configurable quality thresholds. The architecture is designed for extensibility, enabling future support for additional programming languages and detection rules.

## 2. Features

AgentLint provides the following detection capabilities:

**Size Analysis**
- Large Function Detection: Identifies functions exceeding configurable line thresholds
- Large File Detection: Identifies source files exceeding configurable line thresholds

**Documentation Analysis**
- Overcommenting Detection: Calculates comment-to-code ratios and flags excessive documentation
- Redundant Comment Identification: Detects comments that merely restate code functionality
- Missing Documentation Detection: Flags exported functions lacking documentation

**Code Quality Analysis**
- Unused Function Detection: Identifies functions that are defined but not referenced
- Unused Variable Detection: Identifies variables declared but not utilized
- Unreachable Code Detection: Identifies code paths that cannot be executed
- Dead Import Detection: Identifies import statements that are not referenced

## 3. Installation

AgentLint is distributed as a Go binary. Installation is performed via the Go toolchain:

```bash
go install github.com/CiaranMcAleer/AgentLint/cmd/agentlint@latest
```

The binary is installed to `$GOPATH/bin` and is immediately available for use.

## 4. Usage

### 4.1 Basic Usage

The tool operates on directories containing Go source code:

```bash
# Analyze the current directory
agentlint

# Analyze a specific directory
agentlint ./path/to/go/project

# Use a custom configuration file
agentlint -config agentlint.yaml ./myproject

# Output results in JSON format
agentlint -format json -output report.json ./myproject
```

### 4.2 Command Line Options

The following command line options are available:

| Option | Description | Default |
|--------|-------------|---------|
| -config | Path to configuration file | agentlint.yaml |
| -format | Output format (console, json) | console |
| -output | Output file path | stdout |
| -verbose | Enable verbose output | false |
| -version | Display version information | - |
| -help | Display help information | - |

## 5. Configuration

AgentLint behavior is controlled through YAML configuration files. The tool searches for `agentlint.yaml` or `agentlint.yml` in the current directory when no explicit configuration is provided.

### 5.1 Configuration Schema

```yaml
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
    checkRedundant: true
    checkDocCoverage: true

  orphanedCode:
    enabled: true
    checkUnusedFunctions: true
    checkUnusedVariables: true
    checkUnreachableCode: true
    checkDeadImports: true

output:
  format: "console"
  verbose: false

language:
  go:
    ignoreTests: false
```

### 5.2 Rule Configuration

Each rule category supports the following options:

**functionSize**: Controls large function detection
- `enabled`: Enable or disable the rule
- `maxLines`: Maximum permitted function size

**fileSize**: Controls large file detection
- `enabled`: Enable or disable the rule
- `maxLines`: Maximum permitted file size

**overcommenting**: Controls documentation analysis
- `enabled`: Enable or disable the rule
- `maxCommentRatio`: Maximum comment-to-code ratio (0.0 to 1.0)
- `checkRedundant`: Enable redundant comment detection
- `checkDocCoverage`: Enable missing documentation detection

**orphanedCode**: Controls code quality analysis
- `enabled`: Enable or disable the rule
- `checkUnusedFunctions`: Enable unused function detection
- `checkUnusedVariables`: Enable unused variable detection
- `checkUnreachableCode`: Enable unreachable code detection
- `checkDeadImports`: Enable dead import detection

## 6. Detection Rules

### 6.1 Size Rules

**Large Function Rule**
Detects functions exceeding the configured line threshold. Functions exceeding this threshold typically indicate excessive complexity and should be decomposed into smaller, focused units.

**Large File Rule**
Detects files exceeding the configured line threshold. Large files often indicate poor code organization and should be split into multiple focused modules.

### 6.2 Documentation Rules

**Overcommenting Rule**
Calculates the ratio of comment lines to code lines. Files exceeding the configured ratio are flagged as potentially over-documented. Excessive commenting can reduce code readability by obscuring the actual implementation.

**Redundant Comment Rule**
Identifies comments that restate code functionality without adding semantic value. Examples include comments that merely translate variable names or restate control flow logic.

**Missing Documentation Rule**
Identifies exported (public) functions lacking documentation comments. This rule enforces documentation standards for public APIs.

### 6.3 Orphaned Code Rules

**Unused Function Rule**
Identifies functions that are defined but not referenced within the analyzed codebase. Note that this analysis is performed on a per-file basis and may not detect cross-file references.

**Unused Variable Rule**
Identifies variables that are declared but never used. While the Go compiler enforces unused variable detection for local variables, this rule provides additional analysis capabilities.

**Unreachable Code Rule**
Identifies code statements following unconditional return statements. Such code represents logical errors and should be removed.

**Dead Import Rule**
Identifies import statements that are not referenced within the file. Unused imports should be removed to reduce compilation overhead and improve maintainability.

## 7. Output Formats

### 7.1 Console Output

The console formatter provides human-readable output organized by file:

```
AgentLint - LLM Code Smell Detector
========================================
Found 3 issues across 2 files

main.go (2 issues):
  main.go:15: Function 'processData' is too large (75 lines, max 50) [WARN]
  main.go:25: Variable 'tempData' is declared but never used [WARN]

utils.go (1 issues):
  utils.go:10: Function 'helperFunction' is defined but never used [WARN]

Summary:
  Warnings: 3

Analysis complete.
```

### 7.2 JSON Output

The JSON formatter provides structured output suitable for integration with other tools:

```json
{
  "summary": {
    "total_issues": 3,
    "error_count": 0,
    "warning_count": 3,
    "info_count": 0,
    "file_count": 2
  },
  "results": [
    {
      "rule_id": "large-function",
      "rule_name": "Large Function",
      "category": "size",
      "severity": "warning",
      "file_path": "main.go",
      "line": 15,
      "column": 0,
      "message": "Function 'processData' is too large (75 lines, max 50)",
      "suggestion": "Consider breaking down function 'processData' into smaller functions"
    }
  ],
  "timestamp": "2023-12-20T23:41:00Z"
}
```

## 8. Architecture

AgentLint is built on a modular, language-agnostic architecture comprising the following components:

**Core Framework**
Provides language-agnostic interfaces and type definitions. This layer defines theAnalyzer and Rule interfaces that all language implementations must satisfy.

**Language Support**
Pluggable analyzer implementations for different programming languages. The current implementation supports Go. Additional language support can be added by implementing the Analyzer interface.

**Rule Engine**
Extensible rule system for detecting code quality issues. Rules are organized into categories and implement the Rule interface. New rules can be added without modifying core components.

**Configuration Management**
YAML-based configuration system with support for rule-specific parameters. Configuration is validated at startup and defaults are applied for unspecified options.

**Output Formatters**
Multiple output format support through a formatter interface. The console formatter provides human-readable output while the JSON formatter provides structured data suitable for programmatic processing.

## 9. Extending AgentLint

The architecture supports extension in two primary dimensions:

### 9.1 Adding New Rules

New rules are implemented by satisfying the Rule interface:

```go
type Rule interface {
    ID() string
    Name() string
    Description() string
    Category() RuleCategory
    Severity() Severity
    Check(ctx context.Context, node interface{}, config Config) *Result
}
```

Rules are registered with the analyzer during initialization.

### 9.2 Adding New Languages

New language support requires implementing the Analyzer interface:

```go
type Analyzer interface {
    Analyze(ctx context.Context, filePath string, config Config) ([]Result, error)
    SupportedExtensions() []string
    Name() string
}
```

The analyzer handles file parsing and metric calculation, then delegates to registered rules for issue detection.

## 10. Contributing

Contributions are welcome. For significant changes, please open an issue to discuss the proposed modifications before implementation. Pull requests should include appropriate test coverage and documentation updates.

## 11. License

This project is released into the public domain. See the LICENSE file for details.
