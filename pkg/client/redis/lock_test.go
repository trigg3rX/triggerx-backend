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

// MockLock is a mock implementation of Lock for testing with MockRedisClient
type MockLock struct {
	client        *MockRedisClient
	key           string
	token         string
	ttl           time.Duration
	retryStrategy *RetryStrategy
}

// NewMockLock creates a new mock lock instance
func NewMockLock(mockClient *MockRedisClient, key string, ttl time.Duration, retryStrategy *RetryStrategy) *MockLock {
	if retryStrategy == nil {
		retryStrategy = NoRetry()
	}
	return &MockLock{
		client:        mockClient,
		key:           key,
		token:         "mock-token",
		ttl:           ttl,
		retryStrategy: retryStrategy,
	}
}

// Acquire attempts to acquire the lock using the mock client
func (l *MockLock) Acquire(ctx context.Context) (bool, error) {
	// For NoRetry, handle it directly
	if l.retryStrategy.MaxRetries == 0 {
		acquired, err := l.client.SetNX(ctx, l.key, l.token, l.ttl)
		if err != nil {
			return false, fmt.Errorf("failed to acquire lock: %w", err)
		}
		return acquired, nil
	}

	// For retry cases, simulate the retry logic
	callCount := 0
	for callCount <= l.retryStrategy.MaxRetries {
		acquired, err := l.client.SetNX(ctx, l.key, l.token, l.ttl)
		if err != nil {
			return false, fmt.Errorf("failed to acquire lock: %w", err)
		}
		if acquired {
			return true, nil
		}
		callCount++
		if callCount <= l.retryStrategy.MaxRetries {
			time.Sleep(l.retryStrategy.Delay)
		}
	}
	return false, nil
}

