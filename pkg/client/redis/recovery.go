package redis

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// connectionRecoveryLoop monitors connection health and attempts recovery
func (c *Client) connectionRecoveryLoop() {
	ticker := time.NewTicker(c.recoveryConfig.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.checkAndRecoverConnection(context.Background())
	}
}

// checkAndRecoverConnection checks connection health and attempts recovery if needed
func (c *Client) checkAndRecoverConnection(ctx context.Context) {
	c.mu.Lock()
	if c.isRecovering {
		c.mu.Unlock()
		return // Already in recovery mode
	}
	c.mu.Unlock()

	// Quick health check
	if err := c.Ping(ctx); err != nil {
		c.logger.Warn(ctx, "Redis connection unhealthy, starting recovery", observability.String("error", fmt.Sprintf("%v", err)))
		c.mu.Lock()
		c.isRecovering = true
		c.mu.Unlock()
		go c.performConnectionRecovery()
	} else {
		c.mu.Lock()
		c.lastHealthCheck = time.Now()
		c.mu.Unlock()
	}
}

// performConnectionRecovery attempts to recover the Redis connection
func (c *Client) performConnectionRecovery() {
	start := time.Now()
	c.trackRecoveryStart("connection_failed")

	defer func() {
		c.mu.Lock()
		c.isRecovering = false
		c.mu.Unlock()
	}()

	config := c.recoveryConfig
	backoff := time.Second

	ctx := context.Background()

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Info(ctx, "Redis connection recovery attempt %d/%d after %v", observability.Int("attempt", attempt+1), observability.Int("max_retries", config.MaxRetries), observability.Duration("backoff", backoff))
			time.Sleep(backoff)
		}

		// Try to recreate connection
		if err := c.recreateConnection(ctx); err != nil {
			c.logger.Error(ctx, "Redis connection recovery attempt %d failed: %v", observability.Int("attempt", attempt+1), observability.String("error", fmt.Sprintf("%v", err)))

			// Exponential backoff with jitter
			backoff = time.Duration(float64(backoff) * config.BackoffFactor)
			if backoff > config.MaxBackoffDelay {
				backoff = config.MaxBackoffDelay
			}
			continue
		}

		// Test the new connection
		if err := c.CheckConnection(ctx); err != nil {
			c.logger.Error(ctx, "Redis connection recovery test failed", observability.String("error", fmt.Sprintf("%v", err)))
			continue
		}

		c.logger.Info(ctx, "Redis connection recovery successful after %d attempts", observability.Int("attempt", attempt+1))
		c.mu.Lock()
		c.lastHealthCheck = time.Now()
		c.mu.Unlock()

		// Track successful recovery
		duration := time.Since(start)
		c.trackRecoveryEnd(true, attempt+1, duration)
		return
	}

	c.logger.Error(ctx, "Redis connection recovery failed after %d attempts", observability.Int("max_retries", config.MaxRetries))

	// Track failed recovery
	duration := time.Since(start)
	c.trackRecoveryEnd(false, config.MaxRetries, duration)
}

// recreateConnection recreates the Redis connection
func (c *Client) recreateConnection(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close existing connection
	if c.redisClient != nil {
		err := c.redisClient.Close()
		if err != nil {
			c.logger.Error(ctx, "Failed to close Redis client", observability.String("error", fmt.Sprintf("%v", err)))
		}
	}

	// Parse config and create new connection
	opt, err := parseRedisConfig(c.config)
	if err != nil {
		return fmt.Errorf("failed to parse Redis configuration: %w", err)
	}

	c.redisClient = redis.NewClient(opt)
	return nil
}

// GetConnectionStatus returns the current connection status in a strongly-typed struct.
func (c *Client) GetConnectionStatus() *ConnectionStatus {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats := c.redisClient.PoolStats()
	return &ConnectionStatus{
		IsRecovering:     c.isRecovering,
		LastHealthCheck:  c.lastHealthCheck,
		RecoveryEnabled:  c.recoveryConfig.Enabled,
		RecoveryInterval: c.recoveryConfig.CheckInterval,
		PoolStats: PoolHealthStats{
			Hits:       stats.Hits,
			Misses:     stats.Misses,
			Timeouts:   stats.Timeouts,
			TotalConns: stats.TotalConns,
			IdleConns:  stats.IdleConns,
			StaleConns: stats.StaleConns,
		},
	}
}
