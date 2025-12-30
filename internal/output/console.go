package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/agentlint/agentlint/internal/core"
)

// ConsoleFormatter formats results for console output
type ConsoleFormatter struct {
	verbose bool
}

// NewConsoleFormatter creates a new console formatter
func NewConsoleFormatter(verbose bool) *ConsoleFormatter {
	return &ConsoleFormatter{
		verbose: verbose,
	}
}

// Format formats the results for console output
func (f *ConsoleFormatter) Format(results []core.Result) error {
	if len(results) == 0 {
		fmt.Println("No issues found!")
		return nil
	}

	fileResults := groupResultsByFile(results)

	fmt.Printf("Found %d issues across %d files\n\n", len(results), len(fileResults))

	f.printResultsByFile(fileResults)
	f.printSummary(results)

	return nil
}

func groupResultsByFile(results []core.Result) map[string][]core.Result {
	fileResults := make(map[string][]core.Result)
	for _, result := range results {
		fileResults[result.FilePath] = append(fileResults[result.FilePath], result)
	}
	return fileResults
}

func (f *ConsoleFormatter) printResultsByFile(fileResults map[string][]core.Result) {
	for filePath, fileIssues := range fileResults {
		fmt.Printf("%s (%d issues):\n", filePath, len(fileIssues))

		for _, issue := range fileIssues {
			severity := formatSeverity(issue.Severity)
			fmt.Printf("  %s:%d: %s [%s]\n", filePath, issue.Line, issue.Message, severity)

			if f.verbose && issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	}
}

func formatSeverity(severity string) string {
	switch severity {
	case "error":
		return "ERROR"
	case "warning":
		return "WARN"
	case "info":
		return "INFO"
	default:
		return severity
	}
}

type severityCounts struct {
	errors   int
	warnings int
	info     int
}

func countSeverities(results []core.Result) severityCounts {
	var counts severityCounts
	for _, result := range results {
		switch result.Severity {
		case "error":
			counts.errors++
		case "warning":
			counts.warnings++
		case "info":
			counts.info++
		}
	}
	return counts
}

func (f *ConsoleFormatter) printSummary(results []core.Result) {
	counts := countSeverities(results)

	if counts.errors > 0 || counts.warnings > 0 || counts.info > 0 {
		fmt.Println("Summary:")
		if counts.errors > 0 {
			fmt.Printf("  Errors: %d\n", counts.errors)
		}
		if counts.warnings > 0 {
			fmt.Printf("  Warnings: %d\n", counts.warnings)
		}
		if counts.info > 0 {
			fmt.Printf("  Info: %d\n", counts.info)
		}
	}
}

// FormatError formats an error for console output
func (f *ConsoleFormatter) FormatError(err error) error {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return nil
}

// PrintHeader prints a header for the analysis
func (f *ConsoleFormatter) PrintHeader() {
	fmt.Println("AgentLint - LLM Code Smell Detector")
	fmt.Println(strings.Repeat("=", 40))
}

// PrintFooter prints a footer for the analysis
func (f *ConsoleFormatter) PrintFooter() {
	fmt.Println("\nAnalysis complete.")
}
