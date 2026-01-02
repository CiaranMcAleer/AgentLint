package golang

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestCrossFileAnalyzer_NoFalsePositivesForMethodCalls ensures method calls are tracked
func TestCrossFileAnalyzer_NoFalsePositivesForMethodCalls(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with a struct and methods
	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

type Server struct {
	port int
}

func (s *Server) Start() {
	s.initialize()
	s.listen()
}

func (s *Server) initialize() {
	s.port = 8080
}

func (s *Server) listen() {
	_ = s.port
}

func main() {
	s := &Server{}
	s.Start()
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	// Should find no unused functions - all methods are called
	if len(results) > 0 {
		for _, r := range results {
			t.Errorf("False positive: %s at line %d - %s", r.FilePath, r.Line, r.Message)
		}
	}
}

// TestCrossFileAnalyzer_NoFalsePositivesForDirectCalls ensures direct function calls are tracked
func TestCrossFileAnalyzer_NoFalsePositivesForDirectCalls(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

func main() {
	helperA()
}

func helperA() {
	helperB()
}

func helperB() {
	helperC()
}

func helperC() {
	_ = "done"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	if len(results) > 0 {
		for _, r := range results {
			t.Errorf("False positive: %s at line %d - %s", r.FilePath, r.Line, r.Message)
		}
	}
}

// TestCrossFileAnalyzer_NoFalsePositivesForCrossFileCalls ensures cross-file calls are tracked
func TestCrossFileAnalyzer_NoFalsePositivesForCrossFileCalls(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

func main() {
	processData()
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	utilsFile := filepath.Join(tmpDir, "utils.go")
	err = os.WriteFile(utilsFile, []byte(`package main

func processData() {
	validateInput()
	transformData()
}

func validateInput() {
	_ = "validating"
}

func transformData() {
	_ = "transforming"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write utils.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	if len(results) > 0 {
		for _, r := range results {
			t.Errorf("False positive: %s at line %d - %s", r.FilePath, r.Line, r.Message)
		}
	}
}

// TestCrossFileAnalyzer_NoFalsePositivesForFunctionReferences ensures callbacks are tracked
func TestCrossFileAnalyzer_NoFalsePositivesForFunctionReferences(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

func main() {
	handler := myHandler
	process(handler)
	
	// Also test passing directly as argument
	process(anotherHandler)
}

func myHandler() {
	_ = "handling"
}

func anotherHandler() {
	_ = "also handling"
}

func process(fn func()) {
	fn()
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	if len(results) > 0 {
		for _, r := range results {
			t.Errorf("False positive: %s at line %d - %s", r.FilePath, r.Line, r.Message)
		}
	}
}

// TestCrossFileAnalyzer_DetectsRealOrphans ensures truly unused functions are detected
func TestCrossFileAnalyzer_DetectsRealOrphans(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

func main() {
	usedFunction()
}

func usedFunction() {
	_ = "I am used"
}

func orphanedFunction() {
	_ = "Nobody calls me"
}

func anotherOrphan() {
	_ = "I am also unused"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	// Should find exactly 2 orphaned functions
	if len(results) != 2 {
		t.Errorf("Expected 2 orphaned functions, got %d", len(results))
		for _, r := range results {
			t.Logf("Found: %s", r.Message)
		}
	}

	// Verify the correct functions are detected
	orphanNames := make(map[string]bool)
	for _, r := range results {
		if r.RuleID == "cross-file-unused-function" {
			// Extract function name from message
			orphanNames[r.Message] = true
		}
	}

	expectedOrphans := []string{
		"Function 'orphanedFunction' is not called anywhere in the project",
		"Function 'anotherOrphan' is not called anywhere in the project",
	}

	for _, expected := range expectedOrphans {
		if !orphanNames[expected] {
			t.Errorf("Expected to find orphan: %s", expected)
		}
	}
}

// TestCrossFileAnalyzer_DetectsOrphanedMethods ensures unused methods are detected
func TestCrossFileAnalyzer_DetectsOrphanedMethods(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

type Handler struct{}

func (h *Handler) UsedMethod() {
	_ = "I am used"
}

func (h *Handler) orphanedMethod() {
	_ = "Nobody calls me"
}

func main() {
	h := &Handler{}
	h.UsedMethod()
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	// Should find exactly 1 orphaned method
	foundOrphanedMethod := false
	for _, r := range results {
		if r.RuleID == "cross-file-unused-method" {
			foundOrphanedMethod = true
			if r.Message != "Method 'orphanedMethod' on receiver 'Handler' is not called anywhere in the project" {
				t.Errorf("Unexpected message: %s", r.Message)
			}
		}
	}

	if !foundOrphanedMethod {
		t.Error("Expected to find orphaned method 'orphanedMethod'")
	}
}

// TestCrossFileAnalyzer_IgnoresExportedFunctions ensures exported functions are not flagged
func TestCrossFileAnalyzer_ExportedFunctionsNotFlaggedAsOrphans(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package mypackage

// PublicFunction is exported and may be called from outside the package
func PublicFunction() {
	privateHelper()
}

func privateHelper() {
	_ = "I help the public function"
}

// AnotherPublic is also exported
func AnotherPublic() {
	_ = "Also exported"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	// Exported functions should not be flagged, only unexported ones that aren't called
	for _, r := range results {
		// Check that no exported function is flagged
		if contains(r.Message, "PublicFunction") || contains(r.Message, "AnotherPublic") {
			t.Errorf("Exported function should not be flagged: %s", r.Message)
		}
	}
}

// TestCrossFileAnalyzer_IgnoresTestFunctions ensures test functions are not flagged
func TestCrossFileAnalyzer_IgnoresTestFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	// Note: Test files are skipped by AnalyzeDirectory, so we test with
	// functions that have test-like names in regular files
	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

func TestHelper() {
	_ = "test helper"
}

func BenchmarkHelper() {
	_ = "benchmark helper"
}

func ExampleUsage() {
	_ = "example"
}

func main() {
	_ = "main"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	// Test/Benchmark/Example functions should not be flagged
	for _, r := range results {
		if contains(r.Message, "TestHelper") ||
			contains(r.Message, "BenchmarkHelper") ||
			contains(r.Message, "ExampleUsage") {
			t.Errorf("Test/Benchmark/Example function should not be flagged: %s", r.Message)
		}
	}
}

// TestCrossFileAnalyzer_MethodCalledFromAnotherMethod ensures method-to-method calls are tracked
func TestCrossFileAnalyzer_MethodCalledFromAnotherMethod(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

type Service struct{}

func (s *Service) Run() {
	s.step1()
}

func (s *Service) step1() {
	s.step2()
}

func (s *Service) step2() {
	s.step3()
}

func (s *Service) step3() {
	_ = "final step"
}

func main() {
	s := &Service{}
	s.Run()
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	// All methods are chained and called, should find no orphans
	if len(results) > 0 {
		for _, r := range results {
			t.Errorf("False positive: %s at line %d - %s", r.FilePath, r.Line, r.Message)
		}
	}
}

// TestCrossFileAnalyzer_InitFunctionIgnored ensures init functions are not flagged
func TestCrossFileAnalyzer_InitFunctionIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

func init() {
	setupConfig()
}

func setupConfig() {
	_ = "setting up"
}

func main() {
	_ = "running"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	analyzer := NewCrossFileAnalyzer()
	if err := analyzer.AnalyzeDirectory(context.Background(), tmpDir); err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	results := analyzer.FindUnusedFunctions()

	// init() should not be flagged, and setupConfig is called from init
	for _, r := range results {
		if contains(r.Message, "init") {
			t.Errorf("init function should not be flagged: %s", r.Message)
		}
	}

	// setupConfig should also not be flagged since it's called from init
	if len(results) > 0 {
		for _, r := range results {
			t.Errorf("Unexpected orphan: %s", r.Message)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
