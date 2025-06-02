package client

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

// HealthCheck performs a health check against the database server
func (c *DBServerClient) HealthCheck() error {
	url := fmt.Sprintf("%s/status", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %s", resp.Status)
	}

	return nil
}

// GetConditionBasedJobs retrieves condition-based jobs from the database
func (c *DBServerClient) GetConditionBasedJobs() ([]types.ConditionJobData, error) {
	url := fmt.Sprintf("%s/api/v1/jobs/condition", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

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
	url := fmt.Sprintf("%s/api/v1/jobs/%d/status", c.baseURL, jobID)

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	return nil
}

// SendTaskToManager sends a task to the manager for execution
func (c *DBServerClient) SendTaskToManager(jobID int64, triggerValue float64, conditionType string) error {
	url := fmt.Sprintf("%s/api/v1/tasks/create", c.baseURL)

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

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
