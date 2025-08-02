package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type ContainerLifecycle struct {
	manager *Manager
	logger  logging.Logger
	config  config.DockerConfig
}

func NewContainerLifecycle(manager *Manager, logger logging.Logger) *ContainerLifecycle {
	return &ContainerLifecycle{
		manager: manager,
		logger:  logger,
		config:  manager.config,
	}
}

func (cl *ContainerLifecycle) CreateAndStart(ctx context.Context, config *types.ContainerConfig) (*types.ContainerInfo, error) {
	cl.logger.Infof("Creating container with image: %s", config.Image)

	// Create container
	containerID, err := cl.manager.CreateContainer(ctx, config.WorkingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := cl.StartContainer(ctx, containerID); err != nil {
		// Cleanup failed container
		err := cl.manager.CleanupContainer(ctx, containerID)
		if err != nil {
			cl.logger.Errorf("Failed to cleanup container: %v", err)
		}
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get container info
	info, err := cl.GetContainerInfo(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container info: %w", err)
	}

	cl.logger.Infof("Container created and started: %s", containerID)
	return info, nil
}

func (cl *ContainerLifecycle) StartContainer(ctx context.Context, containerID string) error {
	cl.logger.Debugf("Starting container: %s", containerID)

	err := cl.manager.Cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be running
	if err := cl.waitForContainerRunning(ctx, containerID); err != nil {
		return fmt.Errorf("container failed to start: %w", err)
	}

	cl.logger.Debugf("Container started successfully: %s", containerID)
	return nil
}

func (cl *ContainerLifecycle) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	cl.logger.Debugf("Stopping container: %s", containerID)

	timeoutSeconds := int(timeout.Seconds())
	err := cl.manager.Cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	cl.logger.Debugf("Container stopped successfully: %s", containerID)
	return nil
}

func (cl *ContainerLifecycle) RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	cl.logger.Debugf("Restarting container: %s", containerID)

	timeoutSeconds := int(timeout.Seconds())
	err := cl.manager.Cli.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
	if err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}

	// Wait for container to be running
	if err := cl.waitForContainerRunning(ctx, containerID); err != nil {
		return fmt.Errorf("container failed to restart: %w", err)
	}

	cl.logger.Debugf("Container restarted successfully: %s", containerID)
	return nil
}

func (cl *ContainerLifecycle) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	cl.logger.Debugf("Removing container: %s (force: %v)", containerID, force)

	err := cl.manager.Cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	cl.logger.Debugf("Container removed successfully: %s", containerID)
	return nil
}

func (cl *ContainerLifecycle) GetContainerInfo(ctx context.Context, containerID string) (*types.ContainerInfo, error) {
	info, err := cl.manager.GetContainerInfo(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container info: %w", err)
	}

	// Convert to our types
	containerInfo := &types.ContainerInfo{
		ID:      containerID,
		Created: time.Now(), // Use current time as fallback
		Config: &types.ContainerConfig{
			Image:          info.Config.Image,
			WorkingDir:     info.Config.WorkingDir,
			Environment:    info.Config.Env,
			Binds:          info.HostConfig.Binds,
			Resources:      info.HostConfig.Resources,
			Privileged:     info.HostConfig.Privileged,
			NetworkMode:    string(info.HostConfig.NetworkMode),
			SecurityOpt:    info.HostConfig.SecurityOpt,
			ReadOnlyRootFS: info.HostConfig.ReadonlyRootfs,
		},
	}

	return containerInfo, nil
}

