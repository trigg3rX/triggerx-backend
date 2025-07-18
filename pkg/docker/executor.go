package docker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type CodeExecutor struct {
	DockerManager *Manager
	Downloader    *Downloader
	config        ExecutorConfig
	logger        logging.Logger
}

func NewCodeExecutor(ctx context.Context, cfg ExecutorConfig, logger logging.Logger) (*CodeExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	manager := NewManager(cli, cfg.Docker, logger)

	// Ensure the Docker image is available
	logger.Infof("Checking for Docker image: %s", cfg.Docker.Image)

	// Pull the image with retry logic for Docker-in-Docker reliability
	var pullErr error
	for attempts := 1; attempts <= 3; attempts++ {
		logger.Infof("Attempt %d/3: Pulling Docker image %s", attempts, cfg.Docker.Image)
		if err := manager.PullImage(ctx, cfg.Docker.Image); err != nil {
			pullErr = err
			logger.Warnf("Failed to pull image (attempt %d/3): %v", attempts, err)
			time.Sleep(time.Duration(attempts) * time.Second)
			continue
		}
		pullErr = nil
		break
	}

	// Log success or warning
	if pullErr != nil {
		logger.Warnf("Could not pull image %s after 3 attempts. Will try to use local image if available: %v",
			cfg.Docker.Image, pullErr)
	} else {
		logger.Infof("Successfully pulled image: %s", cfg.Docker.Image)
	}

	downloader, err := NewDownloader(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}

	return &CodeExecutor{
		DockerManager: manager,
		Downloader:    downloader,
		config:        cfg,
		logger:        logger,
	}, nil
}

func (e *CodeExecutor) Execute(ctx context.Context, fileURL string, noOfAttesters int) (*ExecutionResult, error) {
	// 1. Download code from IPFS
	codePath, err := e.Downloader.DownloadFile(fileURL, e.logger)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("IPFS download failed: %w", err),
		}, nil
	}

	// Always prepare code for Docker-in-Docker execution
	e.logger.Infof("Original code path: %s", codePath)
	codePath, err = e.ensureDinDCompatiblePath(ctx, codePath)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("failed to prepare code for Docker-in-Docker: %w", err),
		}, nil
	}

	// 2. Create and setup container
	containerID, err := e.DockerManager.CreateContainer(ctx, filepath.Dir(codePath))
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("container creation failed: %w", err),
		}, nil
	}
	defer func() {
		if err := e.DockerManager.CleanupContainer(ctx, containerID); err != nil {
			e.logger.Errorf("failed to cleanup container %s: %v", containerID, err)
		}
	}()

	// Copy the file directly into the container
	if err := e.CopyFileToContainer(ctx, e.DockerManager.Cli, containerID, codePath); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("failed to copy file to container: %w", err),
		}, nil
	}

	result, err := e.MonitorExecution(ctx, e.DockerManager.Cli, containerID, noOfAttesters)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("execution failed: %w", err),
		}, nil
	}

	return result, nil
}

// ensureDinDCompatiblePath ensures the code is in a location accessible by Docker-in-Docker
// If the path is not in /tmp, it copies the code to /tmp
func (e *CodeExecutor) ensureDinDCompatiblePath(ctx context.Context, codePath string) (string, error) {
	// Always move to /tmp for Docker-in-Docker compatibility
	e.logger.Infof("Preparing code from %s for Docker-in-Docker execution", codePath)

	// Create a new temporary directory in /tmp with world-readable permissions
	tmpDir, err := os.MkdirTemp("/tmp", "ipfs-code-dind")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory in /tmp: %w", err)
	}

	// Set permissions to ensure accessibility from any Docker container
	if err := os.Chmod(tmpDir, 0777); err != nil {
		e.logger.Warnf("Failed to set permissions on %s: %v", tmpDir, err)
	}

	// Copy the code file
	srcFile, err := os.Open(codePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source code file: %w", err)
	}
	defer srcFile.Close()

	// Use a standard filename that will be explicitly referenced in the container
	destPath := filepath.Join(tmpDir, "code.go")
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy code file: %w", err)
	}

	// Verify the file was copied correctly by reading it back
	fileContent, err := os.ReadFile(destPath)
	if err != nil {
		e.logger.Warnf("Failed to read back copied file: %v", err)
	} else {
		e.logger.Infof("Successfully copied %d bytes to %s", len(fileContent), destPath)
	}

	// Set permissions on the new file to be world-readable/writable
	if err := os.Chmod(destPath, 0666); err != nil {
		e.logger.Warnf("Failed to set permissions on %s: %v", destPath, err)
	}

	// Create setup script in the new location with executable permissions
	setupScriptPath := filepath.Join(tmpDir, "setup.sh")
	if err := os.WriteFile(setupScriptPath, []byte(SetupScript), 0777); err != nil {
		return "", fmt.Errorf("failed to write setup script: %w", err)
	}

	// Verify the files exist and have correct permissions
	e.logger.Infof("Verifying files in %s:", tmpDir)
	if entries, err := os.ReadDir(tmpDir); err == nil {
		for _, entry := range entries {
			if info, err := os.Stat(filepath.Join(tmpDir, entry.Name())); err == nil {
				e.logger.Infof("  - %s (mode: %v)", entry.Name(), info.Mode())

				// Read first few bytes of the file to verify it's not empty
				if !entry.IsDir() {
					filePath := filepath.Join(tmpDir, entry.Name())
					fileContent, err := os.ReadFile(filePath)
					if err == nil {
						preview := string(fileContent)
						if len(preview) > 50 {
							preview = preview[:50] + "..."
						}
						e.logger.Infof("    - content preview: %s", preview)
					}
				}
			}
		}
	}

	e.logger.Infof("Code successfully prepared for Docker-in-Docker execution: %s", destPath)
	return destPath, nil
}

