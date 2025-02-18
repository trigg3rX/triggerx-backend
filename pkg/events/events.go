package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	JobEventChannel = "job_events"
	KeeperEventChannel = "keeper_events"
)

type JobEvent struct {
	Type    string `json:"type"`
	JobID   int64  `json:"job_id"`
	TaskDefinitionID int    `json:"task_definition_id"`
	TimeFrame int64    `json:"time_frame"`
}

type KeeperEvent struct {
	Type    string `json:"type"`
	KeeperID int64    `json:"keeper_id"`
}

type EventBus struct {
	redis *redis.Client
}

var globalEventBus *EventBus

func InitEventBus(redisAddr string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	globalEventBus = &EventBus{
		redis: rdb,
	}
	return nil
}

func GetEventBus() *EventBus {
	return globalEventBus
}

func (eb *EventBus) PublishJobEvent(ctx context.Context, event JobEvent) error {
	logger := logging.GetLogger(logging.Development, logging.DatabaseProcess)

	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	logger.Infof("Publishing event to channel %s: %s", JobEventChannel, string(eventJSON))

	// Publish to Redis
	result := eb.redis.Publish(ctx, JobEventChannel, eventJSON)
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Get number of clients that received the message
	receivers := result.Val()
	logger.Infof("Event published to %d subscribers", receivers)

	return nil
}

func (eb *EventBus) PublishKeeperEvent(ctx context.Context, event KeeperEvent) error {
	logger := logging.GetLogger(logging.Development, logging.DatabaseProcess)

	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	logger.Infof("Publishing event to channel %s: %s", KeeperEventChannel, string(eventJSON))

	// Publish to Redis
	result := eb.redis.Publish(ctx, KeeperEventChannel, eventJSON)
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Get number of clients that received the message
	receivers := result.Val()
	logger.Infof("Event published to %d subscribers", receivers)

	return nil
}

func (eb *EventBus) Redis() *redis.Client {
	return eb.redis
}
