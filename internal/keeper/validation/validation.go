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
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

type JobValidator struct {
	logger    Logger
	ethClient *ethclient.Client
}

func NewJobValidator(logger Logger, ethClient *ethclient.Client) *JobValidator {
	return &JobValidator{
		logger:    logger,
		ethClient: ethClient,
	}
}

func (v *JobValidator) ValidateTimeBasedJob(job *jobtypes.HandleCreateJobData) (bool, error) {
	const timeTolerance = 1500 * time.Millisecond

	if job.TaskDefinitionID != 1 && job.TaskDefinitionID != 2 {
		return false, fmt.Errorf("not a time-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating time-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	if job.TimeInterval <= 0 {
		return false, fmt.Errorf("invalid time interval: %d (must be positive)", job.TimeInterval)
	}

	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	now := time.Now().UTC()

	if job.LastExecutedAt.IsZero() {
		if job.TimeFrame > 0 {
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

	nextExecution := job.LastExecutedAt.Add(time.Duration(job.TimeInterval) * time.Second).Add(-timeTolerance)

	if now.Before(nextExecution) {
		v.logger.Infof("Not enough time has passed for job %d. Last executed: %s, next execution: %s (with %v tolerance)",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339), nextExecution.Add(timeTolerance).Format(time.RFC3339), timeTolerance)
		return false, nil
	}

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
	if job.TaskDefinitionID != 3 && job.TaskDefinitionID != 4 {
		return false, fmt.Errorf("not an event-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating event-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	if job.TriggerContractAddress == "" {
		return false, fmt.Errorf("missing TriggerContractAddress for event-based job %d", job.JobID)
	}

	if job.TriggerEvent == "" {
		return false, fmt.Errorf("missing TriggerEvent for event-based job %d", job.JobID)
	}

	v.logger.Infof("Validating contract address '%s' and event '%s'", job.TriggerContractAddress, job.TriggerEvent)

	if !common.IsHexAddress(job.TriggerContractAddress) {
		return false, fmt.Errorf("invalid Ethereum contract address: %s", job.TriggerContractAddress)
	}

	contractCode, err := v.ethClient.CodeAt(context.Background(), common.HexToAddress(job.TriggerContractAddress), nil)
	if err != nil {
		return false, fmt.Errorf("failed to check contract existence: %v", err)
	}

	if len(contractCode) == 0 {
		return false, fmt.Errorf("no contract found at address: %s", job.TriggerContractAddress)
	}

	_, err = v.fetchContractABI(job.TriggerContractAddress)
	if err != nil {
		v.logger.Warnf("Could not fetch ABI for contract %s: %v", job.TriggerContractAddress, err)
	}

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
	blockscoutUrl := fmt.Sprintf(
		"https://optimism-sepolia.blockscout.com/api?module=contract&action=getabi&address=%s",
		contractAddress)

	resp, err := http.Get(blockscoutUrl)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var response struct {
				Status  string `json:"status"`
				Message string `json:"message"`
				Result  string `json:"result"`
			}

			err = json.Unmarshal(body, &response)
			if err == nil && response.Status == "1" {
				return response.Result, nil
			}
		}
	}
	OptimismAPIKey := os.Getenv("OPTIMISM_API_KEY")
	if OptimismAPIKey == "" {
		return "", fmt.Errorf("OPTIMISM environment variable not set")
	}

	etherscanUrl := fmt.Sprintf(
		"https://api-sepolia-optimism.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s",
		contractAddress, OptimismAPIKey)

	resp, err = http.Get(etherscanUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch ABI from both APIs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch ABI from both APIs, Etherscan status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Etherscan response body: %v", err)
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse Etherscan JSON response: %v", err)
	}

	if response.Status != "1" {
		return "", fmt.Errorf("error from both APIs, Etherscan error: %s", response.Message)
	}

	return response.Result, nil
}

func (v *JobValidator) ValidateConditionBasedJob(job *jobtypes.HandleCreateJobData) (bool, error) {
	if job.TaskDefinitionID != 5 && job.TaskDefinitionID != 6 {
		return false, fmt.Errorf("not a condition-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating condition-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	if job.ScriptTriggerFunction == "" {
		return false, fmt.Errorf("missing ScriptTriggerFunction for condition-based job %d", job.JobID)
	}

	v.logger.Infof("Fetching condition script from IPFS: %s", job.ScriptTriggerFunction)
	scriptContent, err := ipfs.FetchIPFSContent(config.IpfsHost, job.ScriptTriggerFunction)
	if err != nil {
		return false, fmt.Errorf("failed to fetch condition script: %v", err)
	}
	v.logger.Infof("Successfully fetched condition script for job %d", job.JobID)

	now := time.Now().UTC()
	if job.TimeFrame > 0 {
		endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second)
		if now.After(endTime) {
			v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame)
			return false, nil
		}
	}

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

	tempDir, err := ioutil.TempDir("", "condition-build")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary build directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	v.logger.Infof("Compiling condition script for job %d", job.JobID)
	outputBinary := filepath.Join(tempDir, "condition")
	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to compile condition script: %v, stderr: %s", err, stderr.String())
	}

	v.logger.Infof("Executing condition script for job %d", job.JobID)
	result := exec.Command(outputBinary)
	stdout, err := result.Output()
	if err != nil {
		return false, fmt.Errorf("failed to run condition script: %v", err)
	}

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

func (v *JobValidator) ValidateAndPrepareJob(job *jobtypes.HandleCreateJobData, triggerData *jobtypes.TriggerData) (bool, error) {
	switch job.TaskDefinitionID {
	case 1, 2:
		return v.ValidateTimeBasedJob(job)
	case 3, 4:
		return v.ValidateEventBasedJob(job, nil)
	case 5, 6:
		return v.ValidateConditionBasedJob(job)
	default:
		return false, fmt.Errorf("unsupported task definition ID: %d", job.TaskDefinitionID)
	}
}

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

	logger.Infof("Received Task Validation Request: ProofOfTask=%s, TaskDefinitionID=%d, Performer=%s", taskRequest.ProofOfTask, taskRequest.TaskDefinitionID, taskRequest.Performer)

	var decodedData string
	if strings.HasPrefix(taskRequest.Data, "0x") {
		dataBytes, err := hex.DecodeString(taskRequest.Data[2:])
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
		logger.Infof("Decoded Data CID: %s", decodedData)
	} else {
		decodedData = taskRequest.Data
	}

	ipfsContent, err := ipfs.FetchIPFSContent(config.IpfsHost, decodedData)
	if err != nil {
		logger.Errorf("Failed to fetch IPFS content from ProofOfTask: %v", err)
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content from ProofOfTask: %v", err),
		})
		return
	}

	logger.Infof("Data CID: %s", decodedData)

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

	isValid := false
	var validationErr error

	switch taskRequest.TaskDefinitionID {
	case 1, 2:
		isValid, validationErr = jobValidator.ValidateTimeBasedJob(&ipfsData.JobData)
	case 3, 4:
		isValid, validationErr = jobValidator.ValidateEventBasedJob(&ipfsData.JobData, nil)
	case 5, 6:
		if ipfsData.JobData.ScriptTriggerFunction == "" {
			logger.Warnf("Missing ScriptTriggerFunction for condition-based job %d", ipfsData.JobData.JobID)

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
