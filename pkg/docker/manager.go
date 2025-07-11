package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
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
	Cli    *client.Client
	config DockerConfig
	logger logging.Logger
}

func NewManager(cli *client.Client, config DockerConfig, logger logging.Logger) *Manager {
	return &Manager{
		Cli:    cli,
		config: config,
		logger: logger,
	}
}

func (m *Manager) PullImage(ctx context.Context, imageName string) error {
	_, err := m.Cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		m.logger.Errorf("failed to pull image: %v", err)
		return fmt.Errorf("failed to pull image: %w", err)
	}
	return nil
}

func (m *Manager) CleanupImages(ctx context.Context) error {
	images, err := m.Cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		m.logger.Errorf("failed to list images: %v", err)
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, dockerImage := range images {
		_, err := m.Cli.ImageRemove(ctx, dockerImage.ID, image.RemoveOptions{Force: true})
		if err != nil {
			m.logger.Errorf("failed to remove image: %v", err)
		}
	}
	return nil
}

func (m *Manager) CreateContainer(ctx context.Context, codePath string) (string, error) {
	absPath, err := filepath.Abs(codePath)
	if err != nil {
		m.logger.Errorf("failed to get absolute path: %v", err)
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	setupScriptPath := filepath.Join(filepath.Dir(absPath), "setup.sh")
	if err := os.WriteFile(setupScriptPath, []byte(SetupScript), 0755); err != nil {
		m.logger.Errorf("failed to write setup script: %v", err)
		return "", fmt.Errorf("failed to write setup script: %w", err)
	}

	if err := os.Chmod(setupScriptPath, 0755); err != nil {
		m.logger.Errorf("failed to set permissions for setup script: %v", err)
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
			Memory:   int64(m.config.MemoryLimitBytes()),
			NanoCPUs: int64(m.config.CPULimit * 1e9),
		},
	}

	resp, err := m.Cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		m.logger.Errorf("failed to create container: %v", err)
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func (m *Manager) CleanupContainer(ctx context.Context, containerID string) error {
	if !m.config.AutoCleanup {
		m.logger.Infof("auto cleanup is disabled, skipping container cleanup")
		return nil
	}

	return m.Cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

func (m *Manager) GetContainerInfo(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
	info, err := m.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		m.logger.Errorf("failed to get container info: %v", err)
		return nil, fmt.Errorf("failed to get container info: %w", err)
	}
	return &info, nil
}