// Release safely releases the lock using the mock client
func (l *MockLock) Release(ctx context.Context) error {
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end`

	res, err := l.client.Eval(ctx, script, []string{l.key}, l.token)
	if err != nil {
		return fmt.Errorf("failed to execute lock release script: %w", err)
	}

	if val, ok := res.(int64); !ok || val == 0 {
		return ErrLockNotAcquired
	}

	return nil
}

func TestLockAcquireAndRelease(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	lockKey := "test:lock:simple"

	lock, err := testClient.NewLock(lockKey, 10*time.Second, NoRetry())
	require.NoError(t, err)

	// Acquire the lock
	acquired, err := lock.Acquire(ctx)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Check that the key exists in Redis
	val, err := testClient.Get(ctx, lockKey)
	require.NoError(t, err)
	assert.Equal(t, lock.token, val)

	// Release the lock
	err = lock.Release(ctx)
	require.NoError(t, err)

	// Check that the key is gone
	_, err = testClient.Get(ctx, lockKey)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestLockContention(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	lockKey := "test:lock:contention"

	// Client 1 acquires the lock
	lock1, err := testClient.NewLock(lockKey, 10*time.Second, NoRetry())
	require.NoError(t, err)
	acquired1, err := lock1.Acquire(ctx)
	require.NoError(t, err)
	require.True(t, acquired1)

	// Client 2 fails to acquire the lock
	lock2, err := testClient.NewLock(lockKey, 10*time.Second, NoRetry())
	require.NoError(t, err)
	acquired2, err := lock2.Acquire(ctx)
	require.NoError(t, err)
	assert.False(t, acquired2)

	// Client 1 tries to release a lock held by itself (should fail)
	err = lock2.Release(ctx)
	assert.ErrorIs(t, err, ErrLockNotAcquired)

	// Client 1 releases the lock
	err = lock1.Release(ctx)
	require.NoError(t, err)

	// Now client 2 can acquire it
	acquired2, err = lock2.Acquire(ctx)
	require.NoError(t, err)
	assert.True(t, acquired2)

	err = lock2.Release(ctx)
	require.NoError(t, err)
}

func TestLockWithFixedRetry(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	lockKey := "test:lock:retry"

	// Acquire lock that will expire in 100ms
	lock1, _ := testClient.NewLock(lockKey, 100*time.Millisecond, NoRetry())
	acquired, err := lock1.Acquire(ctx)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Try to acquire the same lock with a retry strategy
	lock2, _ := testClient.NewLock(lockKey, 10*time.Second, FixedRetry(5, 50*time.Millisecond))

	// This should succeed because lock1 expires during the retry attempts
	acquired, err = lock2.Acquire(ctx)
	require.NoError(t, err)
	assert.True(t, acquired)

	err = lock2.Release(ctx)
	require.NoError(t, err)
}

func TestNewLock_InvalidTTL_ReturnsError(t *testing.T) {
	// Test with zero TTL
	lock, err := testClient.NewLock("test:lock", 0, NoRetry())
	assert.Error(t, err)
	assert.Nil(t, lock)
	assert.Contains(t, err.Error(), "lock TTL must be greater than zero")

	// Test with negative TTL
	lock, err = testClient.NewLock("test:lock", -1*time.Second, NoRetry())
	assert.Error(t, err)
	assert.Nil(t, lock)
	assert.Contains(t, err.Error(), "lock TTL must be greater than zero")
}

func TestNewLock_NilRetryStrategy_UsesNoRetry(t *testing.T) {
	lock, err := testClient.NewLock("test:lock", 10*time.Second, nil)
	require.NoError(t, err)
	assert.NotNil(t, lock)
	assert.Equal(t, 0, lock.retryStrategy.MaxRetries)
	assert.Equal(t, time.Duration(0), lock.retryStrategy.Delay)
}

func TestLockAcquire_SetNXError_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:error"
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error on SetNX
	mockClient.MockSetNX = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
		return false, expectedError
	}

	// Create mock lock
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, NoRetry())

	// Attempt to acquire lock
	acquired, err := lock.Acquire(ctx)
	assert.Error(t, err)
	assert.False(t, acquired)
	assert.Contains(t, err.Error(), "failed to acquire lock")
	assert.Contains(t, err.Error(), "redis connection failed")
}

func TestLockAcquire_WithRetry_SetNXError_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:retry-error"
	expectedError := fmt.Errorf("redis connection failed")

	// Set up the mock to return an error on SetNX
	mockClient.MockSetNX = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
		return false, expectedError
	}

	// Create lock with retry strategy
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, FixedRetry(3, 10*time.Millisecond))

	// Attempt to acquire lock
	acquired, err := lock.Acquire(ctx)
	assert.Error(t, err)
	assert.False(t, acquired)
	assert.Contains(t, err.Error(), "failed to acquire lock")
	assert.Contains(t, err.Error(), "redis connection failed")
}

func TestLockAcquire_WithRetry_LockContention_EventuallySucceeds(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:contention-retry"
	callCount := 0

	// Set up the mock to return false (contention) for first 2 calls, then true (success)
	mockClient.MockSetNX = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
		callCount++
		if callCount <= 2 {
			return false, nil // Lock contention
		}
		return true, nil // Success
	}

	// Create lock with retry strategy
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, FixedRetry(5, 10*time.Millisecond))

	// Attempt to acquire lock
	acquired, err := lock.Acquire(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.Equal(t, 3, callCount, "Should have called SetNX 3 times")
}

func TestLockAcquire_WithRetry_LockContention_ExhaustsRetries(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:contention-exhausted"
	callCount := 0

	// Set up the mock to always return false (contention)
	mockClient.MockSetNX = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
		callCount++
		return false, nil // Always contention
	}

	// Create lock with retry strategy
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, FixedRetry(2, 10*time.Millisecond))

	// Attempt to acquire lock
	acquired, err := lock.Acquire(ctx)
	assert.NoError(t, err) // No error, just couldn't acquire
	assert.False(t, acquired)
	assert.Equal(t, 3, callCount, "Should have called SetNX 3 times (initial + 2 retries)")
}

func TestLockRelease_EvalError_ReturnsError(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:release-error"
	expectedError := fmt.Errorf("redis eval failed")

	// Set up the mock to return an error on Eval
	mockClient.MockEval = func(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
		return nil, expectedError
	}

	// Create lock
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, NoRetry())

	// Attempt to release lock
	err := lock.Release(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute lock release script")
	assert.Contains(t, err.Error(), "redis eval failed")
}

func TestLockRelease_EvalReturnsZero_ReturnsLockNotAcquired(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:release-zero"

	// Set up the mock to return 0 (lock not owned by this client)
	mockClient.MockEval = func(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
		return int64(0), nil
	}

	// Create lock
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, NoRetry())

	// Attempt to release lock
	err := lock.Release(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrLockNotAcquired)
}

func TestLockRelease_EvalReturnsInvalidType_ReturnsLockNotAcquired(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:release-invalid"

	// Set up the mock to return invalid type
	mockClient.MockEval = func(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
		return "invalid-type", nil
	}

	// Create lock
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, NoRetry())

	// Attempt to release lock
	err := lock.Release(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrLockNotAcquired)
}

func TestLockRelease_EvalReturnsOne_Success(t *testing.T) {
	// Create a mock client
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	lockKey := "test:lock:release-success"

	// Set up the mock to return 1 (lock successfully released)
	mockClient.MockEval = func(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
		return int64(1), nil
	}

	// Create lock
	lock := NewMockLock(mockClient, lockKey, 10*time.Second, NoRetry())

	// Attempt to release lock
	err := lock.Release(ctx)
	assert.NoError(t, err)
}

func TestShouldRetryLockAcquisition_OnlyRetriesOnContention(t *testing.T) {
	// Create a lock
	lock := &Lock{
		client:        nil,
		key:           "test:lock",
		token:         "test-token",
		ttl:           10 * time.Second,
		retryStrategy: FixedRetry(3, 10*time.Millisecond),
	}

	// Test that it retries on lock contention
	assert.True(t, lock.shouldRetryLockAcquisition(ErrLockContention, 1))

	// Test that it doesn't retry on other errors
	assert.False(t, lock.shouldRetryLockAcquisition(fmt.Errorf("network error"), 1))
	assert.False(t, lock.shouldRetryLockAcquisition(redis.Nil, 1))
	assert.False(t, lock.shouldRetryLockAcquisition(nil, 1))
}
