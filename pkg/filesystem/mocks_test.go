package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMockFileSystem(t *testing.T) {
	mockFS := NewMockFileSystem()

	if mockFS == nil {
		t.Fatal("NewMockFileSystem should return a non-nil instance")
	}

	if mockFS.files == nil {
		t.Error("NewMockFileSystem should initialize files map")
	}

	if mockFS.dirs == nil {
		t.Error("NewMockFileSystem should initialize dirs map")
	}
}

func TestMockFileSystem_Interface(t *testing.T) {
	// Test that MockFileSystem implements FileSystemAPI interface
	var _ FileSystemAPI = &MockFileSystem{}
}

func TestMockFileSystem_ReadWriteFile(t *testing.T) {
	mockFS := NewMockFileSystem()

	filename := "test.txt"
	testData := []byte("Hello, Mock World!")

	// Test WriteFile
	err := mockFS.WriteFile(filename, testData, 0644)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Test ReadFile
	readData, err := mockFS.ReadFile(filename)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("ReadFile returned incorrect data. Expected: %s, Got: %s", testData, readData)
	}
}

func TestMockFileSystem_ReadFile_NotExist(t *testing.T) {
	mockFS := NewMockFileSystem()

	_, err := mockFS.ReadFile("nonexistent.txt")
	if err == nil {
		t.Error("ReadFile should return error for non-existent file")
	}
	if err != os.ErrNotExist {
		t.Errorf("ReadFile should return os.ErrNotExist, got: %v", err)
	}
}

func TestMockFileSystem_MkdirAll(t *testing.T) {
	mockFS := NewMockFileSystem()

	dirPath := "/path/to/directory"

	err := mockFS.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Check that directory was created
	if !mockFS.dirs[dirPath] {
		t.Error("MkdirAll should create directory in dirs map")
	}
}

func TestMockFileSystem_Stat_File(t *testing.T) {
	mockFS := NewMockFileSystem()

	filename := "test.txt"
	testData := []byte("test content")

	// Write file first
	err := mockFS.WriteFile(filename, testData, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test Stat on file
	info, err := mockFS.Stat(filename)
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}

	if info == nil {
		t.Fatal("Stat should return non-nil FileInfo")
	}

	if info.Name() != filename {
		t.Errorf("Expected name %s, got %s", filename, info.Name())
	}

	if info.IsDir() {
		t.Error("File should not be identified as directory")
	}

	if info.Size() != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), info.Size())
	}
}

func TestMockFileSystem_Stat_Directory(t *testing.T) {
	mockFS := NewMockFileSystem()

	dirPath := "/test/dir"

	// Create directory first
	err := mockFS.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Test Stat on directory
	info, err := mockFS.Stat(dirPath)
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}

	if info == nil {
		t.Fatal("Stat should return non-nil FileInfo")
	}

	if info.Name() != dirPath {
		t.Errorf("Expected name %s, got %s", dirPath, info.Name())
	}

	if !info.IsDir() {
		t.Error("Directory should be identified as directory")
	}

	if info.Size() != 0 {
		t.Errorf("Expected directory size 0, got %d", info.Size())
	}
}

func TestMockFileSystem_Stat_NotExist(t *testing.T) {
	mockFS := NewMockFileSystem()

	_, err := mockFS.Stat("nonexistent")
	if err == nil {
		t.Error("Stat should return error for non-existent file")
	}
	if err != os.ErrNotExist {
		t.Errorf("Stat should return os.ErrNotExist, got: %v", err)
	}
}

