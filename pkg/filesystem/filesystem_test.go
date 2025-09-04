package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
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
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

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
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

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
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

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
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

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
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

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
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

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
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

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

func TestMockFileSystem_ConfigurableWriteFileError(t *testing.T) {
	fs := NewMockFileSystem()

	// Set up WriteFile to fail for specific files
	fs.SetWriteFileError(func(filename string, data []byte, perm os.FileMode) error {
		if filename == "/forbidden.txt" {
			return fmt.Errorf("permission denied: %s", filename)
		}
		return nil // Allow other files
	})

	// This should succeed
	err := fs.WriteFile("/allowed.txt", []byte("content"), 0644)
	if err != nil {
		t.Errorf("WriteFile should succeed for allowed file: %v", err)
	}

	// This should fail
	err = fs.WriteFile("/forbidden.txt", []byte("content"), 0644)
	if err == nil {
		t.Error("WriteFile should fail for forbidden file")
	} else if err.Error() != "permission denied: /forbidden.txt" {
		t.Errorf("Expected 'permission denied' error, got: %v", err)
	}

	// Clear the error and try again
	fs.ClearAllErrors()
	err = fs.WriteFile("/forbidden.txt", []byte("content"), 0644)
	if err != nil {
		t.Errorf("WriteFile should succeed after clearing errors: %v", err)
	}
}

func TestMockFileSystem_ConfigurableMkdirAllError(t *testing.T) {
	fs := NewMockFileSystem()

	// Set up MkdirAll to fail for specific paths
	fs.SetMkdirAllError(func(path string, perm os.FileMode) error {
		if path == "/system" || path == "/system/important" {
			return fmt.Errorf("access denied: %s", path)
		}
		return nil
	})

	// This should succeed
	err := fs.MkdirAll("/user/data", 0755)
	if err != nil {
		t.Errorf("MkdirAll should succeed for user path: %v", err)
	}

	// This should fail
	err = fs.MkdirAll("/system/important", 0755)
	if err == nil {
		t.Error("MkdirAll should fail for system path")
	} else if err.Error() != "access denied: /system" {
		t.Errorf("Expected 'access denied: /system' error, got: %v", err)
	}

	// Test that parent directory creation also fails
	err = fs.MkdirAll("/system/important/subdir", 0755)
	if err == nil {
		t.Error("MkdirAll should fail when parent is forbidden")
	}
}

func TestMockFileSystem_ConfigurableStatError(t *testing.T) {
	fs := NewMockFileSystem()

	// Create some files first
	fs.AddFile("/visible.txt", []byte("content"))
	fs.AddDir("/visible_dir")

	// Set up Stat to fail for specific paths
	fs.SetStatError(func(name string) error {
		if name == "/hidden.txt" || name == "/hidden_dir" {
			return os.ErrNotExist
		}
		return nil
	})

	// These should succeed
	if _, err := fs.Stat("/visible.txt"); err != nil {
		t.Errorf("Stat should succeed for visible file: %v", err)
	}
	if _, err := fs.Stat("/visible_dir"); err != nil {
		t.Errorf("Stat should succeed for visible directory: %v", err)
	}

	// These should fail
	if _, err := fs.Stat("/hidden.txt"); err == nil {
		t.Error("Stat should fail for hidden file")
	}
	if _, err := fs.Stat("/hidden_dir"); err == nil {
		t.Error("Stat should fail for hidden directory")
	}
}

func TestMockFileSystem_ConfigurableRemoveError(t *testing.T) {
	fs := NewMockFileSystem()

	// Create some files and directories
	fs.AddFile("/deletable.txt", []byte("content"))
	fs.AddFile("/protected.txt", []byte("content"))
	fs.AddDir("/deletable_dir")
	fs.AddDir("/protected_dir")

	// Set up Remove to fail for protected items
	fs.SetRemoveError(func(name string) error {
		if name == "/protected.txt" || name == "/protected_dir" {
			return fmt.Errorf("cannot remove protected item: %s", name)
		}
		return nil
	})

	// These should succeed
	err := fs.Remove("/deletable.txt")
	if err != nil {
		t.Errorf("Remove should succeed for deletable file: %v", err)
	}
	err = fs.Remove("/deletable_dir")
	if err != nil {
		t.Errorf("Remove should succeed for deletable directory: %v", err)
	}

	// These should fail
	err = fs.Remove("/protected.txt")
	if err == nil {
		t.Error("Remove should fail for protected file")
	} else if err.Error() != "cannot remove protected item: /protected.txt" {
		t.Errorf("Expected protection error, got: %v", err)
	}

	err = fs.Remove("/protected_dir")
	if err == nil {
		t.Error("Remove should fail for protected directory")
	}
}

