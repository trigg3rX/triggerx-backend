package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type JobStreamManager struct {
	client         *Client
	logger         logging.Logger
	consumerGroups map[string]bool
}

func NewJobStreamManager(logger logging.Logger) (*JobStreamManager, error) {
	if !config.IsRedisAvailable() {
		return nil, fmt.Errorf("redis not available")
	}

	client, err := NewRedisClient(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	return &JobStreamManager{
		client:         client,
		logger:         logger,
		consumerGroups: make(map[string]bool),
	}, nil
}

func (jsm *JobStreamManager) Initialize() error {
	streams := []string{
		JobsRunningStream, JobsCompletedStream,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, stream := range streams {
		if err := jsm.client.CreateStreamIfNotExists(ctx, stream, config.GetJobStreamTTL()); err != nil {
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
	}
	return nil
}

func (jsm *JobStreamManager) RegisterConsumerGroup(stream, group string) error {
	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := jsm.consumerGroups[key]; exists {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := jsm.client.CreateConsumerGroup(ctx, stream, group); err != nil {
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}

	jsm.consumerGroups[key] = true
	jsm.logger.Infof("Created consumer group '%s' for stream '%s'", group, stream)
	return nil
}

func (jsm *JobStreamManager) AddJobToRunningStream(job *JobStreamData) error {
	return jsm.addJobToStream(JobsRunningStream, job)
}

func (jsm *JobStreamManager) AddJobToCompletedStream(job *JobStreamData, executionResult map[string]interface{}) error {
	return jsm.addJobToStream(JobsCompletedStream, job)
}

func (jsm *JobStreamManager) addJobToStream(stream string, job *JobStreamData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jobJSON, err := json.Marshal(job)
	if err != nil {
		jsm.logger.Errorf("Failed to marshal job data: %v", err)
		return err
	}

	res, err := jsm.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"job":       jobJSON,
			"created_at": time.Now().Unix(),
		},
	})

	if err != nil {
		jsm.logger.Errorf("Failed to add job to stream %s: %v", stream, err)
		return err
	}

	jsm.logger.Debugf("Job %d added to stream %s with ID %s", job.JobID, stream, res)
	return nil
}

func (jsm *JobStreamManager) ReadJobsFromRunningStream(consumerGroup, consumerName string, count int64) ([]JobStreamData, error) {
	return jsm.readJobsFromStream(JobsRunningStream, consumerGroup, consumerName, count)
}

func (jsm *JobStreamManager) ReadJobsFromCompletedStream(consumerGroup, consumerName string, count int64) ([]JobStreamData, error) {
	jobs, err := jsm.readJobsFromStream(JobsCompletedStream, consumerGroup, consumerName, count)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var readyJobs []JobStreamData
	for _, job := range jobs {
		if job.ExpireTime == nil || job.ExpireTime.Before(now) {
			readyJobs = append(readyJobs, job)
		}
	}
	return readyJobs, nil
}

func (jsm *JobStreamManager) readJobsFromStream(stream, consumerGroup, consumerName string, count int64) ([]JobStreamData, error) {
	if err := jsm.RegisterConsumerGroup(stream, consumerGroup); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streams, err := jsm.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    time.Second,
	})

	if err != nil {
		if err == redis.Nil {
			return []JobStreamData{}, nil
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var jobs []JobStreamData
	for _, stream := range streams {
		for _, message := range stream.Messages {
			jobJSON, exists := message.Values["job"].(string)
			if !exists {
				continue
			}

			var job JobStreamData
			if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
				jsm.logger.Errorf("Failed to unmarshal job data: %v", err)
				continue
			}

			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (jsm *JobStreamManager) AckJobCompleted(stream, consumerGroup, messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return jsm.client.XAck(ctx, stream, consumerGroup, messageID)
}

func (jsm *JobStreamManager) GetStreamInfo() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{JobsRunningStream, JobsCompletedStream}

	for _, stream := range streams {
		length, err := jsm.client.XLen(ctx, stream)
		if err != nil {
			length = -1
		}
		streamLengths[stream] = length
	}

	return map[string]interface{}{
		"available":      config.IsRedisAvailable(),
		"max_length":     config.GetStreamMaxLen(),
		"ttl":            config.GetJobStreamTTL().String(),
		"stream_lengths": streamLengths,
		"max_retries":    MaxRetryAttempts,
		"retry_backoff":  RetryBackoffBase.String(),
	}
}

func (jsm *JobStreamManager) Close() error {
	return jsm.client.Close()
}
