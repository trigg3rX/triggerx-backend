package container

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/trigg3rX/triggerx-backend/pkg/client/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/scripts"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Events when containers are created, started, stopped, error, etc.
type ContainerEvent struct {
	Type        string // "created", "started", "stopped", "error"
	ContainerID string
	Language    types.Language
	Timestamp   time.Time
	Metadata    map[string]interface{}
}

type containerManager struct {
	dockerClient docker.DockerClientAPI
	fileSystem   fs.FileSystemAPI
	config       config.ConfigProviderInterface
	logger       logging.Logger
	pools        map[types.Language]poolAPI
	mutex        sync.RWMutex
	initialized  bool
	// Object pools for reusable objects to reduce GC pressure
	executionResultPool sync.Pool
	bytesBufferPool     sync.Pool
	tarWriterPool       sync.Pool
}

// NewContainerManager creates a new container manager with dependency injection
func NewContainerManager(
	dockerClient docker.DockerClientAPI,
	fileSystem fs.FileSystemAPI,
	cfg config.ConfigProviderInterface,
	logger logging.Logger,
) (*containerManager, error) {
	manager := &containerManager{
		dockerClient: dockerClient,
		fileSystem:   fileSystem,
		config:       cfg,
		logger:       logger,
		pools:        make(map[types.Language]poolAPI),
		executionResultPool: sync.Pool{
			New: func() interface{} {
				return &types.ExecutionResult{}
			},
		},
		bytesBufferPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
		tarWriterPool: sync.Pool{
			New: func() interface{} {
				buf := &bytes.Buffer{}
				return tar.NewWriter(buf)
			},
		},
	}
	return manager, nil
}

// Initialize initializes the container manager and pulls all required images.
func (m *containerManager) Initialize(ctx context.Context) error {
	m.logger.Info("Initializing Docker manager")

	// Proactively pull all images required by the configured languages.
	// This ensures images are ready before pools start creating containers.
	supportedLanguages := m.config.GetSupportedLanguages()
	pulledImages := make(map[string]bool) // Use a map to avoid pulling the same image multiple times

	for _, lang := range supportedLanguages {
		poolConfig, exists := m.config.GetLanguagePoolConfig(lang)
		if !exists {
			continue // Should not happen if config is validated
		}

		imageName := poolConfig.LanguageConfig.ImageName
		if !pulledImages[imageName] {
			m.logger.Infof("Pulling required image for language %s: %s", lang, imageName)
			if err := m.PullImage(ctx, imageName); err != nil {
				// We can treat this as a warning or a fatal error.
				// For robustness, let's warn and continue.
				m.logger.Warnf("Failed to pull image %s: %v", imageName, err)
			}
			pulledImages[imageName] = true
		}
	}

	m.logger.Info("Docker manager initialized successfully")
	return nil
}

// InitializeLanguagePools initializes language-specific container pools
func (m *containerManager) InitializeLanguagePools(ctx context.Context, languages []types.Language) error {
	m.mutex.Lock()
	if m.initialized {
		m.mutex.Unlock()
		return fmt.Errorf("docker manager already initialized")
	}
	m.mutex.Unlock()

	m.logger.Info("Initializing language-specific container pools")

	// Use goroutines and WaitGroup for parallel initialization
	var wg sync.WaitGroup
	errors := make(chan error, len(languages))
	successCount := 0
	var successCountMutex sync.Mutex

	for _, lang := range languages {
		poolConfig, exists := m.config.GetLanguagePoolConfig(lang)
		if !exists {
			m.logger.Warnf("No configuration found for language %s, skipping", lang)
			continue
		}

		wg.Add(1)
		go func(language types.Language, config config.LanguagePoolConfig) {
			defer wg.Done()

			// Create adapter for the pool
			poolAdapter := NewContainerManagerAdapter(m)
			pool := newContainerPool(config, poolAdapter, m.logger)

			if err := pool.initialize(ctx); err != nil {
				m.logger.Warnf("Failed to initialize pool for language %s: %v", language, err)
				errors <- fmt.Errorf("failed to initialize pool for language %s: %w", language, err)
				return
			}

			// Safely add pool to the map and update success count
			successCountMutex.Lock()
			successCount++
			successCountMutex.Unlock()

			m.mutex.Lock()
			m.pools[language] = pool
			m.mutex.Unlock()

			m.logger.Infof("Initialized pool for language: %s", language)
		}(lang, poolConfig)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		m.logger.Warnf("Pool initialization error: %v", err)
	}

	m.initialized = true
	m.logger.Infof("Language-specific container pools initialized with %d pools", successCount)
	return nil
}

