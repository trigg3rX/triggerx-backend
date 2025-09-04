package container

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/client/docker/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func setupPoolTest(t *testing.T) (*containerPool, *mocks.MockDockerClient, *config.MockConfigProviderInterface) {
	mockLogger := logging.NewNoOpLogger()
	mockDockerClient := mocks.NewMockDockerClient()
	mockConfigProvider := config.NewDefaultMockConfigProvider(t)

	// Use a real manager, as the pool depends on it, but it's configured with mocks
	manager, err := NewContainerManager(mockDockerClient, nil, mockConfigProvider, mockLogger)
	require.NoError(t, err)

	poolConfig := mockConfigProvider.GetConfig().Languages[string(types.LanguageGo)]
	pool := newContainerPool(poolConfig, manager, mockLogger)

	return pool, mockDockerClient, mockConfigProvider
}

func TestContainerPool_NewContainerPool_Success(t *testing.T) {
	pool, _, _ := setupPoolTest(t)
	assert.NotNil(t, pool)
	assert.Equal(t, types.LanguageGo, pool.language)
	assert.NotNil(t, pool.containers)
	assert.NotNil(t, pool.stats)
	assert.Equal(t, types.LanguageGo, pool.stats.Language)
	assert.Equal(t, 0, pool.stats.TotalContainers)
	assert.NotNil(t, pool.waitQueue)
	assert.Equal(t, 5, cap(pool.waitQueue))
}

// TestContainerPool_Initialize_Success tests successful initialization of a container pool
func TestContainerPool_Initialize_Success(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	// Add mock image to simulate existing image
	mockDockerClient.AddMockImage("golang:1.21")

	// Act
	ctx := context.Background()
	err := pool.initialize(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mockDockerClient.ImageListCalls))
	assert.Equal(t, 2, len(mockDockerClient.ContainerCreateCalls))
}

// TestContainerPool_Initialize_ImagePullFailure tests initialization failure when image pull fails
func TestContainerPool_Initialize_ImagePullFailure(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	mockDockerClient.ShouldFailImagePull = true

	// Act
	ctx := context.Background()
	err := pool.initialize(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pull image")
}

// TestContainerPool_Initialize_ContextCancelled tests initialization when context is cancelled
func TestContainerPool_Initialize_ContextCancelled(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	// Add mock image to simulate existing image
	mockDockerClient.AddMockImage("golang:1.21")

	// Act
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	err := pool.initialize(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

// TestContainerPool_GetContainer_Success tests successful container retrieval
func TestContainerPool_GetContainer_Success(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	// Add a ready container to the pool
	containerID := "test-container-1"
	pooledContainer := &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusReady,
		LastUsed:   time.Now(),
		WorkingDir: "/tmp/test",
		ImageName:  "golang:1.21",
		Language:   types.LanguageGo,
		CreatedAt:  time.Now(),
	}
	pool.containers[containerID] = pooledContainer

	// Configure mock to return running container
	mockDockerClient.MockContainers[containerID] = &mocks.MockContainer{
		ID: containerID,
		State: container.State{
			Status:  "running",
			Running: true,
		},
	}

	// Act
	ctx := context.Background()
	result, err := pool.getContainer(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, containerID, result.ID)
	assert.Equal(t, types.ContainerStatusRunning, result.Status)
}

// TestContainerPool_GetContainer_NoReadyContainers tests container retrieval when no ready containers exist
func TestContainerPool_GetContainer_NoReadyContainers(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	// Add mock image to simulate existing image
	mockDockerClient.AddMockImage("golang:1.21")

	// Act
	ctx := context.Background()
	result, err := pool.getContainer(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, types.ContainerStatusRunning, result.Status)
	assert.Equal(t, 1, len(mockDockerClient.ContainerCreateCalls))
}

// TestContainerPool_GetContainer_MaxContainersReached tests container retrieval when max containers reached
func TestContainerPool_GetContainer_MaxContainersReached(t *testing.T) {
	// Arrange
	pool, _, _ := setupPoolTest(t)

	// Fill the pool to max capacity with busy containers
	for i := 0; i < 2; i++ {
		containerID := fmt.Sprintf("test-container-%d", i)
		pooledContainer := &types.PooledContainer{
			ID:         containerID,
			Status:     types.ContainerStatusRunning,
			LastUsed:   time.Now(),
			WorkingDir: "/tmp/test",
			ImageName:  "golang:1.21",
			Language:   types.LanguageGo,
			CreatedAt:  time.Now(),
		}
		pool.containers[containerID] = pooledContainer
	}

	// Act
	ctx := context.Background()
	result, err := pool.getContainer(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "timeout waiting for container")
}

// TestContainerPool_GetContainer_ContextTimeout tests container retrieval with context timeout
func TestContainerPool_GetContainer_ContextTimeout(t *testing.T) {
	// Arrange
	pool, _, _ := setupPoolTest(t)

	// Fill the pool to max capacity with busy containers
	for i := 0; i < 2; i++ {
		containerID := fmt.Sprintf("test-container-%d", i)
		pooledContainer := &types.PooledContainer{
			ID:         containerID,
			Status:     types.ContainerStatusRunning,
			LastUsed:   time.Now(),
			WorkingDir: "/tmp/test",
			ImageName:  "golang:1.21",
			Language:   types.LanguageGo,
			CreatedAt:  time.Now(),
		}
		pool.containers[containerID] = pooledContainer
	}

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	result, err := pool.getContainer(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "timeout")
}

// TestContainerPool_ReturnContainer_Success tests successful container return
func TestContainerPool_ReturnContainer_Success(t *testing.T) {
	// Arrange
	pool, _, _ := setupPoolTest(t)

	// Add a container to the pool
	containerID := "test-container-1"
	pooledContainer := &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusRunning,
		LastUsed:   time.Now(),
		WorkingDir: "/tmp/test",
		ImageName:  "golang:1.21",
		Language:   types.LanguageGo,
		CreatedAt:  time.Now(),
	}
	pool.containers[containerID] = pooledContainer

	// Act
	err := pool.returnContainer(pooledContainer)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, types.ContainerStatusReady, pooledContainer.Status)
	assert.Nil(t, pooledContainer.Error)
}

