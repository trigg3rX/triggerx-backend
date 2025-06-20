package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	redisClient "github.com/trigg3rX/triggerx-backend/internal/redis/redis"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type JobStreamManager struct {
	client         *redisClient.Client
	logger         logging.Logger
	consumerGroups map[string]bool
}

func NewJobStreamManager(logger logging.Logger) (*JobStreamManager, error) {
	logger.Info("Initializing JobStreamManager for condition scheduler...")

	if !config.IsRedisAvailable() {
		logger.Error("Redis not available for JobStreamManager initialization")
		metrics.ServiceStatus.WithLabelValues("job_stream_manager").Set(0)
		return nil, fmt.Errorf("redis not available")
	}

	client, err := redisClient.NewRedisClient(logger)
	if err != nil {
		logger.Error("Failed to create Redis client for JobStreamManager", "error", err)
		metrics.ServiceStatus.WithLabelValues("job_stream_manager").Set(0)
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	jsm := &JobStreamManager{
		client:         client,
		logger:         logger,
		consumerGroups: make(map[string]bool),
	}

	logger.Info("JobStreamManager initialized successfully")
	metrics.ServiceStatus.WithLabelValues("job_stream_manager").Set(1)
	return jsm, nil
}

func (jsm *JobStreamManager) Initialize() error {
	jsm.logger.Info("Initializing job streams...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize jobs:running stream (no expiration)
	jsm.logger.Debug("Creating jobs:running stream")
	if err := jsm.client.CreateStreamIfNotExists(ctx, JobsRunningStream, 0); err != nil {
		jsm.logger.Error("Failed to initialize jobs:running stream", "error", err)
		return fmt.Errorf("failed to initialize jobs:running stream: %w", err)
	}

	// Initialize jobs:completed stream (24 hour expiration)
	jsm.logger.Debug("Creating jobs:completed stream") 
	if err := jsm.client.CreateStreamIfNotExists(ctx, JobsCompletedStream, JobsCompletedTTL); err != nil {
		jsm.logger.Error("Failed to initialize jobs:completed stream", "error", err)
		return fmt.Errorf("failed to initialize jobs:completed stream: %w", err)
	}

	jsm.logger.Info("All job streams initialized successfully")
	return nil
}

// AddJobToRunningStream adds a new job to the running stream for condition monitoring
func (jsm *JobStreamManager) AddJobToRunningStream(jobData *JobStreamData, schedulerID int) error {
	jsm.logger.Info("Adding job to running stream",
		"job_id", jobData.JobID,
		"task_definition_id", jobData.TaskDefinitionID,
		"scheduler_id", schedulerID)

	jobJSON, err := json.Marshal(jobData)
	if err != nil {
		jsm.logger.Error("Failed to marshal job data",
			"job_id", jobData.JobID,
			"error", err)
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messageID, err := jsm.client.XAdd(ctx, &redis.XAddArgs{
		Stream: JobsRunningStream,
		ID:     "*",
		Values: map[string]interface{}{
			"job":          string(jobJSON),
			"scheduler_id": schedulerID,
			"created_at":   time.Now().Unix(),
		},
	})

	if err != nil {
		jsm.logger.Error("Failed to add job to running stream",
			"job_id", jobData.JobID,
			"scheduler_id", schedulerID,
			"error", err)
		return fmt.Errorf("failed to add job to running stream: %w", err)
	}

	jsm.logger.Info("Job added to running stream successfully",
		"job_id", jobData.JobID,
		"message_id", messageID,
		"scheduler_id", schedulerID)

	metrics.JobsAddedToStreamTotal.WithLabelValues("running").Inc()
	return nil
}

// ReadJobsFromRunningStream reads jobs from running stream for a specific scheduler
func (jsm *JobStreamManager) ReadJobsFromRunningStream(schedulerID int, consumerName string, count int64) ([]JobStreamData, []string, error) {
	consumerGroup := fmt.Sprintf("scheduler-%d", schedulerID)
	
	jsm.logger.Debug("Reading jobs from running stream",
		"scheduler_id", schedulerID,
		"consumer_group", consumerGroup,
		"consumer_name", consumerName,
		"count", count)

	return jsm.readJobsFromStream(JobsRunningStream, consumerGroup, consumerName, count)
}

// MoveJobToCompleted moves a job from running to completed stream
func (jsm *JobStreamManager) MoveJobToCompleted(jobData *JobStreamData, schedulerID int, messageID string, reason string) error {
	jsm.logger.Info("Moving job to completed stream",
		"job_id", jobData.JobID,
		"scheduler_id", schedulerID,
		"reason", reason)

	// Add to completed stream
	jobData.IsActive = false
	jobJSON, err := json.Marshal(jobData)
	if err != nil {
		jsm.logger.Error("Failed to marshal job data for completed stream",
			"job_id", jobData.JobID,
			"error", err)
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	completedMessageID, err := jsm.client.XAdd(ctx, &redis.XAddArgs{
		Stream: JobsCompletedStream,
		ID:     "*",
		Values: map[string]interface{}{
			"job":          string(jobJSON),
			"scheduler_id": schedulerID,
			"completed_at": time.Now().Unix(),
			"reason":       reason,
		},
	})

	if err != nil {
		jsm.logger.Error("Failed to add job to completed stream",
			"job_id", jobData.JobID,
			"error", err)
		return fmt.Errorf("failed to add job to completed stream: %w", err)
	}

	// Acknowledge original message in running stream
	consumerGroup := fmt.Sprintf("scheduler-%d", schedulerID)
	if err := jsm.AckJobProcessed(JobsRunningStream, consumerGroup, messageID); err != nil {
		jsm.logger.Warn("Failed to acknowledge job in running stream",
			"job_id", jobData.JobID,
			"message_id", messageID,
			"error", err)
		// Don't return error as job was successfully moved to completed
	}

	jsm.logger.Info("Job moved to completed stream successfully",
		"job_id", jobData.JobID,
		"completed_message_id", completedMessageID,
		"reason", reason)

	metrics.JobsAddedToStreamTotal.WithLabelValues("completed").Inc()
	return nil
}

// readJobsFromStream reads jobs from a specific stream using consumer groups
func (jsm *JobStreamManager) readJobsFromStream(stream, consumerGroup, consumerName string, count int64) ([]JobStreamData, []string, error) {
	if err := jsm.RegisterConsumerGroup(stream, consumerGroup); err != nil {
		return nil, nil, err
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streams, err := jsm.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    time.Second,
	})

	duration := time.Since(start)

	if err != nil {
		if err == redis.Nil {
			jsm.logger.Debug("No jobs available in stream",
				"stream", stream,
				"consumer_group", consumerGroup,
				"duration", duration)
			return []JobStreamData{}, []string{}, nil
		}
		jsm.logger.Error("Failed to read from stream",
			"stream", stream,
			"consumer_group", consumerGroup,
			"duration", duration,
			"error", err)
		return nil, nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var jobs []JobStreamData
	var messageIDs []string
	
	for _, stream := range streams {
		for _, message := range stream.Messages {
			jobJSON, exists := message.Values["job"].(string)
			if !exists {
				jsm.logger.Warn("Message missing job data",
					"stream", stream.Stream,
					"message_id", message.ID)
				continue
			}

			var job JobStreamData
			if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
				jsm.logger.Error("Failed to unmarshal job data",
					"stream", stream.Stream,
					"message_id", message.ID,
					"error", err)
				continue
			}

			jobs = append(jobs, job)
			messageIDs = append(messageIDs, message.ID)
			
			jsm.logger.Debug("Job read from stream",
				"job_id", job.JobID,
				"stream", stream.Stream,
				"message_id", message.ID)
		}
	}

	jsm.logger.Info("Jobs read from stream successfully",
		"stream", stream,
		"job_count", len(jobs),
		"duration", duration)

	return jobs, messageIDs, nil
}

// RegisterConsumerGroup registers a consumer group for a stream
func (jsm *JobStreamManager) RegisterConsumerGroup(stream string, group string) error {
	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := jsm.consumerGroups[key]; exists {
		jsm.logger.Debug("Consumer group already exists", "stream", stream, "group", group)
		return nil
	}

	jsm.logger.Info("Registering consumer group", "stream", stream, "group", group)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := jsm.client.CreateConsumerGroup(ctx, stream, group); err != nil {
		jsm.logger.Error("Failed to create consumer group",
			"stream", stream,
			"group", group,
			"error", err)
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}

	jsm.consumerGroups[key] = true
	jsm.logger.Info("Consumer group created successfully", "stream", stream, "group", group)
	return nil
}

// AckJobProcessed acknowledges that a job has been processed
func (jsm *JobStreamManager) AckJobProcessed(stream, consumerGroup, messageID string) error {
	jsm.logger.Debug("Acknowledging job processed",
		"stream", stream,
		"consumer_group", consumerGroup,
		"message_id", messageID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := jsm.client.XAck(ctx, stream, consumerGroup, messageID)
	if err != nil {
		jsm.logger.Error("Failed to acknowledge job",
			"stream", stream,
			"consumer_group", consumerGroup,
			"message_id", messageID,
			"error", err)
		return err
	}

	jsm.logger.Debug("Job acknowledged successfully",
		"stream", stream,
		"message_id", messageID)

	return nil
}

// GetJobStreamInfo returns information about job streams
func (jsm *JobStreamManager) GetJobStreamInfo() map[string]interface{} {
	jsm.logger.Debug("Getting job stream information")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{JobsRunningStream, JobsCompletedStream}

	for _, stream := range streams {
		length, err := jsm.client.XLen(ctx, stream)
		if err != nil {
			jsm.logger.Warn("Failed to get stream length",
				"stream", stream,
				"error", err)
			length = -1
		}
		streamLengths[stream] = length

		// Update stream length metrics
		switch stream {
		case JobsRunningStream:
			metrics.JobStreamLengths.WithLabelValues("running").Set(float64(length))
		case JobsCompletedStream:
			metrics.JobStreamLengths.WithLabelValues("completed").Set(float64(length))
		}
	}

	info := map[string]interface{}{
		"available":              config.IsRedisAvailable(),
		"jobs_completed_ttl":     JobsCompletedTTL.String(),
		"stream_lengths":         streamLengths,
		"consumer_groups":        len(jsm.consumerGroups),
		"max_stream_length":      config.GetStreamMaxLen(),
	}

	jsm.logger.Debug("Job stream information retrieved", "info", info)
	return info
}

func (jsm *JobStreamManager) Close() error {
	jsm.logger.Info("Closing JobStreamManager")

	err := jsm.client.Close()
	if err != nil {
		jsm.logger.Error("Failed to close Redis client", "error", err)
		return err
	}

	jsm.logger.Info("JobStreamManager closed successfully")
	return nil
}