func TestMockFileSystem_ConfigurableRemoveAllError(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a directory structure
	fs.AddDir("/safe")
	fs.AddFile("/safe/file1.txt", []byte("content1"))
	fs.AddFile("/safe/file2.txt", []byte("content2"))

	fs.AddDir("/dangerous")
	fs.AddFile("/dangerous/file1.txt", []byte("content1"))
	fs.AddFile("/dangerous/file2.txt", []byte("content2"))

	// Set up RemoveAll to fail for dangerous paths
	fs.SetRemoveAllError(func(path string) error {
		if path == "/dangerous" {
			return fmt.Errorf("cannot remove dangerous directory: %s", path)
		}
		return nil
	})

	// This should succeed
	err := fs.RemoveAll("/safe")
	if err != nil {
		t.Errorf("RemoveAll should succeed for safe directory: %v", err)
	}

	// Verify safe directory is gone
	if _, err := fs.Stat("/safe"); err == nil {
		t.Error("Safe directory should be removed")
	}

	// This should fail
	err = fs.RemoveAll("/dangerous")
	if err == nil {
		t.Error("RemoveAll should fail for dangerous directory")
	} else if err.Error() != "cannot remove dangerous directory: /dangerous" {
		t.Errorf("Expected danger error, got: %v", err)
	}

	// Verify dangerous directory still exists
	if _, err := fs.Stat("/dangerous"); err != nil {
		t.Error("Dangerous directory should still exist")
	}
}