// TestContainerPool_ReturnContainer_ContainerNotFound tests container return when container not found in pool
func TestContainerPool_ReturnContainer_ContainerNotFound(t *testing.T) {
	// Arrange
	pool, _, _ := setupPoolTest(t)

	// Create a container that's not in the pool
	pooledContainer := &types.PooledContainer{
		ID:         "non-existent-container",
		Status:     types.ContainerStatusRunning,
		LastUsed:   time.Now(),
		WorkingDir: "/tmp/test",
		ImageName:  "golang:1.21",
		Language:   types.LanguageGo,
		CreatedAt:  time.Now(),
	}

	// Act
	err := pool.returnContainer(pooledContainer)

	// Assert
	assert.NoError(t, err) // Should not error, just log warning
}

// TestContainerPool_ReturnContainer_ResetFailure tests container return when reset fails
func TestContainerPool_ReturnContainer_ResetFailure(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	// Add a container to the pool
	containerID := "test-container-1"
	pooledContainer := &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusRunning,
		LastUsed:   time.Now(),
		WorkingDir: "/tmp/test",
		ImageName:  "golang:1.21",
		Language:   types.LanguageGo,
		CreatedAt:  time.Now(),
	}
	pool.containers[containerID] = pooledContainer

	// Configure mock to fail exec create
	mockDockerClient.ShouldFailContainerExecCreate = true

	// Act
	err := pool.returnContainer(pooledContainer)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, types.ContainerStatusError, pooledContainer.Status)
	assert.NotNil(t, pooledContainer.Error)
}

