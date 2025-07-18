package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
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
	m.logger.Infof("Pulling Docker image: %s", imageName)

	// Check if image already exists locally
	images, err := m.Cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		m.logger.Warnf("Failed to list images: %v", err)
	} else {
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if tag == imageName || tag == imageName+":latest" {
					m.logger.Debugf("Image %s already exists locally, skipping pull", imageName)
					return nil
				}
			}
		}
	}

	// Image doesn't exist locally, pull it
	reader, err := m.Cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		m.logger.Errorf("Failed to pull image: %v", err)
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	// Read the output to ensure the pull completes
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		m.logger.Errorf("Error reading image pull response: %v", err)
		return fmt.Errorf("error reading image pull response: %w", err)
	}

	// Log the last few lines of the pull output for debugging
	pullOutput := buf.String()
	lines := strings.Split(pullOutput, "\n")
	if len(lines) > 5 {
		lines = lines[len(lines)-5:]
	}
	for _, line := range lines {
		if line != "" {
			m.logger.Debugf("Pull output: %s", line)
		}
	}

	m.logger.Infof("Successfully pulled image: %s", imageName)
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

	m.logger.Infof("Creating container with code directory: %s", absPath)

	// List directory contents for debugging
	if entries, err := os.ReadDir(absPath); err == nil {
		m.logger.Infof("Directory contents of %s:", absPath)
		for _, entry := range entries {
			m.logger.Infof("  - %s (dir: %v)", entry.Name(), entry.IsDir())
			// Also check permissions
			if info, err := os.Stat(filepath.Join(absPath, entry.Name())); err == nil {
				m.logger.Infof("    - permissions: %v", info.Mode())
			}
		}
	}

	// For Docker-in-Docker, make sure the mount path is absolute and exists on the host
	hostMountPath := absPath
	if !filepath.IsAbs(hostMountPath) {
		hostMountPath, _ = filepath.Abs(hostMountPath)
	}

	m.logger.Infof("Using host mount path: %s", hostMountPath)

	// Create a command that runs the setup script content directly
	// First list directories, then copy code.go to /tmp if needed, then create and run setup script
	setupCommand := `
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

	config := &container.Config{
		Image:      m.config.Image,
		Cmd:        []string{"sh", "-c", setupCommand},
		Tty:        true,
		WorkingDir: "/code",
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/code:rw", hostMountPath),
			"/var/run/docker.sock:/var/run/docker.sock", // Ensure Docker socket is mounted
		},
		Resources: container.Resources{
			Memory:   int64(m.config.MemoryLimitBytes()),
			NanoCPUs: int64(m.config.CPULimit * 1e9),
		},
		Privileged: true, // Add privileged mode for Docker-in-Docker
	}

	m.logger.Infof("Creating container with bind mount: %s:/code", hostMountPath)

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
