package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Scheduler RPC Port
	timeSchedulerRPCPort string

	// Database RPC URL
	dbServerURL string
	// Aggregator RPC URL
	aggregatorRPCUrl string
	// Redis API URL
	redisRPCUrl string

	// Scheduler ID
	timeSchedulerID int

	// Time Durations
	pollingInterval     time.Duration
	pollingLookAhead    time.Duration
	taskBatchSize       int
	performerLockTTL    time.Duration
	taskCacheTTL        time.Duration
	duplicateTaskWindow time.Duration
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                 env.GetEnvBool("DEV_MODE", false),
		timeSchedulerRPCPort:    env.GetEnvString("TIME_SCHEDULER_RPC_PORT", "9005"),
		redisRPCUrl:             env.GetEnvString("REDIS_RPC_URL", "http://localhost:9003"),
		dbServerURL:             env.GetEnvString("DBSERVER_RPC_URL", "http://localhost:9002"),
		aggregatorRPCUrl:        env.GetEnvString("AGGREGATOR_RPC_URL", "http://localhost:9001"),
		pollingInterval:         env.GetEnvDuration("TIME_SCHEDULER_POLLING_INTERVAL", 30*time.Second),
		pollingLookAhead:        env.GetEnvDuration("TIME_SCHEDULER_POLLING_LOOKAHEAD", 40*time.Minute),
		taskBatchSize:            env.GetEnvInt("TIME_SCHEDULER_TASK_BATCH_SIZE", 15),
		performerLockTTL:        env.GetEnvDuration("TIME_SCHEDULER_PERFORMER_LOCK_TTL", 31*time.Second),
		taskCacheTTL:            env.GetEnvDuration("TIME_SCHEDULER_TASK_CACHE_TTL", 1*time.Minute),
		duplicateTaskWindow:     env.GetEnvDuration("TIME_SCHEDULER_DUPLICATE_TASK_WINDOW", 1*time.Minute),
	}
	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func validateConfig() error {
	if !env.IsValidPort(cfg.timeSchedulerRPCPort) {
		return fmt.Errorf("invalid time scheduler RPC port: %s", cfg.timeSchedulerRPCPort)
	}
	if !env.IsValidURL(cfg.dbServerURL) {
		return fmt.Errorf("invalid database server URL: %s", cfg.dbServerURL)
	}
	if !env.IsValidURL(cfg.aggregatorRPCUrl) {
		return fmt.Errorf("invalid aggregator RPC URL: %s", cfg.aggregatorRPCUrl)
	}
	if !env.IsValidURL(cfg.redisRPCUrl) {
		return fmt.Errorf("invalid Redis API URL: %s", cfg.redisRPCUrl)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetSchedulerRPCPort() string {
	return cfg.timeSchedulerRPCPort
}

func GetDBServerURL() string {
	return cfg.dbServerURL
}

func GetAggregatorRPCUrl() string {
	return cfg.aggregatorRPCUrl
}

func GetRedisRPCUrl() string {
	return cfg.redisRPCUrl
}

func GetSchedulerID() int {
	return cfg.timeSchedulerID
}

func GetPollingInterval() time.Duration {
	return cfg.pollingInterval
}

func GetPollingLookAhead() time.Duration {
	return cfg.pollingLookAhead
}

func GetTaskBatchSize() int {
	return cfg.taskBatchSize
}

func GetPerformerLockTTL() time.Duration {
	return cfg.performerLockTTL
}

func GetTaskCacheTTL() time.Duration {
	return cfg.taskCacheTTL
}

func GetDuplicateTaskWindow() time.Duration {
	return cfg.duplicateTaskWindow
}
