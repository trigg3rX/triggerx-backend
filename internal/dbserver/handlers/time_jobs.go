package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// timeNow is a variable that holds the time.Now function
// It can be overridden in tests
var timeNow = time.Now

func (h *Handler) GetTimeBasedJobs(c *gin.Context) {
	pollIntervalStr := c.Query("pollInterval")
	pollInterval, err := strconv.ParseInt(pollIntervalStr, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedJobs] Error parsing poll interval: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll interval"})
		return
	}

	nextExecutionTimestamp := timeNow().Add(time.Duration(pollInterval) * time.Second)

	var jobs []commonTypes.ScheduleTimeJobData

	jobs, err = h.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(nextExecutionTimestamp)
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
