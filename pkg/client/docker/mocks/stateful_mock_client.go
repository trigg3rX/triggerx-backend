package mocks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// mockConn is a safe mock connection that implements net.Conn
type mockConn struct{}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, io.EOF }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// MockDockerClient is a mock implementation of DockerClientAPI for testing
type MockDockerClient struct {
	// Mutex for thread safety
	mutex sync.RWMutex

	// Control behavior
	ShouldFailImageList            bool
	ShouldFailImagePull            bool
	ShouldFailImageRemove          bool
	ShouldFailContainerCreate      bool
	ShouldFailContainerStart       bool
	ShouldFailContainerStop        bool
	ShouldFailContainerRestart     bool
	ShouldFailContainerInspect     bool
	ShouldFailContainerRemove      bool
	ShouldFailContainerStats       bool
	ShouldFailContainerWait        bool
	ShouldFailContainerExecCreate  bool
	ShouldFailContainerExecStart   bool
	ShouldFailContainerExecInspect bool
	ShouldFailContainerExecAttach  bool
	ShouldFailCopyToContainer      bool
	ShouldFailCopyFromContainer    bool
	FailContainerCreateAfter       int

	// Track method calls
	ImageListCalls            []ImageListCall
	ImagePullCalls            []ImagePullCall
	ImageRemoveCalls          []ImageRemoveCall
	ContainerCreateCalls      []ContainerCreateCall
	ContainerStartCalls       []ContainerStartCall
	ContainerStopCalls        []ContainerStopCall
	ContainerRestartCalls     []ContainerRestartCall
	ContainerInspectCalls     []ContainerInspectCall
	ContainerRemoveCalls      []ContainerRemoveCall
	ContainerStatsCalls       []ContainerStatsCall
	ContainerWaitCalls        []ContainerWaitCall
	ContainerExecCreateCalls  []ContainerExecCreateCall
	ContainerExecStartCalls   []ContainerExecStartCall
	ContainerExecInspectCalls []ContainerExecInspectCall
	ContainerExecAttachCalls  []ContainerExecAttachCall
	CopyToContainerCalls      []CopyToContainerCall
	CopyFromContainerCalls    []CopyFromContainerCall

	// Container creation counter for conditional failures
	containerCreateCount int

	// Mock data
	MockImages     []image.Summary
	MockContainers map[string]*MockContainer
	MockExecs      map[string]*MockExec

	// Mock exec responses for testing
	MockExecCreateResponses  map[string]error
	MockExecAttachResponses  map[string]MockHijackedResponse
	MockExecInspectResponses map[string]container.ExecInspect

	// ID counters for generating unique IDs
	nextContainerID int
	nextExecID      int
}

type MockOption func(*MockDockerClient)

// Call tracking structs
type ImageListCall struct {
	Ctx     context.Context
	Options image.ListOptions
}

type ImagePullCall struct {
	Ctx     context.Context
	Ref     string
	Options image.PullOptions
}

type ImageRemoveCall struct {
	Ctx     context.Context
	ImageID string
	Options image.RemoveOptions
}

type ContainerCreateCall struct {
	Ctx              context.Context
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
	Platform         *ocispec.Platform
	ContainerName    string
}

type ContainerStartCall struct {
	Ctx         context.Context
	ContainerID string
	Options     container.StartOptions
}

type ContainerStopCall struct {
	Ctx         context.Context
	ContainerID string
	Options     container.StopOptions
}

type ContainerRestartCall struct {
	Ctx         context.Context
	ContainerID string
	Options     container.StopOptions
}

type ContainerInspectCall struct {
	Ctx         context.Context
	ContainerID string
}

type ContainerStatsCall struct {
	Ctx         context.Context
	ContainerID string
	Stream      bool
}

type ContainerWaitCall struct {
	Ctx         context.Context
	ContainerID string
	Condition   container.WaitCondition
}

type ContainerRemoveCall struct {
	Ctx         context.Context
	ContainerID string
	Options     container.RemoveOptions
}

type ContainerExecCreateCall struct {
	Ctx         context.Context
	ContainerID string
	Config      container.ExecOptions
}

type ContainerExecStartCall struct {
	Ctx    context.Context
	ExecID string
	Config container.ExecStartOptions
}

type ContainerExecInspectCall struct {
	Ctx    context.Context
	ExecID string
}

type ContainerExecAttachCall struct {
	Ctx    context.Context
	ExecID string
	Config container.ExecAttachOptions
}

type CopyToContainerCall struct {
	Ctx         context.Context
	ContainerID string
	DstPath     string
	Content     io.Reader
	Options     container.CopyToContainerOptions
}

type CopyFromContainerCall struct {
	Ctx         context.Context
	ContainerID string
	SrcPath     string
}

// Mock data structures
type MockContainer struct {
	ID         string
	State      container.State
	Config     *container.Config
	HostConfig *container.HostConfig
}

