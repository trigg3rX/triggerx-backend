package dbserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetTimeBasedTasks fetches tasks that need to be executed in the next window
func (c *DBServerClient) GetTimeBasedTasks() ([]types.ScheduleTimeTaskData, error) {
	url := fmt.Sprintf("%s/api/jobs/time", c.dbserverUrl)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time-based tasks: %v", err)
	}

	resp, err := c.httpClient.DoWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time-based tasks: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var tasks []types.ScheduleTimeTaskData
	err = json.Unmarshal(body, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	c.logger.Debugf("Fetched %d time-based tasks", len(tasks))
	return tasks, nil
}
