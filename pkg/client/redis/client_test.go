package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnection(t *testing.T) {
	ctx := context.Background()
	err := testClient.Ping(ctx)
	assert.NoError(t, err)

	err = testClient.CheckConnection(ctx)
	assert.NoError(t, err)
}

func TestGetSetDel(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:get-set-del"
	value := "hello world"

	// Set
	err := testClient.Set(ctx, key, value, 1*time.Minute)
	require.NoError(t, err)

	// Get
	retrieved, err := testClient.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, retrieved)

	// Get non-existent key
	_, err = testClient.Get(ctx, "non-existent-key")
	assert.ErrorIs(t, err, redis.Nil)

	// Del
	err = testClient.Del(ctx, key)
	require.NoError(t, err)

	// Verify deletion
	_, err = testClient.Get(ctx, key)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestGetWithExists(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:get-with-exists"
	value := "i am here"

	// Test when key does not exist
	_, exists, err := testClient.GetWithExists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Set key and test again
	createKey(t, key, value)
	retrieved, exists, err := testClient.GetWithExists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, value, retrieved)
}

func TestDelWithCount(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	createKey(t, "key1", "v1")
	createKey(t, "key2", "v2")
	createKey(t, "key3", "v3")

	deleted, err := testClient.DelWithCount(ctx, "key1", "key2", "non-existent")
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	_, err = testClient.Get(ctx, "key3")
	assert.NoError(t, err) // key3 should still exist

	_, err = testClient.Get(ctx, "key1")
	assert.ErrorIs(t, err, redis.Nil) // key1 should be gone
}

func TestSetNX(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:setnx"
	value := "unique"

	// First SetNX should succeed
	acquired, err := testClient.SetNX(ctx, key, value, 1*time.Minute)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Second SetNX should fail
	acquired, err = testClient.SetNX(ctx, key, "another-value", 1*time.Minute)
	require.NoError(t, err)
	assert.False(t, acquired)
}

func TestTTL(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key := "test:ttl"
	value := "hello world"

	// Test TTL on non-existent key - should return -2
	ttl, err := testClient.TTL(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-2), ttl)

	// Create key with 10-minute expiration (from createKey helper)
	createKey(t, key, value)

	// Get TTL immediately after creation
	ttl, err = testClient.TTL(ctx, key)
	require.NoError(t, err)
	// Should be approximately 10 minutes, allow some tolerance for test execution time
	assert.True(t, ttl > 9*time.Minute && ttl <= 10*time.Minute, "TTL should be around 10 minutes, got: %v", ttl)

	// Wait a bit and check TTL decreases
	time.Sleep(1 * time.Second)

	ttl, err = testClient.TTL(ctx, key)
	require.NoError(t, err)
	// Should be approximately 10 minutes minus 1 second
	assert.True(t, ttl > 9*time.Minute-2*time.Second && ttl <= 10*time.Minute-1*time.Second,
		"TTL should be around 10 minutes minus 1 second, got: %v", ttl)

	// Create a key without expiration
	err = testClient.Set(ctx, "no-expiry-key", "no-expiry-value", 0)
	require.NoError(t, err)

	// Test TTL on key without expiration - should return -1
	ttl, err = testClient.TTL(ctx, "no-expiry-key")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl)

	// Test TTL on non-existent key - should return -2
	ttl, err = testClient.TTL(ctx, "non-existent-key")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-2), ttl)
}

func TestExecutePipeline(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	key1, val1 := "pipe:key1", "pipe-val1"
	key2, val2 := "pipe:key2", "pipe-val2"

	_, err := testClient.ExecutePipeline(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, key1, val1, 0)
		pipe.Set(ctx, key2, val2, 0)
		pipe.Get(ctx, key1)
		return nil
	})
	require.NoError(t, err)

	// Verify results outside the pipeline
	retrieved1, err := testClient.Get(ctx, key1)
	require.NoError(t, err)
	assert.Equal(t, val1, retrieved1)

	retrieved2, err := testClient.Get(ctx, key2)
	require.NoError(t, err)
	assert.Equal(t, val2, retrieved2)
}

func TestTTL_ValidKey_ReturnsTTL(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	key := "test:ttl-key"
	expectedTTL := 30 * time.Second

	// Set up the mock to return a valid TTL
	mockClient.MockTTL = func(ctx context.Context, key string) (time.Duration, error) {
		return expectedTTL, nil
	}

	// Test TTL for valid key
	ttl, err := mockClient.TTL(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, expectedTTL, ttl)
}

func TestTTL_NonExistentKey_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	key := "non-existent-key"

	// Set up the mock to return redis.Nil for non-existent key
	mockClient.MockTTL = func(ctx context.Context, key string) (time.Duration, error) {
		return 0, redis.Nil
	}

	// Test TTL for non-existent key
	ttl, err := mockClient.TTL(ctx, key)
	assert.Error(t, err)
	assert.ErrorIs(t, err, redis.Nil)
	assert.Equal(t, time.Duration(0), ttl)
}

func TestTTL_KeyWithoutExpiration_ReturnsNegativeOne(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	key := "no-expiry-key"

	// Set up the mock to return -1 for key without expiration
	mockClient.MockTTL = func(ctx context.Context, key string) (time.Duration, error) {
		return time.Duration(-1), nil
	}

	// Test TTL for key without expiration
	ttl, err := mockClient.TTL(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl)
}

func TestTTL_WhenError_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	key := "test:ttl-key"
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error
	mockClient.MockTTL = func(ctx context.Context, key string) (time.Duration, error) {
		return 0, expectedError
	}

	// Test TTL when Redis returns an error
	ttl, err := mockClient.TTL(ctx, key)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, time.Duration(0), ttl)
}

func TestCheckConnection_WhenError_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error on CheckConnection
	mockClient.MockCheckConnection = func(ctx context.Context) error {
		return expectedError
	}

	// Test CheckConnection when there's an error
	err := mockClient.CheckConnection(ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestCheckConnection_WhenRedisNil_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()

	// Set up the mock to return redis.Nil
	mockClient.MockCheckConnection = func(ctx context.Context) error {
		return redis.Nil
	}

	// Test CheckConnection when Redis returns redis.Nil
	err := mockClient.CheckConnection(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestCheckConnection_WhenNetworkError_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	networkError := fmt.Errorf("network timeout")

	// Set up the mock to return a network error
	mockClient.MockCheckConnection = func(ctx context.Context) error {
		return networkError
	}

	// Test CheckConnection when there's a network error
	err := mockClient.CheckConnection(ctx)
	assert.Error(t, err)
	assert.Equal(t, networkError, err)
	assert.Contains(t, err.Error(), "network timeout")
}

func TestClose_Success(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)

	// Set up the mock to return no error on Close
	mockClient.MockClose = func() error {
		return nil
	}

	// Test that Close returns no error
	err := mockClient.Close()
	assert.NoError(t, err)
}

func TestClose_WhenError_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)

	// Set up the mock to return an error on Close
	expectedError := fmt.Errorf("connection close failed")
	mockClient.MockClose = func() error {
		return expectedError
	}

	// Test that Close returns the expected error
	err := mockClient.Close()
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}
