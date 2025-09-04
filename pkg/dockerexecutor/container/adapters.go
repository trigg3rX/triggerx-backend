package container

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/client/docker"
)

// ContainerManagerAdapter adapts ContainerManagerAPI to the consumer-defined ContainerManager interface
type ContainerManagerAdapter struct {
	manager ContainerManagerAPI
}

// NewContainerManagerAdapter creates a new adapter for container manager
func NewContainerManagerAdapter(manager ContainerManagerAPI) *ContainerManagerAdapter {
	return &ContainerManagerAdapter{
		manager: manager,
	}
}

// PullImage implements ContainerManager.PullImage
func (a *ContainerManagerAdapter) PullImage(ctx context.Context, imageName string) error {
	return a.manager.PullImage(ctx, imageName)
}

// CleanupContainer implements ContainerManager.CleanupContainer
func (a *ContainerManagerAdapter) CleanupContainer(ctx context.Context, containerID string) error {
	return a.manager.CleanupContainer(ctx, containerID)
}

// GetDockerClient implements ContainerManager.GetDockerClient
func (a *ContainerManagerAdapter) GetDockerClient() docker.DockerClientAPI {
	return a.manager.GetDockerClient()
}
