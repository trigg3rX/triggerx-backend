package execution

import (
	// "bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"

	// "fmt"

	"net/http"

	// "os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/services"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

// ResourceStats holds execution resource usage metrics
type ResourceStats struct {
	MemoryUsage   uint64  `json:"memory_usage"`
	CPUPercentage float64 `json:"cpu_percentage"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
	RxBytes       uint64  `json:"rx_bytes"`
	RxPackets     uint64  `json:"rx_packets"`
	RxErrors      uint64  `json:"rx_errors"`
	RxDropped     uint64  `json:"rx_dropped"`
	TxBytes       uint64  `json:"tx_bytes"`
	TxPackets     uint64  `json:"tx_packets"`
	TxErrors      uint64  `json:"tx_errors"`
	TxDropped     uint64  `json:"tx_dropped"`
	BandwidthRate float64 `json:"bandwidth_rate"`
	TotalFee      float64 `json:"total_fee"`
}

// keeperResponseWrapper wraps execution result bytes to satisfy the proof module's interface
type keeperResponseWrapper struct {
	Data []byte
}

func (krw *keeperResponseWrapper) GetData() []byte {
	return krw.Data
}

// ExecuteTask is the main handler for executing keeper tasks. It:
// 1. Validates and processes the incoming job request
// 2. Executes the job and generates execution proof
// 3. Stores proof on IPFS via Pinata
// 4. Returns execution results with proof details to the attester
func ExecuteTask(c *gin.Context) {
	logger.Info("Executing Task")

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Invalid method",
		})
		return
	}

	var requestBody struct {
		Data string `json:"data"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON body",
		})
		return
	}

	// Handle hex-encoded data (remove "0x" prefix if present)
	hexData := requestBody.Data
	if len(hexData) > 2 && hexData[:2] == "0x" {
		hexData = hexData[2:]
	}

	// Decode the hex string to bytes
	decodedData, err := hex.DecodeString(hexData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hex data",
		})
		return
	}

	decodedDataString := string(decodedData)

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(decodedDataString), &requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse JSON data",
		})
		return
	}

	jobDataRaw := requestData["jobData"]
	triggerDataRaw := requestData["triggerData"]
	performerDataRaw := requestData["performerData"]

	// logger.Infof("jobDataRaw: %v\n", jobDataRaw)
	// logger.Infof("triggerDataRaw: %v\n", triggerDataRaw)
	// logger.Infof("performerDataRaw: %v\n", performerDataRaw)

	// Convert to proper types
	var jobData types.HandleCreateJobData
	jobDataBytes, err := json.Marshal(jobDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job data format"})
		return
	}
	if err := json.Unmarshal(jobDataBytes, &jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse job data"})
		return
	}
	// logger.Infof("jobData: %v\n", jobData)

	var triggerData types.TriggerData
	triggerDataBytes, err := json.Marshal(triggerDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger data format"})
		return
	}
	if err := json.Unmarshal(triggerDataBytes, &triggerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse trigger data"})
		return
	}
	// logger.Infof("triggerData: %v\n", triggerData)

	var performerData types.GetPerformerData
	performerDataBytes, err := json.Marshal(performerDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid performer data format"})
		return
	}
	if err := json.Unmarshal(performerDataBytes, &performerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse performer data"})
		return
	}
	// logger.Infof("performerData: %v\n", performerData)

	// logger.Infof("taskDefinitionId: %v\n", jobData.TaskDefinitionID)
	logger.Infof("performerAddress: %v\n", performerData.KeeperAddress)

	if performerData.KeeperAddress != config.KeeperAddress {
		logger.Infof("I am not the performer for this task, skipping ...")
		c.JSON(http.StatusOK, gin.H{"error": "I am not the performer for this task, skipping ..."})
		return
	}

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		logger.Errorf("Failed to create Docker client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize Docker"})
		return
	}
	defer cli.Close()

	// Create temporary directory for code
	tmpDir, err := os.MkdirTemp("", "task-execution")
	if err != nil {
		logger.Errorf("Failed to create temp directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare execution environment"})
		return
	}
	defer os.RemoveAll(tmpDir)

	// Write the code/task to a file
	codePath := filepath.Join(tmpDir, "task.go")
	taskCode := generateTaskCode(jobData) // You'll need to implement this function to generate the Go code
	if err := os.WriteFile(codePath, []byte(taskCode), 0644); err != nil {
		logger.Errorf("Failed to write task code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare task"})
		return
	}

	// Create and start Docker container
	containerID, err := CreateDockerContainer(context.Background(), cli, codePath)
	if err != nil {
		logger.Errorf("Failed to create container: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize execution environment"})
		return
	}

	// Ensure container cleanup
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := cli.ContainerRemove(cleanupCtx, containerID, dockertypes.ContainerRemoveOptions{Force: true}); err != nil {
			logger.Warnf("Failed to remove container %s: %v", containerID, err)
		}
	}()

	// Monitor resources and execute task
	resourceStats, err := MonitorResources(context.Background(), cli, containerID)
	if err != nil {
		logger.Errorf("Failed to monitor resources: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
		return
	}

	// Create ethClient and execute job as before
	ethClient, err := ethclient.Dial("https://opt-sepolia.g.alchemy.com/v2/E3OSaENxCMNoRBi_quYcmTNPGfRitxQa")
	if err != nil {
		logger.Errorf("Failed to connect to Ethereum client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Ethereum network"})
		return
	}
	defer ethClient.Close()

	// Create job executor with ethClient and etherscan API key
	jobExecutor := NewJobExecutor(ethClient, config.AlchemyAPIKey)
	actionData, err := jobExecutor.Execute(&jobData)
	if err != nil {
		logger.Errorf("Error executing job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job execution failed"})
		return
	}

	// // Update keeper metrics after successful job execution
	// keeperID := os.Getenv("KEEPER_ID")
	// if keeperID == "" {
	//  logger.Warn("KEEPER_ID environment variable not set, using default value")
	// }
	// taskID := triggerData.TaskID

	// // Call the metrics server to store keeper execution metrics
	// if err := StoreKeeperMetrics(keeperID, fmt.Sprintf("%d", taskID)); err != nil {
	//  logger.Warnf("Failed to store keeper metrics: %v", err)
	//  // Continue execution even if metrics storage fails
	// } else {
	//  logger.Infof("Successfully stored metrics for keeper %d and task %d", keeperID, taskID)
	// }

	// Add resource usage and fees to the response
	actionData.ResourceStats = resourceStats
	actionData.TaskID = triggerData.TaskID

	logger.Infof("actionData: %v\n", actionData)

	actionDataBytes, err := json.Marshal(actionData)
	if err != nil {
		logger.Errorf("Error marshaling execution result:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal execution result"})
		return
	}
	krw := &keeperResponseWrapper{Data: actionDataBytes}

	// Mock TLS state for proof generation
	certBytes := []byte("mock certificate data")
	mockCert := &x509.Certificate{Raw: certBytes}
	connState := &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{mockCert},
	}

	tempData := types.IPFSData{
		JobData:     jobData,
		TriggerData: triggerData,
		ActionData:  actionData,
	}

	// Generate and store proof on IPFS, returning content identifier (CID)
	ipfsData, err := proof.GenerateAndStoreProof(krw, connState, tempData)
	if err != nil {
		logger.Errorf("Error generating/storing proof:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	// Generate TLS proof for response verification
	tlsProof, err := proof.GenerateProof(krw, connState)
	if err != nil {
		logger.Errorf("Error generating TLS proof:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	services.SendTask(tlsProof.ResponseHash, ipfsData.ProofData.ActionDataCID, jobData.TaskDefinitionID)

	logger.Infof("CID: %s", ipfsData.ProofData.ActionDataCID)

	ipfsDataBytes, err := json.Marshal(ipfsData)
	if err != nil {
		logger.Errorf("Error marshaling IPFS data:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal IPFS data"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", ipfsDataBytes)
}

// StoreKeeperMetrics makes API calls to update keeper metrics in the database
// func StoreKeeperMetrics(keeperID string, taskID string) error {
// 	// Call the increment-tasks endpoint
// 	incrementTasksURL := fmt.Sprintf("http://localhost:8080/api/keepers/%s/increment-tasks", keeperID)
// 	incrementResp, err := http.Post(incrementTasksURL, "application/json", nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to increment keeper task count: %w", err)
// 	}
// 	defer incrementResp.Body.Close()

// 	if incrementResp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("increment task count API returned non-OK status: %d", incrementResp.StatusCode)
// 	}

// 	// Call the add-points endpoint with the task ID
// 	addPointsURL := fmt.Sprintf("http://localhost:8080/api/keepers/%s/add-points", keeperID)

// 	// Create the request payload with the task ID
// 	payload := struct {
// 		TaskID string `

// Add these new types to handle resource monitoring in the response
type ExecutionResult struct {
	// ... existing fields ...
	ResourceStats *ResourceStats `json:"resource_stats,omitempty"`
}

func generateTaskCode(jobData types.HandleCreateJobData) string {
	// Implement this function to generate the Go code that needs to be executed
	// This should convert the job data into executable Go code
	return fmt.Sprintf(`
package main

import (
	"fmt"
)

func main() {
	// Implementation based on jobData
	fmt.Println("Executing task...")
	// Add actual task implementation here
}
`)
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

	return resp.ID, nil
}

func MonitorResources(ctx context.Context, cli *client.Client, containerID string) (*ResourceStats, error) {
	stats := &ResourceStats{}
	var dockerStartTime = time.Now().UTC()

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
	err = cli.ContainerStart(ctx, containerID, dockertypes.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	// Enhanced log handling
	logReader, err := cli.ContainerLogs(ctx, containerID, dockertypes.ContainerLogsOptions{
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
	var lastStats dockertypes.StatsJSON

	// Start stats collection goroutine
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
				var statsJSON dockertypes.StatsJSON
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

	// Calculate fees
	calculateFees(content, stats, time.Since(dockerStartTime))

	return stats, nil
}

func calculateFees(content []byte, stats *ResourceStats, executionTime time.Duration) {
	// Constants for fee calculation
	const (
		PriceperTG            = 0.0001 // Price per TG unit
		Fixedcost             = 1      // Fixed cost in TG
		TransactionSimulation = 1      // Weight for TransactionSimulation in TG
	)

	var NoOfAttesters int
	contentSizeKB := float64(len(content)) / (1024)
	execTimeInSeconds := executionTime.Seconds()
	memoryUsedMB := float64(stats.MemoryUsage) / (1024 * 1024)

	ComputationCost := (execTimeInSeconds * 2) + (memoryUsedMB / 128 * 1) + (contentSizeKB / 1024 * 1)
	NetworkScalingFactor := (1 + NoOfAttesters)

	TotalTG := (ComputationCost * float64(NetworkScalingFactor)) + Fixedcost + TransactionSimulation
	stats.TotalFee = TotalTG * PriceperTG
}
