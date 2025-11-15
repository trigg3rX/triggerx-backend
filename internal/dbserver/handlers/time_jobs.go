package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTimeBasedTasks(c *gin.Context) {
	pollLookAhead := config.GetPollingLookAhead()
	lookAheadTime := time.Now().Add(time.Duration(pollLookAhead) * time.Second)

	var tasks []commonTypes.ScheduleTimeTaskData
	var err error

	trackDBOp := metrics.TrackDBOperation("read", "time_jobs")
	tasks, err = h.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(lookAheadTime)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedTasks] Error retrieving time based tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve time based tasks",
			"code":  "TIME_TASKS_FETCH_ERROR",
			"tasks": tasks,
		})
		return
	}

	for i := range tasks {
		trackDBOp = metrics.TrackDBOperation("create", "task_data")
		taskID, err := h.taskRepository.CreateTaskDataInDB(&types.CreateTaskDataRequest{
			JobID:            tasks[i].TaskTargetData.JobID.Int,
			TaskDefinitionID: tasks[i].TaskDefinitionID,
		})
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[GetTimeBasedJobs] Error creating task data: %v", err)
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
			h.logger.Errorf("[GetTimeBasedJobs] Error adding task ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to add task ID",
				"code":  "TASK_ID_ADDITION_ERROR",
			})
			continue
		}

		tasks[i].TaskID = taskID
	}

	if len(tasks) != 0 {
		h.logger.Infof("[GetTimeBasedJobs] Successfully retrieved %d time based jobs", len(tasks))
	}
	c.JSON(http.StatusOK, tasks)
}
