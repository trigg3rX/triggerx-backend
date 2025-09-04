package dbserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (c *DBServerClient) CreateTask(ctx context.Context, createTaskData types.CreateTaskDataRequest) (int64, error) {
	url := fmt.Sprintf("%s/api/tasks", c.dbserverUrl)

	jsonPayload, err := json.Marshal(createTaskData)
	if err != nil {
		return -1, fmt.Errorf("failed to marshal create task data: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return -1, fmt.Errorf("failed to create task: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.DoWithRetry(ctx, req)
	if err != nil {
		return -1, fmt.Errorf("failed to create task: %v", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.logger.Errorf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return -1, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		TaskID int64 `json:"task_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return -1, fmt.Errorf("failed to decode response body: %v", err)
	}

	return response.TaskID, nil
}
