package types

import (
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
)

func MemoryLimitBytes(memoryLimit string) uint64 {
	memoryLimitBytes, err := units.RAMInBytes(memoryLimit)
	if err != nil {
		return 0
	}
	return uint64(memoryLimitBytes)
}

func ToContainerResources(memoryLimit string, cpuLimit float64) container.Resources {
	return container.Resources{
		Memory:   int64(MemoryLimitBytes(memoryLimit)),
		NanoCPUs: int64(cpuLimit * 1e9),
	}
}

// GetLanguageFromExtension returns the language based on file extension string
func GetLanguageFromExtension(extension string) Language {
	ext := strings.ToLower(extension)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return GetLanguageFromFile("dummy" + ext)
}

// GetLanguageFromFile returns the language based on file extension
func GetLanguageFromFile(filePath string) Language {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return LanguageGo
	case ".py":
		return LanguagePy
	case ".js":
		return LanguageJS
	case ".ts":
		return LanguageTS
	case ".mjs", ".cjs":
		return LanguageNode
	default:
		return LanguageGo // Default to Go
	}
}
