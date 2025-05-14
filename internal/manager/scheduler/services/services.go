package services

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

var logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
var lastIndex int

type Params struct {
	proofOfTask      string
	data             string
	taskDefinitionId int
	performerAddress string
	signature        string
}

func SendTaskToPerformer(jobData *types.HandleCreateJobData, triggerData *types.TriggerData, performerData types.GetPerformerData) (bool, error) {
	logger.Debugf("Performer %d", performerData.KeeperID)

	privateKey, err := crypto.HexToECDSA(config.DeployerPrivateKey)
	if err != nil {
		logger.Errorf("Error converting private key", "error", err)
	}
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		logger.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	performerAddress := crypto.PubkeyToAddress(*publicKey).Hex()

	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	data := map[string]interface{}{
		"jobData":       jobData,
		"triggerData":   triggerData,
		"performerData": performerData,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("Error marshalling data", "error", err)
		return false, err
	}

	dataPacked, err := arguments.Pack(
		"proofOfTask",
		[]byte(jsonData),
		common.HexToAddress(performerAddress),
		big.NewInt(int64(0)),
	)
	if err != nil {
		logger.Errorf("Error encoding data", "error", err)
		return false, err
	}
	messageHash := crypto.Keccak256Hash(dataPacked)

	sig, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		logger.Errorf("Error signing message", "error", err)
		return false, err
	}
	sig[64] += 27
	serializedSignature := hexutil.Encode(sig)
	logger.Infof("Serialized signature: %s", serializedSignature)

	client, err := rpc.Dial(config.AggregatorRPCAddress)
	if err != nil {
		logger.Errorf("Error dialing RPC", "error", err)
		return false, err
	}

	params := Params{
		proofOfTask:      "proofOfTask",
		data:             "0x" + hex.EncodeToString(jsonData),
		taskDefinitionId: 0,
		performerAddress: performerAddress,
		signature:        serializedSignature,
	}

	var result interface{}

	err = client.Call(&result, "sendCustomMessage", params.data, params.taskDefinitionId)
	if err != nil {
		logger.Errorf("Error making RPC request: %v", err)
	}

	logger.Infof("API response: %v", result)
	return true, nil
}

func GetPerformer() (types.GetPerformerData, error) {
	url := fmt.Sprintf("%s/api/keepers/performers", config.DatabaseRPCAddress)

	logger.Debugf("Fetching performer data from %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorf("Failed to create request for performer data: %v", err)
		return types.GetPerformerData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send request for performer data: %v", err)
		return types.GetPerformerData{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("API returned non-200 status code for performer data: %d, body: %s",
			resp.StatusCode, string(body))
		return types.GetPerformerData{}, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(body))
	}

	var performers []types.GetPerformerData
	if err := json.NewDecoder(resp.Body).Decode(&performers); err != nil {
		logger.Errorf("Failed to decode performers: %v", err)
		return types.GetPerformerData{}, fmt.Errorf("failed to decode performers: %w", err)
	}

	logger.Infof("Found %d valid performers", len(performers))
	if len(performers) == 0 {
		return types.GetPerformerData{}, fmt.Errorf("no performers available")
	}

	var selectedPerformer types.GetPerformerData

	nextIndex := 0
	if config.FoundNextPerformer {
		nextIndex = (lastIndex + 1) % len(performers)
	}

	selectedPerformer = performers[nextIndex]

	logger.Debugf("Selected performer at index %d with ID %d",
		nextIndex, selectedPerformer.KeeperID)

	lastIndex = nextIndex
	config.FoundNextPerformer = true

	logger.Infof("Selected performer ID: %v with address: %s",
		selectedPerformer.KeeperID, selectedPerformer.KeeperAddress)

	return selectedPerformer, nil
}

func GetJobDetails(jobID int64) (types.HandleCreateJobData, error) {
	url := fmt.Sprintf("%s/api/keepers/jobs/%d", config.DatabaseRPCAddress, jobID)

	logger.Debugf("Fetching job details for job %d from %s", jobID, url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorf("Failed to create request for job %d details: %v", jobID, err)
		return types.HandleCreateJobData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send request for job %d details: %v", jobID, err)
		return types.HandleCreateJobData{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("Job details returned non-200 status code for job %d: %d, body: %s",
			jobID, resp.StatusCode, string(body))
		return types.HandleCreateJobData{}, fmt.Errorf("job details returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var jobData types.JobData
	if err := json.NewDecoder(resp.Body).Decode(&jobData); err != nil {
		logger.Errorf("Failed to decode job details for job %d: %v", jobID, err)
		return types.HandleCreateJobData{}, fmt.Errorf("failed to decode job details: %w", err)
	}

	handleCreateJobData := types.HandleCreateJobData{
		JobID:                  jobData.JobID,
		TaskDefinitionID:       jobData.TaskDefinitionID,
		UserID:                 jobData.UserID,
		Priority:               jobData.Priority,
		Security:               jobData.Security,
		LinkJobID:              jobData.LinkJobID,
		ChainStatus:            jobData.ChainStatus,
		TimeFrame:              jobData.TimeFrame,
		Recurring:              jobData.Recurring,
		TimeInterval:           jobData.TimeInterval,
		TriggerChainID:         jobData.TriggerChainID,
		TriggerContractAddress: jobData.TriggerContractAddress,
		TriggerEvent:           jobData.TriggerEvent,
		ScriptIPFSUrl:          jobData.ScriptIPFSUrl,
		ScriptTriggerFunction:  jobData.ScriptTriggerFunction,
		TargetChainID:          jobData.TargetChainID,
		TargetContractAddress:  jobData.TargetContractAddress,
		TargetFunction:         jobData.TargetFunction,
		ArgType:                jobData.ArgType,
		Arguments:              jobData.Arguments,
		ScriptTargetFunction:   jobData.ScriptTargetFunction,
		CreatedAt:              jobData.CreatedAt,
		LastExecutedAt:         jobData.LastExecutedAt,
	}

	logger.Debugf("Successfully retrieved job details for job %d", jobID)
	return handleCreateJobData, nil
}

func CreateTaskData(taskData *types.CreateTaskData) (int64, bool, error) {
	url := fmt.Sprintf("%s/api/tasks", config.DatabaseRPCAddress)

	logger.Debugf("Creating task for job %d with performer %d", taskData.JobID, taskData.TaskPerformerID)

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		logger.Errorf("Failed to marshal task data for job %d: %v", taskData.JobID, err)
		return 0, false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Failed to create request for task creation for job %d: %v", taskData.JobID, err)
		return 0, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send request for task creation for job %d: %v", taskData.JobID, err)
		return 0, false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("Task creation returned non-success status code for job %d: %d, body: %s",
			taskData.JobID, resp.StatusCode, string(body))
		return 0, false, fmt.Errorf("task creation returned status code %d: %s", resp.StatusCode, string(body))
	}

	var response types.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Errorf("Failed to decode task creation response for job %d: %v", taskData.JobID, err)
		return 0, false, fmt.Errorf("failed to decode task ID: %w", err)
	}

	logger.Infof("Successfully created task %d for job %d with performer %d",
		response.TaskID, taskData.JobID, taskData.TaskPerformerID)
	return response.TaskID, true, nil
}
