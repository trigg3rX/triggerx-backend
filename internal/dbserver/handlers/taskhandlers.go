package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/resources"
)

func (h *Handler) CreateTaskData(c *gin.Context) {
	var req types.CreateTaskDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validation: required fields must not be zero
	if req.JobID == 0 || req.TaskDefinitionID == 0 || req.TaskPerformerID == 0 {
		h.logger.Error("Missing required fields in request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	taskID, err := h.taskRepository.CreateTaskDataInDB(&req)
	if err != nil {
		h.logger.Error("Error creating task data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}

func (h *Handler) UpdateTaskExecutionData(c *gin.Context) {
	taskID := c.Param("id")
	h.logger.Infof("[UpdateTaskExecutionData] Updating task execution data for task with ID: %s", taskID)

	var taskData types.UpdateTaskExecutionDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Errorf("[UpdateTaskExecutionData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.ExecutionTimestamp.IsZero() || taskData.ExecutionTxHash == "" {
		h.logger.Errorf("[UpdateTaskExecutionData] Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	if err := h.taskRepository.UpdateTaskExecutionDataInDB(&taskData); err != nil {
		h.logger.Errorf("[UpdateTaskExecutionData] Error updating task execution data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateTaskExecutionData] Successfully updated task execution data for task with ID: %s", taskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task execution data updated successfully"})
}

func (h *Handler) UpdateTaskAttestationData(c *gin.Context) {
	taskID := c.Param("id")
	h.logger.Infof("[UpdateTaskAttestationData] Updating task attestation data for task with ID: %s", taskID)

	var taskData types.UpdateTaskAttestationDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Errorf("[UpdateTaskAttestationData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if taskData.TaskID == 0 || taskData.TaskNumber == 0 || len(taskData.TaskAttesterIDs) == 0 || len(taskData.TpSignature) == 0 || len(taskData.TaSignature) == 0 || taskData.TaskSubmissionTxHash == "" {
		h.logger.Errorf("[UpdateTaskAttestationData] Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	if err := h.taskRepository.UpdateTaskAttestationDataInDB(&taskData); err != nil {
		h.logger.Errorf("[UpdateTaskAttestationData] Error updating task attestation data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateTaskAttestationData] Successfully updated task attestation data for task with ID: %s", taskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task attestation data updated successfully"})
}

func (h *Handler) GetTaskDataByID(c *gin.Context) {
	taskID := c.Param("id")
	h.logger.Infof("[GetTaskDataByID] Fetching task with ID: %s", taskID)

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Error parsing task ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}
	taskData, err := h.taskRepository.GetTaskDataByID(taskIDInt)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Error retrieving task with ID %s: %v", taskID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Successfully retrieved task with ID: %s", taskID)
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	jobID := c.Param("id")
	h.logger.Infof("[GetTasksByJobID] Fetching tasks for job with ID: %s", jobID)

	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error parsing job ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	taskData, err := h.taskRepository.GetTasksByJobID(jobIDInt)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving tasks for job with ID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Successfully retrieved tasks for job with ID: %s", jobID)
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) CalculateTaskFees(ipfsURLs string) (float64, error) {
	if ipfsURLs == "" {
		return 0, fmt.Errorf("missing IPFS URLs")
	}

	urlList := strings.Split(ipfsURLs, ",")
	totalFee := 0.0
	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer func() {
		if err := cli.Close(); err != nil {
			h.logger.Errorf("Error closing Docker client: %v", err)
		}
	}()

	for _, ipfsURL := range urlList {
		ipfsURL = strings.TrimSpace(ipfsURL)
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			codePath, err := resources.DownloadIPFSFile(url)
			if err != nil {
				h.logger.Errorf("Error downloading IPFS file: %v", err)
				return
			}
			defer func() {
				if err := os.RemoveAll(filepath.Dir(codePath)); err != nil {
					h.logger.Errorf("Error removing temporary directory: %v", err)
				}
			}()

			containerID, err := resources.CreateDockerContainer(ctx, cli, codePath)
			if err != nil {
				h.logger.Errorf("Error creating container: %v", err)
				return
			}
			if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
				h.logger.Errorf("Error removing container: %v", err)
			}

			stats, err := resources.MonitorResources(ctx, cli, containerID)
			if err != nil {
				h.logger.Errorf("Error monitoring resources: %v", err)
				return
			}

			mu.Lock()
			totalFee += stats.TotalFee
			mu.Unlock()
		}(ipfsURL)
	}

	wg.Wait()
	return totalFee, nil
}

func (h *Handler) GetTaskFees(c *gin.Context) {
	ipfsURLs := c.Query("ipfs_url")

	totalFee, err := h.CalculateTaskFees(ipfsURLs)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error calculating fees: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_fee": totalFee,
	})
}

func (h *Handler) UpdateTaskFee(c *gin.Context) {
	taskID := c.Param("id")
	h.logger.Infof("[UpdateTaskFee] Updating task fee for task with ID: %s", taskID)

	var taskFee struct {
		Fee float64 `json:"fee"`
	}
	if err := c.ShouldBindJSON(&taskFee); err != nil {
		h.logger.Errorf("[UpdateTaskFee] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Errorf("[UpdateTaskFee] Error parsing task ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}
	if err := h.taskRepository.UpdateTaskFee(taskIDInt, taskFee.Fee); err != nil {
		h.logger.Errorf("[UpdateTaskFee] Error updating task fee: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateTaskFee] Successfully updated task fee for task with ID: %s", taskID)
	c.JSON(http.StatusOK, taskFee)
}
