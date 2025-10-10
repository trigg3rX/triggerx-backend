package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTimeBasedTasks(c *gin.Context) {
	pollLookAhead := config.GetPollingLookAhead()
	lookAheadTime := time.Now().Add(time.Duration(pollLookAhead) * time.Second)

	// Get all time jobs
	trackDBOp := metrics.TrackDBOperation("read", "time_jobs")
	allTimeJobs, err := h.timeJobRepository.List(c.Request.Context())
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedTasks] Error retrieving time jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve time based tasks",
			"code":  "TIME_TASKS_FETCH_ERROR",
		})
		return
	}

	// Filter jobs by next execution timestamp
	var filteredJobs []*types.TimeJobDataEntity
	for _, job := range allTimeJobs {
		if !job.IsCompleted && job.NextExecutionTimestamp.Before(lookAheadTime) {
			filteredJobs = append(filteredJobs, job)
		}
	}

	// Convert to ScheduleTimeTaskData format
	var tasks []types.ScheduleTimeTaskData
	for _, timeJob := range filteredJobs {
		// Create task for this job
		newTask := &types.TaskDataEntity{
			TaskID:               0, // Will be auto-generated
			TaskNumber:           0,
			JobID:                timeJob.JobID,
			TaskDefinitionID:     timeJob.TaskDefinitionID,
			CreatedAt:            time.Now().UTC(),
			TaskOpxPredictedCost: "0",
			TaskOpxActualCost:    "0",
			IsSuccessful:         false,
			IsAccepted:           false,
		}

		trackDBOp = metrics.TrackDBOperation("create", "task_data")
		err := h.taskRepository.Create(c.Request.Context(), newTask)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[GetTimeBasedJobs] Error creating task data: %v", err)
			continue
		}

		// Get created task to get ID
		task, err := h.taskRepository.GetByID(c.Request.Context(), newTask.TaskID)
		if err != nil || task == nil {
			continue
		}

		// Update job with new task ID
		job, err := h.jobRepository.GetByID(c.Request.Context(), timeJob.JobID)
		if err == nil && job != nil {
			job.TaskIDs = append(job.TaskIDs, task.TaskID)
			err = h.jobRepository.Update(c.Request.Context(), job)
			if err != nil {
				h.logger.Errorf("[GetTimeBasedJobs] Error updating job with new task ID: %v", err)
				continue
			}
		}

		// Build response
		scheduleData := types.ScheduleTimeTaskData{
			TaskID:                 task.TaskID,
			TaskDefinitionID:       timeJob.TaskDefinitionID,
			LastExecutedAt:         timeJob.LastExecutedAt,
			ExpirationTime:         timeJob.ExpirationTime,
			NextExecutionTimestamp: timeJob.NextExecutionTimestamp,
			ScheduleType:           timeJob.ScheduleType,
			TimeInterval:           timeJob.TimeInterval,
			CronExpression:         timeJob.CronExpression,
			SpecificSchedule:       timeJob.SpecificSchedule,
			TaskTargetData: types.TaskTargetData{
				JobID:                     timeJob.JobID,
				TaskDefinitionID:          timeJob.TaskDefinitionID,
				TargetChainID:             timeJob.TargetChainID,
				TargetContractAddress:     timeJob.TargetContractAddress,
				TargetFunction:            timeJob.TargetFunction,
				ABI:                       timeJob.ABI,
				ArgType:                   timeJob.ArgType,
				Arguments:                 timeJob.Arguments,
				DynamicArgumentsScriptUrl: timeJob.DynamicArgumentsScriptURL,
			},
		}
		tasks = append(tasks, scheduleData)
	}

	// h.logger.Infof("[GetTimeBasedJobs] Successfully retrieved %d time based jobs", len(tasks))
	c.JSON(http.StatusOK, tasks)
}
