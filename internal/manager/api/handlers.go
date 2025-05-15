package api

import (
	// "crypto/tls"
	// "crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/manager/client/database"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"github.com/ethereum/go-ethereum/ethclient"
	// "github.com/trigg3rX/triggerx-backend/internal/keeper/execution"
	// "github.com/trigg3rX/triggerx-backend/internal/keeper/services"
	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	// "github.com/trigg3rX/triggerx-backend/pkg/proof"
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

// Handlers contains all HTTP handler dependencies
type Handlers struct {
	logger       logging.Logger
	jobScheduler *scheduler.JobScheduler
	dbClient     *database.DatabaseClient
}

// NewHandlers creates a new instance of Handlers
func NewHandlers(logger logging.Logger, jobScheduler *scheduler.JobScheduler, dbClient *database.DatabaseClient) *Handlers {
	return &Handlers{
		logger:       logger,
		jobScheduler: jobScheduler,
		dbClient:     dbClient,
	}
}

// HandleCreateJobEvent handles job creation requests
func (h *Handlers) HandleCreateJobEvent(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Invalid method"})
		return
	}

	var jobData types.HandleCreateJobData
	if err := c.BindJSON(&jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	var err error
	switch jobData.TaskDefinitionID {
	case 1, 2:
		err = h.jobScheduler.StartTimeBasedJob(jobData)
	case 3, 4:
		err = h.jobScheduler.StartEventBasedJob(jobData)
	case 5, 6:
		err = h.jobScheduler.StartConditionBasedJob(jobData)
	default:
		h.logger.Warnf("Unknown task definition ID: %d for job: %d",
			jobData.TaskDefinitionID, jobData.JobID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
		return
	}

	if err != nil {
		h.logger.Errorf("Failed to schedule job %d: %v", jobData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule job"})
		return
	}

	h.logger.Infof("Successfully scheduled job with ID: %d", jobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job scheduled successfully"})
}

// HandleUpdateJobEvent handles job update requests
func (h *Handlers) HandleUpdateJobEvent(c *gin.Context) {
	var updateJobData types.HandleUpdateJobData
	if err := c.BindJSON(&updateJobData); err != nil {
		h.logger.Error("Failed to parse update job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	h.logger.Infof("Job update requested for ID: %d", updateJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job update request received"})
}

// HandlePauseJobEvent handles job pause requests
func (h *Handlers) HandlePauseJobEvent(c *gin.Context) {
	var pauseJobData types.HandlePauseJobData
	if err := c.BindJSON(&pauseJobData); err != nil {
		h.logger.Error("Failed to parse pause job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	h.logger.Infof("Job pause requested for ID: %d", pauseJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job pause request received"})
}

// HandleResumeJobEvent handles job resume requests
func (h *Handlers) HandleResumeJobEvent(c *gin.Context) {
	var resumeJobData types.HandleResumeJobData
	if err := c.BindJSON(&resumeJobData); err != nil {
		h.logger.Error("Failed to parse resume job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	h.logger.Infof("Job resume requested for ID: %d", resumeJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job resume request received"})
}

// HandleJobStateUpdate handles job state update requests
func (h *Handlers) HandleJobStateUpdate(c *gin.Context) {
	var updateData struct {
		JobID     int64     `json:"job_id"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := c.BindJSON(&updateData); err != nil {
		h.logger.Error("Failed to parse job state update data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	h.logger.Infof("Updating state for job ID: %d with timestamp: %v", updateData.JobID, updateData.Timestamp)

	worker := h.jobScheduler.GetWorker(updateData.JobID)
	if worker == nil {
		h.logger.Warnf("No active worker found for job ID: %d", updateData.JobID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found or not active"})
		return
	}

	if err := h.jobScheduler.UpdateJobLastExecutedTime(updateData.JobID, updateData.Timestamp); err != nil {
		h.logger.Errorf("Failed to update job %d last executed time: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job state"})
		return
	}

	if err := h.jobScheduler.UpdateJobStateCache(updateData.JobID, "last_executed", updateData.Timestamp); err != nil {
		h.logger.Warnf("Failed to update job %d state cache: %v", updateData.JobID, err)
	}

	h.logger.Infof("Successfully updated state for job ID: %d", updateData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job state updated successfully"})
}

// ExecuteTask handles P2P task execution messages
func (h *Handlers) ExecuteTask(c *gin.Context) {
	h.logger.Info("Executing Task")

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

	hexData := requestBody.Data
	if len(hexData) > 2 && hexData[:2] == "0x" {
		hexData = hexData[2:]
	}

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
	h.logger.Infof("performerAddress: %v\n", performerData.KeeperAddress)
	h.logger.Info(">>> Oh, I am the performer...")
	h.logger.Info(">>> Don't mind if I do...")

	ethClient, err := ethclient.Dial("https://opt-sepolia.g.alchemy.com/v2/E3OSaENxCMNoRBi_quYcmTNPGfRitxQa")
	if err != nil {
		h.logger.Errorf("Failed to connect to Ethereum client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Ethereum network"})
		return
	}
	defer ethClient.Close()

	// jobExecutor := execution.NewJobExecutor(ethClient, config.GetAlchemyApiKey())

	// actionData, err := jobExecutor.Execute(&jobData)
	// if err != nil {
	// 	h.logger.Errorf("Error executing job: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Job execution failed"})
	// 	return
	// }

	// actionData.TaskID = triggerData.TaskID

	// h.logger.Infof("actionData: %v\n", actionData)

	// actionDataBytes, err := json.Marshal(actionData)
	// if err != nil {
	// 	h.logger.Errorf("Error marshaling execution result:", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal execution result"})
	// 	return
	// }
	// krw := &execution.KeeperResponseWrapper{Data: actionDataBytes}

	// certBytes := []byte("mock certificate data")
	// mockCert := &x509.Certificate{Raw: certBytes}
	// connState := &tls.ConnectionState{
	// 	PeerCertificates: []*x509.Certificate{mockCert},
	// }

	// tempData := types.IPFSData{
	// 	JobData:     jobData,
	// 	TriggerData: triggerData,
	// 	ActionData:  actionData,
	// }

	// ipfsData, err := proof.GenerateAndStoreProof(krw, connState, tempData)
	// if err != nil {
	// 	h.logger.Errorf("Error generating/storing proof:", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
	// 	return
	// }

	// tlsProof, err := proof.GenerateProof(krw, connState)
	// if err != nil {
	// 	h.logger.Errorf("Error generating TLS proof:", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
	// 	return
	// }

	// services.SendTask(tlsProof.ResponseHash, ipfsData.ProofData.ActionDataCID, jobData.TaskDefinitionID)

	// h.logger.Infof("CID: %s", ipfsData.ProofData.ActionDataCID)

	// ipfsDataBytes, err := json.Marshal(ipfsData)
	// if err != nil {
	// 	h.logger.Errorf("Error marshaling IPFS data:", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal IPFS data"})
	// 	return
	// }

	// c.Data(http.StatusOK, "application/octet-stream", ipfsDataBytes)
	c.JSON(http.StatusOK, gin.H{"message": "Task executed successfully"})
}

// ValidateTask handles task validation requests
func (h *Handlers) ValidateTask(c *gin.Context) {
	var taskRequest TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	h.logger.Info("Received Task Validation Request:")
	h.logger.Infof("Task Definition ID: %d", taskRequest.TaskDefinitionID)
	h.logger.Infof("Performer Address: %s", taskRequest.Performer)

	var decodedData string
	if strings.HasPrefix(taskRequest.Data, "0x") {
		dataBytes, err := hex.DecodeString(taskRequest.Data[2:])
		if err != nil {
			h.logger.Errorf("Failed to hex-decode data: %v", err)
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Data:    false,
				Error:   true,
				Message: fmt.Sprintf("Failed to decode hex data: %v", err),
			})
			return
		}
		decodedData = string(dataBytes)
		h.logger.Infof("Decoded Data: %s", decodedData)
	} else {
		decodedData = taskRequest.Data
	}

	ipfsContent, err := ipfs.FetchIPFSContent(config.GetIpfsHost(), decodedData)
	if err != nil {
		h.logger.Errorf("Failed to fetch IPFS content from ProofOfTask: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content from ProofOfTask: %v", err),
		})
		return
	}

	var ipfsData types.IPFSData
	if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
		h.logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse IPFS content: %v", err),
		})
		return
	}

	jobID := ipfsData.JobData.JobID
	executionTimestamp := ipfsData.ActionData.Timestamp
	taskID := ipfsData.ActionData.TaskID
	taskFee := ipfsData.ActionData.TotalFee

	if err := h.dbClient.UpdateTaskFeeInDatabase(taskID, taskFee); err != nil {
		h.logger.Errorf("Failed to update task fee in database: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to update task fee in database: %v", err),
		})
		return
	}

	if err := h.dbClient.UpdateJobLastExecutedTimestamp(jobID, executionTimestamp); err != nil {
		h.logger.Errorf("Failed to update job last executed timestamp in database: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to update job last executed timestamp: %v", err),
		})
		return
	}

	if err := updateJobStateInScheduler(jobID, executionTimestamp); err != nil {
		h.logger.Warnf("Failed to update job state in scheduler: %v", err)
	}

	c.JSON(http.StatusOK, ValidationResponse{
		Data:    true,
		Error:   false,
		Message: fmt.Sprintf("Successfully validated task for job ID %d", jobID),
	})
}

func updateJobStateInScheduler(jobID int64, timestamp time.Time) error {
	updateData := map[string]interface{}{
		"job_id":    jobID,
		"timestamp": timestamp,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal job state update data: %w", err)
	}

	schedulerURL := fmt.Sprintf("http://localhost:%s/job/state/update", config.GetManagerRPCPort())

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

	// logger.Infof("Successfully updated job %d state in scheduler with timestamp %v", jobID, timestamp)
	return nil
}
