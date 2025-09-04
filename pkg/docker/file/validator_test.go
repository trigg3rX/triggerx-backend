package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestNewCodeValidator_ValidInput_ReturnsValidator(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go", ".js"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	logger := logging.NewNoOpLogger()

	// Act
	validator := newCodeValidator(cfg, logger, &fs.OSFileSystem{})

	// Assert
	require.NotNil(t, validator)
	assert.Equal(t, cfg, validator.config)
	assert.Equal(t, logger, validator.logger)
}

func TestCodeValidator_ValidateFile_ValidFile_ReturnsSuccess(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go", ".js"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Create a temporary valid file
	tempDir := t.TempDir()
	validContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(validContent), 0644)
	require.NoError(t, err)

	// Act
	result, err := validator.validateFile(filePath)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
	assert.Greater(t, result.Complexity, 0.0)
}

func TestCodeValidator_ValidateFile_FileTooLarge_ReturnsError(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       100, // Very small limit
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Create a temporary file that exceeds size limit
	tempDir := t.TempDir()
	largeContent := strings.Repeat("a", 200) // 200 bytes > 100 limit
	filePath := filepath.Join(tempDir, "large.go")
	err := os.WriteFile(filePath, []byte(largeContent), 0644)
	require.NoError(t, err)

	// Act
	result, err := validator.validateFile(filePath)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "file size exceeds limit")
}

func TestCodeValidator_ValidateFile_InvalidExtension_ReturnsError(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go", ".js"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Crea	te a temporary file with invalid extension
	tempDir := t.TempDir()
	content := `package main`
	filePath := filepath.Join(tempDir, "test.py") // .py not allowed
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Act
	result, err := validator.validateFile(filePath)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "file extension not allowed")
}

func TestCodeValidator_ValidateFile_BlockedPattern_ReturnsError(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Create a temporary file with blocked pattern
	tempDir := t.TempDir()
	content := `package main

import "os/exec"

func main() {
	exec.Command("rm", "-rf", "/").Run() // This should be blocked
}`
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Act
	result, err := validator.validateFile(filePath)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "dangerous pattern found")
}

func TestCodeValidator_ValidateFile_SuspiciousPattern_ReturnsWarning(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Create a temporary file with suspicious pattern
	tempDir := t.TempDir()
	content := `package main

import "os"

func main() {
	file, _ := os.Open("test.txt") // This should trigger a warning
}`
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Act
	result, err := validator.validateFile(filePath)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsValid) // Should still be valid
	assert.Empty(t, result.Errors)
	assert.Len(t, result.Warnings, 1)
	assert.Contains(t, result.Warnings[0], "suspicious pattern")
}

func TestCodeValidator_ValidateFile_NonexistentFile_ReturnsError(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Act
	result, err := validator.validateFile("/nonexistent/file.go")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get file info")
}

func TestCodeValidator_ValidateFileSize_ValidSize_ReturnsSuccess(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize: 1024,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	tempDir := t.TempDir()
	content := "small content"
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	result := &types.ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Act
	err = validator.validateFileSize(filePath, result)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Errors)
}

func TestCodeValidator_ValidateFileSize_ExceedsLimit_ReturnsError(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize: 10, // Very small limit
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	tempDir := t.TempDir()
	content := "this content is too large for the limit"
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	result := &types.ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Act
	err = validator.validateFileSize(filePath, result)

	// Assert
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "file size exceeds limit")
}

func TestCodeValidator_ValidateFileExtension_ValidExtensions_ReturnsSuccess(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		AllowedExtensions: []string{".go", ".js", ".py"},
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	testCases := []struct {
		name     string
		filePath string
	}{
		{"Go file", "/path/to/file.go"},
		{"JavaScript file", "/path/to/file.js"},
		{"Python file", "/path/to/file.py"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := &types.ValidationResult{
				IsValid:  true,
				Errors:   make([]string, 0),
				Warnings: make([]string, 0),
			}

			// Act
			err := validator.validateFileExtension(tc.filePath, result)

			// Assert
			assert.NoError(t, err)
			assert.True(t, result.IsValid)
			assert.Empty(t, result.Errors)
		})
	}
}

