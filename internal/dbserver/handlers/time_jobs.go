package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTimeBasedTasks(c *gin.Context) {
	pollLookAhead := config.GetPollingLookAhead()
	lookAheadTime := time.Now().Add(time.Duration(pollLookAhead) * time.Second)

	var tasks []commonTypes.ScheduleTimeTaskData
	var err error

	tasks, err = h.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(lookAheadTime)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedTasks] Error retrieving time based tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "tasks": tasks})
		return
	}

	for _, task := range tasks {
		taskID, err := h.taskRepository.CreateTaskDataInDB(&types.CreateTaskDataRequest{
			JobID:            task.JobID,
			TaskDefinitionID: task.TaskDefinitionID,
		})
		if err != nil {
			h.logger.Errorf("[GetTimeBasedJobs] Error creating task data: %v", err)
			continue
		}
		task.TaskID = taskID
	}

	h.logger.Infof("[GetTimeBasedJobs] Successfully retrieved %d time based jobs", len(tasks))
	c.JSON(http.StatusOK, tasks)
}
