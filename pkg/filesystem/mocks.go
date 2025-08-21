package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MockFileSystem is a mock implementation of the file system
type MockFileSystem struct {
	files          map[string][]byte
	dirs           map[string]bool
	readFileResult func(string) ([]byte, error)
	absResult      func(string) (string, error)
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (fs *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	if fs.readFileResult != nil {
		return fs.readFileResult(filename)
	}
	if content, exists := fs.files[filename]; exists {
		return content, nil
	}
	return nil, os.ErrNotExist
}

func (fs *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	fs.files[filename] = data
	return nil
}

func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	fs.dirs[path] = true
	return nil
}

func (fs *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if content, exists := fs.files[name]; exists {
		return &mockFileInfo{name: name, isDir: false, size: int64(len(content))}, nil
	}
	if _, exists := fs.dirs[name]; exists {
		return &mockFileInfo{name: name, isDir: true, size: 0}, nil
	}
	return nil, os.ErrNotExist
}

func (fs *MockFileSystem) Remove(name string) error {
	delete(fs.files, name)
	delete(fs.dirs, name)
	return nil
}

func (fs *MockFileSystem) RemoveAll(path string) error {
	delete(fs.files, path)
	delete(fs.dirs, path)
	return nil
}

func (fs *MockFileSystem) Abs(path string) (string, error) {
	if fs.absResult != nil {
		return fs.absResult(path)
	}
	return path, nil // Simplified mock - just return the path as-is
}

func (fs *MockFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	if _, exists := fs.dirs[dirname]; !exists {
		return nil, os.ErrNotExist
	}

	var entries []os.DirEntry
	// Add files in this directory
	for filePath := range fs.files {
		if filepath.Dir(filePath) == dirname {
			entries = append(entries, &mockDirEntry{
				name:  filepath.Base(filePath),
				isDir: false,
			})
		}
	}
	// Add subdirectories
	for dirPath := range fs.dirs {
		if filepath.Dir(dirPath) == dirname && dirPath != dirname {
			entries = append(entries, &mockDirEntry{
				name:  filepath.Base(dirPath),
				isDir: true,
			})
		}
	}
	return entries, nil
}

// Helper methods for testing
func (fs *MockFileSystem) AddFile(filename string, content []byte) {
	fs.files[filename] = content
}

func (fs *MockFileSystem) AddDir(dirname string) {
	fs.dirs[dirname] = true
}

// SetReadFileResultFunc sets a custom function to handle ReadFile calls
func (fs *MockFileSystem) SetReadFileResultFunc(fn func(string) ([]byte, error)) {
	fs.readFileResult = fn
}

// SetAbsResultFunc sets a custom function to handle Abs calls
func (fs *MockFileSystem) SetAbsResultFunc(fn func(string) (string, error)) {
	fs.absResult = fn
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name  string
	isDir bool
	size  int64
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// mockDirEntry implements os.DirEntry for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string      { return m.name }
func (m *mockDirEntry) IsDir() bool       { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode { return 0644 }
func (m *mockDirEntry) Info() (os.FileInfo, error) {
	return &mockFileInfo{name: m.name, isDir: m.isDir, size: 0}, nil
}

// failingMockFS is a mock that fails MkdirAll operations
type FailingMockFS struct {
	MockFileSystem
}

func (fs *FailingMockFS) MkdirAll(path string, perm os.FileMode) error {
	return fmt.Errorf("permission denied")
}
