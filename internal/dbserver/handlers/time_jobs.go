package handlers

import (
	// "math"
	// "math/big"
	"net/http"
	// "strings"
	"time"

	"github.com/gin-gonic/gin"
	// "github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTimeBasedJobs(c *gin.Context) {
	// var pollInterval int64
	// if err := c.ShouldBindQuery(&pollInterval); err != nil {
	// 	h.logger.Errorf("[GetTimeBasedJobs] Error decoding request body: %v", err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	pollInterval := int64(10)

	var jobs []types.TimeJobData

	iter := h.db.Session().Query(`
		SELECT job_id, time_frame, recurring, schedule_type, time_interval, 
		       cron_expression, specific_schedule, timezone, next_execution_timestamp,
		       target_chain_id, target_contract_address, target_function,
		       abi, arg_type, arguments, dynamic_arguments_script_ipfs_url 
		FROM triggerx.time_job_data 
		WHERE next_execution_timestamp <= ? ALLOW FILTERING`,
		time.Now().Add(time.Duration(pollInterval)*time.Second)).Iter()

	var job types.TimeJobData
	for iter.Scan(
		&job.JobID, &job.TimeFrame, &job.Recurring, &job.ScheduleType,
		&job.TimeInterval, &job.CronExpression, &job.SpecificSchedule,
		&job.Timezone, &job.NextExecutionTimestamp,
		&job.TargetChainID, &job.TargetContractAddress, &job.TargetFunction,
		&job.ABI, &job.ArgType, &job.Arguments, &job.DynamicArgumentsScriptIPFSUrl) {
		jobs = append(jobs, job)
	}

	if err := iter.Close(); err != nil {
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
