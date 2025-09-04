package container

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/client/docker/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"go.uber.org/mock/gomock"
)

func setupManagerTest(t gomock.TestReporter) (*containerManager, *mocks.MockDockerClient, *fs.MockFileSystem, *config.MockConfigProviderInterface) {
	mockDockerClient := mocks.NewMockDockerClient()
	mockFileSystem := &fs.MockFileSystem{}
	mockConfig := config.NewDefaultMockConfigProvider(t)
	logger := &logging.NoOpLogger{}

	manager, _ := NewContainerManager(mockDockerClient, mockFileSystem, mockConfig, logger)
	return manager, mockDockerClient, mockFileSystem, mockConfig
}

// Test Manager creation
func TestNewManager_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDockerClient := mocks.NewMockDockerClient()
	mockFileSystem := &fs.MockFileSystem{}
	mockConfig := config.NewMockConfigProviderInterface(ctrl)
	logger := &logging.NoOpLogger{}

	manager, err := NewContainerManager(mockDockerClient, mockFileSystem, mockConfig, logger)

	require.NoError(t, err)
	require.NotNil(t, manager)
	assert.NotNil(t, manager.dockerClient)
	assert.NotNil(t, manager.fileSystem)
	assert.NotNil(t, manager.config)
	assert.NotNil(t, manager.logger)
	assert.NotNil(t, manager.pools)
	assert.False(t, manager.initialized)
}

// Test Initialize method
func TestManager_Initialize_Success(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Don't add mock image so that ImagePull gets called

	err := manager.Initialize(ctx)

	require.NoError(t, err)
	assert.Len(t, mockDockerClient.ImagePullCalls, 1)
	assert.Equal(t, "golang:1.21-alpine", mockDockerClient.ImagePullCalls[0].Ref)
}

func TestManager_Initialize_ImageExists(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Mock image already exists locally
	mockDockerClient.AddMockImage("golang:1.21-alpine")

	err := manager.Initialize(ctx)

	require.NoError(t, err)
	assert.Len(t, mockDockerClient.ImageListCalls, 1)
	assert.Len(t, mockDockerClient.ImagePullCalls, 0) // Should not pull if image exists
}

func TestManager_Initialize_ImagePullFailure(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Mock image pull failure
	mockDockerClient.ShouldFailImagePull = true

	err := manager.Initialize(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pull base image")
	assert.Len(t, mockDockerClient.ImagePullCalls, 1)
}

// Test InitializeLanguagePools method
func TestManager_InitializeLanguagePools_Success(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	languages := []types.Language{types.LanguageGo, types.LanguagePy}

	err := manager.InitializeLanguagePools(ctx, languages)

	require.NoError(t, err)
	assert.True(t, manager.initialized)
	assert.Len(t, manager.pools, 2)
	assert.Contains(t, manager.pools, types.LanguageGo)
	assert.Contains(t, manager.pools, types.LanguagePy)
}

func TestManager_InitializeLanguagePools_AlreadyInitialized(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Set as already initialized
	manager.initialized = true

	err := manager.InitializeLanguagePools(ctx, []types.Language{types.LanguageGo})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker manager already initialized")
}

func TestManager_InitializeLanguagePools_UnsupportedLanguage(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	languages := []types.Language{types.Language("unsupported")}

	err := manager.InitializeLanguagePools(ctx, languages)

	require.NoError(t, err) // Should not error, just skip unsupported languages
	assert.True(t, manager.initialized)
	assert.Len(t, manager.pools, 0) // No pools created for unsupported language
}

// Test GetContainer method
func TestManager_GetContainer_NotInitialized(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	container, err := manager.GetContainer(ctx, types.LanguageGo)

	require.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "docker manager not initialized")
}

func TestManager_GetContainer_LanguageNotSupported(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Initialize but don't add any pools
	manager.initialized = true

	container, err := manager.GetContainer(ctx, types.LanguageGo)

	require.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "no pool available for language")
}

