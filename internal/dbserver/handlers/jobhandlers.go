package handlers

import (
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateJobData(w http.ResponseWriter, r *http.Request) {
	var tempData types.CreateJobData
	if err := json.NewDecoder(r.Body).Decode(&tempData); err != nil {
		h.logger.Errorf("[CreateJobData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var jobID int64
	if err := h.db.Session().Query(`
		SELECT MAX(job_id) FROM triggerx.job_data`).Scan(&jobID); err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error getting max job ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jobID = jobID + 1

	var existingUserID int64
	var existingJobIDs []int64
	var existingAccountBalance *big.Int
	var existingTokenBalance *big.Int

	err := h.db.Session().Query(`
        SELECT user_id, job_ids, account_balance, token_balance 
        FROM triggerx.user_data 
        WHERE user_address = ? ALLOW FILTERING`,
		tempData.UserAddress).Scan(&existingUserID, &existingJobIDs, 
			&existingAccountBalance, &existingTokenBalance)

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error checking user existence for address %s: %v", tempData.UserAddress, err)
		http.Error(w, "Error checking user existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err == gocql.ErrNotFound {
		h.logger.Infof("[CreateJobData] Creating new user for address %s", tempData.UserAddress)
		var maxUserID int64
		if err := h.db.Session().Query(`
            SELECT MAX(user_id) FROM triggerx.user_data
        `).Scan(&maxUserID); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting max user ID: %v", err)
			http.Error(w, "Error getting max userID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		existingUserID = maxUserID + 1

		if err := h.db.Session().Query(`
            INSERT INTO triggerx.user_data (
                user_id, user_address, created_at, 
				job_ids, account_balance, token_balance,  last_updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			existingUserID, tempData.UserAddress, time.Now(),
			[]int64{jobID}, tempData.StakeAmount, tempData.TokenAmount, time.Now()).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error creating user data for userID %d: %v", existingUserID, err)
			http.Error(w, "Error creating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Created new user with userID %d", existingUserID)
	} else {
		updatedJobIDs := append(existingJobIDs, jobID)

		existingAccountBalance = new(big.Int).Add(existingAccountBalance, tempData.StakeAmount)
		existingTokenBalance = new(big.Int).Add(existingTokenBalance, tempData.TokenAmount)

		if err := h.db.Session().Query(`
            UPDATE triggerx.user_data 
            SET job_ids = ?, account_balance = ?, token_balance = ?, last_updated_at = ?
            WHERE user_id = ?`,
			updatedJobIDs, existingAccountBalance, existingTokenBalance, 
			time.Now(), existingUserID).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user data for userID %d: %v", existingUserID, err)
			http.Error(w, "Error updating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Updated user data for userID %d", existingUserID)
	}

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.job_data (
            job_id, task_definition_id, user_id, priority, security, link_job_id, 
			time_frame, recurring, time_interval, trigger_chain_id, trigger_contract_address, 
            trigger_event, script_ipfs_url, script_trigger_function, target_chain_id, 
            target_contract_address, target_function, arg_type, arguments, script_target_function, 
            status, job_cost_prediction, created_at, last_executed_at, task_ids
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		jobID, tempData.TaskDefinitionID, existingUserID, tempData.Priority, tempData.Security, tempData.LinkJobID, 
		tempData.TimeFrame, tempData.Recurring, tempData.TimeInterval, tempData.TriggerChainID, tempData.TriggerContractAddress, 
		tempData.TriggerEvent, tempData.ScriptIPFSUrl, tempData.ScriptTriggerFunction, tempData.TargetChainID, 
		tempData.TargetContractAddress, tempData.TargetFunction, tempData.ArgType, tempData.Arguments, tempData.ScriptTargetFunction, 
		false, tempData.JobCostPrediction, time.Now(), nil, []int64{}).Exec(); err != nil {
		h.logger.Errorf("[CreateJobData] Error inserting job data for jobID %d: %v", jobID, err)
		http.Error(w, "Error inserting job data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateJobData] Successfully created jobID %d", jobID)

	eventBus := events.GetEventBus()
	if eventBus == nil {
		h.logger.Infof("[CreateJobData] Warning: EventBus is nil, event will not be published")
		return
	}

	h.logger.Infof("[CreateJobData] Publishing job creation event for jobID %d", jobID)
	event := events.JobEvent{
		Type:    "job_created",
		JobID:   jobID,
		TaskDefinitionID: tempData.TaskDefinitionID,
		TimeFrame: tempData.TimeFrame,
	}

	if err := eventBus.PublishJobEvent(r.Context(), event); err != nil {
		h.logger.Infof("[CreateJobData] Warning: Failed to publish job creation event: %v", err)
	} else {
		h.logger.Infof("[CreateJobData] Successfully published job creation event for jobID %d", jobID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"message": "Database Updated Successfully",
		"User": types.CreateJobResponse{
			UserID: existingUserID,
			AccountBalance: existingAccountBalance,
			TokenBalance: existingTokenBalance,
			JobID: jobID,
			TaskDefinitionID: tempData.TaskDefinitionID,
			TimeFrame: tempData.TimeFrame,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("[CreateJobData] Error encoding response for jobID %d: %v", jobID, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateJobData(w http.ResponseWriter, r *http.Request) {
	var tempData types.UpdateJobData
	if err := json.NewDecoder(r.Body).Decode(&tempData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request for jobID %s: %v", tempData.JobID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.job_data 	
        SET time_frame = ?, recurring = ?
        WHERE job_id = ?`,
		tempData.TimeFrame, tempData.Recurring, tempData.JobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating jobID %s: %v", tempData.JobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if eventBus := events.GetEventBus(); eventBus != nil {
		event := events.JobEvent{
			Type:    "job_updated",
			JobID:   tempData.JobID,
			TimeFrame: tempData.TimeFrame,
		}
		if err := eventBus.PublishJobEvent(r.Context(), event); err != nil {
			h.logger.Infof("[UpdateJobData] Warning: Failed to publish job update event: %v", err)
		} else {
			h.logger.Infof("[UpdateJobData] Successfully published job update event for jobID %d", tempData.JobID)
		}
	}

	h.logger.Infof("[UpdateJobData] Successfully updated and published event for jobID %s", tempData.JobID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tempData)
}

func (h *Handler) GetJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[GetJobData] Fetching data for jobID %s", jobID)

	var jobData types.JobData
	if err := h.db.Session().Query(`
        SELECT job_id, task_definition_id, user_id, priority, security, link_job_id, 
               time_frame, recurring, time_interval, trigger_chain_id, trigger_contract_address, 
               trigger_event, script_ipfs_url, script_trigger_function, target_chain_id, 
               target_contract_address, target_function, arg_type, arguments, script_target_function, 
               status, job_cost_prediction, created_at, last_executed_at, task_ids
        FROM triggerx.job_data 
        WHERE job_id = ?`, jobID).Scan(
		&jobData.JobID, &jobData.TaskDefinitionID, &jobData.UserID, &jobData.Priority, &jobData.Security, &jobData.LinkJobID,
		&jobData.TimeFrame, &jobData.Recurring, &jobData.TimeInterval, &jobData.TriggerChainID, &jobData.TriggerContractAddress,
		&jobData.TriggerEvent, &jobData.ScriptIPFSUrl, &jobData.ScriptTriggerFunction, &jobData.TargetChainID,
		&jobData.TargetContractAddress, &jobData.TargetFunction, &jobData.ArgType, &jobData.Arguments, &jobData.ScriptTargetFunction,
		&jobData.Status, &jobData.JobCostPrediction, &jobData.CreatedAt, &jobData.LastExecutedAt, &jobData.TaskIDs); err != nil {
		h.logger.Errorf("[GetJobData] Error retrieving jobID %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetJobData] Successfully retrieved jobID %s", jobID)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) GetJobsByUserAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userAddress := vars["user_address"]
	h.logger.Infof("[GetJobsByUserAddress] Fetching jobs for user_address %s", userAddress)

	type JobSummary struct {
		JobID   int64 `json:"jobID"`
		JobType int   `json:"jobType"`
		Status  bool  `json:"status"`
	}

	var userJobs []JobSummary

	iter := h.db.Session().Query(`
        SELECT jobID, jobType, status 
        FROM triggerx.job_data 
        WHERE user_address = ? ALLOW FILTERING
    `, userAddress).Iter()

	var job JobSummary
	for iter.Scan(&job.JobID, &job.JobType, &job.Status) {
		userJobs = append(userJobs, job)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetJobsByUserAddress] Error retrieving jobs for user_address %s: %v", userAddress, err)
		http.Error(w, "Error retrieving jobs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetJobsByUserAddress] Retrieved %d jobs for user_address %s", len(userJobs), userAddress)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(userJobs); err != nil {
		h.logger.Errorf("[GetJobsByUserAddress] Error encoding response for user_address %s: %v", userAddress, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