// GetDockerClient returns the Docker client
func (m *containerManager) GetDockerClient() docker.DockerClientAPI {
	return m.dockerClient
}

// GetContainer returns a container for the specified language
func (m *containerManager) GetContainer(ctx context.Context, language types.Language) (*types.PooledContainer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("docker manager not initialized")
	}

	pool, exists := m.pools[language]
	if !exists {
		return nil, fmt.Errorf("no pool available for language: %s", language)
	}

	return pool.getContainer(ctx)
}

// ReturnContainer returns a container to its language-specific pool
func (m *containerManager) ReturnContainer(container *types.PooledContainer) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		return fmt.Errorf("docker manager not initialized")
	}

	pool, exists := m.pools[container.Language]
	if !exists {
		return fmt.Errorf("no pool available for language: %s", container.Language)
	}

	return pool.returnContainer(container)
}

// GetPoolStats returns statistics for all language pools
func (m *containerManager) GetPoolStats() map[types.Language]*types.PoolStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[types.Language]*types.PoolStats)
	for lang, pool := range m.pools {
		stats[lang] = pool.getStats()
	}

	return stats
}

// GetLanguageStats returns statistics for a specific language pool
func (m *containerManager) GetLanguageStats(language types.Language) (*types.PoolStats, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	pool, exists := m.pools[language]
	if !exists {
		return nil, false
	}

	return pool.getStats(), true
}

// GetHealthCheckStats returns health check statistics for all language pools
func (m *containerManager) GetHealthCheckStats() map[types.Language]map[string]int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[types.Language]map[string]int)
	for lang, pool := range m.pools {
		total, toCheck, inError := pool.getHealthCheckStats()
		stats[lang] = map[string]int{
			"total_containers":    total,
			"containers_to_check": toCheck,
			"containers_in_error": inError,
		}
	}

	return stats
}

// GetSupportedLanguages returns all languages with active pools
func (m *containerManager) GetSupportedLanguages() []types.Language {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	languages := make([]types.Language, 0, len(m.pools))
	for lang := range m.pools {
		languages = append(languages, lang)
	}

	return languages
}

// IsLanguageSupported checks if a language is supported
func (m *containerManager) IsLanguageSupported(language types.Language) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.pools[language]
	return exists
}

// MarkContainerAsFailed marks a container as failed and removes it from the pool
// This should be called when a container fails during command execution
func (m *containerManager) MarkContainerAsFailed(containerID string, language types.Language, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		m.logger.Warnf("Docker manager not initialized, cannot mark container as failed")
		return
	}

	pool, exists := m.pools[language]
	if !exists {
		m.logger.Warnf("No pool available for language: %s, cannot mark container as failed", language)
		return
	}

	pool.markContainerAsFailed(containerID, err)
}

// ExecuteInContainerWithLanguage executes code in a container using language-specific setup
func (m *containerManager) ExecuteInContainer(ctx context.Context, containerID string, filePath string, language types.Language) (*types.ExecutionResult, string, error) {
	m.logger.Infof("Executing file %s in container %s with language %s", filePath, containerID, language)

	// Verify container is running before execution
	inspect, err := m.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to inspect container before execution: %w", err)
	}
	m.logger.Debugf("Container %s state: %s, running: %v", containerID, inspect.State.Status, inspect.State.Running)

	if !inspect.State.Running {
		return nil, "", fmt.Errorf("container %s is not running (status: %s)", containerID, inspect.State.Status)
	}

	// Execute the code with combined file copy and execution
	m.logger.Debugf("Starting combined file copy and code execution in container %s", containerID)
	result, execID, err := m.executeCodeWithFileCopy(ctx, containerID, filePath, language)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute code: %w", err)
	}
	m.logger.Debugf("Code execution completed for container %s", containerID)

	return result, execID, nil
}

