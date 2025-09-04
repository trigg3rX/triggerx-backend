package file

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var cacheCfg = config.FileCacheConfig{
	CacheDir:          "data/cache",
	MaxCacheSize:      1024 * 1024, // 1MB
	EvictionSize:      100 * 1024,  // 100KB
	EnableCompression: true,
	MaxFileSize:       100 * 1024, // 100KB
}

func TestNewFileCache_ValidConfig_ReturnsCache(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg.CacheDir = tempDir
	logger := logging.NewNoOpLogger()

	// Act
	cache, err := newFileCache(cacheCfg, logger, fs.NewMockFileSystem())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cache)
	assert.Equal(t, tempDir, cache.cacheDir)
	assert.Equal(t, cacheCfg, cache.config)
	assert.Equal(t, logger, cache.logger)
	assert.NotNil(t, cache.fileCache)
	assert.NotNil(t, cache.stats)
	assert.Equal(t, cacheCfg.MaxCacheSize, cache.stats.MaxSize)
}

func TestNewFileCache_EmptyCacheDir_UsesDefaultDirectory(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = ""

	// Act
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cache)
	assert.Equal(t, "/var/lib/triggerx/cache", cache.cacheDir)
}

func TestNewFileCache_InvalidCacheDir_ReturnsError(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = "/nonexistent/path"

	// Act
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), &fs.FailingMockFS{})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, cache)
	assert.Contains(t, err.Error(), "failed to create cache directory")
}

func TestFileCache_GetOrDownloadFile_CacheHit_ReturnsCachedFile(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Pre-populate cache with a file
	key := "test-key"
	content := []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}")
	filePath, err := cache.storeFile(key, "go", content)
	require.NoError(t, err)

	// Mock download function (should not be called)
	downloadCalled := false
	downloadFunc := func() ([]byte, error) {
		downloadCalled = true
		return nil, nil
	}

	// Act
	resultPath, err := cache.getOrDownloadFile(key, "go", downloadFunc)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, filePath, resultPath)
	assert.False(t, downloadCalled, "Download function should not be called for cache hit")
}

func TestFileCache_GetOrDownloadFile_CacheMiss_DownloadsAndStores(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	key := "new-key"
	expectedContent := []byte("package main\n\nfunc main() {\n\tfmt.Println(\"World\")\n}")

	// Mock download function
	downloadCalled := false
	downloadFunc := func() ([]byte, error) {
		downloadCalled = true
		return expectedContent, nil
	}

	// Act
	resultPath, err := cache.getOrDownloadFile(key, "go", downloadFunc)

	// Assert
	assert.NoError(t, err)
	assert.True(t, downloadCalled, "Download function should be called for cache miss")
	assert.NotEmpty(t, resultPath)

	// Verify file was stored
	content, err := cache.fs.ReadFile(resultPath)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content)

	// Verify cache entry exists
	cache.mutex.RLock()
	cachedFile, exists := cache.fileCache[key]
	cache.mutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, key, cachedFile.Hash)
}

func TestFileCache_GetOrDownloadFile_DownloadFails_ReturnsError(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	key := "fail-key"

	// Mock download function that fails
	downloadFunc := func() ([]byte, error) {
		return nil, assert.AnError
	}

	// Act
	resultPath, err := cache.getOrDownloadFile(key, "go", downloadFunc)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, resultPath)
	assert.Contains(t, err.Error(), "download failed")
}

func TestFileCache_AccessCachedFile_FileExists_ReturnsPath(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Create a cached file
	key := "test-key"
	content := []byte("package main")
	filePath, err := cache.storeFile(key, "go", content)
	require.NoError(t, err)

	cache.mutex.RLock()
	cachedFile := cache.fileCache[key]
	cache.mutex.RUnlock()

	// Act
	resultPath, err := cache.accessCachedFile(cachedFile)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, filePath, resultPath)
}

func TestFileCache_AccessCachedFile_FileDeleted_ReturnsError(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Create a cached file
	key := "test-key"
	content := []byte("package main")
	filePath, err := cache.storeFile(key, "go", content)
	require.NoError(t, err)

	// Delete the file from disk
	err = cache.fs.Remove(filePath)
	require.NoError(t, err)

	cache.mutex.RLock()
	cachedFile := cache.fileCache[key]
	cache.mutex.RUnlock()

	// Act
	resultPath, err := cache.accessCachedFile(cachedFile)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, resultPath)
	assert.Contains(t, err.Error(), "cached file not found on disk")

	// Verify cache entry was removed
	cache.mutex.RLock()
	_, exists := cache.fileCache[key]
	cache.mutex.RUnlock()
	assert.False(t, exists)
}

