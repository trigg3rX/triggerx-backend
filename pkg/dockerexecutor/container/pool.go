package container

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/pkg/client/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/scripts"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// ContainerManager defines what the container pool needs from a container manager
type ContainerManager interface {
	PullImage(ctx context.Context, imageName string) error
	CleanupContainer(ctx context.Context, containerID string) error
	GetDockerClient() docker.DockerClientAPI
}

type containerPool struct {
	language          types.Language
	containers        map[string]*types.PooledContainer
	mutex             sync.RWMutex
	config            config.LanguagePoolConfig
	logger            observability.Logger
	manager           ContainerManager
	stats             *types.PoolStats
	statsMutex        sync.RWMutex
	waitQueue         chan struct{}
	readyQueue        chan *types.PooledContainer // Channel for ready containers
	stopHealth        chan struct{}               // Channel to stop health check routine
	creationSemaphore chan struct{}               // Semaphore to control container creation
}

func newContainerPool(ctx context.Context, cfg config.LanguagePoolConfig, manager ContainerManager, logger observability.Logger) *containerPool {
	pool := &containerPool{
		language:          cfg.LanguageConfig.Language,
		containers:        make(map[string]*types.PooledContainer),
		config:            cfg,
		logger:            logger,
		manager:           manager,
		waitQueue:         make(chan struct{}, cfg.BasePoolConfig.MaxContainers),
		readyQueue:        make(chan *types.PooledContainer, cfg.BasePoolConfig.MaxContainers), // Buffer for ready containers
		stopHealth:        make(chan struct{}),
		creationSemaphore: make(chan struct{}, cfg.BasePoolConfig.MaxContainers),
		stats: &types.PoolStats{
			Language:          cfg.LanguageConfig.Language,
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

	// Initialize creation semaphore with tokens equal to MaxContainers
	for i := 0; i < cfg.BasePoolConfig.MaxContainers; i++ {
		pool.creationSemaphore <- struct{}{}
	}

	// Start health check routine
	pool.startHealthCheckRoutine(ctx)

	return pool
}

func (p *containerPool) initialize(ctx context.Context) error {
	// p.logger.Debug(ctx, "Initializing language pool", observability.String("language", string(p.language)), observability.Int("pre_warmed_containers", p.config.BasePoolConfig.MinContainers))

	// Pre-warm containers in parallel for faster initialization
	containerChan := make(chan *types.PooledContainer, p.config.BasePoolConfig.MinContainers)
	errorChan := make(chan error, p.config.BasePoolConfig.MinContainers)

	// Start parallel container creation
	for i := 0; i < p.config.BasePoolConfig.MinContainers; i++ {
		go func(index int) {
			container, err := p.createPreparedContainer(ctx)
			if err != nil {
				p.logger.Warn(ctx, "Failed to create pre-warmed container", observability.Int("index", index), observability.String("language", string(p.language)), observability.Error(err))
				errorChan <- err
				return
			}
			containerChan <- container
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < p.config.BasePoolConfig.MinContainers; i++ {
		select {
		case container := <-containerChan:
			successCount++
			p.logger.Debug(ctx, "Successfully created pre-warmed container", observability.Int("index", successCount), observability.String("container_id", container.ID))
		case err := <-errorChan:
			p.logger.Warn(ctx, "Container creation failed", observability.Error(err))
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during container initialization: %w", ctx.Err())
		}
	}

	p.logger.Debug(ctx, "Language pool initialized", observability.String("language", string(p.language)), observability.Int("success_count", successCount))

	// Populate ready queue with existing ready containers
	p.populateReadyQueue(ctx)

	return nil
}

func (p *containerPool) getContainer(ctx context.Context) (*types.PooledContainer, error) {
	timer := time.NewTimer(p.config.BasePoolConfig.MaxWaitTime)
	defer timer.Stop()

	for {
		select {
		// First, try to get an existing container
		case container := <-p.readyQueue:
			// Trust the container is healthy based on health check routine
			// Only update status and return immediately for better performance
			container.Status = types.ContainerStatusRunning
			container.LastUsed = time.Now()
			p.updateStats(ctx)
			return container, nil

		// If no container is ready, wait.
		default:
			// Try to acquire a creation token from the semaphore
			select {
			case <-p.creationSemaphore:
				// We have permission to create a container, proceed immediately
				container, err := p.createPreparedContainer(ctx)
				if err != nil {
					// Return the token back to the semaphore since creation failed
					p.creationSemaphore <- struct{}{}
					return nil, fmt.Errorf("failed to create container for language %s: %w", p.language, err)
				}
				container.Status = types.ContainerStatusRunning
				container.LastUsed = time.Now()
				return container, nil
			default:
				// No creation tokens available, pool is at capacity
				// Wait for a container to be returned
				select {
				case container := <-p.readyQueue:
					// A container was returned. Loop to the top to grab and validate it.
					container.Status = types.ContainerStatusRunning
					container.LastUsed = time.Now()
					p.updateStats(ctx)
					return container, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-timer.C:
					return nil, fmt.Errorf("timeout waiting for container in language pool %s", p.language)
				}
			}
		}
	}
}

func (p *containerPool) returnContainer(ctx context.Context, container *types.PooledContainer) error {
	p.logger.Debug(ctx, "Returning container to language pool", observability.String("container_id", container.ID), observability.String("language", string(p.language)))

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if pooledContainer, exists := p.containers[container.ID]; exists {
		p.logger.Debug(ctx, "Found container in pool, resetting it", observability.String("container_id", container.ID))

		// Reset container for reuse
		if err := p.resetContainer(ctx, container.ID); err != nil {
			p.logger.Warn(ctx, "Failed to reset container", observability.String("container_id", container.ID), observability.Error(err))
			// Mark as error and remove from pool
			pooledContainer.Status = types.ContainerStatusError
			pooledContainer.Error = err
			p.updateStats(ctx)

			// Try to cleanup the failed container
			if cleanupErr := p.manager.CleanupContainer(context.Background(), container.ID); cleanupErr != nil {
				p.logger.Warn(ctx, "Failed to cleanup failed container", observability.String("container_id", container.ID), observability.Error(cleanupErr))
			} else {
				delete(p.containers, container.ID)
				// Release a token back to the semaphore since we removed a container
				select {
				case p.creationSemaphore <- struct{}{}:
				default:
					// Semaphore is full, which shouldn't happen but handle gracefully
					p.logger.Warn(ctx, "Creation semaphore is full when trying to release token for removed container", observability.String("container_id", container.ID))
				}
			}

			return err
		}

		p.logger.Debug(ctx, "Container reset successfully, marking as ready", observability.String("container_id", container.ID))
		pooledContainer.Status = types.ContainerStatusReady
		pooledContainer.Error = nil // Clear any previous errors
		p.updateStats(ctx)

		// Add container to ready queue for immediate availability
		select {
		case p.readyQueue <- pooledContainer:
			p.logger.Debug(ctx, "Added container to ready queue", observability.String("container_id", container.ID))
		default:
			// Ready queue is full, signal via wait queue as fallback
			select {
			case p.waitQueue <- struct{}{}:
			default:
				// Both queues are full, no one is waiting
			}
		}

		p.logger.Debug(ctx, "Returned container to language pool", observability.String("container_id", container.ID), observability.String("language", string(p.language)))
	} else {
		p.logger.Warn(ctx, "Container not found in language pool", observability.String("container_id", container.ID), observability.String("language", string(p.language)))
	}

	return nil
}

func (p *containerPool) createPreparedContainer(ctx context.Context) (*types.PooledContainer, error) {
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
			p.logger.Warn(ctx, "Failed to cleanup container after initialization failure", observability.String("container_id", containerID), observability.Error(cleanupErr))
		}
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	// Create pooled container
	pooledContainer := &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusReady,
		LastUsed:   time.Now(),
		WorkingDir: tmpDir,
		ImageName:  p.config.LanguageConfig.ImageName,
		Language:   p.language,
		CreatedAt:  time.Now(),
	}

	// Add to pool (semaphore guarantees we won't exceed capacity)
	p.mutex.Lock()
	p.containers[containerID] = pooledContainer
	p.stats.CreatedCount++
	p.updateStats(ctx)
	p.mutex.Unlock()

	// Add to ready queue for immediate availability
	select {
	case p.readyQueue <- pooledContainer:
		p.logger.Debug(ctx, "Added newly created container to ready queue", observability.String("container_id", containerID))
	default:
		// Ready queue is full, container will be available on next GetContainer call
		p.logger.Debug(ctx, "Ready queue full, container will be available on next request", observability.String("container_id", containerID))
	}

	p.logger.Debug(ctx, "Created prepared container", observability.String("language", string(p.language)), observability.String("container_id", containerID))
	return pooledContainer, nil
}

func (p *containerPool) createContainer(ctx context.Context, codePath string) (string, error) {
	absPath, err := filepath.Abs(codePath)
	if err != nil {
		p.logger.Error(ctx, "Failed to get absolute path", observability.Error(err))
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	p.logger.Debug(ctx, "Creating container with code directory", observability.String("code_path", absPath))

	// For Docker-in-Docker, make sure the mount path is absolute and exists on the host
	hostMountPath := absPath
	if !filepath.IsAbs(hostMountPath) {
		hostMountPath, _ = filepath.Abs(hostMountPath)
	}

	// Create a simple keep-alive command that keeps the container running
	keepAliveCommand := `tail -f /dev/null`

	// Merge environment variables from DockerConfig and LanguageConfig
	envVars := make([]string, 0, len(p.config.DockerConfig.Environment)+len(p.config.LanguageConfig.Environment))
	envVars = append(envVars, p.config.DockerConfig.Environment...)
	envVars = append(envVars, p.config.LanguageConfig.Environment...)

	config := &container.Config{
		Image:      p.config.LanguageConfig.ImageName,
		Cmd:        []string{"sh", "-c", keepAliveCommand},
		Tty:        true,
		WorkingDir: "/code",
		Env:        envVars,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/code:rw", hostMountPath),
			"/var/run/dockerexecutor.sock:/var/run/dockerexecutor.sock",
		},
		Resources: container.Resources{
			Memory:   int64(types.MemoryLimitBytes(p.config.DockerConfig.MemoryLimit)),
			NanoCPUs: int64(p.config.DockerConfig.CPULimit * 1e9),
		},
		Privileged: true,
	}

	// Generate a meaningful container name
	containerName := p.generateContainerName()

	resp, err := p.manager.GetDockerClient().ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		p.logger.Error(ctx, "Failed to create container", observability.Error(err))
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID
	p.logger.Debug(ctx, "Container created", observability.String("container_id", containerID))

	// Start the container
	err = p.manager.GetDockerClient().ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		p.logger.Error(ctx, "Failed to start container", observability.Error(err))
		// Try to cleanup the created container
		if cleanupErr := p.manager.GetDockerClient().ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); cleanupErr != nil {
			p.logger.Warn(ctx, "Failed to cleanup container after start failure", observability.Error(cleanupErr))
		}
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be running
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		inspect, err := p.manager.GetDockerClient().ContainerInspect(ctx, containerID)
		if err != nil {
			return "", fmt.Errorf("failed to inspect container after start: %w", err)
		}

		if inspect.State.Running {
			p.logger.Debug(ctx, "Container is running", observability.String("container_id", containerID))
			return containerID, nil
		}

		p.logger.Debug(ctx, "Container not running yet", observability.String("container_id", containerID), observability.Int("attempt", i+1), observability.Int("max_retries", maxRetries), observability.String("status", inspect.State.Status))
		time.Sleep(500 * time.Millisecond)
	}

	return "", fmt.Errorf("container %s failed to start properly", containerID)
}