func TestCodeValidator_ValidateFileExtension_InvalidExtension_ReturnsError(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		AllowedExtensions: []string{".go", ".js"},
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	testCases := []struct {
		name     string
		filePath string
	}{
		{"Python file", "/path/to/file.py"},
		{"No extension", "/path/to/file"},
		{"Unknown extension", "/path/to/file.xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := &types.ValidationResult{
				IsValid:  true,
				Errors:   make([]string, 0),
				Warnings: make([]string, 0),
			}

			// Act
			err := validator.validateFileExtension(tc.filePath, result)

			// Assert
			assert.NoError(t, err)
			assert.False(t, result.IsValid)
			assert.Len(t, result.Errors, 1)
			assert.Contains(t, result.Errors[0], "file extension not allowed")
		})
	}
}

func TestCodeValidator_CalculateComplexity_ValidFile_ReturnsComplexity(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize: 1024,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	tempDir := t.TempDir()
	content := `package main

import "fmt"

func main() {
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			fmt.Println(i)
		}
	}
}`
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Act
	complexity := validator.calculateComplexity(filePath)

	// Assert
	assert.Greater(t, complexity, 0.0)
}

func TestCodeValidator_CalculateComplexity_NonexistentFile_ReturnsZero(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize: 1024,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Act
	complexity := validator.calculateComplexity("/nonexistent/file.go")

	// Assert
	assert.Equal(t, 0.0, complexity)
}

func TestCodeValidator_CalculateContentComplexity_EmptyContent_ReturnsZero(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize: 1024,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	// Act
	complexity := validator.calculateContentComplexity([]byte{})

	// Assert
	assert.Equal(t, 0.0, complexity)
}

func TestCodeValidator_CalculateContentComplexity_ComplexContent_ReturnsHigherComplexity(t *testing.T) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize: 1024,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	simpleContent := `package main
func main() {
	fmt.Println("Hello")
}`

	complexContent := `package main

import (
	"fmt"
	"os"
	"strings"
)

func processData(data []string) []string {
	result := make([]string, 0)
	for _, item := range data {
		if strings.Contains(item, "test") {
			for i := 0; i < len(item); i++ {
				if item[i] == 'a' {
					result = append(result, item)
					break
				}
			}
		}
	}
	return result
}

func main() {
	data := []string{"test1", "test2", "other"}
	result := processData(data)
	for _, item := range result {
		fmt.Println(item)
	}
}`

	// Act
	simpleComplexity := validator.calculateContentComplexity([]byte(simpleContent))
	complexComplexity := validator.calculateContentComplexity([]byte(complexContent))

	// Assert
	assert.Greater(t, complexComplexity, simpleComplexity)
	assert.Greater(t, simpleComplexity, 0.0)
}

// Benchmark tests
func BenchmarkCodeValidator_ValidateFile_SimpleFile(b *testing.B) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	tempDir := b.TempDir()
	content := `package main

func main() {
	fmt.Println("Hello, World!")
}`
	filePath := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(b, err)

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.validateFile(filePath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodeValidator_CalculateContentComplexity_ComplexFile(b *testing.B) {
	// Arrange
	cfg := config.ValidationConfig{
		MaxFileSize:       1024,
		AllowedExtensions: []string{".go"},
		MaxComplexity:     10,
		TimeoutSeconds:    30,
	}
	validator := newCodeValidator(cfg, logging.NewNoOpLogger(), &fs.OSFileSystem{})

	content := `package main

import (
	"fmt"
	"os"
	"strings"
)

func processData(data []string) []string {
	result := make([]string, 0)
	for _, item := range data {
		if strings.Contains(item, "test") {
			for i := 0; i < len(item); i++ {
				if item[i] == 'a' {
					result = append(result, item)
					break
				}
			}
		}
	}
	return result
}

func main() {
	data := []string{"test1", "test2", "other"}
	result := processData(data)
	for _, item := range result {
		fmt.Println(item)
	}
}`

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.calculateContentComplexity([]byte(content))
	}
}
