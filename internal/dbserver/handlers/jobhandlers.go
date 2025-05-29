package handlers

import (
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateJobData(c *gin.Context) {
	var tempJobs []types.CreateJobData
	if err := c.ShouldBindJSON(&tempJobs); err != nil {
		h.logger.Errorf("[CreateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(tempJobs) == 0 {
		h.logger.Error("[CreateJobData] No jobs provided in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No jobs provided"})
		return
	}

	// Set timestamps and timezone for all jobs
	now := time.Now().UTC()
	for i := range tempJobs {
		tempJobs[i].CreatedAt = now
		tempJobs[i].UpdatedAt = now
		tempJobs[i].LastExecutedAt = time.Time{} // Zero time for new jobs
		if tempJobs[i].Timezone == "" {
			tempJobs[i].Timezone = "UTC" // Default to UTC if not specified
		}
	}

	var ipfsURLs []string
	needsDynamicFee := false
	for i, job := range tempJobs {
		if job.ArgType == 2 {
			needsDynamicFee = true
			if job.ScriptIPFSUrl == "" {
				h.logger.Errorf("[CreateJobData] Missing IPFS URL for job %d", i)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing IPFS URL"})
				return
			}
			ipfsURLs = append(ipfsURLs, job.ScriptIPFSUrl)
		}
	}

	var feePerJob float64 = 0.01

	if needsDynamicFee {
		var totalFee float64
		var err error
		totalFee, err = h.CalculateTaskFees(strings.Join(ipfsURLs, ","))
		if err != nil {
			h.logger.Errorf("[CreateJobData] Error calculating fees: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating fees"})
			return
		}

		feePerJob = totalFee / float64(len(ipfsURLs))
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var existingUserID int64
	var existingAccountBalance *big.Int = big.NewInt(0)
	var existingTokenBalance *big.Int = big.NewInt(0)
	var existingJobIDs []int64 = []int64{}
	var newJobIDs []int64
	err := h.db.Session().Query(`
		SELECT user_id, account_balance, token_balance, job_ids
		FROM triggerx.user_data 
		WHERE user_address = ? ALLOW FILTERING`,
		strings.ToLower(tempJobs[0].UserAddress)).Scan(&existingUserID, &existingAccountBalance, &existingTokenBalance, &existingJobIDs)

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error checking user existence for address %s: %v", tempJobs[0].UserAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence: " + err.Error()})
		return
	}

	if err == gocql.ErrNotFound {
		h.logger.Infof("[CreateJobData] Creating new user for address %s", tempJobs[0].UserAddress)
		var maxUserID int64
		if err := h.db.Session().Query(`
			SELECT MAX(user_id) FROM triggerx.user_data
		`).Scan(&maxUserID); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting max user ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting max userID: " + err.Error()})
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
			existingUserID, strings.ToLower(tempJobs[0].UserAddress), time.Now().UTC(),
			[]int64{}, existingAccountBalance, existingTokenBalance, time.Now().UTC(), 0.0).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error creating user data for userID %d: %v", existingUserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user data: " + err.Error()})
			return
		}
		h.logger.Infof("[CreateJobData] Created new user with userID %d", existingUserID)
	} else {
		existingAccountBalance = new(big.Int).Add(existingAccountBalance, tempJobs[0].StakeAmount)
		existingTokenBalance = new(big.Int).Add(existingTokenBalance, tempJobs[0].TokenAmount)

		if err := h.db.Session().Query(`
			UPDATE triggerx.user_data 
			SET account_balance = ?, token_balance = ?, last_updated_at = ?
			WHERE user_id = ?`,
			existingAccountBalance, existingTokenBalance,
			time.Now().UTC(), existingUserID).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user data for userID %d: %v", existingUserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user data: " + err.Error()})
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

		// Insert into main job_data table
		if err := h.db.Session().Query(`
			INSERT INTO triggerx.job_data (
				job_id, job_title, task_definition_id, user_id, link_job_id, chain_status,
				custom, time_frame, recurring, status, job_cost_prediction, task_ids,
				created_at, updated_at, last_executed_at, timezone
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			currentJobID, tempJobs[i].JobTitle, tempJobs[i].TaskDefinitionID, existingUserID, linkJobID, chainStatus,
			tempJobs[i].Custom, tempJobs[i].TimeFrame, tempJobs[i].Recurring, false, tempJobs[i].JobCostPrediction, []int64{},
			tempJobs[i].CreatedAt, tempJobs[i].UpdatedAt, tempJobs[i].LastExecutedAt, tempJobs[i].Timezone).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error inserting job data for jobID %d: %v", currentJobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting job data: " + err.Error()})
			return
		}

		// Insert into appropriate job type table based on task_definition_id
		switch {
		case tempJobs[i].TaskDefinitionID == 1 || tempJobs[i].TaskDefinitionID == 2:
			// Time-based job
			if err := h.db.Session().Query(`
				INSERT INTO triggerx.time_job_data (
					job_id, time_frame, recurring, time_interval,
					target_chain_id, target_contract_address, target_function,
					abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				currentJobID, tempJobs[i].TimeFrame, tempJobs[i].Recurring, tempJobs[i].TimeInterval,
				tempJobs[i].TargetChainID, tempJobs[i].TargetContractAddress, tempJobs[i].TargetFunction,
				tempJobs[i].ABI, tempJobs[i].ArgType, tempJobs[i].Arguments, tempJobs[i].ScriptIPFSUrl).Exec(); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting time job data for jobID %d: %v", currentJobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting time job data: " + err.Error()})
				return
			}
			h.logger.Infof("[CreateJobData] Successfully created time-based job %d with interval %d seconds",
				currentJobID, tempJobs[i].TimeInterval)
		case tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4:
			// Event-based job
			if err := h.db.Session().Query(`
				INSERT INTO triggerx.event_job_data (
					job_id, time_frame, recurring,
					trigger_chain_id, trigger_contract_address, trigger_event,
					target_chain_id, target_contract_address, target_function,
					abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				currentJobID, tempJobs[i].TimeFrame, tempJobs[i].Recurring,
				tempJobs[i].TriggerChainID, tempJobs[i].TriggerContractAddress, tempJobs[i].TriggerEvent,
				tempJobs[i].TargetChainID, tempJobs[i].TargetContractAddress, tempJobs[i].TargetFunction,
				tempJobs[i].ABI, tempJobs[i].ArgType, tempJobs[i].Arguments, tempJobs[i].ScriptIPFSUrl).Exec(); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting event job data for jobID %d: %v", currentJobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting event job data: " + err.Error()})
				return
			}
			h.logger.Infof("[CreateJobData] Successfully created event-based job %d for event %s on contract %s",
				currentJobID, tempJobs[i].TriggerEvent, tempJobs[i].TriggerContractAddress)
		case tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6:
			// Condition-based job
			if err := h.db.Session().Query(`
				INSERT INTO triggerx.condition_job_data (
					job_id, time_frame, recurring,
					condition_type, upper_limit, lower_limit,
					value_source_type, value_source_url,
					target_chain_id, target_contract_address, target_function,
					abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				currentJobID, tempJobs[i].TimeFrame, tempJobs[i].Recurring,
				tempJobs[i].ConditionType, tempJobs[i].UpperLimit, tempJobs[i].LowerLimit,
				tempJobs[i].ValueSourceType, tempJobs[i].ValueSourceUrl,
				tempJobs[i].TargetChainID, tempJobs[i].TargetContractAddress, tempJobs[i].TargetFunction,
				tempJobs[i].ABI, tempJobs[i].ArgType, tempJobs[i].Arguments, tempJobs[i].ScriptIPFSUrl).Exec(); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting condition job data for jobID %d: %v", currentJobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting condition job data: " + err.Error()})
				return
			}
			h.logger.Infof("[CreateJobData] Successfully created condition-based job %d with condition type %s (limits: %f-%f)",
				currentJobID, tempJobs[i].ConditionType, tempJobs[i].LowerLimit, tempJobs[i].UpperLimit)
		default:
			h.logger.Errorf("[CreateJobData] Invalid task definition ID %d for job %d", tempJobs[i].TaskDefinitionID, i)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		pointsToAdd := 10.0
		if tempJobs[i].Custom {
			pointsToAdd = 20.0
		}

		var currentPoints float64
		if err := h.db.Session().Query(`
			SELECT user_points FROM triggerx.user_data 
			WHERE user_id = ?`,
			existingUserID).Scan(&currentPoints); err != nil && err != gocql.ErrNotFound {
			h.logger.Errorf("[CreateJobData] Error getting current points for userID %d: %v", existingUserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting current points: " + err.Error()})
			return
		}

		newPoints := currentPoints + pointsToAdd
		if err := h.db.Session().Query(`
			UPDATE triggerx.user_data 
			SET user_points = ?, last_updated_at = ?
			WHERE user_id = ?`,
			newPoints, time.Now().UTC(), existingUserID).Exec(); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user points for userID %d: %v", existingUserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user points: " + err.Error()})
			return
		}

		createdJobs.JobIDs[i] = currentJobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	// Update user's job_ids
	allJobIDs := append(existingJobIDs, newJobIDs...)
	if err := h.db.Session().Query(`
		UPDATE triggerx.user_data 
		SET job_ids = ?, last_updated_at = ?
		WHERE user_id = ?`,
		allJobIDs, time.Now().UTC(), existingUserID).Exec(); err != nil {
		h.logger.Errorf("[CreateJobData] Error updating user job IDs for userID %d: %v", existingUserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user job IDs: " + err.Error()})
		return
	}
	h.logger.Infof("[CreateJobData] Successfully updated user %d with %d total jobs", existingUserID, len(allJobIDs))

	c.JSON(http.StatusOK, createdJobs)
	h.logger.Infof("[CreateJobData] Successfully completed job creation for user %d with %d new jobs",
		existingUserID, len(tempJobs))
}

