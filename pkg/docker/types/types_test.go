package types

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestGetLanguageFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected Language
	}{
		{name: "go file", filePath: "test.go", expected: LanguageGo},
		{name: "py file", filePath: "test.py", expected: LanguagePy},
		{name: "js file", filePath: "test.js", expected: LanguageJS},
		{name: "ts file", filePath: "test.ts", expected: LanguageTS},
		{name: "node file", filePath: "test.mjs", expected: LanguageNode},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetLanguageFromFile(test.filePath)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetLanguageFromExtension(t *testing.T) {
	tests := []struct {
		name      string
		extension string
		expected  Language
	}{
		{name: "go extension", extension: ".go", expected: LanguageGo},
		{name: "py extension", extension: ".py", expected: LanguagePy},
		{name: "js extension", extension: "js", expected: LanguageJS},
		{name: "ts extension", extension: "ts", expected: LanguageTS},
		{name: "node extension", extension: "mjs", expected: LanguageNode},
		{name: "unknown extension", extension: ".unknown", expected: LanguageGo},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetLanguageFromExtension(test.extension)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestMemoryLimitBytes(t *testing.T) {
	tests := []struct {
		name        string
		memoryLimit string
		expected    uint64
		expectError bool
	}{
		{name: "bytes", 			memoryLimit: "1024", 	expected: 1024, 		expectError: false},
		{name: "kilobytes", 		memoryLimit: "1k", 		expected: 1024, 		expectError: false},
		{name: "megabytes", 		memoryLimit: "1m", 		expected: 1048576, 		expectError: false},
		{name: "gigabytes", 		memoryLimit: "1g", 		expected: 1073741824, 	expectError: false},
		{name: "decimal megabytes", memoryLimit: "1.5m", 	expected: 1572864, 		expectError: false},
		{name: "zero", 				memoryLimit: "0", 		expected: 0, 			expectError: false},
		{name: "empty string", 		memoryLimit: "", 		expected: 0, 			expectError: true},
		{name: "invalid format", 	memoryLimit: "invalid", expected: 0, 			expectError: true},
		{name: "negative value", 	memoryLimit: "-1m", 	expected: 0, 			expectError: true},
		{name: "large value", 		memoryLimit: "2g", 		expected: 2147483648, 	expectError: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MemoryLimitBytes(test.memoryLimit)

			if test.expectError {
				assert.Equal(t, uint64(0), result, "Expected 0 for invalid input")
			} else {
				assert.Equal(t, test.expected, result, "Memory limit conversion failed")
			}
		})
	}
}

func TestToContainerResources(t *testing.T) {
	tests := []struct {
		name        string
		memoryLimit string
		cpuLimit    float64
		expected    container.Resources
	}{
		{
			name:        "basic values",
			memoryLimit: "512m",
			cpuLimit:    1.0,
			expected: container.Resources{
				Memory:   int64(536870912),  // 512MB in bytes
				NanoCPUs: int64(1000000000), // 1.0 * 1e9
			},
		},
		{
			name:        "large memory and CPU",
			memoryLimit: "2g",
			cpuLimit:    4.0,
			expected: container.Resources{
				Memory:   int64(2147483648), // 2GB in bytes
				NanoCPUs: int64(4000000000), // 4.0 * 1e9
			},
		},
		{
			name:        "small values",
			memoryLimit: "64k",
			cpuLimit:    0.5,
			expected: container.Resources{
				Memory:   int64(65536),     // 64KB in bytes
				NanoCPUs: int64(500000000), // 0.5 * 1e9
			},
		},
		{
			name:        "zero values",
			memoryLimit: "0",
			cpuLimit:    0.0,
			expected: container.Resources{
				Memory:   int64(0),
				NanoCPUs: int64(0),
			},
		},
		{
			name:        "decimal CPU",
			memoryLimit: "1g",
			cpuLimit:    2.5,
			expected: container.Resources{
				Memory:   int64(1073741824), // 1GB in bytes
				NanoCPUs: int64(2500000000), // 2.5 * 1e9
			},
		},
		{
			name:        "invalid memory (should return 0)",
			memoryLimit: "invalid",
			cpuLimit:    1.0,
			expected: container.Resources{
				Memory:   int64(0),
				NanoCPUs: int64(1000000000),
			},
		},
		{
			name:        "bytes memory format",
			memoryLimit: "1048576",
			cpuLimit:    1.0,
			expected: container.Resources{
				Memory:   int64(1048576),
				NanoCPUs: int64(1000000000),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ToContainerResources(test.memoryLimit, test.cpuLimit)

			assert.Equal(t, test.expected.Memory, result.Memory, "Memory limit mismatch")
			assert.Equal(t, test.expected.NanoCPUs, result.NanoCPUs, "CPU limit mismatch")
		})
	}
}

func TestToContainerResources_EdgeCases(t *testing.T) {
	t.Run("very large CPU value", func(t *testing.T) {
		result := ToContainerResources("1g", 100.0)
		expected := container.Resources{
			Memory:   int64(1073741824),   // 1GB
			NanoCPUs: int64(100000000000), // 100.0 * 1e9
		}
		assert.Equal(t, expected.Memory, result.Memory)
		assert.Equal(t, expected.NanoCPUs, result.NanoCPUs)
	})

	t.Run("very small CPU value", func(t *testing.T) {
		result := ToContainerResources("1m", 0.001)
		expected := container.Resources{
			Memory:   int64(1048576), // 1MB
			NanoCPUs: int64(1000000), // 0.001 * 1e9
		}
		assert.Equal(t, expected.Memory, result.Memory)
		assert.Equal(t, expected.NanoCPUs, result.NanoCPUs)
	})

	t.Run("empty memory string", func(t *testing.T) {
		result := ToContainerResources("", 1.0)
		expected := container.Resources{
			Memory:   int64(0),
			NanoCPUs: int64(1000000000),
		}
		assert.Equal(t, expected.Memory, result.Memory)
		assert.Equal(t, expected.NanoCPUs, result.NanoCPUs)
	})
}
