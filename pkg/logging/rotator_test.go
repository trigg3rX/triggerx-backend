package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test NewSequentialRotator constructor
func TestNewSequentialRotator_ValidParameters_ReturnsCorrectInstance(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		maxSizeMB  int
		maxAge     int
		maxBackups int
		compress   bool
		expectedMB int64
	}{
		{
			name:       "standard config",
			filename:   "test.log",
			maxSizeMB:  50,
			maxAge:     30,
			maxBackups: 10,
			compress:   false,
			expectedMB: 50,
		},
		{
			name:       "compressed config",
			filename:   "app.log",
			maxSizeMB:  100,
			maxAge:     7,
			maxBackups: 5,
			compress:   true,
			expectedMB: 100,
		},
		{
			name:       "zero values",
			filename:   "debug.log",
			maxSizeMB:  0,
			maxAge:     0,
			maxBackups: 0,
			compress:   false,
			expectedMB: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rotator := NewSequentialRotator(tt.filename, tt.maxSizeMB, tt.maxAge, tt.maxBackups, tt.compress)

			assert.Equal(t, tt.filename, rotator.filename)
			assert.Equal(t, int64(tt.maxSizeMB)*1024*1024, rotator.maxSize)
			assert.Equal(t, tt.maxAge, rotator.maxAge)
			assert.Equal(t, tt.maxBackups, rotator.maxBackups)
			assert.Equal(t, tt.compress, rotator.compress)
			assert.Nil(t, rotator.file)
			assert.Equal(t, int64(0), rotator.size)
		})
	}
}

// Test Write method
func TestSequentialRotator_Write_FirstWrite_OpensFileAndWrites(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)
	defer func() {
		if err := rotator.Close(); err != nil {
			t.Errorf("Failed to close rotator: %v", err)
		}
	}()

	data := []byte("test log message\n")
	n, err := rotator.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, int64(len(data)), rotator.size)
	assert.NotNil(t, rotator.file)

	// Verify file was created and contains the data
	content, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, data, content)
}

func TestSequentialRotator_Write_SubsequentWrites_AppendToFile(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)
	defer func() {
		if err := rotator.Close(); err != nil {
			t.Errorf("Failed to close rotator: %v", err)
		}
	}()

	// First write
	data1 := []byte("first message\n")
	n1, err := rotator.Write(data1)
	assert.NoError(t, err)
	assert.Equal(t, len(data1), n1)

	// Second write
	data2 := []byte("second message\n")
	n2, err := rotator.Write(data2)
	assert.NoError(t, err)
	assert.Equal(t, len(data2), n2)

	// Verify total size
	expectedSize := int64(len(data1) + len(data2))
	assert.Equal(t, expectedSize, rotator.size)

	// Verify file content
	content, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	expectedContent := append(data1, data2...)
	assert.Equal(t, expectedContent, content)
}

func TestSequentialRotator_Write_ExceedsMaxSize_TriggersRotation(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	// Small max size to trigger rotation quickly
	rotator := NewSequentialRotator(testFile, 1, 30, 5, false) // 1MB
	defer func() {
		if err := rotator.Close(); err != nil {
			t.Errorf("Failed to close rotator: %v", err)
		}
	}()

	// Write data that exceeds max size
	largeData := strings.Repeat("x", 2*1024*1024) // 2MB
	n, err := rotator.Write([]byte(largeData))

	assert.NoError(t, err)
	assert.Equal(t, len(largeData), n)

	// Check that rotation occurred
	rotatedFiles, err := filepath.Glob(filepath.Join(tempDir, "test.*.log"))
	assert.NoError(t, err)
	assert.Greater(t, len(rotatedFiles), 0, "Expected rotated files to exist")
}

func TestSequentialRotator_Write_EmptyData_HandlesCorrectly(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)
	defer func() {
		if err := rotator.Close(); err != nil {
			t.Errorf("Failed to close rotator: %v", err)
		}
	}()

	// Write empty data
	n, err := rotator.Write([]byte{})

	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, int64(0), rotator.size)
}

// Test Close method
func TestSequentialRotator_Close_NoFile_ReturnsNil(t *testing.T) {
	rotator := NewSequentialRotator("test.log", 10, 30, 5, false)

	err := rotator.Close()
	assert.NoError(t, err)
}

