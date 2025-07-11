package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/trigg3rX/triggerx-backend-imua/internal/aggregator"
	"github.com/trigg3rX/triggerx-backend-imua/internal/aggregator/types"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// Common RPC errors
var (
	ErrTaskNotFound             = errors.New("400. Task not found")
	ErrOperatorNotRegistered    = errors.New("400. Operator not registered")
	ErrInvalidTaskResponse      = errors.New("400. Invalid task response")
	ErrTaskExpired              = errors.New("400. Task has expired")
	ErrOperatorAlreadyResponded = errors.New("400. Operator has already responded to this task")
	ErrQuorumNotMet             = errors.New("400. Required quorum not met")
	ErrInvalidSignature         = errors.New("400. Invalid signature")
	ErrServerShuttingDown       = errors.New("500. Server is shutting down")
)

// RPCServer handles RPC communication with operators
type RPCServer struct {
	aggregator *aggregator.Aggregator
	logger     logging.Logger
	serverAddr string
	httpServer *http.Server
	isShutdown bool
}

// NewRPCServer creates a new RPC server instance
func NewRPCServer(agg *aggregator.Aggregator, logger logging.Logger, serverAddr string) *RPCServer {
	return &RPCServer{
		aggregator: agg,
		logger:     logger,
		serverAddr: serverAddr,
		isShutdown: false,
	}
}

// Start starts the RPC server
func (s *RPCServer) Start(ctx context.Context) error {
	if s.isShutdown {
		return ErrServerShuttingDown
	}

	// Register RPC methods
	err := rpc.Register(s)
	if err != nil {
		s.logger.Error("Failed to register RPC service", "error", err)
		return fmt.Errorf("failed to register RPC service: %w", err)
	}

	// Setup HTTP handler
	rpc.HandleHTTP()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    s.serverAddr,
		Handler: http.DefaultServeMux,
	}

	// Start server in goroutine
	go func() {
		s.logger.Infof("Starting RPC server on %s", s.serverAddr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("RPC server error", "error", err)
		}
	}()

	s.logger.Info("RPC server started successfully")
	return nil
}

// Stop gracefully shuts down the RPC server
func (s *RPCServer) Stop(ctx context.Context) error {
	s.isShutdown = true

	if s.httpServer != nil {
		s.logger.Info("Shutting down RPC server...")

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error during RPC server shutdown", "error", err)
			return err
		}
	}

	s.logger.Info("RPC server stopped")
	return nil
}

// ============================================================================
// RPC Method Definitions (these must be exported and follow RPC conventions)
// ============================================================================

