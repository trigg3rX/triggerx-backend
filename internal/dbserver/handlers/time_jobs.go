package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTimeBasedJobs(c *gin.Context) {
	var pollInterval int64
	if err := c.ShouldBindQuery(&pollInterval); err != nil {
		h.logger.Errorf("[GetTimeBasedJobs] Error getting poll interval: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nextExecutionTimestamp := time.Now().Add(time.Duration(pollInterval) * time.Second)

	var jobs []commonTypes.ScheduleTimeJobData

	jobs, err := h.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(nextExecutionTimestamp)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedJobs] Error retrieving time based jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if jobs == nil {
		jobs = []commonTypes.ScheduleTimeJobData{}
	}

	for _, job := range jobs {
		if job.DynamicArgumentsScriptUrl != "" {
			job.TaskDefinitionID = 2
		} else {
			job.TaskDefinitionID = 1
		}
	}

	h.logger.Infof("[GetTimeBasedJobs] Successfully retrieved %d time based jobs", len(jobs))
	c.JSON(http.StatusOK, jobs)
}
