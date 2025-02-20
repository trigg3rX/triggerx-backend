package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SendTaskToPerformer sends a job execution request to the performer service running on port 4003.
// It takes job and task metadata, formats them into the expected payload structure, and makes a POST request.
// Returns true if the task was successfully sent and accepted by the performer, false with error otherwise.
func SendTaskToPerformer(jobData *types.Job, triggerData *types.TriggerData) (bool, error) {
	client := &http.Client{}

	// Construct payload with task definition ID and job details required for execution
	payload := map[string]interface{}{
		"job": jobData,
		"trigger": triggerData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9002/task/execute", bytes.NewBuffer(jsonData))
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
