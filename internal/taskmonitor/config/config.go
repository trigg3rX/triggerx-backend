package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Task Monitor RPC port
	taskMonitorRPCPort string

	// Contract Addresses to listen for events
	attestationCenterAddress     string
	testAttestationCenterAddress string

	// RPC URLs for Ethereum and Base
	rpcProvider string
	rpcAPIKey   string

	// ScyllaDB Host and Port
	databaseHostAddress string
	databaseHostPort    string

	// Upstash Redis URL and Rest Token
	upstashRedisUrl       string
	upstashRedisRestToken string

	// Sync Configs Update
	lastBaseBlockUpdated uint64

	// Pinata JWT and Host
	pinataJWT  string
	pinataHost string

	// OpenTelemetry endpoint
	ottempoEndpoint string

	// Notification webhook
	notifyWebhookURL   string
	notifyWebhookToken string

	// SMTP email settings
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPass     string
	smtpFrom     string
	smtpStartTLS bool

	// Common settings
	poolSize     int
	minIdleConns int
	maxRetries   int

	// Timeout settings
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	poolTimeout  time.Duration

	// Stream settings
	streamMaxLen    int
	jobStreamTTL    time.Duration
	taskStreamTTL   time.Duration
	cacheTTL        time.Duration
	cleanupInterval time.Duration

	// Metrics settings
	metricsUpdateInterval time.Duration

	// Timeout and retry settings
	retryDelay            time.Duration
	requestTimeout        time.Duration
	initializationTimeout time.Duration
	maxRetryBackoff       time.Duration
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                      env.GetEnvBool("DEV_MODE", false),
		taskMonitorRPCPort:           env.GetEnvString("TASK_MONITOR_RPC_PORT", "9007"),
		attestationCenterAddress:     env.GetEnvString("ATTESTATION_CENTER_ADDRESS", ""),
		testAttestationCenterAddress: env.GetEnvString("TEST_ATTESTATION_CENTER_ADDRESS", ""),
		rpcProvider:                  env.GetEnvString("RPC_PROVIDER", ""),
		rpcAPIKey:                    env.GetEnvString("RPC_API_KEY", ""),
		databaseHostAddress:          env.GetEnvString("DATABASE_HOST_ADDRESS", ""),
		databaseHostPort:             env.GetEnvString("DATABASE_HOST_PORT", ""),
		upstashRedisUrl:              env.GetEnvString("UPSTASH_REDIS_URL", ""),
		upstashRedisRestToken:        env.GetEnvString("UPSTASH_REDIS_REST_TOKEN", ""),
		pinataJWT:                    env.GetEnvString("PINATA_JWT", ""),
		pinataHost:                   env.GetEnvString("PINATA_HOST", ""),
		poolSize:                     env.GetEnvInt("REDIS_POOL_SIZE", 10),
		minIdleConns:                 env.GetEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		maxRetries:                   env.GetEnvInt("REDIS_MAX_RETRIES", 3),
		dialTimeout:                  env.GetEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		readTimeout:                  env.GetEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		writeTimeout:                 env.GetEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		poolTimeout:                  env.GetEnvDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
		streamMaxLen:                 env.GetEnvInt("REDIS_STREAM_MAX_LEN", 10000),
		jobStreamTTL:                 env.GetEnvDuration("REDIS_JOB_STREAM_TTL", 120*time.Hour),
		taskStreamTTL:                env.GetEnvDuration("REDIS_TASK_STREAM_TTL", 1*time.Hour),
		cacheTTL:                     env.GetEnvDuration("REDIS_CACHE_TTL", 24*time.Hour),
		cleanupInterval:              env.GetEnvDuration("REDIS_CLEANUP_INTERVAL", 10*time.Minute),
		metricsUpdateInterval:        env.GetEnvDuration("REDIS_METRICS_UPDATE_INTERVAL", 30*time.Second),
		retryDelay:                   env.GetEnvDuration("REDIS_RETRY_DELAY", 2*time.Second),
		requestTimeout:               env.GetEnvDuration("REDIS_REQUEST_TIMEOUT", 10*time.Second),
		initializationTimeout:        env.GetEnvDuration("REDIS_INITIALIZATION_TIMEOUT", 10*time.Second),
		maxRetryBackoff:              env.GetEnvDuration("REDIS_MAX_RETRY_BACKOFF", 5*time.Minute),
		ottempoEndpoint:              env.GetEnvString("TEMPO_OTLP_ENDPOINT", "localhost:4318"),
		notifyWebhookURL:             env.GetEnvString("TASK_NOTIFY_WEBHOOK_URL", ""),
		notifyWebhookToken:           env.GetEnvString("TASK_NOTIFY_WEBHOOK_TOKEN", ""),
		smtpHost:                     env.GetEnvString("SMTP_HOST", ""),
		smtpPort:                     env.GetEnvInt("SMTP_PORT", 587),
		smtpUser:                     env.GetEnvString("SMTP_USER", ""),
		smtpPass:                     env.GetEnvString("SMTP_PASS", ""),
		smtpFrom:                     env.GetEnvString("SMTP_FROM", ""),
		smtpStartTLS:                 env.GetEnvBool("SMTP_STARTTLS", true),
	}

	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func SetLastBaseBlockUpdated(blockNumber uint64) {
	cfg.lastBaseBlockUpdated = blockNumber
}

func GetLastBaseBlockUpdated() uint64 {
	return cfg.lastBaseBlockUpdated
}

func GetAttestationCenterAddress() string {
	return cfg.attestationCenterAddress
}

func GetTestAttestationCenterAddress() string {
	return cfg.testAttestationCenterAddress
}

