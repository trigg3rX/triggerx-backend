package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	SetupScript = `#!/bin/sh
cd /code
go mod init code
go mod tidy
echo "START_EXECUTION"
go run code.go 2>&1 || {
    echo "Error executing Go program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
)

type Manager struct {
	Cli  *client.Client
	config  DockerConfig
}

func NewManager(cli *client.Client, config DockerConfig) *Manager {
	return &Manager{
		Cli:    cli,
		config: config,
	}
}

func (m *Manager) CreateContainer(ctx context.Context, codePath string) (string, error) {
	absPath, err := filepath.Abs(codePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	setupScriptPath := filepath.Join(filepath.Dir(absPath), "setup.sh")
	if err := os.WriteFile(setupScriptPath, []byte(SetupScript), 0755); err != nil {
		return "", fmt.Errorf("failed to write setup script: %w", err)
	}

	if err := os.Chmod(setupScriptPath, 0755); err != nil {
		return "", fmt.Errorf("failed to set permissions for setup script: %w", err)
	}

	config := &container.Config{
		Image:      m.config.Image,
		Cmd:        []string{"/code/setup.sh"},
		Tty:        true,
		WorkingDir: "/code",
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/code", absPath),
		},
		Resources: container.Resources{
			Memory:     int64(m.config.MemoryLimitBytes()),
			NanoCPUs:   int64(m.config.CPULimit * 1e9),
		},
	}

	resp, err := m.Cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func (m *Manager) CleanupContainer(ctx context.Context, containerID string) error {
	if !m.config.AutoCleanup {
		return nil
	}
	
	return m.Cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}

func (m *Manager) GetContainerInfo(ctx context.Context, containerID string) (*container.InspectResponse, error) {
	info, err := m.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container info: %w", err)
	}
	return &info, nil
}