func (h *Handler) UpdateJobData(c *gin.Context) {
	var updateData types.UpdateJobData
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set update timestamp
	updateData.UpdatedAt = time.Now().UTC()

	// Get job type first
	var taskDefinitionID int
	if err := h.db.Session().Query(`
		SELECT task_definition_id FROM triggerx.job_data 
		WHERE job_id = ?`,
		updateData.JobID).Scan(&taskDefinitionID); err != nil {
		h.logger.Errorf("[UpdateJobData] Error getting task definition ID for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting job type: " + err.Error()})
		return
	}

	// Update main job_data table with just updated_at
	if err := h.db.Session().Query(`
		UPDATE triggerx.job_data 
		SET updated_at = ?
		WHERE job_id = ?`,
		updateData.UpdatedAt, updateData.JobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
	}

	// Update specific job type table
	switch {
	case taskDefinitionID == 1 || taskDefinitionID == 2:
		// Time-based job
		if err := h.db.Session().Query(`
			UPDATE triggerx.time_job_data 
			SET time_frame = ?, recurring = ?, updated_at = ?
			WHERE job_id = ?`,
			updateData.TimeFrame, updateData.Recurring, updateData.UpdatedAt,
			updateData.JobID).Exec(); err != nil {
			h.logger.Errorf("[UpdateJobData] Error updating time job data for jobID %d: %v", updateData.JobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job data: " + err.Error()})
			return
		}
	case taskDefinitionID == 3 || taskDefinitionID == 4:
		// Event-based job
		if err := h.db.Session().Query(`
			UPDATE triggerx.event_job_data 
			SET time_frame = ?, recurring = ?, updated_at = ?
			WHERE job_id = ?`,
			updateData.TimeFrame, updateData.Recurring, updateData.UpdatedAt,
			updateData.JobID).Exec(); err != nil {
			h.logger.Errorf("[UpdateJobData] Error updating event job data for jobID %d: %v", updateData.JobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job data: " + err.Error()})
			return
		}
	case taskDefinitionID == 5 || taskDefinitionID == 6:
		// Condition-based job
		if err := h.db.Session().Query(`
			UPDATE triggerx.condition_job_data 
			SET time_frame = ?, recurring = ?, updated_at = ?
			WHERE job_id = ?`,
			updateData.TimeFrame, updateData.Recurring, updateData.UpdatedAt,
			updateData.JobID).Exec(); err != nil {
			h.logger.Errorf("[UpdateJobData] Error updating condition job data for jobID %d: %v", updateData.JobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job data: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Job updated successfully",
		"job_id":     updateData.JobID,
		"updated_at": updateData.UpdatedAt,
	})
}

func (h *Handler) UpdateJobLastExecutedAt(c *gin.Context) {
	var updateData struct {
		JobID          int64     `json:"job_id" binding:"required"`
		LastExecutedAt time.Time `json:"last_executed_at" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure timestamp is in UTC
	updateData.LastExecutedAt = updateData.LastExecutedAt.UTC()
	now := time.Now().UTC()

	// Get job type first
	var taskDefinitionID int
	if err := h.db.Session().Query(`
		SELECT task_definition_id FROM triggerx.job_data 
		WHERE job_id = ?`,
		updateData.JobID).Scan(&taskDefinitionID); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error getting task definition ID for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting job type: " + err.Error()})
		return
	}

	// Update main job_data table
	if err := h.db.Session().Query(`
		UPDATE triggerx.job_data 
		SET last_executed_at = ?, updated_at = ?
		WHERE job_id = ?`,
		updateData.LastExecutedAt, now, updateData.JobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
	}

	// Update specific job type table
	switch {
	case taskDefinitionID == 1 || taskDefinitionID == 2:
		// Time-based job
		if err := h.db.Session().Query(`
			UPDATE triggerx.time_job_data 
			SET last_executed_at = ?, updated_at = ?
			WHERE job_id = ?`,
			updateData.LastExecutedAt, now, updateData.JobID).Exec(); err != nil {
			h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating time job data for jobID %d: %v", updateData.JobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job data: " + err.Error()})
			return
		}
	case taskDefinitionID == 3 || taskDefinitionID == 4:
		// Event-based job
		if err := h.db.Session().Query(`
			UPDATE triggerx.event_job_data 
			SET last_executed_at = ?, updated_at = ?
			WHERE job_id = ?`,
			updateData.LastExecutedAt, now, updateData.JobID).Exec(); err != nil {
			h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating event job data for jobID %d: %v", updateData.JobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job data: " + err.Error()})
			return
		}
	case taskDefinitionID == 5 || taskDefinitionID == 6:
		// Condition-based job
		if err := h.db.Session().Query(`
			UPDATE triggerx.condition_job_data 
			SET last_executed_at = ?, updated_at = ?
			WHERE job_id = ?`,
			updateData.LastExecutedAt, now, updateData.JobID).Exec(); err != nil {
			h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating condition job data for jobID %d: %v", updateData.JobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job data: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Last executed time updated successfully",
		"job_id":           updateData.JobID,
		"last_executed_at": updateData.LastExecutedAt,
		"updated_at":       now,
	})
}

func (h *Handler) GetJobData(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		h.logger.Error("[GetJobData] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No job ID provided"})
		return
	}

	var jobData types.JobData
	if err := h.db.Session().Query(`
		SELECT job_id, job_title, task_definition_id, user_id, link_job_id, chain_status,
			custom, time_frame, recurring, status, job_cost_prediction, task_ids
		FROM triggerx.job_data 
		WHERE job_id = ?`,
		jobID).Scan(&jobData.JobID, &jobData.JobTitle, &jobData.TaskDefinitionID, &jobData.UserID,
		&jobData.LinkJobID, &jobData.ChainStatus, &jobData.Custom, &jobData.TimeFrame,
		&jobData.Recurring, &jobData.Status, &jobData.JobCostPrediction, &jobData.TaskIDs); err != nil {
		h.logger.Errorf("[GetJobData] Error getting job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting job data: " + err.Error()})
		return
	}

	// Check task_definition_id to determine job type
	switch {
	case jobData.TaskDefinitionID == 1 || jobData.TaskDefinitionID == 2:
		// Time-based job
		var timeJobData types.TimeJobData
		if err := h.db.Session().Query(`
			SELECT job_id, time_frame, recurring, time_interval,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			FROM triggerx.time_job_data 
			WHERE job_id = ?`,
			jobID).Scan(&timeJobData.JobID, &timeJobData.TimeFrame, &timeJobData.Recurring, &timeJobData.TimeInterval,
			&timeJobData.TargetChainID, &timeJobData.TargetContractAddress, &timeJobData.TargetFunction,
			&timeJobData.ABI, &timeJobData.ArgType, &timeJobData.Arguments, &timeJobData.DynamicArgumentsScriptIPFSUrl); err != nil {
			h.logger.Errorf("[GetJobData] Error getting time job data for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting time job data: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, timeJobData)
		return

	case jobData.TaskDefinitionID == 3 || jobData.TaskDefinitionID == 4:
		// Event-based job
		var eventJobData types.EventJobData
		if err := h.db.Session().Query(`
			SELECT job_id, time_frame, recurring,
				trigger_chain_id, trigger_contract_address, trigger_event,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			FROM triggerx.event_job_data 
			WHERE job_id = ?`,
			jobID).Scan(&eventJobData.JobID, &eventJobData.TimeFrame, &eventJobData.Recurring,
			&eventJobData.TriggerChainID, &eventJobData.TriggerContractAddress, &eventJobData.TriggerEvent,
			&eventJobData.TargetChainID, &eventJobData.TargetContractAddress, &eventJobData.TargetFunction,
			&eventJobData.ABI, &eventJobData.ArgType, &eventJobData.Arguments, &eventJobData.DynamicArgumentsScriptIPFSUrl); err != nil {
			h.logger.Errorf("[GetJobData] Error getting event job data for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting event job data: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, eventJobData)
		return

	case jobData.TaskDefinitionID == 5 || jobData.TaskDefinitionID == 6:
		// Condition-based job
		var conditionJobData types.ConditionJobData
		if err := h.db.Session().Query(`
			SELECT job_id, time_frame, recurring,
				condition_type, upper_limit, lower_limit,
				value_source_type, value_source_url,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			FROM triggerx.condition_job_data 
			WHERE job_id = ?`,
			jobID).Scan(&conditionJobData.JobID, &conditionJobData.TimeFrame, &conditionJobData.Recurring,
			&conditionJobData.ConditionType, &conditionJobData.UpperLimit, &conditionJobData.LowerLimit,
			&conditionJobData.ValueSourceType, &conditionJobData.ValueSourceUrl,
			&conditionJobData.TargetChainID, &conditionJobData.TargetContractAddress, &conditionJobData.TargetFunction,
			&conditionJobData.ABI, &conditionJobData.ArgType, &conditionJobData.Arguments, &conditionJobData.DynamicArgumentsScriptIPFSUrl); err != nil {
			h.logger.Errorf("[GetJobData] Error getting condition job data for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting condition job data: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, conditionJobData)
		return

	default:
		// Return basic job data if task_definition_id is not recognized
		c.JSON(http.StatusOK, jobData)
		return
	}
}

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	userAddress := c.Param("user_address")
	if userAddress == "" {
		h.logger.Error("[GetJobsByUserAddress] No user address provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user address provided"})
		return
	}

	var userID int64
	var jobIDs []int64
	if err := h.db.Session().Query(`
		SELECT user_id, job_ids
		FROM triggerx.user_data 
		WHERE user_address = ? ALLOW FILTERING`,
		strings.ToLower(userAddress)).Scan(&userID, &jobIDs); err != nil {
		h.logger.Errorf("[GetJobsByUserAddress] Error getting user data for address %s: %v", userAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user data: " + err.Error()})
		return
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	var jobs []interface{}
	for _, jobID := range jobIDs {
		// Get basic job data
		var jobData types.JobData
		if err := h.db.Session().Query(`
			SELECT job_id, job_title, task_definition_id, user_id, link_job_id, chain_status,
				custom, time_frame, recurring, status, job_cost_prediction, task_ids
			FROM triggerx.job_data 
			WHERE job_id = ?`,
			jobID).Scan(&jobData.JobID, &jobData.JobTitle, &jobData.TaskDefinitionID, &jobData.UserID,
			&jobData.LinkJobID, &jobData.ChainStatus, &jobData.Custom, &jobData.TimeFrame,
			&jobData.Recurring, &jobData.Status, &jobData.JobCostPrediction, &jobData.TaskIDs); err != nil {
			h.logger.Errorf("[GetJobsByUserAddress] Error getting job data for jobID %d: %v", jobID, err)
			continue
		}

		// Check task_definition_id to determine job type
		switch {
		case jobData.TaskDefinitionID == 1 || jobData.TaskDefinitionID == 2:
			// Time-based job
			var timeJobData types.TimeJobData
			if err := h.db.Session().Query(`
				SELECT job_id, time_frame, recurring, time_interval,
					target_chain_id, target_contract_address, target_function,
					abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
				FROM triggerx.time_job_data 
				WHERE job_id = ?`,
				jobID).Scan(&timeJobData.JobID, &timeJobData.TimeFrame, &timeJobData.Recurring, &timeJobData.TimeInterval,
				&timeJobData.TargetChainID, &timeJobData.TargetContractAddress, &timeJobData.TargetFunction,
				&timeJobData.ABI, &timeJobData.ArgType, &timeJobData.Arguments, &timeJobData.DynamicArgumentsScriptIPFSUrl); err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting time job data for jobID %d: %v", jobID, err)
				continue
			}
			jobs = append(jobs, timeJobData)

		case jobData.TaskDefinitionID == 3 || jobData.TaskDefinitionID == 4:
			// Event-based job
			var eventJobData types.EventJobData
			if err := h.db.Session().Query(`
				SELECT job_id, time_frame, recurring,
					trigger_chain_id, trigger_contract_address, trigger_event,
					target_chain_id, target_contract_address, target_function,
					abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
				FROM triggerx.event_job_data 
				WHERE job_id = ?`,
				jobID).Scan(&eventJobData.JobID, &eventJobData.TimeFrame, &eventJobData.Recurring,
				&eventJobData.TriggerChainID, &eventJobData.TriggerContractAddress, &eventJobData.TriggerEvent,
				&eventJobData.TargetChainID, &eventJobData.TargetContractAddress, &eventJobData.TargetFunction,
				&eventJobData.ABI, &eventJobData.ArgType, &eventJobData.Arguments, &eventJobData.DynamicArgumentsScriptIPFSUrl); err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting event job data for jobID %d: %v", jobID, err)
				continue
			}
			jobs = append(jobs, eventJobData)

		case jobData.TaskDefinitionID == 5 || jobData.TaskDefinitionID == 6:
			// Condition-based job
			var conditionJobData types.ConditionJobData
			if err := h.db.Session().Query(`
				SELECT job_id, time_frame, recurring,
					condition_type, upper_limit, lower_limit,
					value_source_type, value_source_url,
					target_chain_id, target_contract_address, target_function,
					abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
				FROM triggerx.condition_job_data 
				WHERE job_id = ?`,
				jobID).Scan(&conditionJobData.JobID, &conditionJobData.TimeFrame, &conditionJobData.Recurring,
				&conditionJobData.ConditionType, &conditionJobData.UpperLimit, &conditionJobData.LowerLimit,
				&conditionJobData.ValueSourceType, &conditionJobData.ValueSourceUrl,
				&conditionJobData.TargetChainID, &conditionJobData.TargetContractAddress, &conditionJobData.TargetFunction,
				&conditionJobData.ABI, &conditionJobData.ArgType, &conditionJobData.Arguments, &conditionJobData.DynamicArgumentsScriptIPFSUrl); err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting condition job data for jobID %d: %v", jobID, err)
				continue
			}
			jobs = append(jobs, conditionJobData)

		default:
			// Add basic job data if task_definition_id is not recognized
			jobs = append(jobs, jobData)
		}
	}

	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) DeleteJobData(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		h.logger.Error("[DeleteJobData] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No job ID provided"})
		return
	}

	// Delete from all possible job type tables
	if err := h.db.Session().Query(`
		DELETE FROM triggerx.time_job_data 
		WHERE job_id = ?`,
		jobID).Exec(); err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[DeleteJobData] Error deleting time job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting time job data: " + err.Error()})
		return
	}

	if err := h.db.Session().Query(`
		DELETE FROM triggerx.event_job_data 
		WHERE job_id = ?`,
		jobID).Exec(); err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[DeleteJobData] Error deleting event job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting event job data: " + err.Error()})
		return
	}

	if err := h.db.Session().Query(`
		DELETE FROM triggerx.condition_job_data 
		WHERE job_id = ?`,
		jobID).Exec(); err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[DeleteJobData] Error deleting condition job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting condition job data: " + err.Error()})
		return
	}

	// Finally delete from the main job_data table
	if err := h.db.Session().Query(`
		DELETE FROM triggerx.job_data 
		WHERE job_id = ?`,
		jobID).Exec(); err != nil {
		h.logger.Errorf("[DeleteJobData] Error deleting job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting job data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}
