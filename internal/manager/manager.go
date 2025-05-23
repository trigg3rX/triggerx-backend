package manager

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger       logging.Logger
	jobScheduler *scheduler.JobScheduler
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, jobScheduler *scheduler.JobScheduler) *Handler {
	return &Handler{
		logger:       logger.With("component", "health_handler"),
		jobScheduler: jobScheduler,
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	middlewareLogger := logger.With("component", "http_middleware")
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		if status >= 500 {
			middlewareLogger.Error("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		} else if status >= 400 {
			middlewareLogger.Warn("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		} else {
			middlewareLogger.Info("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		}
	}
}

// RegisterRoutes registers all HTTP routes for the health service
func RegisterRoutes(router *gin.Engine, jobScheduler *scheduler.JobScheduler) {
	logger := logging.GetServiceLogger()
	handler := NewHandler(logger, jobScheduler)

	router.GET("/", handler.handleRoot)
	router.POST("/jobs/schedule", handler.ScheduleJob)
	router.POST("/jobs/:id/pause", handler.PauseJob)
	router.POST("/jobs/:id/resume", handler.ResumeJob)
	router.POST("/jobs/:id/cancel", handler.CancelJob)
	router.POST("/p2p/message", handler.HandleP2PMessage)
	router.POST("/task/validate", handler.ValidateTask)
}

func (h *Handler) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Manager Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) ScheduleJob(c *gin.Context) {
	var job types.HandleCreateJobData
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	switch {
	case job.TaskDefinitionID == 1 || job.TaskDefinitionID == 2:
		err = h.jobScheduler.StartTimeBasedJob(job)
	case job.TaskDefinitionID == 3 || job.TaskDefinitionID == 4:
		err = h.jobScheduler.StartEventBasedJob(job)
	case job.TaskDefinitionID == 5 || job.TaskDefinitionID == 6:
		err = h.jobScheduler.StartConditionBasedJob(job)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job scheduled successfully"})
}

func (h *Handler) PauseJob(c *gin.Context) {
	// jobID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
	// 	return
	// }

	// if err := h.jobScheduler.PauseJob(jobID); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	c.JSON(http.StatusOK, gin.H{"message": "Job paused successfully"})
}

func (h *Handler) ResumeJob(c *gin.Context) {
	// jobID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
	// 	return
	// }

	// if err := h.jobScheduler.ResumeJob(jobID); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	c.JSON(http.StatusOK, gin.H{"message": "Job resumed successfully"})
}

func (h *Handler) CancelJob(c *gin.Context) {
	// jobID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
	// 	return
	// }

	// if err := h.jobScheduler.CancelJob(jobID); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	c.JSON(http.StatusOK, gin.H{"message": "Job cancelled successfully"})
}

func (h *Handler) HandleP2PMessage(c *gin.Context) {
	h.logger.Info("Executing task")

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Invalid method",
		})
		return
	}

	var requestBody struct {
		Data string `json:"data"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// Decode hex data
	hexData := requestBody.Data
	if len(hexData) > 2 && hexData[:2] == "0x" {
		hexData = hexData[2:]
	}

	decodedData, err := hex.DecodeString(hexData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hex data"})
		return
	}

	decodedDataString := string(decodedData)

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(decodedDataString), &requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse JSON data",
		})
		return
	}

	jobDataRaw := requestData["jobData"]
	triggerDataRaw := requestData["triggerData"]
	performerDataRaw := requestData["performerData"]

	// Convert to proper types
	var jobData types.HandleCreateJobData
	jobDataBytes, err := json.Marshal(jobDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job data format"})
		return
	}
	if err := json.Unmarshal(jobDataBytes, &jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse job data"})
		return
	}
	h.logger.Infof("jobData: %v\n", jobData)

	var triggerData types.TriggerData
	triggerDataBytes, err := json.Marshal(triggerDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger data format"})
		return
	}
	if err := json.Unmarshal(triggerDataBytes, &triggerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse trigger data"})
		return
	}
	h.logger.Infof("triggerData: %v\n", triggerData)

	var performerData types.GetPerformerData
	performerDataBytes, err := json.Marshal(performerDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid performer data format"})
		return
	}
	if err := json.Unmarshal(performerDataBytes, &performerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse performer data"})
		return
	}
	h.logger.Infof("performerData: %v\n", performerData)

	// // Execute task
	// actionData, err := h.executor.Execute(jobData)
	// if err != nil {
	// 	h.logger.Error("Failed to execute task", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Task execution failed"})
	// 	return
	// }

	// // Set task ID from trigger data
	// actionData.TaskID = triggerData.TaskID

	// // Convert result to bytes
	// resultBytes, err := json.Marshal(actionData)
	// if err != nil {
	// 	h.logger.Error("Failed to marshal result", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process result"})
	// 	return
	// }

	resultBytes := []byte("test")

	c.Data(http.StatusOK, "application/octet-stream", resultBytes)
}

func (h *Handler) ValidateTask(c *gin.Context) {
	// var request types.TaskValidationRequest
	// if err := c.BindJSON(&request); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	// 	return
	// }

	// // Validate task
	// isValid, err := h.validator.ValidateTask(&request)
	// if err != nil {
	// 	h.logger.Error("Task validation failed", "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation failed"})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"isValid": true,
	})
}
