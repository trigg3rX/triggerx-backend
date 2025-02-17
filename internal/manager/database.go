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
	var jobData types.JobData

	jobIDStr := strconv.FormatInt(jobID, 10)

	err := s.dbClient.Session().Query(`
		SELECT jobID, jobType, userID, chainID, 
			   timeFrame, timeInterval, triggerContractAddress, 
			   triggerEvent, targetContractAddress, targetFunction, 
			   argType, arguments, recurring, status, 
			   jobCostPrediction, createdAt, lastExecutedAt, scriptFunction,
			   scriptIPFSUrl, priority, security, taskIDs, linkJobID
		FROM triggerx.job_data 
		WHERE jobID = ?`, jobIDStr).Scan(
		&jobData.JobID, &jobData.JobType, &jobData.UserID,
		&jobData.ChainID, &jobData.TimeFrame, &jobData.TimeInterval,
		&jobData.TriggerContractAddress, &jobData.TriggerEvent,
		&jobData.TargetContractAddress, &jobData.TargetFunction,
		&jobData.ArgType, &jobData.Arguments, &jobData.Recurring,
		&jobData.Status, &jobData.JobCostPrediction,
		&jobData.CreatedAt, &jobData.LastExecutedAt,
		&jobData.ScriptFunction, &jobData.ScriptIPFSUrl,
		&jobData.Priority, &jobData.Security, &jobData.TaskIDs,
		&jobData.LinkJobID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch job data: %v", err)
	}

	job := &types.Job{
		JobID:                  jobData.JobID,
		JobType:                jobData.JobType,
		UserID:                 jobData.UserID,
		ChainID:                jobData.ChainID,
		TimeFrame:              jobData.TimeFrame,
		TimeInterval:           int64(jobData.TimeInterval),
		TriggerContractAddress: jobData.TriggerContractAddress,
		TriggerEvent:           jobData.TriggerEvent,
		TargetContractAddress:  jobData.TargetContractAddress,
		TargetFunction:         jobData.TargetFunction,
		ArgType:                jobData.ArgType,
		Recurring:              jobData.Recurring,
		ScriptFunction:         jobData.ScriptFunction,
		ScriptIPFSUrl:          jobData.ScriptIPFSUrl,
		Status:                 jobData.Status,
		CreatedAt:              jobData.CreatedAt,
		LastExecuted:           jobData.LastExecutedAt,
		Priority:               jobData.Priority,
		Security:               jobData.Security,
		TaskIDs:                jobData.TaskIDs,
		LinkID:                 jobData.LinkJobID,
		Arguments:              make(map[string]interface{}),
	}

	for i, arg := range jobData.Arguments {
		job.Arguments[fmt.Sprintf("arg%d", i)] = arg
	}

	return job, nil
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
func CreateTaskData(taskData *types.TaskData) (status bool, err error) {
	client := &http.Client{}

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return false, fmt.Errorf("failed to marshal task data: %v", err)
	}

	req, err := http.NewRequest("POST", "http://data.triggerx.network:8080/api/tasks", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to create task, status: %d, body: %s", resp.StatusCode, string(body))
	}

	return true, nil

}

// GetPerformer retrieves an available performer node from the keeper network.
// Returns the first available performer or error if none are available.
func GetPerformer() (types.Performer, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://data.triggerx.network:8080/api/keepers/performers", nil)
	if err != nil {
		return types.Performer{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return types.Performer{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return types.Performer{}, fmt.Errorf("failed to get performers, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var performers []types.Performer
	if err := json.NewDecoder(resp.Body).Decode(&performers); err != nil {
		return types.Performer{}, fmt.Errorf("failed to decode performers: %v", err)
	}

	if len(performers) == 0 {
		return types.Performer{}, fmt.Errorf("no performers available")
	}

	return performers[0], nil
}