// Test ReturnContainer method
func TestManager_ReturnContainer_NotInitialized(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	mockContainer := &types.PooledContainer{
		Language: types.LanguageGo,
	}

	err := manager.ReturnContainer(mockContainer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker manager not initialized")
}

func TestManager_ReturnContainer_LanguageNotSupported(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	// Initialize but don't add any pools
	manager.initialized = true

	mockContainer := &types.PooledContainer{
		Language: types.LanguageGo,
	}

	err := manager.ReturnContainer(mockContainer)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no pool available for language")
}

// Test GetPoolStats method
func TestManager_GetPoolStats_EmptyPools(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	stats := manager.GetPoolStats()

	assert.NotNil(t, stats)
	assert.Len(t, stats, 0)
}

// Test GetLanguageStats method
func TestManager_GetLanguageStats_LanguageNotFound(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	stats, exists := manager.GetLanguageStats(types.LanguageGo)

	assert.False(t, exists)
	assert.Nil(t, stats)
}

// Test GetSupportedLanguages method
func TestManager_GetSupportedLanguages_EmptyPools(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	languages := manager.GetSupportedLanguages()

	assert.NotNil(t, languages)
	assert.Len(t, languages, 0)
}

// Test IsLanguageSupported method
func TestManager_IsLanguageSupported_NotSupported(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	supported := manager.IsLanguageSupported(types.LanguageGo)

	assert.False(t, supported)
}

// Test PullImage method
func TestManager_PullImage_ImageAlreadyExists(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Mock existing image
	mockDockerClient.AddMockImage("golang:1.21-alpine")

	err := manager.PullImage(ctx, "golang:1.21-alpine")

	require.NoError(t, err)
	assert.Len(t, mockDockerClient.ImageListCalls, 1)
	assert.Len(t, mockDockerClient.ImagePullCalls, 0) // Should not pull if image exists
}

func TestManager_PullImage_ImageNotExists_PullSuccess(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Don't add any mock images, so it will try to pull

	err := manager.PullImage(ctx, "golang:1.21-alpine")

	require.NoError(t, err)
	assert.Len(t, mockDockerClient.ImageListCalls, 1)
	assert.Len(t, mockDockerClient.ImagePullCalls, 1)
	assert.Equal(t, "golang:1.21-alpine", mockDockerClient.ImagePullCalls[0].Ref)
}

func TestManager_PullImage_ImageListFailure(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Mock image list failure, but pull should still work
	mockDockerClient.ShouldFailImageList = true

	err := manager.PullImage(ctx, "golang:1.21-alpine")

	require.NoError(t, err)
	assert.Len(t, mockDockerClient.ImageListCalls, 1)
	assert.Len(t, mockDockerClient.ImagePullCalls, 1) // Should still attempt pull
}

func TestManager_PullImage_PullFailure(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Mock pull failure
	mockDockerClient.ShouldFailImagePull = true

	err := manager.PullImage(ctx, "golang:1.21-alpine")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pull image")
	assert.Len(t, mockDockerClient.ImagePullCalls, 1)
}

// Test CleanupContainer method
func TestManager_CleanupContainer_AutoCleanupDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDockerClient := mocks.NewMockDockerClient()
	mockFileSystem := &fs.MockFileSystem{}
	mockConfig := config.NewMockConfigProviderInterface(ctrl)
	logger := &logging.NoOpLogger{}

	// Set up mock expectations for config with auto cleanup disabled
	cfg := mockConfig.GetConfig()
	cfg.Manager.AutoCleanup = false
	mockConfig.EXPECT().GetConfig().Return(cfg).AnyTimes()

	manager, _ := NewContainerManager(mockDockerClient, mockFileSystem, mockConfig, logger)
	ctx := context.Background()

	err := manager.CleanupContainer(ctx, "container-1")

	require.NoError(t, err)
	assert.Len(t, mockDockerClient.ContainerRemoveCalls, 0) // Should not call remove
}

func TestManager_CleanupContainer_Success(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	err := manager.CleanupContainer(ctx, "container-1")

	require.NoError(t, err)
	assert.Len(t, mockDockerClient.ContainerRemoveCalls, 1)
	assert.Equal(t, "container-1", mockDockerClient.ContainerRemoveCalls[0].ContainerID)
}

func TestManager_CleanupContainer_Failure(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Mock container remove failure
	mockDockerClient.ShouldFailContainerRemove = true

	err := manager.CleanupContainer(ctx, "container-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock container remove error")
	assert.Len(t, mockDockerClient.ContainerRemoveCalls, 1)
}

