package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Config holds the database client configuration
type Config struct {
	DBServerURL    string
	RequestTimeout time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
}

// DBServerClient represents a client for the database server
type DBServerClient struct {
	logger     logging.Logger
	httpClient *http.Client
	baseURL    string
	config     Config
}

// NewDBServerClient creates a new database server client
func NewDBServerClient(logger logging.Logger, config Config) (*DBServerClient, error) {
	client := &DBServerClient{
		logger:  logger,
		baseURL: config.DBServerURL,
		config:  config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}

	return client, nil
}

// makeRequestWithRetry makes an HTTP request with retry logic and metrics tracking
func (c *DBServerClient) makeRequestWithRetry(req *http.Request, endpoint string) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Track retry attempt
			metrics.TrackDBRetry(endpoint)
			c.logger.Warn("Retrying DB request", "endpoint", endpoint, "attempt", attempt)
			time.Sleep(c.config.RetryDelay)
		}

		// Make the request
		resp, err := c.httpClient.Do(req)

		// Track the request
		if err != nil {
			// Connection error
			metrics.TrackDBConnectionError()
			metrics.TrackDBRequest(req.Method, endpoint, "connection_error")
			lastErr = err
			continue
		}

		// Track successful connection with status code
		statusCode := fmt.Sprintf("%d", resp.StatusCode)
		metrics.TrackDBRequest(req.Method, endpoint, statusCode)

		// Return response regardless of status code - let caller handle status
		return resp, nil
	}

	// All retries exhausted
	return nil, fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// HealthCheck performs a health check against the database server
func (c *DBServerClient) HealthCheck() error {
	endpoint := "/api/health"
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.makeRequestWithRetry(req, endpoint)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %s", resp.Status)
	}

	return nil
}

// GetConditionBasedJobs retrieves condition-based jobs from the database
func (c *DBServerClient) GetConditionBasedJobs() ([]types.ConditionJobData, error) {
	endpoint := "/api/v1/jobs/condition"
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.makeRequestWithRetry(req, endpoint)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var jobs []types.ConditionJobData
	if err := json.Unmarshal(body, &jobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return jobs, nil
}

// UpdateJobStatus updates the status of a job
func (c *DBServerClient) UpdateJobStatus(jobID int64, isRunning bool) error {
	endpoint := fmt.Sprintf("/api/v1/jobs/%d/status", jobID)
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	payload := map[string]interface{}{
		"status": isRunning,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.makeRequestWithRetry(req, endpoint)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	return nil
}

// SendTaskToManager sends a task to the manager for execution
func (c *DBServerClient) SendTaskToManager(jobID int64, triggerValue float64, conditionType string) error {
	endpoint := "/api/v1/tasks/create"
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	payload := map[string]interface{}{
		"job_id":         jobID,
		"trigger_value":  triggerValue,
		"condition_type": conditionType,
		"timestamp":      time.Now().UTC(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.makeRequestWithRetry(req, endpoint)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	c.logger.Info("Task sent to manager successfully", "job_id", jobID, "trigger_value", triggerValue)
	return nil
}

// Close closes the client and cleans up resources
func (c *DBServerClient) Close() error {
	// Close HTTP client connections if needed
	c.httpClient.CloseIdleConnections()
	return nil
}
