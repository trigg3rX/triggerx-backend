package container

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/scripts"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type ContainerPool struct {
	language      types.Language
	containers    map[string]*types.PooledContainer
	mutex         sync.RWMutex
	config        config.LanguagePoolConfig
	logger        logging.Logger
	manager       *Manager
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	stats         *types.PoolStats
	statsMutex    sync.RWMutex
	waitQueue     chan struct{}
}

func NewContainerPool(cfg config.LanguagePoolConfig, manager *Manager, logger logging.Logger) *ContainerPool {
	pool := &ContainerPool{
		language:    cfg.Language,
		containers:  make(map[string]*types.PooledContainer),
		config:      cfg,
		logger:      logger,
		manager:     manager,
		stopCleanup: make(chan struct{}),
		waitQueue:   make(chan struct{}, cfg.MaxContainers),
		stats: &types.PoolStats{
			Language:          cfg.Language,
			TotalContainers:   0,
			ReadyContainers:   0,
			BusyContainers:    0,
			ErrorContainers:   0,
			UtilizationRate:   0.0,
			AverageWaitTime:   0,
			MaxWaitTime:       0,
			ContainerLifetime: 0,
			CreatedCount:      0,
			DestroyedCount:    0,
			LastCleanup:       time.Now(),
		},
	}

	// Start cleanup routine
	pool.startCleanupRoutine()

	// Start health check routine
	pool.startHealthCheckRoutine()

	return pool
}

