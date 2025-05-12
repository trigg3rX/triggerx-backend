package handlers

import (
	"encoding/json"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Create a new Job, and send it to the Manager
// If User doesn't exist, create a new user, or update the existing user
func (h *Handler) CreateJobData(w http.ResponseWriter, r *http.Request) {
	var tempJobs []types.CreateJobData
	if err := json.NewDecoder(r.Body).Decode(&tempJobs); err != nil {
		h.logger.Errorf("[CreateJobData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(tempJobs) == 0 {
		h.logger.Error("[CreateJobData] No jobs provided in request")
		http.Error(w, "No jobs provided", http.StatusBadRequest)
		return
	}

	// Set default value for custom field if not provided
	for i := range tempJobs {
		// The custom field will be false by default since it's a bool
		// We don't need to explicitly set it as Go's zero value for bool is false
		h.logger.Infof("[CreateJobData] Job %d custom field value: %v", i, tempJobs[i].Custom)
	}

	// Collect all IPFS URLs and check if any job needs dynamic fee calculation
	var ipfsURLs []string
	needsDynamicFee := false
	for i, job := range tempJobs {
		if job.ArgType == 2 { // Dynamic ArgType
			needsDynamicFee = true
			if job.ScriptIPFSUrl == "" {
				h.logger.Errorf("[CreateJobData] Missing IPFS URL for job %d", i)
				http.Error(w, "Missing IPFS URL", http.StatusBadRequest)
				return
			}
			ipfsURLs = append(ipfsURLs, job.ScriptIPFSUrl)
		}
	}

	var feePerJob float64 = 0.01 // Default fee per execution

	// Only calculate dynamic fees if needed
	if needsDynamicFee {
		// Calculate fees for jobs that need dynamic fee calculation
		var totalFee float64
		var err error
		totalFee, err = h.CalculateTaskFees(strings.Join(ipfsURLs, ","))
		if err != nil {
			h.logger.Errorf("[CreateJobData] Error calculating fees: %v", err)
			http.Error(w, "Error calculating fees", http.StatusInternalServerError)
			return
		}

		// The totalFee returned is the sum of all fees for dynamic jobs
		feePerJob = totalFee / float64(len(ipfsURLs))
	}

	// Calculate cost prediction for each job
	for i := range tempJobs {
		timeframeInSeconds := float64(tempJobs[i].TimeFrame)
		intervalInSeconds := float64(tempJobs[i].TimeInterval)
		executionCount := math.Ceil(timeframeInSeconds / intervalInSeconds)

		if tempJobs[i].ArgType == 2 {
			tempJobs[i].JobCostPrediction = executionCount * feePerJob
		} else {
			tempJobs[i].JobCostPrediction = executionCount * 0.01
		}
		h.logger.Infof("[CreateJobData] Calculated job cost for job %d: %f (ArgType: %d)", i, tempJobs[i].JobCostPrediction, tempJobs[i].ArgType)
	}

	var lastJobID int64
	if err := h.db.Session().Query(`
		SELECT MAX(job_id) FROM triggerx.job_data`).Scan(&lastJobID); err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error getting max job ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var existingUserID int64
	var existingAccountBalance *big.Int = big.NewInt(0)
	var existingTokenBalance *big.Int = big.NewInt(0)
	var existingJobIDs []int64 = []int64{}
	var newJobIDs []int64
	var err error
	err = h.db.Session().Query(`
		SELECT user_id, account_balance, token_balance, job_ids
		FROM triggerx.user_data 
		WHERE user_address = ? ALLOW FILTERING`,
		tempJobs[0].UserAddress).Scan(&existingUserID, &existingAccountBalance, &existingTokenBalance, &existingJobIDs)

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error checking user existence for address %s: %v", tempJobs[0].UserAddress, err)
		http.Error(w, "Error checking user existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err == gocql.ErrNotFound {
		h.logger.Infof("[CreateJobData] Creating new user for address %s", tempJobs[0].UserAddress)
		var maxUserID int64
		if err := h.db.Session().Query(`
			SELECT MAX(user_id) FROM triggerx.user_data`).Scan(&maxUserID); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting max user ID: %v", err)
			http.Error(w, "Error getting max userID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		existingUserID = maxUserID + 1
		existingAccountBalance = new(big.Int).Add(existingAccountBalance, tempJobs[0].StakeAmount)
		existingTokenBalance = new(big.Int).Add(existingTokenBalance, tempJobs[0].TokenAmount)

		if err := h.db.Session().Query(`
			INSERT INTO triggerx.user_data (
				user_id, user_address, created_at, 
				job_ids, account_balance, token_balance, last_updated_at, user_points
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			existingUserID, tempJobs[0].UserAddress, time.Now().UTC(),
			[]int64{}, existingAccountBalance, existingTokenBalance, time.Now().UTC(), 0.0).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error creating user data for userID %d: %v", existingUserID, err)
			http.Error(w, "Error creating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Created new user with userID %d", existingUserID)
	} else {
		existingAccountBalance = new(big.Int).Add(existingAccountBalance, tempJobs[0].StakeAmount)
		existingTokenBalance = new(big.Int).Add(existingTokenBalance, tempJobs[0].TokenAmount)

		// First get the current user points since it's part of the primary key
		var currentPoints float64
		if err := h.db.Session().Query(`
			SELECT user_points FROM triggerx.user_data 
			WHERE user_id = ?`,
			existingUserID).Scan(&currentPoints); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting current points for userID %d: %v", existingUserID, err)
			http.Error(w, "Error getting current points: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := h.db.Session().Query(`
			UPDATE triggerx.user_data 
			SET account_balance = ?, token_balance = ?, last_updated_at = ?
			WHERE user_id = ? AND user_points = ?`,
			existingAccountBalance, existingTokenBalance,
			time.Now().UTC(), existingUserID, currentPoints).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user data for userID %d: %v", existingUserID, err)
			http.Error(w, "Error updating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[CreateJobData] Updated user data for userID %d", existingUserID)
	}

	createdJobs := types.CreateJobResponse{
		UserID:            existingUserID,
		AccountBalance:    existingAccountBalance,
		TokenBalance:      existingTokenBalance,
		JobIDs:            make([]int64, len(tempJobs)),
		TaskDefinitionIDs: make([]int, len(tempJobs)),
		TimeFrames:        make([]int64, len(tempJobs)),
	}

	for i := len(tempJobs) - 1; i >= 0; i-- {
		lastJobID++
		currentJobID := lastJobID
		newJobIDs = append(newJobIDs, currentJobID)

		chainStatus := 1
		var linkJobID int64 = -1

		if i == 0 {
			chainStatus = 0
		}
		if i < len(tempJobs)-1 {
			linkJobID = createdJobs.JobIDs[i+1]
		}

		if err := h.db.Session().Query(`
			INSERT INTO triggerx.job_data (
				job_id, task_definition_id, user_id, priority, security, link_job_id, chain_status,
				time_frame, recurring, time_interval, trigger_chain_id, trigger_contract_address, 
				trigger_event, script_ipfs_url, script_trigger_function, target_chain_id, 
				target_contract_address, target_function, abi, arg_type, arguments, script_target_function, 
				status, job_cost_prediction, created_at, last_executed_at, task_ids, custom
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			currentJobID, tempJobs[i].TaskDefinitionID, existingUserID, tempJobs[i].Priority, tempJobs[i].Security, linkJobID, chainStatus,
			tempJobs[i].TimeFrame, tempJobs[i].Recurring, tempJobs[i].TimeInterval, tempJobs[i].TriggerChainID, tempJobs[i].TriggerContractAddress,
			tempJobs[i].TriggerEvent, tempJobs[i].ScriptIPFSUrl, tempJobs[i].ScriptTriggerFunction, tempJobs[i].TargetChainID,
			tempJobs[i].TargetContractAddress, tempJobs[i].TargetFunction, tempJobs[i].ABI, tempJobs[i].ArgType, tempJobs[i].Arguments, tempJobs[i].ScriptTargetFunction,
			false, tempJobs[i].JobCostPrediction, time.Now().UTC(), nil, []int64{}, tempJobs[i].Custom).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error inserting job data for jobID %d: %v", currentJobID, err)
			http.Error(w, "Error inserting job data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Add points based on whether the job is custom or not
		pointsToAdd := 10.0
		if tempJobs[i].Custom {
			pointsToAdd = 20.0
		}

		// Get current user points
		var currentPoints float64
		if err := h.db.Session().Query(`
			SELECT user_points FROM triggerx.user_data 
			WHERE user_id = ?`,
			existingUserID).Scan(&currentPoints); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting current points for userID %d: %v", existingUserID, err)
			http.Error(w, "Error getting current points: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update user points with new total using batch operation
		newPoints := currentPoints + pointsToAdd
		batch := h.db.Session().NewBatch(gocql.LoggedBatch)

		// Insert new row with updated points
		batch.Query(`
			INSERT INTO triggerx.user_data (
				user_id, user_points, user_address, created_at,
				job_ids, account_balance, token_balance, last_updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			existingUserID, newPoints, tempJobs[0].UserAddress, time.Now().UTC(),
			existingJobIDs, existingAccountBalance, existingTokenBalance, time.Now().UTC(),
		)

		// Delete old row with previous points
		batch.Query(`
			DELETE FROM triggerx.user_data
			WHERE user_id = ? AND user_points = ?`,
			existingUserID, currentPoints,
		)

		if err := h.db.Session().ExecuteBatch(batch); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user points for userID %d: %v", existingUserID, err)
			http.Error(w, "Error updating user points: "+err.Error(), http.StatusInternalServerError)
			return
		}

		h.logger.Infof("[CreateJobData] Successfully created jobID %d and added %.2f points to user", currentJobID, pointsToAdd)

		h.logger.Infof("[CreateJobData] Sending Job data to Manager for jobID %d", currentJobID)
		jobData := types.HandleCreateJobData{
			JobID:                  currentJobID,
			TaskDefinitionID:       tempJobs[i].TaskDefinitionID,
			UserID:                 existingUserID,
			Priority:               tempJobs[i].Priority,
			Security:               tempJobs[i].Security,
			LinkJobID:              linkJobID,
			ChainStatus:            chainStatus,
			TimeFrame:              tempJobs[i].TimeFrame,
			Recurring:              tempJobs[i].Recurring,
			TimeInterval:           tempJobs[i].TimeInterval,
			TriggerChainID:         tempJobs[i].TriggerChainID,
			TriggerContractAddress: tempJobs[i].TriggerContractAddress,
			TriggerEvent:           tempJobs[i].TriggerEvent,
			ScriptIPFSUrl:          tempJobs[i].ScriptIPFSUrl,
			ScriptTriggerFunction:  tempJobs[i].ScriptTriggerFunction,
			TargetChainID:          tempJobs[i].TargetChainID,
			TargetContractAddress:  tempJobs[i].TargetContractAddress,
			TargetFunction:         tempJobs[i].TargetFunction,
			ABI:                    tempJobs[i].ABI,
			ArgType:                tempJobs[i].ArgType,
			Arguments:              tempJobs[i].Arguments,
			ScriptTargetFunction:   tempJobs[i].ScriptTargetFunction,
			CreatedAt:              time.Now().UTC(),
			LastExecutedAt:         time.Time{},
		}

		success, err := h.SendDataToManager("/job/create", jobData)
		if err != nil {
			h.logger.Errorf("[CreateJobData] Error sending job data to manager for jobID %d: %v", currentJobID, err)
			http.Error(w, "Error sending job data to manager", http.StatusInternalServerError)
			return
		}

		if !success {
			h.logger.Errorf("[CreateJobData] Failed to send job data to manager for jobID %d", currentJobID)
			http.Error(w, "Failed to send job data to manager", http.StatusInternalServerError)
			return
		}

		h.logger.Infof("[CreateJobData] Successfully sent job data to manager for jobID %d", currentJobID)

		createdJobs.JobIDs[i] = currentJobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	existingJobIDs = append(existingJobIDs, newJobIDs...)

	// First get the current user points since it's part of the primary key
	var currentPoints float64
	if err := h.db.Session().Query(`
		SELECT user_points FROM triggerx.user_data 
		WHERE user_id = ?`,
		existingUserID).Scan(&currentPoints); err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error getting current points for userID %d: %v", existingUserID, err)
		http.Error(w, "Error getting current points: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.db.Session().Query(`
		UPDATE triggerx.user_data 
		SET job_ids = ?, last_updated_at = ?
		WHERE user_id = ? AND user_points = ?`,
		existingJobIDs, time.Now().UTC(), existingUserID, currentPoints).Exec(); err != nil {
		h.logger.Errorf("[CreateJobData] Error creating user data for userID %d: %v", existingUserID, err)
		http.Error(w, "Error creating user data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Infof("[CreateJobData] Updated user data for userID %d", existingUserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"message": "Database Updated Successfully",
		"Data":    createdJobs,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("[CreateJobData] Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Update a Job, and send it to the Manager
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

	success, err := h.SendDataToManager("/job/update", tempData)
	if err != nil {
		h.logger.Errorf("[UpdateJobData] Error sending job data to manager for jobID %d: %v", tempData.JobID, err)
		http.Error(w, "Error sending job data to manager", http.StatusInternalServerError)
		return
	}

	if !success {
		h.logger.Errorf("[UpdateJobData] Failed to send job data to manager for jobID %d", tempData.JobID)
		http.Error(w, "Failed to send job data to manager", http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateJobData] Successfully updated and published event for jobID %s", tempData.JobID)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateJobLastExecutedAt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[UpdateJobLastExecutedAt] Updating last executed at for jobID %s", jobID)

	if err := h.db.Session().Query(`
		UPDATE triggerx.job_data 
		SET last_executed_at = ?
		WHERE job_id = ?`,
		time.Now().UTC(), jobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating last executed at for jobID %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateJobLastExecutedAt] Successfully updated last executed at for jobID %s", jobID)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[GetJobData] Fetching data for jobID %s", jobID)

	var jobData types.JobData
	if err := h.db.Session().Query(`
        SELECT job_id, task_definition_id, user_id, priority, security, link_job_id, chain_status,
               time_frame, recurring, time_interval, trigger_chain_id, trigger_contract_address, 
               trigger_event, script_ipfs_url, script_trigger_function, target_chain_id, 
               target_contract_address, target_function, abi, arg_type, arguments, script_target_function, 
               status, job_cost_prediction, created_at, last_executed_at, task_ids
        FROM triggerx.job_data 
        WHERE job_id = ?`, jobID).Scan(
		&jobData.JobID, &jobData.TaskDefinitionID, &jobData.UserID, &jobData.Priority, &jobData.Security, &jobData.LinkJobID, &jobData.ChainStatus,
		&jobData.TimeFrame, &jobData.Recurring, &jobData.TimeInterval, &jobData.TriggerChainID, &jobData.TriggerContractAddress,
		&jobData.TriggerEvent, &jobData.ScriptIPFSUrl, &jobData.ScriptTriggerFunction, &jobData.TargetChainID,
		&jobData.TargetContractAddress, &jobData.TargetFunction, &jobData.ABI, &jobData.ArgType, &jobData.Arguments, &jobData.ScriptTargetFunction,
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
		JobID       int64 `json:"job_id"`
		JobType     int   `json:"job_type"`
		Status      bool  `json:"status"`
		ChainStatus int   `json:"chain_status"`
		LinkJobID   int64 `json:"link_job_id"`
	}

	var userJobs []JobSummary

	// First, get the user_id from the user_address
	var userID int64
	if err := h.db.Session().Query(`
		SELECT user_id 
		FROM triggerx.user_data 
		WHERE user_address = ? ALLOW FILTERING
	`, userAddress).Scan(&userID); err != nil {
		// Instead of returning a 404, return a 200 with a message
		h.logger.Infof("[GetJobsByUserAddress] User address %s not found", userAddress)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"message": "User address not registered",
			"jobs":    userJobs, // Return an empty list of jobs
		}
		w.WriteHeader(http.StatusOK) // Set status to 200
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Errorf("[GetJobsByUserAddress] Error encoding response for user_address %s: %v", userAddress, err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
		return
	}

	h.logger.Infof("[GetJobsByUserAddress] Found user_id %d for user_address %s", userID, userAddress)

	iter := h.db.Session().Query(`
        SELECT job_id, task_definition_id, status, chain_status, link_job_id
        FROM triggerx.job_data 
        WHERE user_id = ? ALLOW FILTERING
    `, userID).Iter()

	var job JobSummary
	for iter.Scan(&job.JobID, &job.JobType, &job.Status, &job.ChainStatus, &job.LinkJobID) {
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

func (h *Handler) DeleteJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	h.logger.Infof("[DeleteJobData] Deleting jobID %s", jobID)

	if err := h.db.Session().Query(`
		UPDATE triggerx.job_data 
        SET status = ?
        WHERE job_id = ?`,
		true, jobID).Exec(); err != nil {
		h.logger.Errorf("[DeleteJobData] Error deleting jobID %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[DeleteJobData] Successfully deleted jobID %s", jobID)
	w.WriteHeader(http.StatusOK)
}
