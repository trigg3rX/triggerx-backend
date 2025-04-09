package manager

import (
	// "bytes"
	// "encoding/json"
	// "fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	// "github.com/trigg3rX/triggerx-backend/internal/manager/config"
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
	case 1, 2: // Time-based jobs
		err = jobScheduler.StartTimeBasedJob(jobData)
	case 3, 4: // Event-based jobs
		err = jobScheduler.StartEventBasedJob(jobData)
	case 5, 6: // Condition-based jobs
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

	// TODO: Implement job update logic using scheduler
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

	// TODO: Implement job pause logic using scheduler
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

	// TODO: Implement job resume logic using scheduler
	logger.Infof("Job resume requested for ID: %d", resumeJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job resume request received"})
}

// HandleJobStateUpdate handles requests to update a job's state in the scheduler
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

	// Retrieve the worker for this job
	worker := jobScheduler.GetWorker(updateData.JobID)
	if worker == nil {
		logger.Warnf("No active worker found for job ID: %d", updateData.JobID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found or not active"})
		return
	}

	// Update the job's last executed timestamp in the worker
	if err := jobScheduler.UpdateJobLastExecutedTime(updateData.JobID, updateData.Timestamp); err != nil {
		logger.Errorf("Failed to update job %d last executed time: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job state"})
		return
	}

	// Update the job's state in the cache
	if err := jobScheduler.UpdateJobStateCache(updateData.JobID, "last_executed", updateData.Timestamp); err != nil {
		logger.Warnf("Failed to update job %d state cache: %v", updateData.JobID, err)
		// Continue even if cache update fails
	}

	logger.Infof("Successfully updated state for job ID: %d", updateData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job state updated successfully"})
}

// func HandleKeeperConnectEvent(c *gin.Context) {
// 	var keeperData types.UpdateKeeperConnectionData
// 	if err := c.BindJSON(&keeperData); err != nil {
// 		logger.Error("Failed to parse keeper data", "error", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
// 		return
// 	}

// 	url := fmt.Sprintf("%s/api/keepers/connection", config.DatabaseIPAddress)

// 	jsonData, err := json.Marshal(keeperData)
// 	if err != nil {
// 		logger.Error("Failed to marshal keeper data", "error", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal keeper data"})
// 		return
// 	}

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		logger.Error("Failed to create HTTP request", "error", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create HTTP request"})
// 		return
// 	}

// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		logger.Error("Failed to send request", "error", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to send request"})
// 		return
// 	}
// 	defer resp.Body.Close()

// 	var response types.UpdateKeeperConnectionDataResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		logger.Error("Failed to decode response", "error", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode response"})
// 		return
// 	}

// 	logger.Infof("Keeper connected: %s", keeperData.KeeperAddress)
// 	c.JSON(http.StatusOK, response)
// }