func (p *ContainerPool) Initialize(ctx context.Context) error {
	// p.logger.Infof("Initializing %s language pool with %d pre-warmed containers", p.language, p.config.PreWarmCount)

	// Pull the language-specific image
	if err := p.pullImage(ctx, p.config.Config.ImageName); err != nil {
		return fmt.Errorf("failed to pull image for language %s: %w", p.language, err)
	}

	// Pre-warm containers in parallel for faster initialization
	containerChan := make(chan *types.PooledContainer, p.config.PreWarmCount)
	errorChan := make(chan error, p.config.PreWarmCount)

	// Start parallel container creation
	for i := 0; i < p.config.PreWarmCount; i++ {
		go func(index int) {
			container, err := p.createPreparedContainer(ctx)
			if err != nil {
				p.logger.Warnf("Failed to create pre-warmed container %d for language %s: %v", index, p.language, err)
				errorChan <- err
				return
			}
			containerChan <- container
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < p.config.PreWarmCount; i++ {
		select {
		case container := <-containerChan:
			successCount++
			p.logger.Debugf("Successfully created pre-warmed container %d: %s", successCount, container.ID)
		case err := <-errorChan:
			p.logger.Warnf("Container creation failed: %v", err)
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during container initialization: %w", ctx.Err())
		}
	}

	p.logger.Infof("%s language pool initialized with %d containers", p.language, successCount)
	return nil
}

func (p *ContainerPool) GetContainer(ctx context.Context) (*types.PooledContainer, error) {
	// Try to get a ready container
	p.mutex.Lock()
	for _, container := range p.containers {
		// p.logger.Debugf("Checking container %s: status=%s, ready=%v", container.ID, container.Status, container.IsReady)

		if container.Status == types.ContainerStatusReady && container.IsReady {
			// p.logger.Debugf("Found ready container %s, verifying it's running", container.ID)

			// Verify container is actually running before returning it
			inspect, err := p.manager.Cli.ContainerInspect(ctx, container.ID)
			if err != nil {
				p.logger.Warnf("Container %s health check failed during get: %v", container.ID, err)
				container.Status = types.ContainerStatusError
				container.Error = err
				continue
			}

			// p.logger.Debugf("Container %s inspect result: status=%s, running=%v", container.ID, inspect.State.Status, inspect.State.Running)

			if !inspect.State.Running {
				p.logger.Warnf("Container %s is not running, marking as error", container.ID)
				container.Status = types.ContainerStatusError
				container.IsReady = false
				continue
			}

			p.logger.Debugf("Container %s verified as running, returning it", container.ID)
			container.Status = types.ContainerStatusRunning
			container.LastUsed = time.Now()
			p.updateStats()
			p.mutex.Unlock()
			return container, nil
		}
	}
	p.mutex.Unlock()

	// No ready containers available, try to create a new one
	if p.getTotalContainerCount() < p.config.MaxContainers {
		container, err := p.createPreparedContainer(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create container for language %s: %w", p.language, err)
		}
		container.Status = types.ContainerStatusRunning
		container.LastUsed = time.Now()
		return container, nil
	}

	// Wait for a container to become available
	select {
	case <-p.waitQueue:
		// Try again after waiting
		return p.GetContainer(ctx)
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for container: %w", ctx.Err())
	case <-time.After(p.config.MaxWaitTime):
		return nil, fmt.Errorf("timeout waiting for container in language pool %s", p.language)
	}
}

func (p *ContainerPool) ReturnContainer(container *types.PooledContainer) error {
	p.logger.Debugf("Returning container %s to %s language pool", container.ID, p.language)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if pooledContainer, exists := p.containers[container.ID]; exists {
		p.logger.Debugf("Found container %s in pool, resetting it", container.ID)

		// Reset container for reuse
		if err := p.resetContainer(container.ID); err != nil {
			p.logger.Warnf("Failed to reset container %s: %v", container.ID, err)
			// Mark as error and remove from pool
			pooledContainer.Status = types.ContainerStatusError
			pooledContainer.Error = err
			pooledContainer.IsReady = false
			p.updateStats()

			// Try to cleanup the failed container
			if cleanupErr := p.manager.CleanupContainer(context.Background(), container.ID); cleanupErr != nil {
				p.logger.Warnf("Failed to cleanup failed container %s: %v", container.ID, cleanupErr)
			} else {
				delete(p.containers, container.ID)
			}

			return err
		}

		p.logger.Debugf("Container %s reset successfully, marking as ready", container.ID)
		pooledContainer.Status = types.ContainerStatusReady
		pooledContainer.IsReady = true
		pooledContainer.Error = nil // Clear any previous errors
		p.updateStats()

		// Signal that a container is available
		select {
		case p.waitQueue <- struct{}{}:
		default:
			// Channel is full, no one is waiting
		}

		p.logger.Debugf("Returned container %s to %s language pool", container.ID, p.language)
	} else {
		p.logger.Warnf("Container %s not found in %s language pool", container.ID, p.language)
	}

	return nil
}

func (p *ContainerPool) createPreparedContainer(ctx context.Context) (*types.PooledContainer, error) {
	// Create a temporary directory for the container
	tmpDir, err := p.createTempDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create container using the language-specific configuration
	containerID, err := p.createContainer(ctx, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Initialize the container with /code folder and basic setup
	if err := p.initializeContainer(ctx, containerID); err != nil {
		// Cleanup container if initialization fails
		if cleanupErr := p.manager.CleanupContainer(ctx, containerID); cleanupErr != nil {
			p.logger.Warnf("Failed to cleanup container %s after initialization failure: %v", containerID, cleanupErr)
		}
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	// Create pooled container
	pooledContainer := &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusReady,
		LastUsed:   time.Now(),
		IsReady:    true,
		WorkingDir: tmpDir,
		ImageName:  p.config.Config.ImageName,
		Language:   p.language,
		CreatedAt:  time.Now(),
	}

	p.mutex.Lock()
	p.containers[containerID] = pooledContainer
	p.stats.CreatedCount++
	p.updateStats()
	p.mutex.Unlock()

	p.logger.Infof("Created prepared container for language %s: %s", p.language, containerID)
	return pooledContainer, nil
}

func (p *ContainerPool) createTempDirectory() (string, error) {
	tmpDir := fmt.Sprintf("/tmp/docker-container-%s-%d", p.language, time.Now().UnixNano())
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	return tmpDir, nil
}

// initializeContainer sets up the container with /code folder and basic initialization
func (p *ContainerPool) initializeContainer(ctx context.Context, containerID string) error {
	p.logger.Debugf("Initializing container %s with /code folder setup", containerID)

	// Create /code directory and initialize basic files
	initScript := scripts.GetInitializationScript(p.language)

	execConfig := &container.ExecOptions{
		Cmd:          []string{"sh", "-c", initScript},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := p.manager.Cli.ContainerExecCreate(ctx, containerID, *execConfig)
	if err != nil {
		return fmt.Errorf("failed to create initialization exec: %w", err)
	}

	execAttachResp, err := p.manager.Cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to initialization exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the initialization command
	if err := p.manager.Cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start initialization exec: %w", err)
	}

	// Wait for initialization to complete
	for {
		inspectResp, err := p.manager.Cli.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect initialization exec: %w", err)
		}
		if !inspectResp.Running {
			if inspectResp.ExitCode != 0 {
				return fmt.Errorf("container initialization failed with exit code: %d", inspectResp.ExitCode)
			}
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	p.logger.Debugf("Container %s initialized successfully", containerID)

	// Verify container is ready by running a quick test
	if err := p.verifyContainerReady(ctx, containerID); err != nil {
		return fmt.Errorf("container verification failed: %w", err)
	}

	return nil
}

// verifyContainerReady runs a quick test to ensure the container is fully ready
func (p *ContainerPool) verifyContainerReady(ctx context.Context, containerID string) error {
	p.logger.Debugf("Verifying container %s is ready", containerID)

	// Run a simple test command based on language
	var testCmd string
	switch p.language {
	case types.LanguageGo:
		testCmd = "cd /code && go version"
	case types.LanguagePy:
		testCmd = "cd /code && python --version"
	case types.LanguageJS, types.LanguageNode:
		testCmd = "cd /code && node --version"
	case types.LanguageTS:
		testCmd = "cd /code && tsc --version"
	default:
		testCmd = "cd /code && echo 'ready'"
	}

	execConfig := &container.ExecOptions{
		Cmd:          []string{"sh", "-c", testCmd},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := p.manager.Cli.ContainerExecCreate(ctx, containerID, *execConfig)
	if err != nil {
		return fmt.Errorf("failed to create verification exec: %w", err)
	}

	execAttachResp, err := p.manager.Cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to verification exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the verification command
	if err := p.manager.Cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start verification exec: %w", err)
	}

	// Wait for verification to complete with timeout
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("container verification timed out")
		default:
			inspectResp, err := p.manager.Cli.ContainerExecInspect(ctx, execResp.ID)
			if err != nil {
				return fmt.Errorf("failed to inspect verification exec: %w", err)
			}
			if !inspectResp.Running {
				if inspectResp.ExitCode != 0 {
					return fmt.Errorf("container verification failed with exit code: %d", inspectResp.ExitCode)
				}
				p.logger.Debugf("Container %s verified as ready", containerID)
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (p *ContainerPool) pullImage(ctx context.Context, imageName string) error {
	// Check if image already exists locally
	images, err := p.manager.Cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		p.logger.Warnf("Failed to list images: %v", err)
	} else {
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if tag == imageName || tag == imageName+":latest" {
					p.logger.Debugf("Image %s already exists locally, skipping pull", imageName)
					return nil
				}
			}
		}
	}

	// Image doesn't exist locally, pull it
	reader, err := p.manager.Cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		p.logger.Errorf("Failed to pull image: %v", err)
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	p.logger.Infof("Successfully pulled image: %s", imageName)
	return nil
}

func (p *ContainerPool) createContainer(ctx context.Context, codePath string) (string, error) {
	absPath, err := filepath.Abs(codePath)
	if err != nil {
		p.logger.Errorf("failed to get absolute path: %v", err)
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	p.logger.Debugf("Creating container with code directory: %s", absPath)

	// For Docker-in-Docker, make sure the mount path is absolute and exists on the host
	hostMountPath := absPath
	if !filepath.IsAbs(hostMountPath) {
		hostMountPath, _ = filepath.Abs(hostMountPath)
	}

	// Create a simple keep-alive command that keeps the container running
	keepAliveCommand := `tail -f /dev/null`

	config := &container.Config{
		Image:      p.config.Config.ImageName,
		Cmd:        []string{"sh", "-c", keepAliveCommand},
		Tty:        true,
		WorkingDir: "/code",
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/code:rw", hostMountPath),
			"/var/run/docker.sock:/var/run/docker.sock",
		},
		Resources: container.Resources{
			Memory:   1024 * 1024 * 1024, // 1GB default
			NanoCPUs: 1e9,                // 1 CPU default
		},
		Privileged: true,
	}

	resp, err := p.manager.Cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		p.logger.Errorf("failed to create container: %v", err)
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID
	p.logger.Infof("Container created with ID: %s", containerID)

	// Start the container
	err = p.manager.Cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		p.logger.Errorf("failed to start container: %v", err)
		// Try to cleanup the created container
		if cleanupErr := p.manager.Cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); cleanupErr != nil {
			p.logger.Warnf("Failed to cleanup container after start failure: %v", cleanupErr)
		}
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be running
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		inspect, err := p.manager.Cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return "", fmt.Errorf("failed to inspect container after start: %w", err)
		}

		if inspect.State.Running {
			p.logger.Infof("Container %s is running", containerID)
			return containerID, nil
		}

		p.logger.Debugf("Container %s not running yet (attempt %d/%d), status: %s", containerID, i+1, maxRetries, inspect.State.Status)
		time.Sleep(500 * time.Millisecond)
	}

	return "", fmt.Errorf("container %s failed to start properly", containerID)
}

