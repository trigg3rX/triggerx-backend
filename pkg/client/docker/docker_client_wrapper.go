package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	cerrdefs "github.com/containerd/errdefs"
)

// DockerClientWrapper wraps the real Docker client to implement the DockerClientAPI interface
type DockerClientWrapper struct {
	client *client.Client
}

// NewDockerClientWrapper creates a new wrapper around the Docker client
func NewDockerClientWrapper(cli *client.Client) DockerClientAPI {
	return &DockerClientWrapper{client: cli}
}

// Image operations
func (w *DockerClientWrapper) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	return w.client.ImageList(ctx, options)
}

func (w *DockerClientWrapper) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	return w.client.ImagePull(ctx, ref, options)
}

func (w *DockerClientWrapper) ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	return w.client.ImageRemove(ctx, imageID, options)
}

// Container operations
func (w *DockerClientWrapper) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	return w.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func (w *DockerClientWrapper) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return w.client.ContainerStart(ctx, containerID, options)
}

func (w *DockerClientWrapper) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return w.client.ContainerStop(ctx, containerID, options)
}

func (w *DockerClientWrapper) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	return w.client.ContainerRestart(ctx, containerID, options)
}

func (w *DockerClientWrapper) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return w.client.ContainerInspect(ctx, containerID)
}

func (w *DockerClientWrapper) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return w.client.ContainerRemove(ctx, containerID, options)
}

func (w *DockerClientWrapper) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	return w.client.ContainerStats(ctx, containerID, stream)
}

func (w *DockerClientWrapper) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	return w.client.ContainerWait(ctx, containerID, condition)
}

// Container execution operations
func (w *DockerClientWrapper) ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	return w.client.ContainerExecCreate(ctx, containerID, config)
}

func (w *DockerClientWrapper) ContainerExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error {
	return w.client.ContainerExecStart(ctx, execID, config)
}

func (w *DockerClientWrapper) ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	return w.client.ContainerExecInspect(ctx, execID)
}

func (w *DockerClientWrapper) ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error) {
	return w.client.ContainerExecAttach(ctx, execID, config)
}

func (w *DockerClientWrapper) CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error {
	return w.client.CopyToContainer(ctx, containerID, dstPath, content, options)
}

func (w *DockerClientWrapper) CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error) {
	return w.client.CopyFromContainer(ctx, containerID, srcPath)
}

func (w *DockerClientWrapper) IsErrNotFound(err error) bool {
	return cerrdefs.IsNotFound(err)
}
