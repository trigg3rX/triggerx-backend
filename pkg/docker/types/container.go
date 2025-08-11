package types

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
)

type Language string

const (
	LanguageGo   Language = "go"
	LanguagePy   Language = "py"
	LanguageJS   Language = "js"
	LanguageTS   Language = "ts"
	LanguageNode Language = "node"
)

type ContainerStatus string

const (
	ContainerStatusPending ContainerStatus = "pending"
	ContainerStatusRunning ContainerStatus = "running"
	ContainerStatusStopped ContainerStatus = "stopped"
	ContainerStatusReady   ContainerStatus = "ready"
	ContainerStatusError   ContainerStatus = "error"
)

type PooledContainer struct {
	ID         string          `json:"id"`
	Status     ContainerStatus `json:"status"`
	LastUsed   time.Time       `json:"last_used"`
	IsReady    bool            `json:"is_ready"`
	WorkingDir string          `json:"working_dir"`
	ImageName  string          `json:"image_name"`
	Language   Language        `json:"language"`
	CreatedAt  time.Time       `json:"created_at"`
	Error      error           `json:"error,omitempty"`
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

// GetLanguageFromExtension returns the language based on file extension string
func GetLanguageFromExtension(extension string) Language {
	ext := strings.ToLower(extension)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return GetLanguageFromFile("dummy" + ext)
}

type ContainerConfig struct {
	Image          string              `json:"image"`
	WorkingDir     string              `json:"working_dir"`
	Environment    []string            `json:"environment"`
	Binds          []string            `json:"binds"`
	Resources      container.Resources `json:"resources"`
	Privileged     bool                `json:"privileged"`
	NetworkMode    string              `json:"network_mode"`
	SecurityOpt    []string            `json:"security_opt"`
	ReadOnlyRootFS bool                `json:"read_only_root_fs"`
}

type ContainerInfo struct {
	ID      string               `json:"id"`
	Config  *ContainerConfig     `json:"config"`
	Stats   *DockerResourceStats `json:"stats,omitempty"`
	Created time.Time            `json:"created"`
}
