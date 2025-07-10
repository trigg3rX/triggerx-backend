package redis

import (
	"context"
	"fmt"
	"time"
)

// GetHealthStatus returns detailed health status of the Redis connection
func (c *Client) GetHealthStatus(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		Connected:  false,
		LastPing:   time.Time{},
		Errors:     []string{},
		Type:       "local",
		ServerInfo: make(map[string]interface{}),
	}

	if c.config.IsUpstash {
		status.Type = "upstash"
	}

	// Test ping with latency measurement
	start := time.Now()
	err := c.Ping()
	latency := time.Since(start)

	if err != nil {
		status.Errors = append(status.Errors, err.Error())
		status.Connected = false
	} else {
		status.Connected = true
		status.LastPing = time.Now()
		status.PingLatency = latency
	}

	// Get server info if connected
	if status.Connected {
		if info, err := c.getServerInfo(ctx); err == nil {
			status.ServerInfo = info
		} else {
			status.Errors = append(status.Errors, fmt.Sprintf("failed to get server info: %v", err))
		}
	}

	return status
}

// getServerInfo retrieves Redis server information
func (c *Client) getServerInfo(ctx context.Context) (map[string]interface{}, error) {
	var info map[string]interface{}
	err := c.executeWithRetry(ctx, func() error {
		result, err := c.redisClient.Info(ctx).Result()
		if err != nil {
			return err
		}

		// Parse basic info
		info = map[string]interface{}{
			"raw_info":  result,
			"timestamp": time.Now(),
		}

		// Try to get memory info
		if memInfo, err := c.redisClient.Info(ctx, "memory").Result(); err == nil {
			info["memory_info"] = memInfo
		}

		// Try to get stats
		if statsInfo, err := c.redisClient.Info(ctx, "stats").Result(); err == nil {
			info["stats_info"] = statsInfo
		}

		return nil
	}, "GetServerInfo")

	return info, err
}

// IsHealthy performs a comprehensive health check
func (c *Client) IsHealthy(ctx context.Context) bool {
	status := c.GetHealthStatus(ctx)
	return status.Connected && len(status.Errors) == 0
}

// PerformHealthCheck performs a detailed health check with specific tests
func (c *Client) PerformHealthCheck(ctx context.Context) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	// Test 1: Basic ping
	start := time.Now()
	pingErr := c.Ping()
	pingLatency := time.Since(start)
	results["ping"] = map[string]interface{}{
		"success": pingErr == nil,
		"latency": pingLatency,
		"error":   nil,
	}
	if pingErr != nil {
		results["ping"].(map[string]interface{})["error"] = pingErr.Error()
	}

	// Test 2: Set/Get operation
	testKey := fmt.Sprintf("health_check_%d", time.Now().Unix())
	testValue := "health_test_value"

	start = time.Now()
	setErr := c.Set(ctx, testKey, testValue, 30*time.Second)
	setLatency := time.Since(start)
	results["set"] = map[string]interface{}{
		"success": setErr == nil,
		"latency": setLatency,
		"error":   nil,
	}
	if setErr != nil {
		results["set"].(map[string]interface{})["error"] = setErr.Error()
	}

	// Test 3: Get operation
	if setErr == nil {
		start = time.Now()
		getValue, getErr := c.Get(ctx, testKey)
		getLatency := time.Since(start)
		results["get"] = map[string]interface{}{
			"success":     getErr == nil && getValue == testValue,
			"latency":     getLatency,
			"error":       nil,
			"value_match": getValue == testValue,
		}
		if getErr != nil {
			results["get"].(map[string]interface{})["error"] = getErr.Error()
		}

		// Cleanup
		c.Del(ctx, testKey)
	}

	// Test 4: Connection pool status
	stats := c.redisClient.PoolStats()
	results["pool_stats"] = map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}

	// Overall health status
	allHealthy := true
	if pingErr != nil || setErr != nil {
		allHealthy = false
	}

	results["overall"] = map[string]interface{}{
		"healthy":   allHealthy,
		"timestamp": time.Now(),
	}

	return results, nil
}