// TestContainerPool_GetStats_Success tests successful stats retrieval
func TestContainerPool_GetStats_Success(t *testing.T) {
	// Arrange
	pool, _, _ := setupPoolTest(t)

	// Add containers to the pool
	pool.containers["ready-1"] = &types.PooledContainer{
		ID:       "ready-1",
		Status:   types.ContainerStatusReady,
		Language: types.LanguageGo,
	}
	pool.containers["running-1"] = &types.PooledContainer{
		ID:       "running-1",
		Status:   types.ContainerStatusRunning,
		Language: types.LanguageGo,
	}
	pool.containers["error-1"] = &types.PooledContainer{
		ID:       "error-1",
		Status:   types.ContainerStatusError,
		Language: types.LanguageGo,
	}

	// Act
	pool.updateStats()
	stats := pool.getStats()

	// Assert
	assert.NotNil(t, stats)
	assert.Equal(t, types.LanguageGo, stats.Language)
	assert.Equal(t, 3, stats.TotalContainers)
	assert.Equal(t, 1, stats.ReadyContainers)
	assert.Equal(t, 1, stats.BusyContainers)
	assert.Equal(t, 1, stats.ErrorContainers)
	assert.Equal(t, 1.0/3.0, stats.UtilizationRate)
}

// TestContainerPool_Close_Success tests successful pool closure
func TestContainerPool_Close_Success(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	// Add a container to the pool
	containerID := "test-container-1"
	pool.containers[containerID] = &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusReady,
		WorkingDir: "/tmp/test",
		ImageName:  "golang:1.21",
		Language:   types.LanguageGo,
		CreatedAt:  time.Now(),
	}

	// Act
	err := pool.close(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pool.containers))
	assert.Equal(t, 1, len(mockDockerClient.ContainerRemoveCalls))
}

// TestContainerPool_createTempDirectory_Success tests successful temporary directory creation
func TestContainerPool_createTempDirectory_Success(t *testing.T) {
	// Arrange
	pool, _, _ := setupPoolTest(t)

	// Act
	tmpDir, err := pool.createTempDirectory()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, tmpDir)
	assert.Contains(t, tmpDir, "docker-container-go")

	// Cleanup
	err = os.RemoveAll(tmpDir)
	assert.NoError(t, err)
}

// TestContainerPool_ConcurrentAccess tests concurrent access to the container pool
func TestContainerPool_ConcurrentAccess(t *testing.T) {
	// Arrange
	pool, mockDockerClient, _ := setupPoolTest(t)

	// Add mock image
	mockDockerClient.AddMockImage("golang:1.21")

	// Act - Run concurrent operations
	var wg sync.WaitGroup
	results := make([]*types.PooledContainer, 5)
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ctx := context.Background()
			container, err := pool.getContainer(ctx)
			results[index] = container
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// Assert
	successCount := 0
	for i := 0; i < 5; i++ {
		if errors[i] == nil && results[i] != nil {
			successCount++
		}
	}
	assert.Greater(t, successCount, 0, "At least one container should be successfully retrieved")
}

// TestContainerPool_StatsUpdate tests that stats are properly updated
func TestContainerPool_StatsUpdate(t *testing.T) {
	// Arrange
	pool, _, _ := setupPoolTest(t)

	// Add containers with different statuses
	pool.containers["ready-1"] = &types.PooledContainer{
		ID:       "ready-1",
		Status:   types.ContainerStatusReady,
		Language: types.LanguageGo,
	}
	pool.containers["running-1"] = &types.PooledContainer{
		ID:       "running-1",
		Status:   types.ContainerStatusRunning,
		Language: types.LanguageGo,
	}

	// Act
	pool.updateStats()
	stats := pool.getStats()

	// Assert
	assert.Equal(t, 2, stats.TotalContainers)
	assert.Equal(t, 1, stats.ReadyContainers)
	assert.Equal(t, 1, stats.BusyContainers)
	assert.Equal(t, 0.5, stats.UtilizationRate) // 1 busy / 2 total
}