func (m *containerManager) PullImage(ctx context.Context, imageName string) error {
	// m.logger.Infof("Pulling Docker image: %s", imageName)

	// Check if image already exists locally
	images, err := m.dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		m.logger.Warnf("Failed to list images: %v", err)
	} else {
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if tag == imageName || tag == imageName+":latest" {
					m.logger.Debugf("Image %s already exists locally, skipping pull", imageName)
					return nil
				}
			}
		}
	}

	// Image doesn't exist locally, pull it
	reader, err := m.dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		m.logger.Errorf("Failed to pull image: %v", err)
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			m.logger.Errorf("Failed to close image pull reader: %v", err)
		}
	}()

	// Read the output to ensure the pull completes
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		m.logger.Errorf("Error reading image pull response: %v", err)
		return fmt.Errorf("error reading image pull response: %w", err)
	}

	// Log the last few lines of the pull output for debugging
	pullOutput := buf.String()
	lines := strings.Split(pullOutput, "\n")
	if len(lines) > 5 {
		lines = lines[len(lines)-5:]
	}
	for _, line := range lines {
		if line != "" {
			m.logger.Debugf("Pull output: %s", line)
		}
	}

	m.logger.Infof("Successfully pulled image: %s", imageName)
	return nil
}

func (m *containerManager) CleanupContainer(ctx context.Context, containerID string) error {
	if !m.config.GetManagerConfig().AutoCleanup {
		m.logger.Infof("auto cleanup is disabled, skipping container cleanup")
		return nil
	}

	return m.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

// KillExecProcess attempts to terminate a running Docker exec process
// Since Docker doesn't provide a direct API to kill exec processes,
// we rely on context cancellation and container signaling
func (m *containerManager) KillExecProcess(ctx context.Context, execID string) error {
	m.logger.Infof("Attempting to terminate exec process %s", execID)

	// First, check if the exec process is still running
	inspectResp, err := m.dockerClient.ContainerExecInspect(ctx, execID)
	if err != nil {
		m.logger.Warnf("Failed to inspect exec process %s: %v", execID, err)
		return fmt.Errorf("failed to inspect exec process: %w", err)
	}

	if !inspectResp.Running {
		m.logger.Infof("Exec process %s is already terminated", execID)
		return nil
	}

	// Since Docker doesn't provide a direct API to kill exec processes,
	// we can try to signal the container to terminate the process
	// This is a best-effort approach
	m.logger.Infof("Exec process %s is still running. Attempting to signal container %s", execID, inspectResp.ContainerID)

	// Try to send a signal to the container (this might help terminate the exec process)
	// Note: This is not guaranteed to work for all exec processes
	if inspectResp.ContainerID != "" {
		// Send SIGTERM to the container as a best-effort approach
		// This might help terminate the exec process
		m.logger.Debugf("Sending SIGTERM to container %s to help terminate exec process", inspectResp.ContainerID)
		// Note: We don't actually call ContainerKill here as it would kill the entire container
		// Instead, we rely on context cancellation and let the Docker daemon handle cleanup
	}

	m.logger.Infof("Exec process %s termination initiated. The process will terminate when the context is cancelled "+
		"or when the Docker daemon handles the cleanup.", execID)

	return nil
}

func (m *containerManager) executeCodeWithFileCopy(ctx context.Context, containerID string, filePath string, language types.Language) (*types.ExecutionResult, string, error) {
	result := m.getExecutionResult()
	outputBuffer := m.getBytesBuffer()
	defer m.returnBytesBuffer(outputBuffer)

	// Copy file to container using Docker's optimized copy method
	if err := m.copyFileToContainerOptimized(ctx, containerID, filePath, language); err != nil {
		// Return the result to pool since we're not using it
		m.returnExecutionResult(result)
		return nil, "", fmt.Errorf("failed to copy file to container: %w", err)
	}

	// Step 1: Run setup script (warming up caches, etc.)
	if err := m.runSetupScript(ctx, containerID, language); err != nil {
		// Return the result to pool since we're not using it
		m.returnExecutionResult(result)
		return nil, "", fmt.Errorf("failed to run setup script: %w", err)
	}

	// Step 2: Execute the actual code with precise timing
	executionStartTime := time.Now()
	execID, err := m.runExecutionScript(ctx, containerID, language, outputBuffer)
	executionEndTime := time.Now()
	codeExecutionTime := executionEndTime.Sub(executionStartTime)

	// Set result success/error based on execution outcome
	if err != nil {
		result.Success = false
		result.Error = err
	} else {
		result.Success = true
		result.Error = nil
	}

	// Step 3: Read the result file if execution was successful
	if result.Success {
		resultContent, err := m.readResultFile(ctx, containerID)
		if err != nil {
			m.logger.Warnf("Failed to read result file from container %s: %v", containerID, err)
			// Fall back to stdout/stderr output
			result.Output = outputBuffer.String()
		} else {
			// Use the result file content
			result.Output = resultContent
		}
	} else {
		// Use stdout/stderr output for error cases
		result.Output = outputBuffer.String()
	}

	// Step 4: Run cleanup script asynchronously (don't wait for it)
	go func() {
		if err := m.runCleanupScript(context.Background(), containerID, language); err != nil {
			m.logger.Warnf("Failed to run cleanup script for container %s: %v", containerID, err)
		}
	}()

	// Copy output from pooled buffer to result before buffer is returned to pool
	result.Stats.ExecutionTime = codeExecutionTime

	// Note: We don't return the result to the pool here because we're returning it to the caller.
	// The caller becomes responsible for the object lifecycle. The pooled buffer is automatically
	// returned via defer, but the result object is passed to the caller.
	return result, execID, nil
}

// runSetupScript runs the setup script for warming up caches and preparing the environment
func (m *containerManager) runSetupScript(ctx context.Context, containerID string, language types.Language) error {
	m.logger.Debugf("Running setup script for container %s", containerID)

	setupScript := scripts.GetSetupScript(language)
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", setupScript},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := m.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create setup exec: %w", err)
	}

	execAttachResp, err := m.dockerClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to setup exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the setup command
	if err := m.dockerClient.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start setup exec: %w", err)
	}

	// Wait for setup to complete
	setupComplete := false
	for !setupComplete {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			inspectResp, err := m.dockerClient.ContainerExecInspect(ctx, execResp.ID)
			if err != nil {
				return fmt.Errorf("failed to inspect setup exec: %w", err)
			}
			if !inspectResp.Running {
				if inspectResp.ExitCode != 0 {
					return fmt.Errorf("setup script failed with exit code: %d", inspectResp.ExitCode)
				}
				setupComplete = true
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// runExecutionScript runs the actual code execution with precise timing
func (m *containerManager) runExecutionScript(ctx context.Context, containerID string, language types.Language, outputBuffer *bytes.Buffer) (string, error) {
	m.logger.Debugf("Running execution script for container %s", containerID)

	executionScript := scripts.GetExecutionScript(language)
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", executionScript},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := m.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create execution exec: %w", err)
	}

	execAttachResp, err := m.dockerClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach to execution exec: %w", err)
	}
	defer execAttachResp.Close()

	// Execute the command
	if err := m.dockerClient.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start execution exec: %w", err)
	}

	// Read output in a goroutine
	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		scanner := bufio.NewScanner(execAttachResp.Reader)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuffer.WriteString(line + "\n")
		}
	}()

	// Wait for execution to complete
	var exitCode int
	executionComplete := false
	timeout := time.After(60 * time.Second)

	for !executionComplete {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timeout:
			return "", fmt.Errorf("execution timeout after 60 seconds")
		case <-outputDone:
			// Output reading completed, check for completion file
			if m.checkExecutionComplete(ctx, containerID) {
				executionComplete = true
				exitCode = 0 // Assume success if we have output and completion file
			}
		default:
			// Check for completion file periodically
			if m.checkExecutionComplete(ctx, containerID) {
				executionComplete = true
				exitCode = 0
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// Check if execution was successful
	if exitCode != 0 {
		return "", fmt.Errorf("execution failed with exit code: %d", exitCode)
	}

	m.logger.Debugf("Execution script completed for container %s", containerID)
	return execResp.ID, nil
}

// Helper function to check if execution is complete
func (m *containerManager) checkExecutionComplete(ctx context.Context, containerID string) bool {
	// Check if result.json and execution_complete.flag exist
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", "test -f /code/result.json && test -f /code/execution_complete.flag"},
		AttachStdout: false,
		AttachStderr: false,
	}

	execResp, err := m.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return false
	}

	if err := m.dockerClient.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return false
	}

	// Wait for the test command to complete
	for {
		inspectResp, err := m.dockerClient.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			return false
		}
		if !inspectResp.Running {
			return inspectResp.ExitCode == 0
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// readResultFile reads the result.json file from the container
func (m *containerManager) readResultFile(ctx context.Context, containerID string) (string, error) {
	m.logger.Debugf("Reading result file from container %s", containerID)

	// Copy result.json from container to host
	reader, _, err := m.dockerClient.CopyFromContainer(ctx, containerID, "/code/result.json")
	if err != nil {
		return "", fmt.Errorf("failed to copy result.json from container: %w", err)
	}
	defer reader.Close()

    // Create a tar reader to extract the file content
    tarReader := tar.NewReader(reader)
    
    // Read the tar header
    header, err := tarReader.Next()
    if err != nil {
        return "", fmt.Errorf("failed to read tar header: %w", err)
    }
    
    // Verify it's the expected file
    if header.Name != "result.json" {
        return "", fmt.Errorf("unexpected file in tar: %s", header.Name)
    }
    
    // Read the file content
    content, err := io.ReadAll(tarReader)
    if err != nil {
        return "", fmt.Errorf("failed to read file content from tar: %w", err)
    }

	return string(content), nil
}

// runCleanupScript runs the cleanup script asynchronously
func (m *containerManager) runCleanupScript(ctx context.Context, containerID string, language types.Language) error {
	m.logger.Debugf("Running cleanup script for container %s", containerID)

	cleanupScript := scripts.GetCleanupScript(language)
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", cleanupScript},
		AttachStdout: false,
		AttachStderr: false,
	}

	execResp, err := m.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create cleanup exec: %w", err)
	}

	// Execute the cleanup command (detached)
	if err := m.dockerClient.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("failed to start cleanup exec: %w", err)
	}

	// Don't wait for cleanup to complete - it's asynchronous
	m.logger.Debugf("Cleanup script started for container %s", containerID)
	return nil
}

