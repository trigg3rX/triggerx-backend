package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

type Config struct {
	devMode bool

	// Port at which health service will be running
	healthRPCPort string

	// Bot token for Telegram notifications
	botToken string
	// Email user for notifications
	emailUser     string
	emailPassword string

	// ScyllaDB Host and Port
	databaseHostAddress string
	databaseHostPort    string

	// ScyllaDB Authentication
	databaseUsername string
	databasePassword string

	// ScyllaDB SSL/TLS Configuration
	databaseSSLEnabled        bool
	databaseSSLCertPath       string
	databaseSSLKeyPath        string
	databaseSSLCAPath         string
	databaseSSLInsecureSkipVerify bool

	// OTel exporter endpoint
	otelExporterEndpoint string

	// IPFS configuration
	pinataHost string
	pinataJWT  string

	// Manager Signing Address
	managerSigningAddress string

	// Etherscan API Key
	etherscanAPIKey string

	// Alchemy API Key
	alchemyAPIKey string

	// Task Execution Address
	taskExecutionAddress     string
	testTaskExecutionAddress string
	imuaTaskExecutionAddress string
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                  env.GetEnvBool("DEV_MODE", false),
		healthRPCPort:            env.GetEnvString("HEALTH_RPC_PORT", "9003"),
		botToken:                 env.GetEnvString("BOT_TOKEN", ""),
		emailUser:                env.GetEnvString("EMAIL_USER", ""),
		emailPassword:            env.GetEnvString("EMAIL_PASS", ""),
		databaseHostAddress:      env.GetEnvString("DATABASE_HOST_ADDRESS", "localhost"),
		databaseHostPort:         env.GetEnvString("DATABASE_HOST_PORT", "9042"),
		databaseUsername:         env.GetEnvString("DATABASE_USERNAME", ""),
		databasePassword:         env.GetEnvString("DATABASE_PASSWORD", ""),
		databaseSSLEnabled:       env.GetEnvBool("DATABASE_SSL_ENABLED", false),
		databaseSSLCertPath:      env.GetEnvString("DATABASE_SSL_CERT_PATH", ""),
		databaseSSLKeyPath:       env.GetEnvString("DATABASE_SSL_KEY_PATH", ""),
		databaseSSLCAPath:       env.GetEnvString("DATABASE_SSL_CA_PATH", ""),
		databaseSSLInsecureSkipVerify: env.GetEnvBool("DATABASE_SSL_INSECURE_SKIP_VERIFY", false),
		otelExporterEndpoint:     env.GetEnvString("OTEL_EXPORTER_ENDPOINT", "localhost:4318"),
		pinataHost:               env.GetEnvString("PINATA_HOST", ""),
		pinataJWT:                env.GetEnvString("PINATA_JWT", ""),
		managerSigningAddress:    env.GetEnvString("MANAGER_SIGNING_ADDRESS", ""),
		etherscanAPIKey:          env.GetEnvString("ETHERSCAN_API_KEY", ""),
		alchemyAPIKey:            env.GetEnvString("ALCHEMY_API_KEY", ""),
		taskExecutionAddress:     env.GetEnvString("TASK_EXECUTION_ADDRESS", ""),
		testTaskExecutionAddress: env.GetEnvString("TEST_TASK_EXECUTION_ADDRESS", ""),
		imuaTaskExecutionAddress: env.GetEnvString("IMUA_TASK_EXECUTION_ADDRESS", ""),
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
	if !env.IsValidPort(cfg.healthRPCPort) {
		return fmt.Errorf("invalid Health RPC Port: %s", cfg.healthRPCPort)
	}
	if env.IsEmpty(cfg.pinataHost) {
		return fmt.Errorf("invalid Pinata Host: %s", cfg.pinataHost)
	}
	if env.IsEmpty(cfg.pinataJWT) {
		return fmt.Errorf("invalid Pinata JWT: %s", cfg.pinataJWT)
	}
	if env.IsEmpty(cfg.etherscanAPIKey) {
		return fmt.Errorf("invalid Etherscan API Key: %s", cfg.etherscanAPIKey)
	}
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("invalid Alchemy API Key: %s", cfg.alchemyAPIKey)
	}
	if !env.IsValidIPAddress(cfg.databaseHostAddress) {
		return fmt.Errorf("invalid database host address: %s", cfg.databaseHostAddress)
	}
	if !env.IsValidPort(cfg.databaseHostPort) {
		return fmt.Errorf("invalid database host port: %s", cfg.databaseHostPort)
	}
	if !env.IsValidHostPort(cfg.otelExporterEndpoint) {
		return fmt.Errorf("invalid OTEL exporter endpoint: %s (must be a valid host:port, e.g., localhost:4318)", cfg.otelExporterEndpoint)
	}
	if !env.IsValidEthAddress(cfg.taskExecutionAddress) {
		return fmt.Errorf("invalid task execution address: %s", cfg.taskExecutionAddress)
	}
	if !env.IsValidEthAddress(cfg.testTaskExecutionAddress) {
		return fmt.Errorf("invalid test task execution address: %s", cfg.testTaskExecutionAddress)
	}
	if !env.IsValidEthAddress(cfg.imuaTaskExecutionAddress) {
		return fmt.Errorf("invalid Imua task execution address: %s", cfg.imuaTaskExecutionAddress)
	}
	if !cfg.devMode {
		if !env.IsValidEmail(cfg.emailUser) {
			return fmt.Errorf("invalid email user: %s", cfg.emailUser)
		}
		if env.IsEmpty(cfg.emailPassword) {
			return fmt.Errorf("invalid email password: %s", cfg.emailPassword)
		}
		if env.IsEmpty(cfg.botToken) {
			return fmt.Errorf("invalid bot token: %s", cfg.botToken)
		}
	}
	return nil
}

func GetHealthRPCPort() string {
	return cfg.healthRPCPort
}

func GetDatabaseHostAddress() string {
	return cfg.databaseHostAddress
}

func GetDatabaseHostPort() string {
	return cfg.databaseHostPort
}

func GetOTELExporterEndpoint() string {
	return cfg.otelExporterEndpoint
}

func GetBotToken() string {
	return cfg.botToken
}

func GetEmailUser() string {
	return cfg.emailUser
}

func GetEmailPassword() string {
	return cfg.emailPassword
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetPinataHost() string {
	return cfg.pinataHost
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func GetEtherscanAPIKey() string {
	return cfg.etherscanAPIKey
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

func GetTaskExecutionAddress() string {
	return cfg.taskExecutionAddress
}

func GetTestTaskExecutionAddress() string {
	return cfg.testTaskExecutionAddress
}

func GetImuaTaskExecutionAddress() string {
	return cfg.imuaTaskExecutionAddress
}

func GetManagerSigningAddress() string {
	return cfg.managerSigningAddress
}

// GetVersion returns the service version
// TODO: This should be set from build flags or environment variable
func GetVersion() string {
	return "1.0.0"
}

func GetDatabaseUsername() string {
	return cfg.databaseUsername
}

func GetDatabasePassword() string {
	return cfg.databasePassword
}

func GetDatabaseSSLEnabled() bool {
	return cfg.databaseSSLEnabled
}

func GetDatabaseSSLCertPath() string {
	return cfg.databaseSSLCertPath
}

func GetDatabaseSSLKeyPath() string {
	return cfg.databaseSSLKeyPath
}

func GetDatabaseSSLCAPath() string {
	return cfg.databaseSSLCAPath
}

func GetDatabaseSSLInsecureSkipVerify() bool {
	return cfg.databaseSSLInsecureSkipVerify
}
