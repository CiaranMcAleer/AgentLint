# AgentLint

A CLI-based linter written in Go, designed to detect LLM-generated code bad smells. The MVP focuses on Go projects with a flexible architecture to support multiple languages in the future.

## Features

- **Large Function Detection**: Flags functions that exceed configurable line limits
- **Large File Detection**: Identifies files that are too large
- **Overcommenting Analysis**: Detects excessive comments and redundant documentation
- **Orphaned Code Detection**: Finds unused functions, variables, unreachable code, and dead imports
- **Configurable Rules**: All detection rules are configurable via YAML files
- **Multiple Output Formats**: Supports console and JSON output formats
- **Extensible Architecture**: Designed for easy addition of new languages and rules

## Installation

```bash
go install github.com/agentlint/agentlint/cmd/agentlint@latest
```

## Usage

### Basic Usage

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

### Command Line Options

```
-config string    Path to configuration file
-format string    Output format (console, json) (default "console")
-output string    Output file (default: stdout)
-verbose          Verbose output
-version          Show version information
-help             Show help information
```

## Configuration

AgentLint uses a YAML configuration file to customize rule behavior. If no configuration file is provided, it will look for `agentlint.yaml` or `agentlint.yml` in the current directory.

### Example Configuration

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

## Detection Rules

### Size Rules

- **Large Function**: Detects functions with more than the configured number of lines
- **Large File**: Detects files with more than the configured number of lines

### Comment Rules

- **Overcommenting**: Flags files with excessive comment-to-code ratios
- **Redundant Comments**: Identifies comments that simply restate what the code does
- **Missing Documentation**: Flags exported functions without documentation

### Orphaned Code Rules

- **Unused Functions**: Finds functions that are defined but never called
- **Unused Variables**: Identifies variables that are declared but never used
- **Unreachable Code**: Detects code that can never be executed
- **Dead Imports**: Finds import statements that are never used

## Output Formats

### Console Output

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

### JSON Output

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

## Architecture

AgentLint is built with a modular, language-agnostic architecture:

- **Core Framework**: Language-agnostic interfaces and types
- **Language Support**: Pluggable language analyzers (currently Go)
- **Rule Engine**: Extensible rule system for detecting code smells
- **Configuration**: Flexible YAML-based configuration
- **Output Formatters**: Multiple output format support

The architecture is designed to make adding new languages and rules straightforward, enabling future expansion beyond Go.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is released into the public domain. See the [LICENSE](LICENSE) file for details.
