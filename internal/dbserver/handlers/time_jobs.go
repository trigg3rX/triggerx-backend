package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
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

	var jobs []types.TimeJobData

	jobs, err = h.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(nextExecutionTimestamp)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedJobs] Error retrieving time based jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if jobs == nil {
		jobs = []types.TimeJobData{}
	}

	h.logger.Infof("[GetTimeBasedJobs] Successfully retrieved %d time based jobs", len(jobs))
	c.JSON(http.StatusOK, jobs)
}
