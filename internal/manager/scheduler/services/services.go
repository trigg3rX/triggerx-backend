package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.ManagerProcess)

// SendTaskToPerformer sends a job execution request to the performer service running on port 4003.
// It takes job and task metadata, formats them into the expected payload structure, and makes a POST request.
// Returns true if the task was successfully sent and accepted by the performer, false with error otherwise.
func SendTaskToPerformer(jobData *types.HandleCreateJobData, triggerData *types.TriggerData, connectionAddress string) (bool, error) {
	client := &http.Client{}

	payload := map[string]interface{}{
		"job":     jobData,
		"trigger": triggerData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", connectionAddress, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("performer returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return true, nil
}

func GetPerformerData() (types.GetPerformerData, error) {
	// url := "https://data.triggerx.network/keepers/performers"
	url := "http://localhost:8080/api/keepers/performers"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.GetPerformerData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.GetPerformerData{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var performers []types.GetPerformerData
	if err := json.NewDecoder(resp.Body).Decode(&performers); err != nil {
		return types.GetPerformerData{}, fmt.Errorf("failed to decode performers: %w", err)
	}

	if len(performers) == 0 {
		return types.GetPerformerData{}, fmt.Errorf("no performers available")
	}

	randomIndex := time.Now().UnixNano() % int64(len(performers))
	selectedPerformer := performers[randomIndex]

	logger.Infof("Selected performer ID: %v", selectedPerformer.KeeperID)

	return selectedPerformer, nil
}

func GetJobDetails(jobID int64) (types.HandleCreateJobData, error) {
	url := fmt.Sprintf("http://localhost:8080/api/keepers/jobs/%d", jobID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.HandleCreateJobData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.HandleCreateJobData{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return types.HandleCreateJobData{}, fmt.Errorf("job details returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// First decode into JobData structure
	var jobData types.JobData
	if err := json.NewDecoder(resp.Body).Decode(&jobData); err != nil {
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

	return handleCreateJobData, nil
}

func CreateTaskData(taskData *types.CreateTaskData) (int64, bool, error) {
	// url := "https://data.triggerx.network/keepers/tasks"
	url := "http://localhost:8080/api/tasks"

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return 0, false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response types.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, false, fmt.Errorf("failed to decode task ID: %w", err)
	}

	return response.TaskID, true, nil
}
