package dbserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (c *DBServerClient) UpdateTaskExecutionData(ctx context.Context, taskExecutionData types.UpdateTaskExecutionDataRequest) (bool, error) {
	url := fmt.Sprintf("%s/api/tasks/execution/%d", c.dbserverUrl, taskExecutionData.TaskID)

	jsonPayload, err := json.Marshal(taskExecutionData)
	if err != nil {
		return false, fmt.Errorf("failed to marshal task execution data: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	_, err = c.httpClient.DoWithRetry(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to update task execution data: %v", err)
	}

	return true, nil
}
