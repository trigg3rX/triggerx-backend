package api

import (
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateJobData(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[CreateJobData] Received request method: %s", r.Method)

	if r.Method == http.MethodOptions {
		h.logger.Infof("[CreateJobData] Handling preflight request")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Errorf("[CreateJobData] Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	createdAt := time.Now().UTC()
	lastUpdatedAt := time.Now().UTC()

	var tempJob types.CreateJobData
	if err := json.Unmarshal(body, &tempJob); err != nil {
		h.logger.Errorf("[CreateJobData] Error decoding JSON for jobID %d: %v", tempJob.JobID, err)
		http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	jobData := types.JobData{
		JobID:               	tempJob.JobID,
		JobType:             	int(tempJob.JobType),
		ChainID:             	int(tempJob.ChainID),
		TimeFrame:           	tempJob.TimeFrame,
		TimeInterval:        	int(tempJob.TimeInterval),
		TriggerContractAddress: tempJob.TriggerContractAddress,
		TriggerEvent:         	tempJob.TriggerEvent,
		TargetContractAddress: tempJob.TargetContractAddress,
		TargetFunction:      	tempJob.TargetFunction,
		ArgType:             	int(tempJob.ArgType),
		Arguments:           	tempJob.Arguments,
		Recurring:           	tempJob.Recurring,
		ScriptFunction:      	tempJob.ScriptFunction,
		ScriptIPFSUrl:       	tempJob.ScriptIPFSUrl,
		Status:              	true,
		JobCostPrediction:   	tempJob.JobCostPrediction,
		Priority:            	tempJob.Priority,
		Security:            	tempJob.Security,
		LinkJobID:              tempJob.LinkJobID,
	}

	h.logger.Infof("[CreateJobData] Processing job creation for jobID %d", jobData.JobID)

	var existingUserID int64
	var existingJobIDs []int64
	var existingStakeAmount *big.Int
	var existingUserBalance *big.Int
	var updatedJobIDs []int64
	var newStakeAmount *big.Int
	var updateaccountBalance *big.Int

	err = h.db.Session().Query(`
        SELECT userID, jobIDs, stakeAmount, accountBalance 
        FROM triggerx.user_data 
        WHERE userAddress = ? ALLOW FILTERING`,
		tempJob.UserAddress).Scan(&existingUserID, &existingJobIDs, &existingStakeAmount, &existingUserBalance)

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error checking user existence for address %s: %v", tempJob.UserAddress, err)
		http.Error(w, "Error checking user existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err == gocql.ErrNotFound {
		h.logger.Infof("[CreateJobData] Creating new user for address %s", tempJob.UserAddress)
		var maxUserID int64
		if err := h.db.Session().Query(`
            SELECT MAX(userID) FROM triggerx.user_data
        `).Scan(&maxUserID); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting max user ID: %v", err)
			http.Error(w, "Error getting max userID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		existingUserID = maxUserID + 1

		stakeAmountGwei := new(big.Float).SetInt(tempJob.StakeAmount)
		stakeAmountInt, _ := stakeAmountGwei.Int(nil)

		if err := h.db.Session().Query(`
            INSERT INTO triggerx.user_data (
                userID, userAddress, jobIDs, stakeAmount, createdAt, lastUpdatedAt, accountBalance
            ) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			existingUserID, tempJob.UserAddress, []int64{jobData.JobID}, stakeAmountInt,
			createdAt, lastUpdatedAt, tempJob.StakeAmount).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error creating user data for userID %d: %v", existingUserID, err)
			http.Error(w, "Error creating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Created new user with userID %d", existingUserID)
	} else {
		updatedJobIDs := append(existingJobIDs, jobData.JobID)

		updateaccountBalance := new(big.Int).Add(existingUserBalance, tempJob.StakeAmount)
		newStakeFloat := new(big.Float).SetInt(tempJob.StakeAmount)
		newStakeInt, _ := newStakeFloat.Int(nil)
		newStakeAmount := new(big.Int).Add(existingStakeAmount, newStakeInt)

		if err := h.db.Session().Query(`
            UPDATE triggerx.user_data 
            SET jobIDs = ?, stakeAmount = ?, accountBalance = ?
            WHERE userID = ?`,
			updatedJobIDs, newStakeAmount, updateaccountBalance, existingUserID).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user data for userID %d: %v", existingUserID, err)
			http.Error(w, "Error updating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Updated user data for userID %d", existingUserID)
	}

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.job_data (
            jobID, jobType, userID, chainID, timeFrame, 
			timeInterval, triggerContractAddress, triggerEvent, 
            targetContractAddress, targetFunction, argType, arguments, recurring, 
            scriptFunction, scriptIPFSUrl, status, jobCostPrediction, createdAt, 
            lastExecutedAt, priority, security, taskIDs, linkJobID
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		jobData.JobID, jobData.JobType, existingUserID, jobData.ChainID, jobData.TimeFrame, 
		jobData.TimeInterval, jobData.TriggerContractAddress, jobData.TriggerEvent,
		jobData.TargetContractAddress, jobData.TargetFunction, jobData.ArgType, jobData.Arguments, jobData.Recurring,
		jobData.ScriptFunction, jobData.ScriptIPFSUrl, jobData.Status, jobData.JobCostPrediction,
		jobData.CreatedAt, jobData.LastExecutedAt, jobData.Priority, jobData.Security, jobData.TaskIDs, jobData.LinkJobID).Exec(); err != nil {
		h.logger.Errorf("[CreateJobData] Error inserting job data for jobID %d: %v", jobData.JobID, err)
		http.Error(w, "Error inserting job data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateJobData] Successfully created jobID %d", jobData.JobID)

	eventBus := events.GetEventBus()
	if eventBus == nil {
		h.logger.Infof("[CreateJobData] Warning: EventBus is nil, event will not be published")
		return
	}

	h.logger.Infof("[CreateJobData] Publishing job creation event for jobID %d", jobData.JobID)
	event := events.JobEvent{
		Type:    "job_created",
		JobID:   jobData.JobID,
		JobType: jobData.JobType,
		ChainID: jobData.ChainID,
	}

	if err := eventBus.PublishJobEvent(r.Context(), event); err != nil {
		h.logger.Infof("[CreateJobData] Warning: Failed to publish job creation event: %v", err)
	} else {
		h.logger.Infof("[CreateJobData] Successfully published job creation event for jobID %d", jobData.JobID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"message": "Database Updated Successfully",
		"User": map[string]interface{}{
			"userID":                existingUserID,
			"userAddress":           tempJob.UserAddress,
			"jobIDs":                updatedJobIDs,
			"stakeAmount":           newStakeAmount,
			"accountBalance":        updateaccountBalance,
			"createdAt":             createdAt,
			"lastUpdatedAt":         lastUpdatedAt,
		},
		"Job": map[string]interface{}{
			"jobID":                 jobData.JobID,
			"jobType":               jobData.JobType,
			"userID":                existingUserID,
			"chainID":               jobData.ChainID,
			"timeFrame":             jobData.TimeFrame,
			"timeInterval":          jobData.TimeInterval,
			"triggerContractAddress":jobData.TriggerContractAddress,
			"triggerEvent":          jobData.TriggerEvent,
			"targetContractAddress": jobData.TargetContractAddress,
			"targetFunction":        jobData.TargetFunction,
			"argType":               jobData.ArgType,
			"arguments":             jobData.Arguments,
			"recurring":             jobData.Recurring,
			"status":                jobData.Status,
			"jobCostPrediction":     jobData.JobCostPrediction,
			"scriptFunction":        jobData.ScriptFunction,
			"scriptIPFSUrl":         jobData.ScriptIPFSUrl,
			"priority":              jobData.Priority,
			"security":              jobData.Security,
			"taskIDs":               jobData.TaskIDs,
			"linkJobID":             jobData.LinkJobID,
			"createdAt":             createdAt,
			"lastExecutedAt":        lastUpdatedAt,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("[CreateJobData] Error encoding response for jobID %d: %v", jobData.JobID, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[UpdateJobData] Updating jobID %s", jobID)

	var jobData types.JobData
	if err := json.NewDecoder(r.Body).Decode(&jobData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request for jobID %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.job_data 
        SET jobType = ?, userID = ?, chainID = ?, 
            timeFrame = ?, timeInterval = ?, triggerContractAddress = ?,
            triggerEvent = ?, targetContractAddress = ?, targetFunction = ?, 
            argType = ?, arguments = ?, recurring = ?, scriptFunction = ?,
            scriptIPFSUrl = ?, status = ?, jobCostPrediction = ?, 
            priority = ?, security = ?, taskIDs = ?, linkJobID = ?,
            lastExecutedAt = ?
        WHERE jobID = ?`,
		jobData.JobType, jobData.UserID, jobData.ChainID,
		jobData.TimeFrame, jobData.TimeInterval, jobData.TriggerContractAddress,
		jobData.TriggerEvent, jobData.TargetContractAddress, jobData.TargetFunction,
		jobData.ArgType, jobData.Arguments, jobData.Recurring, jobData.ScriptFunction,
		jobData.ScriptIPFSUrl, jobData.Status, jobData.JobCostPrediction,
		jobData.Priority, jobData.Security, jobData.TaskIDs, jobData.LinkJobID,
		jobData.LastExecutedAt, jobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating jobID %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if eventBus := events.GetEventBus(); eventBus != nil {
		event := events.JobEvent{
			Type:    "job_updated",
			JobID:   jobData.JobID,
			JobType: jobData.JobType,
			ChainID: jobData.ChainID,
		}
		if err := eventBus.PublishJobEvent(r.Context(), event); err != nil {
			h.logger.Infof("[UpdateJobData] Warning: Failed to publish job update event: %v", err)
		} else {
			h.logger.Infof("[UpdateJobData] Successfully published job update event for jobID %d", jobData.JobID)
		}
	}

	h.logger.Infof("[UpdateJobData] Successfully updated and published event for jobID %s", jobID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) GetJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[GetJobData] Fetching data for jobID %s", jobID)

	var jobData types.JobData
	if err := h.db.Session().Query(`
        SELECT jobID, jobType, userID, chainID, timeFrame, 
               timeInterval, triggerContractAddress, triggerEvent,
               targetContractAddress, targetFunction, argType, arguments,
               recurring, scriptFunction, scriptIPFSUrl, status,
               jobCostPrediction, priority, security, taskIDs,
               linkJobID, createdAt, lastExecutedAt
        FROM triggerx.job_data 
        WHERE jobID = ?`, jobID).Scan(
		&jobData.JobID, &jobData.JobType, &jobData.UserID, &jobData.ChainID,
		&jobData.TimeFrame, &jobData.TimeInterval, &jobData.TriggerContractAddress,
		&jobData.TriggerEvent, &jobData.TargetContractAddress, &jobData.TargetFunction,
		&jobData.ArgType, &jobData.Arguments, &jobData.Recurring, &jobData.ScriptFunction,
		&jobData.ScriptIPFSUrl, &jobData.Status, &jobData.JobCostPrediction,
		&jobData.Priority, &jobData.Security, &jobData.TaskIDs,
		&jobData.LinkJobID, &jobData.CreatedAt, &jobData.LastExecutedAt); err != nil {
		h.logger.Errorf("[GetJobData] Error retrieving jobID %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetJobData] Successfully retrieved jobID %s", jobID)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[GetAllJobs] Fetching all jobs")
	var jobs []types.JobData
	iter := h.db.Session().Query(`
        SELECT jobID, jobType, userID, chainID, timeFrame,
               timeInterval, triggerContractAddress, triggerEvent,
               targetContractAddress, targetFunction, argType, arguments,
               recurring, scriptFunction, scriptIPFSUrl, status,
               jobCostPrediction, priority, security, taskIDs,
               linkJobID, createdAt, lastExecutedAt
        FROM triggerx.job_data`).Iter()

	var job types.JobData
	for iter.Scan(
		&job.JobID, &job.JobType, &job.UserID, &job.ChainID,
		&job.TimeFrame, &job.TimeInterval, &job.TriggerContractAddress,
		&job.TriggerEvent, &job.TargetContractAddress, &job.TargetFunction,
		&job.ArgType, &job.Arguments, &job.Recurring, &job.ScriptFunction,
		&job.ScriptIPFSUrl, &job.Status, &job.JobCostPrediction,
		&job.Priority, &job.Security, &job.TaskIDs,
		&job.LinkJobID, &job.CreatedAt, &job.LastExecutedAt) {
		jobs = append(jobs, job)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetAllJobs] Error retrieving jobs: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetAllJobs] Successfully retrieved %d jobs", len(jobs))
	json.NewEncoder(w).Encode(jobs)
}

func (h *Handler) GetLatestJobID(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[GetLatestJobID] Fetching latest job ID")
	var latestJobID int64

	if err := h.db.Session().Query(`
        SELECT MAX(jobID) FROM triggerx.job_data
    `).Scan(&latestJobID); err != nil {
		if err == gocql.ErrNotFound {
			h.logger.Infof("[GetLatestJobID] No jobs found, starting with jobID 0")
			latestJobID = 0
			json.NewEncoder(w).Encode(map[string]int64{"latest_jobID": latestJobID})
			return
		}
		h.logger.Errorf("[GetLatestJobID] Error fetching latest job ID: %v", err)
		http.Error(w, "Error fetching latest job ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetLatestJobID] Latest jobID is %d", latestJobID) 
	json.NewEncoder(w).Encode(map[string]int64{"latest_jobID": latestJobID})
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