func TestContainerPool_Initialize_PartialFailure(t *testing.T) {
	// Arrange
	mockLogger := logging.NewNoOpLogger()
	mockDockerClient := mocks.NewMockDockerClient()
	mockManager := &containerManager{
		dockerClient: mockDockerClient,
		logger:       mockLogger,
		config:       config.NewDefaultMockConfigProvider(t),
	}
	cfg := config.LanguagePoolConfig{
		BasePoolConfig: config.BasePoolConfig{
			MaxContainers: 5,
			MinContainers: 1,
		},
		DockerConfig: config.DockerContainerConfig{
			Image:          "golang:1.21",
			TimeoutSeconds: 300,
			MemoryLimit:    "1024m",
			CPULimit:       1.0,
			NetworkMode:    "bridge",
		},
		LanguageConfig: config.LanguageConfig{
			Language:  types.LanguageGo,
			ImageName: "golang:1.21",
		},
	}
	pool := newContainerPool(cfg, mockManager, mockLogger)

	// Add mock image to simulate existing image
	mockDockerClient.AddMockImage("golang:1.21")
	// Fail the second container creation
	mockDockerClient.FailContainerCreateAfter = 1

	// Act
	ctx := context.Background()
	err := pool.initialize(ctx)

	// Assert
	assert.NoError(t, err) // Initialization should not fail, just create fewer containers
	assert.Equal(t, 2, len(pool.containers), "Should have successfully created 2 containers")
	// Create calls: 1 success, 1 failure, 1 success
	assert.Equal(t, 3, len(mockDockerClient.ContainerCreateCalls))
}