func (p *ContainerPool) resetContainer(containerID string) error {
	p.logger.Debugf("Resetting container %s", containerID)

	// Instead of full restart, just clean up the code files
	resetScript := p.getResetScript()

	execConfig := &container.ExecOptions{
		Cmd:          []string{"sh", "-c", resetScript},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := p.manager.Cli.ContainerExecCreate(context.Background(), containerID, *execConfig)
	if err != nil {
		return fmt.Errorf("failed to create reset exec: %w", err)
	}

	execAttachResp, err := p.manager.Cli.ContainerExecAttach(context.Background(), execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to reset exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the reset command
	if err := p.manager.Cli.ContainerExecStart(context.Background(), execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start reset exec: %w", err)
	}

	// Wait for reset to complete
	for {
		inspectResp, err := p.manager.Cli.ContainerExecInspect(context.Background(), execResp.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect reset exec: %w", err)
		}
		if !inspectResp.Running {
			if inspectResp.ExitCode != 0 {
				return fmt.Errorf("container reset failed with exit code: %d", inspectResp.ExitCode)
			}
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	p.logger.Debugf("Container %s reset successfully", containerID)
	return nil
}

// getResetScript returns the script to reset a container
func (p *ContainerPool) getResetScript() string {
	switch p.language {
	case types.LanguageGo:
		return `#!/bin/sh
set -e
cd /code
rm -f code.go
echo 'package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}' > code.go
echo "Container reset successfully"
`
	case types.LanguagePy:
		return `#!/bin/sh
set -e
cd /code
rm -f code.py
echo 'print("Hello, World!")' > code.py
echo "Container reset successfully"
`
	case types.LanguageJS, types.LanguageNode:
		return `#!/bin/sh
set -e
cd /code
rm -f code.js
echo 'console.log("Hello, World!");' > code.js
echo "Container reset successfully"
`
	case types.LanguageTS:
		return `#!/bin/sh
set -e
cd /code
rm -f code.ts
echo 'console.log("Hello, World!");' > code.ts
echo "Container reset successfully"
`
	default:
		return `#!/bin/sh
set -e
cd /code
echo "Container reset successfully"
`
	}
}

func (p *ContainerPool) startCleanupRoutine() {
	p.cleanupTicker = time.NewTicker(p.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-p.cleanupTicker.C:
				p.cleanup()
			case <-p.stopCleanup:
				p.cleanupTicker.Stop()
				return
			}
		}
	}()
}

func (p *ContainerPool) startHealthCheckRoutine() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				p.healthCheck()
			case <-p.stopCleanup:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *ContainerPool) cleanup() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	containersToRemove := make([]string, 0)

	// Find containers that have been idle too long
	for id, container := range p.containers {
		if container.Status == types.ContainerStatusReady &&
			now.Sub(container.LastUsed) > p.config.IdleTimeout {
			containersToRemove = append(containersToRemove, id)
		}
	}

	// Remove idle containers (but keep minimum number)
	readyCount := p.getReadyContainerCount()
	for _, id := range containersToRemove {
		if readyCount <= p.config.MinContainers {
			break
		}

		if err := p.manager.CleanupContainer(context.Background(), id); err != nil {
			p.logger.Warnf("Failed to cleanup container %s: %v", id, err)
			continue
		}

		delete(p.containers, id)
		readyCount--
		p.stats.DestroyedCount++
	}

	if len(containersToRemove) > 0 {
		p.logger.Infof("Cleaned up %d idle containers from %s language pool", len(containersToRemove), p.language)
	}

	p.stats.LastCleanup = now
	p.updateStats()
}

