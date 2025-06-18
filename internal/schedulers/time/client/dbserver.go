package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type DBServerClient struct {
	baseURL    string
	httpClient *retry.HTTPClient
	logger     logging.Logger
}

type APIResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewDBServerClient(logger logging.Logger, baseURL string) (*DBServerClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("DBServerURL is required")
	}

	httpClient, err := retry.NewHTTPClient(retry.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %v", err)
	}

	return &DBServerClient{
		logger:     logger,
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

// GetTimeBasedJobs fetches jobs that need to be executed in the next window
func (c *DBServerClient) GetTimeBasedJobs() ([]types.ScheduleTimeTaskData, error) {
	url := fmt.Sprintf("%s/api/jobs/time", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time-based jobs: %v", err)
	}

	resp, err := c.httpClient.DoWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time-based jobs: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	var jobs []types.ScheduleTimeTaskData
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	c.logger.Debugf("Fetched %d time-based jobs", len(jobs))
	return jobs, nil
}

// UpdateJobNextExecution updates the next execution timestamp for a job
func (c *DBServerClient) UpdateJobNextExecution(jobID int64, nextExecution time.Time) error {
	url := fmt.Sprintf("%s/api/jobs/%d/lastexecuted", c.baseURL, jobID)

	payload := map[string]string{
		"next_execution_timestamp": nextExecution.Format(time.RFC3339),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	_, err = c.httpClient.DoWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to update job %d next execution: %v", jobID, err)
	}

	c.logger.Debugf("Updated job %d next execution to %v", jobID, nextExecution)
	return nil
}

// UpdateJobStatus updates the status of a job
func (c *DBServerClient) UpdateJobStatus(jobID int64, status bool) error {
	url := fmt.Sprintf("%s/api/jobs/%d/status/%t", c.baseURL, jobID, status)

	payload := map[string]bool{
		"status": status,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	_, err = c.httpClient.DoWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to update job %d status: %v", jobID, err)
	}

	c.logger.Debugf("Updated job %d status to %v", jobID, status)
	return nil
}

// HealthCheck checks if the database server is healthy
func (c *DBServerClient) HealthCheck() error {
	url := fmt.Sprintf("%s/api/health", c.baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	resp, err := c.httpClient.DoWithRetry(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status code %d", resp.StatusCode)
	}

	return nil
}

// Close closes the HTTP client
func (c *DBServerClient) Close() {
	c.httpClient.Close()
	c.logger.Debug("Database client closed")
}
