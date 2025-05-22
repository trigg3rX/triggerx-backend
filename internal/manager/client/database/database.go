package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Common errors
var (
	ErrNoPerformersAvailable = fmt.Errorf("no performers available")
	ErrInvalidResponse       = fmt.Errorf("invalid response from database service")
	ErrRequestFailed         = fmt.Errorf("request to database service failed")
)

// DatabaseClientConfig holds the configuration for DatabaseClient
type DatabaseClientConfig struct {
	RPCAddress  string
	HTTPTimeout time.Duration
}

// DatabaseClient handles communication with the database service
type DatabaseClient struct {
	logger     logging.Logger
	httpClient *http.Client
	config     DatabaseClientConfig
	lastIndex  int
	mu         sync.Mutex // Add mutex for thread safety
}

// NewDatabaseClient creates a new instance of DatabaseClient
func NewDatabaseClient(logger logging.Logger, cfg DatabaseClientConfig) (*DatabaseClient, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg.RPCAddress == "" {
		return nil, fmt.Errorf("RPC address cannot be empty")
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}

	// Create a transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	return &DatabaseClient{
		logger: logger,
		httpClient: &http.Client{
			Timeout:   cfg.HTTPTimeout,
			Transport: transport,
		},
		config:    cfg,
		lastIndex: 0,
	}, nil
}

// Close cleans up any resources used by the client
func (c *DatabaseClient) Close() error {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
	return nil
}

func (c *DatabaseClient) doRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.config.RPCAddress, endpoint)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrRequestFailed, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to send request: %v", ErrRequestFailed, err)
	}

	return resp, nil
}

func (c *DatabaseClient) GetPerformer() (types.GetPerformerData, error) {
	c.logger.Debug("Fetching performer data")

	resp, err := c.doRequest("GET", "/api/keepers/performers", nil)
	if err != nil {
		c.logger.Error("Failed to get performers", "error", err)
		return types.GetPerformerData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("API returned non-200 status code",
			"status", resp.StatusCode,
			"body", string(body))
		return types.GetPerformerData{}, fmt.Errorf("%w: status code %d", ErrInvalidResponse, resp.StatusCode)
	}

	var performers []types.GetPerformerData
	if err := json.NewDecoder(resp.Body).Decode(&performers); err != nil {
		c.logger.Error("Failed to decode performers response", "error", err)
		return types.GetPerformerData{}, fmt.Errorf("%w: failed to decode response: %v", ErrInvalidResponse, err)
	}

	if len(performers) == 0 {
		c.logger.Warn("No performers available")
		return types.GetPerformerData{}, ErrNoPerformersAvailable
	}

	selectedPerformer := c.selectNextPerformer(performers)

	c.logger.Info("Selected performer",
		"id", selectedPerformer.KeeperID,
		"address", selectedPerformer.KeeperAddress)

	return selectedPerformer, nil
}

func (c *DatabaseClient) selectNextPerformer(performers []types.GetPerformerData) types.GetPerformerData {
	c.mu.Lock()
	defer c.mu.Unlock()

	nextIndex := 0
	if config.GetFoundNextPerformer() {
		nextIndex = (c.lastIndex + 1) % len(performers)
	}

	selectedPerformer := performers[nextIndex]
	c.lastIndex = nextIndex
	config.SetFoundNextPerformer(true)

	return selectedPerformer
}

