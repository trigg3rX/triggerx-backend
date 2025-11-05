package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	redisclient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// RewardsCache wraps Redis client for rewards-specific operations
type RewardsCache struct {
	client redisclient.RedisClientInterface
	logger logging.Logger
}

// NewRewardsCache creates a new rewards cache using the existing Redis client
func NewRewardsCache(client redisclient.RedisClientInterface, logger logging.Logger) *RewardsCache {
	return &RewardsCache{
		client: client,
		logger: logger.With("component", "rewards_cache"),
	}
}

// IncrementDailyUptime increments the daily uptime counter for a keeper by seconds
func (c *RewardsCache) IncrementDailyUptime(ctx context.Context, keeperAddress string, seconds int64) error {
	key := fmt.Sprintf("keeper:daily_uptime:%s", keeperAddress)

	// Get current value
	currentVal, err := c.client.Get(ctx, key)
	if err != nil {
		// Key doesn't exist yet, set it to the increment value
		if err := c.client.Set(ctx, key, seconds, 48*time.Hour); err != nil {
			return fmt.Errorf("failed to set initial daily uptime: %w", err)
		}
		return nil
	}

	// Parse current value
	current, err := strconv.ParseInt(currentVal, 10, 64)
	if err != nil {
		c.logger.Warn("Failed to parse current uptime, resetting",
			"keeper", keeperAddress,
			"value", currentVal,
			"error", err)
		current = 0
	}

	// Increment and save
	newValue := current + seconds
	if err := c.client.Set(ctx, key, newValue, 48*time.Hour); err != nil {
		return fmt.Errorf("failed to increment daily uptime: %w", err)
	}

	return nil
}

// GetDailyUptime retrieves the daily uptime for a keeper in seconds
func (c *RewardsCache) GetDailyUptime(ctx context.Context, keeperAddress string) (int64, error) {
	key := fmt.Sprintf("keeper:daily_uptime:%s", keeperAddress)

	val, err := c.client.Get(ctx, key)
	if err != nil {
		// Key doesn't exist, return 0
		return 0, nil
	}

	uptime, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uptime value: %w", err)
	}

	return uptime, nil
}

// GetAllDailyUptimes retrieves daily uptime for all keepers
func (c *RewardsCache) GetAllDailyUptimes(ctx context.Context) (map[string]int64, error) {
	pattern := "keeper:daily_uptime:*"

	// Scan for all daily uptime keys
	keys, err := c.client.ScanAll(ctx, &redisclient.ScanOptions{
		Pattern: pattern,
		Count:   100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan daily uptime keys: %w", err)
	}

	uptimes := make(map[string]int64)
	for _, key := range keys {
		val, err := c.client.Get(ctx, key)
		if err != nil {
			c.logger.Warn("Failed to get uptime for key",
				"key", key,
				"error", err)
			continue
		}

		uptime, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			c.logger.Warn("Failed to parse uptime value",
				"key", key,
				"error", err)
			continue
		}

		// Extract keeper address from key (keeper:daily_uptime:{address})
		address := key[len("keeper:daily_uptime:"):]
		uptimes[address] = uptime
	}

	return uptimes, nil
}

// ResetDailyUptime resets the daily uptime counter for a keeper
func (c *RewardsCache) ResetDailyUptime(ctx context.Context, keeperAddress string) error {
	key := fmt.Sprintf("keeper:daily_uptime:%s", keeperAddress)
	return c.client.Del(ctx, key)
}

// ResetAllDailyUptimes resets all daily uptime counters
func (c *RewardsCache) ResetAllDailyUptimes(ctx context.Context) error {
	pattern := "keeper:daily_uptime:*"

	// Scan for all daily uptime keys
	keys, err := c.client.ScanAll(ctx, &redisclient.ScanOptions{
		Pattern: pattern,
		Count:   100,
	})
	if err != nil {
		return fmt.Errorf("failed to scan daily uptime keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...); err != nil {
			return fmt.Errorf("failed to delete daily uptime keys: %w", err)
		}
		c.logger.Info("Reset all daily uptime counters", "count", len(keys))
	}

	return nil
}

// SetLastRewardsDistribution sets the timestamp of the last rewards distribution
func (c *RewardsCache) SetLastRewardsDistribution(ctx context.Context, timestamp time.Time) error {
	key := "rewards:last_distribution"
	return c.client.Set(ctx, key, timestamp.Format(time.RFC3339), 0)
}

// GetLastRewardsDistribution gets the timestamp of the last rewards distribution
func (c *RewardsCache) GetLastRewardsDistribution(ctx context.Context) (time.Time, error) {
	key := "rewards:last_distribution"

	val, err := c.client.Get(ctx, key)
	if err != nil {
		// Key doesn't exist
		return time.Time{}, nil
	}

	timestamp, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return timestamp, nil
}

// SetCurrentPeriodStart sets the start timestamp of the current rewards period
func (c *RewardsCache) SetCurrentPeriodStart(ctx context.Context, timestamp time.Time) error {
	key := "rewards:current_period_start"
	return c.client.Set(ctx, key, timestamp.Format(time.RFC3339), 0)
}

// GetCurrentPeriodStart gets the start timestamp of the current rewards period
func (c *RewardsCache) GetCurrentPeriodStart(ctx context.Context) (time.Time, error) {
	key := "rewards:current_period_start"

	val, err := c.client.Get(ctx, key)
	if err != nil {
		// Key doesn't exist
		return time.Time{}, nil
	}

	timestamp, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return timestamp, nil
}

// Close closes the underlying Redis connection
func (c *RewardsCache) Close() error {
	return c.client.Close()
}
