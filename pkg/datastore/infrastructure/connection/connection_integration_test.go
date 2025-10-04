//go:build integration
// +build integration

package connection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TestIntegration_Connection tests actual database connection
func TestIntegration_Connection(t *testing.T) {
	config := NewConfig("localhost", "9043") // Test database port
	logger := logging.NewNoOpLogger()

	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	require.NotNil(t, conn)

	defer conn.Close()

	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = conn.HealthCheck(ctx)
	assert.NoError(t, err)
}

// TestIntegration_HealthCheck tests actual health check
func TestIntegration_HealthCheck(t *testing.T) {
	config := NewConfig("localhost", "9043")
	logger := logging.NewNoOpLogger()

	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	require.NotNil(t, conn)

	defer conn.Close()

	// Test multiple health checks
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = conn.HealthCheck(ctx)
		cancel()

		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}
}

// TestIntegration_Reconnection tests actual reconnection scenario
func TestIntegration_Reconnection(t *testing.T) {
	config := NewConfig("localhost", "9043")
	config.HealthCheckInterval = time.Millisecond * 100
	logger := logging.NewNoOpLogger()

	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	require.NotNil(t, conn)

	defer conn.Close()

	// Test that health checker is running
	time.Sleep(200 * time.Millisecond)

	// Verify health status
	if manager, ok := conn.(*scyllaConnectionManager); ok {
		assert.True(t, manager.healthCheckCount > 0)
	}
}
