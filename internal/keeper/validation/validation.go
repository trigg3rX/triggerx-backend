package validation

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

// Logger interface for logging
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

// JobValidator handles validation of jobs
type JobValidator struct {
	logger    Logger
	ethClient *ethclient.Client
}

// NewJobValidator creates a new job validator
func NewJobValidator(logger Logger, ethClient *ethclient.Client) *JobValidator {
	return &JobValidator{
		logger:    logger,
		ethClient: ethClient,
	}
}

// ValidateTimeBasedJob checks if a time-based job (task definitions 1 and 2) should be executed
// based on its time interval, timeframe, and last execution time
func (v *JobValidator) ValidateTimeBasedJob(job *jobtypes.HandleCreateJobData) (bool, error) {
	// Define tolerance constant (3 seconds)
	const timeTolerance = 3 * time.Second

	// Ensure this is a time-based job
	if job.TaskDefinitionID != 1 && job.TaskDefinitionID != 2 {
		return false, fmt.Errorf("not a time-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating time-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	// Check for zero or negative values
	if job.TimeInterval <= 0 {
		return false, fmt.Errorf("invalid time interval: %d (must be positive)", job.TimeInterval)
	}

	// For non-recurring jobs, check if job has already been executed and shouldn't run again
	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	now := time.Now().UTC()

	// Check if this is the job's first execution
	if job.LastExecutedAt.IsZero() {
		// For first execution, check if it's within the timeframe from creation
		if job.TimeFrame > 0 {
			// Add tolerance to timeframe check
			endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second).Add(timeTolerance)
			if now.After(endTime) {
				v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds, with %v tolerance)",
					job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame, timeTolerance)
				return false, nil
			}
		}

		v.logger.Infof("Job %d is eligible for first execution", job.JobID)
		return true, nil
	}

	// Calculate the next scheduled execution time with tolerance
	nextExecution := job.LastExecutedAt.Add(time.Duration(job.TimeInterval) * time.Second).Add(-timeTolerance)
	
	// Store current time for logging but don't update job.LastExecutedAt yet
	// (this should be done by the caller after successful execution)
	//currentTime := now
	
	// Check if enough time has passed since the last execution (with tolerance)
	if now.Before(nextExecution) {
		v.logger.Infof("Not enough time has passed for job %d. Last executed: %s, next execution: %s (with %v tolerance)",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339), nextExecution.Add(timeTolerance).Format(time.RFC3339), timeTolerance)
		return false, nil
	}

	// If timeframe is set, check if job is still within its timeframe (with tolerance)
	if job.TimeFrame > 0 {
		endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second).Add(timeTolerance)
		if now.After(endTime) {
			v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds, with %v tolerance)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame, timeTolerance)
			return false, nil
		}
	}

	v.logger.Infof("Job %d is eligible for execution", job.JobID)
	return true, nil
}

func (v *JobValidator) ValidateEventBasedJob(job *jobtypes.HandleCreateJobData, ipfsData *jobtypes.IPFSData) (bool, error) {
	// Ensure this is an event-based job
	if job.TaskDefinitionID != 3 && job.TaskDefinitionID != 4 {
		return false, fmt.Errorf("not an event-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating event-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	// For non-recurring jobs, check if job has already been executed and shouldn't run again
	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	// Fetch IPFS content based on the job's proof of task
	ipfsContent, err := fetchIPFSContent(ipfsData.ActionData.IPFSDataCID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch IPFS content: %v", err)
	}

	// Extract trigger transaction hash from IPFS content
	triggerTxHash, err := ExtractTriggerTxHash(ipfsContent)
	if err != nil {
		return false, fmt.Errorf("failed to extract trigger transaction hash from IPFS: %v", err)
	}

	// Fetch event transaction hash from blockchain
	eventTxHash, err := v.fetchEventTransactionHash(job.TriggerContractAddress, job.TriggerEvent)
	if err != nil {
		return false, fmt.Errorf("failed to fetch event transaction hash: %v", err)
	}

	// Compare the transaction hashes
	if eventTxHash != triggerTxHash {
		v.logger.Infof("Event transaction hash (%s) doesn't match trigger transaction hash from IPFS (%s)",
			eventTxHash, triggerTxHash)
		return false, nil
	}

	v.logger.Infof("Transaction hash validation successful for job %d", job.JobID)

	// Check if job is within its timeframe
	now := time.Now().UTC()
	if job.TimeFrame > 0 {
		endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second)
		if now.After(endTime) {
			v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame)
			return false, nil
		}
	}

	v.logger.Infof("Event-based job %d is validated successfully", job.JobID)
	return true, nil
}

