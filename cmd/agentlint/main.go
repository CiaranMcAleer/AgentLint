package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentlint/agentlint/internal/core"
	"github.com/agentlint/agentlint/internal/languages"
	"github.com/agentlint/agentlint/internal/languages/golang"
	"github.com/agentlint/agentlint/internal/output"
)

func main() {
	// Define command line flags
	var (
		// Output options
		outputFormat = flag.String("format", "console", "Output format (console, json)")
		outputFile   = flag.String("output", "", "Output file (default: stdout)")
		verbose      = flag.Bool("verbose", false, "Verbose output")
		
		// Function size rules
		funcSizeEnabled   = flag.Bool("enable-func-size", true, "Enable large function detection")
		funcSizeMaxLines  = flag.Int("func-max-lines", 50, "Maximum number of lines for a function")
		
		// File size rules
		fileSizeEnabled   = flag.Bool("enable-file-size", true, "Enable large file detection")
		fileSizeMaxLines  = flag.Int("file-max-lines", 500, "Maximum number of lines for a file")
		
		// Comment rules
		commentEnabled        = flag.Bool("enable-comments", true, "Enable overcommenting detection")
		commentMaxRatio       = flag.Float64("comment-max-ratio", 0.3, "Maximum comment-to-code ratio")
		commentCheckRedundant = flag.Bool("check-redundant", true, "Check for redundant comments")
		commentCheckDoc       = flag.Bool("check-docs", true, "Check for missing documentation")
		
		// Orphaned code rules
		orphanedEnabled           = flag.Bool("enable-orphaned", true, "Enable orphaned code detection")
		orphanedCheckUnusedFuncs  = flag.Bool("check-unused-funcs", true, "Check for unused functions")
		orphanedCheckUnusedVars   = flag.Bool("check-unused-vars", true, "Check for unused variables")
		orphanedCheckUnreachable  = flag.Bool("check-unreachable", true, "Check for unreachable code")
		orphanedCheckDeadImports  = flag.Bool("check-dead-imports", true, "Check for dead imports")
		
		// Go-specific options
		goIgnoreTests = flag.Bool("ignore-tests", false, "Ignore test files during analysis")
		
		// General options
		version = flag.Bool("version", false, "Show version information")
		help    = flag.Bool("help", false, "Show help information")
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

	// Create configuration from CLI flags
	cfg := core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  *funcSizeEnabled,
				MaxLines: *funcSizeMaxLines,
			},
			FileSize: core.FileSizeConfig{
				Enabled:  *fileSizeEnabled,
				MaxLines: *fileSizeMaxLines,
			},
			Overcommenting: core.OvercommentingConfig{
				Enabled:           *commentEnabled,
				MaxCommentRatio:   *commentMaxRatio,
				CheckRedundant:    *commentCheckRedundant,
				CheckDocCoverage:  *commentCheckDoc,
			},
			OrphanedCode: core.OrphanedCodeConfig{
				Enabled:              *orphanedEnabled,
				CheckUnusedFunctions: *orphanedCheckUnusedFuncs,
				CheckUnusedVariables: *orphanedCheckUnusedVars,
				CheckUnreachableCode: *orphanedCheckUnreachable,
				CheckDeadImports:     *orphanedCheckDeadImports,
			},
		},
		Output: core.OutputConfig{
			Format: *outputFormat,
			Verbose: *verbose,
		},
		Language: core.LanguageConfig{
			Go: core.GoConfig{
				IgnoreTests: *goIgnoreTests,
			},
		},
	}

	// Create context
	ctx := context.Background()

	// Initialize language registry
	registry := languages.NewRegistry()
	
	// Register Go analyzer
	goAnalyzer := golang.NewAnalyzer(cfg)
	registry.Register(goAnalyzer)

	// Create file scanner
	goScanner := golang.NewFileScanner()

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
	fmt.Println("Output Options:")
	fmt.Println("  -format string       Output format (console, json) (default \"console\")")
	fmt.Println("  -output string       Output file (default: stdout)")
	fmt.Println("  -verbose             Verbose output")
	fmt.Println()
	fmt.Println("Function Size Rules:")
	fmt.Println("  -enable-func-size    Enable large function detection (default true)")
	fmt.Println("  -func-max-lines      Maximum number of lines for a function (default 50)")
	fmt.Println()
	fmt.Println("File Size Rules:")
	fmt.Println("  -enable-file-size    Enable large file detection (default true)")
	fmt.Println("  -file-max-lines      Maximum number of lines for a file (default 500)")
	fmt.Println()
	fmt.Println("Comment Rules:")
	fmt.Println("  -enable-comments     Enable overcommenting detection (default true)")
	fmt.Println("  -comment-max-ratio   Maximum comment-to-code ratio (default 0.3)")
	fmt.Println("  -check-redundant     Check for redundant comments (default true)")
	fmt.Println("  -check-docs          Check for missing documentation (default true)")
	fmt.Println()
	fmt.Println("Orphaned Code Rules:")
	fmt.Println("  -enable-orphaned    Enable orphaned code detection (default true)")
	fmt.Println("  -check-unused-funcs  Check for unused functions (default true)")
	fmt.Println("  -check-unused-vars   Check for unused variables (default true)")
	fmt.Println("  -check-unreachable   Check for unreachable code (default true)")
	fmt.Println("  -check-dead-imports  Check for dead imports (default true)")
	fmt.Println()
	fmt.Println("Go-specific Options:")
	fmt.Println("  -ignore-tests        Ignore test files during analysis (default false)")
	fmt.Println()
	fmt.Println("General Options:")
	fmt.Println("  -version             Show version information")
	fmt.Println("  -help                Show help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  agentlint ./myproject")
	fmt.Println("  agentlint -format json -output report.json ./myproject")
	fmt.Println("  agentlint -func-max-lines 30 -file-max-lines 200 ./myproject")
	fmt.Println("  agentlint -enable-comments=false -check-unused-funcs=false ./myproject")
}