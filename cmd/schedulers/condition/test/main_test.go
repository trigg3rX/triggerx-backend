package test

// import (
// 	"context"
// 	"net/http"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/api"
// 	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/client"
// 	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/scheduler"
// 	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
// )

// func TestConditionSchedulerInitialization(t *testing.T) {
// 	// Use a real DBServerClient, but with dummy config
// 	logConfig := logging.LoggerConfig{
// 		ProcessName:   "condition-scheduler-test",
// 		IsDevelopment: true,
// 	}
// 	logger, err := logging.NewZapLogger(logConfig)
// 	if err != nil {
// 		panic("Failed to initialize logger: " + err.Error())
// 	}
// 	dbClient := &client.DBServerClient{}
// 	managerID := "test-condition-scheduler"
// 	sched, err := scheduler.NewConditionBasedScheduler(managerID, logger, dbClient)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, sched)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	go func() { sched.Start(ctx) }()
// 	time.Sleep(100 * time.Millisecond)
// 	sched.Stop()
// }

// func TestConditionSchedulerServer(t *testing.T) {
// 	logConfig := logging.LoggerConfig{
// 		ProcessName:   "condition-scheduler-test",
// 		IsDevelopment: true,
// 	}
// 	logger, err := logging.NewZapLogger(logConfig)
// 	if err != nil {
// 		panic("Failed to initialize logger: " + err.Error())
// 	}
// 	dbClient := &client.DBServerClient{}
// 	managerID := "test-condition-scheduler"
// 	sched, err := scheduler.NewConditionBasedScheduler(managerID, logger, dbClient)
// 	assert.NoError(t, err)
// 	server := api.NewServer(api.Config{
// 		Port: "8080",
// 	}, api.Dependencies{
// 		Logger:    logger,
// 		Scheduler: sched,
// 	})
// 	go func() { _ = server.Start() }()
// 	time.Sleep(200 * time.Millisecond)
// 	resp, err := http.Get("http://localhost:8080/status")
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	err = server.Stop(ctx)
// 	assert.NoError(t, err)
// }
