package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	// Redis key prefix for storing active job data
	activeJobKeyPrefix = "condition_scheduler:active_job:"
	// Redis key prefix for storing last trigger timestamp
	lastTriggerKeyPrefix = "condition_scheduler:last_trigger:"
)

// Client wraps the Redis client for condition scheduler operations
type Client struct {
	client *redis.Client
	logger observability.Logger
}

// NewClient creates a new Redis client for the condition scheduler service
func NewClient(logger observability.Logger) (*Client, error) {
	opt, err := redis.ParseURL(config.GetUpstashRedisUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	if config.GetUpstashRedisRestToken() != "" {
		opt.Password = config.GetUpstashRedisRestToken()
	}

	client := redis.NewClient(opt)

	redisClient := &Client{
		client: client,
		logger: logger,
	}

	if err := redisClient.CheckConnection(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

// CheckConnection validates the Redis connection
func (c *Client) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		c.logger.Error(ctx, "Failed to connect to Redis", observability.Error(err))
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.logger.Info(ctx, "Successfully connected to Redis")
	return nil
}

// ActiveJobData represents the minimal job data stored in Redis for recovery
type ActiveJobData struct {
	JobID            string    `json:"job_id"`
	TaskDefinitionID int       `json:"task_definition_id"`
	ScheduledAt      time.Time `json:"scheduled_at"`
}

// StoreActiveJob stores an active job in Redis for crash recovery
func (c *Client) StoreActiveJob(ctx context.Context, schedulerID string, jobID string, taskDefID int) error {
	key := fmt.Sprintf("%s%s:%s", activeJobKeyPrefix, schedulerID, jobID)

	data := ActiveJobData{
		JobID:            jobID,
		TaskDefinitionID: taskDefID,
		ScheduledAt:      time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Store with 24 hour TTL
	err = c.client.Set(ctx, key, jsonData, 24*time.Hour).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to store active job",
			observability.Error(err),
			observability.String("job_id", jobID),
		)
		return fmt.Errorf("failed to store active job: %w", err)
	}

	c.logger.Debug(ctx, "Stored active job in Redis",
		observability.String("job_id", jobID),
	)
	return nil
}

// RemoveActiveJob removes an active job from Redis
func (c *Client) RemoveActiveJob(ctx context.Context, schedulerID string, jobID string) error {
	key := fmt.Sprintf("%s%s:%s", activeJobKeyPrefix, schedulerID, jobID)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to remove active job",
			observability.Error(err),
			observability.String("job_id", jobID),
		)
		return fmt.Errorf("failed to remove active job: %w", err)
	}

	c.logger.Debug(ctx, "Removed active job from Redis",
		observability.String("job_id", jobID),
	)
	return nil
}

// GetActiveJobs retrieves all active jobs for a scheduler (for crash recovery)
func (c *Client) GetActiveJobs(ctx context.Context, schedulerID string) ([]ActiveJobData, error) {
	pattern := fmt.Sprintf("%s%s:*", activeJobKeyPrefix, schedulerID)

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Error(ctx, "Failed to get active job keys",
			observability.Error(err),
		)
		return nil, fmt.Errorf("failed to get active job keys: %w", err)
	}

	var activeJobs []ActiveJobData
	for _, key := range keys {
		val, err := c.client.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			c.logger.Warn(ctx, "Failed to get job data",
				observability.Error(err),
				observability.String("key", key),
			)
			continue
		}

		var jobData ActiveJobData
		if err := json.Unmarshal([]byte(val), &jobData); err != nil {
			c.logger.Warn(ctx, "Failed to unmarshal job data",
				observability.Error(err),
				observability.String("key", key),
			)
			continue
		}

		activeJobs = append(activeJobs, jobData)
	}

	return activeJobs, nil
}

// SetLastTriggerTime stores the last trigger timestamp for a job (for cooldown)
func (c *Client) SetLastTriggerTime(ctx context.Context, schedulerID string, jobID string, timestamp time.Time) error {
	key := fmt.Sprintf("%s%s:%s", lastTriggerKeyPrefix, schedulerID, jobID)

	// Store with 1 hour TTL (should be enough for cooldown purposes)
	err := c.client.Set(ctx, key, timestamp.Format(time.RFC3339Nano), time.Hour).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to set last trigger time",
			observability.Error(err),
			observability.String("job_id", jobID),
		)
		return fmt.Errorf("failed to set last trigger time: %w", err)
	}

	return nil
}

// GetLastTriggerTime retrieves the last trigger timestamp for a job
func (c *Client) GetLastTriggerTime(ctx context.Context, schedulerID string, jobID string) (time.Time, error) {
	key := fmt.Sprintf("%s%s:%s", lastTriggerKeyPrefix, schedulerID, jobID)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return time.Time{}, nil
	} else if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last trigger time: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339Nano, val)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last trigger time: %w", err)
	}

	return timestamp, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}