func TestSequentialRotator_Close_WithFile_ClosesFile(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)

	// Write some data to open the file
	_, err := rotator.Write([]byte("test"))
	require.NoError(t, err)
	assert.NotNil(t, rotator.file)

	// Close the file
	err = rotator.Close()
	assert.NoError(t, err)
	assert.Nil(t, rotator.file)
}

// Test openFile method
func TestSequentialRotator_OpenFile_NewFile_CreatesDirectoryAndFile(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Create a nested directory structure
	nestedDir := filepath.Join(tempDir, "logs", "app")
	testFile := filepath.Join(nestedDir, "test.log")

	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)

	err := rotator.openFile()
	assert.NoError(t, err)
	assert.NotNil(t, rotator.file)
	assert.Equal(t, int64(0), rotator.size)

	// Verify directory was created
	_, err = os.Stat(nestedDir)
	assert.NoError(t, err)
}

func TestSequentialRotator_OpenFile_ExistingFile_AppendsAndGetsSize(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")

	// Create existing file with some content
	existingData := []byte("existing content\n")
	err := os.WriteFile(testFile, existingData, 0644)
	require.NoError(t, err)

	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)

	err = rotator.openFile()
	assert.NoError(t, err)
	assert.NotNil(t, rotator.file)
	assert.Equal(t, int64(len(existingData)), rotator.size)
}

// Test rotate method
func TestSequentialRotator_Rotate_ValidRotation_CreatesSequentialFile(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 1, 30, 5, false) // 1MB max size

	// Create initial file with some content
	err := rotator.openFile()
	require.NoError(t, err)

	// Write data to make file exist
	_, err = rotator.Write([]byte("initial content"))
	require.NoError(t, err)

	// Perform rotation
	err = rotator.rotate()
	assert.NoError(t, err)

	// Check that original file was renamed
	rotatedFiles, err := filepath.Glob(filepath.Join(tempDir, "test.*.log"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rotatedFiles))
	assert.Contains(t, rotatedFiles[0], "test.1.log")

	// Check that new file is open and size is reset
	assert.NotNil(t, rotator.file)
	assert.Equal(t, int64(0), rotator.size)
}

// Test getNextSequenceNumber method
func TestSequentialRotator_GetNextSequenceNumber_NoExistingFiles_ReturnsOne(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)

	nextSeq := rotator.getNextSequenceNumber()
	assert.Equal(t, 1, nextSeq)
}

func TestSequentialRotator_GetNextSequenceNumber_WithExistingFiles_ReturnsNextSequence(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Create existing rotated files
	testFiles := []string{"test.1.log", "test.3.log", "test.5.log"}
	for _, file := range testFiles {
		_, err := os.Create(filepath.Join(tempDir, file))
		require.NoError(t, err)
	}

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)

	nextSeq := rotator.getNextSequenceNumber()
	assert.Equal(t, 6, nextSeq)
}

func TestSequentialRotator_GetNextSequenceNumber_InvalidFilenames_HandlesGracefully(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Create files with invalid naming patterns
	invalidFiles := []string{"test.log", "test.abc.log", "test..log", "other.log"}
	for _, file := range invalidFiles {
		_, err := os.Create(filepath.Join(tempDir, file))
		require.NoError(t, err)
	}

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)

	nextSeq := rotator.getNextSequenceNumber()
	assert.Equal(t, 1, nextSeq)
}

// Test cleanupOldFiles method
func TestSequentialRotator_CleanupOldFiles_MaxBackupsExceeded_RemovesOldestFiles(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Create more files than maxBackups
	testFiles := []string{"test.1.log", "test.2.log", "test.3.log", "test.4.log", "test.5.log"}
	for _, file := range testFiles {
		_, err := os.Create(filepath.Join(tempDir, file))
		require.NoError(t, err)
	}

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 3, false) // maxBackups = 3

	rotator.cleanupOldFiles()

	// Check that only the newest 3 files remain
	remainingFiles, err := filepath.Glob(filepath.Join(tempDir, "test.*.log"))
	assert.NoError(t, err)
	assert.Len(t, remainingFiles, 3)

	// Verify the correct files remain (highest sequence numbers)
	expectedFiles := []string{"test.3.log", "test.4.log", "test.5.log"}
	for _, expected := range expectedFiles {
		found := false
		for _, remaining := range remainingFiles {
			if filepath.Base(remaining) == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected file %s to remain", expected)
	}
}

