package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentlint/agentlint/internal/config"
	"github.com/agentlint/agentlint/internal/core"
	"github.com/agentlint/agentlint/internal/languages"
	"github.com/agentlint/agentlint/internal/languages/go"
	"github.com/agentlint/agentlint/internal/output"
)

func main() {
	// Define command line flags
	var (
		configPath   = flag.String("config", "", "Path to configuration file")
		outputFormat = flag.String("format", "console", "Output format (console, json)")
		outputFile   = flag.String("output", "", "Output file (default: stdout)")
		verbose      = flag.Bool("verbose", false, "Verbose output")
		version      = flag.Bool("version", false, "Show version information")
		help         = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	// Show version information
	if *version {
		fmt.Println("AgentLint v0.1.0")
		fmt.Println("A linter for detecting LLM code bad smells")
		return
	}

	// Show help information
	if *help {
		showHelp()
		return
	}

	// Get the path to analyze
	path := "."
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}

	// Make path absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Path does not exist: %s\n", absPath)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Override config with command line options
	if *verbose {
		cfg.Output.Verbose = true
	}
	if *outputFormat != "" {
		cfg.Output.Format = *outputFormat
	}

	// Create context
	ctx := context.Background()

	// Initialize language registry
	registry := languages.NewRegistry()
	
	// Register Go analyzer
	goAnalyzer := go.NewAnalyzer(cfg)
	registry.Register(goAnalyzer)

	// Create file scanner
	goScanner := go.NewFileScanner()

	// Scan for files
	fmt.Printf("Scanning %s...\n", absPath)
	filesByLanguage, err := goScanner.ScanForRegistry(ctx, absPath, registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning files: %v\n", err)
		os.Exit(1)
	}

	// Analyze files
	var allResults []core.Result
	totalFiles := 0

	for language, files := range filesByLanguage {
		analyzer, exists := registry.GetAnalyzer(language)
		if !exists {
			continue
		}

		fmt.Printf("Analyzing %d %s files...\n", len(files), language)
		totalFiles += len(files)

		for _, file := range files {
			results, err := analyzer.Analyze(ctx, file, cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing file %s: %v\n", file, err)
				continue
			}
			allResults = append(allResults, results...)
		}
	}

	// Create output formatter
	var formatter output.Formatter
	switch cfg.Output.Format {
	case "json":
		formatter = output.NewJSONFormatter(cfg.Output.Verbose)
	case "console":
		fallthrough
	default:
		formatter = output.NewConsoleFormatter(cfg.Output.Verbose)
	}

	// Set output destination
	var outputFileHandle *os.File
	if *outputFile != "" {
		outputFileHandle, err = os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer outputFileHandle.Close()
		os.Stdout = outputFileHandle
	}

	// Print results
	formatter.PrintHeader()
	if err := formatter.Format(allResults); err != nil {
		formatter.FormatError(err)
		os.Exit(1)
	}
	formatter.PrintFooter()

	// Exit with error code if issues were found
	if len(allResults) > 0 {
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("AgentLint - A linter for detecting LLM code bad smells")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agentlint [flags] [path]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -config string    Path to configuration file")
	fmt.Println("  -format string    Output format (console, json) (default \"console\")")
	fmt.Println("  -output string    Output file (default: stdout)")
	fmt.Println("  -verbose          Verbose output")
	fmt.Println("  -version          Show version information")
	fmt.Println("  -help             Show help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  agentlint ./myproject")
	fmt.Println("  agentlint -format json -output report.json ./myproject")
	fmt.Println("  agentlint -config agentlint.yaml ./myproject")
}