func TestMockFileSystem_Remove(t *testing.T) {
	mockFS := NewMockFileSystem()

	// Test removing file
	filename := "test.txt"
	err := mockFS.WriteFile(filename, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = mockFS.Remove(filename)
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// Verify file is removed
	_, err = mockFS.ReadFile(filename)
	if err != os.ErrNotExist {
		t.Error("File should be removed from files map")
	}

	// Test removing directory
	dirPath := "/test/dir"
	err = mockFS.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	err = mockFS.Remove(dirPath)
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// Verify directory is removed
	_, err = mockFS.Stat(dirPath)
	if err != os.ErrNotExist {
		t.Error("Directory should be removed from dirs map")
	}
}

func TestMockFileSystem_RemoveAll(t *testing.T) {
	mockFS := NewMockFileSystem()

	// Create file and directory
	filename := "test.txt"
	dirPath := "/test/dir"

	err := mockFS.WriteFile(filename, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = mockFS.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Test RemoveAll on file
	err = mockFS.RemoveAll(filename)
	if err != nil {
		t.Errorf("RemoveAll failed: %v", err)
	}

	// Test RemoveAll on directory
	err = mockFS.RemoveAll(dirPath)
	if err != nil {
		t.Errorf("RemoveAll failed: %v", err)
	}

	// Verify both are removed
	_, err = mockFS.ReadFile(filename)
	if err != os.ErrNotExist {
		t.Error("File should be removed")
	}

	_, err = mockFS.Stat(dirPath)
	if err != os.ErrNotExist {
		t.Error("Directory should be removed")
	}
}

func TestMockFileSystem_Abs(t *testing.T) {
	mockFS := NewMockFileSystem()

	testPath := "/test/path"

	absPath, err := mockFS.Abs(testPath)
	if err != nil {
		t.Errorf("Abs failed: %v", err)
	}

	// Mock implementation just returns the path as-is
	if absPath != testPath {
		t.Errorf("Expected %s, got %s", testPath, absPath)
	}
}

func TestMockFileSystem_ReadDir(t *testing.T) {
	mockFS := NewMockFileSystem()

	// Create directory first
	dirPath := "/test"
	err := mockFS.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Create files in the directory
	file1 := filepath.Join(dirPath, "file1.txt")
	file2 := filepath.Join(dirPath, "file2.txt")

	err = mockFS.WriteFile(file1, []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = mockFS.WriteFile(file2, []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Create subdirectory
	subdir := filepath.Join(dirPath, "subdir")
	err = mockFS.MkdirAll(subdir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Test ReadDir
	entries, err := mockFS.ReadDir(dirPath)
	if err != nil {
		t.Errorf("ReadDir failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Check entries
	entryMap := make(map[string]bool)
	for _, entry := range entries {
		entryMap[entry.Name()] = entry.IsDir()
	}

	if isDir, exists := entryMap["file1.txt"]; !exists || isDir {
		t.Error("file1.txt should exist and not be a directory")
	}

	if isDir, exists := entryMap["file2.txt"]; !exists || isDir {
		t.Error("file2.txt should exist and not be a directory")
	}

	if isDir, exists := entryMap["subdir"]; !exists || !isDir {
		t.Error("subdir should exist and be a directory")
	}
}

func TestMockFileSystem_ReadDir_NotExist(t *testing.T) {
	mockFS := NewMockFileSystem()

	_, err := mockFS.ReadDir("/nonexistent")
	if err == nil {
		t.Error("ReadDir should return error for non-existent directory")
	}
	if err != os.ErrNotExist {
		t.Errorf("ReadDir should return os.ErrNotExist, got: %v", err)
	}
}

func TestMockFileSystem_AddFile(t *testing.T) {
	mockFS := NewMockFileSystem()

	filename := "helper_test.txt"
	content := []byte("helper content")

	mockFS.AddFile(filename, content)

	// Verify file was added
	readContent, err := mockFS.ReadFile(filename)
	if err != nil {
		t.Errorf("ReadFile failed after AddFile: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Expected content %s, got %s", content, readContent)
	}
}

func TestMockFileSystem_AddDir(t *testing.T) {
	mockFS := NewMockFileSystem()

	dirname := "/helper/dir"

	mockFS.AddDir(dirname)

	// Verify directory was added
	info, err := mockFS.Stat(dirname)
	if err != nil {
		t.Errorf("Stat failed after AddDir: %v", err)
	}

	if !info.IsDir() {
		t.Error("AddDir should create a directory")
	}
}

func TestMockFileInfo(t *testing.T) {
	info := &mockFileInfo{
		name:  "test.txt",
		isDir: false,
		size:  100,
	}

	if info.Name() != "test.txt" {
		t.Errorf("Expected name 'test.txt', got '%s'", info.Name())
	}

	if info.Size() != 100 {
		t.Errorf("Expected size 100, got %d", info.Size())
	}

	if info.IsDir() {
		t.Error("File should not be directory")
	}

	if info.Mode() != 0644 {
		t.Errorf("Expected mode 0644, got %v", info.Mode())
	}

	if info.Sys() != nil {
		t.Error("Sys() should return nil")
	}

	// Test that ModTime returns a time (we can't test exact value)
	modTime := info.ModTime()
	if modTime.IsZero() {
		t.Error("ModTime should not be zero")
	}
}

func TestMockFileInfo_Directory(t *testing.T) {
	info := &mockFileInfo{
		name:  "testdir",
		isDir: true,
		size:  0,
	}

	if !info.IsDir() {
		t.Error("Directory should be identified as directory")
	}

	if info.Size() != 0 {
		t.Errorf("Directory size should be 0, got %d", info.Size())
	}
}

func TestMockDirEntry(t *testing.T) {
	entry := &mockDirEntry{
		name:  "test.txt",
		isDir: false,
	}

	if entry.Name() != "test.txt" {
		t.Errorf("Expected name 'test.txt', got '%s'", entry.Name())
	}

	if entry.IsDir() {
		t.Error("File should not be directory")
	}

	if entry.Type() != 0644 {
		t.Errorf("Expected type 0644, got %v", entry.Type())
	}

	// Test Info method
	info, err := entry.Info()
	if err != nil {
		t.Errorf("Info() failed: %v", err)
	}

	if info.Name() != entry.Name() {
		t.Errorf("Info().Name() should match entry.Name()")
	}

	if info.IsDir() != entry.IsDir() {
		t.Errorf("Info().IsDir() should match entry.IsDir()")
	}
}

func TestMockDirEntry_Directory(t *testing.T) {
	entry := &mockDirEntry{
		name:  "testdir",
		isDir: true,
	}

	if !entry.IsDir() {
		t.Error("Directory should be identified as directory")
	}

	info, err := entry.Info()
	if err != nil {
		t.Errorf("Info() failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("Info() should return directory info")
	}
}

// Integration test for MockFileSystem
func TestMockFileSystem_Integration(t *testing.T) {
	mockFS := NewMockFileSystem()

	// Create directory structure
	baseDir := "/project"
	srcDir := filepath.Join(baseDir, "src")
	testDir := filepath.Join(baseDir, "test")

	err := mockFS.MkdirAll(baseDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed for base dir: %v", err)
	}

	err = mockFS.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	err = mockFS.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Create files
	mainFile := filepath.Join(srcDir, "main.go")
	testFile := filepath.Join(testDir, "main_test.go")

	err = mockFS.WriteFile(mainFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = mockFS.WriteFile(testFile, []byte("package main_test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// List base directory
	entries, err := mockFS.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in base directory, got %d", len(entries))
	}

	// Verify src directory contents
	srcEntries, err := mockFS.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("ReadDir failed for src: %v", err)
	}

	if len(srcEntries) != 1 {
		t.Errorf("Expected 1 entry in src directory, got %d", len(srcEntries))
	}

	if srcEntries[0].Name() != "main.go" {
		t.Errorf("Expected 'main.go', got '%s'", srcEntries[0].Name())
	}

	// Read and verify file contents
	content, err := mockFS.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != "package main" {
		t.Errorf("Expected 'package main', got '%s'", content)
	}

	// Test file stat
	info, err := mockFS.Stat(mainFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Size() != int64(len("package main")) {
		t.Errorf("Expected size %d, got %d", len("package main"), info.Size())
	}

	// Clean up
	err = mockFS.RemoveAll(baseDir)
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Verify cleanup
	_, err = mockFS.Stat(baseDir)
	if err != os.ErrNotExist {
		t.Error("Base directory should be removed")
	}
}

func TestMockFileSystem_RemoveAll_RecursiveDeletion(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a directory structure
	fs.AddDir("/my/dir")
	fs.AddDir("/my/dir/subdir")
	fs.AddFile("/my/dir/file.txt", []byte("content"))
	fs.AddFile("/my/dir/subdir/nested.txt", []byte("nested"))
	fs.AddFile("/my/dir/subdir/another.txt", []byte("another"))

	// RemoveAll should delete everything under /my/dir
	err := fs.RemoveAll("/my/dir")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Verify all files and directories under /my/dir are gone
	if _, err := fs.Stat("/my/dir"); err == nil {
		t.Error("Directory /my/dir should not exist after RemoveAll")
	}
	if _, err := fs.Stat("/my/dir/subdir"); err == nil {
		t.Error("Directory /my/dir/subdir should not exist after RemoveAll")
	}
	if _, err := fs.Stat("/my/dir/file.txt"); err == nil {
		t.Error("File /my/dir/file.txt should not exist after RemoveAll")
	}
	if _, err := fs.Stat("/my/dir/subdir/nested.txt"); err == nil {
		t.Error("File /my/dir/subdir/nested.txt should not exist after RemoveAll")
	}
	if _, err := fs.Stat("/my/dir/subdir/another.txt"); err == nil {
		t.Error("File /my/dir/subdir/another.txt should not exist after RemoveAll")
	}
}

func TestMockFileSystem_Remove_DirectoryNotEmpty(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a directory with content
	fs.AddDir("/my/dir")
	fs.AddFile("/my/dir/file.txt", []byte("content"))

	// Remove should fail for non-empty directory
	err := fs.Remove("/my/dir")
	if err == nil {
		t.Error("Remove should fail for non-empty directory")
	}

	// Verify directory still exists
	if _, err := fs.Stat("/my/dir"); err != nil {
		t.Error("Directory /my/dir should still exist after failed Remove")
	}
	if _, err := fs.Stat("/my/dir/file.txt"); err != nil {
		t.Error("File /my/dir/file.txt should still exist after failed Remove")
	}
}

func TestMockFileSystem_Remove_EmptyDirectory(t *testing.T) {
	fs := NewMockFileSystem()

	// Create an empty directory
	fs.AddDir("/my/empty/dir")

	// Remove should succeed for empty directory
	err := fs.Remove("/my/empty/dir")
	if err != nil {
		t.Fatalf("Remove failed for empty directory: %v", err)
	}

	// Verify directory is gone
	if _, err := fs.Stat("/my/empty/dir"); err == nil {
		t.Error("Directory /my/empty/dir should not exist after Remove")
	}
}

func TestMockFileSystem_Remove_File(t *testing.T) {
	fs := NewMockFileSystem()

	// Create a file
	fs.AddFile("/my/file.txt", []byte("content"))

	// Remove should succeed for file
	err := fs.Remove("/my/file.txt")
	if err != nil {
		t.Fatalf("Remove failed for file: %v", err)
	}

	// Verify file is gone
	if _, err := fs.Stat("/my/file.txt"); err == nil {
		t.Error("File /my/file.txt should not exist after Remove")
	}
}

func TestMockFileSystem_MkdirAll_RecursiveCreation(t *testing.T) {
	fs := NewMockFileSystem()

	// Create nested directory structure
	err := fs.MkdirAll("/a/b/c/d", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Verify all parent directories exist
	paths := []string{"/a", "/a/b", "/a/b/c", "/a/b/c/d"}
	for _, path := range paths {
		if info, err := fs.Stat(path); err != nil {
			t.Errorf("Directory %s should exist: %v", path, err)
		} else if !info.IsDir() {
			t.Errorf("%s should be a directory", path)
		}
	}
}

func TestMockFileSystem_MkdirAll_ExistingDirectories(t *testing.T) {
	fs := NewMockFileSystem()

	// Create some directories first
	fs.AddDir("/existing")
	fs.AddDir("/existing/sub")

	// MkdirAll should not fail when directories already exist
	err := fs.MkdirAll("/existing/sub/new", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed with existing directories: %v", err)
	}

	// Verify new directory was created
	if info, err := fs.Stat("/existing/sub/new"); err != nil {
		t.Errorf("Directory /existing/sub/new should exist: %v", err)
	} else if !info.IsDir() {
		t.Error("/existing/sub/new should be a directory")
	}
}

func TestMockFileSystem_WriteFile_CreatesParentDirectories(t *testing.T) {
	fs := NewMockFileSystem()

	// Write file to non-existent directory structure
	err := fs.WriteFile("/new/path/to/file.txt", []byte("content"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify parent directories were created
	paths := []string{"/new", "/new/path", "/new/path/to"}
	for _, path := range paths {
		if info, err := fs.Stat(path); err != nil {
			t.Errorf("Directory %s should exist: %v", path, err)
		} else if !info.IsDir() {
			t.Errorf("%s should be a directory", path)
		}
	}

	// Verify file was created
	if info, err := fs.Stat("/new/path/to/file.txt"); err != nil {
		t.Errorf("File /new/path/to/file.txt should exist: %v", err)
	} else if info.IsDir() {
		t.Error("/new/path/to/file.txt should be a file")
	}
}

func TestMockFileSystem_RemoveAll_RootDirectory(t *testing.T) {
	fs := NewMockFileSystem()

	// Create files in different locations
	fs.AddFile("/file1.txt", []byte("content1"))
	fs.AddFile("/dir1/file2.txt", []byte("content2"))
	fs.AddFile("/dir2/file3.txt", []byte("content3"))
	fs.AddDir("/dir1")
	fs.AddDir("/dir2")

	// RemoveAll on root should remove everything
	err := fs.RemoveAll("/")
	if err != nil {
		t.Fatalf("RemoveAll on root failed: %v", err)
	}

	// Verify all files and directories are gone
	paths := []string{"/file1.txt", "/dir1", "/dir1/file2.txt", "/dir2", "/dir2/file3.txt"}
	for _, path := range paths {
		if _, err := fs.Stat(path); err == nil {
			t.Errorf("Path %s should not exist after RemoveAll on root", path)
		}
	}
}

func TestMockFileSystem_RemoveAll_PartialPath(t *testing.T) {
	fs := NewMockFileSystem()

	// Create structure with multiple directories
	fs.AddDir("/shared")
	fs.AddDir("/shared/dir1")
	fs.AddDir("/shared/dir2")
	fs.AddFile("/shared/dir1/file1.txt", []byte("content1"))
	fs.AddFile("/shared/dir2/file2.txt", []byte("content2"))
	fs.AddFile("/other/file.txt", []byte("other"))

	// RemoveAll on /shared/dir1 should only remove that subtree
	err := fs.RemoveAll("/shared/dir1")
	if err != nil {
		t.Fatalf("RemoveAll on /shared/dir1 failed: %v", err)
	}

	// Verify /shared/dir1 and its contents are gone
	if _, err := fs.Stat("/shared/dir1"); err == nil {
		t.Error("/shared/dir1 should not exist after RemoveAll")
	}
	if _, err := fs.Stat("/shared/dir1/file1.txt"); err == nil {
		t.Error("/shared/dir1/file1.txt should not exist after RemoveAll")
	}

	// Verify other paths still exist
	remainingPaths := []string{"/shared", "/shared/dir2", "/shared/dir2/file2.txt", "/other/file.txt"}
	for _, path := range remainingPaths {
		if _, err := fs.Stat(path); err != nil {
			t.Errorf("Path %s should still exist after RemoveAll on /shared/dir1: %v", path, err)
		}
	}
}
