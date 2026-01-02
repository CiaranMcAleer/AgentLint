package output_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/output"
)

func generateTestResults(count int) []core.Result {
	results := make([]core.Result, count)
	for i := 0; i < count; i++ {
		results[i] = core.Result{
			RuleID:     "test-rule",
			RuleName:   "Test Rule",
			Category:   "test",
			Severity:   "warning",
			FilePath:   "/path/to/file.go",
			Line:       i + 1,
			Column:     10,
			Message:    "This is a test message for benchmarking",
			Suggestion: "Consider fixing this issue",
		}
	}
	return results
}

func generateTestResultsMultiFile(fileCount, issuesPerFile int) []core.Result {
	results := make([]core.Result, 0, fileCount*issuesPerFile)
	severities := []string{"error", "warning", "info"}
	for f := 0; f < fileCount; f++ {
		for i := 0; i < issuesPerFile; i++ {
			results = append(results, core.Result{
				RuleID:     "test-rule",
				RuleName:   "Test Rule",
				Category:   "test",
				Severity:   severities[i%3],
				FilePath:   "/path/to/file" + string(rune('A'+f)) + ".go",
				Line:       i + 1,
				Column:     10,
				Message:    "This is a test message for benchmarking",
				Suggestion: "Consider fixing this issue",
			})
		}
	}
	return results
}

func BenchmarkNewConsoleFormatter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = output.NewConsoleFormatter(true)
	}
}

func BenchmarkNewJSONFormatter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = output.NewJSONFormatter(true)
	}
}

// formatterBenchCase defines a benchmark test case for formatter benchmarks
type formatterBenchCase struct {
	name    string
	results []core.Result
	verbose bool
}

func BenchmarkConsoleFormatter_Format(b *testing.B) {
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	cases := []formatterBenchCase{
		{"Empty", []core.Result{}, false},
		{"10Results", generateTestResults(10), false},
		{"100Results", generateTestResults(100), false},
		{"100ResultsVerbose", generateTestResults(100), true},
		{"MultiFile", generateTestResultsMultiFile(10, 10), false},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			formatter := output.NewConsoleFormatter(tc.verbose)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = formatter.Format(tc.results)
			}
		})
	}
}

func BenchmarkJSONFormatter_Format(b *testing.B) {
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.Run("Empty", func(b *testing.B) {
		formatter := output.NewJSONFormatter(false)
		results := []core.Result{}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = formatter.Format(results)
		}
	})

	b.Run("10Results", func(b *testing.B) {
		formatter := output.NewJSONFormatter(false)
		results := generateTestResults(10)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = formatter.Format(results)
		}
	})

	b.Run("100Results", func(b *testing.B) {
		formatter := output.NewJSONFormatter(false)
		results := generateTestResults(100)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = formatter.Format(results)
		}
	})

	b.Run("1000Results", func(b *testing.B) {
		formatter := output.NewJSONFormatter(false)
		results := generateTestResults(1000)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = formatter.Format(results)
		}
	})
}

func BenchmarkConsoleFormatter_FormatError(b *testing.B) {
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() { os.Stderr = oldStderr }()

	formatter := output.NewConsoleFormatter(false)
	err := io.EOF

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatter.FormatError(err)
	}
}

func BenchmarkJSONFormatter_FormatError(b *testing.B) {
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	formatter := output.NewJSONFormatter(false)
	err := io.EOF

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatter.FormatError(err)
	}
}

func BenchmarkConsoleFormatter_PrintHeader(b *testing.B) {
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	formatter := output.NewConsoleFormatter(false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.PrintHeader()
	}
}

func BenchmarkConsoleFormatter_PrintFooter(b *testing.B) {
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	formatter := output.NewConsoleFormatter(false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.PrintFooter()
	}
}

// Helper to silence output for benchmarks
var _ = bytes.Buffer{}
