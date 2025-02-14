package manager

import (
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (s *JobScheduler) GetJobDetails(jobID int64) (*types.Job, error) {
	var jobData types.JobData

	jobIDStr := strconv.FormatInt(jobID, 10)

	// Query the database using the job ID
	err := s.dbClient.Session().Query(`
		SELECT job_id, jobType, user_address, chain_id, 
			   time_frame, time_interval, contract_address, 
			   target_function, arg_type, arguments, status, 
			   job_cost_prediction, script_function, script_ipfs_url
		FROM triggerx.job_data 
		WHERE job_id = ?`, jobIDStr).Scan(
		&jobData.JobID, &jobData.JobType, &jobData.UserAddress,
		&jobData.ChainID, &jobData.TimeFrame, &jobData.TimeInterval,
		&jobData.ContractAddress, &jobData.TargetFunction,
		&jobData.ArgType, &jobData.Arguments, &jobData.Status,
		&jobData.JobCostPrediction, &jobData.ScriptFunction,
		&jobData.ScriptIpfsUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch job data: %v", err)
	}

	// Convert database model to Job struct
	job := &types.Job{
		JobID:             strconv.FormatInt(jobData.JobID, 10),
		JobType:           jobData.JobType,
		ChainID:           strconv.Itoa(jobData.ChainID),
		ContractAddress:   jobData.ContractAddress,
		TimeFrame:         jobData.TimeFrame,
		TimeInterval:      int64(jobData.TimeInterval),
		TargetFunction:    jobData.TargetFunction,
		ArgType:           strconv.Itoa(jobData.ArgType),
		Arguments:         make(map[string]interface{}),
		ScriptFunction:    jobData.ScriptFunction,
		ScriptIpfsUrl:     jobData.ScriptIpfsUrl,
		CreatedAt:         jobData.CreatedAt,
		LastExecuted:      jobData.LastExecutedAt,
	}

	// Convert arguments array to map
	for i, arg := range jobData.Arguments {
		job.Arguments[fmt.Sprintf("arg%d", i)] = arg
	}

	return job, nil
}