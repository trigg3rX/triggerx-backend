package manager

import (
	"fmt"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

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
