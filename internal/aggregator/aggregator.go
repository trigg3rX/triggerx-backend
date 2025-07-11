package aggregator

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/internal/aggregator/types"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// Aggregator manages tasks, operators, and response aggregation
type Aggregator struct {
	logger logging.Logger
	mu     sync.RWMutex

	// Task management
	tasks         map[types.TaskIndex]*types.Task
	taskResponses map[types.TaskIndex][]*types.TaskResponse
	nextTaskIndex types.TaskIndex

	// Operator management
	operators map[string]*types.OperatorInfo

	// Configuration
	config *types.AggregatorConfig

	// Channels for communication
	taskChannel     chan *types.Task
	responseChannel chan *types.TaskResponse
	shutdownChannel chan struct{}

	// Metrics
	stats *types.AggregatorStats
}

// NewAggregator creates a new aggregator instance
func NewAggregator(logger logging.Logger, config *types.AggregatorConfig) *Aggregator {
	if config == nil {
		config = &types.AggregatorConfig{
			MaxConcurrentTasks: 100,
			DefaultTimeout:     5 * time.Minute,
			MinOperators:       1,
			MaxOperators:       100,
		}
	}

	return &Aggregator{
		logger:          logger,
		tasks:           make(map[types.TaskIndex]*types.Task),
		taskResponses:   make(map[types.TaskIndex][]*types.TaskResponse),
		operators:       make(map[string]*types.OperatorInfo),
		nextTaskIndex:   1,
		config:          config,
		taskChannel:     make(chan *types.Task, config.MaxConcurrentTasks),
		responseChannel: make(chan *types.TaskResponse, config.MaxConcurrentTasks*10),
		shutdownChannel: make(chan struct{}),
		stats: &types.AggregatorStats{
			TotalTasks:      0,
			CompletedTasks:  0,
			FailedTasks:     0,
			ActiveOperators: 0,
			LastTaskCreated: time.Time{},
		},
	}
}

// Start begins the aggregator service
func (a *Aggregator) Start(ctx context.Context) error {
	a.logger.Info("Starting aggregator service...")

	// Start background workers
	go a.taskProcessor(ctx)
	go a.responseProcessor(ctx)
	go a.taskExpiryChecker(ctx)
	go a.statsUpdater(ctx)

	a.logger.Info("Aggregator service started successfully")
	return nil
}

// Stop gracefully shuts down the aggregator
func (a *Aggregator) Stop() error {
	a.logger.Info("Stopping aggregator service...")
	close(a.shutdownChannel)
	return nil
}

// CreateTask creates a new task and adds it to the processing queue
func (a *Aggregator) CreateTask(req *types.NewTaskRequest) (*types.Task, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if we've reached the maximum concurrent tasks
	if len(a.tasks) >= a.config.MaxConcurrentTasks {
		return nil, fmt.Errorf("maximum concurrent tasks limit reached (%d)", a.config.MaxConcurrentTasks)
	}

	// Set default timeout if not provided
	timeout := req.Timeout
	if timeout == 0 {
		timeout = a.config.DefaultTimeout
	}

	// Create new task
	task := &types.Task{
		Index:          a.nextTaskIndex,
		Data:           req.Data,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(timeout),
		Status:         types.TaskStatusPending,
		RequiredQuorum: req.RequiredQuorum,
		SubmitterAddr:  req.SubmitterAddr,
		ResponseCount:  0,
		BlockNumber:    types.BlockNumber(time.Now().Unix()), // Simplified block number
	}

	// Store task
	a.tasks[task.Index] = task
	a.taskResponses[task.Index] = make([]*types.TaskResponse, 0)
	a.nextTaskIndex++

	// Update stats
	a.stats.TotalTasks++
	a.stats.LastTaskCreated = task.CreatedAt

	// Queue task for processing
	select {
	case a.taskChannel <- task:
		a.logger.Infof("Created task %d: %s", task.Index, task.Data)
	default:
		return nil, fmt.Errorf("task queue is full")
	}

	return task, nil
}

