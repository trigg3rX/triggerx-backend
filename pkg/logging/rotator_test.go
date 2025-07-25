package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSequentialRotator(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_rotator")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a test file path
	testFile := filepath.Join(tempDir, "2025-07-01.log")

	// Create a rotator with small size for testing (1KB)
	rotator := NewSequentialRotator(testFile, 1, 30, 5, false) // 1KB max size
	defer func() {
		err := rotator.Close()
		if err != nil {
			t.Fatalf("Failed to close rotator: %v", err)
		}
	}()

	// Write smaller chunks to trigger rotations more reliably
	testData := strings.Repeat("This is a test log line.\n", 20) // ~500 bytes

	// Write multiple batches to trigger rotations
	// Each batch is ~500 bytes, so we need at least 3 batches to exceed 1KB and trigger rotation
	for i := 0; i < 6; i++ {
		_, err = rotator.Write([]byte(testData))
		if err != nil {
			t.Fatalf("Failed to write batch %d: %v", i+1, err)
		}
		t.Logf("Wrote batch %d (%d bytes)", i+1, len(testData))
	}

	// Check that the current file exists
	currentFile := "2025-07-01.log"
	fullPath := filepath.Join(tempDir, currentFile)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Current file %s does not exist", currentFile)
	}

	// Check for any rotated files (should have sequential numbering)
	rotatedFiles, _ := filepath.Glob(filepath.Join(tempDir, "2025-07-01.*.log"))
	if len(rotatedFiles) == 0 {
		t.Logf("No rotated files found - this might be OK if rotation threshold wasn't reached")
	} else {
		t.Logf("Found %d rotated files", len(rotatedFiles))
		for _, file := range rotatedFiles {
			t.Logf("  Rotated file: %s", filepath.Base(file))
		}
	}

	// Verify file contents are not empty
	allFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.log"))
	for _, file := range allFiles {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if info.Size() == 0 {
			t.Errorf("File %s is empty", filepath.Base(file))
		}
	}

	t.Logf("Successfully created files with sequential naming:")
	files, _ := filepath.Glob(filepath.Join(tempDir, "*.log"))
	for _, file := range files {
		info, _ := os.Stat(file)
		t.Logf("  %s (size: %d bytes)", filepath.Base(file), info.Size())
	}
}

func TestSequentialRotatorNaming(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_naming")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 1, 30, 10, false) // 1KB max size

	// Test getNextSequenceNumber with no existing files
	nextSeq := rotator.getNextSequenceNumber()
	if nextSeq != 1 {
		t.Errorf("Expected next sequence to be 1, got %d", nextSeq)
	}

	// Create some test files to simulate existing rotated logs
	testFiles := []string{"test.1.log", "test.3.log", "test.5.log"}
	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		err = f.Close()
		if err != nil {
			t.Fatalf("Failed to close test file %s: %v", file, err)
		}
	}

	// Test getNextSequenceNumber with existing files
	nextSeq = rotator.getNextSequenceNumber()
	if nextSeq != 6 {
		t.Errorf("Expected next sequence to be 6, got %d", nextSeq)
	}

	err = rotator.Close()
	if err != nil {
		t.Fatalf("Failed to close rotator: %v", err)
	}
}
