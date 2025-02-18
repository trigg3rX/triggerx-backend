package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetJobDetails retrieves job configuration and metadata from the database.
// Maps the raw database fields into a structured Job type, including argument parsing.
func (s *JobScheduler) GetJobDetails(jobID int64) (*types.Job, error) {
	var jobData types.Job

	jobIDStr := strconv.FormatInt(jobID, 10)

	var arguments map[int]interface{}

	err := s.dbClient.Session().Query(`
		SELECT job_id, task_definition_id, priority, security, time_frame, recurring, link_job_id,
			   time_interval, trigger_chain_id, trigger_contract_address, trigger_event, script_ipfs_url,
			   script_trigger_function, target_chain_id, target_contract_address, target_function, arg_type, arguments,
			   script_target_function, created_at, last_executed
		FROM triggerx.job_data 
		WHERE jobID = ?`, jobIDStr).Scan(
		&jobData.JobID, &jobData.TaskDefinitionID, &jobData.Priority, &jobData.Security, &jobData.TimeFrame,
		&jobData.TriggerChainID, &jobData.TargetChainID, &jobData.TimeFrame,
		&jobData.TimeInterval, &jobData.TriggerContractAddress,
		&jobData.TriggerEvent, &jobData.TargetContractAddress,
		&jobData.TargetFunction, &jobData.ArgType, arguments,
		&jobData.ScriptTargetFunction, &jobData.CreatedAt, &jobData.LastExecuted)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch job data: %v", err)
	}

	for i, arg := range arguments {
		jobData.Arguments[fmt.Sprintf("arg%d", i)] = arg
	}

	return &jobData, nil
}

// UpdateJobStatus updates the status field for a job in the database.
// Used to track job lifecycle states (pending, running, completed, failed etc).
func (s *JobScheduler) UpdateJobStatus(jobID int64, status string) error {
	jobIDStr := strconv.FormatInt(jobID, 10)

	err := s.dbClient.Session().Query(`
		UPDATE triggerx.job_data 
		SET status = ? 
		WHERE jobID = ?`, status, jobIDStr).Scan()

	if err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}

	return nil
}

// UpdateJobLastExecuted updates the last execution timestamp for a job.
// Critical for tracking execution history and scheduling recurring jobs.
func (s *JobScheduler) UpdateJobLastExecuted(jobID int64, lastExecuted time.Time) error {
	jobIDStr := strconv.FormatInt(jobID, 10)

	err := s.dbClient.Session().Query(`
		UPDATE triggerx.job_data 
		SET lastExecutedAt = ? 
		WHERE jobID = ?`, lastExecuted, jobIDStr).Scan()

	if err != nil {
		return fmt.Errorf("failed to update job last executed: %v", err)
	}

	return nil
}

// CreateTaskData sends a POST request to create a new task in the task management service.
// Returns success status and any errors encountered during task creation.
func CreateTaskData(taskData *types.CreateTaskData) (taskID int64, status bool, err error) {
	client := &http.Client{}

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return 0, false, fmt.Errorf("failed to marshal task data: %v", err)
	}

	req, err := http.NewRequest("POST", "http://data.triggerx.network:8080/api/tasks", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, false, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, false, fmt.Errorf("failed to create task, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var taskResponse types.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResponse); err != nil {
		return 0, false, fmt.Errorf("failed to decode response: %v", err)
	}

	return taskResponse.TaskID, true, nil
}
