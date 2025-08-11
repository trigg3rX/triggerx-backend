package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type codeValidator struct {
	config config.ValidationConfig
	logger logging.Logger
}

func newCodeValidator(cfg config.ValidationConfig, logger logging.Logger) *codeValidator {
	return &codeValidator{
		config: cfg,
		logger: logger,
	}
}

// validateFile validates a file by checking its size, extension, and content (meant to be when writing file to cache)
func (v *codeValidator) validateFile(filePath string) (*types.ValidationResult, error) {
	result := &types.ValidationResult{
		IsValid:    true,
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
		Complexity: 0.0,
	}

	// Check file size
	if err := v.validateFileSize(filePath, result); err != nil {
		return result, err
	}

	// Check file extension
	if err := v.validateFileExtension(filePath, result); err != nil {
		return result, err
	}

	// Read and validate file content
	if err := v.validateFileContent(filePath, result); err != nil {
		return result, err
	}

	// Calculate complexity
	result.Complexity = v.calculateComplexity(filePath)

	return result, nil
}

func (v *codeValidator) validateFileContent(filePath string, result *types.ValidationResult) error {
	content, err := os.ReadFile(filePath) // Read the entire file into memory
	if err != nil {
		return fmt.Errorf("failed to read file for validation: %w", err)
	}

    // Delegate the actual pattern matching to the function that handles byte slices
	v.validateContentPatterns(content, result)

	return nil
}

func (v *codeValidator) validateFileSize(filePath string, result *types.ValidationResult) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() > v.config.MaxFileSize {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file size exceeds limit: %d bytes", fileInfo.Size()))
	}

	return nil
}

func (v *codeValidator) validateFileExtension(filePath string, result *types.ValidationResult) error {
	ext := filepath.Ext(filePath)

	// Check if extension is allowed
	allowed := false
	for _, allowedExt := range v.config.AllowedExtensions {
		if ext == allowedExt {
			allowed = true
			break
		}
	}

	if !allowed {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file extension not allowed: %s", ext))
	}

	return nil
}

func (v *codeValidator) validateContentPatterns(content []byte, result *types.ValidationResult) {
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	for lineNumber, line := range lines {
		lineNumber++ // Convert to 1-based indexing

		// Check for dangerous patterns
		for _, pattern := range v.config.BlockedPatterns {
			if strings.Contains(line, pattern) {
				result.IsValid = false
				result.Errors = append(result.Errors,
					fmt.Sprintf("dangerous pattern found at line %d: %s", lineNumber, pattern))
			}
		}

		// Check for suspicious patterns (warnings)
		v.checkSuspiciousPatterns(line, lineNumber, result)
	}
}

func (v *codeValidator) checkSuspiciousPatterns(line string, lineNumber int, result *types.ValidationResult) {
	suspiciousPatterns := []string{
		"http://",
		"ftp://",
		"file://",
		"os.Open",
		"ioutil.ReadFile",
		"os.ReadFile",
		"exec.",
		"syscall.",
		"runtime.",
		"unsafe.",
		"reflect.",
		"cgo.",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(line, pattern) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("suspicious pattern at line %d: %s", lineNumber, pattern))
		}
	}
}

func (v *codeValidator) calculateComplexity(filePath string) float64 {
	file, err := os.Open(filePath)
	if err != nil {
		v.logger.Warnf("Failed to open file for complexity calculation: %v", err)
		return 0.0
	}
	defer func() {
		err := file.Close()
		if err != nil {
			v.logger.Errorf("Failed to close file: %v", err)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		v.logger.Warnf("Failed to get file info for complexity calculation: %v", err)
		return 0.0
	}

	// Read file content for complexity calculation
	content := make([]byte, fileInfo.Size())
	_, err = file.Read(content)
	if err != nil {
		v.logger.Warnf("Failed to read file for complexity calculation: %v", err)
		return 0.0
	}

	return v.calculateContentComplexity(content)
}

func (v *codeValidator) calculateContentComplexity(content []byte) float64 {
	contentStr := string(content)

	// Basic complexity calculation based on:
	// 1. File size (in KB)
	// 2. Number of lines
	// 3. Number of functions
	// 4. Number of imports
	// 5. Number of loops and conditionals

	sizeKB := float64(len(content)) / 1024.0
	lines := strings.Split(contentStr, "\n")
	numLines := float64(len(lines))

	// Count functions (basic pattern matching)
	functionCount := float64(strings.Count(contentStr, "func "))

	// Count imports
	importCount := float64(strings.Count(contentStr, "import "))

	// Count loops and conditionals
	loopCount := float64(strings.Count(contentStr, "for ") + strings.Count(contentStr, "range "))
	conditionalCount := float64(strings.Count(contentStr, "if ") + strings.Count(contentStr, "switch "))

	// Calculate complexity score
	complexity := sizeKB*0.1 + // Size factor
		numLines*0.01 + // Line count factor
		functionCount*0.5 + // Function factor
		importCount*0.2 + // Import factor
		(loopCount+conditionalCount)*0.3 // Control flow factor

	return complexity
}
