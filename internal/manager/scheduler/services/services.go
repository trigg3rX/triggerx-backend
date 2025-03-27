// File: services/services.go

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.ManagerProcess)

// SendTaskToPerformer sends a job execution request to the performer service running on port 4003.
// It takes job and task metadata, formats them into the expected payload structure, and makes a POST request.
// Returns true if the task was successfully sent and accepted by the performer, false with error otherwise.
func SendTaskToPerformer(jobData *types.HandleCreateJobData, triggerData *types.TriggerData, connectionAddress string) (bool, error) {
	// Validate connection address
	if connectionAddress == "" {
		logger.Errorf("Cannot send task for job %d: connection address is empty", jobData.JobID)
		return false, fmt.Errorf("connection address is empty for job %d", jobData.JobID)
	}

	logger.Infof("Sending task for job %d to performer at address: %s", jobData.JobID, connectionAddress)

	// Ensure connection address has a protocol
	if !strings.HasPrefix(connectionAddress, "http://") && !strings.HasPrefix(connectionAddress, "https://") {
		connectionAddress = "http://" + connectionAddress
		logger.Debugf("Added HTTP protocol to connection address: %s", connectionAddress)
	}

	// Add the task execution endpoint to the connection address
	// connectionAddress = connectionAddress
	logger.Debugf("Final endpoint URL: %s", connectionAddress)

	client := &http.Client{
		Timeout: 10 * time.Second, // Add a timeout to prevent hanging requests
	}

	payload := map[string]interface{}{
		"job":     jobData,
		"trigger": triggerData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Failed to marshal payload for job %d: %v", jobData.JobID, err)
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	executionURL := fmt.Sprintf("%s/task/execute", connectionAddress)

	req, err := http.NewRequest("POST", executionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Failed to create request for job %d to %s: %v", jobData.JobID, connectionAddress, err)
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	logger.Debugf("Sending HTTP request to %s for job %d", connectionAddress, jobData.JobID)
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send request for job %d to %s: %v", jobData.JobID, connectionAddress, err)
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("Performer returned non-200 status code for job %d: %d, body: %s",
			jobData.JobID, resp.StatusCode, string(body))
		return false, fmt.Errorf("performer returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	logger.Infof("Successfully sent task for job %d to performer at %s", jobData.JobID, connectionAddress)
	return true, nil
}

func GetPerformerData() (types.GetPerformerData, error) {
	// url := "https://data.triggerx.network/api/keepers/performers"
	url := "http://51.21.200.252:9002/api/keepers/performers"

	logger.Debugf("Fetching performer data from %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorf("Failed to create request for performer data: %v", err)
		return types.GetPerformerData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second, // Add a timeout to prevent hanging requests
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

	logger.Debugf("Retrieved %d performers from API", len(performers))

	// Filter out performers with empty connection addresses
	var validPerformers []types.GetPerformerData
	for _, p := range performers {
		if p.ConnectionAddress != "" {
			validPerformers = append(validPerformers, p)
		} else {
			logger.Warnf("Performer ID %d has empty connection address, skipping", p.KeeperID)
		}
	}

	if len(validPerformers) == 0 {
		logger.Errorf("No performers available with valid connection addresses")
		return types.GetPerformerData{}, fmt.Errorf("no performers available with valid connection addresses")
	}

	logger.Infof("Found %d performers with valid connection addresses", len(validPerformers))

	randomIndex := time.Now().UnixNano() % int64(len(validPerformers))
	selectedPerformer := validPerformers[randomIndex]

	logger.Infof("Selected performer ID: %v with connection address: %s",
		selectedPerformer.KeeperID, selectedPerformer.ConnectionAddress)

	return selectedPerformer, nil
}

func GetJobDetails(jobID int64) (types.HandleCreateJobData, error) {
	url := fmt.Sprintf("http://51.21.200.252:9002/api/keepers/jobs/%d", jobID)

	logger.Debugf("Fetching job details for job %d from %s", jobID, url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorf("Failed to create request for job %d details: %v", jobID, err)
		return types.HandleCreateJobData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second, // Add a timeout to prevent hanging requests
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

	// First decode into JobData structure
	var jobData types.JobData
	if err := json.NewDecoder(resp.Body).Decode(&jobData); err != nil {
		logger.Errorf("Failed to decode job details for job %d: %v", jobID, err)
		return types.HandleCreateJobData{}, fmt.Errorf("failed to decode job details: %w", err)
	}

	// Convert JobData to HandleCreateJobData
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
	// url := "https://data.triggerx.network/api/tasks"
	url := "http://51.21.200.252:9002/api/tasks"

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
		Timeout: 10 * time.Second, // Add a timeout to prevent hanging requests
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
