package yaml

import (
	"os"
	"path/filepath"
	"testing"
)

// IntegrationTestStruct represents a complex struct for integration testing
type IntegrationTestStruct struct {
	Server       ServerConfig       `yaml:"server" validate:"required"`
	HealthCheck  HealthCheckConfig  `yaml:"health_check" validate:"required"`
	Database     DatabaseConfig     `yaml:"database" validate:"required"`
	Notification NotificationConfig `yaml:"notification"`
}

type ServerConfig struct {
	HTTPPort string `yaml:"http_port" validate:"required,port"`
	GRPCPort string `yaml:"grpc_port" validate:"required,port"`
}

type HealthCheckConfig struct {
	KeeperTimeout string `yaml:"keeper_timeout" validate:"required,duration"`
	CheckInterval string `yaml:"check_interval" validate:"required,duration"`
	SyncInterval  string `yaml:"sync_interval" validate:"required,duration"`
	MaxRetries    int    `yaml:"max_retries" validate:"min=1,max=10"`
	RetryBackoff  string `yaml:"retry_backoff" validate:"duration"`
}

type DatabaseConfig struct {
	Host              string `yaml:"host" validate:"required,ip"`
	Port              string `yaml:"port" validate:"required,port"`
	Keyspace          string `yaml:"keyspace" validate:"required,min=1"`
	ReplicationFactor int    `yaml:"replication_factor" validate:"min=1,max=10"`
	ConsistencyLevel  string `yaml:"consistency_level" validate:"oneof=one|two|three|quorum|all|local_quorum|each_quorum|local_one"`
	ConnectionTimeout string `yaml:"connection_timeout" validate:"duration"`
	QueryTimeout      string `yaml:"query_timeout" validate:"duration"`
}

type NotificationConfig struct {
	SMTPHost      string `yaml:"smtp_host" validate:"ip"`
	SMTPPort      string `yaml:"smtp_port" validate:"port"`
	Timeout       string `yaml:"timeout" validate:"duration"`
	RetryAttempts int    `yaml:"retry_attempts" validate:"min=1,max=10"`
}

func TestIntegration_LoadAndValidate(t *testing.T) {
	// Create a comprehensive YAML config
	yamlContent := `
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "70s"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 3
  retry_backoff: "1s"
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
`

	tempFile := createTempYAML(t, yamlContent)
	defer func() {
		err := os.Remove(tempFile)
		if err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()
	

	// Load configuration
	var config IntegrationTestStruct
	err := LoadYAML(tempFile, &config)
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}

	// Validate configuration
	err = ValidateConfig(&config)
	if err != nil {
		t.Fatalf("ValidateHealthConfig failed: %v", err)
	}

	// Verify loaded values
	if config.Server.HTTPPort != "8080" {
		t.Errorf("Expected HTTPPort 8080, got %s", config.Server.HTTPPort)
	}
	if config.Server.GRPCPort != "9090" {
		t.Errorf("Expected GRPCPort 9090, got %s", config.Server.GRPCPort)
	}
	if config.HealthCheck.KeeperTimeout != "70s" {
		t.Errorf("Expected KeeperTimeout 70s, got %s", config.HealthCheck.KeeperTimeout)
	}
	if config.HealthCheck.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.HealthCheck.MaxRetries)
	}
	if config.Database.Host != "localhost" {
		t.Errorf("Expected Database.Host localhost, got %s", config.Database.Host)
	}
	if config.Database.Keyspace != "triggerx_health" {
		t.Errorf("Expected Database.Keyspace triggerx_health, got %s", config.Database.Keyspace)
	}
	if config.Notification.SMTPHost != "smtp.gmail.com" {
		t.Errorf("Expected Notification.SMTPHost smtp.gmail.com, got %s", config.Notification.SMTPHost)
	}
}

