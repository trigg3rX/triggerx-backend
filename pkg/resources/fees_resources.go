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
	MemoryUsage       uint64  `json:"memory_usage"`
	CPUPercentage     float64 `json:"cpu_percentage"`
	NetworkRx         uint64  `json:"network_rx"`
	NetworkTx         uint64  `json:"network_tx"`
	BlockRead         uint64  `json:"block_read"`
	BlockWrite        uint64  `json:"block_write"`
	RxBytes           uint64  `json:"rx_bytes"`
	RxPackets         uint64  `json:"rx_packets"`
	RxErrors          uint64  `json:"rx_errors"`
	RxDropped         uint64  `json:"rx_dropped"`
	TxBytes           uint64  `json:"tx_bytes"`
	TxPackets         uint64  `json:"tx_packets"`
	TxErrors          uint64  `json:"tx_errors"`
	TxDropped         uint64  `json:"tx_dropped"`
	BandwidthRate     float64 `json:"bandwidth_rate"`
	TotalFee          float64 `json:"total_fee"`
	StaticComplexity  float64 `json:"static_complexity"`
	DynamicComplexity float64 `json:"dynamic_complexity"`
	ComplexityIndex   float64 `json:"complexity_index"`
	GasFees           float64 `json:"gas_fees"`
	Output            string  `json:"output"`
	Status            bool    `json:"status"`
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
	absCodePath, err := filepath.Abs(codePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}

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

	scriptPath := filepath.Join(filepath.Dir(absCodePath), "setup.sh")
	if err := os.WriteFile(scriptPath, []byte(setupScript), 0755); err != nil {
		return "", fmt.Errorf("failed to write setup script: %v", err)
	}

	if err := os.Chmod(scriptPath, 0755); err != nil {
		return "", fmt.Errorf("failed to set permissions on setup script: %v", err)
	}

	config := &container.Config{
		Image:      "golang:latest",
		Cmd:        []string{"/code/setup.sh"},
		Tty:        true,
		WorkingDir: "/code",
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/code", filepath.Dir(absCodePath)),
		},
	}

	fmt.Printf("Code: %s\n", absCodePath)
	fmt.Printf("Script: %s\n", scriptPath)
	fmt.Printf("Mount path: %s\n", filepath.Dir(absCodePath))

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	return resp.ID, nil
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

	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %v", err)
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
	err = cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	logReader, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %v", err)
	}
	defer logReader.Close()

	errChan := make(chan error, 1)
	doneChan := make(chan bool)
	var lastStats types.StatsJSON

	go func() {
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
		stats.CPUPercentage = (cpuDelta / systemDelta) * float64(len(lastStats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	stats.MemoryUsage = lastStats.MemoryStats.Usage

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

	stats.BandwidthRate = float64(stats.RxBytes + stats.TxBytes)

	for _, bioStat := range lastStats.BlkioStats.IoServiceBytesRecursive {
		if bioStat.Op == "Read" {
			stats.BlockRead += bioStat.Value
		} else if bioStat.Op == "Write" {
			stats.BlockWrite += bioStat.Value
		}
	}

	if !executionStarted || codeExecutionTime == 0 {
		codeExecutionTime = time.Since(executionStartTime)
		fmt.Println("Warning: Could not determine precise code execution time, using container execution time instead")
	}

	stats.TotalFee = calculateFees(content, stats, codeExecutionTime)

	stats.Output = outputBuffer.String()

	stats.Status = executionStarted && codeExecutionTime > 0

	return stats, nil
}