func (e *CodeExecutor) MonitorExecution(ctx context.Context, cli *client.Client, containerID string, noOfAttesters int) (*ExecutionResult, error) {
	result := &ExecutionResult{}
	var executionStartTime time.Time
	var executionEndTime time.Time
	var codeExecutionTime time.Duration
	var dockerStartTime = time.Now().UTC()
	executionStarted := false
	var outputBuffer bytes.Buffer

	containerInfo, err := e.DockerManager.GetContainerInfo(ctx, containerID)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("failed to get container info: %w", err),
		}, nil
	}

	var codePath string
	for _, mount := range containerInfo.Mounts {
		if mount.Destination == "/code" {
			codePath = filepath.Join(mount.Source, "code.go")
			break
		}
	}

	content, err := os.ReadFile(codePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read code file: %v", err)
	}

	containerStartTime := time.Now().UTC()
	err = cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	logReader, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %v", err)
	}
	defer func() {
		if err := logReader.Close(); err != nil {
			fmt.Printf("Warning: failed to close log reader: %v\n", err)
		}
	}()

	errChan := make(chan error, 1)
	doneChan := make(chan bool)
	var lastStats container.StatsResponse

	go func() {
		statsReader, err := cli.ContainerStats(ctx, containerID, true)
		if err != nil {
			errChan <- fmt.Errorf("failed to get stats stream: %v", err)
			return
		}
		defer func() {
			if err := statsReader.Body.Close(); err != nil {
				fmt.Printf("Warning: failed to close stats reader: %v\n", err)
			}
		}()

		decoder := json.NewDecoder(statsReader.Body)
		for {
			select {
			case <-doneChan:
				return
			default:
				var statsJSON container.StatsResponse
				if err := decoder.Decode(&statsJSON); err != nil {
					if err != io.EOF {
						errChan <- fmt.Errorf("failed to decode stats: %v", err)
					}
				}
				lastStats = statsJSON
				if lastStats.MemoryStats.Usage > uint64(0.9*float64(e.config.Docker.MemoryLimitBytes())) {
					result.Warnings = append(result.Warnings, "Memory usage approaching limit")
				}
				if lastStats.CPUStats.CPUUsage.TotalUsage > uint64(0.9*float64(e.config.Docker.CPULimit*1e9)) {
					result.Warnings = append(result.Warnings, "CPU usage approaching limit")
				}
				if lastStats.MemoryStats.MaxUsage > uint64(1.01*float64(e.config.Docker.MemoryLimitBytes())) {
					errChan <- fmt.Errorf("container was killed due to exceeding memory limit")
					return
				}
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(logReader)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "START_EXECUTION") {
				executionStartTime = time.Now().UTC()
				executionStarted = true
				fmt.Println("Code execution started")
			} else if strings.Contains(line, "END_EXECUTION") && executionStarted {
				executionEndTime = time.Now().UTC()
				codeExecutionTime = executionEndTime.Sub(executionStartTime)
				fmt.Printf("Code execution completed in: %v\n", codeExecutionTime)
			} else if executionStarted {
				outputBuffer.WriteString(line + "\n")
			}

			fmt.Printf("Container Log: %s\n", line)
		}
	}()

	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		close(doneChan)
		return nil, fmt.Errorf("container wait error: %v", err)
	case status := <-statusCh:
		close(doneChan)
		if status.StatusCode != 0 {
			return nil, fmt.Errorf("container exited with status %d", status.StatusCode)
		}
	case err := <-errChan:
		close(doneChan)
		return nil, fmt.Errorf("error during stats collection: %v", err)
	case <-ctx.Done():
		close(doneChan)
		return nil, fmt.Errorf("operation timed out: %v", ctx.Err())
	}

	dockerTime := time.Since(dockerStartTime)
	containerTime := time.Since(containerStartTime)

	fmt.Printf("\nTiming Breakdown:\n")
	fmt.Printf("Code Execution Time: %v\n", codeExecutionTime)
	fmt.Printf("Container Runtime: %v\n", containerTime)
	fmt.Printf("Total Docker Processing Time: %v\n", dockerTime)
	fmt.Printf("Docker Overhead: %v\n", dockerTime-codeExecutionTime)

	if lastStats.CPUStats.SystemUsage != 0 {
		cpuDelta := float64(lastStats.CPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(lastStats.CPUStats.SystemUsage)
		// Fix: Use 1 as default CPU count if PercpuUsage is empty
		cpuCount := len(lastStats.CPUStats.CPUUsage.PercpuUsage)
		if cpuCount == 0 {
			cpuCount = 1 // Default to 1 CPU if we can't determine the count
		}
		result.Stats.CPUPercentage = (cpuDelta / systemDelta) * float64(cpuCount) * 100.0
	} else {
		// Fallback calculation if SystemUsage is 0
		if len(lastStats.CPUStats.CPUUsage.PercpuUsage) > 0 {
			// Use a simpler calculation based on total CPU usage
			result.Stats.CPUPercentage = float64(lastStats.CPUStats.CPUUsage.TotalUsage) / 1e9 * 100.0
		} else {
			// Final fallback: calculate based on total usage and execution time
			if codeExecutionTime > 0 {
				// Convert nanoseconds to seconds and calculate percentage
				cpuSeconds := float64(lastStats.CPUStats.CPUUsage.TotalUsage) / 1e9
				result.Stats.CPUPercentage = (cpuSeconds / codeExecutionTime.Seconds()) * 100.0
			}
		}
	}

	result.Stats.MemoryUsage = lastStats.MemoryStats.Usage

	for _, nw := range lastStats.Networks {
		result.Stats.RxBytes += nw.RxBytes
		result.Stats.RxPackets += nw.RxPackets
		result.Stats.RxErrors += nw.RxErrors
		result.Stats.RxDropped += nw.RxDropped
		result.Stats.TxBytes += nw.TxBytes
		result.Stats.TxPackets += nw.TxPackets
		result.Stats.TxErrors += nw.TxErrors
		result.Stats.TxDropped += nw.TxDropped
	}

	result.Stats.BandwidthRate = float64(result.Stats.RxBytes + result.Stats.TxBytes)

	for _, bioStat := range lastStats.BlkioStats.IoServiceBytesRecursive {
		switch bioStat.Op {
		case "Read":
			result.Stats.BlockRead += bioStat.Value
		case "Write":
			result.Stats.BlockWrite += bioStat.Value
		}
	}

	if !executionStarted || codeExecutionTime == 0 {
		codeExecutionTime = time.Since(executionStartTime)
		fmt.Println("Warning: Could not determine precise code execution time, using container execution time instead")
	}

	result.Stats.NoOfAttesters = noOfAttesters

	result.Stats.TotalCost = e.calculateFees(content, &result.Stats, codeExecutionTime)

	result.Output = outputBuffer.String()

	result.Success = executionStarted && codeExecutionTime > 0

	return result, nil
}

