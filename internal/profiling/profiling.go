package profiling

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"
)

var (
	mu          sync.Mutex
	profEnabled bool
	cpuProfile  *os.File
	memProfile  *os.File
	traceFile   *os.File
)

func StartCPUProfile(filename string) error {
	mu.Lock()
	defer mu.Unlock()

	if profEnabled {
		return fmt.Errorf("profiling already enabled")
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %v", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("could not start CPU profile: %v", err)
	}

	cpuProfile = f
	profEnabled = true
	return nil
}

func StopCPUProfile() {
	mu.Lock()
	defer mu.Unlock()

	if profEnabled && cpuProfile != nil {
		pprof.StopCPUProfile()
		cpuProfile.Close()
		cpuProfile = nil
	}
	profEnabled = false
}

func StartMemProfile(filename string) error {
	mu.Lock()
	defer mu.Unlock()

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %v", err)
	}

	memProfile = f
	return nil
}

func WriteMemProfile() error {
	mu.Lock()
	defer mu.Unlock()

	if memProfile != nil {
		runtime.GC()
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			return fmt.Errorf("could not write memory profile: %v", err)
		}
	}
	return nil
}

func CloseMemProfile() {
	mu.Lock()
	defer mu.Unlock()

	if memProfile != nil {
		memProfile.Close()
		memProfile = nil
	}
}

func StartTrace(filename string) error {
	mu.Lock()
	defer mu.Unlock()

	if profEnabled {
		return fmt.Errorf("profiling already enabled")
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create trace file: %v", err)
	}

	if err := trace.Start(f); err != nil {
		f.Close()
		return fmt.Errorf("could not start trace: %v", err)
	}

	traceFile = f
	profEnabled = true
	return nil
}

func StopTrace() {
	mu.Lock()
	defer mu.Unlock()

	if profEnabled && traceFile != nil {
		trace.Stop()
		traceFile.Close()
		traceFile = nil
	}
	profEnabled = false
}

type ProfileStats struct {
	NumGoroutine int
	NumCPU       int
	MemAlloc     uint64
	MemSys       uint64
	MemLookups   uint64
	MemMallocs   uint64
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
}

func GetStats() ProfileStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return ProfileStats{
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemAlloc:     memStats.Alloc,
		MemSys:       memStats.Sys,
		MemLookups:   memStats.Lookups,
		MemMallocs:   memStats.Mallocs,
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		HeapIdle:     memStats.HeapIdle,
		HeapInuse:    memStats.HeapInuse,
	}
}

func PrintStats(stats ProfileStats) {
	fmt.Printf("=== Profile Statistics ===\n")
	fmt.Printf("Goroutines: %d\n", stats.NumGoroutine)
	fmt.Printf("CPU Cores: %d\n", stats.NumCPU)
	fmt.Printf("Memory Alloc: %d bytes (%.2f KB)\n", stats.MemAlloc, float64(stats.MemAlloc)/1024)
	fmt.Printf("Memory Sys: %d bytes (%.2f KB)\n", stats.MemSys, float64(stats.MemSys)/1024)
	fmt.Printf("Heap Alloc: %d bytes (%.2f KB)\n", stats.HeapAlloc, float64(stats.HeapAlloc)/1024)
	fmt.Printf("Heap Sys: %d bytes (%.2f KB)\n", stats.HeapSys, float64(stats.HeapSys)/1024)
	fmt.Printf("Heap Inuse: %d bytes (%.2f KB)\n", stats.HeapInuse, float64(stats.HeapInuse)/1024)
	fmt.Printf("==========================\n")
}

type TimingStats struct {
	StartTime   time.Time
	EndTime     time.Time
	Elapsed     time.Duration
	FileCount   int
	ResultCount int
}

func NewTimingStats() *TimingStats {
	return &TimingStats{
		StartTime: time.Now(),
	}
}

func (t *TimingStats) Finish(fileCount, resultCount int) {
	t.EndTime = time.Now()
	t.Elapsed = t.EndTime.Sub(t.StartTime)
	t.FileCount = fileCount
	t.ResultCount = resultCount
}

func (t *TimingStats) Print() {
	fmt.Printf("=== Timing Statistics ===\n")
	fmt.Printf("Total Time: %v\n", t.Elapsed)
	fmt.Printf("Files Analyzed: %d\n", t.FileCount)
	fmt.Printf("Results Found: %d\n", t.ResultCount)
	fmt.Printf("Files/Second: %.2f\n", float64(t.FileCount)/t.Elapsed.Seconds())
	fmt.Printf("Results/Second: %.2f\n", float64(t.ResultCount)/t.Elapsed.Seconds())
	fmt.Printf("==========================\n")
}
