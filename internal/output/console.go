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

	// Group results by file
	fileResults := make(map[string][]core.Result)
	for _, result := range results {
		fileResults[result.FilePath] = append(fileResults[result.FilePath], result)
	}

	// Print summary
	fmt.Printf("Found %d issues across %d files\n\n", len(results), len(fileResults))

	// Print results by file
	for filePath, fileIssues := range fileResults {
		fmt.Printf("%s (%d issues):\n", filePath, len(fileIssues))
		
		for _, issue := range fileIssues {
			severity := issue.Severity
			switch severity {
			case "error":
				severity = "ERROR"
			case "warning":
				severity = "WARN"
			case "info":
				severity = "INFO"
			}
			
			fmt.Printf("  %s:%d: %s [%s]\n", filePath, issue.Line, issue.Message, severity)
			
			if f.verbose && issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	}

	// Print summary by severity
	errorCount := 0
	warningCount := 0
	infoCount := 0
	
	for _, result := range results {
		switch result.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		case "info":
			infoCount++
		}
	}
	
	if errorCount > 0 || warningCount > 0 || infoCount > 0 {
		fmt.Println("Summary:")
		if errorCount > 0 {
			fmt.Printf("  Errors: %d\n", errorCount)
		}
		if warningCount > 0 {
			fmt.Printf("  Warnings: %d\n", warningCount)
		}
		if infoCount > 0 {
			fmt.Printf("  Info: %d\n", infoCount)
		}
	}

	return nil
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