func TestFileCache_StoreFile_ValidContent_StoresSuccessfully(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	key := "test-key"
	content := []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}")

	// Act
	filePath, err := cache.storeFile(key, "go", content)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, filePath)

	// Verify file exists on disk
	_, err = cache.fs.Stat(filePath)
	assert.NoError(t, err)

	// Verify content is correct
	storedContent, err := cache.fs.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, storedContent)

	// Verify cache entry
	cache.mutex.RLock()
	cachedFile, exists := cache.fileCache[key]
	cache.mutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, key, cachedFile.Hash)
	assert.Equal(t, int64(len(content)), cachedFile.Size)
}

func TestFileCache_StoreFile_KeyWithSpecialChars_SanitizesFilename(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	key := "https://example.com/path/to/file"
	content := []byte("package main")

	// Act
	filePath, err := cache.storeFile(key, "go", content)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, filePath)
	assert.Contains(t, filePath, "https___example.com_path_to_file.go")
}

func TestFileCache_EnsureSpace_EnoughSpace_NoEviction(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Add a small file
	key1 := "key1"
	content1 := []byte("small content")
	_, err = cache.storeFile(key1, "go", content1)
	require.NoError(t, err)

	// Act - try to ensure space for another small file
	err = cache.ensureSpace(100)

	// Assert
	assert.NoError(t, err)

	// Verify no files were evicted
	cache.mutex.RLock()
	fileCount := len(cache.fileCache)
	cache.mutex.RUnlock()
	assert.Equal(t, 1, fileCount)
}

func TestFileCache_EnsureSpace_NotEnoughSpace_EvictsOldestFiles(t *testing.T) {
	// Arrange
	// Create a local copy of the config to avoid affecting other tests
	localCacheCfg := cacheCfg
	localCacheCfg.CacheDir = t.TempDir()
	// Set a small cache size to force eviction
	localCacheCfg.MaxCacheSize = 50 // Very small cache size
	cache, err := newFileCache(localCacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Add two files
	key1 := "key1"
	content1 := []byte("first file content")
	_, err = cache.storeFile(key1, "go", content1)
	require.NoError(t, err)

	// Wait a bit to ensure different access times
	time.Sleep(10 * time.Millisecond)

	key2 := "key2"
	content2 := []byte("second file content")
	_, err = cache.storeFile(key2, "go", content2)
	require.NoError(t, err)

	// Verify both files exist
	cache.mutex.RLock()
	initialCount := len(cache.fileCache)
	cache.mutex.RUnlock()
	assert.Equal(t, 2, initialCount)

	// Act - try to ensure space for a large file
	err = cache.ensureSpace(200)

	// Assert
	assert.NoError(t, err)

	// Verify some files were evicted
	cache.mutex.RLock()
	finalCount := len(cache.fileCache)
	cache.mutex.RUnlock()
	assert.Less(t, finalCount, initialCount)
}

func TestFileCache_LoadExistingFiles_EmptyDirectory_LoadsNothing(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Act
	err = cache.loadExistingFiles()

	// Assert
	assert.NoError(t, err)
	cache.mutex.RLock()
	fileCount := len(cache.fileCache)
	cache.mutex.RUnlock()
	assert.Equal(t, 0, fileCount)
}

func TestFileCache_LoadExistingFiles_WithGoFiles_LoadsFiles(t *testing.T) {
	// Arrange
	tempDir := "/tmp/test"
	cacheCfg.CacheDir = tempDir
	mockFS := fs.NewMockFileSystem()

	// Pre-populate mock filesystem with files
	file1 := filepath.Join(tempDir, "test1.go")
	file2 := filepath.Join(tempDir, "test2.go")
	file3 := filepath.Join(tempDir, "test.txt")

	mockFS.AddDir(tempDir)
	mockFS.AddFile(file1, []byte("package main"))
	mockFS.AddFile(file2, []byte("package main"))
	mockFS.AddFile(file3, []byte("not a go file"))

	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), mockFS)
	require.NoError(t, err)

	// Act
	err = cache.loadExistingFiles()

	// Assert
	assert.NoError(t, err)
	cache.mutex.RLock()
	fileCount := len(cache.fileCache)
	cache.mutex.RUnlock()
	assert.Equal(t, 2, fileCount) // Only .go files should be loaded
}