func (cl *ContainerLifecycle) GetContainerStats(ctx context.Context, containerID string) (*types.DockerResourceStats, error) {
	stats, err := cl.manager.Cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer func() {
		err := stats.Body.Close()
		if err != nil {
			cl.logger.Errorf("Failed to close container stats body: %v", err)
		}
	}()

	// Docker stats stream JSON objects; decode the first one
	var dockerStats container.StatsResponse
	dec := json.NewDecoder(stats.Body)
	if err := dec.Decode(&dockerStats); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to decode container stats: %w", err)
	}

	// Calculate CPU usage percent (see Docker's formula)
	var cpuPercent float64
	cpuDelta := float64(dockerStats.CPUStats.CPUUsage.TotalUsage - dockerStats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(dockerStats.CPUStats.SystemUsage - dockerStats.PreCPUStats.SystemUsage)
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(dockerStats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	// Network stats (sum all interfaces)
	var rxBytes, txBytes uint64
	if dockerStats.Networks != nil {
		for _, v := range dockerStats.Networks {
			rxBytes += v.RxBytes
			txBytes += v.TxBytes
		}
	}

	resourceStats := &types.DockerResourceStats{
		MemoryUsage:   dockerStats.MemoryStats.Usage,
		CPUPercentage: cpuPercent,
		NetworkRx:     rxBytes,
		NetworkTx:     txBytes,
		BlockRead:     sumBlkioServiceBytes(dockerStats.BlkioStats.IoServiceBytesRecursive, "Read"),
		BlockWrite:    sumBlkioServiceBytes(dockerStats.BlkioStats.IoServiceBytesRecursive, "Write"),
		RxBytes:       rxBytes,
		TxBytes:       txBytes,
		ExecutionTime: time.Duration(dockerStats.Read.UnixNano()), // This is the stats timestamp; adjust as needed
	}

	return resourceStats, nil
}

func (cl *ContainerLifecycle) IsContainerRunning(ctx context.Context, containerID string) (bool, error) {
	info, err := cl.manager.GetContainerInfo(ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("failed to get container info: %w", err)
	}

	return info.State.Running, nil
}

func (cl *ContainerLifecycle) WaitForContainer(ctx context.Context, containerID string, condition container.WaitCondition) error {
	cl.logger.Debugf("Waiting for container: %s (condition: %s)", containerID, condition)

	statusCh, errCh := cl.manager.Cli.ContainerWait(ctx, containerID, condition)

	select {
	case err := <-errCh:
		return fmt.Errorf("container wait error: %w", err)
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with status %d", status.StatusCode)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for container: %w", ctx.Err())
	}
}

func (cl *ContainerLifecycle) waitForContainerRunning(ctx context.Context, containerID string) error {
	timeout := 30 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		running, err := cl.IsContainerRunning(ctx, containerID)
		if err != nil {
			cl.logger.Warnf("Failed to check container status: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if running {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("container failed to start within %v", timeout)
}

func (cl *ContainerLifecycle) HealthCheck(ctx context.Context, containerID string) error {
	cl.logger.Debugf("Performing health check on container: %s", containerID)

	// Check if container is running
	running, err := cl.IsContainerRunning(ctx, containerID)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !running {
		return fmt.Errorf("container is not running")
	}

	// Get container stats to check resource usage
	stats, err := cl.GetContainerStats(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container stats: %w", err)
	}

	// Check memory usage
	memoryLimit := cl.config.MemoryLimitBytes()
	if stats.MemoryUsage > uint64(0.9*float64(memoryLimit)) {
		cl.logger.Warnf("Container %s memory usage is high: %d/%d bytes",
			containerID, stats.MemoryUsage, memoryLimit)
	}

	// Check CPU usage
	if stats.CPUPercentage > 90.0 {
		cl.logger.Warnf("Container %s CPU usage is high: %.2f%%",
			containerID, stats.CPUPercentage)
	}

	cl.logger.Debugf("Health check passed for container: %s", containerID)
	return nil
}

func (cl *ContainerLifecycle) CleanupContainer(ctx context.Context, containerID string) error {
	cl.logger.Debugf("Cleaning up container: %s", containerID)

	// Stop container if running
	running, err := cl.IsContainerRunning(ctx, containerID)
	if err != nil {
		cl.logger.Warnf("Failed to check if container is running: %v", err)
	} else if running {
		if err := cl.StopContainer(ctx, containerID, 10*time.Second); err != nil {
			cl.logger.Warnf("Failed to stop container: %v", err)
		}
	}

	// Remove container
	if err := cl.RemoveContainer(ctx, containerID, true); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	cl.logger.Debugf("Container cleanup completed: %s", containerID)
	return nil
}

// Helper to sum block I/O stats for a given op ("Read" or "Write")
func sumBlkioServiceBytes(entries []container.BlkioStatEntry, op string) uint64 {
	var sum uint64
	for _, entry := range entries {
		if entry.Op == op {
			sum += entry.Value
		}
	}
	return sum
}
