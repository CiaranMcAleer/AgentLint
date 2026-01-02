package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/golang"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/python"
	"github.com/CiaranMcAleer/AgentLint/internal/output"
	"github.com/CiaranMcAleer/AgentLint/internal/profiling"
)

func main() {
	flags := parseFlags()
	if flags.showHelp {
		showHelp()
		return
	}
	if flags.showVersion {
		printVersion()
		return
	}

	setupProfiling(flags)
	setupWorkers(flags)

	path := resolvePath()
	cfg := buildConfig(flags)
	cfg.Language.Go.IgnoreTests = flags.goIgnoreTests
	ctx := context.Background()

	registry := setupAnalyzer(cfg)
	scanner := languages.NewMultiScanner(registry)
	timing := profiling.NewTimingStats()

	filesByLanguage, err := scanFiles(ctx, path, scanner)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning files: %v\n", err)
		os.Exit(1)
	}

	allResults := analyzeFiles(ctx, filesByLanguage, registry, cfg)
	printResults(timing, allResults, flags, cfg)
}

func printVersion() {
	fmt.Println("AgentLint v0.1.0")
	fmt.Println("A linter for detecting LLM code bad smells")
}

func setupProfiling(flags *parsedFlags) {
	if flags.cpuProfile != "" {
		if err := profiling.StartCPUProfile(flags.cpuProfile); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer profiling.StopCPUProfile()
	}

	if flags.memProfile != "" {
		if err := profiling.StartMemProfile(flags.memProfile); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting memory profile: %v\n", err)
			os.Exit(1)
		}
		defer profiling.CloseMemProfile()
	}

	if flags.traceProfile != "" {
		if err := profiling.StartTrace(flags.traceProfile); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting trace: %v\n", err)
			os.Exit(1)
		}
		defer profiling.StopTrace()
	}
}

func setupWorkers(flags *parsedFlags) {
	if flags.workers > 0 {
		runtime.GOMAXPROCS(flags.workers)
	}
}

func resolvePath() string {
	path := "."
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Path does not exist: %s\n", absPath)
		os.Exit(1)
	}

	return absPath
}

func printResults(timing *profiling.TimingStats, allResults []core.Result, flags *parsedFlags, cfg core.Config) {
	timing.Finish(len(allResults), len(allResults))
	if flags.verbose {
		timing.Print()
		profiling.PrintStats(profiling.GetStats())
	}

	if flags.memProfile != "" {
		profiling.WriteMemProfile()
	}

	outputResults(cfg, allResults)

	if len(allResults) > 0 {
		os.Exit(1)
	}
}

type parsedFlags struct {
	outputFormat             string
	outputFile               string
	verbose                  bool
	funcSizeEnabled          bool
	funcSizeMaxLines         int
	fileSizeEnabled          bool
	fileSizeMaxLines         int
	commentEnabled           bool
	commentMaxRatio          float64
	commentCheckRedundant    bool
	commentCheckDoc          bool
	orphanedEnabled          bool
	orphanedCheckUnusedFuncs bool
	orphanedCheckUnusedVars  bool
	orphanedCheckUnreachable bool
	orphanedCheckDeadImports bool
	goIgnoreTests            bool
	showVersion              bool
	showHelp                 bool
	cpuProfile               string
	memProfile               string
	traceProfile             string
	workers                  int
}

