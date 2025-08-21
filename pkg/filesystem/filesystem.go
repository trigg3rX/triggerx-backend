package fs

import (
	"os"
	"path/filepath"
)

// FileSystemAPI is the interface for the file system
type FileSystemAPI interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Remove(name string) error
	RemoveAll(path string) error
	Abs(path string) (string, error)
	ReadDir(dirname string) ([]os.DirEntry, error)
}

// OSFileSystem is the real implementation of the file system
type OSFileSystem struct{}

func (fs *OSFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (fs *OSFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (fs *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (fs *OSFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

func (fs *OSFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}