// OperatorRegistrationRequest represents an operator registration request
type OperatorRegistrationRequest struct {
	OperatorID string `json:"operator_id"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	Stake      string `json:"stake"`
	Signature  string `json:"signature"`
}

// OperatorRegistrationResponse represents the response to operator registration
type OperatorRegistrationResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// RegisterOperator handles operator registration via RPC
func (s *RPCServer) RegisterOperator(req *OperatorRegistrationRequest, reply *OperatorRegistrationResponse) error {
	if s.isShutdown {
		return ErrServerShuttingDown
	}

	s.logger.Infof("Received operator registration request from %s", req.OperatorID)

	// Validate input
	if req.OperatorID == "" || req.Address == "" || req.PublicKey == "" {
		reply.Success = false
		reply.Message = "Missing required fields"
		return ErrInvalidTaskResponse
	}

	// Parse operator address
	operatorAddr := common.HexToAddress(req.Address)
	if operatorAddr == (common.Address{}) {
		reply.Success = false
		reply.Message = "Invalid operator address"
		return ErrInvalidTaskResponse
	}

	// Create operator info
	operatorInfo := &types.OperatorInfo{
		ID:        req.OperatorID,
		Address:   operatorAddr,
		PublicKey: req.PublicKey,
		Stake:     req.Stake,
	}

	// Register with aggregator
	err := s.aggregator.RegisterOperator(operatorInfo)
	if err != nil {
		s.logger.Error("Failed to register operator", "operator_id", req.OperatorID, "error", err)
		reply.Success = false
		reply.Message = err.Error()
		return err
	}

	reply.Success = true
	reply.Message = "Operator registered successfully"
	reply.Timestamp = time.Now().Unix()

	s.logger.Infof("Successfully registered operator %s", req.OperatorID)
	return nil
}

// TaskResponseSubmission represents a task response from an operator
type TaskResponseSubmission struct {
	TaskIndex    uint32 `json:"task_index"`
	OperatorID   string `json:"operator_id"`
	OperatorAddr string `json:"operator_address"`
	Response     string `json:"response"`
	Signature    string `json:"signature"`
}

// TaskResponseReply represents the response to a task response submission
type TaskResponseReply struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	TaskIndex uint32 `json:"task_index"`
	Timestamp int64  `json:"timestamp"`
}

// SubmitTaskResponse handles task response submission from operators via RPC
func (s *RPCServer) SubmitTaskResponse(submission *TaskResponseSubmission, reply *TaskResponseReply) error {
	if s.isShutdown {
		return ErrServerShuttingDown
	}

	s.logger.Infof("Received task response from operator %s for task %d", submission.OperatorID, submission.TaskIndex)

	// Validate input
	if submission.OperatorID == "" || submission.Response == "" || submission.Signature == "" {
		reply.Success = false
		reply.Message = "Missing required fields"
		reply.TaskIndex = submission.TaskIndex
		return ErrInvalidTaskResponse
	}

	// Parse operator address
	operatorAddr := common.HexToAddress(submission.OperatorAddr)
	if operatorAddr == (common.Address{}) {
		reply.Success = false
		reply.Message = "Invalid operator address"
		reply.TaskIndex = submission.TaskIndex
		return ErrInvalidTaskResponse
	}

	// Create task response
	taskResponse := &types.TaskResponse{
		TaskIndex:    types.TaskIndex(submission.TaskIndex),
		OperatorID:   submission.OperatorID,
		OperatorAddr: operatorAddr,
		Response:     submission.Response,
		Signature:    submission.Signature,
	}

	// Submit to aggregator
	err := s.aggregator.SubmitTaskResponse(taskResponse)
	if err != nil {
		s.logger.Error("Failed to submit task response",
			"operator_id", submission.OperatorID,
			"task_index", submission.TaskIndex,
			"error", err)
		reply.Success = false
		reply.Message = err.Error()
		reply.TaskIndex = submission.TaskIndex
		return err
	}

	reply.Success = true
	reply.Message = "Task response submitted successfully"
	reply.TaskIndex = submission.TaskIndex
	reply.Timestamp = time.Now().Unix()

	s.logger.Infof("Successfully processed task response from operator %s for task %d",
		submission.OperatorID, submission.TaskIndex)
	return nil
}

// TaskInfoRequest represents a request for task information
type TaskInfoRequest struct {
	TaskIndex uint32 `json:"task_index"`
}

// TaskInfoResponse represents task information response
type TaskInfoResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Task      *types.Task `json:"task,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// GetTaskInfo retrieves information about a specific task via RPC
func (s *RPCServer) GetTaskInfo(req *TaskInfoRequest, reply *TaskInfoResponse) error {
	if s.isShutdown {
		return ErrServerShuttingDown
	}

	s.logger.Infof("Received task info request for task %d", req.TaskIndex)

	// Get task from aggregator
	task, err := s.aggregator.GetTask(types.TaskIndex(req.TaskIndex))
	if err != nil {
		s.logger.Error("Failed to get task info", "task_index", req.TaskIndex, "error", err)
		reply.Success = false
		reply.Message = err.Error()
		reply.Timestamp = time.Now().Unix()
		return ErrTaskNotFound
	}

	reply.Success = true
	reply.Message = "Task information retrieved successfully"
	reply.Task = task
	reply.Timestamp = time.Now().Unix()

	return nil
}

// HealthCheckRequest represents a health check request
type HealthCheckRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
	Stats      *types.AggregatorStats `json:"stats,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
	ServerTime string                 `json:"server_time"`
}

// HealthCheck provides health status via RPC
func (s *RPCServer) HealthCheck(req *HealthCheckRequest, reply *HealthCheckResponse) error {
	if s.isShutdown {
		reply.Success = false
		reply.Message = "Server is shutting down"
		reply.Timestamp = time.Now().Unix()
		reply.ServerTime = time.Now().Format(time.RFC3339)
		return ErrServerShuttingDown
	}

	s.logger.Debug("Received health check request")

	// Get aggregator stats
	stats := s.aggregator.GetStats()

	reply.Success = true
	reply.Message = "Aggregator RPC server is healthy"
	reply.Stats = stats
	reply.Timestamp = time.Now().Unix()
	reply.ServerTime = time.Now().Format(time.RFC3339)

	return nil
}

// TaskListRequest represents a request for task listing
type TaskListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// GetTasks retrieves a list of tasks via RPC
func (s *RPCServer) GetTasks(req *TaskListRequest, reply *types.TaskListResponse) error {
	if s.isShutdown {
		return ErrServerShuttingDown
	}

	s.logger.Infof("Received task list request (page: %d, size: %d)", req.Page, req.PageSize)

	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get tasks from aggregator
	taskList, err := s.aggregator.GetTasks(page, pageSize)
	if err != nil {
		s.logger.Error("Failed to get task list", "error", err)
		return fmt.Errorf("failed to retrieve tasks: %w", err)
	}

	*reply = *taskList
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// GetAddress returns the server address
func (s *RPCServer) GetAddress() string {
	return s.serverAddr
}

// IsRunning returns true if the server is running
func (s *RPCServer) IsRunning() bool {
	return s.httpServer != nil && !s.isShutdown
}

// GetStats returns server statistics
func (s *RPCServer) GetServerStats() map[string]interface{} {
	return map[string]interface{}{
		"server_address": s.serverAddr,
		"is_running":     s.IsRunning(),
		"is_shutdown":    s.isShutdown,
		"start_time":     time.Now().Format(time.RFC3339),
	}
}
