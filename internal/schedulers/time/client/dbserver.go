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
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type DBServerClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logging.Logger
	config     Config
}

type Config struct {
	DBServerURL    string
	RequestTimeout time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
}

type APIResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewDBServerClient(logger logging.Logger, config Config) (*DBServerClient, error) {
	if config.DBServerURL == "" {
		return nil, fmt.Errorf("DBServerURL is required")
	}

	// Set defaults
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	// Configure HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
	}

	return &DBServerClient{
		logger:  logger,
		config:  config,
		baseURL: config.DBServerURL,
		httpClient: &http.Client{
			Timeout:   config.RequestTimeout,
			Transport: transport,
		},
	}, nil
}

// GetTimeBasedJobs fetches jobs that need to be executed in the next window
func (c *DBServerClient) GetTimeBasedJobs() ([]types.ScheduleTimeJobData, error) {
	url := fmt.Sprintf("%s/api/jobs/time", c.baseURL)

	var jobs []types.ScheduleTimeJobData
	err := c.doWithRetry("GET", url, nil, &jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time-based jobs: %v", err)
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

	err := c.doWithRetry("PUT", url, payload, nil)
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

	err := c.doWithRetry("PUT", url, payload, nil)
	if err != nil {
		return fmt.Errorf("failed to update job %d status: %v", jobID, err)
	}

	c.logger.Debugf("Updated job %d status to %v", jobID, status)
	return nil
}

// doWithRetry performs HTTP requests with retry logic
func (c *DBServerClient) doWithRetry(method, url string, payload interface{}, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debugf("Retrying request (attempt %d/%d) after %v", attempt, c.config.MaxRetries, c.config.RetryDelay)
			time.Sleep(c.config.RetryDelay)
		}

		err := c.doRequest(method, url, payload, result)
		if err == nil {
			return nil // Success
		}

		lastErr = err
		c.logger.Warnf("Request failed (attempt %d/%d): %v", attempt+1, c.config.MaxRetries+1, err)
	}

	return fmt.Errorf("request failed after %d attempts: %v", c.config.MaxRetries+1, lastErr)
}

// doRequest performs a single HTTP request
func (c *DBServerClient) doRequest(method, url string, payload interface{}, result interface{}) error {
	var body io.Reader

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "triggerx-scheduler/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResp APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Error != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, apiResp.Error)
		}
		return fmt.Errorf("HTTP error: status code %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response if result is provided
	if result != nil {
		// Try to parse as API response first
		var apiResp APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Data != nil {
			// Re-marshal the data field and unmarshal into result
			dataBytes, err := json.Marshal(apiResp.Data)
			if err != nil {
				return fmt.Errorf("failed to marshal API response data: %v", err)
			}
			if err := json.Unmarshal(dataBytes, result); err != nil {
				return fmt.Errorf("failed to unmarshal API response data: %v", err)
			}
		} else {
			// Try to parse directly
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %v", err)
			}
		}
	}

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

	resp, err := c.httpClient.Do(req)
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
	c.httpClient.CloseIdleConnections()
	c.logger.Debug("Database client closed")
}
