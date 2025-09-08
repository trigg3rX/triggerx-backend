package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/aggregator"
	"github.com/trigg3rX/triggerx-backend/internal/aggregator/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// APIHandlers contains the aggregator instance and provides HTTP handlers
type APIHandlers struct {
	aggregator *aggregator.Aggregator
	logger     logging.Logger
}

// NewAPIHandlers creates a new API handlers instance
func NewAPIHandlers(agg *aggregator.Aggregator, logger logging.Logger) *APIHandlers {
	return &APIHandlers{
		aggregator: agg,
		logger:     logger,
	}
}

// RegisterRoutes registers all HTTP routes for the aggregator API
func RegisterRoutes(router *gin.Engine, logger logging.Logger) {
	// This will be updated once we have the aggregator instance
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Aggregator service is running",
			"timestamp":   time.Now().Unix(),
			"server_time": time.Now().Format(time.RFC3339),
		})
	})
}

// RegisterRoutesWithAggregator registers all HTTP routes with an aggregator instance
func RegisterRoutesWithAggregator(router *gin.Engine, agg *aggregator.Aggregator, logger logging.Logger) {
	handlers := NewAPIHandlers(agg, logger)

	// API version group
	v1 := router.Group("/api/v1")
	{
		// Health and status endpoints
		v1.GET("/health", handlers.HealthCheck)
		v1.GET("/stats", handlers.GetStats)

		// Task management endpoints
		tasks := v1.Group("/tasks")
		{
			tasks.POST("/", handlers.CreateTask)
			tasks.GET("/", handlers.GetTasks)
			tasks.GET("/:taskIndex", handlers.GetTask)
		}

		// Operator management endpoints
		operators := v1.Group("/operators")
		{
			operators.POST("/register", handlers.RegisterOperator)
			operators.POST("/response", handlers.SubmitTaskResponse)
		}
	}

	// Legacy endpoints for backward compatibility
	router.GET("/health", handlers.HealthCheck)
}

// ============================================================================
// Health and Status Handlers
// ============================================================================

// HealthCheck returns the health status of the aggregator
func (h *APIHandlers) HealthCheck(c *gin.Context) {
	stats := h.aggregator.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"message":     "Aggregator service is running",
		"timestamp":   time.Now().Unix(),
		"server_time": time.Now().Format(time.RFC3339),
		"stats":       stats,
		"status":      "healthy",
	})
}

// GetStats returns aggregator statistics
func (h *APIHandlers) GetStats(c *gin.Context) {
	stats := h.aggregator.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      stats,
		"timestamp": time.Now().Unix(),
	})
}

// ============================================================================
// Task Management Handlers
// ============================================================================

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Data           string `json:"data" binding:"required"`
	RequiredQuorum uint8  `json:"required_quorum"`
	Timeout        string `json:"timeout,omitempty"` // Duration string like "5m", "1h"
	SubmitterAddr  string `json:"submitter_address" binding:"required"`
}

// CreateTask handles task creation requests
func (h *APIHandlers) CreateTask(c *gin.Context) {
	var req CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid task creation request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Parse submitter address
	submitterAddr := common.HexToAddress(req.SubmitterAddr)
	if submitterAddr == (common.Address{}) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid submitter address",
		})
		return
	}

	// Parse timeout if provided
	var timeout time.Duration
	if req.Timeout != "" {
		var err error
		timeout, err = time.ParseDuration(req.Timeout)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid timeout format",
				"details": "Use format like '5m', '1h', '30s'",
			})
			return
		}
	}

	// Set default quorum if not provided
	requiredQuorum := req.RequiredQuorum
	if requiredQuorum == 0 {
		requiredQuorum = 1
	}

	// Create task request
	taskReq := &types.NewTaskRequest{
		Data:           req.Data,
		RequiredQuorum: requiredQuorum,
		Timeout:        timeout,
		SubmitterAddr:  submitterAddr,
	}

	// Create task
	task, err := h.aggregator.CreateTask(taskReq)
	if err != nil {
		h.logger.Error("Failed to create task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create task",
			"details": err.Error(),
		})
		return
	}

	h.logger.Infof("Task created successfully: %d", task.Index)
	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"message":   "Task created successfully",
		"data":      task,
		"timestamp": time.Now().Unix(),
	})
}