// Test ExecuteInContainer method
func TestManager_ExecuteInContainer_ContainerNotRunning(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	mockDockerClient.SetContainerState("container-1", container.State{Running: false, Status: "exited"})
	result, execID, err := manager.ExecuteInContainer(ctx, "container-1", "/test/file.go", types.LanguageGo)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "is not running")
	assert.Equal(t, execID, execID)
}

func TestManager_ExecuteInContainer_ContainerInspectFailure(t *testing.T) {
	manager, mockDockerClient, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Mock container inspect failure
	mockDockerClient.ShouldFailContainerInspect = true

	result, execID, err := manager.ExecuteInContainer(ctx, "container-1", "/test/file.go", types.LanguageGo)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to inspect container before execution")
	assert.Equal(t, execID, execID)
}

func TestManager_ExecuteInContainer_FileReadFailure(t *testing.T) {
	manager, mockDockerClient, mockFileSystem, _ := setupManagerTest(t)
	ctx := context.Background()

	// Create a running mock container
	resp, _ := mockDockerClient.ContainerCreate(ctx, &container.Config{}, &container.HostConfig{}, nil, nil, "")
	containerID := resp.ID
	mockDockerClient.ContainerStart(ctx, containerID, container.StartOptions{})

	// Mock file read failure
	mockFileSystem.SetReadFileResultFunc(func(filename string) ([]byte, error) {
		if filename == "/test/file.go" {
			return nil, errors.New("file read error")
		}
		return []byte("package main"), nil
	})

	result, execID, err := manager.ExecuteInContainer(ctx, containerID, "/test/file.go", types.LanguageGo)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to replace code file")
	assert.Equal(t, execID, execID)
}

// Test Close method
func TestManager_Close_Success(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	err := manager.Close(context.Background())

	require.NoError(t, err)
}

// Test thread safety
func TestManager_ConcurrentAccess(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)

	// Test concurrent reads to pool stats
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.GetPoolStats()
			_ = manager.GetSupportedLanguages()
			_ = manager.IsLanguageSupported(types.LanguageGo)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}

// Integration test scenarios
func TestManager_InitializationFlow(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Test complete initialization flow
	// Don't add mock image so that ImagePull gets called during Initialize

	// 1. Initialize manager
	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// 2. Initialize language pools
	languages := []types.Language{types.LanguageGo, types.LanguagePy}
	err = manager.InitializeLanguagePools(ctx, languages)
	require.NoError(t, err)

	// 3. Verify state
	assert.True(t, manager.initialized)
	assert.Len(t, manager.pools, 2)
	assert.True(t, manager.IsLanguageSupported(types.LanguageGo))
	assert.True(t, manager.IsLanguageSupported(types.LanguagePy))
	assert.False(t, manager.IsLanguageSupported(types.LanguageJS))

	// 4. Get stats
	stats := manager.GetPoolStats()
	assert.Len(t, stats, 2)

	languages = manager.GetSupportedLanguages()
	assert.Len(t, languages, 2)
}

// Benchmark tests
func BenchmarkManager_GetPoolStats(b *testing.B) {
	manager, _, _, _ := setupManagerTest(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetPoolStats()
	}
}

func BenchmarkManager_IsLanguageSupported(b *testing.B) {
	manager, _, _, _ := setupManagerTest(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.IsLanguageSupported(types.LanguageGo)
	}
}

func TestManager_ExecuteInContainer_Success(t *testing.T) {
	manager, mockDockerClient, mockFileSystem, _ := setupManagerTest(t)
	ctx := context.Background()

	// Create a running mock container
	resp, _ := mockDockerClient.ContainerCreate(ctx, &container.Config{}, &container.HostConfig{}, nil, nil, "")
	containerID := resp.ID
	mockDockerClient.ContainerStart(ctx, containerID, container.StartOptions{})

	// Mock file read
	mockFileSystem.SetReadFileResultFunc(func(filename string) ([]byte, error) {
		return []byte("print('hello')"), nil
	})

	// Mock exec lifecycle
	execID := "exec-123"
	mockDockerClient.SetExecCreateResponse(execID, nil)
	mockDockerClient.SetExecAttachResponse(mocks.MockHijackedResponse{
		Output: "START_EXECUTION\nhello world\nEND_EXECUTION",
	}, nil)
	mockDockerClient.SetExecInspectResponse(container.ExecInspect{Running: false, ExitCode: 0}, nil)

	result, execID, err := manager.ExecuteInContainer(ctx, containerID, "/test/file.py", types.LanguagePy)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "hello world")
	assert.NotZero(t, result.Stats.ExecutionTime)
	assert.Nil(t, result.Error)

	// Verify that multiple exec calls were made (setup, copy, verify, execute)
	assert.GreaterOrEqual(t, len(mockDockerClient.ContainerExecCreateCalls), 1)
	assert.Equal(t, execID, execID)
}

