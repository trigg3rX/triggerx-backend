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

const (
	JobsReadyTimeStream      = "jobs:ready:time"
	JobsRetryTimeStream      = "jobs:retry:time"
	JobsReadyEventStream     = "jobs:ready:event"
	JobsRetryEventStream     = "jobs:retry:event"
	JobsReadyConditionStream = "jobs:ready:condition"
	JobsRetryConditionStream = "jobs:retry:condition"
)

// AddJobToStream adds a job to the specified Redis stream (legacy function for backward compatibility)
func AddJobToStream(stream string, job interface{}) error {
	if !config.IsRedisAvailable() {
		// Log the attempt but don't fail the operation
		logger := logging.GetServiceLogger()
		logger.Warnf("Redis not available, skipping job stream addition to %s", stream)
		return nil
	}

	client := GetClient()
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	logger := logging.GetServiceLogger()

	jobJSON, err := json.Marshal(job)
	if err != nil {
		logger.Errorf("Failed to marshal job: %v", err)
		return err
	}

	res, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"job":        jobJSON,
			"created_at": time.Now().Unix(),
		},
	}).Result()

	if err != nil {
		logger.Errorf("Failed to add job to stream %s: %v", stream, err)
		return err
	}

	logger.Infof("Job added to stream %s with ID %s", stream, res)

	// Set TTL on the stream key (refreshes on each add)
	expireRes := client.Expire(ctx, stream, config.GetStreamTTL())
	if expireRes.Err() != nil {
		logger.Warnf("Failed to set TTL on stream %s: %v", stream, expireRes.Err())
	}

	// Check stream length for overflow
	lenRes := client.XLen(ctx, stream)
	if lenRes.Err() == nil && lenRes.Val() >= int64(config.GetStreamMaxLen()) {
		logger.Warnf("Stream %s reached max length (%d)", stream, config.GetStreamMaxLen())
	}

	return nil
}

// AddJobToStream adds a job to the specified Redis stream (instance method)
func (c *Client) AddJobToStream(stream string, job interface{}) error {
	ctx := context.Background()

	jobJSON, err := json.Marshal(job)
	if err != nil {
		c.logger.Errorf("Failed to marshal job: %v", err)
		return err
	}

	res, err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"job":        jobJSON,
			"created_at": time.Now().Unix(),
		},
	}).Result()

	if err != nil {
		c.logger.Errorf("Failed to add job to stream %s: %v", stream, err)
		return err
	}

	c.logger.Infof("Job added to stream %s with ID %s", stream, res)

	// Set TTL on the stream key (refreshes on each add)
	expireRes := c.client.Expire(ctx, stream, config.GetStreamTTL())
	if expireRes.Err() != nil {
		c.logger.Warnf("Failed to set TTL on stream %s: %v", stream, expireRes.Err())
	}

	// Check stream length for overflow
	lenRes := c.client.XLen(ctx, stream)
	if lenRes.Err() == nil && lenRes.Val() >= int64(config.GetStreamMaxLen()) {
		c.logger.Warnf("Stream %s reached max length (%d)", stream, config.GetStreamMaxLen())
	}

	return nil
}

// Global convenience functions for backward compatibility
var defaultLogger logging.Logger

// SetDefaultLogger sets the default logger for global functions
func SetDefaultLogger(logger logging.Logger) {
	defaultLogger = logger
}

// AddJobToStreamGlobal adds a job to stream using the global enhanced client
func AddJobToStreamGlobal(stream string, job interface{}) error {
	if !config.IsRedisAvailable() {
		if defaultLogger == nil {
			defaultLogger = logging.GetServiceLogger()
		}
		defaultLogger.Warnf("Redis not available, skipping job stream addition to %s", stream)
		return nil
	}

	if defaultLogger == nil {
		defaultLogger = logging.GetServiceLogger()
	}

	client := GetGlobalClient(defaultLogger)
	if client == nil {
		return fmt.Errorf("global Redis client not initialized")
	}

	return client.AddJobToStream(stream, job)
}

// StreamInfo returns information about Redis streams configuration
func GetStreamInfo() map[string]interface{} {
	return map[string]interface{}{
		"available":  config.IsRedisAvailable(),
		"max_length": config.GetStreamMaxLen(),
		"ttl":        config.GetStreamTTL().String(),
		"streams": map[string]string{
			"jobs_ready_time":      JobsReadyTimeStream,
			"jobs_retry_time":      JobsRetryTimeStream,
			"jobs_ready_event":     JobsReadyEventStream,
			"jobs_retry_event":     JobsRetryEventStream,
			"jobs_ready_condition": JobsReadyConditionStream,
			"jobs_retry_condition": JobsRetryConditionStream,
		},
	}
}