// generateContainerName creates a meaningful name for the container
func (p *containerPool) generateContainerName() string {
	// Use atomic counter to ensure uniqueness even in parallel scenarios
	uuid := uuid.New().String()
	return fmt.Sprintf("triggerx-executor-%s-%s", p.language, uuid[:8])
}

func (p *containerPool) populateReadyQueue(ctx context.Context) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Add all ready containers to the ready queue
	for _, container := range p.containers {
		if container.Status == types.ContainerStatusReady {
			select {
			case p.readyQueue <- container:
				p.logger.Debug(ctx, "Added existing ready container to ready queue", observability.String("container_id", container.ID))
			default:
				// Ready queue is full, stop adding more
				p.logger.Debug(ctx, "Ready queue full, stopping population")
				return
			}
		}
	}
}

func (p *containerPool) createTempDirectory() (string, error) {
	tmpDir := fmt.Sprintf("/tmp/docker-container-%s-%d", p.language, time.Now().UnixNano())
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	return tmpDir, nil
}

// initializeContainer sets up the container with /code folder and basic initialization
func (p *containerPool) initializeContainer(ctx context.Context, containerID string) error {
	p.logger.Debug(ctx, "Initializing container with /code folder setup", observability.String("container_id", containerID))

	// Create /code directory and initialize basic files
	initScript := scripts.GetInitializationScript(p.language)

	execConfig := &container.ExecOptions{
		Cmd:          []string{"sh", "-c", initScript},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := p.manager.GetDockerClient().ContainerExecCreate(ctx, containerID, *execConfig)
	if err != nil {
		return fmt.Errorf("failed to create initialization exec: %w", err)
	}

	execAttachResp, err := p.manager.GetDockerClient().ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to initialization exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the initialization command
	if err := p.manager.GetDockerClient().ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start initialization exec: %w", err)
	}

	// Wait for initialization to complete
	for {
		inspectResp, err := p.manager.GetDockerClient().ContainerExecInspect(ctx, execResp.ID)
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

	p.logger.Debug(ctx, "Container initialized successfully", observability.String("container_id", containerID))

	// Verify container is ready by running a quick test
	if err := p.verifyContainerReady(ctx, containerID); err != nil {
		return fmt.Errorf("container verification failed: %w", err)
	}

	return nil
}

// verifyContainerReady runs a quick test to ensure the container is fully ready
func (p *containerPool) verifyContainerReady(ctx context.Context, containerID string) error {
	p.logger.Debug(ctx, "Verifying container is ready", observability.String("container_id", containerID))

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

	execResp, err := p.manager.GetDockerClient().ContainerExecCreate(ctx, containerID, *execConfig)
	if err != nil {
		return fmt.Errorf("failed to create verification exec: %w", err)
	}

	execAttachResp, err := p.manager.GetDockerClient().ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to verification exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the verification command
	if err := p.manager.GetDockerClient().ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start verification exec: %w", err)
	}

	// Wait for verification to complete with timeout
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("container verification timed out")
		default:
			inspectResp, err := p.manager.GetDockerClient().ContainerExecInspect(ctx, execResp.ID)
			if err != nil {
				return fmt.Errorf("failed to inspect verification exec: %w", err)
			}
			if !inspectResp.Running {
				if inspectResp.ExitCode != 0 {
					return fmt.Errorf("container verification failed with exit code: %d", inspectResp.ExitCode)
				}
				p.logger.Debug(ctx, "Container verified as ready", observability.String("container_id", containerID))
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (p *containerPool) resetContainer(ctx context.Context, containerID string) error {
	p.logger.Debug(ctx, "Resetting container", observability.String("container_id", containerID))

	// Instead of full restart, just clean up the code files
	resetScript := scripts.GetCleanupScript(p.language)

	execConfig := &container.ExecOptions{
		Cmd:          []string{"sh", "-c", resetScript},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := p.manager.GetDockerClient().ContainerExecCreate(context.Background(), containerID, *execConfig)
	if err != nil {
		return fmt.Errorf("failed to create reset exec: %w", err)
	}

	execAttachResp, err := p.manager.GetDockerClient().ContainerExecAttach(context.Background(), execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to reset exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the reset command
	if err := p.manager.GetDockerClient().ContainerExecStart(context.Background(), execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start reset exec: %w", err)
	}

	// Wait for reset to complete
	for {
		inspectResp, err := p.manager.GetDockerClient().ContainerExecInspect(context.Background(), execResp.ID)
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

	p.logger.Debug(ctx, "Container reset successfully", observability.String("container_id", containerID))
	return nil
}

func (p *containerPool) startHealthCheckRoutine(ctx context.Context) {
	ticker := time.NewTicker(p.config.BasePoolConfig.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				p.healthCheck(ctx)
			case <-p.stopHealth:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *containerPool) healthCheck(ctx context.Context) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	containersChecked := 0
	containersWithIssues := 0
	containersToRemove := make([]string, 0)

	for id, container := range p.containers {
		containersChecked++

		// Check if container is still running
		inspect, err := p.manager.GetDockerClient().ContainerInspect(context.Background(), id)
		if err != nil {
			p.logger.Warn(ctx, "Container health check failed", observability.String("container_id", id), observability.Error(err))
			container.Status = types.ContainerStatusError
			container.Error = err
			container.Status = types.ContainerStatusError
			containersWithIssues++
			containersToRemove = append(containersToRemove, id)
			continue
		}

		if !inspect.State.Running {
			p.logger.Warn(ctx, "Container is not running, marking for removal", observability.String("container_id", id))
			container.Status = types.ContainerStatusStopped
			container.Status = types.ContainerStatusError
			containersWithIssues++
			containersToRemove = append(containersToRemove, id)
			continue
		}

		// Container is running, ensure it's marked as ready if it was in error state
		if container.Status == types.ContainerStatusError {
			p.logger.Debug(ctx, "Container recovered from error state", observability.String("container_id", id))
			container.Status = types.ContainerStatusReady
			container.Error = nil
			container.Status = types.ContainerStatusReady
		}
	}

	// Remove unhealthy containers from the pool
	for _, id := range containersToRemove {
		p.logger.Debug(ctx, "Removing unhealthy container from pool", observability.String("container_id", id))
		delete(p.containers, id)

		// Release a token back to the semaphore since we removed a container
		select {
		case p.creationSemaphore <- struct{}{}:
		default:
			// Semaphore is full, which shouldn't happen but handle gracefully
			p.logger.Warn(ctx, "Creation semaphore is full when trying to release token for unhealthy container", observability.String("container_id", id))
		}

		// Try to cleanup the container
		if err := p.manager.CleanupContainer(context.Background(), id); err != nil {
			p.logger.Warn(ctx, "Failed to cleanup unhealthy container", observability.String("container_id", id), observability.Error(err))
		}
	}

	if containersChecked > 0 {
		totalContainers := len(p.containers)
		checkPercentage := float64(containersChecked) / float64(totalContainers+len(containersToRemove)) * 100
		p.logger.Debug(ctx, "Health check completed for pool", observability.String("language", string(p.language)), observability.Int("containers_checked", containersChecked), observability.Float64("check_percentage", checkPercentage), observability.Int("containers_with_issues", containersWithIssues), observability.Int("containers_removed", len(containersToRemove)))
	}

	p.updateStats(ctx)
}

// getHealthCheckStats returns statistics about the last health check
func (p *containerPool) getHealthCheckStats() (int, int, int) {
	totalContainers := len(p.containers)
	containersToCheck := 0
	containersInError := 0

	for _, container := range p.containers {
		shouldCheck := container.Status == types.ContainerStatusError ||
			container.Status == types.ContainerStatusStopped

		if shouldCheck {
			containersToCheck++
		}

		if container.Status == types.ContainerStatusError {
			containersInError++
		}
	}

	return totalContainers, containersToCheck, containersInError
}

// markContainerAsFailed marks a container as failed and removes it from the pool
// This should be called when a container fails during command execution
func (p *containerPool) markContainerAsFailed(ctx context.Context, containerID string, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if container, exists := p.containers[containerID]; exists {
		p.logger.Warn(ctx, "Marking container as failed due to execution error", observability.String("container_id", containerID), observability.Error(err))
		container.Status = types.ContainerStatusError
		container.Error = err
		container.Status = types.ContainerStatusError

		// Remove from pool immediately
		delete(p.containers, containerID)
		p.updateStats(ctx)

		// Release a token back to the semaphore since we removed a container
		select {
		case p.creationSemaphore <- struct{}{}:
		default:
			// Semaphore is full, which shouldn't happen but handle gracefully
			p.logger.Warn(ctx, "Creation semaphore is full when trying to release token for failed container", observability.String("container_id", containerID))
		}

		// Cleanup the failed container
		go func() {
			if cleanupErr := p.manager.CleanupContainer(context.Background(), containerID); cleanupErr != nil {
				p.logger.Warn(ctx, "Failed to cleanup failed container", observability.String("container_id", containerID), observability.Error(cleanupErr))
			}
		}()
	}
}

func (p *containerPool) updateStats(ctx context.Context) {
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

func (p *containerPool) getStats() *types.PoolStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *p.stats
	return &stats
}

func (p *containerPool) close(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Stop health check routine
	close(p.stopHealth)

	// Cleanup all containers
	for id := range p.containers {
		if err := p.manager.CleanupContainer(ctx, id); err != nil {
			p.logger.Warn(ctx, "Failed to cleanup container", observability.String("container_id", id), observability.Error(err))
		}
	}

	p.containers = make(map[string]*types.PooledContainer)
	p.logger.Debug(ctx, "Closed language pool", observability.String("language", string(p.language)))
	return nil
}