type MockExec struct {
	ID          string
	ContainerID string
	Running     bool
	ExitCode    int
	Config      container.ExecOptions
}

// MockHijackedResponse represents a mock hijacked response for testing
type MockHijackedResponse struct {
	Output string
	Error  error
}

// Create option functions
func WithFailingContainerStart() MockOption {
    return func(m *MockDockerClient) {
        m.ShouldFailContainerStart = true
    }
}

func WithExecInspectResponse(execID string, resp container.ExecInspect) MockOption {
    return func(m *MockDockerClient) {
        m.MockExecInspectResponses[execID] = resp
    }
}

// NewMockDockerClient creates a new mock Docker client
func NewMockDockerClient(opts ...MockOption) *MockDockerClient {
    mock := &MockDockerClient{
		MockContainers:           make(map[string]*MockContainer),
		MockExecs:                make(map[string]*MockExec),
		MockExecCreateResponses:  make(map[string]error),
		MockExecAttachResponses:  make(map[string]MockHijackedResponse),
		MockExecInspectResponses: make(map[string]container.ExecInspect),
		nextContainerID:          1,
		nextExecID:               1,
	}
	for _, opt := range opts {
        opt(mock)
    }
    return mock
}

// Reset clears all call tracking and mock data
func (m *MockDockerClient) Reset() {
	m.ImageListCalls = nil
	m.ImagePullCalls = nil
	m.ImageRemoveCalls = nil
	m.ContainerCreateCalls = nil
	m.ContainerStartCalls = nil
	m.ContainerStopCalls = nil
	m.ContainerRestartCalls = nil
	m.ContainerInspectCalls = nil
	m.ContainerRemoveCalls = nil
	m.ContainerStatsCalls = nil
	m.ContainerWaitCalls = nil
	m.ContainerExecCreateCalls = nil
	m.ContainerExecStartCalls = nil
	m.ContainerExecInspectCalls = nil
	m.ContainerExecAttachCalls = nil
	m.CopyToContainerCalls = nil
	m.CopyFromContainerCalls = nil

	m.MockImages = nil
	m.MockContainers = make(map[string]*MockContainer)
	m.MockExecs = make(map[string]*MockExec)
	m.MockExecCreateResponses = make(map[string]error)
	m.MockExecAttachResponses = make(map[string]MockHijackedResponse)
	m.MockExecInspectResponses = make(map[string]container.ExecInspect)
	m.nextContainerID = 1
	m.nextExecID = 1
	m.containerCreateCount = 0

	// Reset failure flags
	m.ShouldFailImageList = false
	m.ShouldFailImagePull = false
	m.ShouldFailImageRemove = false
	m.ShouldFailContainerCreate = false
	m.ShouldFailContainerStart = false
	m.ShouldFailContainerStop = false
	m.ShouldFailContainerRestart = false
	m.ShouldFailContainerInspect = false
	m.ShouldFailContainerRemove = false
	m.ShouldFailContainerStats = false
	m.ShouldFailContainerWait = false
	m.ShouldFailContainerExecCreate = false
	m.ShouldFailContainerExecStart = false
	m.ShouldFailContainerExecInspect = false
	m.ShouldFailContainerExecAttach = false
	m.ShouldFailCopyToContainer = false
	m.ShouldFailCopyFromContainer = false
	m.FailContainerCreateAfter = 0
}

// Image operations
func (m *MockDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	m.ImageListCalls = append(m.ImageListCalls, ImageListCall{Ctx: ctx, Options: options})

	if m.ShouldFailImageList {
		return nil, fmt.Errorf("mock image list error")
	}

	return m.MockImages, nil
}

func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	m.ImagePullCalls = append(m.ImagePullCalls, ImagePullCall{Ctx: ctx, Ref: ref, Options: options})

	if m.ShouldFailImagePull {
		return nil, fmt.Errorf("mock image pull error")
	}

	// Return a mock reader with some pull output
	pullOutput := `{"status":"Pulling from library/golang"}`
	return io.NopCloser(strings.NewReader(pullOutput)), nil
}

func (m *MockDockerClient) ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	m.ImageRemoveCalls = append(m.ImageRemoveCalls, ImageRemoveCall{Ctx: ctx, ImageID: imageID, Options: options})

	if m.ShouldFailImageRemove {
		return nil, fmt.Errorf("mock image remove error")
	}

	return []image.DeleteResponse{{Deleted: imageID}}, nil
}

