package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	logger logging.Logger
	prefix string
}

// NewRedisCache creates a new instance of RedisCache
func NewRedisCache(addr string, password string, db int, prefix string, logger logging.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		logger: logger,
		prefix: prefix,
	}, nil
}

func (c *RedisCache) jobKey(jobID int64) string {
	return fmt.Sprintf("%s:job:%d", c.prefix, jobID)
}

func (c *RedisCache) jobsKey() string {
	return fmt.Sprintf("%s:jobs", c.prefix)
}

// Get retrieves a job's state from Redis
func (c *RedisCache) Get(ctx context.Context, jobID int64) (*JobState, error) {
	data, err := c.client.Get(ctx, c.jobKey(jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no state found for job %d", jobID)
		}
		return nil, fmt.Errorf("failed to get job state: %w", err)
	}

	var state JobState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job state: %w", err)
	}

	return &state, nil
}

// Set stores a job's state in Redis
func (c *RedisCache) Set(ctx context.Context, jobID int64, state *JobState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal job state: %w", err)
	}

	pipe := c.client.Pipeline()
	pipe.Set(ctx, c.jobKey(jobID), data, 0)
	pipe.SAdd(ctx, c.jobsKey(), jobID)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to set job state: %w", err)
	}

	return nil
}

// Update updates specific fields in a job's state
func (c *RedisCache) Update(ctx context.Context, jobID int64, field string, value interface{}) error {
	state, err := c.Get(ctx, jobID)
	if err != nil {
		return err
	}

	switch field {
	case "last_executed":
		if timestamp, ok := value.(time.Time); ok {
			state.LastExecuted = timestamp
		} else {
			return fmt.Errorf("invalid value type for last_executed")
		}
	case "status":
		if status, ok := value.(string); ok {
			state.Status = status
		} else {
			return fmt.Errorf("invalid value type for status")
		}
	default:
		if state.Metadata == nil {
			state.Metadata = make(map[string]interface{})
		}
		state.Metadata[field] = value
	}

	return c.Set(ctx, jobID, state)
}

// Delete removes a job's state from Redis
func (c *RedisCache) Delete(ctx context.Context, jobID int64) error {
	pipe := c.client.Pipeline()
	pipe.Del(ctx, c.jobKey(jobID))
	pipe.SRem(ctx, c.jobsKey(), jobID)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete job state: %w", err)
	}

	return nil
}

// GetAll retrieves all job states from Redis
func (c *RedisCache) GetAll(ctx context.Context) (map[int64]*JobState, error) {
	jobIDs, err := c.client.SMembers(ctx, c.jobsKey()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job IDs: %w", err)
	}

	states := make(map[int64]*JobState)
	for _, jobIDStr := range jobIDs {
		var jobID int64
		if _, err := fmt.Sscanf(jobIDStr, "%d", &jobID); err != nil {
			c.logger.Warnf("Invalid job ID format: %s", jobIDStr)
			continue
		}

		state, err := c.Get(ctx, jobID)
		if err != nil {
			c.logger.Warnf("Failed to get state for job %d: %v", jobID, err)
			continue
		}
		states[jobID] = state
	}

	return states, nil
}

// SaveState is a no-op for Redis as data is already persisted
func (c *RedisCache) SaveState() error {
	return nil
}

// LoadState is a no-op for Redis as data is already persisted
func (c *RedisCache) LoadState() error {
	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
