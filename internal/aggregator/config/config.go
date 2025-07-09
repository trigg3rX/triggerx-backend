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

	// Port configuration
	aggregatorRPCPort string
	aggregatorP2PPort string

	// Task management configuration
	maxConcurrentTasks  int
	defaultTaskTimeout  time.Duration
	taskCleanupInterval time.Duration

	// Operator configuration
	minOperators            int
	maxOperators            int
	operatorTimeoutDuration time.Duration

	// Performance and limits
	requestTimeout   time.Duration
	maxRetryAttempts int

	// Blockchain configuration (for future use)
	ethRPCURL string
	chainID   int64

	// Security configuration
	enableSignatureValidation bool
	requireStakedOperators    bool
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Parse duration values
	defaultTaskTimeout, err := parseDurationWithDefault(env.GetEnv("AGGREGATOR_DEFAULT_TASK_TIMEOUT", "5m"), 5*time.Minute)
	if err != nil {
		return fmt.Errorf("invalid AGGREGATOR_DEFAULT_TASK_TIMEOUT: %w", err)
	}

	taskCleanupInterval, err := parseDurationWithDefault(env.GetEnv("AGGREGATOR_TASK_CLEANUP_INTERVAL", "1h"), 1*time.Hour)
	if err != nil {
		return fmt.Errorf("invalid AGGREGATOR_TASK_CLEANUP_INTERVAL: %w", err)
	}

	operatorTimeoutDuration, err := parseDurationWithDefault(env.GetEnv("AGGREGATOR_OPERATOR_TIMEOUT", "30s"), 30*time.Second)
	if err != nil {
		return fmt.Errorf("invalid AGGREGATOR_OPERATOR_TIMEOUT: %w", err)
	}

	requestTimeout, err := parseDurationWithDefault(env.GetEnv("AGGREGATOR_REQUEST_TIMEOUT", "30s"), 30*time.Second)
	if err != nil {
		return fmt.Errorf("invalid AGGREGATOR_REQUEST_TIMEOUT: %w", err)
	}

	cfg = Config{
		devMode:           env.GetEnvBool("DEV_MODE", false),
		aggregatorRPCPort: env.GetEnv("AGGREGATOR_RPC_PORT", "9007"),
		aggregatorP2PPort: env.GetEnv("AGGREGATOR_P2P_PORT", "9008"),

		// Task management
		maxConcurrentTasks:  env.GetEnvInt("AGGREGATOR_MAX_CONCURRENT_TASKS", 100),
		defaultTaskTimeout:  defaultTaskTimeout,
		taskCleanupInterval: taskCleanupInterval,

		// Operator configuration
		minOperators:            env.GetEnvInt("AGGREGATOR_MIN_OPERATORS", 1),
		maxOperators:            env.GetEnvInt("AGGREGATOR_MAX_OPERATORS", 100),
		operatorTimeoutDuration: operatorTimeoutDuration,

		// Performance and limits
		requestTimeout:   requestTimeout,
		maxRetryAttempts: env.GetEnvInt("AGGREGATOR_MAX_RETRY_ATTEMPTS", 3),

		// Blockchain configuration
		ethRPCURL: env.GetEnv("ETH_RPC_URL", ""),
		chainID:   int64(env.GetEnvInt("CHAIN_ID", 1)),

		// Security configuration
		enableSignatureValidation: env.GetEnvBool("AGGREGATOR_ENABLE_SIGNATURE_VALIDATION", false),
		requireStakedOperators:    env.GetEnvBool("AGGREGATOR_REQUIRE_STAKED_OPERATORS", false),
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
	if !env.IsValidPort(cfg.aggregatorRPCPort) {
		return fmt.Errorf("invalid Aggregator RPC Port: %s", cfg.aggregatorRPCPort)
	}
	if !env.IsValidPort(cfg.aggregatorP2PPort) {
		return fmt.Errorf("invalid Aggregator P2P Port: %s", cfg.aggregatorP2PPort)
	}

	// Validate task configuration
	if cfg.maxConcurrentTasks < 1 {
		return fmt.Errorf("maxConcurrentTasks must be at least 1")
	}
	if cfg.defaultTaskTimeout < time.Second {
		return fmt.Errorf("defaultTaskTimeout must be at least 1 second")
	}

	// Validate operator configuration
	if cfg.minOperators < 1 {
		return fmt.Errorf("minOperators must be at least 1")
	}
	if cfg.maxOperators < cfg.minOperators {
		return fmt.Errorf("maxOperators must be greater than or equal to minOperators")
	}

	// Validate performance configuration
	if cfg.maxRetryAttempts < 0 {
		return fmt.Errorf("maxRetryAttempts must be non-negative")
	}
	if cfg.requestTimeout < time.Second {
		return fmt.Errorf("requestTimeout must be at least 1 second")
	}

	return nil
}

// parseDurationWithDefault parses a duration string, returning the default if empty or invalid
func parseDurationWithDefault(durationStr string, defaultDuration time.Duration) (time.Duration, error) {
	if durationStr == "" {
		return defaultDuration, nil
	}
	return time.ParseDuration(durationStr)
}

// Basic getters
func IsDevMode() bool {
	return cfg.devMode
}

func GetAggregatorRPCPort() string {
	return cfg.aggregatorRPCPort
}

func GetAggregatorP2PPort() string {
	return cfg.aggregatorP2PPort
}

// Task management getters
func GetMaxConcurrentTasks() int {
	return cfg.maxConcurrentTasks
}

func GetDefaultTaskTimeout() time.Duration {
	return cfg.defaultTaskTimeout
}

func GetTaskCleanupInterval() time.Duration {
	return cfg.taskCleanupInterval
}

// Operator configuration getters
func GetMinOperators() int {
	return cfg.minOperators
}

func GetMaxOperators() int {
	return cfg.maxOperators
}

func GetOperatorTimeoutDuration() time.Duration {
	return cfg.operatorTimeoutDuration
}

// Performance configuration getters
func GetRequestTimeout() time.Duration {
	return cfg.requestTimeout
}

func GetMaxRetryAttempts() int {
	return cfg.maxRetryAttempts
}

// Blockchain configuration getters
func GetEthRPCURL() string {
	return cfg.ethRPCURL
}

func GetChainID() int64 {
	return cfg.chainID
}

// Security configuration getters
func IsSignatureValidationEnabled() bool {
	return cfg.enableSignatureValidation
}

func IsStakedOperatorsRequired() bool {
	return cfg.requireStakedOperators
}

// GetAggregatorConfig returns a complete configuration object for the aggregator
func GetAggregatorConfig() map[string]interface{} {
	return map[string]interface{}{
		"dev_mode":                  cfg.devMode,
		"rpc_port":                  cfg.aggregatorRPCPort,
		"p2p_port":                  cfg.aggregatorP2PPort,
		"max_concurrent_tasks":      cfg.maxConcurrentTasks,
		"default_task_timeout":      cfg.defaultTaskTimeout.String(),
		"task_cleanup_interval":     cfg.taskCleanupInterval.String(),
		"min_operators":             cfg.minOperators,
		"max_operators":             cfg.maxOperators,
		"operator_timeout_duration": cfg.operatorTimeoutDuration.String(),
		"request_timeout":           cfg.requestTimeout.String(),
		"max_retry_attempts":        cfg.maxRetryAttempts,
		"eth_rpc_url":               cfg.ethRPCURL,
		"chain_id":                  cfg.chainID,
		"signature_validation":      cfg.enableSignatureValidation,
		"require_staked_operators":  cfg.requireStakedOperators,
	}
}
