package resources

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func pullDockerImage(ctx context.Context, cli *client.Client, image string) error {
	fmt.Printf("Pulling Docker image %s...\n", image)
	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}
	defer reader.Close()

	// Create a buffer to store the status output
	d := json.NewDecoder(reader)
	type Event struct {
		Status         string `json:"status"`
		Error          string `json:"error,omitempty"`
		Progress       string `json:"progress,omitempty"`
		ProgressDetail struct {
			Current int `json:"current"`
			Total   int `json:"total"`
		} `json:"progressDetail"`
	}

	// Read the output stream
	for {
		var event Event
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode pull status: %v", err)
		}

		if event.Error != "" {
			return fmt.Errorf("error pulling image: %s", event.Error)
		}

		if event.Status != "" {
			if event.Progress != "" {
				fmt.Printf("\r%s: %s", event.Status, event.Progress)
			} else {
				fmt.Println(event.Status)
			}
		}
	}

	fmt.Println("\nImage pull complete.")
	return nil
}

type ResourceStats struct {
	MemoryUsage   uint64  `json:"memory_usage"`
	CPUPercentage float64 `json:"cpu_percentage"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
	// New bandwidth metrics
	RxBytes       uint64  `json:"rx_bytes"`
	RxPackets     uint64  `json:"rx_packets"`
	RxErrors      uint64  `json:"rx_errors"`
	RxDropped     uint64  `json:"rx_dropped"`
	TxBytes       uint64  `json:"tx_bytes"`
	TxPackets     uint64  `json:"tx_packets"`
	TxErrors      uint64  `json:"tx_errors"`
	TxDropped     uint64  `json:"tx_dropped"`
	BandwidthRate float64 `json:"bandwidth_rate"` // bytes per second
	// Fee calculation fields
	TotalFee          float64 `json:"total_fee"`
	StaticComplexity  float64 `json:"static_complexity"`
	DynamicComplexity float64 `json:"dynamic_complexity"`
	ComplexityIndex   float64 `json:"complexity_index"`
	GasFees           float64 `json:"gas_fees"`
	Output            string  `json:"output"` // Add this field for script output
	Status            bool    `json:"status"` // Add this field for condition status
}

func DownloadIPFSFile(ipfsURL string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ipfs-code")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	resp, err := http.Get(ipfsURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	filePath := filepath.Join(tmpDir, "code.go")
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}
	return filePath, nil
}

func CreateDockerContainer(ctx context.Context, cli *client.Client, codePath string) (string, error) {
	// Get absolute path to ensure proper mounting
	absCodePath, err := filepath.Abs(codePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Create a script to set up and run the Go code
	setupScript := `#!/bin/sh
    cd /code
    go mod init code
    go mod tidy
    echo "START_EXECUTION"
    go run code.go 2>&1 || {
        echo "Error executing Go program. Exit code: $?"
        exit 1
    }
    echo "END_EXECUTION"
    `

	// Write setup script to the same directory as the code
	scriptPath := filepath.Join(filepath.Dir(absCodePath), "setup.sh")
	if err := os.WriteFile(scriptPath, []byte(setupScript), 0755); err != nil {
		return "", fmt.Errorf("failed to write setup script: %v", err)
	}

	absScriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for setup script: %v", err)
	}

	// Ensure the script has executable permissions
	if err := os.Chmod(absScriptPath, 0755); err != nil {
		return "", fmt.Errorf("failed to set permissions on setup script: %v", err)
	}

	config := &container.Config{
		Image:      "golang:latest",
		Cmd:        []string{"./setup.sh"},
		Tty:        true,
		WorkingDir: "/code",
	}
	fmt.Println("Code: ", absCodePath)
	fmt.Println("Script: ", absScriptPath)

	// Add HostConfig to mount the code and script into the container
	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/code", filepath.Dir(absScriptPath)),
		},
	}

	// Create the container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	return resp.ID, nil
}

func calculateFees(content []byte, stats *ResourceStats, executionTime time.Duration) float64 {
	// Constants for fee calculation
	const (
		PriceperTG            = 0.0001 // Price per TG unit
		Fixedcost             = 1      // Fixed cost in TG
		TransactionSimulation = 1      // Weight for TransactionSimulation in TG
	)

	var NoOfAttesters int
	// var OtherFactors int
	// Convert content size to KB for static complexity
	contentSizeKB := float64(len(content)) / (1024)

	// Calculate static complexity based on the code content in KB
	staticComplexity := contentSizeKB // Simple complexity based on code length

	// Calculate resource metrics
	execTimeInSeconds := executionTime.Seconds()
	memoryUsedMB := float64(stats.MemoryUsage) / (1024 * 1024) // Convert to MB

	ComputationCost := (execTimeInSeconds * 2) + (memoryUsedMB / 128 * 1) + (staticComplexity / 1024 * 1)
	NetworkScalingFactor := (1 + NoOfAttesters)

	// Calculate TotalTG using the new formula
	TotalTG := (ComputationCost * float64(NetworkScalingFactor)) + Fixedcost + TransactionSimulation

	// Calculate total fee based on TG units
	totalFee := TotalTG * PriceperTG

	// Update stats with fee information
	stats.TotalFee = totalFee

	return totalFee
}

func MonitorResources(ctx context.Context, cli *client.Client, containerID string) (*ResourceStats, error) {
	stats := &ResourceStats{}
	var executionStartTime time.Time
	var executionEndTime time.Time
	var codeExecutionTime time.Duration
	var dockerStartTime = time.Now().UTC()
	executionStarted := false
	var outputBuffer bytes.Buffer

	// Get container info to find the mounted directory
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %v", err)
	}

	// Find the mounted code file path
	var codePath string
	for _, mount := range containerInfo.Mounts {
		if mount.Destination == "/code" {
			codePath = filepath.Join(mount.Source, "code.go")
			break
		}
	}

	// Read the code content
	content, err := os.ReadFile(codePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read code file: %v", err)
	}

	// Start container
	containerStartTime := time.Now().UTC()
	err = cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	// Enhanced log handling
	logReader, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		// Timestamps: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %v", err)
	}
	defer logReader.Close()

	// Create channels for error handling and completion
	errChan := make(chan error, 1)
	doneChan := make(chan bool)
	var lastStats types.StatsJSON

	// Start stats collection goroutine
	go func() {
		// Get initial stats
		statsReader, err := cli.ContainerStats(ctx, containerID, true)
		if err != nil {
			errChan <- fmt.Errorf("failed to get stats stream: %v", err)
			return
		}
		defer statsReader.Body.Close()

		decoder := json.NewDecoder(statsReader.Body)
		for {
			select {
			case <-doneChan:
				return
			default:
				var statsJSON types.StatsJSON
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

	// Modify the log handling goroutine to capture output
	go func() {
		scanner := bufio.NewScanner(logReader)
		for scanner.Scan() {
			line := scanner.Text()

			// Check for timing markers
			if strings.Contains(line, "START_EXECUTION") {
				executionStartTime = time.Now().UTC()
				executionStarted = true
				fmt.Println("Code execution started")
			} else if strings.Contains(line, "END_EXECUTION") && executionStarted {
				executionEndTime = time.Now().UTC()
				codeExecutionTime = executionEndTime.Sub(executionStartTime)
				fmt.Printf("Code execution completed in: %v\n", codeExecutionTime)
			} else if executionStarted {
				// Capture output between START_EXECUTION and END_EXECUTION
				outputBuffer.WriteString(line + "\n")
			}

			// Print all container logs
			fmt.Printf("Container Log: %s\n", line)
		}
	}()

	// Wait for container to finish
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

	// Calculate timing metrics
	dockerTime := time.Since(dockerStartTime)
	containerTime := time.Since(containerStartTime)

	// Print timing information
	fmt.Printf("\nTiming Breakdown:\n")
	fmt.Printf("Code Execution Time: %v\n", codeExecutionTime)
	fmt.Printf("Container Runtime: %v\n", containerTime)
	fmt.Printf("Total Docker Processing Time: %v\n", dockerTime)
	fmt.Printf("Docker Overhead: %v\n", dockerTime-codeExecutionTime)

	// Calculate final statistics
	if lastStats.CPUStats.SystemUsage != 0 {
		cpuDelta := float64(lastStats.CPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(lastStats.CPUStats.SystemUsage)
		stats.CPUPercentage = (cpuDelta / systemDelta) * float64(len(lastStats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	stats.MemoryUsage = lastStats.MemoryStats.Usage

	// Network stats
	for _, nw := range lastStats.Networks {
		stats.RxBytes += nw.RxBytes
		stats.RxPackets += nw.RxPackets
		stats.RxErrors += nw.RxErrors
		stats.RxDropped += nw.RxDropped
		stats.TxBytes += nw.TxBytes
		stats.TxPackets += nw.TxPackets
		stats.TxErrors += nw.TxErrors
		stats.TxDropped += nw.TxDropped
	}

	// Calculate bandwidth rate based on total transfer
	stats.BandwidthRate = float64(stats.RxBytes + stats.TxBytes)

	// Block I/O stats
	for _, bioStat := range lastStats.BlkioStats.IoServiceBytesRecursive {
		if bioStat.Op == "Read" {
			stats.BlockRead += bioStat.Value
		} else if bioStat.Op == "Write" {
			stats.BlockWrite += bioStat.Value
		}
	}

	// Use actual code execution time for fee calculation
	if !executionStarted || codeExecutionTime == 0 {
		// Fallback to container execution time if markers weren't found
		codeExecutionTime = time.Since(executionStartTime)
		fmt.Println("Warning: Could not determine precise code execution time, using container execution time instead")
	}

	// Calculate fees using the measured execution time
	calculateFees(content, stats, codeExecutionTime)

	// Set the captured output in stats
	stats.Output = outputBuffer.String()

	// Set status based on successful execution
	stats.Status = executionStarted && codeExecutionTime > 0

	return stats, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("Usage: program <ipfs-url>")
	}

	ipfsURL := os.Args[1]
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	if err := pullDockerImage(ctx, cli, "golang:latest"); err != nil {
		return fmt.Errorf("failed to pull Docker image: %w", err)
	}

	codePath, err := DownloadIPFSFile(ipfsURL)
	if err != nil {
		return fmt.Errorf("failed to download IPFS file: %w", err)
	}

	fmt.Printf("Downloaded file path: %s\n", codePath)
	content, err := os.ReadFile(codePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	fmt.Printf("File content length: %d bytes\n", len(content))

	// Move the defer after the error check to ensure cleanup
	defer func() {
		if err := os.RemoveAll(filepath.Dir(codePath)); err != nil {
			fmt.Printf("Warning: failed to cleanup temporary files: %v\n", err)
		}
	}()

	containerID, err := CreateDockerContainer(ctx, cli, codePath)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Create cleanup function with timeout context
	cleanup := func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := cli.ContainerRemove(cleanupCtx, containerID, types.ContainerRemoveOptions{Force: true}); err != nil {
			fmt.Printf("Warning: failed to remove container %s: %v\n", containerID, err)
		}
	}
	defer cleanup()

	fmt.Println("\nStarting container and monitoring resources...")
	stats, err := MonitorResources(ctx, cli, containerID)
	if err != nil {
		return fmt.Errorf("failed to monitor resources: %w", err)
	}

	printResourceStats(stats)
	return nil
}

func printResourceStats(stats *ResourceStats) {
	fmt.Printf("\nResource Usage:\n")
	fmt.Printf("Memory: %.2f MB\n", float64(stats.MemoryUsage)/(1024*1024))
	fmt.Printf("CPU: %.2f%%\n", stats.CPUPercentage)
	fmt.Printf("Network Statistics:\n")
	fmt.Printf("  Received: %.2f MB (%.0f packets)\n", float64(stats.RxBytes)/(1024*1024), float64(stats.RxPackets))
	fmt.Printf("  Transmitted: %.2f MB (%.0f packets)\n", float64(stats.TxBytes)/(1024*1024), float64(stats.TxPackets))
	fmt.Printf("  Bandwidth Rate: %.2f MB/s\n", stats.BandwidthRate/(1024*1024))
	fmt.Printf("  Errors: Rx=%d, Tx=%d\n", stats.RxErrors, stats.TxErrors)
	fmt.Printf("  Dropped: Rx=%d, Tx=%d\n", stats.RxDropped, stats.TxDropped)
	fmt.Printf("Disk I/O:\n")
	fmt.Printf("  Read: %.2f MB\n", float64(stats.BlockRead)/(1024*1024))
	fmt.Printf("  Write: %.2f MB\n", float64(stats.BlockWrite)/(1024*1024))

	fmt.Printf("\nFee Calculation:\n")
	fmt.Printf("Static Complexity: %.6f\n", stats.StaticComplexity)
	fmt.Printf("Dynamic Complexity: %.6f\n", stats.DynamicComplexity)
	fmt.Printf("Complexity Index: %.6f\n", stats.ComplexityIndex)
	fmt.Printf("Gas Fees: $%.4f\n", stats.GasFees)
	fmt.Printf("Total Fee: $%.4f\n", stats.TotalFee)
}
