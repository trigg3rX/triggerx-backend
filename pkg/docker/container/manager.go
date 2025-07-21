package container

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
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
	Cli        *client.Client
	config     config.DockerConfig
	logger     logging.Logger
	pool       *ContainerPool
	lifecycle  *ContainerLifecycle
	poolConfig config.PoolConfig
}

func NewManager(cli *client.Client, cfg config.ExecutorConfig, logger logging.Logger) (*Manager, error) {
	manager := &Manager{
		Cli:        cli,
		config:     cfg.Docker,
		logger:     logger,
		poolConfig: cfg.Pool,
	}

	// Create lifecycle manager
	manager.lifecycle = NewContainerLifecycle(manager, logger)

	// Create container pool
	manager.pool = NewContainerPool(cfg.Pool, manager, logger)

	return manager, nil
}

func (m *Manager) Initialize(ctx context.Context) error {
	m.logger.Info("Initializing Docker manager")

	// Pull the base image
	if err := m.PullImage(ctx, m.config.Image); err != nil {
		return fmt.Errorf("failed to pull base image: %w", err)
	}

	// Initialize container pool
	if err := m.pool.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize container pool: %w", err)
	}

	m.logger.Info("Docker manager initialized successfully")
	return nil
}

func (m *Manager) GetContainer(ctx context.Context) (*types.PooledContainer, error) {
	return m.pool.GetContainer(ctx)
}

func (m *Manager) ReturnContainer(container *types.PooledContainer) error {
	return m.pool.ReturnContainer(container)
}

func (m *Manager) ExecuteInContainer(ctx context.Context, containerID string, filePath string) (*types.ExecutionResult, error) {
	m.logger.Infof("Executing file %s in container %s", filePath, containerID)

	// Verify container is running before execution
	inspect, err := m.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container before execution: %w", err)
	}
	m.logger.Debugf("Container %s state: %s, running: %v", containerID, inspect.State.Status, inspect.State.Running)

	if !inspect.State.Running {
		return nil, fmt.Errorf("container %s is not running (status: %s)", containerID, inspect.State.Status)
	}

	// Copy the file to the container
	// m.logger.Debugf("Starting file copy to container %s", containerID)
	if err := m.copyFileToContainer(ctx, containerID, filePath); err != nil {
		return nil, fmt.Errorf("failed to copy file to container: %w", err)
	}
	m.logger.Debugf("File copy completed for container %s", containerID)

	// Execute the code
	m.logger.Debugf("Starting code execution in container %s", containerID)
	result, err := m.executeCode(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}
	m.logger.Debugf("Code execution completed for container %s", containerID)

	return result, nil
}

func (m *Manager) PullImage(ctx context.Context, imageName string) error {
	// m.logger.Infof("Pulling Docker image: %s", imageName)

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

	m.logger.Debugf("Creating container with code directory: %s", absPath)

	// List directory contents for debugging
	// if entries, err := os.ReadDir(absPath); err == nil {
	// 	m.logger.Debugf("Directory contents of %s:", absPath)
	// 	for _, entry := range entries {
	// 		m.logger.Debugf("  - %s (dir: %v)", entry.Name(), entry.IsDir())
	// 		if info, err := os.Stat(filepath.Join(absPath, entry.Name())); err == nil {
	// 			m.logger.Debugf("    - permissions: %v", info.Mode())
	// 		}
	// 	}
	// }

	// For Docker-in-Docker, make sure the mount path is absolute and exists on the host
	hostMountPath := absPath
	if !filepath.IsAbs(hostMountPath) {
		hostMountPath, _ = filepath.Abs(hostMountPath)
	}

	// m.logger.Debugf("Using host mount path: %s", hostMountPath)

	// Create a simple keep-alive command that keeps the container running
	// We'll execute the actual code later via exec
	keepAliveCommand := `tail -f /dev/null`

	config := &container.Config{
		Image:      m.config.Image,
		Cmd:        []string{"sh", "-c", keepAliveCommand},
		Tty:        true, // Don't allocate TTY for keep-alive
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

	// m.logger.Debugf("Creating container with bind mount: %s:/code", hostMountPath)

	resp, err := m.Cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		m.logger.Errorf("failed to create container: %v", err)
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID
	m.logger.Infof("Container created with ID: %s", containerID)

	// Start the container
	// m.logger.Infof("Starting container: %s", containerID)
	err = m.Cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		m.logger.Errorf("failed to start container: %v", err)
		// Try to cleanup the created container
		if cleanupErr := m.Cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); cleanupErr != nil {
			m.logger.Warnf("Failed to cleanup container after start failure: %v", cleanupErr)
		}
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be running
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		inspect, err := m.Cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return "", fmt.Errorf("failed to inspect container after start: %w", err)
		}

		if inspect.State.Running {
			m.logger.Infof("Container %s is running", containerID)
			return containerID, nil
		}

		m.logger.Debugf("Container %s not running yet (attempt %d/%d), status: %s", containerID, i+1, maxRetries, inspect.State.Status)
		time.Sleep(500 * time.Millisecond)
	}

	return "", fmt.Errorf("container %s failed to start properly", containerID)
}