func (c *DatabaseClient) GetJobDetails(jobID int64) (types.HandleCreateJobData, error) {
	c.logger.Debug("Fetching job details", "jobID", jobID)

	resp, err := c.doRequest("GET", fmt.Sprintf("/api/keepers/jobs/%d", jobID), nil)
	if err != nil {
		c.logger.Error("Failed to get job details", "jobID", jobID, "error", err)
		return types.HandleCreateJobData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Job details request failed",
			"jobID", jobID,
			"status", resp.StatusCode,
			"body", string(body))
		return types.HandleCreateJobData{}, fmt.Errorf("%w: status code %d", ErrInvalidResponse, resp.StatusCode)
	}

	// First get the basic job data
	var jobData types.JobData
	if err := json.NewDecoder(resp.Body).Decode(&jobData); err != nil {
		c.logger.Error("Failed to decode job details", "jobID", jobID, "error", err)
		return types.HandleCreateJobData{}, fmt.Errorf("%w: failed to decode response: %v", ErrInvalidResponse, err)
	}

	// Create base handleCreateJobData with common fields
	handleCreateJobData := types.HandleCreateJobData{
		JobID:             jobData.JobID,
		TaskDefinitionID:  jobData.TaskDefinitionID,
		UserID:            jobData.UserID,
		LinkJobID:         jobData.LinkJobID,
		ChainStatus:       jobData.ChainStatus,
		JobTitle:          jobData.JobTitle,
		Custom:            jobData.Custom,
		TimeFrame:         jobData.TimeFrame,
		Recurring:         jobData.Recurring,
		Status:            jobData.Status,
		JobCostPrediction: jobData.JobCostPrediction,
	}

	// Based on task_definition_id, decode the specific job type data
	switch {
	case jobData.TaskDefinitionID == 1 || jobData.TaskDefinitionID == 2:
		// Time-based job
		var timeJobData types.TimeJobData
		if err := json.NewDecoder(resp.Body).Decode(&timeJobData); err != nil {
			c.logger.Error("Failed to decode time job details", "jobID", jobID, "error", err)
			return types.HandleCreateJobData{}, fmt.Errorf("%w: failed to decode time job response: %v", ErrInvalidResponse, err)
		}
		handleCreateJobData.TimeInterval = timeJobData.TimeInterval
		handleCreateJobData.TargetChainID = timeJobData.TargetChainID
		handleCreateJobData.TargetContractAddress = timeJobData.TargetContractAddress
		handleCreateJobData.TargetFunction = timeJobData.TargetFunction
		handleCreateJobData.ABI = timeJobData.ABI
		handleCreateJobData.ArgType = timeJobData.ArgType
		handleCreateJobData.Arguments = timeJobData.Arguments
		handleCreateJobData.ScriptIPFSUrl = timeJobData.DynamicArgumentsScriptIPFSUrl

	case jobData.TaskDefinitionID == 3 || jobData.TaskDefinitionID == 4:
		// Event-based job
		var eventJobData types.EventJobData
		if err := json.NewDecoder(resp.Body).Decode(&eventJobData); err != nil {
			c.logger.Error("Failed to decode event job details", "jobID", jobID, "error", err)
			return types.HandleCreateJobData{}, fmt.Errorf("%w: failed to decode event job response: %v", ErrInvalidResponse, err)
		}
		handleCreateJobData.TriggerChainID = eventJobData.TriggerChainID
		handleCreateJobData.TriggerContractAddress = eventJobData.TriggerContractAddress
		handleCreateJobData.TriggerEvent = eventJobData.TriggerEvent
		handleCreateJobData.TargetChainID = eventJobData.TargetChainID
		handleCreateJobData.TargetContractAddress = eventJobData.TargetContractAddress
		handleCreateJobData.TargetFunction = eventJobData.TargetFunction
		handleCreateJobData.ABI = eventJobData.ABI
		handleCreateJobData.ArgType = eventJobData.ArgType
		handleCreateJobData.Arguments = eventJobData.Arguments
		handleCreateJobData.ScriptIPFSUrl = eventJobData.DynamicArgumentsScriptIPFSUrl

	case jobData.TaskDefinitionID == 5 || jobData.TaskDefinitionID == 6:
		// Condition-based job
		var conditionJobData types.ConditionJobData
		if err := json.NewDecoder(resp.Body).Decode(&conditionJobData); err != nil {
			c.logger.Error("Failed to decode condition job details", "jobID", jobID, "error", err)
			return types.HandleCreateJobData{}, fmt.Errorf("%w: failed to decode condition job response: %v", ErrInvalidResponse, err)
		}
		handleCreateJobData.ConditionType = conditionJobData.ConditionType
		handleCreateJobData.UpperLimit = conditionJobData.UpperLimit
		handleCreateJobData.LowerLimit = conditionJobData.LowerLimit
		handleCreateJobData.ValueSourceType = conditionJobData.ValueSourceType
		handleCreateJobData.ValueSourceUrl = conditionJobData.ValueSourceUrl
		handleCreateJobData.TargetChainID = conditionJobData.TargetChainID
		handleCreateJobData.TargetContractAddress = conditionJobData.TargetContractAddress
		handleCreateJobData.TargetFunction = conditionJobData.TargetFunction
		handleCreateJobData.ABI = conditionJobData.ABI
		handleCreateJobData.ArgType = conditionJobData.ArgType
		handleCreateJobData.Arguments = conditionJobData.Arguments
		handleCreateJobData.ScriptIPFSUrl = conditionJobData.DynamicArgumentsScriptIPFSUrl
	}

	c.logger.Debug("Successfully retrieved job details", "jobID", jobID)
	return handleCreateJobData, nil
}

