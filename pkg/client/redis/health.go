package redis

import (
	"context"
	"fmt"
	"time"
)

// GetHealthStatus returns detailed health status of the Redis connection
func (c *Client) GetHealthStatus(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		Connected: false,
		LastPing:  time.Time{},
		Errors:    []string{},
	}

	// Test ping with latency measurement
	start := time.Now()
	err := c.Ping(ctx)
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

// getServerInfo retrieves Redis server information into a strongly-typed struct in one round trip.
func (c *Client) getServerInfo(ctx context.Context) (*ServerInfo, error) {
	var info *ServerInfo
	err := c.executeWithRetry(ctx, func() error {
		// Fetch all info sections at once to minimize network round trips.
		result, err := c.redisClient.Info(ctx, "all").Result()
		if err != nil {
			return err
		}

		info = &ServerInfo{
			RawInfo:   result,
			Timestamp: time.Now(),
		}

		// The full result string already contains memory and stats info,
		// but if you need them separated for the struct, you can still
		// issue specific commands or parse the main string.
		// For efficiency, storing the raw result is often enough.
		// Let's assume for now the RawInfo is the primary goal.
		// If you absolutely need them as separate fields, the original way is fine,
		// but this single 'all' call is faster.

		return nil
	}, "GetServerInfo")

	return info, err
}

// IsHealthy performs a comprehensive health check
func (c *Client) IsHealthy(ctx context.Context) bool {
	status := c.GetHealthStatus(ctx)
	return status.Connected && len(status.Errors) == 0
}

// PerformHealthCheck performs a detailed health check and returns a strongly-typed result.
func (c *Client) PerformHealthCheck(ctx context.Context) (*HealthCheckResult, error) {
	results := &HealthCheckResult{}

	// Test 1: Basic ping
	start := time.Now()
	pingErr := c.Ping(ctx)
	results.Ping.Latency = time.Since(start)
	results.Ping.Success = pingErr == nil
	if pingErr != nil {
		results.Ping.Error = pingErr.Error()
	}

	// Test 2 & 3: Atomic Set/Get/Delete operation using Lua script
	// This reduces network round trips from 3 (Set + Get + Del) to 1 (Lua script)
	testKey := fmt.Sprintf("health_check_%d", time.Now().Unix())
	testValue := "health_test_value"

	start = time.Now()
	setSuccess, getSuccess, valueMatch, cleanupErr := c.performAtomicHealthTest(ctx, testKey, testValue, 30*time.Second)
	results.Set.Latency = time.Since(start)
	results.Set.Success = setSuccess

	results.Get.Latency = time.Since(start) // Same latency as Set since it's atomic
	results.Get.Success = getSuccess
	results.Get.ValueMatch = valueMatch

	if cleanupErr != "" {
		results.Cleanup.Error = cleanupErr
	}

	// Test 4: Connection pool status
	stats := c.redisClient.PoolStats()
	results.PoolStats = PoolHealthStats{
		Hits:       stats.Hits,
		Misses:     stats.Misses,
		Timeouts:   stats.Timeouts,
		TotalConns: stats.TotalConns,
		IdleConns:  stats.IdleConns,
		StaleConns: stats.StaleConns,
	}

	// Overall health status
	results.Overall.Healthy = (pingErr == nil && setSuccess)
	results.Overall.Timestamp = time.Now()

	return results, nil
}

// performAtomicHealthTest performs Set/Get/Delete operations atomically using Lua script
// Returns: setSuccess, getSuccess, valueMatch, cleanupError
func (c *Client) performAtomicHealthTest(ctx context.Context, key, value string, ttl time.Duration) (bool, bool, bool, string) {
	// Lua script that atomically sets a key, gets its value, and deletes it
	// Returns: [set_result, get_result, delete_result]
	// set_result: 1 if key was set, 0 if failed
	// get_result: the value that was retrieved, or nil if failed
	// delete_result: 1 if key was deleted, 0 if failed
	script := `
	local set_result = redis.call("set", KEYS[1], ARGV[1], "ex", ARGV[2])
	if set_result == "OK" then
		local get_result = redis.call("get", KEYS[1])
		local delete_result = redis.call("del", KEYS[1])
		return {1, get_result, delete_result}
	else
		return {0, nil, 0}
	end`

	ttlSeconds := int(ttl.Seconds())

	result, err := c.Eval(ctx, script, []string{key}, value, ttlSeconds)
	if err != nil {
		return false, false, false, err.Error()
	}

	// Parse the result array
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) >= 3 {
		setSuccess := resultArray[0] == int64(1)
		getValue := resultArray[1]
		deleteSuccess := resultArray[2] == int64(1)

		getSuccess := getValue != nil
		valueMatch := getSuccess && getValue == value

		var cleanupErr string
		if !deleteSuccess {
			cleanupErr = "failed to delete test key"
		}

		return setSuccess, getSuccess, valueMatch, cleanupErr
	}

	return false, false, false, "invalid script result format"
}