func parseFlags() *parsedFlags {
	f := &parsedFlags{}

	flag.StringVar(&f.outputFormat, "format", "console", "Output format (console, json)")
	flag.StringVar(&f.outputFile, "output", "", "Output file (default: stdout)")
	flag.BoolVar(&f.verbose, "verbose", false, "Verbose output")

	flag.BoolVar(&f.funcSizeEnabled, "enable-func-size", true, "Enable large function detection")
	flag.IntVar(&f.funcSizeMaxLines, "func-max-lines", 50, "Maximum number of lines for a function")

	flag.BoolVar(&f.fileSizeEnabled, "enable-file-size", true, "Enable large file detection")
	flag.IntVar(&f.fileSizeMaxLines, "file-max-lines", 500, "Maximum number of lines for a file")

	flag.BoolVar(&f.commentEnabled, "enable-comments", true, "Enable overcommenting detection")
	flag.Float64Var(&f.commentMaxRatio, "comment-max-ratio", 0.3, "Maximum comment-to-code ratio")
	flag.BoolVar(&f.commentCheckRedundant, "check-redundant", true, "Check for redundant comments")
	flag.BoolVar(&f.commentCheckDoc, "check-docs", true, "Check for missing documentation")

	flag.BoolVar(&f.orphanedEnabled, "enable-orphaned", true, "Enable orphaned code detection")
	flag.BoolVar(&f.orphanedCheckUnusedFuncs, "check-unused-funcs", true, "Check for unused functions")
	flag.BoolVar(&f.orphanedCheckUnusedVars, "check-unused-vars", true, "Check for unused variables")
	flag.BoolVar(&f.orphanedCheckUnreachable, "check-unreachable", true, "Check for unreachable code")
	flag.BoolVar(&f.orphanedCheckDeadImports, "check-dead-imports", true, "Check for dead imports")

	flag.BoolVar(&f.goIgnoreTests, "ignore-tests", false, "Ignore test files during analysis")
	flag.StringVar(&f.cpuProfile, "cpuprofile", "", "Write CPU profile to file")
	flag.StringVar(&f.memProfile, "memprofile", "", "Write memory profile to file")
	flag.StringVar(&f.traceProfile, "trace", "", "Write execution trace to file")
	flag.IntVar(&f.workers, "workers", 0, "Number of worker threads (0 = auto)")
	flag.BoolVar(&f.showVersion, "version", false, "Show version information")
	flag.BoolVar(&f.showHelp, "help", false, "Show help information")

	flag.Parse()

	return f
}

func buildConfig(f *parsedFlags) core.Config {
	return core.Config{
		Rules: core.RulesConfig{
			FunctionSize: core.FunctionSizeConfig{
				Enabled:  f.funcSizeEnabled,
				MaxLines: f.funcSizeMaxLines,
			},
			FileSize: core.FileSizeConfig{
				Enabled:  f.fileSizeEnabled,
				MaxLines: f.fileSizeMaxLines,
			},
			Overcommenting: core.OvercommentingConfig{
				Enabled:          f.commentEnabled,
				MaxCommentRatio:  f.commentMaxRatio,
				CheckRedundant:   f.commentCheckRedundant,
				CheckDocCoverage: f.commentCheckDoc,
			},
			OrphanedCode: core.OrphanedCodeConfig{
				Enabled:              f.orphanedEnabled,
				CheckUnusedFunctions: f.orphanedCheckUnusedFuncs,
				CheckUnusedVariables: f.orphanedCheckUnusedVars,
				CheckUnreachableCode: f.orphanedCheckUnreachable,
				CheckDeadImports:     f.orphanedCheckDeadImports,
			},
		},
		Output: core.OutputConfig{
			Format:  f.outputFormat,
			Verbose: f.verbose,
		},
		Language: core.LanguageConfig{
			Go: core.GoConfig{
				IgnoreTests: f.goIgnoreTests,
			},
		},
	}
}

func setupAnalyzer(cfg core.Config) *languages.Registry {
	registry := languages.NewRegistry()

	// Register Go analyzer
	goAnalyzer := golang.NewAnalyzer(cfg)
	registry.Register(goAnalyzer)

	// Register Python analyzer
	pythonAnalyzer := python.NewAnalyzer(cfg)
	registry.Register(pythonAnalyzer)

	return registry
}

func scanFiles(ctx context.Context, absPath string, scanner *languages.MultiScanner) (map[string][]string, error) {
	fmt.Printf("Scanning %s...\n", absPath)
	return scanner.Scan(ctx, absPath)
}

func analyzeFiles(ctx context.Context, filesByLanguage map[string][]string, registry *languages.Registry, cfg core.Config) []core.Result {
	var allResults []core.Result

	for language, files := range filesByLanguage {
		analyzer, exists := registry.GetAnalyzer(language)
		if !exists {
			continue
		}

		fmt.Printf("Analyzing %d %s files...\n", len(files), language)

		if language == "go" && len(files) > 1 {
			parallelAnalyzer := golang.NewParallelAnalyzer(cfg, 0)
			results := parallelAnalyzer.AnalyzeFiles(ctx, files, cfg)
			allResults = append(allResults, results...)
		} else {
			for _, file := range files {
				results, err := analyzer.Analyze(ctx, file, cfg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error analyzing file %s: %v\n", file, err)
					continue
				}
				allResults = append(allResults, results...)
			}
		}
	}

	return allResults
}

