package docker

//go:generate mockgen -source=interface.go -destination=mocks/mock_docker_client.go -package=mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// DockerClientAPI defines the interface for the Docker client.
type DockerClientAPI interface {
	// ContainerCreate creates a new container with the given configuration.
	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)

	// ContainerStart starts a container with the given ID.
	ContainerStart(
		ctx context.Context,
		containerID string,
		options container.StartOptions,
	) error

	// ContainerStop stops a container with the given ID.
	ContainerStop(
		ctx context.Context,
		containerID string,
		options container.StopOptions,
	) error

	// ContainerRestart restarts a container with the given ID.
	ContainerRestart(
		ctx context.Context,
		containerID string,
		options container.StopOptions,
	) error

	// ContainerRemove removes a container with the given ID.
	ContainerRemove(
		ctx context.Context,
		containerID string,
		options container.RemoveOptions,
	) error

	// ContainerInspect inspects a container with the given ID.
	ContainerInspect(
		ctx context.Context,
		containerID string,
	) (container.InspectResponse, error)

	// ContainerStats gets the stats of a container with the given ID.
	ContainerStats(
		ctx context.Context,
		containerID string,
		stream bool,
	) (container.StatsResponseReader, error)

	// ContainerWait waits for a container with the given ID.
	ContainerWait(
		ctx context.Context,
		containerID string,
		condition container.WaitCondition,
	) (<-chan container.WaitResponse, <-chan error)

	// ContainerExecCreate creates a new exec instance for a container.
	ContainerExecCreate(
		ctx context.Context,
		container string,
		config container.ExecOptions,
	) (container.ExecCreateResponse, error)

	// ContainerExecAttach attaches to an exec instance.
	ContainerExecAttach(
		ctx context.Context,
		execID string,
		config container.ExecAttachOptions,
	) (types.HijackedResponse, error)

	// ContainerExecStart starts an exec instance.
	ContainerExecStart(
		ctx context.Context,
		execID string,
		config container.ExecStartOptions,
	) error

	// ContainerExecInspect inspects an exec instance.
	ContainerExecInspect(
		ctx context.Context,
		execID string,
	) (container.ExecInspect, error)

	// ImagePull pulls an image with the given reference.
	ImagePull(
		ctx context.Context,
		refStr string,
		options image.PullOptions,
	) (io.ReadCloser, error)

	// ImageList lists all images.
	ImageList(
		ctx context.Context,
		options image.ListOptions,
	) ([]image.Summary, error)

	// ImageRemove removes an image with the given ID.
	ImageRemove(
		ctx context.Context,
		imageID string,
		options image.RemoveOptions,
	) ([]image.DeleteResponse, error)

	// CopyToContainer copies files to a container.
	CopyToContainer(
		ctx context.Context,
		containerID string,
		dstPath string,
		content io.Reader,
		options container.CopyToContainerOptions,
	) error

	// CopyFromContainer copies files from a container.
	CopyFromContainer(
		ctx context.Context,
		containerID string,
		srcPath string,
	) (io.ReadCloser, container.PathStat, error)
}
