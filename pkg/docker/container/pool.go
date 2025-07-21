package container

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type ContainerPool struct {
	containers    map[string]*types.PooledContainer
	mutex         sync.RWMutex
	config        config.PoolConfig
	logger        logging.Logger
	manager       *Manager
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	stats         *types.PoolStats
	statsMutex    sync.RWMutex
	waitQueue     chan struct{}
}

func NewContainerPool(cfg config.PoolConfig, manager *Manager, logger logging.Logger) *ContainerPool {
	pool := &ContainerPool{
		containers:  make(map[string]*types.PooledContainer),
		config:      cfg,
		logger:      logger,
		manager:     manager,
		stopCleanup: make(chan struct{}),
		waitQueue:   make(chan struct{}, cfg.MaxContainers),
		stats: &types.PoolStats{
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
	// p.logger.Infof("Initializing container pool with %d pre-warmed containers", p.config.PreWarmCount)

	// Pre-warm containers
	for i := 0; i < p.config.PreWarmCount; i++ {
		if _, err := p.createPreparedContainer(ctx); err != nil {
			p.logger.Warnf("Failed to create pre-warmed container %d: %v", i, err)
			continue
		}
	}

	p.logger.Infof("Container pool initialized with %d containers", p.getReadyContainerCount())
	return nil
}

func (p *ContainerPool) GetContainer(ctx context.Context) (*types.PooledContainer, error) {
	// p.logger.Debugf("Getting container from pool, total containers: %d", len(p.containers))

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
			return nil, fmt.Errorf("failed to create container: %w", err)
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
		return nil, fmt.Errorf("timeout waiting for container after %v", p.config.MaxWaitTime)
	}
}

func (p *ContainerPool) ReturnContainer(container *types.PooledContainer) error {
	p.logger.Debugf("Returning container %s to pool", container.ID)

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

		p.logger.Debugf("Returned container %s to pool", container.ID)
	} else {
		p.logger.Warnf("Container %s not found in pool", container.ID)
	}

	return nil
}

func (p *ContainerPool) createPreparedContainer(ctx context.Context) (*types.PooledContainer, error) {
	// Create a temporary directory for the container
	tmpDir, err := p.createTempDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create container
	containerID, err := p.manager.CreateContainer(ctx, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Create pooled container
	pooledContainer := &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusReady,
		LastUsed:   time.Now(),
		IsReady:    true,
		WorkingDir: tmpDir,
		ImageName:  p.manager.config.Image,
		CreatedAt:  time.Now(),
	}

	p.mutex.Lock()
	p.containers[containerID] = pooledContainer
	p.stats.CreatedCount++
	p.updateStats()
	p.mutex.Unlock()

	p.logger.Infof("Created prepared container: %s", containerID)
	return pooledContainer, nil
}

func (p *ContainerPool) resetContainer(containerID string) error {
	p.logger.Debugf("Resetting container %s", containerID)

	// Stop the container
	timeout := 10
	// p.logger.Debugf("Stopping container %s", containerID)
	err := p.manager.Cli.ContainerStop(context.Background(), containerID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	// p.logger.Debugf("Container %s stopped successfully", containerID)

	// Start the container again
	// p.logger.Debugf("Starting container %s", containerID)
	err = p.manager.Cli.ContainerStart(context.Background(), containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	// p.logger.Debugf("Container %s start command sent", containerID)

	// Wait for container to be running and verify its state
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		inspect, err := p.manager.Cli.ContainerInspect(context.Background(), containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect container after restart: %w", err)
		}

		// p.logger.Debugf("Container %s inspect attempt %d: status=%s, running=%v", containerID, i+1, inspect.State.Status, inspect.State.Running)

		if inspect.State.Running {
			// p.logger.Debugf("Container %s is running after reset", containerID)
			return nil
		}

		// Wait a bit before retrying
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("container %s failed to start properly after reset", containerID)
}

func (p *ContainerPool) createTempDirectory() (string, error) {
	// Create a temporary directory with proper permissions
	tmpDir := fmt.Sprintf("/tmp/docker-container-%d", time.Now().UnixNano())

	// Create directory with world-readable permissions for Docker-in-Docker
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	return tmpDir, nil
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
		p.logger.Infof("Cleaned up %d idle containers", len(containersToRemove))
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
	p.logger.Info("Container pool closed")
	return nil
}
