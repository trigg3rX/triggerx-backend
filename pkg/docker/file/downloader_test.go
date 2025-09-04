package file

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestNewDownloader_ValidConfig_ReturnsDownloader(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}
	logger := logging.NewNoOpLogger()

	// Act
	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logger, fs.NewMockFileSystem())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, downloader)
	assert.Equal(t, httpClient, downloader.client)
	assert.Equal(t, logger, downloader.logger)
	assert.NotNil(t, downloader.cache)
	assert.NotNil(t, downloader.validator)
}

func TestNewDownloader_InvalidCacheConfig_ReturnsError(t *testing.T) {
	// Arrange
	cacheCfg := config.FileCacheConfig{
		CacheDir:          "/nonexistent/path",
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}
	logger := logging.NewNoOpLogger()
	mockFS := &fs.FailingMockFS{}

	// Act
	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logger, mockFS)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, downloader)
	assert.Contains(t, err.Error(), "failed to create file cache")
}

func TestDownloader_DownloadFile_ValidURL_DownloadsSuccessfully(t *testing.T) {
	// Arrange
	tempDir := "/tmp/test"
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}
	mockFS := fs.NewMockFileSystem()

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), mockFS)
	require.NoError(t, err)

	// Mock HTTP response
	expectedContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(expectedContent)),
	}
	httpClient.On("Get", "https://example.com/test.go").Return(mockResponse, nil)

	key := "test-key"
	url := "https://example.com/test.go"

	// Act
	result, err := downloader.downloadFile(key, url, "go")

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, key, result.Hash)
	assert.Equal(t, expectedContent, string(result.Content))
	assert.True(t, result.IsCached)
	assert.NotEmpty(t, result.FilePath)
	assert.Greater(t, result.Size, int64(0))
	assert.NotNil(t, result.Validation)
	assert.True(t, result.Validation.IsValid)

	// Verify file was stored in cache through mock filesystem
	_, err = mockFS.Stat(result.FilePath)
	assert.NoError(t, err)

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadFile_HTTPError_ReturnsError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Mock HTTP error
	httpClient.On("Get", "https://example.com/error.go").Return(nil, assert.AnError)

	key := "error-key"
	url := "https://example.com/error.go"

	// Act
	result, err := downloader.downloadFile(key, url, "go")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to download or store file in cache")

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadFile_HTTPStatusError_ReturnsError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Mock HTTP 404 response
	mockResponse := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("Not Found")),
	}
	httpClient.On("Get", "https://example.com/notfound.go").Return(mockResponse, nil)

	key := "notfound-key"
	url := "https://example.com/notfound.go"

	// Act
	result, err := downloader.downloadFile(key, url, "go")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unexpected status code: 404")

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadFile_ValidationFails_ReturnsResultWithErrors(t *testing.T) {
	// Arrange
	tempDir := "/tmp/test"
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       10, // Very small limit
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Mock HTTP response with large content
	largeContent := strings.Repeat("package main\n", 100) // Large content
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(largeContent)),
	}
	httpClient.On("Get", "https://example.com/large.go").Return(mockResponse, nil)

	key := "large-key"
	url := "https://example.com/large.go"

	// Act
	result, err := downloader.downloadFile(key, url, "go")

	// Assert
	assert.NoError(t, err) // Download succeeds
	require.NotNil(t, result)
	assert.Equal(t, largeContent, string(result.Content))
	assert.False(t, result.Validation.IsValid)
	assert.Len(t, result.Validation.Errors, 1)
	assert.Contains(t, result.Validation.Errors[0], "file size exceeds limit")

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadFile_BlockedPattern_ReturnsResultWithErrors(t *testing.T) {
	// Arrange
	tempDir := "/tmp/test"
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Mock HTTP response with blocked pattern
	dangerousContent := `package main

import "os/exec"

func main() {
	exec.Command("rm", "-rf", "/").Run() // This should be blocked
}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(dangerousContent)),
	}
	httpClient.On("Get", "https://example.com/dangerous.go").Return(mockResponse, nil)

	key := "dangerous-key"
	url := "https://example.com/dangerous.go"

	// Act
	result, err := downloader.downloadFile(key, url, "go")

	// Assert
	assert.NoError(t, err) // Download succeeds
	require.NotNil(t, result)
	assert.Equal(t, dangerousContent, string(result.Content))
	assert.False(t, result.Validation.IsValid)
	assert.Len(t, result.Validation.Errors, 1)
	assert.Contains(t, result.Validation.Errors[0], "dangerous pattern found")

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadFile_CacheHit_ReturnsCachedFile(t *testing.T) {
	// Arrange
	tempDir := "/tmp/test"
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Pre-populate cache
	key := "cached-key"
	url := "https://example.com/cached.go"
	content := `package main

func main() {
	fmt.Println("Cached content")
}`

	// First download to populate cache
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(content)),
	}
	httpClient.On("Get", url).Return(mockResponse, nil).Once()

	// First call to populate cache
	result1, err := downloader.downloadFile(key, url, "go")
	require.NoError(t, err)
	require.NotNil(t, result1)

	// Second call should hit cache (no HTTP call)
	result2, err := downloader.downloadFile(key, url, "go")

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, key, result2.Hash)
	assert.Equal(t, content, string(result2.Content))
	assert.True(t, result2.IsCached)
	assert.Equal(t, result1.FilePath, result2.FilePath)

	// Verify HTTP client was only called once
	httpClient.AssertNumberOfCalls(t, "Get", 1)
}

func TestDownloader_DownloadContent_ValidResponse_ReturnsContent(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	expectedContent := "Hello, World!"
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(expectedContent)),
	}
	httpClient.On("Get", "https://example.com/test.txt").Return(mockResponse, nil)

	// Act
	content, err := downloader.downloadContent("https://example.com/test.txt")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(content))

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadContent_HTTPError_ReturnsError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	httpClient.On("Get", "https://example.com/error.txt").Return(nil, assert.AnError)

	// Act
	content, err := downloader.downloadContent("https://example.com/error.txt")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, content)
	assert.Contains(t, err.Error(), "failed to download file")

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadContent_NonOKStatus_ReturnsError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	mockResponse := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("Server Error")),
	}
	httpClient.On("Get", "https://example.com/server-error.txt").Return(mockResponse, nil)

	// Act
	content, err := downloader.downloadContent("https://example.com/server-error.txt")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, content)
	assert.Contains(t, err.Error(), "unexpected status code: 500")

	httpClient.AssertExpectations(t)
}

func TestDownloader_DownloadContent_ReadBodyError_ReturnsError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Create a response with a body that will fail to read
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&failingReader{}),
	}
	httpClient.On("Get", "https://example.com/failing.txt").Return(mockResponse, nil)

	// Act
	content, err := downloader.downloadContent("https://example.com/failing.txt")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, content)
	assert.Contains(t, err.Error(), "failed to read response body")

	httpClient.AssertExpectations(t)
}

func TestDownloader_Close_ValidDownloader_ClosesSuccessfully(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}

	downloader, err := newDownloader(cacheCfg, validationCfg, &httppkg.MockHTTPClient{}, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(t, err)

	// Act
	err = downloader.close()

	// Assert
	assert.NoError(t, err)
}

func TestDownloader_Close_NilCache_ClosesSuccessfully(t *testing.T) {
	// Arrange
	downloader := &downloader{
		client:    &httppkg.MockHTTPClient{},
		cache:     nil,
		validator: &codeValidator{},
		logger:    logging.NewNoOpLogger(),
	}

	// Act
	err := downloader.close()

	// Assert
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkDownloader_DownloadFile_ValidURL(b *testing.B) {
	// Arrange
	tempDir := b.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(b, err)

	content := `package main

func main() {
	fmt.Println("Hello, World!")
}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(content)),
	}
	httpClient.On("Get", "https://example.com/benchmark.go").Return(mockResponse, nil)

	key := "benchmark-key"
	url := "https://example.com/benchmark.go"

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := downloader.downloadFile(key, url, "go")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDownloader_DownloadContent_SmallFile(b *testing.B) {
	// Arrange
	tempDir := b.TempDir()
	cacheCfg := config.FileCacheConfig{
		CacheDir:          tempDir,
		MaxCacheSize:      1024 * 1024,
		EvictionSize:      100 * 1024,
		EnableCompression: false,
		MaxFileSize:       100 * 1024,
	}
	validationCfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	httpClient := &httppkg.MockHTTPClient{}

	downloader, err := newDownloader(cacheCfg, validationCfg, httpClient, logging.NewNoOpLogger(), fs.NewMockFileSystem())
	require.NoError(b, err)

	content := "Hello, World!"
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(content)),
	}
	httpClient.On("Get", "https://example.com/small.txt").Return(mockResponse, nil)

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := downloader.downloadContent("https://example.com/small.txt")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper types for testing
type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}
