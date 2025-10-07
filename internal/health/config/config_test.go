package config

import (
	"os"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	// Create a temporary YAML config file
	tempFile := createTempConfig(t, `
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "70s"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 3
  retry_backoff: "1s"

database:
  host: "localhost"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
  connection_timeout: "10s"
  query_timeout: "30s"

notification:
  smtp_host: "smtp.gmail.com"
  smtp_port: "587"
  timeout: "30s"
  retry_attempts: 3
`)
	defer os.Remove(tempFile)

	// Create test env file (contains all required env vars including DEV_MODE=true)
	envFile := createTestEnvFile(t)
	defer os.Remove(envFile)

	// Clear any existing env vars to ensure clean state
	defer clearEnvVars()

	// Test initialization
	err := InitWithEnvFile(tempFile, envFile)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify configuration was loaded
	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("Config is nil")
	}

	// Verify server config
	if cfg.Server.HTTPPort != "8080" {
		t.Errorf("Expected HTTPPort 8080, got %s", cfg.Server.HTTPPort)
	}
	if cfg.Server.GRPCPort != "9090" {
		t.Errorf("Expected GRPCPort 9090, got %s", cfg.Server.GRPCPort)
	}

	// Verify health check config
	if cfg.HealthCheck.KeeperTimeout != "70s" {
		t.Errorf("Expected KeeperTimeout 70s, got %s", cfg.HealthCheck.KeeperTimeout)
	}
	if cfg.HealthCheck.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", cfg.HealthCheck.MaxRetries)
	}

	// Verify database config
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected Database.Host localhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Keyspace != "triggerx_health" {
		t.Errorf("Expected Database.Keyspace triggerx_health, got %s", cfg.Database.Keyspace)
	}

	// Verify secrets were loaded
	if cfg.Secrets.BotToken != "test_bot_token" {
		t.Errorf("Expected BotToken test_bot_token, got %s", cfg.Secrets.BotToken)
	}
	if cfg.Secrets.EmailUser != "test@example.com" {
		t.Errorf("Expected EmailUser test@example.com, got %s", cfg.Secrets.EmailUser)
	}
	// DevMode is true because the env file created by createTestEnvFile sets DEV_MODE=true
	if cfg.DevMode != true {
		t.Errorf("Expected DevMode true, got %t", cfg.DevMode)
	}
}

