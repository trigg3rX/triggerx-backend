package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateJobData(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[CreateJobData] Received request method: %s", r.Method)

	// Handle preflight
	if r.Method == http.MethodOptions {
		h.logger.Infof("[CreateJobData] Handling preflight request")
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Errorf("[CreateJobData] Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	created_at := time.Now().UTC()
	last_updated_at := time.Now().UTC()

	// Create a temporary struct to handle string chain_id
	type tempJobData struct {
		JobID               int64    `json:"job_id"`
		JobType             int64    `json:"jobType"`
		UserAddress         string   `json:"user_address"`
		ChainID             string   `json:"chain_id"`
		TimeFrame           int64    `json:"time_frame"`
		TimeInterval        int64    `json:"time_interval"`
		ContractAddress     string   `json:"contract_address"`
		TargetFunction      string   `json:"target_function"`
		TargetEvent         string   `json:"target_event"`
		ArgType             int64    `json:"arg_type"`
		Arguments           []string `json:"arguments"`
		Status              bool     `json:"status"`
		JobCostPrediction   int64    `json:"job_cost_prediction"`
		ScriptFunction      string   `json:"script_function"`
		ScriptIpfsUrl       string   `json:"script_ipfs_url"`
		StakeAmount         float64  `json:"stake_amount"`
		UserBalance         float64  `json:"user_balance"`
		DisputePeriodBlocks string   `json:"dispute_period_blocks"`
		Priority            int      `json:"priority"`
		Security            int      `json:"security"`
		TaskIDs             []int64  `json:"task_ids"`
	}

	var tempJob tempJobData
	if err := json.Unmarshal(body, &tempJob); err != nil {
		h.logger.Errorf("[CreateJobData] Error decoding JSON for job_id %d: %v", tempJob.JobID, err)
		http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Convert UserBalance to big.Float for precise decimal handling
	userBalanceDecimal := new(big.Float).SetFloat64(tempJob.UserBalance)

	// Convert to big.Int for storage as varint in the database
	userBalanceInt, _ := userBalanceDecimal.Int(nil)

	// Convert hex string to int64
	chainID, err := strconv.ParseInt(tempJob.ChainID[2:], 16, 64) // Remove "0x" prefix and parse as hex
	if err != nil {
		h.logger.Errorf("[CreateJobData] Error parsing chain_id for job_id %d: %v", tempJob.JobID, err)
		http.Error(w, "Invalid chain_id format", http.StatusBadRequest)
		return
	}

	// Convert DisputePeriodBlocks from string to *big.Int
	disputePeriodBlocksInt := new(big.Int)
	disputePeriodBlocksInt.SetString(tempJob.DisputePeriodBlocks, 10)

	// Create the actual JobData struct
	jobData := types.JobData{
		JobID:               tempJob.JobID,
		JobType:             int(tempJob.JobType),
		UserAddress:         tempJob.UserAddress,
		ChainID:             int(chainID),
		TimeFrame:           tempJob.TimeFrame,
		TimeInterval:        int(tempJob.TimeInterval),
		ContractAddress:     tempJob.ContractAddress,
		TargetFunction:      tempJob.TargetFunction,
		TargetEvent:         tempJob.TargetEvent,
		ArgType:             int(tempJob.ArgType),
		Arguments:           tempJob.Arguments,
		Status:              tempJob.Status,
		JobCostPrediction:   int(tempJob.JobCostPrediction),
		ScriptFunction:      tempJob.ScriptFunction,
		ScriptIpfsUrl:       tempJob.ScriptIpfsUrl,
		DisputePeriodBlocks: disputePeriodBlocksInt,
		Priority:            tempJob.Priority,
		Security:            tempJob.Security,
		TaskIDs:             tempJob.TaskIDs,
	}

	h.logger.Infof("[CreateJobData] Processing job creation for job_id %d", jobData.JobID)

	// Check if user exists by user_address
	var existingUserID int64
	var existingJobIDs []int64
	var existingStakeAmount *big.Int
	var existingUserBalance *big.Int

	err = h.db.Session().Query(`
        SELECT user_id, job_ids, stake_amount, account_balance 
        FROM triggerx.user_data 
        WHERE user_address = ? ALLOW FILTERING`,
		jobData.UserAddress).Scan(&existingUserID, &existingJobIDs, &existingStakeAmount, &existingUserBalance)

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error checking user existence for address %s: %v", jobData.UserAddress, err)
		http.Error(w, "Error checking user existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get new user ID if user doesn't exist
	if err == gocql.ErrNotFound {
		h.logger.Infof("[CreateJobData] Creating new user for address %s", jobData.UserAddress)
		var maxUserID int64
		if err := h.db.Session().Query(`
            SELECT MAX(user_id) FROM triggerx.user_data
        `).Scan(&maxUserID); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting max user ID: %v", err)
			http.Error(w, "Error getting max user ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		existingUserID = maxUserID + 1

		// Convert stake amount to Gwei and store as varint
		stakeAmountGwei := new(big.Float).SetFloat64(tempJob.StakeAmount)

		stakeAmountInt, _ := stakeAmountGwei.Int(nil)

		if err := h.db.Session().Query(`
            INSERT INTO triggerx.user_data (
                user_id, user_address, job_ids, stake_amount, created_at, last_updated_at, account_balance
            ) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			existingUserID, jobData.UserAddress, []int64{jobData.JobID}, stakeAmountInt,
			time.Now().UTC(), time.Now().UTC(), userBalanceInt).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error creating user data for user_id %d: %v", existingUserID, err)
			http.Error(w, "Error creating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Created new user with user_id %d", existingUserID)
	} else {
		// Update existing user's job IDs and add to existing stake amount
		updatedJobIDs := append(existingJobIDs, jobData.JobID)

		// Convert new stake amount to big.Int and add to existing
		updateaccountBalance := new(big.Int).Add(existingUserBalance, userBalanceInt)
		newStakeFloat := new(big.Float).SetFloat64(tempJob.StakeAmount)
		newStakeInt, _ := newStakeFloat.Int(nil)
		newStakeAmount := new(big.Int).Add(existingStakeAmount, newStakeInt)

		if err := h.db.Session().Query(`
            UPDATE triggerx.user_data 
            SET job_ids = ?, stake_amount = ?, account_balance = ?
            WHERE user_id = ?`,
			updatedJobIDs, newStakeAmount, updateaccountBalance, existingUserID).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user data for user_id %d: %v", existingUserID, err)
			http.Error(w, "Error updating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Updated user data for user_id %d", existingUserID)
	}

	// Create the job
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.job_data (
            job_id, jobType, user_id, chain_id, 
            time_frame, time_interval, contract_address, target_function, 
            target_event, arg_type, arguments, status, job_cost_prediction,
            script_function, script_ipfs_url, user_address,
            created_at, last_executed_at, dispute_period_blocks, priority, security, task_ids
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		jobData.JobID, jobData.JobType, existingUserID, jobData.ChainID,
		jobData.TimeFrame, jobData.TimeInterval, jobData.ContractAddress, jobData.TargetFunction,
		jobData.TargetEvent, jobData.ArgType, jobData.Arguments, jobData.Status, jobData.JobCostPrediction,
		jobData.ScriptFunction, jobData.ScriptIpfsUrl, jobData.UserAddress,
		created_at, last_updated_at, jobData.DisputePeriodBlocks, jobData.Priority, jobData.Security, jobData.TaskIDs).Exec(); err != nil {
		h.logger.Errorf("[CreateJobData] Error inserting job data for job_id %d: %v", jobData.JobID, err)
		http.Error(w, "Error inserting job data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateJobData] Successfully created job_id %d", jobData.JobID)

	// After successfully creating the job
	eventBus := events.GetEventBus()
	if eventBus == nil {
		h.logger.Infof("[CreateJobData] Warning: EventBus is nil, event will not be published")
		return
	}

	h.logger.Infof("[CreateJobData] Publishing job creation event for job_id %d", jobData.JobID)
	event := events.JobEvent{
		Type:    "job_created",
		JobID:   jobData.JobID,
		JobType: jobData.JobType,
		ChainID: jobData.ChainID,
	}

	if err := eventBus.PublishJobEvent(r.Context(), event); err != nil {
		h.logger.Infof("[CreateJobData] Warning: Failed to publish job creation event: %v", err)
	} else {
		h.logger.Infof("[CreateJobData] Successfully published job creation event for job_id %d", jobData.JobID)
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Convert BigInt to string for JSON response
	response := map[string]interface{}{
		"message": "Job created successfully",
		"job": map[string]interface{}{
			"job_id":                jobData.JobID,
			"jobType":               jobData.JobType,
			"user_id":               existingUserID,
			"chain_id":              fmt.Sprintf("%d", jobData.ChainID), // Convert to string if needed
			"time_frame":            jobData.TimeFrame,
			"time_interval":         jobData.TimeInterval,
			"contract_address":      jobData.ContractAddress,
			"target_function":       jobData.TargetFunction,
			"target_event":          jobData.TargetEvent,
			"arg_type":              jobData.ArgType,
			"arguments":             jobData.Arguments,
			"status":                jobData.Status,
			"job_cost_prediction":   jobData.JobCostPrediction,
			"script_function":       jobData.ScriptFunction,
			"script_ipfs_url":       jobData.ScriptIpfsUrl,
			"dispute_period_blocks": jobData.DisputePeriodBlocks,
			"priority":              jobData.Priority,
			"security":              jobData.Security,
			"task_ids":              jobData.TaskIDs,
			"time_check":            jobData.TimeCheck,
			"user_address":          jobData.UserAddress,
			"created_at":            time.Now().UTC(),
			"last_executed_at":      time.Now().UTC(),
		},
	}

	// Return response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("[CreateJobData] Error encoding response for job_id %d: %v", jobData.JobID, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[UpdateJobData] Updating job_id %s", jobID)

	var jobData types.JobData
	if err := json.NewDecoder(r.Body).Decode(&jobData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request for job_id %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.job_data 
        SET jobType = ?, user_id = ?, chain_id = ?, 
            time_frame = ?, time_interval = ?, contract_address = ?,
            target_function = ?, target_event = ?, arg_type = ?, arguments = ?,
            status = ?, job_cost_prediction = ?, last_executed_at = ?
        WHERE job_id = ?`,
		jobData.JobType, jobData.UserID, jobData.ChainID,
		jobData.TimeFrame, jobData.TimeInterval, jobData.ContractAddress,
		jobData.TargetFunction, jobData.TargetEvent, jobData.ArgType, jobData.Arguments,
		jobData.Status, jobData.JobCostPrediction, jobData.LastExecutedAt, jobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating job_id %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// After successfully updating the job
	if eventBus := events.GetEventBus(); eventBus != nil {
		event := events.JobEvent{
			Type:    "job_updated",
			JobID:   jobData.JobID,
			JobType: jobData.JobType,
			ChainID: jobData.ChainID,
		}
		if err := eventBus.PublishJobEvent(r.Context(), event); err != nil {
			h.logger.Infof("[UpdateJobData] Warning: Failed to publish job update event: %v", err)
			// Continue execution - don't fail the request due to event publishing
		} else {
			h.logger.Infof("[UpdateJobData] Successfully published job update event for job_id %d", jobData.JobID)
		}
	}

	h.logger.Infof("[UpdateJobData] Successfully updated and published event for job_id %s", jobID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) GetJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[GetJobData] Fetching data for job_id %s", jobID)

	var jobData types.JobData
	if err := h.db.Session().Query(`
        SELECT job_id, jobType, user_id, chain_id, time_frame, 
               time_interval, contract_address, target_function, 
               target_event, arg_type, arguments, status, job_cost_prediction
        FROM triggerx.job_data 
        WHERE job_id = ?`, jobID).Scan(
		&jobData.JobID, &jobData.JobType, &jobData.UserID, &jobData.ChainID,
		&jobData.TimeFrame, &jobData.TimeInterval, &jobData.ContractAddress,
		&jobData.TargetFunction, &jobData.TargetEvent, &jobData.ArgType,
		&jobData.Arguments, &jobData.Status, &jobData.JobCostPrediction); err != nil {
		h.logger.Errorf("[GetJobData] Error retrieving job_id %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetJobData] Successfully retrieved job_id %s", jobID)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[GetAllJobs] Fetching all jobs")
	var jobs []types.JobData
	iter := h.db.Session().Query(`SELECT * FROM triggerx.job_data`).Iter()

	var job types.JobData
	for iter.Scan(
		&job.JobID, &job.JobType, &job.UserID, &job.ChainID,
		&job.TimeFrame, &job.TimeInterval, &job.ContractAddress,
		&job.TargetFunction, &job.TargetEvent, &job.ArgType, &job.Arguments,
		&job.Status, &job.JobCostPrediction) {
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

	// Query to get the maximum job_id
	if err := h.db.Session().Query(`
        SELECT MAX(job_id) FROM triggerx.job_data
    `).Scan(&latestJobID); err != nil {
		if err == gocql.ErrNotFound {
			h.logger.Infof("[GetLatestJobID] No jobs found, starting with job_id 0")
			latestJobID = 0
			json.NewEncoder(w).Encode(map[string]int64{"latest_job_id": latestJobID})
			return
		}
		h.logger.Errorf("[GetLatestJobID] Error fetching latest job ID: %v", err)
		http.Error(w, "Error fetching latest job ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetLatestJobID] Latest job_id is %d", latestJobID)
	json.NewEncoder(w).Encode(map[string]int64{"latest_job_id": latestJobID})
}

func (h *Handler) GetJobsByUserAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userAddress := vars["user_address"]
	h.logger.Infof("[GetJobsByUserAddress] Fetching jobs for user_address %s", userAddress)

	type JobSummary struct {
		JobID   int64 `json:"job_id"`
		JobType int   `json:"jobType"`
		Status  bool  `json:"status"`
	}

	var userJobs []JobSummary

	iter := h.db.Session().Query(`
        SELECT job_id, jobType, status 
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
