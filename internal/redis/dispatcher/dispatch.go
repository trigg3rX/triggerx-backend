package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DispatchPayload represents the data structure sent to the aggregator
type DispatchPayload struct {
	JobData     *types.HandleCreateJobData `json:"job_data"`
	TriggerData *types.TriggerData         `json:"trigger_data"`
	Metadata    *DispatchMetadata          `json:"metadata"`
}

// DispatchMetadata contains additional dispatch information
type DispatchMetadata struct {
	DispatchID    string    `json:"dispatch_id"`
	JobID         int64     `json:"job_id"`
	TriggerType   string    `json:"trigger_type"` // "time", "event", "condition"
	Timestamp     time.Time `json:"timestamp"`
	AttemptNumber int       `json:"attempt_number"`
	StreamSource  string    `json:"stream_source"` // Which Redis stream this came from
}

// DispatchResult represents the result of a dispatch operation
type DispatchResult struct {
	Success     bool      `json:"success"`
	DispatchID  string    `json:"dispatch_id"`
	JobID       int64     `json:"job_id"`
	Error       error     `json:"error,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	RetryNeeded bool      `json:"retry_needed"`
}

// Dispatcher handles sending job data to the aggregator
type Dispatcher struct {
	logger           logging.Logger
	redisClient      *redis.Client
	aggregatorClient *aggregator.AggregatorClient
	config           DispatcherConfig
}

// DispatcherConfig holds configuration for the dispatcher
type DispatcherConfig struct {
	MaxRetryAttempts int           `json:"max_retry_attempts"`
	RetryDelay       time.Duration `json:"retry_delay"`
	DispatchTimeout  time.Duration `json:"dispatch_timeout"`
	BatchSize        int           `json:"batch_size"`
	EnableDeadLetter bool          `json:"enable_dead_letter"`
	MetricsEnabled   bool          `json:"metrics_enabled"`
}

// NewDispatcher creates a new dispatcher instance
func NewDispatcher(logger logging.Logger, redisClient *redis.Client, aggregatorClient *aggregator.AggregatorClient, config DispatcherConfig) *Dispatcher {
	// Set default values if not provided
	if config.MaxRetryAttempts == 0 {
		config.MaxRetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.DispatchTimeout == 0 {
		config.DispatchTimeout = 30 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}

	return &Dispatcher{
		logger:           logger,
		redisClient:      redisClient,
		aggregatorClient: aggregatorClient,
		config:           config,
	}
}

// DispatchToAggregator sends job and trigger data to the aggregator using SendCustomMessage
func (d *Dispatcher) DispatchToAggregator(ctx context.Context, jobData *types.HandleCreateJobData, triggerData *types.TriggerData, metadata *DispatchMetadata) (*DispatchResult, error) {
	d.logger.Infof("Dispatching job to aggregator: JobID=%d, DispatchID=%s, Attempt=%d",
		jobData.JobID, metadata.DispatchID, metadata.AttemptNumber)

	// Create dispatch payload
	payload := &DispatchPayload{
		JobData:     jobData,
		TriggerData: triggerData,
		Metadata:    metadata,
	}

	// Marshal payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		d.logger.Errorf("Failed to marshal dispatch payload: %v", err)
		return &DispatchResult{
			Success:     false,
			DispatchID:  metadata.DispatchID,
			JobID:       jobData.JobID,
			Error:       fmt.Errorf("failed to marshal payload: %w", err),
			Timestamp:   time.Now(),
			RetryNeeded: false, // Marshaling errors don't need retry
		}, err
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, d.config.DispatchTimeout)
	defer cancel()

	// Send to aggregator using SendCustomMessage
	success, err := d.sendToAggregator(ctxWithTimeout, payloadJSON, jobData.TaskDefinitionID)

	result := &DispatchResult{
		Success:     success,
		DispatchID:  metadata.DispatchID,
		JobID:       jobData.JobID,
		Timestamp:   time.Now(),
		RetryNeeded: !success && d.shouldRetry(err, metadata.AttemptNumber),
	}

	if err != nil {
		result.Error = err
		d.logger.Errorf("Failed to dispatch job %d to aggregator: %v", jobData.JobID, err)
	} else {
		d.logger.Infof("Successfully dispatched job %d to aggregator", jobData.JobID)
	}

	return result, err
}

// sendToAggregator performs the actual RPC call to the aggregator
func (d *Dispatcher) sendToAggregator(ctx context.Context, payloadJSON []byte, taskDefinitionID int) (bool, error) {
	// Convert JSON payload to hex string for RPC call
	dataHex := fmt.Sprintf("0x%x", payloadJSON)

	// Create a custom RPC call context - we'll use the aggregator client's RPC connection
	// but call sendCustomMessage directly
	success, err := d.callSendCustomMessage(ctx, dataHex, taskDefinitionID)
	if err != nil {
		return false, fmt.Errorf("sendCustomMessage RPC call failed: %w", err)
	}

	return success, nil
}

// callSendCustomMessage makes the direct RPC call to sendCustomMessage
func (d *Dispatcher) callSendCustomMessage(ctx context.Context, data string, taskDefinitionID int) (bool, error) {
	// Import the RPC client to make direct calls
	rpcClient, err := d.getRPCClient()
	if err != nil {
		return false, fmt.Errorf("failed to get RPC client: %w", err)
	}
	defer rpcClient.Close()

	// Make direct RPC call to sendCustomMessage
	var result interface{}
	err = rpcClient.CallContext(ctx, &result, "sendCustomMessage", data, taskDefinitionID)
	if err != nil {
		return false, fmt.Errorf("sendCustomMessage RPC call failed: %w", err)
	}

	d.logger.Debugf("SendCustomMessage call successful for taskDefinitionID: %d, result: %v", taskDefinitionID, result)
	return true, nil
}

// getRPCClient creates an RPC client using the aggregator configuration
func (d *Dispatcher) getRPCClient() (RPCClient, error) {
	// We need access to the aggregator's RPC address
	// For now, we'll create a wrapper interface that allows us to access the client

	// This is a temporary solution - ideally the aggregator client should expose SendCustomMessage
	return NewRPCClientWrapper(d.aggregatorClient)
}

// RPCClient interface for RPC operations
type RPCClient interface {
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
	Close() error
}

// RPCClientWrapper wraps the aggregator client to provide RPC access
type RPCClientWrapper struct {
	aggregatorClient *aggregator.AggregatorClient
}

// NewRPCClientWrapper creates a new RPC client wrapper
func NewRPCClientWrapper(aggregatorClient *aggregator.AggregatorClient) (RPCClient, error) {
	return &RPCClientWrapper{
		aggregatorClient: aggregatorClient,
	}, nil
}

// CallContext implements the RPC call - this is a placeholder that uses the existing aggregator client
func (w *RPCClientWrapper) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	// For now, we'll use the existing SendTaskResult method as a workaround
	// In production, you'd want to add SendCustomMessage to the aggregator client

	if method == "sendCustomMessage" && len(args) >= 2 {
		data, ok1 := args[0].(string)
		taskDefID, ok2 := args[1].(int)

		if !ok1 || !ok2 {
			return fmt.Errorf("invalid arguments for sendCustomMessage")
		}

		// Use SendTaskResult as a workaround
		taskResult := &aggregator.TaskResult{
			ProofOfTask:      "dispatch_proof",
			Data:             data,
			TaskDefinitionID: taskDefID,
			PerformerAddress: "dispatcher",
		}

		return w.aggregatorClient.SendTaskResult(ctx, taskResult)
	}

	return fmt.Errorf("unsupported RPC method: %s", method)
}

// Close implements the close method
func (w *RPCClientWrapper) Close() error {
	// The aggregator client handles its own cleanup
	return nil
}

// shouldRetry determines if a dispatch should be retried based on error and attempt count
func (d *Dispatcher) shouldRetry(err error, attemptNumber int) bool {
	if attemptNumber >= d.config.MaxRetryAttempts {
		return false
	}

	// Don't retry on certain errors (like invalid payload)
	if err == nil {
		return false
	}

	// Add logic for specific error types that shouldn't be retried
	errorStr := err.Error()
	nonRetryableErrors := []string{
		"invalid payload",
		"marshaling",
		"authentication failed",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if len(errorStr) > 0 && len(nonRetryable) > 0 &&
			len(errorStr) >= len(nonRetryable) && errorStr[:len(nonRetryable)] == nonRetryable {
			return false
		}
	}

	return true
}

// ProcessReadyJobs processes jobs from the Redis ready stream and dispatches them
func (d *Dispatcher) ProcessReadyJobs(ctx context.Context, streamName string) error {
	d.logger.Infof("Starting to process ready jobs from stream: %s", streamName)

	// This method would read from Redis streams and dispatch jobs
	// Implementation would depend on your specific Redis stream reading pattern

	// Example structure:
	// 1. Read from Redis stream
	// 2. Parse job data
	// 3. Call DispatchToAggregator
	// 4. Handle results (move to completed/retry streams)

	d.logger.Warnf("ProcessReadyJobs: Implementation pending - stream reading logic needed")
	return fmt.Errorf("not implemented: ProcessReadyJobs for stream %s", streamName)
}

// HandleDispatchFailure moves failed dispatches to retry stream or dead letter queue
func (d *Dispatcher) HandleDispatchFailure(ctx context.Context, result *DispatchResult, payload *DispatchPayload) error {
	if result.RetryNeeded {
		// Move to retry stream
		return d.moveToRetryStream(ctx, payload, result)
	} else {
		// Move to dead letter queue
		return d.moveToDeadLetterQueue(ctx, payload, result)
	}
}

// moveToRetryStream adds the job to the retry stream with incremented attempt count
func (d *Dispatcher) moveToRetryStream(ctx context.Context, payload *DispatchPayload, result *DispatchResult) error {
	payload.Metadata.AttemptNumber++
	payload.Metadata.Timestamp = time.Now()

	// Add to retry stream
	err := d.redisClient.AddJobToStream(redis.JobsRetryStream, payload)
	if err != nil {
		d.logger.Errorf("Failed to add job %d to retry stream: %v", payload.JobData.JobID, err)
		return err
	}

	d.logger.Infof("Job %d moved to retry stream (attempt %d)",
		payload.JobData.JobID, payload.Metadata.AttemptNumber)
	return nil
}

// moveToDeadLetterQueue adds the job to the failed stream
func (d *Dispatcher) moveToDeadLetterQueue(ctx context.Context, payload *DispatchPayload, result *DispatchResult) error {
	if !d.config.EnableDeadLetter {
		d.logger.Warnf("Dead letter queue disabled, dropping failed job %d", payload.JobData.JobID)
		return nil
	}

	// Add failure metadata
	failureData := map[string]interface{}{
		"payload":        payload,
		"failure_reason": result.Error.Error(),
		"final_attempt":  payload.Metadata.AttemptNumber,
		"failed_at":      time.Now(),
	}

	err := d.redisClient.AddJobToStream(redis.JobsFailedStream, failureData)
	if err != nil {
		d.logger.Errorf("Failed to add job %d to dead letter queue: %v", payload.JobData.JobID, err)
		return err
	}

	d.logger.Warnf("Job %d moved to dead letter queue after %d attempts",
		payload.JobData.JobID, payload.Metadata.AttemptNumber)
	return nil
}

// GetDispatchStats returns statistics about dispatch operations
func (d *Dispatcher) GetDispatchStats() map[string]interface{} {
	if !d.config.MetricsEnabled {
		return map[string]interface{}{
			"metrics_enabled": false,
		}
	}

	// Implementation would track and return dispatch metrics
	return map[string]interface{}{
		"metrics_enabled":     true,
		"max_retry_attempts":  d.config.MaxRetryAttempts,
		"retry_delay":         d.config.RetryDelay.String(),
		"dispatch_timeout":    d.config.DispatchTimeout.String(),
		"batch_size":          d.config.BatchSize,
		"dead_letter_enabled": d.config.EnableDeadLetter,
	}
}