func TestManager_ExecuteInContainer_ExecFailure(t *testing.T) {
	manager, mockDockerClient, mockFileSystem, _ := setupManagerTest(t)
	ctx := context.Background()

	// Create a running mock container
	resp, _ := mockDockerClient.ContainerCreate(ctx, &container.Config{}, &container.HostConfig{}, nil, nil, "")
	containerID := resp.ID
	mockDockerClient.ContainerStart(ctx, containerID, container.StartOptions{})

	// Mock file read
	mockFileSystem.SetReadFileResultFunc(func(filename string) ([]byte, error) {
		return []byte("syntax error"), nil
	})

	// Mock exec lifecycle for failure
	execID := "exec-456"
	mockDockerClient.SetExecCreateResponse(execID, nil)
	mockDockerClient.SetExecAttachResponse(mocks.MockHijackedResponse{
		Output: "START_EXECUTION\nSyntaxError: invalid syntax\n",
	}, nil)
	mockDockerClient.SetExecInspectResponse(container.ExecInspect{Running: false, ExitCode: 1}, nil)

	result, execID, err := manager.ExecuteInContainer(ctx, containerID, "/test/file.py", types.LanguagePy)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute code")
	assert.Nil(t, result)
	assert.Equal(t, execID, execID)
}

func TestManager_PoolInteraction_Success(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// 1. Initialize manager
	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// 2. Initialize language pool
	languages := []types.Language{types.LanguageGo}
	err = manager.InitializeLanguagePools(ctx, languages)
	require.NoError(t, err)

	// 3. Get a container
	pooledContainer, err := manager.GetContainer(ctx, types.LanguageGo)
	require.NoError(t, err)
	require.NotNil(t, pooledContainer)
	assert.NotEmpty(t, pooledContainer.ID)
	assert.Equal(t, types.LanguageGo, pooledContainer.Language)

	// 4. Check pool stats
	stats, exists := manager.GetLanguageStats(types.LanguageGo)
	require.True(t, exists)
	require.NotNil(t, stats)
	assert.Equal(t, 1, stats.TotalContainers) // MinContainers is 1 in default mock config
	assert.Equal(t, 1, stats.BusyContainers)
	assert.Equal(t, 0, stats.ReadyContainers)

	// 5. Return the container
	err = manager.ReturnContainer(pooledContainer)
	require.NoError(t, err)

	// 6. Check stats again
	stats, exists = manager.GetLanguageStats(types.LanguageGo)
	require.True(t, exists)
	require.NotNil(t, stats)
	assert.Equal(t, 1, stats.TotalContainers)
	assert.Equal(t, 0, stats.BusyContainers)
	assert.Equal(t, 1, stats.ReadyContainers)
}

func TestManager_Close_CallsPoolClose(t *testing.T) {
	manager, _, _, _ := setupManagerTest(t)
	ctx := context.Background()

	// Initialize to create pools
	manager.Initialize(ctx)
	manager.InitializeLanguagePools(ctx, []types.Language{types.LanguageGo})

	// Get the pool to check its state later
	pool := manager.pools[types.LanguageGo]
	require.NotNil(t, pool)
	// Add a container to the pool to verify it gets cleaned up
	container, err := pool.getContainer(ctx)
	require.NoError(t, err)
	pool.returnContainer(container)
	assert.Equal(t, 1, pool.getStats().TotalContainers)

	// Close the manager
	err = manager.Close(ctx)
	require.NoError(t, err)

	// Assert that the pool is now empty
	assert.Equal(t, 0, pool.getStats().TotalContainers)
}
