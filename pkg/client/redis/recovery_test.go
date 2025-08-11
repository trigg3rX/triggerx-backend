//go:build integration_slow
// +build integration_slow

package redis

import (
	"context"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TestGetConnectionStatus checks the initial state of the connection status.
func TestGetConnectionStatus(t *testing.T) {
	// Uses the global testClient from main_test.go
	status := testClient.GetConnectionStatus()

	require.NotNil(t, status)
	assert.False(t, status.IsRecovering, "Should not be recovering initially")
	assert.True(t, status.RecoveryEnabled, "Recovery should be enabled by default")
	assert.NotZero(t, status.LastHealthCheck)
	assert.NotZero(t, status.PoolStats.TotalConns)
}

func TestConnectionRecovery_SuccessfulCycle_Slow(t *testing.T) {
	// This build tag ensures this slow test doesn't run with the default `go test`.

	// Skip this test unless explicitly requested via build tags
	if testing.Short() {
		t.Skip("Skipping slow integration test in short mode")
	}

	t.Log("Starting slow connection recovery test. This will stop/start the Redis Docker container.")

	// --- Setup ---

	// Ensure the redis container is running before we start and clean it up after.
	t.Cleanup(func() {
		t.Log("Ensuring redis container is running after test completion...")
		exec.Command("docker", "compose", "start", "redis").Run()
		time.Sleep(3 * time.Second) // Give it a moment to stabilize for other potential tests
	})

	// Create a new client with fast recovery settings for this test
	logger := logging.NewNoOpLogger()
	config := testClient.config // Use the same config as the global client
	fastRecoveryClient, err := NewRedisClient(logger, config)
	require.NoError(t, err)

	// Configure for a very fast recovery attempt to speed up the test
	fastRecoveryClient.SetConnectionRecoveryConfig(&ConnectionRecoveryConfig{
		Enabled:         true,
		CheckInterval:   10 * time.Second, // Interval doesn't matter as we trigger manually
		MaxRetries:      10,               // Plenty of retries
		BackoffFactor:   1.5,
		MaxBackoffDelay: 3 * time.Second,
	})

	// Use monitoring hooks to get signals from the recovery process
	var wg sync.WaitGroup
	wg.Add(2) // We expect two signals: start and end

	var recoveryStarted bool
	var recoveryEndedSuccess bool
	var recoveryAttempts int

	fastRecoveryClient.SetMonitoringHooks(&MonitoringHooks{
		OnRecoveryStart: func(reason string) {
			t.Logf("MonitoringHook: Recovery started (reason: %s)", reason)
			recoveryStarted = true
			wg.Done()
		},
		OnRecoveryEnd: func(success bool, attempts int, duration time.Duration) {
			t.Logf("MonitoringHook: Recovery ended (success: %v, attempts: %d)", success, attempts)
			recoveryEndedSuccess = success
			recoveryAttempts = attempts
			wg.Done()
		},
	})

	// --- Step 1: Simulate Redis Server Failure ---
	t.Log("Stopping redis container...")
	cmd := exec.Command("docker", "compose", "stop", "redis")
	err = cmd.Run()
	require.NoError(t, err, "Failed to stop redis container")
	time.Sleep(1 * time.Second) // Give a moment for the connection to be fully dropped

	// --- Step 2: Trigger Recovery ---
	t.Log("Pinging to trigger the recovery process...")
	// This ping will fail and kick off `performConnectionRecovery` in a goroutine
	err = fastRecoveryClient.Ping(context.Background())
	require.Error(t, err, "Ping should fail when Redis is down")

	// --- Step 3: Wait for Recovery to Start ---
	// We wait for the OnRecoveryStart hook to be called
	// This has a timeout to prevent the test from hanging indefinitely
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		// We proceed, but only after OnRecoveryStart has fired.
	case <-time.After(20 * time.Second):
		t.Fatal("Timeout waiting for recovery hooks")
	}

	require.True(t, recoveryStarted, "The recovery process should have been started")
	// At this point, the client is in its background retry loop

	// --- Step 4: Simulate Redis Server Recovery ---
	t.Log("Starting redis container...")
	cmd = exec.Command("docker", "compose", "start", "redis")
	err = cmd.Run()
	require.NoError(t, err, "Failed to start redis container")
	t.Log("Redis container started, waiting for recovery to succeed...")

	// --- Step 5: Wait for Recovery to End ---
	// The first `wg.Done()` was called by OnRecoveryStart. Now we wait for OnRecoveryEnd.
	select {
	case <-c:
		// This will unblock after OnRecoveryEnd calls wg.Done()
	case <-time.After(20 * time.Second):
		t.Fatal("Timeout waiting for recovery to complete after server restart")
	}

	// --- Step 6: Assert Final State ---
	assert.True(t, recoveryEndedSuccess, "Recovery should have been successful")
	assert.Greater(t, recoveryAttempts, 0, "Recovery should have taken at least one attempt")

	// Final check: The client should be healthy again
	err = fastRecoveryClient.Ping(context.Background())
	assert.NoError(t, err, "Ping should succeed after successful recovery")
}