func (e *CodeExecutor) CopyFileToContainer(ctx context.Context, cli *client.Client, containerID, sourcePath string) error {
	// Read the file content
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create a tar archive containing the file
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add file to tar archive
	hdr := &tar.Header{
		Name: "code.go",
		Mode: 0644,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("failed to write file to tar: %w", err)
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Copy the tar archive to the container
	if err := cli.CopyToContainer(ctx, containerID, "/code", &buf, container.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	return nil
}

func (e *CodeExecutor) calculateFees(content []byte, stats *ResourceStats, executionTime time.Duration) float64 {
	// static complexity is the size of the file in KB
	staticComplexity := float64(len(content)) / (1024)

	execTimeInSeconds := executionTime.Seconds()
	// memory used is the memory usage of the container in MB
	memoryUsedMB := float64(stats.MemoryUsage) / (1024 * 1024)

	computationCost := (execTimeInSeconds * 2) + (memoryUsedMB / 128 * 1) + (staticComplexity / 1024 * 1)
	networkScalingFactor := (1 + stats.NoOfAttesters)

	totalTG := (computationCost * float64(networkScalingFactor)) + e.config.Fees.FixedCost + e.config.Fees.TransactionSimulation + e.config.Fees.OverheadCost

	totalFee := totalTG * e.config.Fees.PricePerTG

	stats.TotalCost = totalFee

	return totalFee
}

func (e *CodeExecutor) Close() error {
	if err := e.DockerManager.CleanupImages(context.Background()); err != nil {
		e.logger.Error("Error closing code executor", "error", err)
		return err
	}
	e.logger.Info("[2/3] Process: Code executor Closed")
	return nil
}
