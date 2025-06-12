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

	// Scheduler Private Key and Address
	timeSchedulerSigningKey     string
	timeSchedulerSigningAddress string

	// Time Durations
	pollingInterval     time.Duration
	pollingLookAhead    time.Duration
	jobBatchSize        int
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
		timeSchedulerRPCPort:    env.GetEnv("TIME_SCHEDULER_RPC_PORT", "9004"),
		dbServerURL:             env.GetEnv("DBSERVER_RPC_URL", "http://localhost:9002"),
		aggregatorRPCUrl:        env.GetEnv("AGGREGATOR_RPC_URL", "http://localhost:9003"),
		timeSchedulerSigningKey:     env.GetEnv("TIME_SCHEDULER_SIGNING_KEY", ""),
		timeSchedulerSigningAddress: env.GetEnv("TIME_SCHEDULER_ADDRESS", ""),
		pollingInterval:         env.GetEnvDuration("TIME_SCHEDULER_POLLING_INTERVAL", 30*time.Second),
		pollingLookAhead:        env.GetEnvDuration("TIME_SCHEDULER_POLLING_LOOK_AHEAD", 40*time.Minute),
		jobBatchSize:            env.GetEnvInt("TIME_SCHEDULER_JOB_BATCH_SIZE", 15),
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
	if !env.IsValidEthKeyPair(cfg.timeSchedulerSigningKey, cfg.timeSchedulerSigningAddress) {
		return fmt.Errorf("invalid time scheduler signing key pair address: %s", cfg.timeSchedulerSigningAddress)
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

func GetSchedulerSigningKey() string {
	return cfg.timeSchedulerSigningKey
}

func GetSchedulerSigningAddress() string {
	return cfg.timeSchedulerSigningAddress
}

func GetPollingInterval() time.Duration {
	return cfg.pollingInterval
}

func GetPollingLookAhead() time.Duration {
	return cfg.pollingLookAhead
}

func GetJobBatchSize() int {
	return cfg.jobBatchSize
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
