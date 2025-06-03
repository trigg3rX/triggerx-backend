package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		LogDir:          logging.BaseDataDir,
		ProcessName:     logging.RedisProcess,
		Environment:     getEnvironment(),
		UseColors:       true,
		MinStdoutLevel:  getLogLevel(),
		MinFileLogLevel: getLogLevel(),
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting Redis service main...",
		"mode", getEnvironment(),
	)
	defer func() {
		logger.Info("Shutting down Redis service main...")
		if err := logging.Shutdown(); err != nil {
			logger.Warnf("Logger shutdown error: %v", err)
		}
	}()

	// Handle SIGINT/SIGTERM for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Infof("Received signal: %v, shutting down...", sig)
		os.Exit(0)
	}()

	// Ping Redis
	if err := redisx.GetRedisClient().Ping(context.Background()).Err(); err != nil {
		logger.Errorf("Redis is not reachable: %v", err)
		fmt.Fprintf(os.Stderr, "Redis is not reachable: %v\n", err)
		os.Exit(1)
	}
	logger.Info("Redis ping successful.")

	// Log Redis persistence config
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	info, err := redisx.GetRedisClient().Info(ctx, "persistence").Result()
	if err != nil {
		logger.Warnf("Failed to fetch Redis persistence info: %v", err)
	} else {
		logger.Infof("Redis persistence info:\n%s", info)
	}

	// Add a test job to the jobs:ready stream
	testJob := map[string]interface{}{
		"type":      "test",
		"timestamp": time.Now().Unix(),
	}
	if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, testJob); err != nil {
		logger.Errorf("Failed to add test job to stream: %v", err)
	} else {
		logger.Info("Test job added to jobs:ready stream.")
	}

	// Block forever
	select {}
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
