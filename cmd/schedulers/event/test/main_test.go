package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// MockDBServerClient is a mock implementation of the DBServerClient
type MockDBServerClient struct {
	mock.Mock
	*client.DBServerClient
}

func (m *MockDBServerClient) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDBServerClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestEventSchedulerInitialization(t *testing.T) {
	logConfig := logging.LoggerConfig{
		ProcessName:     "event-scheduler-test",
		IsDevelopment:   true,
	}
	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	dbClient := &client.DBServerClient{}
	managerID := "test-event-scheduler"
	sched, err := scheduler.NewEventBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		t.Skipf("Skipping test: could not initialize event scheduler (no chain clients): %v", err)
	}
	assert.NotNil(t, sched)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { sched.Start(ctx) }()
	time.Sleep(100 * time.Millisecond)
	sched.Stop()
}

func TestEventSchedulerServer(t *testing.T) {
	logConfig := logging.LoggerConfig{
		ProcessName:     "event-scheduler-test",
		IsDevelopment:   true,
	}
	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	dbClient := &client.DBServerClient{}
	managerID := "test-event-scheduler"
	sched, err := scheduler.NewEventBasedScheduler(managerID, logger, dbClient)
	if err != nil {
		t.Skipf("Skipping test: could not initialize event scheduler (no chain clients): %v", err)
	}
	server := api.NewServer(api.Config{
		Port: "8081",
	}, api.Dependencies{
		Logger:    logger,
		Scheduler: sched,
	})
	go func() { _ = server.Start() }()
	time.Sleep(200 * time.Millisecond)
	resp, err := http.Get("http://localhost:8081/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Stop(ctx)
	assert.NoError(t, err)
}