// SubmitTaskResponse processes a response from an operator
func (a *Aggregator) SubmitTaskResponse(response *types.TaskResponse) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Validate task exists
	task, exists := a.tasks[response.TaskIndex]
	if !exists {
		return fmt.Errorf("task %d not found", response.TaskIndex)
	}

	// Check if task is still active
	if task.Status != types.TaskStatusPending && task.Status != types.TaskStatusProcessing {
		return fmt.Errorf("task %d is no longer accepting responses (status: %s)", response.TaskIndex, task.Status)
	}

	// Check if task has expired
	if time.Now().After(task.ExpiresAt) {
		task.Status = types.TaskStatusExpired
		return fmt.Errorf("task %d has expired", response.TaskIndex)
	}

	// Validate operator
	operator, exists := a.operators[response.OperatorID]
	if !exists {
		return fmt.Errorf("operator %s not registered", response.OperatorID)
	}

	if !operator.IsActive {
		return fmt.Errorf("operator %s is not active", response.OperatorID)
	}

	// Check for duplicate responses from the same operator
	for _, existingResponse := range a.taskResponses[response.TaskIndex] {
		if existingResponse.OperatorID == response.OperatorID {
			return fmt.Errorf("operator %s has already submitted a response for task %d", response.OperatorID, response.TaskIndex)
		}
	}

	// Add response timestamp
	response.SubmittedAt = time.Now()
	response.IsValid = a.validateResponse(task, response)

	// Store response
	a.taskResponses[response.TaskIndex] = append(a.taskResponses[response.TaskIndex], response)
	task.ResponseCount++

	// Update operator activity
	operator.LastActivity = time.Now()

	// Queue response for processing
	select {
	case a.responseChannel <- response:
		a.logger.Infof("Received response from operator %s for task %d", response.OperatorID, response.TaskIndex)
	default:
		return fmt.Errorf("response queue is full")
	}

	return nil
}

// RegisterOperator registers a new operator
func (a *Aggregator) RegisterOperator(operator *types.OperatorInfo) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if operator already exists
	if _, exists := a.operators[operator.ID]; exists {
		return fmt.Errorf("operator %s already registered", operator.ID)
	}

	// Check maximum operators limit
	if len(a.operators) >= a.config.MaxOperators {
		return fmt.Errorf("maximum operators limit reached (%d)", a.config.MaxOperators)
	}

	// Set registration time and activate
	operator.RegisteredAt = time.Now()
	operator.LastActivity = time.Now()
	operator.IsActive = true

	// Store operator
	a.operators[operator.ID] = operator
	a.stats.ActiveOperators++

	a.logger.Infof("Registered operator %s at address %s", operator.ID, operator.Address.Hex())
	return nil
}

// GetTask retrieves a task by index
func (a *Aggregator) GetTask(taskIndex types.TaskIndex) (*types.Task, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	task, exists := a.tasks[taskIndex]
	if !exists {
		return nil, fmt.Errorf("task %d not found", taskIndex)
	}

	return task, nil
}

// GetTasks retrieves tasks with pagination
func (a *Aggregator) GetTasks(page, pageSize int) (*types.TaskListResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	totalCount := len(a.tasks)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	tasks := make([]*types.Task, 0, len(a.tasks))
	for _, task := range a.tasks {
		tasks = append(tasks, task)
	}

	// Sort by creation time (newest first)
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].CreatedAt.Before(tasks[j].CreatedAt) {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}

	// Apply pagination
	var paginatedTasks []types.Task
	if start < len(tasks) {
		if end > len(tasks) {
			end = len(tasks)
		}
		for i := start; i < end; i++ {
			paginatedTasks = append(paginatedTasks, *tasks[i])
		}
	}

	return &types.TaskListResponse{
		Tasks:      paginatedTasks,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetStats returns current aggregator statistics
func (a *Aggregator) GetStats() *types.AggregatorStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	statsCopy := *a.stats
	return &statsCopy
}