func TestInit_DefaultConfigPath(t *testing.T) {
	// Create default config file
	defaultFile := "config/health.yaml"

	// Ensure directory exists
	err := os.MkdirAll("config", 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	defer os.RemoveAll("config")

	// Create default config file
	err = os.WriteFile(defaultFile, []byte(`
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "70s"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 3
  retry_backoff: "1s"

database:
  host: "localhost"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
  connection_timeout: "10s"
  query_timeout: "30s"

notification:
  smtp_host: "smtp.gmail.com"
  smtp_port: "587"
  timeout: "30s"
  retry_attempts: 3
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create default config file: %v", err)
	}

	// Set required environment variables
	setTestEnvVars()
	defer clearEnvVars()

	// Create test env file
	envFile := createTestEnvFile(t)
	defer os.Remove(envFile)

	// Test initialization with empty config path
	err = InitWithEnvFile("", envFile)
	if err != nil {
		t.Fatalf("Init with default path failed: %v", err)
	}

	// Verify configuration was loaded
	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("Config is nil")
	}
}

func TestInit_InvalidYAML(t *testing.T) {
	// Create invalid YAML file
	tempFile := createTempConfig(t, `
invalid: yaml: content
  - this is broken
`)
	defer os.Remove(tempFile)

	setTestEnvVars()
	defer clearEnvVars()

	err := Init(tempFile)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestInit_MissingRequiredFields(t *testing.T) {
	// Create config missing required fields
	tempFile := createTempConfig(t, `
server:
  http_port: "8080"
  # Missing grpc_port

health_check:
  keeper_timeout: "70s"
  # Missing other required fields

database:
  host: "localhost"
  # Missing other required fields
`)
	defer os.Remove(tempFile)

	setTestEnvVars()
	defer clearEnvVars()

	err := Init(tempFile)
	if err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestInit_MissingSecrets(t *testing.T) {
	// Create valid YAML config
	tempFile := createTempConfig(t, `
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "70s"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 3
  retry_backoff: "1s"

database:
  host: "localhost"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
  connection_timeout: "10s"
  query_timeout: "30s"

notification:
  smtp_host: "smtp.gmail.com"
  smtp_port: "587"
  timeout: "30s"
  retry_attempts: 3
`)
	defer os.Remove(tempFile)

	// Create env file with DEV_MODE=false and no secrets
	envContent := `DEV_MODE=false
`
	envFile, err := os.CreateTemp("", "test_missing_secrets_*.env")
	if err != nil {
		t.Fatalf("Failed to create temp env file: %v", err)
	}
	_, err = envFile.WriteString(envContent)
	if err != nil {
		t.Fatalf("Failed to write to temp env file: %v", err)
	}
	envFile.Close()
	defer os.Remove(envFile.Name())

	// Don't set environment variables - should fail validation when DEV_MODE=false
	err = InitWithEnvFile(tempFile, envFile.Name())
	if err == nil {
		t.Error("Expected error for missing secrets when DEV_MODE=false")
	}
}

func TestInit_DevMode(t *testing.T) {
	// Create config with dev mode enabled
	tempFile := createTempConfig(t, `
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "70s"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 3
  retry_backoff: "1s"

database:
  host: "localhost"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
  connection_timeout: "10s"
  query_timeout: "30s"

notification:
  smtp_host: "smtp.gmail.com"
  smtp_port: "587"
  timeout: "30s"
  retry_attempts: 3
`)
	defer os.Remove(tempFile)

	// Set dev mode environment variable
	os.Setenv("DEV_MODE", "true")
	defer os.Unsetenv("DEV_MODE")

	// Create test env file
	envFile := createTestEnvFile(t)
	defer os.Remove(envFile)

	// Don't set other environment variables - should pass in dev mode
	err := InitWithEnvFile(tempFile, envFile)
	if err != nil {
		t.Fatalf("Init failed in dev mode: %v", err)
	}

	// Verify dev mode is set (loaded from env file)
	if !IsDevMode() {
		t.Error("Expected dev mode to be enabled")
	}
}

func TestGetKeeperTimeout(t *testing.T) {
	setupTestConfig(t)

	timeout := GetKeeperTimeout()
	expected := 70 * time.Second
	if timeout != expected {
		t.Errorf("Expected %v, got %v", expected, timeout)
	}
}

func TestGetCheckInterval(t *testing.T) {
	setupTestConfig(t)

	interval := GetCheckInterval()
	expected := 30 * time.Second
	if interval != expected {
		t.Errorf("Expected %v, got %v", expected, interval)
	}
}

func TestGetSyncInterval(t *testing.T) {
	setupTestConfig(t)

	interval := GetSyncInterval()
	expected := 5 * time.Minute
	if interval != expected {
		t.Errorf("Expected %v, got %v", expected, interval)
	}
}

func TestGetMaxRetries(t *testing.T) {
	setupTestConfig(t)

	retries := GetMaxRetries()
	expected := 3
	if retries != expected {
		t.Errorf("Expected %d, got %d", expected, retries)
	}
}

func TestGetRetryBackoff(t *testing.T) {
	setupTestConfig(t)

	backoff := GetRetryBackoff()
	expected := 1 * time.Second
	if backoff != expected {
		t.Errorf("Expected %v, got %v", expected, backoff)
	}
}

func TestGetDatabaseConfig(t *testing.T) {
	setupTestConfig(t)

	// Test database getters
	if GetDatabaseHost() != "localhost" {
		t.Errorf("Expected localhost, got %s", GetDatabaseHost())
	}
	if GetDatabasePort() != "9042" {
		t.Errorf("Expected 9042, got %s", GetDatabasePort())
	}
	if GetDatabaseKeyspace() != "triggerx_health" {
		t.Errorf("Expected triggerx_health, got %s", GetDatabaseKeyspace())
	}
	if GetDatabaseReplicationFactor() != 3 {
		t.Errorf("Expected 3, got %d", GetDatabaseReplicationFactor())
	}
	if GetDatabaseConsistencyLevel() != "quorum" {
		t.Errorf("Expected quorum, got %s", GetDatabaseConsistencyLevel())
	}
}

func TestGetNotificationConfig(t *testing.T) {
	setupTestConfig(t)

	// Test notification getters
	if GetSMTPHost() != "smtp.gmail.com" {
		t.Errorf("Expected smtp.gmail.com, got %s", GetSMTPHost())
	}
	if GetSMTPPort() != "587" {
		t.Errorf("Expected 587, got %s", GetSMTPPort())
	}
	if GetNotificationTimeout() != 30*time.Second {
		t.Errorf("Expected 30s, got %v", GetNotificationTimeout())
	}
	if GetNotificationRetryAttempts() != 3 {
		t.Errorf("Expected 3, got %d", GetNotificationRetryAttempts())
	}
}

func TestGetSecrets(t *testing.T) {
	setupTestConfig(t)

	// Test secret getters
	if GetBotToken() != "test_bot_token" {
		t.Errorf("Expected test_bot_token, got %s", GetBotToken())
	}
	if GetEmailUser() != "test@example.com" {
		t.Errorf("Expected test@example.com, got %s", GetEmailUser())
	}
	if GetEmailPassword() != "test_password" {
		t.Errorf("Expected test_password, got %s", GetEmailPassword())
	}
	if GetEtherscanAPIKey() != "test_etherscan_key" {
		t.Errorf("Expected test_etherscan_key, got %s", GetEtherscanAPIKey())
	}
	if GetAlchemyAPIKey() != "test_alchemy_key" {
		t.Errorf("Expected test_alchemy_key, got %s", GetAlchemyAPIKey())
	}
	if GetPinataHost() != "test.pinata.cloud" {
		t.Errorf("Expected test.pinata.cloud, got %s", GetPinataHost())
	}
	if GetPinataJWT() != "test_pinata_jwt" {
		t.Errorf("Expected test_pinata_jwt, got %s", GetPinataJWT())
	}
}

func TestValidateSecrets(t *testing.T) {
	tests := []struct {
		name    string
		secrets SecretsConfig
		wantErr bool
	}{
		{
			name: "valid secrets",
			secrets: SecretsConfig{
				BotToken:                 "test_token",
				EmailUser:                "test@example.com",
				EmailPassword:            "test_password",
				EtherscanAPIKey:          "test_key",
				AlchemyAPIKey:            "test_key",
				PinataHost:               "test.pinata.cloud",
				PinataJWT:                "test_jwt",
				ManagerSigningAddress:    "0x1234567890123456789012345678901234567890",
				TaskExecutionAddress:     "0x2345678901234567890123456789012345678901",
				TestTaskExecutionAddress: "0x3456789012345678901234567890123456789012",
				ImuaTaskExecutionAddress: "0x4567890123456789012345678901234567890123",
				JWTPrivateKey:            "test_private_key",
			},
			wantErr: false,
		},
		{
			name: "missing bot token",
			secrets: SecretsConfig{
				EmailUser:                "test@example.com",
				EmailPassword:            "test_password",
				EtherscanAPIKey:          "test_key",
				AlchemyAPIKey:            "test_key",
				PinataHost:               "test.pinata.cloud",
				PinataJWT:                "test_jwt",
				ManagerSigningAddress:    "0x1234567890123456789012345678901234567890",
				TaskExecutionAddress:     "0x2345678901234567890123456789012345678901",
				TestTaskExecutionAddress: "0x3456789012345678901234567890123456789012",
				ImuaTaskExecutionAddress: "0x4567890123456789012345678901234567890123",
				JWTPrivateKey:            "test_private_key",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			secrets: SecretsConfig{
				BotToken:                 "test_token",
				EmailUser:                "invalid-email",
				EmailPassword:            "test_password",
				EtherscanAPIKey:          "test_key",
				AlchemyAPIKey:            "test_key",
				PinataHost:               "test.pinata.cloud",
				PinataJWT:                "test_jwt",
				ManagerSigningAddress:    "0x1234567890123456789012345678901234567890",
				TaskExecutionAddress:     "0x2345678901234567890123456789012345678901",
				TestTaskExecutionAddress: "0x3456789012345678901234567890123456789012",
				ImuaTaskExecutionAddress: "0x4567890123456789012345678901234567890123",
				JWTPrivateKey:            "test_private_key",
			},
			wantErr: true,
		},
		{
			name: "invalid ethereum address",
			secrets: SecretsConfig{
				BotToken:                 "test_token",
				EmailUser:                "test@example.com",
				EmailPassword:            "test_password",
				EtherscanAPIKey:          "test_key",
				AlchemyAPIKey:            "test_key",
				PinataHost:               "test.pinata.cloud",
				PinataJWT:                "test_jwt",
				ManagerSigningAddress:    "invalid_address",
				TaskExecutionAddress:     "0x2345678901234567890123456789012345678901",
				TestTaskExecutionAddress: "0x3456789012345678901234567890123456789012",
				ImuaTaskExecutionAddress: "0x4567890123456789012345678901234567890123",
				JWTPrivateKey:            "test_private_key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecrets(tt.secrets)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSecrets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions
func createTempConfig(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return file.Name()
}

func createTestEnvFile(t *testing.T) string {
	// Create a temporary .env file
	envContent := `BOT_TOKEN=test_bot_token
EMAIL_USER=test@example.com
EMAIL_PASS=test_password
ETHERSCAN_API_KEY=test_etherscan_key
ALCHEMY_API_KEY=test_alchemy_key
PINATA_HOST=test.pinata.cloud
PINATA_JWT=test_pinata_jwt
MANAGER_SIGNING_ADDRESS=0x1234567890123456789012345678901234567890
TASK_EXECUTION_ADDRESS=0x2345678901234567890123456789012345678901
TEST_TASK_EXECUTION_ADDRESS=0x3456789012345678901234567890123456789012
IMUA_TASK_EXECUTION_ADDRESS=0x4567890123456789012345678901234567890123
JWT_PRIVATE_KEY=test_jwt_private_key
DEV_MODE=true
`

	envFile, err := os.CreateTemp("", "test_*.env")
	if err != nil {
		t.Fatalf("Failed to create temp env file: %v", err)
	}

	_, err = envFile.WriteString(envContent)
	if err != nil {
		t.Fatalf("Failed to write to temp env file: %v", err)
	}

	err = envFile.Close()
	if err != nil {
		t.Fatalf("Failed to close temp env file: %v", err)
	}

	return envFile.Name()
}

func setTestEnvVars() {
	os.Setenv("BOT_TOKEN", "test_bot_token")
	os.Setenv("EMAIL_USER", "test@example.com")
	os.Setenv("EMAIL_PASS", "test_password")
	os.Setenv("ETHERSCAN_API_KEY", "test_etherscan_key")
	os.Setenv("ALCHEMY_API_KEY", "test_alchemy_key")
	os.Setenv("PINATA_HOST", "test.pinata.cloud")
	os.Setenv("PINATA_JWT", "test_pinata_jwt")
	os.Setenv("MANAGER_SIGNING_ADDRESS", "0x1234567890123456789012345678901234567890")
	os.Setenv("TASK_EXECUTION_ADDRESS", "0x2345678901234567890123456789012345678901")
	os.Setenv("TEST_TASK_EXECUTION_ADDRESS", "0x3456789012345678901234567890123456789012")
	os.Setenv("IMUA_TASK_EXECUTION_ADDRESS", "0x4567890123456789012345678901234567890123")
	os.Setenv("JWT_PRIVATE_KEY", "test_jwt_private_key")
	os.Setenv("DEV_MODE", "false")
}

func clearEnvVars() {
	os.Unsetenv("BOT_TOKEN")
	os.Unsetenv("EMAIL_USER")
	os.Unsetenv("EMAIL_PASS")
	os.Unsetenv("ETHERSCAN_API_KEY")
	os.Unsetenv("ALCHEMY_API_KEY")
	os.Unsetenv("PINATA_HOST")
	os.Unsetenv("PINATA_JWT")
	os.Unsetenv("MANAGER_SIGNING_ADDRESS")
	os.Unsetenv("TASK_EXECUTION_ADDRESS")
	os.Unsetenv("TEST_TASK_EXECUTION_ADDRESS")
	os.Unsetenv("IMUA_TASK_EXECUTION_ADDRESS")
	os.Unsetenv("JWT_PRIVATE_KEY")
	os.Unsetenv("DEV_MODE")
}

func setupTestConfig(t *testing.T) {
	// Create a temporary YAML config file
	tempFile := createTempConfig(t, `
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "70s"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 3
  retry_backoff: "1s"

database:
  host: "localhost"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
  connection_timeout: "10s"
  query_timeout: "30s"

notification:
  smtp_host: "smtp.gmail.com"
  smtp_port: "587"
  timeout: "30s"
  retry_attempts: 3
`)
	defer os.Remove(tempFile)

	// Create test env file
	envFile := createTestEnvFile(t)
	defer os.Remove(envFile)

	// Initialize config
	err := InitWithEnvFile(tempFile, envFile)
	if err != nil {
		t.Fatalf("Setup test config failed: %v", err)
	}
}

func TestGetConfig(t *testing.T) {
	setupTestConfig(t)

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}

	// Verify it's the same config we set up
	if cfg.Server.HTTPPort != "8080" {
		t.Errorf("Expected HTTPPort 8080, got %s", cfg.Server.HTTPPort)
	}
}

func TestGetHTTPPort(t *testing.T) {
	setupTestConfig(t)

	port := GetHTTPPort()
	if port != "8080" {
		t.Errorf("Expected 8080, got %s", port)
	}
}

func TestGetGRPCPort(t *testing.T) {
	setupTestConfig(t)

	port := GetGRPCPort()
	if port != "9090" {
		t.Errorf("Expected 9090, got %s", port)
	}
}

func TestIsDevMode(t *testing.T) {
	setupTestConfig(t)

	// setupTestConfig uses createTestEnvFile which sets DEV_MODE=true
	if !IsDevMode() {
		t.Error("Expected dev mode to be true (set by createTestEnvFile)")
	}
}