// Container operations
func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerCreateCalls = append(m.ContainerCreateCalls, ContainerCreateCall{
		Ctx: ctx, Config: config, HostConfig: hostConfig,
		NetworkingConfig: networkingConfig, Platform: platform, ContainerName: containerName})

	m.containerCreateCount++

	if m.ShouldFailContainerCreate || (m.FailContainerCreateAfter > 0 && m.containerCreateCount > m.FailContainerCreateAfter) {
		return container.CreateResponse{}, fmt.Errorf("mock container create error")
	}

	// Generate a unique container ID
	containerID := fmt.Sprintf("container-%d", m.nextContainerID)
	m.nextContainerID++

	// Store the mock container
	m.MockContainers[containerID] = &MockContainer{
		ID:         containerID,
		Config:     config,
		HostConfig: hostConfig,
		State: container.State{
			Status:  "created",
			Running: false,
		},
	}

	return container.CreateResponse{ID: containerID}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerStartCalls = append(m.ContainerStartCalls, ContainerStartCall{Ctx: ctx, ContainerID: containerID, Options: options})

	if m.ShouldFailContainerStart {
		return fmt.Errorf("mock container start error")
	}

	// Update container state to running
	if mockContainer, exists := m.MockContainers[containerID]; exists {
		mockContainer.State.Status = "running"
		mockContainer.State.Running = true
	}

	return nil
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerStopCalls = append(m.ContainerStopCalls, ContainerStopCall{Ctx: ctx, ContainerID: containerID, Options: options})

	if m.ShouldFailContainerStop {
		return fmt.Errorf("mock container stop error")
	}

	// Update container state to stopped
	if mockContainer, exists := m.MockContainers[containerID]; exists {
		mockContainer.State.Status = "exited"
		mockContainer.State.Running = false
	}

	return nil
}

func (m *MockDockerClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerRestartCalls = append(m.ContainerRestartCalls, ContainerRestartCall{Ctx: ctx, ContainerID: containerID, Options: options})

	if m.ShouldFailContainerRestart {
		return fmt.Errorf("mock container restart error")
	}

	// Update container state to running after restart
	if mockContainer, exists := m.MockContainers[containerID]; exists {
		mockContainer.State.Status = "running"
		mockContainer.State.Running = true
	}

	return nil
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerInspectCalls = append(m.ContainerInspectCalls, ContainerInspectCall{Ctx: ctx, ContainerID: containerID})

	if m.ShouldFailContainerInspect {
		return container.InspectResponse{}, fmt.Errorf("mock container inspect error")
	}

	// Return mock container info
	if mockContainer, exists := m.MockContainers[containerID]; exists {
		inspectResp := container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				ID:    containerID,
				State: &mockContainer.State,
			},
		}
		// Set Config and HostConfig separately to avoid struct literal issues
		inspectResp.Config = mockContainer.Config
		inspectResp.HostConfig = mockContainer.HostConfig
		return inspectResp, nil
	}

	return container.InspectResponse{}, fmt.Errorf("container not found: %s", containerID)
}

func (m *MockDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerStatsCalls = append(m.ContainerStatsCalls, ContainerStatsCall{Ctx: ctx, ContainerID: containerID, Stream: stream})

	if m.ShouldFailContainerStats {
		return container.StatsResponseReader{}, fmt.Errorf("mock container stats error")
	}

	// Return mock stats
	mockStats := `{"read":"2023-01-01T00:00:00Z","memory":{"usage":1048576},"cpu":{"usage":{"total":1000000}}}`
	reader := io.NopCloser(strings.NewReader(mockStats))

	return container.StatsResponseReader{Body: reader}, nil
}

func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerWaitCalls = append(m.ContainerWaitCalls, ContainerWaitCall{Ctx: ctx, ContainerID: containerID, Condition: condition})

	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)

	if m.ShouldFailContainerWait {
		errCh <- fmt.Errorf("mock container wait error")
	} else {
		statusCh <- container.WaitResponse{StatusCode: 0}
	}

	return statusCh, errCh
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerRemoveCalls = append(m.ContainerRemoveCalls, ContainerRemoveCall{Ctx: ctx, ContainerID: containerID, Options: options})

	if m.ShouldFailContainerRemove {
		return fmt.Errorf("mock container remove error")
	}

	// Remove container from mock storage
	delete(m.MockContainers, containerID)

	return nil
}

// Container execution operations
func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerExecCreateCalls = append(m.ContainerExecCreateCalls, ContainerExecCreateCall{Ctx: ctx, ContainerID: containerID, Config: config})

	if m.ShouldFailContainerExecCreate {
		return container.ExecCreateResponse{}, fmt.Errorf("mock container exec create error")
	}

	// Generate a unique exec ID
	execID := fmt.Sprintf("exec-%d", m.nextExecID)
	m.nextExecID++

	// Store the mock exec
	m.MockExecs[execID] = &MockExec{
		ID:          execID,
		ContainerID: containerID,
		Running:     false,
		ExitCode:    0,
		Config:      config,
	}

	// Check if there's a specific error for this exec
	if err, exists := m.MockExecCreateResponses[execID]; exists && err != nil {
		return container.ExecCreateResponse{}, err
	}

	return container.ExecCreateResponse{ID: execID}, nil
}

