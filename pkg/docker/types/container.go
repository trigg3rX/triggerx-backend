package types

import (
	"time"

	"github.com/docker/docker/api/types/container"
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
	CreatedAt  time.Time       `json:"created_at"`
	Error      error           `json:"error,omitempty"`
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
