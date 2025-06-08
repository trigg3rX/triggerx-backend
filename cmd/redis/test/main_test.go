package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestMain(t *testing.T) {
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
	if err := os.Setenv("REDIS_ADDR", "localhost:6379"); err != nil {
		t.Fatalf("Failed to set REDIS_ADDR: %v", err)
	}
	if err := os.Setenv("REDIS_PASSWORD", ""); err != nil {
		t.Fatalf("Failed to set REDIS_PASSWORD: %v", err)
	}
	if err := os.Setenv("REDIS_POOL_SIZE", "10"); err != nil {
		t.Fatalf("Failed to set REDIS_POOL_SIZE: %v", err)
	}
	if err := os.Setenv("REDIS_MIN_IDLE_CONNS", "2"); err != nil {
		t.Fatalf("Failed to set REDIS_MIN_IDLE_CONNS: %v", err)
	}
	if err := os.Setenv("REDIS_MAX_RETRIES", "3"); err != nil {
		t.Fatalf("Failed to set REDIS_MAX_RETRIES: %v", err)
	}
	if err := os.Setenv("REDIS_DIAL_TIMEOUT_SEC", "5"); err != nil {
		t.Fatalf("Failed to set REDIS_DIAL_TIMEOUT_SEC: %v", err)
	}
	if err := os.Setenv("REDIS_READ_TIMEOUT_SEC", "3"); err != nil {
		t.Fatalf("Failed to set REDIS_READ_TIMEOUT_SEC: %v", err)
	}
	if err := os.Setenv("REDIS_WRITE_TIMEOUT_SEC", "3"); err != nil {
		t.Fatalf("Failed to set REDIS_WRITE_TIMEOUT_SEC: %v", err)
	}
	if err := os.Setenv("REDIS_POOL_TIMEOUT_SEC", "4"); err != nil {
		t.Fatalf("Failed to set REDIS_POOL_TIMEOUT_SEC: %v", err)
	}
	if err := os.Setenv("REDIS_STREAM_MAX_LEN", "10000"); err != nil {
		t.Fatalf("Failed to set REDIS_STREAM_MAX_LEN: %v", err)
	}
	if err := os.Setenv("REDIS_JOB_STREAM_TTL_HOURS", "120"); err != nil {
		t.Fatalf("Failed to set REDIS_JOB_STREAM_TTL_HOURS: %v", err)
	}
	if err := os.Setenv("REDIS_TASK_STREAM_TTL_HOURS", "1"); err != nil {
		t.Fatalf("Failed to set REDIS_TASK_STREAM_TTL_HOURS: %v", err)
	}

	// Debug: print env and .env presence
	fmt.Println("[DEBUG] Current working directory:", getCurrentDir())
	fmt.Println("[DEBUG] DEV_MODE:", os.Getenv("DEV_MODE"))
	fmt.Println("[DEBUG] REDIS_ADDR:", os.Getenv("REDIS_ADDR"))

	// Look for .env in root directory
	rootDir := findRootDir()
	envPath := filepath.Join(rootDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		fmt.Println("[DEBUG] .env file found at:", envPath)
		// Print .env contents for debugging
		if envBytes, err := os.ReadFile(envPath); err == nil {
			fmt.Println("[DEBUG] .env contents:", string(envBytes))
		}
	} else {
		fmt.Println("[DEBUG] .env file NOT found at:", envPath, "error:", err)
	}

	// Test configuration initialization
	t.Run("Config Initialization", func(t *testing.T) {
		// Change to root directory for config initialization
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(originalDir); err != nil {
				t.Errorf("Failed to change back to original directory: %v", err)
			}
		}()

		if err := os.Chdir(rootDir); err != nil {
			t.Fatalf("Failed to change to root directory: %v", err)
		}

		err = config.Init()
		if err != nil && !os.IsNotExist(err) && !isNoEnvFileError(err) {
			assert.NoError(t, err, "Config initialization should not fail")
		}
		// Debug: print Redis config after initialization
		fmt.Println("[DEBUG] Redis config after init:")
		fmt.Println("- IsRedisAvailable:", config.IsRedisAvailable())
		fmt.Println("- GetRedisType:", config.GetRedisType())
		fmt.Println("- IsLocalRedisEnabled:", config.IsLocalRedisEnabled())
		fmt.Println("- GetRedisAddr:", config.GetRedisAddr())
	})

	logConfig := logging.LoggerConfig{
		ProcessName:     logging.RedisProcess,
		IsDevelopment:   true,
	}
	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	// Test logger initialization
	t.Run("Logger Initialization", func(t *testing.T) {
		if logger == nil {
			t.Fatalf("Logger should not be nil")
		}
		assert.NoError(t, err, "Logger initialization should not fail")
	})

	// Test Redis client creation and connection
	t.Run("Redis Client", func(t *testing.T) {
		client, err := redisx.NewRedisClient(logger)
		if err != nil {
			t.Fatalf("Redis client creation failed: %v", err)
		}
		if client == nil {
			t.Fatal("Redis client is nil")
		}
		logger.Info("Redis client created successfully")

		// Test Redis connection
		err = client.Ping()
		assert.NoError(t, err, "Redis ping should succeed")
		logger.Info("Redis ping successful")

		// Test Redis stream operations
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Add a test job to the stream
		testJob := map[string]interface{}{
			"type":      "test",
			"timestamp": time.Now().Unix(),
			"message":   "Redis service test job",
		}
		jobBytes, err := json.Marshal(testJob)
		assert.NoError(t, err, "Marshalling test job to JSON should succeed")

		res, err := client.XAdd(ctx, &redisclient.XAddArgs{
			Stream: redisx.JobsRunningStream,
			MaxLen: 1000,
			Approx: true,
			Values: map[string]interface{}{
				"job":        string(jobBytes),
				"created_at": time.Now().Unix(),
			},
		})
		assert.NoError(t, err, "Adding test job to stream should succeed")
		assert.NotEmpty(t, res, "Stream ID should not be empty")
		logger.Info("Test job added to stream successfully", "stream_id", res)

		// Test Redis info
		info := redisx.GetRedisInfo()
		assert.NotNil(t, info, "Redis info should not be nil")
		logger.Info("Redis info retrieved successfully", "info", info)

		// Test graceful shutdown
		err = client.Close()
		assert.NoError(t, err, "Redis client should close gracefully")
		logger.Info("Redis client closed successfully")
	})
}

// isNoEnvFileError checks if the error is about missing .env file
func isNoEnvFileError(err error) bool {
	return err != nil && (err.Error() == "error loading .env file: open .env: no such file or directory" || os.IsNotExist(err))
}

// getCurrentDir returns the current working directory
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error getting current directory: %v", err)
	}
	return dir
}

// findRootDir finds the root directory of the project
func findRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Look for go.mod file to identify root directory
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
