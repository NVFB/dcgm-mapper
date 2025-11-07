package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteMappingFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test data
	processes := []GPUProcess{
		{GPU: "GPU-12345678-1234-1234-1234-123456789012", PID: "1001"},
		{GPU: "GPU-12345678-1234-1234-1234-123456789012", PID: "1002"},
		{GPU: "GPU-87654321-4321-4321-4321-210987654321", PID: "2001"},
	}

	// Write mapping files
	err := writeMappingFiles(tempDir, processes)
	if err != nil {
		t.Fatalf("writeMappingFiles() error = %v", err)
	}

	// Verify GPU-12345678-1234-1234-1234-123456789012 file
	gpu1File := filepath.Join(tempDir, "GPU-12345678-1234-1234-1234-123456789012")
	content1, err := os.ReadFile(gpu1File)
	if err != nil {
		t.Fatalf("Failed to read GPU file: %v", err)
	}
	lines1 := strings.Split(strings.TrimSpace(string(content1)), "\n")
	if len(lines1) != 2 {
		t.Errorf("Expected 2 PIDs for GPU1, got %d", len(lines1))
	}
	if lines1[0] != "1001" || lines1[1] != "1002" {
		t.Errorf("Expected PIDs [1001, 1002], got %v", lines1)
	}

	// Verify GPU-87654321-4321-4321-4321-210987654321 file
	gpu2File := filepath.Join(tempDir, "GPU-87654321-4321-4321-4321-210987654321")
	content2, err := os.ReadFile(gpu2File)
	if err != nil {
		t.Fatalf("Failed to read GPU file: %v", err)
	}
	lines2 := strings.Split(strings.TrimSpace(string(content2)), "\n")
	if len(lines2) != 1 {
		t.Errorf("Expected 1 PID for GPU2, got %d", len(lines2))
	}
	if lines2[0] != "2001" {
		t.Errorf("Expected PID 2001, got %s", lines2[0])
	}
}

func TestWriteMappingFilesEmpty(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test with empty processes
	processes := []GPUProcess{}

	// Write mapping files
	err := writeMappingFiles(tempDir, processes)
	if err != nil {
		t.Fatalf("writeMappingFiles() with empty processes error = %v", err)
	}

	// Verify directory exists but is empty
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected empty directory, got %d entries", len(entries))
	}
}

func TestWriteMappingFilesSingleGPU(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test data with single GPU, multiple PIDs
	processes := []GPUProcess{
		{GPU: "GPU-AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA", PID: "100"},
		{GPU: "GPU-AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA", PID: "200"},
		{GPU: "GPU-AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA", PID: "300"},
	}

	// Write mapping files
	err := writeMappingFiles(tempDir, processes)
	if err != nil {
		t.Fatalf("writeMappingFiles() error = %v", err)
	}

	// Verify only one file was created
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 file, got %d", len(entries))
	}

	// Verify file contents
	gpuFile := filepath.Join(tempDir, "GPU-AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA")
	content, err := os.ReadFile(gpuFile)
	if err != nil {
		t.Fatalf("Failed to read GPU file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 PIDs, got %d", len(lines))
	}
}

func TestWriteMappingFilesCreatesDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "subdir", "nested")

	// Test data
	processes := []GPUProcess{
		{GPU: "GPU-TEST", PID: "999"},
	}

	// Write mapping files to non-existent directory
	err := writeMappingFiles(nonExistentDir, processes)
	if err != nil {
		t.Fatalf("writeMappingFiles() should create directory, error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", nonExistentDir)
	}

	// Verify file exists
	gpuFile := filepath.Join(nonExistentDir, "GPU-TEST")
	if _, err := os.Stat(gpuFile); os.IsNotExist(err) {
		t.Errorf("GPU file was not created: %s", gpuFile)
	}
}

// TestGetGPUProcesses tests the getGPUProcesses function
// This test will be skipped if nvidia-smi is not available
func TestGetGPUProcesses(t *testing.T) {
	processes, err := getGPUProcesses()
	
	// If nvidia-smi is not available, skip the test
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") ||
			strings.Contains(err.Error(), "No devices were found") {
			t.Skip("nvidia-smi not available, skipping test")
		}
		// For other errors, we still want to know
		t.Logf("getGPUProcesses() returned error: %v (this may be expected if no GPUs are in use)", err)
		return
	}

	// If no error, processes should be a valid slice (can be empty)
	if processes == nil {
		t.Error("getGPUProcesses() returned nil processes without error")
	}

	// If there are processes, validate their structure
	for i, proc := range processes {
		if proc.GPU == "" {
			t.Errorf("Process %d has empty GPU field", i)
		}
		if proc.PID == "" {
			t.Errorf("Process %d has empty PID field", i)
		}
		// GPU UUID should start with "GPU-" in nvidia-smi output
		if !strings.HasPrefix(proc.GPU, "GPU-") {
			t.Errorf("Process %d has invalid GPU UUID format: %s", i, proc.GPU)
		}
	}
}

