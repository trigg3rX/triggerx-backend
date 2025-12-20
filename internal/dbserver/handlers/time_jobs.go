package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) GetTimeBasedTasks(c *gin.Context) {
	pollLookAhead := config.GetPollingLookAhead()
	lookAheadTime := time.Now().Add(time.Duration(pollLookAhead) * time.Second)

	var tasks []commonTypes.ScheduleTimeTaskData
	var err error

	// Get regular time-based jobs
	trackDBOp := metrics.TrackDBOperation("read", "time_jobs")
	tasks, err = h.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(lookAheadTime)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTimeBasedTasks] Error retrieving time based tasks", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve time based tasks",
			"code":  "TIME_TASKS_FETCH_ERROR",
			"tasks": tasks,
		})
		return
	}

	// Get custom jobs (TaskDefinitionID = 7)
	if h.customJobRepository != nil {
		trackDBOp = metrics.TrackDBOperation("read", "custom_jobs")
		customJobs, err := h.customJobRepository.GetCustomJobsDueForExecution(lookAheadTime)
		trackDBOp(err)
		if err != nil {
			h.logger.Warn(c.Request.Context(), "[GetCustomBasedTasks] Error retrieving custom jobs", observability.Error(err))
			// Don't fail, just log and continue with time jobs only
		} else {
			// Convert custom jobs to ScheduleTimeTaskData format
			for _, customJob := range customJobs {
				taskData := h.convertCustomJobToScheduleTimeTaskData(c.Request.Context(), &customJob)
				tasks = append(tasks, taskData)
			}
			h.logger.Info(c.Request.Context(), "[getCustomBasedTasks] Retrieved %d custom jobs", observability.Int("custom_jobs_count", len(customJobs)))
		}
	}

	for i := range tasks {
		trackDBOp = metrics.TrackDBOperation("create", "task_data")
		taskID, err := h.taskRepository.CreateTaskDataInDB(c.Request.Context(), &types.CreateTaskDataRequest{
			JobID:            tasks[i].TaskTargetData.JobID.Int,
			TaskDefinitionID: tasks[i].TaskDefinitionID,
		})
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[GetTimeBasedJobs] Error creating task data", observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create task data",
				"code":  "TASK_CREATION_ERROR",
			})
			continue
		}

		trackDBOp = metrics.TrackDBOperation("update", "add_task_id")
		err = h.taskRepository.AddTaskIDToJob(tasks[i].TaskTargetData.JobID.Int, taskID)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[GetTimeBasedJobs] Error adding task ID", observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to add task ID",
				"code":  "TASK_ID_ADDITION_ERROR",
			})
			continue
		}

		tasks[i].TaskID = taskID
	}

	if len(tasks) != 0 {
		h.logger.Info(c.Request.Context(), "[GetTimeBasedJobs] Successfully retrieved %d time based jobs", observability.Int("time_based_jobs_count", len(tasks)))
	}
	c.JSON(http.StatusOK, tasks)
}

// convertCustomJobToScheduleTimeTaskData converts a CustomJobData to ScheduleTimeTaskData format
func (h *Handler) convertCustomJobToScheduleTimeTaskData(ctx context.Context, customJob *commonTypes.CustomJobData) commonTypes.ScheduleTimeTaskData {
	// Fetch storage for this custom job
	storage, err := h.scriptStorageRepository.GetStorageByJobID(customJob.JobID.ToBigInt())
	if err != nil {
		h.logger.Warn(ctx, "[GetTimeBasedTasks] Failed to get storage for job %s: %v", observability.String("job_id", customJob.JobID.String()), observability.Error(err))
		storage = make(map[string]string) // Continue with empty storage
	}

	return commonTypes.ScheduleTimeTaskData{
		TaskID:                 0, // Will be assigned during task creation
		TaskDefinitionID:       7, // Custom job task definition ID
		LastExecutedAt:         customJob.LastExecutedAt,
		ExpirationTime:         customJob.ExpirationTime,
		NextExecutionTimestamp: customJob.NextExecutionTime,
		ScheduleType:           "interval",
		TimeInterval:           customJob.TimeInterval,
		CronExpression:         "",
		SpecificSchedule:       "",
		TaskTargetData: commonTypes.TaskTargetData{
			JobID:                     customJob.JobID,
			TaskID:                    0, // Will be assigned later
			TaskDefinitionID:          7,
			TargetChainID:             customJob.TargetChainID, // Will be filled by script output
			TargetContractAddress:     "", // Will be filled by script output
			TargetFunction:            "", // Will be filled by script output
			ABI:                       "",
			ArgType:                   0,
			Arguments:                 []string{},
			DynamicArgumentsScriptUrl: customJob.CustomScriptUrl, // Use this field to pass script URL
			IsImua:                    false,
			// Custom script fields
			ScriptStorage:             storage,              // Storage from database
			ScriptLanguage:            customJob.ScriptLanguage,
		},
		IsImua: false,
	}
}
