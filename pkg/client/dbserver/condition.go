package dbserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (c *DBServerClient) CreateTask(createTaskData types.CreateTaskRequest) (types.CreateTaskResponse, error) {
	url := fmt.Sprintf("%s/api/tasks", c.dbserverUrl)

	jsonPayload, err := json.Marshal(createTaskData)
	if err != nil {
		return types.CreateTaskResponse{}, fmt.Errorf("failed to marshal create task data: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return types.CreateTaskResponse{}, fmt.Errorf("failed to create task: %v", err)
	}

	resp, err := c.httpClient.DoWithRetry(req)
	if err != nil {
		return types.CreateTaskResponse{}, fmt.Errorf("failed to create task: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.CreateTaskResponse{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var createTaskResponse types.CreateTaskResponse
	err = json.Unmarshal(body, &createTaskResponse)
	if err != nil {
		return types.CreateTaskResponse{}, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return createTaskResponse, nil
}