func (c *DatabaseClient) CreateTaskData(taskData *types.CreateTaskData) (int64, bool, error) {
	c.logger.Debug("Creating task",
		"jobID", taskData.JobID,
		"performerID", taskData.TaskPerformerID)

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		c.logger.Error("Failed to marshal task data", "error", err)
		return 0, false, fmt.Errorf("%w: failed to marshal request: %v", ErrRequestFailed, err)
	}

	resp, err := c.doRequest("POST", "/api/tasks", bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create task", "error", err)
		return 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Task creation failed",
			"status", resp.StatusCode,
			"body", string(body))
		return 0, false, fmt.Errorf("%w: status code %d", ErrInvalidResponse, resp.StatusCode)
	}

	var response types.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Error("Failed to decode task creation response", "error", err)
		return 0, false, fmt.Errorf("%w: failed to decode response: %v", ErrInvalidResponse, err)
	}

	c.logger.Info("Successfully created task",
		"taskID", response.TaskID,
		"jobID", taskData.JobID,
		"performerID", taskData.TaskPerformerID)
	return response.TaskID, true, nil
}

func (c *DatabaseClient) UpdateTaskFeeInDatabase(taskID int64, taskFee float64) error {
	c.logger.Debug("Updating task fee", "taskID", taskID, "fee", taskFee)

	requestBody, err := json.Marshal(map[string]float64{
		"fee": taskFee,
	})
	if err != nil {
		c.logger.Error("Failed to marshal task fee data", "error", err)
		return fmt.Errorf("%w: failed to marshal request: %v", ErrRequestFailed, err)
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/tasks/%d/fee", taskID), bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.Error("Failed to update task fee", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Task fee update failed",
			"taskID", taskID,
			"status", resp.StatusCode,
			"body", string(body))
		return fmt.Errorf("%w: status code %d", ErrInvalidResponse, resp.StatusCode)
	}

	c.logger.Info("Successfully updated task fee", "taskID", taskID, "fee", taskFee)
	return nil
}

func (c *DatabaseClient) UpdateJobLastExecutedTimestamp(jobID int64, timestamp time.Time) error {
	c.logger.Debug("Updating job last executed timestamp",
		"jobID", jobID,
		"timestamp", timestamp)

	requestBody, err := json.Marshal(map[string]string{
		"lastExecutedAt": timestamp.Format(time.RFC3339),
	})
	if err != nil {
		c.logger.Error("Failed to marshal timestamp data", "error", err)
		return fmt.Errorf("%w: failed to marshal request: %v", ErrRequestFailed, err)
	}

	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/jobs/%d/lastexecuted", jobID), bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.Error("Failed to update job timestamp", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Job timestamp update failed",
			"jobID", jobID,
			"status", resp.StatusCode,
			"body", string(body))
		return fmt.Errorf("%w: status code %d", ErrInvalidResponse, resp.StatusCode)
	}

	c.logger.Info("Successfully updated job timestamp",
		"jobID", jobID,
		"timestamp", timestamp)
	return nil
}
