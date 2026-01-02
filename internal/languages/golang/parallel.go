package golang

import (
	"context"
	"runtime"
	"sync"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
)

type ParallelAnalyzer struct {
	analyzer  *Analyzer
	workerNum int
}

func NewParallelAnalyzer(config core.Config, workers int) *ParallelAnalyzer {
	if workers <= 0 {
		workers = defaultWorkerCount()
	}
	return &ParallelAnalyzer{
		analyzer:  NewAnalyzer(config),
		workerNum: workers,
	}
}

func defaultWorkerCount() int {
	maxProcs := runtime.NumCPU()
	if maxProcs > 4 {
		return maxProcs - 1
	}
	return maxProcs
}

type analyzeJob struct {
	filePath string
	result   chan []core.Result
	error    chan error
}

type analyzeResult struct {
	results []core.Result
	err     error
}

func (a *ParallelAnalyzer) AnalyzeFiles(ctx context.Context, filePaths []string, config core.Config) []core.Result {
	if len(filePaths) == 0 {
		return nil
	}

	jobChan := make(chan analyzeJob, len(filePaths))
	resultChan := make(chan analyzeResult, len(filePaths))
	var wg sync.WaitGroup

	for i := 0; i < a.workerNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.worker(ctx, jobChan, resultChan, config)
		}()
	}

	go func() {
		for _, filePath := range filePaths {
			job := analyzeJob{
				filePath: filePath,
				result:   make(chan []core.Result, 1),
				error:    make(chan error, 1),
			}
			jobChan <- job
		}
		close(jobChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Pre-allocate with estimated capacity (avg 2 results per file)
	allResults := make([]core.Result, 0, len(filePaths)*2)
	for result := range resultChan {
		if result.err != nil {
			continue
		}
		allResults = append(allResults, result.results...)
	}

	return allResults
}

func (a *ParallelAnalyzer) worker(ctx context.Context, jobChan <-chan analyzeJob, resultChan chan<- analyzeResult, config core.Config) {
	for job := range jobChan {
		results, err := a.analyzer.Analyze(ctx, job.filePath, config)
		resultChan <- analyzeResult{
			results: results,
			err:     err,
		}
	}
}

func (a *ParallelAnalyzer) Analyze(ctx context.Context, filePath string, config core.Config) ([]core.Result, error) {
	return a.analyzer.Analyze(ctx, filePath, config)
}

func (a *ParallelAnalyzer) SupportedExtensions() []string {
	return a.analyzer.SupportedExtensions()
}

func (a *ParallelAnalyzer) Name() string {
	return a.analyzer.Name()
}

func (a *ParallelAnalyzer) WorkerCount() int {
	return a.workerNum
}