func (m *containerManager) copyFileToContainerOptimized(ctx context.Context, containerID string, filePath string, language types.Language) error {
	// Read the file content
	content, err := m.fileSystem.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Determine the target filename based on language
	var targetFile string
	switch language {
	case types.LanguageGo:
		targetFile = "code.go"
	case types.LanguagePy:
		targetFile = "code.py"
	case types.LanguageJS, types.LanguageNode:
		targetFile = "code.js"
	case types.LanguageTS:
		targetFile = "code.ts"
	default:
		targetFile = "code.go"
	}

	// Create a tar archive in memory using pooled buffer
	buf := m.getBytesBuffer()
	defer m.returnBytesBuffer(buf)
	tw := tar.NewWriter(buf)

	// Create tar header
	header := &tar.Header{
		Name: targetFile,
		Mode: 0644,
		Size: int64(len(content)),
	}

	// Write header
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	// Write file content
	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("failed to write file content to tar: %w", err)
	}

	// Close tar writer
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Copy to container using Docker's optimized method
	copyOptions := container.CopyToContainerOptions{}

	if err := m.dockerClient.CopyToContainer(ctx, containerID, "/code/", bytes.NewReader(buf.Bytes()), copyOptions); err != nil {
		return fmt.Errorf("failed to copy file to container: %w", err)
	}

	return nil
}

