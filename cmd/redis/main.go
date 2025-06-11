package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.RedisProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting Redis service main...",
		"mode", config.IsDevMode(),
	)

	// Create Redis client
	client, err := redisx.NewRedisClient(logger)
	if err != nil {
		logger.Errorf("Failed to create Redis client: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to create Redis client: %v\n", err)
		os.Exit(1)
	}

	// Test Redis connection
	if err := client.Ping(); err != nil {
		logger.Errorf("Redis is not reachable: %v", err)
		fmt.Fprintf(os.Stderr, "Redis is not reachable: %v\n", err)
		if err := client.Close(); err != nil {
			logger.Errorf("Error closing Redis client: %v", err)
		}
		os.Exit(1)
	}
	logger.Info("Redis ping successful.")

	// Add a test job to the jobs running stream
	testJob := map[string]interface{}{
		"type":      "test",
		"timestamp": time.Now().Unix(),
		"message":   "Redis service test job",
	}

	testJobJSON, err := json.Marshal(testJob)
	if err != nil {
		logger.Errorf("Failed to marshal test job: %v", err)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := client.XAdd(ctx, &redis.XAddArgs{
			Stream: redisx.JobsRunningStream,
			MaxLen: int64(config.GetStreamMaxLen()),
			Approx: true,
			Values: map[string]interface{}{
				"job":        testJobJSON,
				"created_at": time.Now().Unix(),
			},
		})
		cancel()

		if err != nil {
			logger.Errorf("Failed to add test job to stream: %v", err)
		} else {
			logger.Infof("Test job added to %s stream with ID %s", redisx.JobsRunningStream, res)
		}
	}

	// Log Redis info
	redisInfo := redisx.GetRedisInfo()
	logger.Infof("Redis configuration: %+v", redisInfo)

	// Set up graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Create shutdown context with timeout
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown in a goroutine
	go func() {
		sig := <-shutdownChan
		logger.Infof("Received shutdown signal: %v", sig)

		// Add shutdown event to Redis stream
		shutdownEvent := map[string]interface{}{
			"event_type":  "redis_service_shutdown",
			"signal":      sig.String(),
			"shutdown_at": time.Now().Unix(),
			"graceful":    true,
		}

		shutdownEventJSON, err := json.Marshal(shutdownEvent)
		if err != nil {
			logger.Warnf("Failed to marshal shutdown event: %v", err)
		} else {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
			_, err := client.XAdd(shutdownCtx, &redis.XAddArgs{
				Stream: redisx.JobsRunningStream,
				MaxLen: int64(config.GetStreamMaxLen()),
				Approx: true,
				Values: map[string]interface{}{
					"event":      shutdownEventJSON,
					"created_at": time.Now().Unix(),
				},
			})
			shutdownCancel()

			if err != nil {
				logger.Warnf("Failed to add shutdown event to stream: %v", err)
			} else {
				logger.Info("Shutdown event added to Redis stream")
			}
		}

		// Cancel the main context to trigger shutdown
		cancel()
	}()

	logger.Info("Redis service is running.")

	// Wait for shutdown signal or context cancellation
	<-ctx.Done()

	// Graceful shutdown sequence
	logger.Info("Starting graceful shutdown...")
	shutdownStart := time.Now()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Close Redis client connection
	logger.Info("Closing Redis client connection...")
	if err := client.Close(); err != nil {
		logger.Errorf("Error closing Redis client: %v", err)
	} else {
		logger.Info("Redis client closed successfully")
	}

	shutdownDuration := time.Since(shutdownStart)
	fmt.Printf("Redis service shutdown completed in %v\n", shutdownDuration)

	// Ensure we exit cleanly
	select {
	case <-shutdownCtx.Done():
		fmt.Fprintf(os.Stderr, "Shutdown timeout exceeded\n")
		os.Exit(1)
	default:
		os.Exit(0)
	}
}