func (v *JobValidator) fetchEventTransactionHash(contractAddress, eventName string) (string, error) {
	v.logger.Infof("Fetching transaction hash for event %s from contract %s", eventName, contractAddress)

	v.logger.Infof("Fetching transaction hash for event %s from contract %s", eventName, contractAddress)

	// Create an Ethereum client if needed
	// This assumes you have access to an ethClient in the JobValidator struct
	// If not, you'd need to add it as a dependency
	if v.ethClient == nil {
		return "", fmt.Errorf("Ethereum client not initialized in job validator")
	}

	// Convert contract address string to common.Address
	contractAddr := common.HexToAddress(contractAddress)

	// Get the contract ABI to create a proper filter for the event
	abiJSON, err := v.fetchContractABI(contractAddress)
	if err != nil {
		return "", fmt.Errorf("failed to fetch contract ABI: %v", err)
	}

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return "", fmt.Errorf("failed to parse contract ABI: %v", err)
	}

	// Find the event in the ABI
	eventABI, exist := parsedABI.Events[eventName]
	if !exist {
		return "", fmt.Errorf("event %s not found in contract ABI", eventName)
	}

	// Create a filter query for the event
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{eventABI.ID}},
		// Limit to recent blocks, adjust as needed
		FromBlock: big.NewInt(0).Sub(nil, big.NewInt(10000)), // Last 10000 blocks
		ToBlock:   nil,                                       // Latest block
	}

	// Query the logs
	logs, err := v.ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		return "", fmt.Errorf("failed to filter logs: %v", err)
	}

	if len(logs) == 0 {
		return "", fmt.Errorf("no logs found for event %s", eventName)
	}

	// Get the most recent event (last in the array)
	latestLog := logs[len(logs)-1]

	v.logger.Infof("Found transaction hash for event %s: %s", eventName, latestLog.TxHash.Hex())

	return latestLog.TxHash.Hex(), nil
}

func (v *JobValidator) fetchContractABI(contractAddress string) (string, error) {
	// Example using Blockscout API for Optimism Sepolia network
	blockscoutUrl := fmt.Sprintf(
		"https://optimism-sepolia.blockscout.com/api?module=contract&action=getabi&address=%s",
		contractAddress)

	resp, err := http.Get(blockscoutUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch ABI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch ABI, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if response.Status != "1" {
		return "", fmt.Errorf("error from API: %s", response.Message)
	}

	return response.Result, nil
}

// ValidateAndPrepareJob validates job based on its task definition ID and prepares it for execution
func (v *JobValidator) ValidateAndPrepareJob(job *jobtypes.HandleCreateJobData, triggerData *jobtypes.TriggerData) (bool, error) {
	switch job.TaskDefinitionID {
	case 1, 2: // Time-based jobs
		return v.ValidateTimeBasedJob(job)
	case 3, 4: // Event-based jobs
		// Placeholder for event-based validation logic
		return v.ValidateEventBasedJob(job, &jobtypes.IPFSData{})
	case 5, 6: // Condition-based jobs
		// Placeholder for condition-based validation logic
		return true, nil
	default:
		return false, fmt.Errorf("unsupported task definition ID: %d", job.TaskDefinitionID)
	}
}

// HTTP API related validation structs and handlers

type ProofResponse struct {
	ProofHash string `json:"proofHash"`
	CID       string `json:"cid"`
}

type ValidationResult struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

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
	logger.Infof("Proof of Task: %s", taskRequest.ProofOfTask)
	logger.Infof("Data: %s", taskRequest.Data)
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
	ipfsContent, err := fetchIPFSContent(decodedData)
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
	logger.Infof("Data CID: %s", decodedData)

	// No need to fetch the data content separately since we're already getting
	// the complete IPFSData from the ProofOfTask content

	// Parse IPFS data into IPFSData struct
	var ipfsData jobtypes.IPFSData
	if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
		logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse IPFS content: %v", err),
		})
		return
	}

	// Create a job validator
	ethClient, err := ethclient.Dial("https://opt-sepolia.g.alchemy.com/v2/E3OSaENxCMNoRBi_quYcmTNPGfRitxQa")
	if err != nil {
		logger.Errorf("Failed to connect to Ethereum client: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to connect to Ethereum client: %v", err),
		})
		return
	}
	defer ethClient.Close()

	jobValidator := NewJobValidator(logger, ethClient)

	// Validate job based on task definition ID
	isValid := false
	var validationErr error

	switch taskRequest.TaskDefinitionID {
	case 1, 2: // Time-based jobs
		isValid, validationErr = jobValidator.ValidateTimeBasedJob(&ipfsData.JobData)
	case 3, 4: // Event-based jobs
		isValid, validationErr = jobValidator.ValidateEventBasedJob(&ipfsData.JobData, &ipfsData)
	case 5, 6: // Condition-based jobs
		// For future implementation
		isValid = true
	default:
		validationErr = fmt.Errorf("unsupported task definition ID: %d", taskRequest.TaskDefinitionID)
	}

	if validationErr != nil {
		logger.Errorf("Validation error: %v", validationErr)
		c.JSON(http.StatusOK, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: validationErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ValidationResponse{
		Data:    isValid,
		Error:   false,
		Message: "",
	})
}

func fetchIPFSContent(cid string) (string, error) {
	ipfsGateway := "https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/"
	resp, err := http.Get(ipfsGateway + cid)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IPFS content: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch IPFS content: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

func ExtractTriggerTxHash(ipfsContent string) (string, error) {
	// Parse the JSON content into a map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(ipfsContent), &data); err != nil {
		return "", fmt.Errorf("failed to parse IPFS content: %v", err)
	}

	// Try to navigate the structure to find trigger_data.trigger_tx_hash
	triggerData, ok := data["trigger_data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("trigger_data not found or not an object")
	}

	triggerTxHash, ok := triggerData["trigger_tx_hash"].(string)
	if !ok {
		return "", fmt.Errorf("trigger_tx_hash not found or not a string")
	}

	return triggerTxHash, nil
}
