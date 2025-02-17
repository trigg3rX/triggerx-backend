package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func SendTaskToPerformer(jobData *types.Job, taskData *types.TaskData) (bool, error) {
	client := &http.Client{}

	// Prepare the request payload matching the performer's expected structure
	payload := map[string]interface{}{
		"taskDefinitionId": taskData.TaskDefinitionID,
		"job": map[string]interface{}{
			"job_id":          jobData.JobID,
			"targetFunction":  jobData.TargetFunction,
			"arguments":       jobData.Arguments,
			"chainID":         jobData.ChainID,
			"contractAddress": jobData.TargetContractAddress,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", "http://localhost:4003/task/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("performer returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return true, nil
}


