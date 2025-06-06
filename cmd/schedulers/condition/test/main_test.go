package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func init() {
	// Initialize logger for tests
	logConfig := logging.LoggerConfig{
		LogDir:          "/tmp/triggerx-test-logs",
		ProcessName:     "condition-scheduler-test",
		Environment:     logging.Development,
		UseColors:       true,
		MinStdoutLevel:  logging.DebugLevel,
		MinFileLogLevel: logging.DebugLevel,
	}
	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

func TestConditionSchedulerInitialization(t *testing.T) {
	logger := logging.GetServiceLogger()
	// Use a real DBServerClient, but with dummy config
	dbClient := &client.DBServerClient{}
	managerID := "test-condition-scheduler"
	sched, err := scheduler.NewConditionBasedScheduler(managerID, logger, dbClient)
	assert.NoError(t, err)
	assert.NotNil(t, sched)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { sched.Start(ctx) }()
	time.Sleep(100 * time.Millisecond)
	sched.Stop()
}

func TestConditionSchedulerServer(t *testing.T) {
	logger := logging.GetServiceLogger()
	dbClient := &client.DBServerClient{}
	managerID := "test-condition-scheduler"
	sched, err := scheduler.NewConditionBasedScheduler(managerID, logger, dbClient)
	assert.NoError(t, err)
	server := api.NewServer(api.Config{
		Port: "8080",
	}, api.Dependencies{
		Logger:    logger,
		Scheduler: sched,
	})
	go func() { _ = server.Start() }()
	time.Sleep(200 * time.Millisecond)
	resp, err := http.Get("http://localhost:8080/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Stop(ctx)
	assert.NoError(t, err)
}

func TestEnvironmentAndLogLevel(t *testing.T) {
	env := getEnvironment()
	assert.Contains(t, []logging.LogLevel{logging.Development, logging.Production}, env)
	level := getLogLevel()
	assert.Contains(t, []logging.Level{logging.DebugLevel, logging.InfoLevel}, level)
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
