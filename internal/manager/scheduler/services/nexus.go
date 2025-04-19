package services

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"crypto/tls"
	"crypto/x509"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	keeperConfig "github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/services"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TaskValidationRequest struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID uint16 `json:"taskDefinitionId"`
	Performer        string `json:"performer"`
}

type ValidationResponse struct {
	Data    bool   `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}

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
	logger.Info(">>> Oh, I am the performer...")
	logger.Info(">>> Don't mind if I do...")

	// Create ethClient using config
	ethClient, err := ethclient.Dial("https://opt-sepolia.g.alchemy.com/v2/E3OSaENxCMNoRBi_quYcmTNPGfRitxQa")
	if err != nil {
		logger.Errorf("Failed to connect to Ethereum client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Ethereum network"})
		return
	}
	defer ethClient.Close()

	// Create job executor with ethClient and etherscan API key
	jobExecutor := execution.NewJobExecutor(ethClient, keeperConfig.AlchemyAPIKey)

	actionData, err := jobExecutor.Execute(&jobData)
	if err != nil {
		logger.Errorf("Error executing job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job execution failed"})
		return
	}

	// // Update keeper metrics after successful job execution
	// keeperID := os.Getenv("KEEPER_ID")
	// if keeperID == "" {
	// 	logger.Warn("KEEPER_ID environment variable not set, using default value")
	// }
	// taskID := triggerData.TaskID

	// // Call the metrics server to store keeper execution metrics
	// if err := StoreKeeperMetrics(keeperID, fmt.Sprintf("%d", taskID)); err != nil {
	// 	logger.Warnf("Failed to store keeper metrics: %v", err)
	// 	// Continue execution even if metrics storage fails
	// } else {
	// 	logger.Infof("Successfully stored metrics for keeper %d and task %d", keeperID, taskID)
	// }

	actionData.TaskID = triggerData.TaskID

	logger.Infof("actionData: %v\n", actionData)

	actionDataBytes, err := json.Marshal(actionData)
	if err != nil {
		logger.Errorf("Error marshaling execution result:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal execution result"})
		return
	}
	krw := &execution.KeeperResponseWrapper{Data: actionDataBytes}

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

func ValidateTask(c *gin.Context) {
	var taskRequest TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	logger.Info("Received Task Validation Request:")
	// logger.Infof("Proof of Task: %s", taskRequest.ProofOfTask)
	// logger.Infof("Data: %s", taskRequest.Data)
	logger.Infof("Task Definition ID: %d", taskRequest.TaskDefinitionID)
	logger.Infof("Performer Address: %s", taskRequest.Performer)

	// Decode the data if it's hex-encoded (with 0x prefix)
	var decodedData string
	if strings.HasPrefix(taskRequest.Data, "0x") {
		dataBytes, err := hex.DecodeString(taskRequest.Data[2:]) // Remove "0x" prefix before decoding
		if err != nil {
			logger.Errorf("Failed to hex-decode data: %v", err)
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Data:    false,
				Error:   true,
				Message: fmt.Sprintf("Failed to decode hex data: %v", err),
			})
			return
		}
		decodedData = string(dataBytes)
		logger.Infof("Decoded Data: %s", decodedData)
	} else {
		decodedData = taskRequest.Data
	}

	// Fetch the ActionData from IPFS using CID from the proof of task
	ipfsContent, err := ipfs.FetchIPFSContent(decodedData)
	if err != nil {
		logger.Errorf("Failed to fetch IPFS content from ProofOfTask: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content from ProofOfTask: %v", err),
		})
		return
	}

	// Log the decoded data CID for debugging
	// logger.Infof("Data CID: %s", decodedData)

	// Parse IPFS data into IPFSData struct
	var ipfsData types.IPFSData
	if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
		logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse IPFS content: %v", err),
		})
		return
	}

	// Extract job ID and execution timestamp
	jobID := ipfsData.JobData.JobID
	executionTimestamp := ipfsData.ActionData.Timestamp
	taskID := ipfsData.ActionData.TaskID
	taskFee := ipfsData.ActionData.TotalFee

	if err := updateTaskFeeInDatabase(taskID, taskFee); err != nil {
		logger.Errorf("Failed to update task fee in database: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to update task fee in database: %v", err),
		})
	}

	// Update the last executed timestamp in the database
	if err := updateJobLastExecutedTimestamp(jobID, executionTimestamp); err != nil {
		logger.Errorf("Failed to update job last executed timestamp in database: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to update job last executed timestamp: %v", err),
		})
	}

	// Update job's last execution time in the running worker
	if err := updateJobStateInScheduler(jobID, executionTimestamp); err != nil {
		logger.Warnf("Failed to update job state in scheduler: %v", err)
		// Continue processing even if scheduler update fails
	}

	// Return success response
	c.JSON(http.StatusOK, ValidationResponse{
		Data:    true,
		Error:   false,
		Message: fmt.Sprintf("Successfully validated task for job ID %d", jobID),
	})
}

func updateTaskFeeInDatabase(taskID int64, taskFee float64) error {
	databaseURL := fmt.Sprintf("%s/api/tasks/%d/fee", config.DatabaseIPAddress, taskID)

	requestBody, err := json.Marshal(map[string]float64{
		"fee": taskFee,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal task fee data: %w", err)
	}

	// Send a PUT request to update the task fee
	req, err := http.NewRequest(http.MethodPut, databaseURL, strings.NewReader(string(requestBody)))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	logger.Infof("Successfully updated task fee for task ID %d to %f", taskID, taskFee)
	return nil
}

// updateJobLastExecutedTimestamp updates the last_executed_at timestamp in the database
func updateJobLastExecutedTimestamp(jobID int64, timestamp time.Time) error {
	databaseURL := fmt.Sprintf("%s/api/jobs/%d/lastexecuted", config.DatabaseIPAddress, jobID)

	// Create the request body
	requestBody, err := json.Marshal(map[string]string{
		"timestamp": timestamp.Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal timestamp data: %w", err)
	}

	// Send a PUT request to update the last executed timestamp
	req, err := http.NewRequest(http.MethodPut, databaseURL, strings.NewReader(string(requestBody)))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	logger.Infof("Successfully updated last executed timestamp for job %d in database", jobID)
	return nil
}

// updateJobStateInScheduler updates the last execution time in the running scheduler worker
func updateJobStateInScheduler(jobID int64, timestamp time.Time) error {
	// Define the update data structure
	updateData := map[string]interface{}{
		"job_id":    jobID,
		"timestamp": timestamp,
	}

	// Marshal the update data
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal job state update data: %w", err)
	}

	// Send the update to the scheduler through an internal endpoint
	schedulerURL := fmt.Sprintf("http://localhost:%s/job/state/update", config.ManagerRPCPort)

	req, err := http.NewRequest(http.MethodPost, schedulerURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create scheduler update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send scheduler update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code from scheduler update: %d", resp.StatusCode)
	}

	logger.Infof("Successfully updated job %d state in scheduler with timestamp %v", jobID, timestamp)
	return nil
}
