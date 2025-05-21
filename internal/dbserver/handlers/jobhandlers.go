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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting job data: " + err.Error()})
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

		success, err := h.SendDataToManager("/jobs/schedule", jobData)
		if err != nil {
			h.logger.Errorf("[CreateJobData] Error sending job data to manager for jobID %d: %v", currentJobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending job data to manager"})
			return
		}

		if !success {
			h.logger.Errorf("[CreateJobData] Failed to send job data to manager for jobID %d", currentJobID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send job data to manager"})
			return
		}

		h.logger.Infof("[CreateJobData] Successfully sent job data to manager for jobID %d", currentJobID)

		createdJobs.JobIDs[i] = currentJobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	existingJobIDs = append(existingJobIDs, newJobIDs...)

	if err := h.db.Session().Query(`
		UPDATE triggerx.user_data 
		SET job_ids = ?, last_updated_at = ?
		WHERE user_id = ?`,
		existingJobIDs, time.Now().UTC(), existingUserID).Exec(); err != nil {
		h.logger.Errorf("[CreateJobData] Error creating user data for userID %d: %v", existingUserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user data: " + err.Error()})
		return
	}
	h.logger.Infof("[CreateJobData] Updated user data for userID %d", existingUserID)

	response := map[string]interface{}{
		"message": "Database Updated Successfully",
		"Data":    createdJobs,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) UpdateJobData(c *gin.Context) {
	var tempData types.UpdateJobData
	if err := c.ShouldBindJSON(&tempData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request for jobID %s: %v", tempData.JobID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Session().Query(`
		UPDATE triggerx.job_data 	
		SET time_frame = ?, recurring = ?
		WHERE job_id = ?`,
		tempData.TimeFrame, tempData.Recurring, tempData.JobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating jobID %s: %v", tempData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	success, err := h.SendDataToManager("/job/update", tempData)
	if err != nil {
		h.logger.Errorf("[UpdateJobData] Error sending job data to manager for jobID %d: %v", tempData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending job data to manager"})
		return
	}

	if !success {
		h.logger.Errorf("[UpdateJobData] Failed to send job data to manager for jobID %d", tempData.JobID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send job data to manager"})
		return
	}

	h.logger.Infof("[UpdateJobData] Successfully updated and published event for jobID %s", tempData.JobID)
	c.Status(http.StatusOK)
}

func (h *Handler) UpdateJobLastExecutedAt(c *gin.Context) {
	jobID := c.Param("id")
	h.logger.Infof("[UpdateJobLastExecutedAt] Updating last executed at for jobID %s", jobID)

	if err := h.db.Session().Query(`
		UPDATE triggerx.job_data 
		SET last_executed_at = ?
		WHERE job_id = ?`,
		time.Now().UTC(), jobID).Exec(); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating last executed at for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateJobLastExecutedAt] Successfully updated last executed at for jobID %s", jobID)
	c.Status(http.StatusOK)
}

func (h *Handler) GetJobData(c *gin.Context) {
	jobID := c.Param("id")
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetJobData] Successfully retrieved jobID %s", jobID)
	c.JSON(http.StatusOK, jobData)
}

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("user_address"))
	h.logger.Infof("[GetJobsByUserAddress] Fetching jobs for user_address %s", userAddress)

	type JobSummary struct {
		JobID       int64 `json:"job_id"`
		JobType     int   `json:"job_type"`
		Status      bool  `json:"status"`
		ChainStatus int   `json:"chain_status"`
		LinkJobID   int64 `json:"link_job_id"`
	}

	var userJobs []JobSummary

	var userID int64
	if err := h.db.Session().Query(`
		SELECT user_id 
		FROM triggerx.user_data 
		WHERE user_address = ? ALLOW FILTERING
	`, userAddress).Scan(&userID); err != nil {
		h.logger.Infof("[GetJobsByUserAddress] User address %s not found", userAddress)
		c.JSON(http.StatusOK, gin.H{
			"message": "User address not registered",
			"jobs":    userJobs,
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving jobs: " + err.Error()})
		return
	}

	h.logger.Infof("[GetJobsByUserAddress] Retrieved %d jobs for user_address %s", len(userJobs), userAddress)

	c.JSON(http.StatusOK, userJobs)
}

func (h *Handler) DeleteJobData(c *gin.Context) {
	jobID := c.Param("id")
	h.logger.Infof("[DeleteJobData] Deleting jobID %s", jobID)

	if err := h.db.Session().Query(`
		UPDATE triggerx.job_data 
        SET status = ?
        WHERE job_id = ?`,
		true, jobID).Exec(); err != nil {
		h.logger.Errorf("[DeleteJobData] Error deleting jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[DeleteJobData] Successfully deleted jobID %s", jobID)
	c.Status(http.StatusOK)
}
