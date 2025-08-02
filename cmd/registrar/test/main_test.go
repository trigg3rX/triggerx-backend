package test

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
// 	"github.com/trigg3rX/triggerx-backend/internal/registrar/events"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// )

// func TestMain(t *testing.T) {
// 	// Set working directory to project root
// 	root, err := getProjectRoot()
// 	if err == nil {
// 		_ = os.Chdir(root)
// 	}

// 	// Debug: print current working directory
// 	cwd, _ := os.Getwd()
// 	fmt.Println("[DEBUG] Current working directory:", cwd)

// 	// Set environment variables for testing
// 	if err := os.Setenv("ENV", "development"); err != nil {
// 		t.Fatalf("Failed to set ENV: %v", err)
// 	}
// 	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
// 		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
// 	}
// 	if err := os.Setenv("DEV_MODE", "true"); err != nil {
// 		t.Fatalf("Failed to set DEV_MODE: %v", err)
// 	}
// 	if err := os.Setenv("AVS_GOVERNANCE_ADDRESS", "0x0C77B6273F4852200b17193837960b2f253518FC"); err != nil {
// 		t.Fatalf("Failed to set AVS_GOVERNANCE_ADDRESS: %v", err)
// 	}
// 	if err := os.Setenv("ATTESTATION_CENTER_ADDRESS", "0x710DAb96f318b16F0fC9962D3466C00275414Ff0"); err != nil {
// 		t.Fatalf("Failed to set ATTESTATION_CENTER_ADDRESS: %v", err)
// 	}
// 	if err := os.Setenv("POLLING_INTERVAL", "5m"); err != nil {
// 		t.Fatalf("Failed to set POLLING_INTERVAL: %v", err)
// 	}
// 	if err := os.Setenv("LAST_REWARDS_UPDATE", "2025-05-14T06:31:00Z"); err != nil {
// 		t.Fatalf("Failed to set LAST_REWARDS_UPDATE: %v", err)
// 	}

// 	// Test configuration initialization
// 	t.Run("Config Initialization", func(t *testing.T) {
// 		// Skip config initialization test if .env file is not present
// 		if _, err := os.Stat(".env"); os.IsNotExist(err) {
// 			t.Skip("Skipping config initialization test as .env file is not present")
// 			return
// 		}
// 		err := config.Init()
// 		assert.NoError(t, err, "Config initialization should not fail")
// 	})

// 	logConfig := logging.LoggerConfig{
// 		ProcessName:   "registrar",
// 		IsDevelopment: true,
// 	}
// 	logger, err := logging.NewZapLogger(logConfig)
// 	if err != nil {
// 		t.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	// Test logger initialization
// 	t.Run("Logger Initialization", func(t *testing.T) {
// 		if logger == nil {
// 			t.Fatalf("Logger should not be nil")
// 		}
// 		assert.NoError(t, err, "Logger initialization should not fail")
// 	})

// 	// Test event listener functionality
// 	t.Run("Event Listener", func(t *testing.T) {
// 		// Test default configuration creation
// 		t.Run("Default Config Creation", func(t *testing.T) {
// 			defaultConfig := events.GetDefaultConfig()
// 			assert.NotNil(t, defaultConfig, "Default config should not be nil")
// 			assert.NotEmpty(t, defaultConfig.Chains, "Default config should have chains")
// 			assert.Greater(t, defaultConfig.ProcessingWorkers, 0, "Should have processing workers")
// 			assert.Greater(t, defaultConfig.EventBufferSize, 0, "Should have event buffer size")
// 			logger.Info("Default config created successfully")
// 		})

// 		// Test event listener creation
// 		t.Run("Event Listener Creation", func(t *testing.T) {
// 			defaultConfig := events.GetDefaultConfig()
// 			listener := events.NewContractEventListener(logger, defaultConfig)
// 			assert.NotNil(t, listener, "Event listener should be created successfully")
// 			logger.Info("Event listener created successfully")

// 			// Test getting status without starting
// 			status := listener.GetStatus()
// 			assert.NotNil(t, status, "Status should not be nil")
// 			assert.Contains(t, status, "running", "Status should contain running field")
// 			assert.Contains(t, status, "processing_workers", "Status should contain processing_workers field")
// 			assert.Contains(t, status, "event_buffer_size", "Status should contain event_buffer_size field")
// 			assert.Equal(t, false, status["running"], "Listener should not be running initially")
// 			logger.Info("Event listener status retrieved successfully")
// 		})

// 		// Test configuration validation
// 		t.Run("Configuration Validation", func(t *testing.T) {
// 			defaultConfig := events.GetDefaultConfig()

// 			// Validate chain configurations
// 			assert.NotEmpty(t, defaultConfig.Chains, "Should have chain configurations")
// 			for _, chain := range defaultConfig.Chains {
// 				assert.NotEmpty(t, chain.ChainID, "Chain should have ID")
// 				assert.NotEmpty(t, chain.Name, "Chain should have name")
// 			}

// 			// Validate reconnect config
// 			assert.Greater(t, defaultConfig.ReconnectConfig.MaxRetries, 0, "Should have max retries")
// 			assert.Greater(t, defaultConfig.ReconnectConfig.BaseDelay.Nanoseconds(), int64(0), "Should have base delay")

// 			// Validate contract addresses
// 			assert.NotEmpty(t, defaultConfig.ContractAddresses, "Should have contract addresses")
// 			logger.Info("Configuration validation completed successfully")
// 		})
// 	})
// }

// // getProjectRoot attempts to find the project root by looking for the .git directory or go.mod file
// func getProjectRoot() (string, error) {
// 	dir, err := os.Getwd()
// 	if err != nil {
// 		return "", err
// 	}
// 	for {
// 		if fileExists(filepath.Join(dir, ".git")) || fileExists(filepath.Join(dir, "go.mod")) {
// 			return dir, nil
// 		}
// 		parent := filepath.Dir(dir)
// 		if parent == dir {
// 			break
// 		}
// 		dir = parent
// 	}
// 	return "", os.ErrNotExist
// }

// func fileExists(path string) bool {
// 	_, err := os.Stat(path)
// 	return err == nil
// }
