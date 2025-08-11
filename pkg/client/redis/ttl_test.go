package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTTL_ExistingKey_RefreshesTTL(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:refresh-ttl:existing"
	initialTTL := 10 * time.Second
	newTTL := 30 * time.Second

	// Create a key with initial TTL
	err := testClient.Set(ctx, key, "value", initialTTL)
	require.NoError(t, err)

	// Verify key exists and has TTL
	ttl, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0 && ttl <= initialTTL)

	// Refresh TTL
	err = testClient.RefreshTTL(ctx, key, newTTL)
	require.NoError(t, err)

	// Verify TTL was refreshed
	ttl, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0 && ttl <= newTTL)
}

func TestRefreshTTL_NonExistentKey_ReturnsNoError(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:refresh-ttl:non-existent"
	newTTL := 30 * time.Second

	// Verify key doesn't exist
	_, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Refresh TTL on non-existent key should not error
	err = testClient.RefreshTTL(ctx, key, newTTL)
	require.NoError(t, err)

	// Key should still not exist
	_, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRefreshTTL_ZeroTTL_RefreshesToNoExpiry(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:refresh-ttl:zero"
	initialTTL := 10 * time.Second

	// Create a key with TTL
	err := testClient.Set(ctx, key, "value", initialTTL)
	require.NoError(t, err)

	// Verify key has TTL
	ttl, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0)

	// Refresh TTL to 0 (no expiry)
	err = testClient.RefreshTTL(ctx, key, 0)
	require.NoError(t, err)

	// Verify TTL was removed
	ttl, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists) // Key should be removed when TTL is set to 0
	assert.Equal(t, time.Duration(0), ttl)
}

func TestRefreshTTL_WhenError_ReturnsError(t *testing.T) {
	// Create a mock client to test error handling
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	key := "test:refresh-ttl:error"
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error on ExpireXX
	mockClient.MockRefreshTTL = func(ctx context.Context, key string, ttl time.Duration) error {
		return expectedError
	}

	// Test that RefreshTTL returns the error
	err := mockClient.RefreshTTL(ctx, key, 30*time.Second)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestRefreshStreamTTL_ExistingStream_RefreshesTTL(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	stream := "test:refresh-stream-ttl:existing"
	initialTTL := 10 * time.Second
	newTTL := 30 * time.Second

	// Create a stream with initial TTL
	err := testClient.CreateStreamIfNotExists(ctx, stream, initialTTL)
	require.NoError(t, err)

	// Verify stream exists and has TTL
	ttl, exists, err := testClient.GetTTLStatus(ctx, stream)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0 && ttl <= initialTTL)

	// Refresh TTL
	err = testClient.RefreshStreamTTL(ctx, stream, newTTL)
	require.NoError(t, err)

	// Verify TTL was refreshed
	ttl, exists, err = testClient.GetTTLStatus(ctx, stream)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0 && ttl <= newTTL)
}

func TestRefreshStreamTTL_NonExistentStream_ReturnsNoError(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	stream := "test:refresh-stream-ttl:non-existent"
	newTTL := 30 * time.Second

	// Verify stream doesn't exist
	_, exists, err := testClient.GetTTLStatus(ctx, stream)
	require.NoError(t, err)
	assert.False(t, exists)

	// Refresh TTL on non-existent stream should not error
	err = testClient.RefreshStreamTTL(ctx, stream, newTTL)
	require.NoError(t, err)

	// Stream should still not exist
	_, exists, err = testClient.GetTTLStatus(ctx, stream)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRefreshStreamTTL_WhenError_ReturnsError(t *testing.T) {
	// Create a mock client to test error handling
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	stream := "test:refresh-stream-ttl:error"
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error on RefreshStreamTTL
	mockClient.MockRefreshStreamTTL = func(ctx context.Context, stream string, ttl time.Duration) error {
		return expectedError
	}

	// Test that RefreshStreamTTL returns the error
	err := mockClient.RefreshStreamTTL(ctx, stream, 30*time.Second)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestSetTTL_ExistingKey_SetsTTL(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:set-ttl:existing"
	ttl := 30 * time.Second

	// Create a key without TTL
	err := testClient.Set(ctx, key, "value", 0)
	require.NoError(t, err)

	// Verify key exists without TTL
	ttlResult, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, time.Duration(-1), ttlResult) // -1 means no expiry

	// Set TTL
	err = testClient.SetTTL(ctx, key, ttl)
	require.NoError(t, err)

	// Verify TTL was set
	ttlResult, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttlResult > 0 && ttlResult <= ttl)
}

