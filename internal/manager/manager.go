package manager

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)
	jobScheduler *scheduler.JobScheduler
)

func JobSchedulerInit() {
	var err error
	jobScheduler, err = scheduler.NewJobScheduler(logger)
	if err != nil {
		logger.Fatalf("Failed to initialize job scheduler: %v", err)
	}
}

func HandleCreateJobEvent(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Invalid method"})
		return
	}

	var jobData types.HandleCreateJobData
	if err := c.BindJSON(&jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	var err error
	switch jobData.TaskDefinitionID {
	case 1, 2:
		err = jobScheduler.StartTimeBasedJob(jobData)
	case 3, 4:
		err = jobScheduler.StartEventBasedJob(jobData)
	case 5, 6:
		err = jobScheduler.StartConditionBasedJob(jobData)
	default:
		logger.Warnf("Unknown task definition ID: %d for job: %d",
			jobData.TaskDefinitionID, jobData.JobID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
		return
	}

	if err != nil {
		logger.Errorf("Failed to schedule job %d: %v", jobData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule job"})
		return
	}

	logger.Infof("Successfully scheduled job with ID: %d", jobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job scheduled successfully"})
}

func HandleUpdateJobEvent(c *gin.Context) {
	var updateJobData types.HandleUpdateJobData
	if err := c.BindJSON(&updateJobData); err != nil {
		logger.Error("Failed to parse update job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	logger.Infof("Job update requested for ID: %d", updateJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job update request received"})
}

func HandlePauseJobEvent(c *gin.Context) {
	var pauseJobData types.HandlePauseJobData
	if err := c.BindJSON(&pauseJobData); err != nil {
		logger.Error("Failed to parse pause job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	logger.Infof("Job pause requested for ID: %d", pauseJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job pause request received"})
}

func HandleResumeJobEvent(c *gin.Context) {
	var resumeJobData types.HandleResumeJobData
	if err := c.BindJSON(&resumeJobData); err != nil {
		logger.Error("Failed to parse resume job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	logger.Infof("Job resume requested for ID: %d", resumeJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job resume request received"})
}

func HandleJobStateUpdate(c *gin.Context) {
	var updateData struct {
		JobID     int64     `json:"job_id"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := c.BindJSON(&updateData); err != nil {
		logger.Error("Failed to parse job state update data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	logger.Infof("Updating state for job ID: %d with timestamp: %v", updateData.JobID, updateData.Timestamp)

	worker := jobScheduler.GetWorker(updateData.JobID)
	if worker == nil {
		logger.Warnf("No active worker found for job ID: %d", updateData.JobID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found or not active"})
		return
	}

	if err := jobScheduler.UpdateJobLastExecutedTime(updateData.JobID, updateData.Timestamp); err != nil {
		logger.Errorf("Failed to update job %d last executed time: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job state"})
		return
	}

	if err := jobScheduler.UpdateJobStateCache(updateData.JobID, "last_executed", updateData.Timestamp); err != nil {
		logger.Warnf("Failed to update job %d state cache: %v", updateData.JobID, err)
	}

	logger.Infof("Successfully updated state for job ID: %d", updateData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job state updated successfully"})
}
