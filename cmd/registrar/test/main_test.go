package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestMain(t *testing.T) {
	// Set working directory to project root
	root, err := getProjectRoot()
	if err == nil {
		_ = os.Chdir(root)
	}

	// Debug: print current working directory
	cwd, _ := os.Getwd()
	fmt.Println("[DEBUG] Current working directory:", cwd)

	// Debug: check for ABI file existence
	abiPath := filepath.Join("pkg", "bindings", "abi", "AvsGovernance.json")
	if _, err := os.Stat(abiPath); err == nil {
		fmt.Println("[DEBUG] ABI file found at:", abiPath)
	} else {
		fmt.Println("[DEBUG] ABI file NOT found at:", abiPath, "error:", err)
	}

	// Set environment variables for testing
	if err := os.Setenv("ENV", "development"); err != nil {
		t.Fatalf("Failed to set ENV: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("DEV_MODE", "true"); err != nil {
		t.Fatalf("Failed to set DEV_MODE: %v", err)
	}
	if err := os.Setenv("AVS_GOVERNANCE_ADDRESS", "0x0C77B6273F4852200b17193837960b2f253518FC"); err != nil {
		t.Fatalf("Failed to set AVS_GOVERNANCE_ADDRESS: %v", err)
	}
	if err := os.Setenv("ATTESTATION_CENTER_ADDRESS", "0x710DAb96f318b16F0fC9962D3466C00275414Ff0"); err != nil {
		t.Fatalf("Failed to set ATTESTATION_CENTER_ADDRESS: %v", err)
	}
	if err := os.Setenv("POLLING_INTERVAL", "5m"); err != nil {
		t.Fatalf("Failed to set POLLING_INTERVAL: %v", err)
	}
	if err := os.Setenv("LAST_REWARDS_UPDATE", "2025-05-14T06:31:00Z"); err != nil {
		t.Fatalf("Failed to set LAST_REWARDS_UPDATE: %v", err)
	}

	// Test configuration initialization
	t.Run("Config Initialization", func(t *testing.T) {
		// Skip config initialization test if .env file is not present
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			t.Skip("Skipping config initialization test as .env file is not present")
			return
		}
		err := config.Init()
		assert.NoError(t, err, "Config initialization should not fail")
	})

	// Test logger initialization
	t.Run("Logger Initialization", func(t *testing.T) {
		logConfig := logging.LoggerConfig{
			LogDir:          logging.BaseDataDir,
			ProcessName:     "registrar",
			Environment:     logging.Development,
			UseColors:       true,
			MinStdoutLevel:  logging.DebugLevel,
			MinFileLogLevel: logging.DebugLevel,
		}
		err := logging.InitServiceLogger(logConfig)
		assert.NoError(t, err, "Logger initialization should not fail")
	})

	// Test event processor initialization
	t.Run("Event Processor", func(t *testing.T) {
		logger := logging.GetServiceLogger()
		processor := events.NewEventProcessor(logger)
		assert.NotNil(t, processor, "Event processor should be created successfully")
		logger.Info("Event processor created successfully")

		// Test ABI loading
		t.Run("ABI Loading", func(t *testing.T) {
			// Test ABI methods
			methods := events.AvsGovernanceABI.Methods
			assert.NotEmpty(t, methods, "AvsGovernance ABI should have methods")

			methods = events.AttestationCenterABI.Methods
			assert.NotEmpty(t, methods, "AttestationCenter ABI should have methods")
		})
	})

	// Test environment and log level
	t.Run("Environment and Log Level", func(t *testing.T) {
		env := getEnvironment()
		assert.Contains(t, []logging.LogLevel{logging.Development, logging.Production}, env, "Environment should be either Development or Production")

		level := getLogLevel()
		assert.Contains(t, []logging.Level{logging.DebugLevel, logging.InfoLevel}, level, "Log level should be either Debug or Info")
	})
}

// getProjectRoot attempts to find the project root by looking for the .git directory or go.mod file
func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(dir, ".git")) || fileExists(filepath.Join(dir, "go.mod")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getEnvironment() logging.LogLevel {
	if config.IsDevMode() {
		return logging.Development
	}
	return logging.Production
}

func getLogLevel() logging.Level {
	if config.IsDevMode() {
		return logging.DebugLevel
	}
	return logging.InfoLevel
}
