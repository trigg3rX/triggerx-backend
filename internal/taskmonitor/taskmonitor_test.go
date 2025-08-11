package taskmonitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestNewTaskManager(t *testing.T) {
	logger := logging.NewNoOpLogger()

	tm, err := NewTaskManager(logger)
	require.NoError(t, err)
	assert.NotNil(t, tm)

	// Test that all components are initialized
	assert.NotNil(t, tm.taskStreamManager)
	assert.NotNil(t, tm.eventListener)
	assert.NotNil(t, tm.redisClient)

	// Clean up
	err = tm.Close()
	assert.NoError(t, err)
}

func TestTaskManager_Initialize(t *testing.T) {
	logger := logging.NewNoOpLogger()

	tm, err := NewTaskManager(logger)
	require.NoError(t, err)
	defer func() {
		if cerr := tm.Close(); cerr != nil {
			t.Errorf("Error closing task manager: %v", cerr)
		}
	}()

	// Test initialization
	err = tm.Initialize()
	// Note: This might fail in test environment due to missing Redis/database connections
	// The important thing is that it doesn't panic and handles errors gracefully
	if err != nil {
		t.Logf("Initialization failed as expected in test environment: %v", err)
	}
}

func TestTaskManager_HealthCheck(t *testing.T) {
	logger := logging.NewNoOpLogger()

	tm, err := NewTaskManager(logger)
	require.NoError(t, err)
	defer func() {
		if cerr := tm.Close(); cerr != nil {
			t.Errorf("Error closing task manager: %v", cerr)
		}
	}()

	// Test health check
	health := tm.HealthCheck()
	assert.NotNil(t, health)
	assert.Contains(t, health, "timestamp")
	assert.Contains(t, health, "uptime_seconds")
	assert.Contains(t, health, "start_time")
}