func outputResults(cfg core.Config, allResults []core.Result) {
	var formatter output.Formatter
	switch cfg.Output.Format {
	case "json":
		formatter = output.NewJSONFormatter(cfg.Output.Verbose)
	case "console":
		fallthrough
	default:
		formatter = output.NewConsoleFormatter(cfg.Output.Verbose)
	}

	var outputFileHandle *os.File
	if cfg.Output.Format == "json" && cfg.Output.Format != "console" {
		var err error
		outputFileHandle, err = os.Create(cfg.Output.Format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer outputFileHandle.Close()
		os.Stdout = outputFileHandle
	}

	formatter.PrintHeader()
	if err := formatter.Format(allResults); err != nil {
		formatter.FormatError(err)
		os.Exit(1)
	}
	formatter.PrintFooter()
}

func showHelp() {
	fmt.Println("AgentLint - A linter for detecting LLM code bad smells")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agentlint [flags] [path]")
	fmt.Println()
	printOutputOptions()
	printFunctionSizeOptions()
	printFileSizeOptions()
	printCommentOptions()
	printOrphanedOptions()
	printGoOptions()
	printPerformanceOptions()
	printGeneralOptions()
	printExamples()
}

func printOutputOptions() {
	fmt.Println("Output Options:")
	fmt.Println("  -format string       Output format (console, json) (default \"console\")")
	fmt.Println("  -output string       Output file (default: stdout)")
	fmt.Println("  -verbose             Verbose output")
	fmt.Println()
}

func printFunctionSizeOptions() {
	fmt.Println("Function Size Rules:")
	fmt.Println("  -enable-func-size    Enable large function detection (default true)")
	fmt.Println("  -func-max-lines      Maximum number of lines for a function (default 50)")
	fmt.Println()
}

func printFileSizeOptions() {
	fmt.Println("File Size Rules:")
	fmt.Println("  -enable-file-size    Enable large file detection (default true)")
	fmt.Println("  -file-max-lines      Maximum number of lines for a file (default 500)")
	fmt.Println()
}

func printCommentOptions() {
	fmt.Println("Comment Rules:")
	fmt.Println("  -enable-comments     Enable overcommenting detection (default true)")
	fmt.Println("  -comment-max-ratio   Maximum comment-to-code ratio (default 0.3)")
	fmt.Println("  -check-redundant     Check for redundant comments (default true)")
	fmt.Println("  -check-docs          Check for missing documentation (default true)")
	fmt.Println()
}

func printOrphanedOptions() {
	fmt.Println("Orphaned Code Rules:")
	fmt.Println("  -enable-orphaned    Enable orphaned code detection (default true)")
	fmt.Println("  -check-unused-funcs  Check for unused functions (default true)")
	fmt.Println("  -check-unused-vars   Check for unused variables (default true)")
	fmt.Println("  -check-unreachable   Check for unreachable code (default true)")
	fmt.Println("  -check-dead-imports  Check for dead imports (default true)")
	fmt.Println()
}

func printGoOptions() {
	fmt.Println("Go-specific Options:")
	fmt.Println("  -ignore-tests        Ignore test files during analysis (default false)")
	fmt.Println()
}

func printPerformanceOptions() {
	fmt.Println("Performance Options:")
	fmt.Println("  -cpuprofile string   Write CPU profile to file")
	fmt.Println("  -memprofile string   Write memory profile to file")
	fmt.Println("  -trace string        Write execution trace to file")
	fmt.Println("  -workers int         Number of worker threads (0 = auto)")
	fmt.Println()
}

func printGeneralOptions() {
	fmt.Println("General Options:")
	fmt.Println("  -version             Show version information")
	fmt.Println("  -help                Show help information")
	fmt.Println()
}

func printExamples() {
	fmt.Println("Examples:")
	fmt.Println("  agentlint ./myproject")
	fmt.Println("  agentlint -format json -output report.json ./myproject")
	fmt.Println("  agentlint -func-max-lines 30 -file-max-lines 200 ./myproject")
	fmt.Println("  agentlint -enable-comments=false -check-unused-funcs=false ./myproject")
}
