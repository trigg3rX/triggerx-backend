package file

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type CodeValidator struct {
	config config.ValidationConfig
	logger logging.Logger
}

func NewCodeValidator(cfg config.ValidationConfig, logger logging.Logger) *CodeValidator {
	return &CodeValidator{
		config: cfg,
		logger: logger,
	}
}

func (v *CodeValidator) ValidateFile(filePath string) (*types.ValidationResult, error) {
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

func (v *CodeValidator) ValidateContent(content []byte) (*types.ValidationResult, error) {
	result := &types.ValidationResult{
		IsValid:    true,
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
		Complexity: 0.0,
	}

	// Check content size
	if int64(len(content)) > v.config.MaxFileSize {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file size exceeds limit: %d bytes", len(content)))
		return result, nil
	}

	// Validate content for dangerous patterns
	v.validateContentPatterns(content, result)

	// Calculate complexity
	result.Complexity = v.calculateContentComplexity(content)

	return result, nil
}

func (v *CodeValidator) validateFileSize(filePath string, result *types.ValidationResult) error {
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

func (v *CodeValidator) validateFileExtension(filePath string, result *types.ValidationResult) error {
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

func (v *CodeValidator) validateFileContent(filePath string, result *types.ValidationResult) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

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

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

func (v *CodeValidator) validateContentPatterns(content []byte, result *types.ValidationResult) {
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

func (v *CodeValidator) checkSuspiciousPatterns(line string, lineNumber int, result *types.ValidationResult) {
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

func (v *CodeValidator) calculateComplexity(filePath string) float64 {
	file, err := os.Open(filePath)
	if err != nil {
		v.logger.Warnf("Failed to open file for complexity calculation: %v", err)
		return 0.0
	}
	defer file.Close()

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

func (v *CodeValidator) calculateContentComplexity(content []byte) float64 {
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

func (v *CodeValidator) IsFileValid(filePath string) (bool, error) {
	result, err := v.ValidateFile(filePath)
	if err != nil {
		return false, err
	}
	return result.IsValid, nil
}

func (v *CodeValidator) IsContentValid(content []byte) (bool, error) {
	result, err := v.ValidateContent(content)
	if err != nil {
		return false, err
	}
	return result.IsValid, nil
}