// GetTasks handles task listing requests
func (h *APIHandlers) GetTasks(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get tasks
	taskList, err := h.aggregator.GetTasks(page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get tasks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve tasks",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      taskList,
		"timestamp": time.Now().Unix(),
	})
}

// GetTask handles individual task retrieval requests
func (h *APIHandlers) GetTask(c *gin.Context) {
	taskIndexStr := c.Param("taskIndex")

	taskIndex, err := strconv.ParseUint(taskIndexStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid task index",
			"details": "Task index must be a valid number",
		})
		return
	}

	// Get task
	task, err := h.aggregator.GetTask(types.TaskIndex(taskIndex))
	if err != nil {
		h.logger.Error("Failed to get task", "task_index", taskIndex, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Task not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      task,
		"timestamp": time.Now().Unix(),
	})
}

// ============================================================================
// Operator Management Handlers
// ============================================================================

// RegisterOperatorRequest represents the request body for operator registration
type RegisterOperatorRequest struct {
	OperatorID string `json:"operator_id" binding:"required"`
	Address    string `json:"address" binding:"required"`
	PublicKey  string `json:"public_key" binding:"required"`
	Stake      string `json:"stake"`
	Signature  string `json:"signature"`
}

// RegisterOperator handles operator registration requests
func (h *APIHandlers) RegisterOperator(c *gin.Context) {
	var req RegisterOperatorRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid operator registration request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Parse operator address
	operatorAddr := common.HexToAddress(req.Address)
	if operatorAddr == (common.Address{}) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid operator address",
		})
		return
	}

	// Create operator info
	operatorInfo := &types.OperatorInfo{
		ID:        req.OperatorID,
		Address:   operatorAddr,
		PublicKey: req.PublicKey,
		Stake:     req.Stake,
	}

	// Register operator
	err := h.aggregator.RegisterOperator(operatorInfo)
	if err != nil {
		h.logger.Error("Failed to register operator", "operator_id", req.OperatorID, "error", err)
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Failed to register operator",
			"details": err.Error(),
		})
		return
	}

	h.logger.Infof("Operator registered successfully: %s", req.OperatorID)
	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"message":     "Operator registered successfully",
		"operator_id": req.OperatorID,
		"timestamp":   time.Now().Unix(),
	})
}

// SubmitTaskResponseRequest represents the request body for task response submission
type SubmitTaskResponseRequest struct {
	TaskIndex    uint32 `json:"task_index" binding:"required"`
	OperatorID   string `json:"operator_id" binding:"required"`
	OperatorAddr string `json:"operator_address" binding:"required"`
	Response     string `json:"response" binding:"required"`
	Signature    string `json:"signature" binding:"required"`
}

// SubmitTaskResponse handles task response submissions from operators
func (h *APIHandlers) SubmitTaskResponse(c *gin.Context) {
	var req SubmitTaskResponseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid task response submission", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Parse operator address
	operatorAddr := common.HexToAddress(req.OperatorAddr)
	if operatorAddr == (common.Address{}) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid operator address",
		})
		return
	}

	// Create task response
	taskResponse := &types.TaskResponse{
		TaskIndex:    types.TaskIndex(req.TaskIndex),
		OperatorID:   req.OperatorID,
		OperatorAddr: operatorAddr,
		Response:     req.Response,
		Signature:    req.Signature,
	}

	// Submit response
	err := h.aggregator.SubmitTaskResponse(taskResponse)
	if err != nil {
		h.logger.Error("Failed to submit task response",
			"operator_id", req.OperatorID,
			"task_index", req.TaskIndex,
			"error", err)

		status := http.StatusInternalServerError
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		} else if err.Error() == "task has expired" || err.Error() == "operator has already responded" {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{
			"success": false,
			"error":   "Failed to submit task response",
			"details": err.Error(),
		})
		return
	}

	h.logger.Infof("Task response submitted successfully: operator %s, task %d", req.OperatorID, req.TaskIndex)
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Task response submitted successfully",
		"task_index":  req.TaskIndex,
		"operator_id": req.OperatorID,
		"timestamp":   time.Now().Unix(),
	})
}