func TestFileCache_LoadMetadata_ValidMetadata_LoadsSuccessfully(t *testing.T) {
	// Arrange
	tempDir := "/tmp/test"
	cacheCfg.CacheDir = tempDir
	mockFS := fs.NewMockFileSystem()

	// Pre-populate mock filesystem
	metadataPath := filepath.Join(tempDir, "cache_metadata.json")
	testFilePath := filepath.Join(tempDir, "test1.go")

	mockFS.AddDir(tempDir)
	mockFS.AddFile(testFilePath, []byte("package main"))

	// Write metadata file
	metadataContent := `{"key1":{"path":"` + testFilePath + `","hash":"key1","last_accessed":"` + time.Now().Format(time.RFC3339) + `","size":100,"created_at":"` + time.Now().Format(time.RFC3339) + `"}}`
	mockFS.AddFile(metadataPath, []byte(metadataContent))

	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), mockFS)
	require.NoError(t, err)

	// Act
	err = cache.loadMetadata(metadataPath)

	// Assert
	assert.NoError(t, err)
	cache.mutex.RLock()
	cachedFile, exists := cache.fileCache["key1"]
	cache.mutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, "key1", cachedFile.Hash)
	assert.Equal(t, int64(100), cachedFile.Size)
}

func TestFileCache_SaveMetadata_ValidCache_SavesSuccessfully(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg.CacheDir = tempDir
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Add a file to cache
	key := "test-key"
	content := []byte("package main")
	_, err = cache.storeFile(key, "go", content)
	require.NoError(t, err)

	// Act
	err = cache.saveMetadata()

	// Assert
	assert.NoError(t, err)

	// Verify metadata file exists
	metadataPath := filepath.Join(tempDir, "cache_metadata.json")
	_, err = cache.fs.Stat(metadataPath)
	assert.NoError(t, err)
}

func TestFileCache_UpdateHitRate_ValidStats_UpdatesCorrectly(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Set some stats
	cache.statsMutex.Lock()
	cache.stats.HitCount = 8
	cache.stats.MissCount = 2
	cache.statsMutex.Unlock()

	// Act
	cache.updateHitRate()

	// Assert
	cache.statsMutex.RLock()
	hitRate := cache.stats.HitRate
	cache.statsMutex.RUnlock()
	assert.Equal(t, 0.8, hitRate) // 8 hits / 10 total = 0.8
}

func TestFileCache_GetCacheStats_ValidCache_ReturnsStats(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Add some files to cache
	key1 := "key1"
	content1 := []byte("content1")
	_, err = cache.storeFile(key1, "go", content1)
	require.NoError(t, err)

	key2 := "key2"
	content2 := []byte("content2")
	_, err = cache.storeFile(key2, "go", content2)
	require.NoError(t, err)

	// Act
	stats := cache.getCacheStats()

	// Assert
	assert.NotNil(t, stats)
	assert.Equal(t, int(2), stats.ItemCount)
	assert.Equal(t, cacheCfg.MaxCacheSize, stats.MaxSize)
	assert.Greater(t, stats.Size, int64(0))
}

func TestFileCache_Close_ValidCache_ClosesSuccessfully(t *testing.T) {
	// Arrange
	cacheCfg.CacheDir = t.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Act
	err = cache.close()

	// Assert
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkFileCache_StoreFile_SmallFile(b *testing.B) {
	// Arrange
	cacheCfg.CacheDir = b.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(b, err)

	content := []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}")

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		_, err := cache.storeFile(key, "go", content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileCache_GetOrDownloadFile_CacheHit(b *testing.B) {
	// Arrange
	cacheCfg.CacheDir = b.TempDir()
	cache, err := newFileCache(cacheCfg, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(b, err)

	// Pre-populate cache
	key := "test-key"
	content := []byte("package main")
	_, err = cache.storeFile(key, "go", content)
	require.NoError(b, err)

	downloadFunc := func() ([]byte, error) {
		return nil, nil
	}

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cache.getOrDownloadFile(key, "go", downloadFunc)
		if err != nil {
			b.Fatal(err)
		}
	}
}
