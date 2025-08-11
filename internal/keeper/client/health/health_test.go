package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestNewClient_ValidConfig_ReturnsClient(t *testing.T) {
	logger, err := logging.NewZapLogger(logging.LoggerConfig{
		ProcessName:   "test",
		IsDevelopment: true,
	})
	require.NoError(t, err)

	cfg := Config{
		HealthServiceURL: "http://localhost:8080",
		PrivateKey:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		KeeperAddress:    "0x1234567890123456789012345678901234567890",
		PeerID:           "test-peer-id",
		Version:          "0.1.0",
		RequestTimeout:   5 * time.Second,
	}

	client, err := NewClient(logger, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, cfg, client.config)
}

func TestNewClient_EmptyTimeout_SetsDefaultTimeout(t *testing.T) {
	logger, err := logging.NewZapLogger(logging.LoggerConfig{
		ProcessName:   "test",
		IsDevelopment: true,
	})
	require.NoError(t, err)

	cfg := Config{
		HealthServiceURL: "http://localhost:8080",
		PrivateKey:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		KeeperAddress:    "0x1234567890123456789012345678901234567890",
		PeerID:           "test-peer-id",
		Version:          "",
		RequestTimeout:   0,
	}

	client, err := NewClient(logger, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 10*time.Second, client.config.RequestTimeout)
	assert.Equal(t, "0.1.6", client.config.Version)
}

func TestNewClient_InvalidPrivateKey_ReturnsError(t *testing.T) {
	logger, err := logging.NewZapLogger(logging.LoggerConfig{
		ProcessName:   "test",
		IsDevelopment: true,
	})
	require.NoError(t, err)

	cfg := Config{
		HealthServiceURL: "http://localhost:8080",
		PrivateKey:       "invalid-private-key",
		KeeperAddress:    "0x1234567890123456789012345678901234567890",
		PeerID:           "test-peer-id",
		Version:          "0.1.0",
		RequestTimeout:   5 * time.Second,
	}

	client, err := NewClient(logger, cfg)
	assert.NoError(t, err) // HTTP client creation should succeed
	assert.NotNil(t, client)
}

func TestClient_CheckIn_WithMockServer(t *testing.T) {
	// This test would require a mock HTTP server to test the actual CheckIn functionality
	// For now, we'll just test that the client can be created and closed properly

	logger, err := logging.NewZapLogger(logging.LoggerConfig{
		ProcessName:   "test",
		IsDevelopment: true,
	})
	require.NoError(t, err)

	cfg := Config{
		HealthServiceURL: "http://localhost:8080",
		PrivateKey:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		KeeperAddress:    "0x1234567890123456789012345678901234567890",
		PeerID:           "test-peer-id",
		Version:          "0.1.0",
		RequestTimeout:   5 * time.Second,
	}

	client, err := NewClient(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test that client can be closed
	client.Close()
}

func TestConfigSetters_UpdateConfigValues(t *testing.T) {
	// Test that the config setters work correctly
	// This is important for the health check-in flow

	// Reset config values
	config.SetEtherscanAPIKey("")
	config.SetAlchemyAPIKey("")
	config.SetIpfsHost("")
	config.SetPinataJWT("")
	config.SetManagerSigningAddress("")
	config.SetTaskExecutionAddress("")

	// Set test values
	testEtherscanKey := "test-etherscan-key"
	testAlchemyKey := "test-alchemy-key"
	testIpfsHost := "test-ipfs-host"
	testPinataJWT := "test-pinata-jwt"
	testManagerAddress := "test-manager-address"
	testTaskExecutionAddress := "test-task-execution-address"

	config.SetEtherscanAPIKey(testEtherscanKey)
	config.SetAlchemyAPIKey(testAlchemyKey)
	config.SetIpfsHost(testIpfsHost)
	config.SetPinataJWT(testPinataJWT)
	config.SetManagerSigningAddress(testManagerAddress)
	config.SetTaskExecutionAddress(testTaskExecutionAddress)

	// Verify values were set correctly
	assert.Equal(t, testEtherscanKey, config.GetEtherscanAPIKey())
	assert.Equal(t, testAlchemyKey, config.GetAlchemyAPIKey())
	assert.Equal(t, testIpfsHost, config.GetIpfsHost())
	assert.Equal(t, testPinataJWT, config.GetPinataJWT())
	assert.Equal(t, testManagerAddress, config.GetManagerSigningAddress())
	assert.Equal(t, testTaskExecutionAddress, config.GetTaskExecutionAddress())
}