func TestContainerPool_GetContainer_HealthCheckFails(t *testing.T) {
	// Arrange
	mockLogger := logging.NewNoOpLogger()
	mockDockerClient := mocks.NewMockDockerClient()
	mockManager := &containerManager{
		dockerClient: mockDockerClient,
		logger:       mockLogger,
		config:       config.NewDefaultMockConfigProvider(t),
	}
	cfg := config.LanguagePoolConfig{
		BasePoolConfig: config.BasePoolConfig{
			MaxContainers: 2,
			MinContainers: 1,
		},
		DockerConfig: config.DockerContainerConfig{
			Image:          "golang:1.21",
			TimeoutSeconds: 300,
			MemoryLimit:    "1024m",
			CPULimit:       1.0,
			NetworkMode:    "bridge",
		},
		LanguageConfig: config.LanguageConfig{
			Language:  types.LanguageGo,
			ImageName: "golang:1.21",
		},
	}
	pool := newContainerPool(cfg, mockManager, mockLogger)
	ctx := context.Background()

	// Add a ready but unhealthy (not running) container to the pool
	unhealthyContainerID := "unhealthy-container"
	pool.containers[unhealthyContainerID] = &types.PooledContainer{
		ID:     unhealthyContainerID,
		Status: types.ContainerStatusReady,
	}
	mockDockerClient.SetContainerState(unhealthyContainerID, container.State{Running: false, Status: "exited"})

	// Act
	// The first Get will find the unhealthy one, mark it as error, and then create a new one.
	result, err := pool.getContainer(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// The unhealthy container should have been marked as error
	assert.Equal(t, types.ContainerStatusError, pool.containers[unhealthyContainerID].Status)
	// A new, healthy container should have been created and returned
	assert.NotEqual(t, unhealthyContainerID, result.ID)
	assert.Equal(t, 2, len(pool.containers)) // 1 error, 1 busy
}

func TestContainerPool_ReturnContainer_SignalsWaiter(t *testing.T) {
	// Arrange
	mockLogger := logging.NewNoOpLogger()
	mockDockerClient := mocks.NewMockDockerClient()
	mockManager := &containerManager{
		dockerClient: mockDockerClient,
		logger:       mockLogger,
		config:       config.NewDefaultMockConfigProvider(t),
	}
	cfg := config.LanguagePoolConfig{
		BasePoolConfig: config.BasePoolConfig{
			MaxContainers: 1, // Only one container allowed
			MinContainers: 1,
			MaxWaitTime:   100 * time.Millisecond,
		},
		DockerConfig: config.DockerContainerConfig{
			Image:          "golang:1.21",
			TimeoutSeconds: 300,
			MemoryLimit:    "1024m",
			CPULimit:       1.0,
			NetworkMode:    "bridge",
		},
		LanguageConfig: config.LanguageConfig{
			Language:  types.LanguageGo,
			ImageName: "golang:1.21",
		},
	}
	pool := newContainerPool(cfg, mockManager, mockLogger)
	ctx := context.Background()
	err := pool.initialize(ctx) // Creates one container
	assert.NoError(t, err)

	// Get the only container, making the pool busy
	c1, err := pool.getContainer(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, c1)

	waiterChan := make(chan struct{})
	// Start a goroutine that waits for a container
	go func() {
		_, err := pool.getContainer(ctx)
		assert.NoError(t, err)
		close(waiterChan)
	}()

	// Give the waiter a moment to block
	time.Sleep(20 * time.Millisecond)

	// Act: Return the container. This should unblock the waiter.
	err = pool.returnContainer(c1)
	assert.NoError(t, err)

	// Assert: The waiter goroutine should complete
	select {
	case <-waiterChan:
		// Success
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waiter was not signaled after container was returned")
	}
}

func TestContainerPool_healthCheck_UpdatesUnhealthyContainer(t *testing.T) {
	// Arrange
	mockLogger := logging.NewNoOpLogger()
	mockDockerClient := mocks.NewMockDockerClient()
	mockManager := &containerManager{
		dockerClient: mockDockerClient,
		logger:       mockLogger,
		config:       config.NewDefaultMockConfigProvider(t),
	}
	cfg := config.LanguagePoolConfig{
		BasePoolConfig: config.BasePoolConfig{
			HealthCheckInterval: 5 * time.Minute, // Prevent auto-check during test
		},
		DockerConfig: config.DockerContainerConfig{
			Image:          "golang:1.21",
			TimeoutSeconds: 300,
			MemoryLimit:    "1024m",
			CPULimit:       1.0,
			NetworkMode:    "bridge",
		},
		LanguageConfig: config.LanguageConfig{
			Language:  types.LanguageGo,
			ImageName: "golang:1.21",
		},
	}
	pool := newContainerPool(cfg, mockManager, mockLogger)

	// Add a ready container that is about to be found as stopped
	containerID := "test-container"
	pool.containers[containerID] = &types.PooledContainer{ID: containerID, Status: types.ContainerStatusReady}
	mockDockerClient.SetContainerState(containerID, container.State{Running: false, Status: "exited"})

	// Act
	pool.healthCheck() // Manually trigger health check

	// Assert
	checkedContainer := pool.containers[containerID]
	assert.NotNil(t, checkedContainer)
	assert.Equal(t, types.ContainerStatusStopped, checkedContainer.Status)
}

// BenchmarkContainerPool_GetContainer benchmarks container retrieval performance
func BenchmarkContainerPool_GetContainer(b *testing.B) {
	// Arrange
	mockLogger := logging.NewNoOpLogger()
	mockDockerClient := mocks.NewMockDockerClient()
	mockManager := &containerManager{
		dockerClient: mockDockerClient,
		logger:       mockLogger,
		config:       config.NewDefaultMockConfigProvider(b),
	}
	cfg := config.LanguagePoolConfig{
		BasePoolConfig: config.BasePoolConfig{
			MaxContainers:       10,
			MinContainers:       1,
			MaxWaitTime:         30 * time.Second,
			HealthCheckInterval: 30 * time.Second,
		},
		DockerConfig: config.DockerContainerConfig{
			Image:          "golang:1.21",
			TimeoutSeconds: 300,
			MemoryLimit:    "1024m",
			CPULimit:       1.0,
			NetworkMode:    "bridge",
		},
		LanguageConfig: config.LanguageConfig{
			Language:  types.LanguageGo,
			ImageName: "golang:1.21",
		},
	}
	pool := newContainerPool(cfg, mockManager, mockLogger)

	// Add a ready container
	containerID := "benchmark-container"
	pool.containers[containerID] = &types.PooledContainer{
		ID:         containerID,
		Status:     types.ContainerStatusReady,
		LastUsed:   time.Now(),
		WorkingDir: "/tmp/test",
		ImageName:  "golang:1.21",
		Language:   types.LanguageGo,
		CreatedAt:  time.Now(),
	}

	// Configure mock to return running container
	mockDockerClient.MockContainers[containerID] = &mocks.MockContainer{
		ID: containerID,
		State: container.State{
			Status:  "running",
			Running: true,
		},
	}

	ctx := context.Background()

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container, err := pool.getContainer(ctx)
		if err != nil {
			b.Fatalf("Failed to get container: %v", err)
		}
		if container == nil {
			b.Fatal("Container is nil")
		}
		// Return container for next iteration
		err = pool.returnContainer(container)
		assert.NoError(b, err)
	}
}