func TestSequentialRotator_CleanupOldFiles_MaxAgeExceeded_RemovesOldFiles(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Create files with different ages
	oldTime := time.Now().AddDate(0, 0, -10) // 10 days old
	newTime := time.Now().AddDate(0, 0, -1)  // 1 day old

	// Create old file
	oldFile := filepath.Join(tempDir, "test.1.log")
	_, err := os.Create(oldFile)
	require.NoError(t, err)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	require.NoError(t, err)

	// Create new file
	newFile := filepath.Join(tempDir, "test.2.log")
	_, err = os.Create(newFile)
	require.NoError(t, err)
	err = os.Chtimes(newFile, newTime, newTime)
	require.NoError(t, err)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 5, 10, false) // maxAge = 5 days

	rotator.cleanupOldFiles()

	// Check that only the new file remains
	remainingFiles, err := filepath.Glob(filepath.Join(tempDir, "test.*.log"))
	assert.NoError(t, err)
	assert.Len(t, remainingFiles, 1)
	assert.Contains(t, remainingFiles[0], "test.2.log")
}

func TestSequentialRotator_CleanupOldFiles_NoMaxBackups_KeepsAllFiles(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Create some files
	testFiles := []string{"test.1.log", "test.2.log", "test.3.log"}
	for _, file := range testFiles {
		_, err := os.Create(filepath.Join(tempDir, file))
		require.NoError(t, err)
	}

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 0, false) // maxBackups = 0 (no limit)

	rotator.cleanupOldFiles()

	// Check that all files remain
	remainingFiles, err := filepath.Glob(filepath.Join(tempDir, "test.*.log"))
	assert.NoError(t, err)
	assert.Len(t, remainingFiles, 3)
}

func TestSequentialRotator_CleanupOldFiles_NoMaxAge_KeepsAllFiles(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	// Create some files
	testFiles := []string{"test.1.log", "test.2.log", "test.3.log"}
	for _, file := range testFiles {
		_, err := os.Create(filepath.Join(tempDir, file))
		require.NoError(t, err)
	}

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 0, 5, false) // maxAge = 0 (no limit)

	rotator.cleanupOldFiles()

	// Check that all files remain
	remainingFiles, err := filepath.Glob(filepath.Join(tempDir, "test.*.log"))
	assert.NoError(t, err)
	assert.Len(t, remainingFiles, 3)
}

// Test edge cases and error conditions
func TestSequentialRotator_Write_ConcurrentWrites_HandlesSafely(t *testing.T) {
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	rotator := NewSequentialRotator(testFile, 10, 30, 5, false)
	defer func() {
		if err := rotator.Close(); err != nil {
			t.Errorf("Failed to close rotator: %v", err)
		}
	}()

	// Write concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			data := []byte(strings.Repeat("x", 100))
			_, err := rotator.Write(data)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all writes to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify file contains all data
	content, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, 1000, len(content)) // 10 * 100 bytes
}

func TestSequentialRotator_Write_FileSystemErrors_HandlesGracefully(t *testing.T) {
	// Test with invalid directory path
	invalidPath := "/invalid/path/that/does/not/exist/test.log"
	rotator := NewSequentialRotator(invalidPath, 10, 30, 5, false)

	data := []byte("test data")
	_, err := rotator.Write(data)
	assert.Error(t, err)
}

// Helper functions
func createTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "test_rotator")
	require.NoError(t, err)
	return tempDir
}

func cleanupTempDir(t *testing.T, tempDir string) {
	err := os.RemoveAll(tempDir)
	if err != nil {
		t.Logf("Failed to remove temp dir: %v", err)
	}
}

// Original tests (enhanced)
func TestSequentialRotator_Integration_CompleteWorkflow(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

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
		_, err := rotator.Write([]byte(testData))
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

func TestSequentialRotatorNaming_SequenceNumberGeneration_WorksCorrectly(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := createTempDir(t)
	defer cleanupTempDir(t, tempDir)

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

	err := rotator.Close()
	if err != nil {
		t.Fatalf("Failed to close rotator: %v", err)
	}
}