func (m *Manager) CleanupContainer(ctx context.Context, containerID string) error {
	if !m.config.AutoCleanup {
		m.logger.Infof("auto cleanup is disabled, skipping container cleanup")
		return nil
	}

	return m.Cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

func (m *Manager) GetContainerInfo(ctx context.Context, containerID string) (*dockertypes.ContainerJSON, error) {
	info, err := m.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		m.logger.Errorf("failed to get container info: %v", err)
		return nil, fmt.Errorf("failed to get container info: %w", err)
	}
	return &info, nil
}

func (m *Manager) copyFileToContainer(ctx context.Context, containerID string, filePath string) error {
	// m.logger.Debugf("Copying file %s to container %s", filePath, containerID)

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	// m.logger.Debugf("Read %d bytes from file %s", len(content), filePath)

	// First, ensure the /code directory exists and check container state
	setupCmd := []string{"sh", "-c", "mkdir -p /code && ls -la /code && pwd && whoami"}
	// m.logger.Debugf("Setup command for container %s: %v", containerID, setupCmd)

	setupExecConfig := &container.ExecOptions{
		Cmd:          setupCmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// m.logger.Debugf("Creating setup exec for container %s", containerID)
	setupExecResp, err := m.Cli.ContainerExecCreate(ctx, containerID, *setupExecConfig)
	if err != nil {
		return fmt.Errorf("failed to create setup exec: %w", err)
	}

	// Execute the setup command
	err = m.Cli.ContainerExecStart(ctx, setupExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start setup exec: %w", err)
	}

	// Wait for setup completion
	for {
		inspectResp, err := m.Cli.ContainerExecInspect(ctx, setupExecResp.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect setup exec: %w", err)
		}

		if !inspectResp.Running {
			// m.logger.Debugf("Setup exec completed with exit code %d", inspectResp.ExitCode)
			if inspectResp.ExitCode != 0 {
				return fmt.Errorf("setup failed with exit code: %d", inspectResp.ExitCode)
			}
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Now copy the file using a simpler approach
	// Escape the content properly for shell
	escapedContent := strings.ReplaceAll(string(content), "'", "'\"'\"'")
	copyCmd := []string{"sh", "-c", fmt.Sprintf("echo '%s' > /code/code.go", escapedContent)}
	// m.logger.Debugf("Copy command for container %s: %v", containerID, copyCmd)

	execConfig := &container.ExecOptions{
		Cmd:          copyCmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// m.logger.Debugf("Creating exec for container %s", containerID)
	execResp, err := m.Cli.ContainerExecCreate(ctx, containerID, *execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}
	// m.logger.Debugf("Created exec %s for container %s", execResp.ID, containerID)

	// Execute the copy command
	// m.logger.Debugf("Starting exec %s for container %s", execResp.ID, containerID)
	err = m.Cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}
	// m.logger.Debugf("Started exec %s for container %s", execResp.ID, containerID)

	// Wait for completion
	// m.logger.Debugf("Waiting for exec %s to complete", execResp.ID)
	for {
		inspectResp, err := m.Cli.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect exec: %w", err)
		}

		if !inspectResp.Running {
			// m.logger.Debugf("Exec %s completed with exit code %d", execResp.ID, inspectResp.ExitCode)
			if inspectResp.ExitCode != 0 {
				return fmt.Errorf("file copy failed with exit code: %d", inspectResp.ExitCode)
			}
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Verify the file was copied
	verifyCmd := []string{"sh", "-c", "ls -la /code/code.go && wc -l /code/code.go"}
	// m.logger.Debugf("Verify command for container %s: %v", containerID, verifyCmd)

	verifyExecConfig := &container.ExecOptions{
		Cmd:          verifyCmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	verifyExecResp, err := m.Cli.ContainerExecCreate(ctx, containerID, *verifyExecConfig)
	if err != nil {
		return fmt.Errorf("failed to create verify exec: %w", err)
	}

	err = m.Cli.ContainerExecStart(ctx, verifyExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start verify exec: %w", err)
	}

	for {
		inspectResp, err := m.Cli.ContainerExecInspect(ctx, verifyExecResp.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect verify exec: %w", err)
		}

		if !inspectResp.Running {
			// m.logger.Debugf("Verify exec completed with exit code %d", inspectResp.ExitCode)
			if inspectResp.ExitCode != 0 {
				return fmt.Errorf("file verification failed with exit code: %d", inspectResp.ExitCode)
			}
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	// m.logger.Debugf("File copy completed successfully for container %s", containerID)
	return nil
}

func (m *Manager) executeCode(ctx context.Context, containerID string) (*types.ExecutionResult, error) {
	result := &types.ExecutionResult{}
	var executionStartTime time.Time
	var executionEndTime time.Time
	var codeExecutionTime time.Duration
	executionStarted := false
	var outputBuffer bytes.Buffer

	execCmd := []string{"sh", "-c", SetupScript}

	execConfig := &container.ExecOptions{
		Cmd:          execCmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := m.Cli.ContainerExecCreate(ctx, containerID, *execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	execAttachResp, err := m.Cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer execAttachResp.Close()
	scanner := bufio.NewScanner(execAttachResp.Reader)

	// Execute the command
	err = m.Cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start exec: %w", err)
	}

	for scanner.Scan() {
		line := scanner.Text()
		// m.logger.Debugf("Container Log: %s\n", line)

		if strings.Contains(line, "START_EXECUTION") {
			executionStartTime = time.Now().UTC()
			executionStarted = true
		} else if strings.Contains(line, "END_EXECUTION") && executionStarted {
			executionEndTime = time.Now().UTC()
			codeExecutionTime = executionEndTime.Sub(executionStartTime)
			// m.logger.Debugf("Code execution completed in: %v\n", codeExecutionTime)
			break
		} else if executionStarted {
			outputBuffer.WriteString(line)
			m.logger.Debugf("Container Log: %s\n", line)
		}
	}

	// Wait for exec to finish (in case END_EXECUTION is not printed)
	for {
		inspectResp, err := m.Cli.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect exec: %w", err)
		}
		if !inspectResp.Running {
			if !executionStarted || codeExecutionTime == 0 {
				codeExecutionTime = time.Since(executionStartTime)
				m.logger.Debugf("Warning: Could not determine precise code execution time, using container execution time instead")
			}
			result.Success = inspectResp.ExitCode == 0
			if inspectResp.ExitCode != 0 {
				result.Error = fmt.Errorf("execution failed with exit code: %d", inspectResp.ExitCode)
			}
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	result.Output = outputBuffer.String()
	result.Stats.ExecutionTime = codeExecutionTime
	return result, nil
}

func (m *Manager) GetPoolStats() *types.PoolStats {
	return m.pool.GetStats()
}

func (m *Manager) Close() error {
	// m.logger.Info("Closing Docker manager")

	// Close container pool
	if m.pool != nil {
		if err := m.pool.Close(); err != nil {
			m.logger.Warnf("Failed to close container pool: %v", err)
		}
	}

	m.logger.Info("Docker manager closed")
	return nil
}