func GetRPCProvider() string {
	return cfg.rpcProvider
}

func GetRPCAPIKey() string {
	return cfg.rpcAPIKey
}

func GetPinataHost() string {
	return cfg.pinataHost
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func GetTaskMonitorRPCPort() string {
	return cfg.taskMonitorRPCPort
}

func GetUpstashRedisUrl() string {
	return cfg.upstashRedisUrl
}

func GetUpstashRedisRestToken() string {
	return cfg.upstashRedisRestToken
}

func GetStreamMaxLen() int {
	return cfg.streamMaxLen
}

func GetJobStreamTTL() time.Duration {
	return cfg.jobStreamTTL
}

func GetTaskStreamTTL() time.Duration {
	return cfg.taskStreamTTL
}

func GetPoolSize() int {
	return cfg.poolSize
}

func GetMinIdleConns() int {
	return cfg.minIdleConns
}

func GetMaxRetries() int {
	return cfg.maxRetries
}

func GetDialTimeout() time.Duration {
	return cfg.dialTimeout
}

func GetReadTimeout() time.Duration {
	return cfg.readTimeout
}

func GetWriteTimeout() time.Duration {
	return cfg.writeTimeout
}

func GetPoolTimeout() time.Duration {
	return cfg.poolTimeout
}

func GetCacheTTL() time.Duration {
	return cfg.cacheTTL
}

func GetCleanupInterval() time.Duration {
	return cfg.cleanupInterval
}

func GetMetricsUpdateInterval() time.Duration {
	return cfg.metricsUpdateInterval
}

func GetRetryDelay() time.Duration {
	return cfg.retryDelay
}

func GetRequestTimeout() time.Duration {
	return cfg.requestTimeout
}

func GetInitializationTimeout() time.Duration {
	return cfg.initializationTimeout
}

func GetMaxRetryBackoff() time.Duration {
	return cfg.maxRetryBackoff
}

func GetOTTempoEndpoint() string {
	return cfg.ottempoEndpoint
}

func GetNotifyWebhookURL() string {
	return cfg.notifyWebhookURL
}

func GetNotifyWebhookToken() string {
	return cfg.notifyWebhookToken
}

func GetSMTPHost() string   { return cfg.smtpHost }
func GetSMTPPort() int      { return cfg.smtpPort }
func GetSMTPUser() string   { return cfg.smtpUser }
func GetSMTPPass() string   { return cfg.smtpPass }
func GetSMTPFrom() string   { return cfg.smtpFrom }
func GetSMTPStartTLS() bool { return cfg.smtpStartTLS }

// GetRedisClientConfig returns a RedisConfig for the new Redis client
func GetRedisClientConfig() redisClient.RedisConfig {
	return redisClient.RedisConfig{
		UpstashConfig: redisClient.UpstashConfig{
			URL:   cfg.upstashRedisUrl,
			Token: cfg.upstashRedisRestToken,
		},
		ConnectionSettings: redisClient.ConnectionSettings{
			PoolSize:         cfg.poolSize,
			MaxIdleConns:     0, // Let Redis client manage this
			MinIdleConns:     cfg.minIdleConns,
			MaxRetries:       cfg.maxRetries,
			DialTimeout:      cfg.dialTimeout,
			ReadTimeout:      cfg.readTimeout,
			WriteTimeout:     cfg.writeTimeout,
			PoolTimeout:      cfg.poolTimeout,
			PingTimeout:      2 * time.Second,  // Default ping timeout
			HealthTimeout:    5 * time.Second,  // Default health check timeout
			OperationTimeout: 10 * time.Second, // Default operation timeout
		},
	}
}

// Get Chain Configs
func GetChainRPCUrl(isRPC bool, chainID string) string {
	var protocol string
	if isRPC {
		protocol = "https://"
	} else {
		protocol = "wss://"
	}
	var domain string
	if cfg.rpcProvider == "alchemy" {
		switch chainID {
		// Testnets
		case "17000":
			domain = "eth-holesky.g.alchemy.com/v2/"
		case "11155111":
			domain = "eth-sepolia.g.alchemy.com/v2/"
		case "11155420":
			domain = "opt-sepolia.g.alchemy.com/v2/"
		case "84532":
			domain = "base-sepolia.g.alchemy.com/v2/"
		case "421614":
			domain = "arb-sepolia.g.alchemy.com/v2/"

		// Mainnets
		case "1":
			domain = "eth-mainnet.g.alchemy.com/v2/"
		case "10":
			domain = "opt-mainnet.g.alchemy.com/v2/"
		case "8453":
			domain = "base-mainnet.g.alchemy.com/v2/"
		case "42161":
			domain = "arb-mainnet.g.alchemy.com/v2/"
		default:
			return ""
		}
	}
	if cfg.rpcProvider == "blast" {
		switch chainID {
		// Testnets
		case "17000":
			domain = "eth-holesky.blastapi.io/"
		case "11155111":
			domain = "eth-sepolia.blastapi.io/"
		case "11155420":
			domain = "optimism-sepolia.blastapi.io/"
		case "84532":
			domain = "base-sepolia.blastapi.io/"
		case "421614":
			domain = "arb-sepolia.blastapi.io/"

		// Mainnets
		case "1":
			domain = "eth-mainnet.blastapi.io/"
		case "10":
			domain = "optimism-mainnet.blastapi.io/"
		case "8453":
			domain = "base-mainnet.blastapi.io/"
		case "42161":
			domain = "arbitrum-one.blastapi.io/"
		default:
			return ""
		}
	}
	return fmt.Sprintf("%s%s%s", protocol, domain, cfg.rpcAPIKey)
}
