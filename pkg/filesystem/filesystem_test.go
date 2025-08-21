package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOSFileSystem_Interface(t *testing.T) {
	// Test that OSFileSystem implements FileSystemAPI interface
	var _ FileSystemAPI = &OSFileSystem{}
}

func TestOSFileSystem_ReadWriteFile(t *testing.T) {
	fs := &OSFileSystem{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("Hello, World!")

	// Test WriteFile
	err = fs.WriteFile(testFile, testData, 0644)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Test ReadFile
	readData, err := fs.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("ReadFile returned incorrect data. Expected: %s, Got: %s", testData, readData)
	}
}

func TestOSFileSystem_ReadFile_NotExist(t *testing.T) {
	fs := &OSFileSystem{}

	_, err := fs.ReadFile("nonexistent_file.txt")
	if err == nil {
		t.Error("ReadFile should return error for non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("ReadFile should return os.ErrNotExist, got: %v", err)
	}
}

func TestOSFileSystem_MkdirAll(t *testing.T) {
	fs := &OSFileSystem{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testPath := filepath.Join(tempDir, "nested", "dir", "structure")

	// Test MkdirAll
	err = fs.MkdirAll(testPath, 0755)
	if err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Verify directory exists
	info, err := fs.Stat(testPath)
	if err != nil {
		t.Errorf("Stat failed after MkdirAll: %v", err)
	}
	if !info.IsDir() {
		t.Error("MkdirAll should create a directory")
	}
}

func TestOSFileSystem_Stat(t *testing.T) {
	fs := &OSFileSystem{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test Stat on directory
	info, err := fs.Stat(tempDir)
	if err != nil {
		t.Errorf("Stat failed on directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Stat should identify directory correctly")
	}

	// Test Stat on file
	testFile := filepath.Join(tempDir, "test.txt")
	err = fs.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	info, err = fs.Stat(testFile)
	if err != nil {
		t.Errorf("Stat failed on file: %v", err)
	}
	if info.IsDir() {
		t.Error("Stat should identify file correctly")
	}
	if info.Size() != 4 {
		t.Errorf("Stat should return correct file size. Expected: 4, Got: %d", info.Size())
	}
}

func TestOSFileSystem_Stat_NotExist(t *testing.T) {
	fs := &OSFileSystem{}

	_, err := fs.Stat("nonexistent_file.txt")
	if err == nil {
		t.Error("Stat should return error for non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Stat should return os.ErrNotExist, got: %v", err)
	}
}

func TestOSFileSystem_Remove(t *testing.T) {
	fs := &OSFileSystem{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	err = fs.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test Remove
	err = fs.Remove(testFile)
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// Verify file is removed
	_, err = fs.Stat(testFile)
	if err == nil {
		t.Error("File should be removed")
	}
	if !os.IsNotExist(err) {
		t.Errorf("File should not exist after Remove, got error: %v", err)
	}
}

func TestOSFileSystem_RemoveAll(t *testing.T) {
	fs := &OSFileSystem{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Fallback cleanup

	// Create nested structure
	nestedDir := filepath.Join(tempDir, "nested")
	err = fs.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	testFile := filepath.Join(nestedDir, "test.txt")
	err = fs.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test RemoveAll
	err = fs.RemoveAll(nestedDir)
	if err != nil {
		t.Errorf("RemoveAll failed: %v", err)
	}

	// Verify directory is removed
	_, err = fs.Stat(nestedDir)
	if err == nil {
		t.Error("Directory should be removed")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Directory should not exist after RemoveAll, got error: %v", err)
	}
}

func TestOSFileSystem_Abs(t *testing.T) {
	fs := &OSFileSystem{}

	// Test with relative path
	relPath := "test/path"
	absPath, err := fs.Abs(relPath)
	if err != nil {
		t.Errorf("Abs failed: %v", err)
	}
	if !filepath.IsAbs(absPath) {
		t.Errorf("Abs should return absolute path. Got: %s", absPath)
	}

	// Test with already absolute path
	if filepath.IsAbs(relPath) {
		t.Skip("Skipping absolute path test on this system")
	}

	// The returned path should contain the relative path
	if !filepath.IsAbs(absPath) {
		t.Errorf("Abs should return absolute path for relative input")
	}
}

func TestOSFileSystem_ReadDir(t *testing.T) {
	fs := &OSFileSystem{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some files and directories
	err = fs.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test1"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	err = fs.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("test2"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	err = fs.MkdirAll(filepath.Join(tempDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Test ReadDir
	entries, err := fs.ReadDir(tempDir)
	if err != nil {
		t.Errorf("ReadDir failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("ReadDir should return 3 entries. Got: %d", len(entries))
	}

	// Check that we have the expected entries
	entryNames := make(map[string]bool)
	for _, entry := range entries {
		entryNames[entry.Name()] = entry.IsDir()
	}

	if _, exists := entryNames["file1.txt"]; !exists || entryNames["file1.txt"] {
		t.Error("file1.txt should be present and not a directory")
	}
	if _, exists := entryNames["file2.txt"]; !exists || entryNames["file2.txt"] {
		t.Error("file2.txt should be present and not a directory")
	}
	if _, exists := entryNames["subdir"]; !exists || !entryNames["subdir"] {
		t.Error("subdir should be present and be a directory")
	}
}

func TestOSFileSystem_ReadDir_NotExist(t *testing.T) {
	fs := &OSFileSystem{}

	_, err := fs.ReadDir("nonexistent_directory")
	if err == nil {
		t.Error("ReadDir should return error for non-existent directory")
	}
}

// Integration test that combines multiple operations
func TestOSFileSystem_Integration(t *testing.T) {
	fs := &OSFileSystem{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested directory structure
	nestedPath := filepath.Join(tempDir, "level1", "level2")
	err = fs.MkdirAll(nestedPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Write a file in the nested directory
	testFile := filepath.Join(nestedPath, "integration_test.txt")
	testData := []byte("Integration test data")
	err = fs.WriteFile(testFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Get absolute path
	absPath, err := fs.Abs(testFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Read the file using absolute path
	readData, err := fs.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read file using absolute path: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("Data mismatch. Expected: %s, Got: %s", testData, readData)
	}

	// Check directory contents
	entries, err := fs.ReadDir(nestedPath)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry in directory, got %d", len(entries))
	}

	if entries[0].Name() != "integration_test.txt" {
		t.Errorf("Expected file name 'integration_test.txt', got '%s'", entries[0].Name())
	}

	// Clean up by removing the nested structure
	err = fs.RemoveAll(filepath.Join(tempDir, "level1"))
	if err != nil {
		t.Fatalf("Failed to remove nested structure: %v", err)
	}

	// Verify cleanup
	_, err = fs.Stat(nestedPath)
	if err == nil {
		t.Error("Nested directory should be removed")
	}
}