func (p *ContainerPool) healthCheck() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for id, container := range p.containers {
		// Check if container is still running
		inspect, err := p.manager.Cli.ContainerInspect(context.Background(), id)
		if err != nil {
			p.logger.Warnf("Container %s health check failed: %v", id, err)
			container.Status = types.ContainerStatusError
			container.Error = err
			continue
		}

		if !inspect.State.Running {
			container.Status = types.ContainerStatusStopped
			container.IsReady = false
		}
	}

	p.updateStats()
}

func (p *ContainerPool) getTotalContainerCount() int {
	return len(p.containers)
}

func (p *ContainerPool) getReadyContainerCount() int {
	count := 0
	for _, container := range p.containers {
		if container.Status == types.ContainerStatusReady && container.IsReady {
			count++
		}
	}
	return count
}

func (p *ContainerPool) updateStats() {
	readyCount := 0
	busyCount := 0
	errorCount := 0

	for _, container := range p.containers {
		switch container.Status {
		case types.ContainerStatusReady:
			readyCount++
		case types.ContainerStatusRunning:
			busyCount++
		case types.ContainerStatusError:
			errorCount++
		}
	}

	p.statsMutex.Lock()
	p.stats.ReadyContainers = readyCount
	p.stats.BusyContainers = busyCount
	p.stats.ErrorContainers = errorCount
	p.stats.TotalContainers = len(p.containers)

	if p.stats.TotalContainers > 0 {
		p.stats.UtilizationRate = float64(busyCount) / float64(p.stats.TotalContainers)
	}
	p.statsMutex.Unlock()
}

func (p *ContainerPool) GetStats() *types.PoolStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *p.stats
	return &stats
}

func (p *ContainerPool) GetLanguage() types.Language {
	return p.language
}

func (p *ContainerPool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Stop cleanup routine
	if p.cleanupTicker != nil {
		close(p.stopCleanup)
	}

	// Cleanup all containers
	for id := range p.containers {
		if err := p.manager.CleanupContainer(context.Background(), id); err != nil {
			p.logger.Warnf("Failed to cleanup container %s: %v", id, err)
		}
	}

	p.containers = make(map[string]*types.PooledContainer)
	p.logger.Infof("Closed %s language pool", p.language)
	return nil
}
