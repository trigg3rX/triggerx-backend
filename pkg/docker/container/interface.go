// file: container/interfaces.go
package container

//go:generate mockgen -source=interface.go -destination=mock_manager.go -package=container

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/client/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

// poolAPI defines the internal interface for a container pool.
// This allows the manager to be tested independently of the pool implementation.
type poolAPI interface {
	initialize(ctx context.Context) error
	getContainer(ctx context.Context) (*types.PooledContainer, error)
	returnContainer(container *types.PooledContainer) error
	getStats() *types.PoolStats
	getHealthCheckStats() (total, toCheck, inError int)
	markContainerAsFailed(containerID string, err error)
	close(ctx context.Context) error
}

// ContainerManagerAPI is the primary interface for interacting with the container management system.
// NOTE: CreateContainer has been removed as it's now an internal detail of the pools.
type ContainerManagerAPI interface {
	Initialize(ctx context.Context) error
	InitializeLanguagePools(ctx context.Context, languages []types.Language) error
	GetDockerClient() docker.DockerClientAPI
	GetContainer(ctx context.Context, language types.Language) (*types.PooledContainer, error)
	ReturnContainer(container *types.PooledContainer) error
	GetPoolStats() map[types.Language]*types.PoolStats
	GetLanguageStats(language types.Language) (*types.PoolStats, bool)
	GetHealthCheckStats() map[types.Language]map[string]int
	GetSupportedLanguages() []types.Language
	IsLanguageSupported(language types.Language) bool
	ExecuteInContainer(ctx context.Context, containerID string, filePath string, language types.Language) (*types.ExecutionResult, string, error)
	PullImage(ctx context.Context, imageName string) error
	CleanupContainer(ctx context.Context, containerID string) error
	KillExecProcess(ctx context.Context, execID string) error
	MarkContainerAsFailed(containerID string, language types.Language, err error)
	Close(ctx context.Context) error
}

// Ensure ContainerManager implements the interface at compile time.
var _ ContainerManagerAPI = (*containerManager)(nil)

// Ensure containerPool implements the internal interface at compile time.
var _ poolAPI = (*containerPool)(nil)