func (m *MockDockerClient) ContainerExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerExecStartCalls = append(m.ContainerExecStartCalls, ContainerExecStartCall{Ctx: ctx, ExecID: execID, Config: config})

	if m.ShouldFailContainerExecStart {
		return fmt.Errorf("mock container exec start error")
	}

	// Update exec state to running, then completed
	if mockExec, exists := m.MockExecs[execID]; exists {
		mockExec.Running = true
		// Simulate execution completion
		go func() {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			mockExec.Running = false
		}()
	}

	return nil
}

func (m *MockDockerClient) ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerExecInspectCalls = append(m.ContainerExecInspectCalls, ContainerExecInspectCall{Ctx: ctx, ExecID: execID})

	if m.ShouldFailContainerExecInspect {
		return container.ExecInspect{}, fmt.Errorf("mock container exec inspect error")
	}

	// Check if there's a specific response for this exec
	if mockResponse, exists := m.MockExecInspectResponses[execID]; exists {
		return mockResponse, nil
	}

	// Return mock exec info
	if mockExec, exists := m.MockExecs[execID]; exists {
		return container.ExecInspect{
			ExecID:      execID,
			Running:     mockExec.Running,
			ExitCode:    mockExec.ExitCode,
			ContainerID: mockExec.ContainerID,
		}, nil
	}

	return container.ExecInspect{}, fmt.Errorf("exec not found: %s", execID)
}

func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ContainerExecAttachCalls = append(m.ContainerExecAttachCalls, ContainerExecAttachCall{Ctx: ctx, ExecID: execID, Config: config})

	if m.ShouldFailContainerExecAttach {
		return types.HijackedResponse{}, fmt.Errorf("mock container exec attach error")
	}

	// Check if there's a specific response for this exec
	if mockResponse, exists := m.MockExecAttachResponses[execID]; exists {
		if mockResponse.Error != nil {
			return types.HijackedResponse{}, mockResponse.Error
		}

		reader := io.NopCloser(strings.NewReader(mockResponse.Output))
		return types.HijackedResponse{
			Reader: bufio.NewReader(reader),
			Conn:   &mockConn{},
		}, nil
	}

	return types.HijackedResponse{
		Reader: bufio.NewReader(strings.NewReader("")),
		Conn:   &mockConn{},
	}, nil
}

func (m *MockDockerClient) CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CopyToContainerCalls = append(m.CopyToContainerCalls, CopyToContainerCall{Ctx: ctx, ContainerID: containerID, DstPath: dstPath, Content: content, Options: options})

	if m.ShouldFailCopyToContainer {
		return fmt.Errorf("mock container copy to container error")
	}

	return nil
}

func (m *MockDockerClient) CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error) {

	m.CopyFromContainerCalls = append(m.CopyFromContainerCalls, CopyFromContainerCall{Ctx: ctx, ContainerID: containerID, SrcPath: srcPath})

	if m.ShouldFailCopyFromContainer {
		return nil, container.PathStat{}, fmt.Errorf("mock container copy from container error")
	}

	return nil, container.PathStat{}, nil
}

func (m *MockDockerClient) IsErrNotFound(err error) bool {
    // Simulate the behavior for the mock
    if err == nil {
        return false
    }
    return strings.Contains(err.Error(), "not found")
}

// Helper methods for setting up mock data
func (m *MockDockerClient) AddMockImage(repoTag string) {
	m.MockImages = append(m.MockImages, image.Summary{
		ID:       "mock-image-id",
		RepoTags: []string{repoTag},
	})
}

func (m *MockDockerClient) SetMockExecExitCode(execID string, exitCode int) {
	if mockExec, exists := m.MockExecs[execID]; exists {
		mockExec.ExitCode = exitCode
	}
}

// SetContainerState sets the state of a mock container
func (m *MockDockerClient) SetContainerState(containerID string, state container.State) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if mockContainer, exists := m.MockContainers[containerID]; exists {
		mockContainer.State = state
	}
}

// SetExecCreateResponse sets the response for a specific exec create call
func (m *MockDockerClient) SetExecCreateResponse(execID string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.MockExecCreateResponses[execID] = err
}

// SetExecAttachResponse sets the response for a specific exec attach call
func (m *MockDockerClient) SetExecAttachResponse(response MockHijackedResponse, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if err != nil {
		response.Error = err
	}
	// Use a default exec ID if none is provided
	execID := "default-exec"
	m.MockExecAttachResponses[execID] = response
}

// SetExecInspectResponse sets the response for a specific exec inspect call
func (m *MockDockerClient) SetExecInspectResponse(response container.ExecInspect, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// Use a default exec ID if none is provided
	execID := "default-exec"
	m.MockExecInspectResponses[execID] = response
}