// validateResponse validates a task response (simplified implementation)
func (a *Aggregator) validateResponse(task *types.Task, response *types.TaskResponse) bool {
	// Basic validation - in a real implementation, this would include signature verification
	if response.Response == "" {
		return false
	}

	// Validate operator address matches registered address
	operator, exists := a.operators[response.OperatorID]
	if !exists || operator.Address != response.OperatorAddr {
		return false
	}

	return true
}

// taskProcessor handles task processing in the background
func (a *Aggregator) taskProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.shutdownChannel:
			return
		case task := <-a.taskChannel:
			a.processTask(task)
		}
	}
}

// processTask handles individual task processing
func (a *Aggregator) processTask(task *types.Task) {
	a.mu.Lock()
	task.Status = types.TaskStatusProcessing
	a.mu.Unlock()

	a.logger.Infof("Processing task %d", task.Index)

	// Simulate task broadcasting to operators
	// In a real implementation, this would send the task to registered operators

	// For now, we just mark it as processing and wait for responses
}

// responseProcessor handles response processing and aggregation
func (a *Aggregator) responseProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.shutdownChannel:
			return
		case response := <-a.responseChannel:
			a.processResponse(response)
		}
	}
}

// processResponse handles individual response processing
func (a *Aggregator) processResponse(response *types.TaskResponse) {
	a.mu.Lock()
	defer a.mu.Unlock()

	task := a.tasks[response.TaskIndex]
	responses := a.taskResponses[response.TaskIndex]

	// Count valid responses
	validResponses := 0
	for _, resp := range responses {
		if resp.IsValid {
			validResponses++
		}
	}

	// Check if we have enough responses to meet quorum
	if validResponses >= int(task.RequiredQuorum) {
		task.Status = types.TaskStatusCompleted
		a.stats.CompletedTasks++
		a.logger.Infof("Task %d completed with %d valid responses", task.Index, validResponses)
	}
}

// taskExpiryChecker periodically checks for expired tasks
func (a *Aggregator) taskExpiryChecker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.shutdownChannel:
			return
		case <-ticker.C:
			a.checkExpiredTasks()
		}
	}
}

// checkExpiredTasks marks expired tasks as failed
func (a *Aggregator) checkExpiredTasks() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for _, task := range a.tasks {
		if (task.Status == types.TaskStatusPending || task.Status == types.TaskStatusProcessing) && now.After(task.ExpiresAt) {
			task.Status = types.TaskStatusExpired
			a.stats.FailedTasks++
			expiredCount++
		}
	}

	if expiredCount > 0 {
		a.logger.Infof("Marked %d tasks as expired", expiredCount)
	}
}

// statsUpdater periodically updates aggregator statistics
func (a *Aggregator) statsUpdater(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.shutdownChannel:
			return
		case <-ticker.C:
			a.updateStats()
		}
	}
}

// updateStats calculates and updates performance statistics
func (a *Aggregator) updateStats() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update active operator count
	activeCount := 0
	for _, operator := range a.operators {
		if operator.IsActive {
			activeCount++
		}
	}
	a.stats.ActiveOperators = activeCount

	// Calculate average response time (simplified)
	totalResponseTime := time.Duration(0)
	responseCount := 0

	for taskIndex, responses := range a.taskResponses {
		task := a.tasks[taskIndex]
		for _, response := range responses {
			if response.IsValid {
				responseTime := response.SubmittedAt.Sub(task.CreatedAt)
				totalResponseTime += responseTime
				responseCount++
			}
		}
	}

	if responseCount > 0 {
		a.stats.AverageResponseTime = totalResponseTime / time.Duration(responseCount)
	}
}

// generateTaskID generates a unique task ID (simplified)
func (a *Aggregator) generateTaskID() string {
	// Generate a random task ID
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return fmt.Sprintf("task_%x", randomBytes)
}
