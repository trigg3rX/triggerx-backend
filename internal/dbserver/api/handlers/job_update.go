package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) DeleteJobData(c *gin.Context) {
	logger := h.getLogger(c)
	jobID := c.Param("id")
	if !types.IsValidJobID(jobID) {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "Invalid job ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("PUT [DeleteJobData] Job ID: %s", jobID)

	// Track database operation
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	job, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
	trackDBOp(err)
	if err != nil || job == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	// Track job status update
	job.Status = string(types.JobStatusDeleted)
	job.UpdatedAt = time.Now().UTC()

	trackDBOp = metrics.TrackDBOperation("update", "job_data")
	err = h.jobRepository.Update(c.Request.Context(), job)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch job.TaskDefinitionID {
	case 1, 2:
		trackDBOp = metrics.TrackDBOperation("update", "time_job")
		timeJob, err := h.timeJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
		if err == nil && timeJob != nil {
			timeJob.IsCompleted = true
			err = h.timeJobRepository.Update(c.Request.Context(), timeJob)
		}
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4:
		trackDBOp = metrics.TrackDBOperation("update", "event_job")
		eventJob, err := h.eventJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
		if err == nil && eventJob != nil {
			eventJob.IsCompleted = true
			err = h.eventJobRepository.Update(c.Request.Context(), eventJob)
		}
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobID)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to event scheduler: " + err.Error()})
			return
		}
	case 5, 6:
		trackDBOp = metrics.TrackDBOperation("update", "condition_job")
		conditionJob, err := h.conditionJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
		if err == nil && conditionJob != nil {
			conditionJob.IsCompleted = true
			err = h.conditionJobRepository.Update(c.Request.Context(), conditionJob)
		}
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobID)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to condition scheduler: " + err.Error()})
			return
		}
	}

	logger.Infof("PUT [DeleteJobData] Successful, job ID: %s", jobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func (h *Handler) UpdateJobDataFromUser(c *gin.Context) {
	logger := h.getLogger(c)
	var req types.UpdateJobDataFromUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("PUT [UpdateJobDataFromUser] Job ID: %s", req.JobID)

	// Get the job first
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	job, err := h.jobRepository.GetByID(c.Request.Context(), req.JobID)
	trackDBOp(err)
	if err != nil || job == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}
	var notify bool
	expirationTime := job.UpdatedAt.Add(time.Duration(req.TimeFrame) * time.Second)
	if job.Recurring == req.Recurring {
		// No need to notify the condition scheduler
		notify = false
	} else {
		// Need to notify the condition scheduler
		notify = true
		job.Recurring = req.Recurring
	}

	// Update job fields from request
	job.JobTitle = req.JobTitle
	job.UpdatedAt = time.Now().UTC()
	job.TimeFrame = req.TimeFrame
	job.JobCostPrediction = req.JobCostPrediction

	trackDBOp = metrics.TrackDBOperation("update", "job_data")
	err = h.jobRepository.Update(c.Request.Context(), job)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}

	// For time-based jobs, update time_interval and next_execution_timestamp
	switch job.TaskDefinitionID {
	case 1, 2:
		trackDBOp = metrics.TrackDBOperation("update", "time_job")
		timeJob := &types.TimeJobDataEntity{
			JobID:                  req.JobID,
			TimeInterval:           req.TimeInterval,
			NextExecutionTimestamp: job.UpdatedAt.Add(time.Duration(req.TimeInterval) * time.Second),
			ExpirationTime:         expirationTime,
		}
		err = h.timeJobRepository.Update(c.Request.Context(), timeJob)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
			return
		}
		logger.Infof("PUT [UpdateJobDataFromUser] Successful, job ID: %s", req.JobID)
		c.JSON(http.StatusOK, gin.H{
			"message":    "Job updated successfully",
			"job_id":     req.JobID,
			"updated_at": time.Now().UTC(),
		})
	case 3, 4:
		updateData := &types.EventJobDataEntity{
			JobID:          req.JobID,
			Recurring:      req.Recurring,
			ExpirationTime: expirationTime,
		}
		err = h.eventJobRepository.Update(c.Request.Context(), updateData)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		}

		if notify {
			success, err := h.notifyUpdateToConditionScheduler(req.JobID, req.Recurring, expirationTime)
			if err != nil {
				logger.Errorf("Error sending update to condition scheduler: %v", err)
			}
			if !success {
				logger.Errorf("Error sending update to condition scheduler: %v", err)
			}
		}
		logger.Infof("PUT [UpdateJobDataFromUser] Successful, job ID: %s", req.JobID)
		c.JSON(http.StatusOK, gin.H{
			"message":    "Job updated successfully",
			"job_id":     req.JobID,
			"updated_at": time.Now().UTC(),
		})
	case 5, 6:
		updateData := &types.ConditionJobDataEntity{
			JobID:          req.JobID,
			Recurring:      req.Recurring,
			ExpirationTime: expirationTime,
		}
		err = h.conditionJobRepository.Update(c.Request.Context(), updateData)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		}
		if notify {
			success, err := h.notifyUpdateToConditionScheduler(req.JobID, req.Recurring, expirationTime)
			if err != nil {
				logger.Errorf("Error sending update to condition scheduler: %v", err)
			}
			if !success {
				logger.Errorf("Error sending update to condition scheduler: %v", err)
			}
		}
		logger.Infof("PUT [UpdateJobDataFromUser] Successful, job ID: %s", req.JobID)
		c.JSON(http.StatusOK, gin.H{
			"message":    "Job updated successfully",
			"job_id":     req.JobID,
			"updated_at": time.Now().UTC(),
		})
	}
}
