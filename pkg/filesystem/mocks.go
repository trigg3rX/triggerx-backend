package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MockFileSystem is a mock implementation of the file system
type MockFileSystem struct {
	mu             sync.RWMutex
	files          map[string][]byte
	dirs           map[string]bool
	readFileResult func(string) ([]byte, error)
	absResult      func(string) (string, error)

	// Configurable error functions for testing
	writeFileError func(string, []byte, os.FileMode) error
	mkdirAllError  func(string, os.FileMode) error
	statError      func(string) error
	removeError    func(string) error
	removeAllError func(string) error
	readDirError   func(string) error
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
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if content, exists := fs.files[filename]; exists {
		return content, nil
	}
	return nil, os.ErrNotExist
}

func (fs *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	// Check for configurable error first
	if fs.writeFileError != nil {
		if err := fs.writeFileError(filename, data, perm); err != nil {
			return err
		}
	}

	// Ensure parent directory exists
	dir := filepath.Dir(filename)
	if dir != "." && dir != "/" {
		if err := fs.MkdirAll(dir, perm); err != nil {
			return err
		}
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.files[filename] = data
	return nil
}

func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	// Normalize path
	path = filepath.Clean(path)
	if path == "." || path == "/" {
		return nil
	}

	// Create all parent directories recursively
	parts := strings.Split(path, string(filepath.Separator))
	currentPath := ""

	for i, part := range parts {
		if part == "" {
			if i == 0 {
				currentPath = "/"
			}
			continue
		}

		switch currentPath {
		case "":
			currentPath = part
		case "/":
			currentPath = "/" + part
		default:
			currentPath = filepath.Join(currentPath, part)
		}

		// Check for configurable error for this specific path
		if fs.mkdirAllError != nil {
			if err := fs.mkdirAllError(currentPath, perm); err != nil {
				return err
			}
		}

		fs.mu.Lock()
		fs.dirs[currentPath] = true
		fs.mu.Unlock()
	}

	return nil
}

func (fs *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	// Check for configurable error first
	if fs.statError != nil {
		if err := fs.statError(name); err != nil {
			return nil, err
		}
	}

	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if content, exists := fs.files[name]; exists {
		return &mockFileInfo{name: name, isDir: false, size: int64(len(content))}, nil
	}
	if _, exists := fs.dirs[name]; exists {
		return &mockFileInfo{name: name, isDir: true, size: 0}, nil
	}
	return nil, os.ErrNotExist
}

func (fs *MockFileSystem) Remove(name string) error {
	// Check for configurable error first
	if fs.removeError != nil {
		if err := fs.removeError(name); err != nil {
			return err
		}
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if it's a directory
	if _, isDir := fs.dirs[name]; isDir {
		// Check if directory is empty
		if fs.hasChildren(name) {
			return fmt.Errorf("remove %s: directory not empty", name)
		}
	}

	delete(fs.files, name)
	delete(fs.dirs, name)
	return nil
}

func (fs *MockFileSystem) RemoveAll(path string) error {
	// Check for configurable error first
	if fs.removeAllError != nil {
		if err := fs.removeAllError(path); err != nil {
			return err
		}
	}

	// Remove all files and directories under this path
	path = filepath.Clean(path)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Remove all files that are under this path
	for filePath := range fs.files {
		if fs.isPathUnder(filePath, path) {
			delete(fs.files, filePath)
		}
	}

	// Remove all directories that are under this path
	for dirPath := range fs.dirs {
		if fs.isPathUnder(dirPath, path) {
			delete(fs.dirs, dirPath)
		}
	}

	return nil
}

// isPathUnder checks if childPath is under parentPath
func (fs *MockFileSystem) isPathUnder(childPath, parentPath string) bool {
	childPath = filepath.Clean(childPath)
	parentPath = filepath.Clean(parentPath)

	if parentPath == "/" {
		return childPath != "/"
	}

	return strings.HasPrefix(childPath, parentPath+string(filepath.Separator)) || childPath == parentPath
}

// hasChildren checks if a directory has any children (files or subdirectories)
func (fs *MockFileSystem) hasChildren(dirPath string) bool {
	dirPath = filepath.Clean(dirPath)

	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Check for files in this directory
	for filePath := range fs.files {
		if filepath.Dir(filePath) == dirPath {
			return true
		}
	}

	// Check for subdirectories
	for path := range fs.dirs {
		if filepath.Dir(path) == dirPath && path != dirPath {
			return true
		}
	}

	return false
}

func (fs *MockFileSystem) Abs(path string) (string, error) {
	if fs.absResult != nil {
		return fs.absResult(path)
	}
	return path, nil // Simplified mock - just return the path as-is
}

func (fs *MockFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	// Check for configurable error first
	if fs.readDirError != nil {
		if err := fs.readDirError(dirname); err != nil {
			return nil, err
		}
	}

	fs.mu.RLock()
	defer fs.mu.RUnlock()

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
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.files[filename] = content
}

func (fs *MockFileSystem) AddDir(dirname string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.dirs[dirname] = true
}

// SetReadFileResultFunc sets a custom function to handle ReadFile calls
func (fs *MockFileSystem) SetReadFileResultFunc(fn func(string) ([]byte, error)) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.readFileResult = fn
}

// SetAbsResultFunc sets a custom function to handle Abs calls
func (fs *MockFileSystem) SetAbsResultFunc(fn func(string) (string, error)) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.absResult = fn
}

// SetWriteFileError sets a function that will be called before WriteFile operations
func (fs *MockFileSystem) SetWriteFileError(fn func(string, []byte, os.FileMode) error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.writeFileError = fn
}

// SetMkdirAllError sets a function that will be called before MkdirAll operations
func (fs *MockFileSystem) SetMkdirAllError(fn func(string, os.FileMode) error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.mkdirAllError = fn
}

// SetStatError sets a function that will be called before Stat operations
func (fs *MockFileSystem) SetStatError(fn func(string) error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.statError = fn
}

// SetRemoveError sets a function that will be called before Remove operations
func (fs *MockFileSystem) SetRemoveError(fn func(string) error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.removeError = fn
}

// SetRemoveAllError sets a function that will be called before RemoveAll operations
func (fs *MockFileSystem) SetRemoveAllError(fn func(string) error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.removeAllError = fn
}

// SetReadDirError sets a function that will be called before ReadDir operations
func (fs *MockFileSystem) SetReadDirError(fn func(string) error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.readDirError = fn
}

// ClearAllErrors clears all error functions
func (fs *MockFileSystem) ClearAllErrors() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.writeFileError = nil
	fs.mkdirAllError = nil
	fs.statError = nil
	fs.removeError = nil
	fs.removeAllError = nil
	fs.readDirError = nil
	fs.readFileResult = nil
	fs.absResult = nil
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
