package file

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type codeValidator struct {
	config config.ValidationConfig
	logger logging.Logger
	fs     fs.FileSystemAPI
}

func newCodeValidator(cfg config.ValidationConfig, logger logging.Logger, fs fs.FileSystemAPI) *codeValidator {
	return &codeValidator{
		config: cfg,
		logger: logger,
		fs:     fs,
	}
}

// validateFile validates a file by checking its size, extension
func (v *codeValidator) validateFile(filePath string) (*types.ValidationResult, error) {
	result := &types.ValidationResult{
		IsValid:    true,
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
		Complexity: 0.0,
	}

	// Check file size
	if err := v.validateFileSize(filePath, result); err != nil {
		return nil, err
	}

	// Check file extension
	if err := v.validateFileExtension(filePath, result); err != nil {
		return nil, err
	}

	// Calculate and validate complexity
	complexity := v.calculateComplexity(filePath)
	result.Complexity = complexity

	if complexity > v.config.MaxComplexity {
		result.IsValid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("file complexity %.2f exceeds limit %.2f", complexity, v.config.MaxComplexity))
	}

	return result, nil
}

// validateFileContent is intentionally removed.
// Security is handled by container sandboxing and seccomp profiles, not naive pattern matching.
// Pattern matching can be trivially bypassed and provides a false sense of security.

func (v *codeValidator) validateFileSize(filePath string, result *types.ValidationResult) error {
	fileInfo, err := v.fs.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Check if file is empty
	if fileInfo.Size() == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "file is empty")
		return nil
	}

	// Check if file exceeds size limit
	if fileInfo.Size() > v.config.MaxFileSize {
		result.IsValid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("file size %d bytes exceeds limit %d bytes", fileInfo.Size(), v.config.MaxFileSize))
	}

	// Add warning for very large files (over 50% of limit)
	if fileInfo.Size() > v.config.MaxFileSize/2 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("file size %d bytes is over 50%% of the limit", fileInfo.Size()))
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

func (v *codeValidator) calculateComplexity(filePath string) float64 {
	content, err := v.fs.ReadFile(filePath)
	if err != nil {
		v.logger.Warnf("Failed to read file for complexity calculation: %v", err)
		return 0.0
	}

	return v.calculateContentComplexity(content)
}

func (v *codeValidator) calculateContentComplexity(content []byte) float64 {
	// Handle empty content
	if len(content) == 0 {
		return 0.0
	}

	contentStr := string(content)

	// Enhanced complexity calculation based on:
	// 1. File size (in KB) - larger files are more complex
	// 2. Number of lines - more lines = more complexity
	// 3. Number of functions/methods - more functions = more complexity
	// 4. Number of imports/dependencies - more imports = more complexity
	// 5. Number of control flow structures - more loops/conditionals = more complexity
	// 6. Nesting depth - deeper nesting = more complexity

	sizeKB := float64(len(content)) / 1024.0
	lines := strings.Split(contentStr, "\n")
	numLines := float64(len(lines))

	// Count functions/methods (Go, Python, JavaScript, TypeScript)
	functionCount := float64(
		// Go
		strings.Count(contentStr, "func ") +
		// Python
		strings.Count(contentStr, "def ") +
		strings.Count(contentStr, "class ") +
		strings.Count(contentStr, "async def ") +
		strings.Count(contentStr, "lambda ") +
		// JavaScript/TypeScript
		strings.Count(contentStr, "function ") +
		strings.Count(contentStr, "function*") +
		strings.Count(contentStr, "async function") +
		strings.Count(contentStr, "() =>") +
		strings.Count(contentStr, "=> {") +
		strings.Count(contentStr, "constructor(") +
		strings.Count(contentStr, "get ") +
		strings.Count(contentStr, "set "))

	// Count imports/dependencies (Go, Python, JavaScript, TypeScript)
	importCount := float64(
		// Go
		strings.Count(contentStr, "import ") +
		strings.Count(contentStr, "import(") +
		// Python
		strings.Count(contentStr, "from ") +
		strings.Count(contentStr, "import ") +
		// JavaScript/TypeScript
		strings.Count(contentStr, "require(") +
		strings.Count(contentStr, "import {") +
		strings.Count(contentStr, "import * as") +
		strings.Count(contentStr, "import type") +
		strings.Count(contentStr, "export {") +
		strings.Count(contentStr, "export *") +
		strings.Count(contentStr, "export default"))

	// Count control flow structures (Go, Python, JavaScript, TypeScript)
	controlFlowCount := float64(
		// Common to all languages
		strings.Count(contentStr, "for ") +
		strings.Count(contentStr, "while ") +
		strings.Count(contentStr, "if ") +
		strings.Count(contentStr, "else ") +
		strings.Count(contentStr, "elif ") +
		strings.Count(contentStr, "switch ") +
		strings.Count(contentStr, "case ") +
		strings.Count(contentStr, "default:") +
		strings.Count(contentStr, "try ") +
		strings.Count(contentStr, "catch ") +
		strings.Count(contentStr, "except ") +
		strings.Count(contentStr, "finally ") +
		strings.Count(contentStr, "with ") +
		// Go specific
		strings.Count(contentStr, "range ") +
		strings.Count(contentStr, "select ") +
		strings.Count(contentStr, "go ") +
		strings.Count(contentStr, "defer ") +
		// Python specific
		strings.Count(contentStr, "for ") +
		strings.Count(contentStr, "async for ") +
		strings.Count(contentStr, "async with ") +
		// JavaScript/TypeScript specific
		strings.Count(contentStr, "do {") +
		strings.Count(contentStr, "await ") +
		strings.Count(contentStr, "yield "))

	// Estimate nesting depth by counting braces/brackets
	openBraces := strings.Count(contentStr, "{")
	closeBraces := strings.Count(contentStr, "}")
	nestingDepth := float64(openBraces - closeBraces)
	if nestingDepth < 0 {
		nestingDepth = 0
	}

	// Calculate weighted complexity score
	complexity := sizeKB*0.05 + // Size factor (reduced weight)
		numLines*0.02 + // Line count factor
		functionCount*0.8 + // Function factor (increased weight)
		importCount*0.3 + // Import factor
		controlFlowCount*0.4 + // Control flow factor
		nestingDepth*0.5 // Nesting depth factor

	return complexity
}