func TestIntegration_LoadAndValidateWithErrors(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
	}{
		{
			name: "invalid port",
			yamlContent: `
server:
  http_port: "99999"  # Invalid port
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
`,
			expectError: true,
		},
		{
			name: "invalid duration",
			yamlContent: `
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "invalid_duration"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 3

database:
  host: "localhost"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
`,
			expectError: true,
		},
		{
			name: "invalid IP address",
			yamlContent: `
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
  host: "invalid.ip.address"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
`,
			expectError: true,
		},
		{
			name: "invalid consistency level",
			yamlContent: `
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
  consistency_level: "invalid_level"
`,
			expectError: true,
		},
		{
			name: "max retries out of range",
			yamlContent: `
server:
  http_port: "8080"
  grpc_port: "9090"

health_check:
  keeper_timeout: "70s"
  check_interval: "30s"
  sync_interval: "5m"
  max_retries: 15  # Out of range

database:
  host: "localhost"
  port: "9042"
  keyspace: "triggerx_health"
  replication_factor: 3
  consistency_level: "quorum"
`,
			expectError: true,
		},
		{
			name: "missing required field",
			yamlContent: `
server:
  http_port: "8080"
  # Missing grpc_port

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
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := createTempYAMLIntegration(t, tt.yamlContent)
			defer func() {
				err := os.Remove(tempFile)
				if err != nil {
					t.Errorf("Failed to remove temp file: %v", err)
				}
			}()

			// Load configuration
			var config IntegrationTestStruct
			err := LoadYAML(tempFile, &config)
			if err != nil {
				t.Fatalf("LoadYAML failed: %v", err)
			}

			// Validate configuration
			err = ValidateConfig(&config)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateHealthConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestIntegration_EnvironmentSpecificConfig(t *testing.T) {
	// Create base config
	baseContent := `
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
`

	// Create environment-specific overrides
	envContent := `
server:
  http_port: "8081"  # Override HTTP port

health_check:
  max_retries: 5  # Override max retries

database:
  host: "127.0.0.1"  # Override database host with valid IP
  consistency_level: "all"     # Override consistency level
  connection_timeout: "15s"    # Override connection timeout
  query_timeout: "45s"         # Override query timeout
`

	// Create temporary files
	baseFile := createTempYAMLIntegration(t, baseContent)
	defer func() {
		err := os.Remove(baseFile)
		if err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	// Create environment file in same directory
	baseDir := filepath.Dir(baseFile)
	baseName := filepath.Base(baseFile)
	baseNameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]
	envFile := filepath.Join(baseDir, baseNameWithoutExt+".prod.yaml")

	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create environment file: %v", err)
	}
	defer func() {
		err := os.Remove(envFile)
		if err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	// Load with environment-specific overrides
	var config IntegrationTestStruct
	err = LoadEnvironmentSpecificYAML(baseFile, &config, "prod")
	if err != nil {
		t.Fatalf("LoadEnvironmentSpecificYAML failed: %v", err)
	}

	// Validate configuration
	err = ValidateConfig(&config)
	if err != nil {
		t.Fatalf("ValidateHealthConfig failed: %v", err)
	}

	// Verify overrides were applied
	if config.Server.HTTPPort != "8081" {
		t.Errorf("Expected overridden HTTPPort 8081, got %s", config.Server.HTTPPort)
	}
	if config.HealthCheck.MaxRetries != 5 {
		t.Errorf("Expected overridden MaxRetries 5, got %d", config.HealthCheck.MaxRetries)
	}
	if config.Database.Host != "127.0.0.1" {
		t.Errorf("Expected overridden Database.Host 127.0.0.1, got %s", config.Database.Host)
	}
	if config.Database.ConsistencyLevel != "all" {
		t.Errorf("Expected overridden ConsistencyLevel all, got %s", config.Database.ConsistencyLevel)
	}

	// Verify non-overridden values remain from base
	if config.Server.GRPCPort != "9090" {
		t.Errorf("Expected base GRPCPort 9090, got %s", config.Server.GRPCPort)
	}
	if config.HealthCheck.KeeperTimeout != "70s" {
		t.Errorf("Expected base KeeperTimeout 70s, got %s", config.HealthCheck.KeeperTimeout)
	}
}

func TestIntegration_SaveAndLoad(t *testing.T) {
	// Original config
	originalConfig := IntegrationTestStruct{
		Server: ServerConfig{
			HTTPPort: "8080",
			GRPCPort: "9090",
		},
		HealthCheck: HealthCheckConfig{
			KeeperTimeout: "70s",
			CheckInterval: "30s",
			SyncInterval:  "5m",
			MaxRetries:    3,
			RetryBackoff:  "1s",
		},
		Database: DatabaseConfig{
			Host:              "localhost",
			Port:              "9042",
			Keyspace:          "triggerx_health",
			ReplicationFactor: 3,
			ConsistencyLevel:  "quorum",
			ConnectionTimeout: "10s",
			QueryTimeout:      "30s",
		},
		Notification: NotificationConfig{
			SMTPHost:      "smtp.gmail.com",
			SMTPPort:      "587",
			Timeout:       "30s",
			RetryAttempts: 3,
		},
	}

	// Save config to file
	tempFile := filepath.Join(t.TempDir(), "integration_test.yaml")
	err := SaveYAML(tempFile, originalConfig)
	if err != nil {
		t.Fatalf("SaveYAML failed: %v", err)
	}

	// Load config from file
	var loadedConfig IntegrationTestStruct
	err = LoadYAML(tempFile, &loadedConfig)
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}

	// Validate loaded config
	err = ValidateConfig(&loadedConfig)
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}

	// Verify configs match
	if loadedConfig != originalConfig {
		t.Errorf("Loaded config doesn't match original:\nOriginal: %+v\nLoaded: %+v", originalConfig, loadedConfig)
	}
}

// Helper function to create temporary YAML file for integration tests
func createTempYAMLIntegration(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "integration_test_*.yaml")
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
