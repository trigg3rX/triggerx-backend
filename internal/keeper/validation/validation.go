package validation

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	const timeTolerance = 1500 * time.Millisecond

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
			// endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second).Add(timeTolerance)
			// if now.After(endTime) {
			// 	v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds, with %v tolerance)",
			// 		job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame, timeTolerance)
			// 	return false, nil
			// }
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
		if job.LastExecutedAt.After(endTime) {
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

	// Check if TriggerContractAddress and TriggerEvent are provided
	if job.TriggerContractAddress == "" {
		return false, fmt.Errorf("missing TriggerContractAddress for event-based job %d", job.JobID)
	}

	if job.TriggerEvent == "" {
		return false, fmt.Errorf("missing TriggerEvent for event-based job %d", job.JobID)
	}

	v.logger.Infof("Validating contract address '%s' and event '%s'", job.TriggerContractAddress, job.TriggerEvent)

	// Verify the contract address is valid
	if !common.IsHexAddress(job.TriggerContractAddress) {
		return false, fmt.Errorf("invalid Ethereum contract address: %s", job.TriggerContractAddress)
	}

	// Check if the contract exists on chain
	contractCode, err := v.ethClient.CodeAt(context.Background(), common.HexToAddress(job.TriggerContractAddress), nil)
	if err != nil {
		return false, fmt.Errorf("failed to check contract existence: %v", err)
	}

	if len(contractCode) == 0 {
		return false, fmt.Errorf("no contract found at address: %s", job.TriggerContractAddress)
	}

	// Optional: Check if the contract ABI contains the specified event
	// (This is a simple check to see if we can get the ABI, but we don't validate the event itself
	// since that would require parsing the ABI)
	_, err = v.fetchContractABI(job.TriggerContractAddress)
	if err != nil {
		v.logger.Warnf("Could not fetch ABI for contract %s: %v", job.TriggerContractAddress, err)
		// We don't fail validation just because we can't fetch the ABI
	}

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

	v.logger.Infof("Event-based job %d validated successfully", job.JobID)
	return true, nil
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

// ValidateConditionBasedJob validates a condition-based job by executing the condition script
// and checking if it returns true
func (v *JobValidator) ValidateConditionBasedJob(job *jobtypes.HandleCreateJobData) (bool, error) {
	// Ensure this is a condition-based job
	if job.TaskDefinitionID != 5 && job.TaskDefinitionID != 6 {
		return false, fmt.Errorf("not a condition-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating condition-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	// For non-recurring jobs, check if job has already been executed and shouldn't run again
	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	// Check if the ScriptTriggerFunction is provided
	if job.ScriptTriggerFunction == "" {
		return false, fmt.Errorf("missing ScriptTriggerFunction for condition-based job %d", job.JobID)
	}

	// Fetch and execute the condition script
	v.logger.Infof("Fetching condition script from IPFS: %s", job.ScriptTriggerFunction)
	scriptContent, err := fetchIPFSContent(job.ScriptTriggerFunction)
	if err != nil {
		return false, fmt.Errorf("failed to fetch condition script: %v", err)
	}
	v.logger.Infof("Successfully fetched condition script for job %d", job.JobID)

	// Check if job is within its timeframe before executing script
	now := time.Now().UTC()
	if job.TimeFrame > 0 {
		endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second)
		if now.After(endTime) {
			v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame)
			return false, nil
		}
	}

	// Create a temporary file for the script
	tempFile, err := ioutil.TempFile("", "condition-*.go")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
		return false, fmt.Errorf("failed to write script to file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		return false, fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Create a temp directory for the script's build output
	tempDir, err := ioutil.TempDir("", "condition-build")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary build directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Compile the script
	v.logger.Infof("Compiling condition script for job %d", job.JobID)
	outputBinary := filepath.Join(tempDir, "condition")
	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to compile condition script: %v, stderr: %s", err, stderr.String())
	}

	// Run the compiled script
	v.logger.Infof("Executing condition script for job %d", job.JobID)
	result := exec.Command(outputBinary)
	stdout, err := result.Output()
	if err != nil {
		return false, fmt.Errorf("failed to run condition script: %v", err)
	}

	// Parse the output to determine if condition is satisfied
	// Look for a line containing "Condition satisfied: true" or "Condition satisfied: false"
	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Condition satisfied: true") {
			v.logger.Infof("Condition script reported satisfaction for job %d", job.JobID)
			return true, nil
		} else if strings.Contains(line, "Condition satisfied: false") {
			v.logger.Infof("Condition script reported non-satisfaction for job %d", job.JobID)
			return false, nil
		}
	}

	// If no explicit condition found, try parsing as JSON
	var conditionResult struct {
		Satisfied bool `json:"satisfied"`
	}
	if err := json.Unmarshal(stdout, &conditionResult); err != nil {
		v.logger.Warnf("Could not determine condition result from output for job %d: %s", job.JobID, string(stdout))
		return false, fmt.Errorf("could not determine condition result from output: %s", string(stdout))
	}

	v.logger.Infof("Condition script for job %d returned satisfied: %v", job.JobID, conditionResult.Satisfied)
	return conditionResult.Satisfied, nil
}

// ValidateAndPrepareJob validates job based on its task definition ID and prepares it for execution
func (v *JobValidator) ValidateAndPrepareJob(job *jobtypes.HandleCreateJobData, triggerData *jobtypes.TriggerData) (bool, error) {
	switch job.TaskDefinitionID {
	case 1, 2: // Time-based jobs
		return v.ValidateTimeBasedJob(job)
	case 3, 4: // Event-based jobs
		return v.ValidateEventBasedJob(job, nil)
	case 5, 6: // Condition-based jobs
		return v.ValidateConditionBasedJob(job)
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
		isValid, validationErr = jobValidator.ValidateEventBasedJob(&ipfsData.JobData, nil)
	case 5, 6: // Condition-based jobs
		// For condition-based jobs, make sure we have the ScriptTriggerFunction
		if ipfsData.JobData.ScriptTriggerFunction == "" {
			logger.Warnf("Missing ScriptTriggerFunction for condition-based job %d", ipfsData.JobData.JobID)

			// Try to extract from trigger data if available
			scriptURL, ok := ipfsData.TriggerData.ConditionParams["script_url"].(string)
			if ok && scriptURL != "" {
				logger.Infof("Found script URL in TriggerData.ConditionParams: %s", scriptURL)
				ipfsData.JobData.ScriptTriggerFunction = scriptURL
			} else {
				validationErr = fmt.Errorf("missing ScriptTriggerFunction for condition-based job")
				break
			}
		}

		logger.Infof("Validating condition-based job with script: %s", ipfsData.JobData.ScriptTriggerFunction)
		isValid, validationErr = jobValidator.ValidateConditionBasedJob(&ipfsData.JobData)
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

func fetchIPFSContent(cidOrUrl string) (string, error) {
	// Determine if input is a full URL or just a CID
	var requestUrl string
	if strings.HasPrefix(cidOrUrl, "http://") || strings.HasPrefix(cidOrUrl, "https://") {
		// Input is already a full URL
		requestUrl = cidOrUrl
	} else {
		// Input is a CID, prepend the IPFS gateway
		ipfsGateway := "https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/"
		requestUrl = ipfsGateway + cidOrUrl
	}

	resp, err := http.Get(requestUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IPFS content: %v", err)
	}
	defer resp.Body.Close()

	// Rest of the function remains the same
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch IPFS content: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}