func TestMockFileSystem_ConfigurableReadDirError(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a directory with files
	fs.AddDir("/readable")
	fs.AddFile("/readable/file1.txt", []byte("content1"))
	fs.AddFile("/readable/file2.txt", []byte("content2"))

	fs.AddDir("/unreadable")
	fs.AddFile("/unreadable/file1.txt", []byte("content1"))

	// Set up ReadDir to fail for unreadable directories
	fs.SetReadDirError(func(dirname string) error {
		if dirname == "/unreadable" {
			return fmt.Errorf("permission denied: %s", dirname)
		}
		return nil
	})

	// This should succeed
	entries, err := fs.ReadDir("/readable")
	if err != nil {
		t.Errorf("ReadDir should succeed for readable directory: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// This should fail
	if err == nil {
		t.Error("ReadDir should fail for unreadable directory")
	} else if err.Error() != "permission denied: /unreadable" {
		t.Errorf("Expected permission error, got: %v", err)
	}
}

func TestMockFileSystem_MultipleErrorTypes(t *testing.T) {
	fs := NewMockFileSystem()

	// Set up multiple error types
	fs.SetWriteFileError(func(filename string, data []byte, perm os.FileMode) error {
		if filename == "/disk_full.txt" {
			return fmt.Errorf("disk full")
		}
		return nil
	})

	fs.SetMkdirAllError(func(path string, perm os.FileMode) error {
		if path == "/no_permission" {
			return fmt.Errorf("no permission")
		}
		return nil
	})

	fs.SetStatError(func(name string) error {
		if name == "/corrupted" {
			return fmt.Errorf("file corrupted")
		}
		return nil
	})

	// Test all error types
	err := fs.WriteFile("/disk_full.txt", []byte("content"), 0644)
	if err == nil || err.Error() != "disk full" {
		t.Errorf("Expected 'disk full' error, got: %v", err)
	}

	err = fs.MkdirAll("/no_permission", 0755)
	if err == nil || err.Error() != "no permission" {
		t.Errorf("Expected 'no permission' error, got: %v", err)
	}

	_, err = fs.Stat("/corrupted")
	if err == nil || err.Error() != "file corrupted" {
		t.Errorf("Expected 'file corrupted' error, got: %v", err)
	}

	// Test that normal operations still work
	err = fs.WriteFile("/normal.txt", []byte("content"), 0644)
	if err != nil {
		t.Errorf("Normal WriteFile should work: %v", err)
	}

	err = fs.MkdirAll("/normal_dir", 0755)
	if err != nil {
		t.Errorf("Normal MkdirAll should work: %v", err)
	}

	_, err = fs.Stat("/normal.txt")
	if err != nil {
		t.Errorf("Normal Stat should work: %v", err)
	}
}

func TestMockFileSystem_ClearAllErrors(t *testing.T) {
	fs := NewMockFileSystem()

	// Set up errors
	fs.SetWriteFileError(func(filename string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("write error")
	})
	fs.SetMkdirAllError(func(path string, perm os.FileMode) error {
		return fmt.Errorf("mkdir error")
	})
	fs.SetStatError(func(name string) error {
		return fmt.Errorf("stat error")
	})

	// Verify errors are active
	err := fs.WriteFile("/test.txt", []byte("content"), 0644)
	if err == nil || err.Error() != "write error" {
		t.Errorf("Expected 'write error', got: %v", err)
	}

	err = fs.MkdirAll("/test", 0755)
	if err == nil || err.Error() != "mkdir error" {
		t.Errorf("Expected 'mkdir error', got: %v", err)
	}

	_, err = fs.Stat("/test.txt")
	if err == nil || err.Error() != "stat error" {
		t.Errorf("Expected 'stat error', got: %v", err)
	}

	// Clear all errors
	fs.ClearAllErrors()

	// Verify operations now work normally
	err = fs.WriteFile("/test.txt", []byte("content"), 0644)
	if err != nil {
		t.Errorf("WriteFile should work after clearing errors: %v", err)
	}

	err = fs.MkdirAll("/test", 0755)
	if err != nil {
		t.Errorf("MkdirAll should work after clearing errors: %v", err)
	}

	_, err = fs.Stat("/test.txt")
	if err != nil {
		t.Errorf("Stat should work after clearing errors: %v", err)
	}
}

func TestMockFileSystem_ErrorFunctionWithLogic(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a counter to track calls
	callCount := 0

	// Set up an error function that fails every 3rd call
	fs.SetWriteFileError(func(filename string, data []byte, perm os.FileMode) error {
		callCount++
		if callCount%3 == 0 {
			return fmt.Errorf("intermittent failure on call %d", callCount)
		}
		return nil
	})

	// Test multiple calls
	for i := 1; i <= 6; i++ {
		err := fs.WriteFile(fmt.Sprintf("/file_%d.txt", i), []byte("content"), 0644)
		if i%3 == 0 {
			// Should fail on 3rd and 6th calls
			if err == nil {
				t.Errorf("Call %d should fail", i)
			} else if err.Error() != fmt.Sprintf("intermittent failure on call %d", i) {
				t.Errorf("Call %d: expected 'intermittent failure', got: %v", i, err)
			}
		} else {
			// Should succeed on other calls
			if err != nil {
				t.Errorf("Call %d should succeed: %v", i, err)
			}
		}
	}
}


func TestMockFileSystem_ThreadSafety(t *testing.T) {
	fs := NewMockFileSystem()

	// Test concurrent reads and writes
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Start multiple goroutines that read and write files
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				filename := fmt.Sprintf("/goroutine_%d/file_%d.txt", id, j)
				content := []byte(fmt.Sprintf("content from goroutine %d, file %d", id, j))

				// Write file
				if err := fs.WriteFile(filename, content, 0644); err != nil {
					t.Errorf("WriteFile failed in goroutine %d: %v", id, err)
					return
				}

				// Read file
				if readContent, err := fs.ReadFile(filename); err != nil {
					t.Errorf("ReadFile failed in goroutine %d: %v", id, err)
					return
				} else if string(readContent) != string(content) {
					t.Errorf("Content mismatch in goroutine %d: expected %s, got %s", id, string(content), string(readContent))
					return
				}

				// Check if file exists
				if info, err := fs.Stat(filename); err != nil {
					t.Errorf("Stat failed in goroutine %d: %v", id, err)
					return
				} else if info.IsDir() {
					t.Errorf("File %s should not be a directory in goroutine %d", filename, id)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMockFileSystem_ConcurrentDirectoryOperations(t *testing.T) {
	fs := NewMockFileSystem()

	var wg sync.WaitGroup
	numGoroutines := 5
	numOperations := 50

	// Start multiple goroutines that create and remove directories
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				dirPath := fmt.Sprintf("/goroutine_%d/dir_%d", id, j)

				// Create directory
				if err := fs.MkdirAll(dirPath, 0755); err != nil {
					t.Errorf("MkdirAll failed in goroutine %d: %v", id, err)
					return
				}

				// Check if directory exists
				if info, err := fs.Stat(dirPath); err != nil {
					t.Errorf("Stat failed in goroutine %d: %v", id, err)
					return
				} else if !info.IsDir() {
					t.Errorf("Path %s should be a directory in goroutine %d", dirPath, id)
					return
				}

				// Create a file in the directory
				filename := dirPath + "/test.txt"
				content := []byte("test content")
				if err := fs.WriteFile(filename, content, 0644); err != nil {
					t.Errorf("WriteFile failed in goroutine %d: %v", id, err)
					return
				}

				// Try to remove directory (should fail because it's not empty)
				if err := fs.Remove(dirPath); err == nil {
					t.Errorf("Remove should fail for non-empty directory in goroutine %d", id)
					return
				}

				// Remove the file first
				if err := fs.Remove(filename); err != nil {
					t.Errorf("Remove file failed in goroutine %d: %v", id, err)
					return
				}

				// Now remove the empty directory
				if err := fs.Remove(dirPath); err != nil {
					t.Errorf("Remove directory failed in goroutine %d: %v", id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMockFileSystem_ConcurrentRemoveAll(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a complex directory structure
	for i := 0; i < 10; i++ {
		dirPath := fmt.Sprintf("/test/dir_%d", i)
		if err := fs.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		for j := 0; j < 5; j++ {
			filename := fmt.Sprintf("%s/file_%d.txt", dirPath, j)
			content := []byte(fmt.Sprintf("content %d", j))
			if err := fs.WriteFile(filename, content, 0644); err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
		}
	}

	var wg sync.WaitGroup
	numGoroutines := 5

	// Start multiple goroutines that call RemoveAll on different paths
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine removes a different subdirectory
			pathToRemove := fmt.Sprintf("/test/dir_%d", id)
			if err := fs.RemoveAll(pathToRemove); err != nil {
				t.Errorf("RemoveAll failed in goroutine %d: %v", id, err)
				return
			}

			// Verify the directory and its contents are gone
			if _, err := fs.Stat(pathToRemove); err == nil {
				t.Errorf("Directory %s should not exist after RemoveAll in goroutine %d", pathToRemove, id)
			}
		}(i)
	}

	wg.Wait()

	// Verify that other directories still exist
	for i := numGoroutines; i < 10; i++ {
		path := fmt.Sprintf("/test/dir_%d", i)
		if _, err := fs.Stat(path); err != nil {
			t.Errorf("Directory %s should still exist: %v", path, err)
		}
	}
}

func TestMockFileSystem_ConcurrentReadDir(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a directory with multiple files
	dirPath := "/test"
	if err := fs.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create multiple files
	for i := 0; i < 20; i++ {
		filename := fmt.Sprintf("%s/file_%d.txt", dirPath, i)
		content := []byte(fmt.Sprintf("content %d", i))
		if err := fs.WriteFile(filename, content, 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 20

	// Start multiple goroutines that read the directory
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				entries, err := fs.ReadDir(dirPath)
				if err != nil {
					t.Errorf("ReadDir failed in goroutine %d: %v", id, err)
					return
				}

				// Should have 20 files
				if len(entries) != 20 {
					t.Errorf("Expected 20 entries, got %d in goroutine %d", len(entries), id)
					return
				}

				// All entries should be files, not directories
				for _, entry := range entries {
					if entry.IsDir() {
						t.Errorf("Entry %s should not be a directory in goroutine %d", entry.Name(), id)
						return
					}
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMockFileSystem_ConcurrentAddFileAndDir(t *testing.T) {
	fs := NewMockFileSystem()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 50

	// Start multiple goroutines that use AddFile and AddDir
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Add a file
				filename := fmt.Sprintf("/goroutine_%d/file_%d.txt", id, j)
				content := []byte(fmt.Sprintf("content %d", j))
				fs.AddFile(filename, content)

				// Add a directory
				dirPath := fmt.Sprintf("/goroutine_%d/dir_%d", id, j)
				fs.AddDir(dirPath)

				// Verify both exist
				if _, err := fs.Stat(filename); err != nil {
					t.Errorf("File %s should exist in goroutine %d: %v", filename, id, err)
					return
				}

				if _, err := fs.Stat(dirPath); err != nil {
					t.Errorf("Directory %s should exist in goroutine %d: %v", dirPath, id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}
