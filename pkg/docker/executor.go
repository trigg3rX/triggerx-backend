package docker

import (
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
)

type CodeExecutor struct {
	dockerManager *Manager
	ipfsDownloader *Downloader
	config        ExecutorConfig
}

func NewCodeExecutor(cfg ExecutorConfig) (*CodeExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &CodeExecutor{
		dockerManager:  NewManager(cli, cfg.Docker),
		ipfsDownloader: NewDownloader(cfg.IPFS),
		config:        cfg,
	}, nil
}

func (e *CodeExecutor) Execute(ctx context.Context, ipfsCID string) (*ExecutionResult, error) {
	// 1. Download code from IPFS
	codePath, err := e.ipfsDownloader.DownloadFile(ctx, ipfsCID)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("IPFS download failed: %w", err),
		}, nil
	}

	// 2. Create and setup container
	containerID, err := e.dockerManager.CreateContainer(ctx, filepath.Dir(codePath))
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("container creation failed: %w", err),
		}, nil
	}
	defer e.dockerManager.CleanupContainer(ctx, containerID)

	result, err := e.monitorExecution(ctx, containerID)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Errorf("execution failed: %w", err),
		}, nil
	}
	
	return result, nil
}

func (e *CodeExecutor) monitorExecution(ctx context.Context, containerID string) (*ExecutionResult, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	result := &ExecutionResult{}
	var executionStartTime time.Time
	var executionEndTime time.Time
	var codeExecutionTime time.Duration
	var dockerStartTime = time.Now().UTC()
	executionStarted := false
	var outputBuffer bytes.Buffer

	containerInfo, err := e.dockerManager.GetContainerInfo(ctx, containerID)
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
					return
				}
				lastStats = statsJSON
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
		result.Stats.CPUPercentage = (cpuDelta / systemDelta) * float64(len(lastStats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
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

	result.Stats.TotalCost = calculateFees(content, &result.Stats, codeExecutionTime)

	result.Output = outputBuffer.String()

	result.Success = executionStarted && codeExecutionTime > 0

	return result, nil
}

func calculateFees(content []byte, stats *ResourceStats, executionTime time.Duration) float64 {
	const (
		PriceperTG            = 0.0001
		Fixedcost             = 1
		TransactionSimulation = 1
		OverheadCost          = 0.1
	)

	var NoOfAttesters int
	contentSizeKB := float64(len(content)) / (1024)

	staticComplexity := contentSizeKB

	execTimeInSeconds := executionTime.Seconds()
	memoryUsedMB := float64(stats.MemoryUsage) / (1024 * 1024)

	ComputationCost := (execTimeInSeconds * 2) + (memoryUsedMB / 128 * 1) + (staticComplexity / 1024 * 1)
	NetworkScalingFactor := (1 + NoOfAttesters)

	TotalTG := (ComputationCost * float64(NetworkScalingFactor)) + Fixedcost + TransactionSimulation + OverheadCost

	totalFee := TotalTG * PriceperTG

	stats.TotalCost = totalFee

	return totalFee
}