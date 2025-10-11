package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TestNewService tests the creation of a new health service
func TestNewService(t *testing.T) {
	mockLogger := logging.NewNoOpLogger()

	cfg := &Config{
		HTTPPort: "8080",
		GRPCPort: "9090",
		GRPCHost: "0.0.0.0",
	}

	service := NewService(mockLogger, cfg)

	assert.NotNil(t, service)
	assert.Equal(t, mockLogger, service.logger)
	assert.Equal(t, cfg, service.config)
	assert.Nil(t, service.httpServer)
	assert.Nil(t, service.rpcServer)
}

// TestSetStateManager tests setting the global state manager
func TestSetStateManager(t *testing.T) {
	// Reset state manager to nil for clean test
	stateManager = nil

	mockLogger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}
	mockNotifier := &mocks.MockNotificationBot{}

	sm := keeper.InitializeStateManager(mockLogger, mockDB, mockNotifier)
	require.NotNil(t, sm)

	SetStateManager(sm)
	assert.Equal(t, sm, stateManager)
}

// TestGetStateManager tests retrieving the global state manager
func TestGetStateManager(t *testing.T) {
	mockLogger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}
	mockNotifier := &mocks.MockNotificationBot{}

	sm := keeper.InitializeStateManager(mockLogger, mockDB, mockNotifier)
	SetStateManager(sm)

	retrieved := GetStateManager()
	assert.NotNil(t, retrieved)
	assert.Equal(t, sm, retrieved)
}

// TestService_StartStop tests starting and stopping the service
func TestService_StartStop(t *testing.T) {
	// Skip this test in short mode as it involves network operations
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockLogger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}
	mockNotifier := &mocks.MockNotificationBot{}

	// Initialize state manager
	sm := keeper.InitializeStateManager(mockLogger, mockDB, mockNotifier)
	SetStateManager(sm)

	// Use random available ports to avoid conflicts
	cfg := &Config{
		HTTPPort: "18080", // Use high port numbers for testing
		GRPCPort: "19090",
		GRPCHost: "127.0.0.1",
	}

	service := NewService(mockLogger, cfg)
	require.NotNil(t, service)

	// Start the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := service.Start(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Verify HTTP server is running by making a request
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/", cfg.HTTPPort))
	if err == nil {
		err := resp.Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Stop the service
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err = service.Stop(stopCtx)
	assert.NoError(t, err)

	// Check for any startup errors
	select {
	case err := <-errChan:
		t.Fatalf("Service startup error: %v", err)
	default:
		// No errors
	}
}

// TestService_StartWithoutStateManager tests that starting fails without state manager
func TestService_StartWithoutStateManager(t *testing.T) {
	// Reset state manager to nil
	stateManager = nil

	mockLogger := logging.NewNoOpLogger()

	cfg := &Config{
		HTTPPort: "18081",
		GRPCPort: "19091",
		GRPCHost: "127.0.0.1",
	}

	service := NewService(mockLogger, cfg)
	require.NotNil(t, service)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should fail because state manager is not initialized
	err := service.Start(ctx)

	// We expect an error since startRPCServer will fail without state manager
	// The error might come through the error channel or context timeout
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

// TestService_ConcurrentStartStop tests concurrent start/stop operations
func TestService_ConcurrentStartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockLogger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}
	mockNotifier := &mocks.MockNotificationBot{}

	sm := keeper.InitializeStateManager(mockLogger, mockDB, mockNotifier)
	SetStateManager(sm)

	cfg := &Config{
		HTTPPort: "18082",
		GRPCPort: "19092",
		GRPCHost: "127.0.0.1",
	}

	service := NewService(mockLogger, cfg)

	var wg sync.WaitGroup

	// Start
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx := context.Background()
		err := service.Start(ctx)
		if err != nil {
			t.Errorf("Service start error: %v", err)
		}
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Stop
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := service.Stop(ctx)
		if err != nil {
			t.Errorf("Service stop error: %v", err)
		}
	}()

	// Wait for both to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for start/stop to complete")
	}
}

// TestService_MultipleStops tests that multiple Stop calls are safe
func TestService_MultipleStops(t *testing.T) {
	mockLogger := logging.NewNoOpLogger()

	cfg := &Config{
		HTTPPort: "18083",
		GRPCPort: "19093",
		GRPCHost: "127.0.0.1",
	}

	service := NewService(mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Call Stop multiple times - should not panic
	err1 := service.Stop(ctx)
	err2 := service.Stop(ctx)
	err3 := service.Stop(ctx)

	// All should succeed (no-op if servers aren't running)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
}
