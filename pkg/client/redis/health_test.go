//go:build integration_slow
// +build integration_slow

package redis

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthFunctions_Healthy verifies all health check functions when Redis is running.
func TestHealthFunctions_Healthy(t *testing.T) {
	ctx := context.Background()

	t.Run("GetHealthStatus", func(t *testing.T) {
		status := testClient.GetHealthStatus(ctx)

		require.NotNil(t, status)
		assert.True(t, status.Connected, "Should be connected")
		assert.Empty(t, status.Errors, "Should have no errors")
		assert.NotZero(t, status.PingLatency, "Ping latency should be recorded")
		require.NotNil(t, status.ServerInfo, "ServerInfo should be populated")
		assert.NotEmpty(t, status.ServerInfo.RawInfo, "Raw server info should not be empty")
	})

	t.Run("IsHealthy", func(t *testing.T) {
		healthy := testClient.IsHealthy(ctx)
		assert.True(t, healthy, "IsHealthy should return true")
	})

	t.Run("PerformHealthCheck", func(t *testing.T) {
		results, err := testClient.PerformHealthCheck(ctx)

		require.NoError(t, err, "PerformHealthCheck should not return a top-level error")
		require.NotNil(t, results)

		// Overall
		assert.True(t, results.Overall.Healthy, "Overall health should be true")
		assert.False(t, results.Overall.Timestamp.IsZero())

		// Ping
		assert.True(t, results.Ping.Success, "Ping check should be successful")
		assert.Empty(t, results.Ping.Error, "Ping error should be empty")
		assert.Greater(t, results.Ping.Latency, time.Duration(0))

		// Set/Get/Cleanup via Lua
		assert.True(t, results.Set.Success, "Set check should be successful")
		assert.True(t, results.Get.Success, "Get check should be successful")
		assert.True(t, results.Get.ValueMatch, "Get value should match the set value")
		assert.Empty(t, results.Cleanup.Error, "Cleanup should have no errors")

		// Pool Stats
		assert.NotZero(t, results.PoolStats.TotalConns, "Should have active connections in the pool")
	})
}

// TestHealthFunctions_Unhealthy_Slow verifies all health check functions when Redis is down.
// To run this test specifically, use the 'integration_slow' build tag:
// go test -v -tags=integration_slow ./...
func TestHealthFunctions_Unhealthy_Slow(t *testing.T) {

	t.Log("Starting slow unhealthy health check test. This will stop the Redis Docker container.")

	// Ensure the redis container is stopped, and guarantee it's started again after the test.
	require.NoError(t, exec.Command("docker", "compose", "stop", "redis").Run())
	t.Cleanup(func() {
		t.Log("Ensuring redis container is running after test completion...")
		exec.Command("docker", "compose", "start", "redis").Run()
		time.Sleep(3 * time.Second) // Give Redis time to initialize for subsequent tests.
	})
	time.Sleep(1 * time.Second) // Give connections time to drop.

	// Use a context with a short timeout to prevent the test from hanging for too long.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("GetHealthStatus", func(t *testing.T) {
		status := testClient.GetHealthStatus(ctx)

		require.NotNil(t, status)
		assert.False(t, status.Connected, "Should not be connected")
		assert.NotEmpty(t, status.Errors, "Should have errors reported")
		assert.Contains(t, status.Errors[0], "connection refused", "Error message should indicate connection failure")
	})

	t.Run("IsHealthy", func(t *testing.T) {
		healthy := testClient.IsHealthy(ctx)
		assert.False(t, healthy, "IsHealthy should return false")
	})

	t.Run("PerformHealthCheck", func(t *testing.T) {
		results, err := testClient.PerformHealthCheck(ctx)

		// A top-level error might be returned from the retry mechanism if the context times out.
		// The key is to check the detailed results struct.
		if err != nil {
			t.Logf("PerformHealthCheck returned expected top-level error: %v", err)
		}
		require.NotNil(t, results)

		// Overall
		assert.False(t, results.Overall.Healthy, "Overall health should be false")

		// Ping
		assert.False(t, results.Ping.Success, "Ping check should fail")
		assert.NotEmpty(t, results.Ping.Error, "Ping error should be populated")

		// Set/Get/Cleanup via Lua
		assert.False(t, results.Set.Success, "Set check should fail")
		assert.False(t, results.Get.Success, "Get check should fail")
		assert.False(t, results.Get.ValueMatch, "Get value match should be false")
	})
}
