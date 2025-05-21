package handlers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/ipfs"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskHandler handles task-related requests
type TaskHandler struct {
	logger    logging.Logger
	executor  execution.TaskExecutor
	validator validation.TaskValidator
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(logger logging.Logger, executor execution.TaskExecutor, validator validation.TaskValidator) *TaskHandler {
	return &TaskHandler{
		logger:    logger,
		executor:  executor,
		validator: validator,
	}
}

// ExecuteTask handles task execution requests
func (h *TaskHandler) ExecuteTask(c *gin.Context) {
	h.logger.Info("Executing task")

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// Decode hex data
	hexData := requestBody.Data
	if len(hexData) > 2 && hexData[:2] == "0x" {
		hexData = hexData[2:]
	}

	decodedData, err := hex.DecodeString(hexData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hex data"})
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
	var resultBytes []byte

	// Convert to proper types
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
	// h.logger.Infof("performerData: %v\n", performerData)

	if config.GetKeeperAddress() != performerData.KeeperAddress {
		h.logger.Infof("I am not the performer: %s", performerData.KeeperAddress)
		c.JSON(http.StatusOK, gin.H{"message": "I am not the performer"})
		return
	} else {
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
		h.logger.Infof("jobData: %v\n", jobData)
	
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
		h.logger.Infof("triggerData: %v\n", triggerData)
	
		// Execute task
		actionData, err := h.executor.Execute(&jobData)
		if err != nil {
			h.logger.Error("Failed to execute task", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
			return
		}
	
		// Set task ID from trigger data
		actionData.TaskID = triggerData.TaskID
	
		// Convert result to bytes
		resultBytes, err = json.Marshal(actionData)
		if err != nil {
			h.logger.Error("Failed to marshal result", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process result"})
			return
		}
	}

	c.Data(http.StatusOK, "application/octet-stream", resultBytes)
}

// ValidateTask handles task validation requests
func (h *TaskHandler) ValidateTask(c *gin.Context) {
	var taskRequest types.TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, types.ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	// Combine request info logs into one concise Infof
	h.logger.Infof("Received Task Validation Request: ProofOfTask=%s, TaskDefinitionID=%d, Performer=%s", taskRequest.ProofOfTask, taskRequest.TaskDefinitionID, taskRequest.Performer)

	// Decode the data if it's hex-encoded (with 0x prefix)
	var decodedData string
	if strings.HasPrefix(taskRequest.Data, "0x") {
		dataBytes, err := hex.DecodeString(taskRequest.Data[2:]) // Remove "0x" prefix before decoding
		if err != nil {
			h.logger.Errorf("Failed to hex-decode data: %v", err)
			c.JSON(http.StatusBadRequest, types.ValidationResponse{
				Data:    false,
				Error:   true,
				Message: fmt.Sprintf("Failed to decode hex data: %v", err),
			})
			return
		}
		decodedData = string(dataBytes)
		h.logger.Infof("Decoded Data CID: %s", decodedData)
	} else {
		decodedData = taskRequest.Data
	}

	// Fetch the ActionData from IPFS using CID from the proof of task
	ipfsContent, err := ipfs.FetchIPFSContent(config.GetIpfsHost(), decodedData)
	if err != nil {
		h.logger.Errorf("Failed to fetch IPFS content from ProofOfTask: %v", err)
		c.JSON(http.StatusInternalServerError, types.ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content from ProofOfTask: %v", err),
		})
		return
	}

	// Log the decoded data CID for context
	h.logger.Infof("Data CID: %s", decodedData)

	// Parse IPFS data into IPFSData struct
	var ipfsData types.IPFSData
	if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
		h.logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
		c.JSON(http.StatusInternalServerError, types.ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse IPFS content: %v", err),
		})
		return
	}

	// Create a job validator
	ethClient, err := ethclient.Dial("https://opt-sepolia.g.alchemy.com/v2/E3OSaENxCMNoRBi_quYcmTNPGfRitxQa")
	if err != nil {
		h.logger.Errorf("Failed to connect to Ethereum client: %v", err)
		c.JSON(http.StatusInternalServerError, types.ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to connect to Ethereum client: %v", err),
		})
		return
	}
	defer ethClient.Close()

	jobValidator := validation.NewTaskValidator(h.logger, ethClient)

	// Validate job based on task definition ID
	isValid := false
	var validationErr error

	switch taskRequest.TaskDefinitionID {
	case 1, 2: // Time-based jobs
		isValid, validationErr = jobValidator.ValidateTimeBasedTask(&ipfsData.JobData)
	case 3, 4: // Event-based jobs
		isValid, validationErr = jobValidator.ValidateEventBasedTask(&ipfsData.JobData, &ipfsData)
	case 5, 6: // Condition-based jobs
		// For condition-based jobs, make sure we have the ScriptTriggerFunction
		if ipfsData.JobData.ScriptTriggerFunction == "" {
			h.logger.Warnf("Missing ScriptTriggerFunction for condition-based job %d", ipfsData.JobData.JobID)

			// Try to extract from trigger data if available
			scriptURL, ok := ipfsData.TriggerData.ConditionParams["script_url"].(string)
			if ok && scriptURL != "" {
				h.logger.Infof("Found script URL in TriggerData.ConditionParams: %s", scriptURL)
				ipfsData.JobData.ScriptTriggerFunction = scriptURL
			} else {
				validationErr = fmt.Errorf("missing ScriptTriggerFunction for condition-based job")
				break
			}
		}

		h.logger.Infof("Validating condition-based job with script: %s", ipfsData.JobData.ScriptTriggerFunction)
		isValid, validationErr = jobValidator.ValidateConditionBasedTask(&ipfsData.JobData, &ipfsData)
	default:
		validationErr = fmt.Errorf("unsupported task definition ID: %d", taskRequest.TaskDefinitionID)
	}

	if validationErr != nil {
		h.logger.Errorf("Validation error: %v", validationErr)
		c.JSON(http.StatusOK, types.ValidationResponse{
			Data:    false,
			Error:   true,
			Message: validationErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ValidationResponse{
		Data:    isValid,
		Error:   false,
		Message: "",
	})
}