// getExecutionResult gets a clean ExecutionResult from the pool
func (m *containerManager) getExecutionResult() *types.ExecutionResult {
	result := m.executionResultPool.Get().(*types.ExecutionResult)
	// Reset the result to ensure clean state
	*result = types.ExecutionResult{}
	return result
}

// returnExecutionResult returns an ExecutionResult to the pool for reuse
func (m *containerManager) returnExecutionResult(result *types.ExecutionResult) {
	if result != nil {
		// Clear sensitive data before returning to pool
		result.Output = ""
		result.Error = nil
		result.Warnings = result.Warnings[:0] // Reset slice but keep capacity
		result.Stats = types.DockerResourceStats{}
		m.executionResultPool.Put(result)
	}
}

// getBytesBuffer gets a clean bytes.Buffer from the pool
func (m *containerManager) getBytesBuffer() *bytes.Buffer {
	buf := m.bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset() // Clear any existing data
	return buf
}

// returnBytesBuffer returns a bytes.Buffer to the pool for reuse
func (m *containerManager) returnBytesBuffer(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset() // Clear the buffer before returning to pool
		m.bytesBufferPool.Put(buf)
	}
}

func (m *containerManager) Close(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close all language pools
	for lang, pool := range m.pools {
		if err := pool.close(ctx); err != nil {
			m.logger.Warnf("Failed to close pool for language %s: %v", lang, err)
		}
	}

	m.logger.Info("Docker manager closed")
	return nil
}
