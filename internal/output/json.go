package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/agentlint/agentlint/internal/core"
)

// JSONFormatter formats results as JSON
type JSONFormatter struct {
	verbose bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(verbose bool) *JSONFormatter {
	return &JSONFormatter{
		verbose: verbose,
	}
}

// JSONOutput represents the structure of JSON output
type JSONOutput struct {
	Summary   Summary              `json:"summary"`
	Results   []core.Result        `json:"results"`
	Errors    []string             `json:"errors,omitempty"`
	Timestamp string               `json:"timestamp"`
}

// Summary contains summary information about the analysis
type Summary struct {
	TotalIssues int `json:"total_issues"`
	ErrorCount  int `json:"error_count"`
	WarnCount   int `json:"warning_count"`
	InfoCount   int `json:"info_count"`
	FileCount   int `json:"file_count"`
}

// Format formats the results as JSON
func (f *JSONFormatter) Format(results []core.Result) error {
	// Calculate summary
	summary := Summary{
		TotalIssues: len(results),
		ErrorCount:  0,
		WarnCount:   0,
		InfoCount:   0,
		FileCount:   0,
	}

	fileSet := make(map[string]bool)
	for _, result := range results {
		switch result.Severity {
		case "error":
			summary.ErrorCount++
		case "warning":
			summary.WarnCount++
		case "info":
			summary.InfoCount++
		}
		fileSet[result.FilePath] = true
	}
	summary.FileCount = len(fileSet)

	// Create output structure
	output := JSONOutput{
		Summary:   summary,
		Results:   results,
		Timestamp: getCurrentTimestamp(),
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Print to stdout
	fmt.Println(string(jsonData))
	return nil
}

// FormatError formats an error as JSON
func (f *JSONFormatter) FormatError(err error) error {
	errorOutput := JSONOutput{
		Summary: Summary{
			TotalIssues: 0,
			ErrorCount:  0,
			WarnCount:   0,
			InfoCount:   0,
			FileCount:   0,
		},
		Results:   []core.Result{},
		Errors:    []string{err.Error()},
		Timestamp: getCurrentTimestamp(),
	}

	jsonData, marshalErr := json.MarshalIndent(errorOutput, "", "  ")
	if marshalErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal error JSON: %v\n", marshalErr)
		return err
	}

	fmt.Println(string(jsonData))
	return err
}

// PrintHeader prints a header for the analysis (no-op for JSON)
func (f *JSONFormatter) PrintHeader() {
	// No header for JSON output
}

// PrintFooter prints a footer for the analysis (no-op for JSON)
func (f *JSONFormatter) PrintFooter() {
	// No footer for JSON output
}

// getCurrentTimestamp returns the current timestamp in ISO 8601 format
func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", 0) // Placeholder - would use time.Now().Format(time.RFC3339)
}