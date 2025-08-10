package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// ErrLockNotAcquired is returned when an attempt to release a lock that was not acquired fails.
var ErrLockNotAcquired = errors.New("cannot release a lock that is not acquired")

// ErrLockContention is returned when a lock cannot be acquired due to contention (not network errors).
var ErrLockContention = errors.New("lock is held by another client")

// RetryStrategy defines the retry behavior for acquiring a lock.
type RetryStrategy struct {
	MaxRetries int
	Delay      time.Duration
}

// NoRetry is a strategy that does not attempt to retry.
func NoRetry() *RetryStrategy {
	return &RetryStrategy{MaxRetries: 0, Delay: 0}
}

// FixedRetry is a strategy that retries a fixed number of times with a constant delay.
func FixedRetry(maxRetries int, delay time.Duration) *RetryStrategy {
	return &RetryStrategy{MaxRetries: maxRetries, Delay: delay}
}

// Lock represents a distributed lock object.
type Lock struct {
	client        *Client
	key           string
	token         string
	ttl           time.Duration
	retryStrategy *RetryStrategy
}

// NewLock creates a new distributed lock instance.
// key: The Redis key to use for the lock.
// ttl: The time-to-live for the lock. It must be greater than zero.
// retryStrategy: The strategy for retrying to acquire the lock if it's already held.
func (c *Client) NewLock(key string, ttl time.Duration, retryStrategy *RetryStrategy) (*Lock, error) {
	if ttl <= 0 {
		return nil, errors.New("lock TTL must be greater than zero")
	}
	if retryStrategy == nil {
		retryStrategy = NoRetry()
	}
	return &Lock{
		client: c,
		key:    key,
		// The token is a unique value for this specific lock attempt (fencing token).
		token:         uuid.NewString(),
		ttl:           ttl,
		retryStrategy: retryStrategy,
	}, nil
}

// Acquire attempts to acquire the lock. It will retry based on the configured RetryStrategy.
func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	// For NoRetry, handle it directly without the retry package
	if l.retryStrategy.MaxRetries == 0 {
		// Attempt to set the key with our unique token, only if it does not exist (NX).
		acquired, err := l.client.SetNX(ctx, l.key, l.token, l.ttl)
		if err != nil {
			// This is a network or Redis error, not a lock contention issue.
			return false, fmt.Errorf("failed to acquire lock: %w", err)
		}
		return acquired, nil
	}

	// Convert RetryStrategy to generic RetryConfig for retry cases
	retryConfig := &retry.RetryConfig{
		MaxRetries:      l.retryStrategy.MaxRetries,
		InitialDelay:    l.retryStrategy.Delay,
		MaxDelay:        l.retryStrategy.Delay, // Use the same delay for all retries
		BackoffFactor:   1.0,                   // No exponential backoff for locks
		JitterFactor:    0.0,                   // No jitter for locks
		LogRetryAttempt: false,                 // Don't log retry attempts for locks
		ShouldRetry:     l.shouldRetryLockAcquisition,
	}

	// Create the lock acquisition operation
	operation := func() (bool, error) {
		// Attempt to set the key with our unique token, only if it does not exist (NX).
		acquired, err := l.client.SetNX(ctx, l.key, l.token, l.ttl)
		if err != nil {
			// This is a network or Redis error, not a lock contention issue.
			return false, fmt.Errorf("failed to acquire lock: %w", err)
		}

		if acquired {
			// We got the lock.
			return true, nil
		}

		// Lock is held by another client - this is expected contention
		return false, ErrLockContention
	}

	// Use the generic retry package
	result, err := retry.Retry(ctx, operation, retryConfig, l.client.logger)
	if err != nil {
		if errors.Is(err, ErrLockContention) {
			// Lock contention after all retries - this is not an error, just couldn't acquire
			return false, nil
		}
		// Network or other error
		return false, err
	}

	return result, nil
}

// shouldRetryLockAcquisition determines if we should retry lock acquisition
func (l *Lock) shouldRetryLockAcquisition(err error) bool {
	// Only retry on lock contention, not on network errors
	return errors.Is(err, ErrLockContention)
}

// Release safely releases the lock. It uses a Lua script to ensure that it only
// deletes the lock if it is still the owner (i.e., the token matches).
func (l *Lock) Release(ctx context.Context) error {
	// This Lua script atomically gets the key, compares its value to our token,
	// and deletes it if they match. This prevents deleting a lock acquired by another client.
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end`

	// We use Eval to run the script. It is wrapped in our client's retry logic for network errors.
	res, err := l.client.Eval(ctx, script, []string{l.key}, l.token)
	if err != nil {
		return fmt.Errorf("failed to execute lock release script: %w", err)
	}

	// The script returns 1 if the key was deleted, 0 otherwise.
	if val, ok := res.(int64); !ok || val == 0 {
		return ErrLockNotAcquired
	}

	return nil
}