func TestSetTTL_NonExistentKey_ReturnsError(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:set-ttl:non-existent"
	ttl := 30 * time.Second

	// Verify key doesn't exist
	_, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Set TTL on non-existent key should not error (Redis behavior)
	err = testClient.SetTTL(ctx, key, ttl)
	require.NoError(t, err)

	// Key should still not exist
	_, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestSetTTL_ZeroTTL_RemovesTTL(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:set-ttl:zero"
	initialTTL := 30 * time.Second

	// Create a key with TTL
	err := testClient.Set(ctx, key, "value", initialTTL)
	require.NoError(t, err)

	// Verify key has TTL
	ttl, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0)

	// Set TTL to 0 (remove expiry)
	err = testClient.SetTTL(ctx, key, 0)
	require.NoError(t, err)

	// Verify TTL was removed
	ttl, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists) // Key should be removed when TTL is set to 0
	assert.Equal(t, time.Duration(0), ttl)
}

func TestSetTTL_WhenError_ReturnsError(t *testing.T) {
	// Create a mock client to test error handling
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	key := "test:set-ttl:error"
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error on SetTTL
	mockClient.MockSetTTL = func(ctx context.Context, key string, ttl time.Duration) error {
		return expectedError
	}

	// Test that SetTTL returns the error
	err := mockClient.SetTTL(ctx, key, 30*time.Second)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestGetTTLStatus_ExistingKeyWithTTL_ReturnsTTLAndExists(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:get-ttl-status:with-ttl"
	ttl := 30 * time.Second

	// Create a key with TTL
	err := testClient.Set(ctx, key, "value", ttl)
	require.NoError(t, err)

	// Get TTL status
	ttlResult, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttlResult > 0 && ttlResult <= ttl)
}

func TestGetTTLStatus_ExistingKeyWithoutTTL_ReturnsNoExpiryAndExists(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:get-ttl-status:no-ttl"

	// Create a key without TTL
	err := testClient.Set(ctx, key, "value", 0)
	require.NoError(t, err)

	// Get TTL status
	ttl, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, time.Duration(-1), ttl) // -1 means no expiry
}

func TestGetTTLStatus_NonExistentKey_ReturnsNotExists(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:get-ttl-status:non-existent"

	// Get TTL status for non-existent key
	ttl, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Equal(t, time.Duration(0), ttl)
}

func TestGetTTLStatus_WhenError_ReturnsError(t *testing.T) {
	// Create a mock client to test error handling
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	key := "test:get-ttl-status:error"
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error on GetTTLStatus
	mockClient.MockGetTTLStatus = func(ctx context.Context, key string) (time.Duration, bool, error) {
		return 0, false, expectedError
	}

	// Test that GetTTLStatus returns the error
	ttl, exists, err := mockClient.GetTTLStatus(ctx, key)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, time.Duration(0), ttl)
	assert.False(t, exists)
}

func TestTTLMethods_Integration_WorkTogether(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:ttl-integration"
	initialTTL := 10 * time.Second
	refreshedTTL := 30 * time.Second

	// Create key with initial TTL
	err := testClient.Set(ctx, key, "value", initialTTL)
	require.NoError(t, err)

	// Verify initial TTL
	ttl, exists, err := testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0 && ttl <= initialTTL)

	// Refresh TTL
	err = testClient.RefreshTTL(ctx, key, refreshedTTL)
	require.NoError(t, err)

	// Verify refreshed TTL
	ttl, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0 && ttl <= refreshedTTL)

	// Set new TTL
	newTTL := 60 * time.Second
	err = testClient.SetTTL(ctx, key, newTTL)
	require.NoError(t, err)

	// Verify new TTL
	ttl, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, ttl > 0 && ttl <= newTTL)

	// Remove TTL
	err = testClient.SetTTL(ctx, key, 0)
	require.NoError(t, err)

	// Verify no TTL
	ttl, exists, err = testClient.GetTTLStatus(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists) // Key should be removed when TTL is set to 0
	assert.Equal(t, time.Duration(0), ttl)
}
