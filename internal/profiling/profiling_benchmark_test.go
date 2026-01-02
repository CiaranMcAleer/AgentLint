package profiling_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CiaranMcAleer/AgentLint/internal/profiling"
)

func BenchmarkGetStats(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = profiling.GetStats()
	}
}

func BenchmarkNewTimingStats(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = profiling.NewTimingStats()
	}
}

func BenchmarkTimingStats_Finish(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ts := profiling.NewTimingStats()
		ts.Finish(100, 50)
	}
}

func BenchmarkTimingStats_Print(b *testing.B) {
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	ts := profiling.NewTimingStats()
	time.Sleep(time.Millisecond)
	ts.Finish(100, 50)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.Print()
	}
}

func BenchmarkPrintStats(b *testing.B) {
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	stats := profiling.GetStats()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiling.PrintStats(stats)
	}
}

func BenchmarkCPUProfile_StartStop(b *testing.B) {
	tmpDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cpuFile := filepath.Join(tmpDir, "cpu.prof")
		_ = profiling.StartCPUProfile(cpuFile)
		profiling.StopCPUProfile()
		os.Remove(cpuFile)
	}
}

func BenchmarkMemProfile_StartWrite(b *testing.B) {
	tmpDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memFile := filepath.Join(tmpDir, "mem.prof")
		_ = profiling.StartMemProfile(memFile)
		_ = profiling.WriteMemProfile()
		profiling.CloseMemProfile()
		os.Remove(memFile)
	}
}

func BenchmarkProfileStats_Access(b *testing.B) {
	stats := profiling.GetStats()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stats.NumGoroutine
		_ = stats.NumCPU
		_ = stats.MemAlloc
		_ = stats.MemSys
		_ = stats.HeapAlloc
		_ = stats.HeapSys
		_ = stats.HeapInuse
	}
}
