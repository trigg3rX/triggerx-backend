package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	// "strings"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
}

func downloadIPFSFile(ipfsURL string) (string, error) {
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

func createDockerContainer(ctx context.Context, cli *client.Client, codePath string) (string, error) {
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
    echo "Starting Go program execution..."
    go run code.go 2>&1 || {
        echo "Error executing Go program. Exit code: $?"
        exit 1
    }
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

func monitorResources(ctx context.Context, cli *client.Client, containerID string) (*ResourceStats, error) {
	stats := &ResourceStats{}

	// Start container
	err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Enhanced log handling
	logReader, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
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

	// Print container output
	go func() {
		scanner := bufio.NewScanner(logReader)
		for scanner.Scan() {
			fmt.Printf("Container Log: %s\n", scanner.Text())
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

	return stats, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: program <ipfs-url>")
		os.Exit(1)
	}

	ipfsURL := os.Args[1]
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		fmt.Printf("Failed to create Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	if err := pullDockerImage(ctx, cli, "golang:latest"); err != nil {
		fmt.Printf("Failed to pull Docker image: %v\n", err)
		os.Exit(1)
	}

	codePath, err := downloadIPFSFile(ipfsURL)
	if err != nil {
		fmt.Printf("Failed to download IPFS file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded file path: %s\n", codePath)
	content, err := os.ReadFile(codePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("File content length: %d bytes\n", len(content))
	defer os.RemoveAll(filepath.Dir(codePath))

	containerID, err := createDockerContainer(ctx, cli, codePath)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		os.Exit(1)
	}
	defer cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})

	fmt.Println("\nStarting container and monitoring resources...")
	stats, err := monitorResources(ctx, cli, containerID)
	if err != nil {
		fmt.Printf("Failed to monitor resources: %v\n", err)
		os.Exit(1)
	}

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
}
