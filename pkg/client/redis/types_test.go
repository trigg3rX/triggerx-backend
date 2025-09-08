package redis

import (
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionSettings_DefaultValues(t *testing.T) {
	settings := ConnectionSettings{}

	// Test that zero values are reasonable defaults
	assert.Equal(t, 0, settings.PoolSize)
	assert.Equal(t, 0, settings.MaxIdleConns)
	assert.Equal(t, 0, settings.MinIdleConns)
	assert.Equal(t, 0, settings.MaxRetries)
	assert.Equal(t, time.Duration(0), settings.DialTimeout)
	assert.Equal(t, time.Duration(0), settings.ReadTimeout)
	assert.Equal(t, time.Duration(0), settings.WriteTimeout)
	assert.Equal(t, time.Duration(0), settings.PoolTimeout)
	assert.Equal(t, time.Duration(0), settings.PingTimeout)
	assert.Equal(t, time.Duration(0), settings.HealthTimeout)
	assert.Equal(t, time.Duration(0), settings.OperationTimeout)
}

func TestConnectionSettings_WithValues(t *testing.T) {
	settings := ConnectionSettings{
		PoolSize:         10,
		MaxIdleConns:     5,
		MinIdleConns:     2,
		MaxRetries:       3,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     10 * time.Second,
		PoolTimeout:      30 * time.Second,
		PingTimeout:      1 * time.Second,
		HealthTimeout:    5 * time.Second,
		OperationTimeout: 30 * time.Second,
	}

	assert.Equal(t, 10, settings.PoolSize)
	assert.Equal(t, 5, settings.MaxIdleConns)
	assert.Equal(t, 2, settings.MinIdleConns)
	assert.Equal(t, 3, settings.MaxRetries)
	assert.Equal(t, 5*time.Second, settings.DialTimeout)
	assert.Equal(t, 10*time.Second, settings.ReadTimeout)
	assert.Equal(t, 10*time.Second, settings.WriteTimeout)
	assert.Equal(t, 30*time.Second, settings.PoolTimeout)
	assert.Equal(t, 1*time.Second, settings.PingTimeout)
	assert.Equal(t, 5*time.Second, settings.HealthTimeout)
	assert.Equal(t, 30*time.Second, settings.OperationTimeout)
}

func TestUpstashConfig_DefaultValues(t *testing.T) {
	config := UpstashConfig{}

	assert.Empty(t, config.URL)
	assert.Empty(t, config.Token)
}

func TestUpstashConfig_WithValues(t *testing.T) {
	config := UpstashConfig{
		URL:   "redis://localhost:6379",
		Token: "test-token",
	}

	assert.Equal(t, "redis://localhost:6379", config.URL)
	assert.Equal(t, "test-token", config.Token)
}

func TestRedisConfig_DefaultValues(t *testing.T) {
	config := RedisConfig{}

	assert.Empty(t, config.UpstashConfig.URL)
	assert.Empty(t, config.UpstashConfig.Token)
	assert.Equal(t, ConnectionSettings{}, config.ConnectionSettings)
}

func TestRedisConfig_WithValues(t *testing.T) {
	config := RedisConfig{
		UpstashConfig: UpstashConfig{
			URL:   "redis://localhost:6379",
			Token: "test-token",
		},
		ConnectionSettings: ConnectionSettings{
			PoolSize:    10,
			MaxRetries:  3,
			DialTimeout: 5 * time.Second,
		},
	}

	assert.Equal(t, "redis://localhost:6379", config.UpstashConfig.URL)
	assert.Equal(t, "test-token", config.UpstashConfig.Token)
	assert.Equal(t, 10, config.ConnectionSettings.PoolSize)
	assert.Equal(t, 3, config.ConnectionSettings.MaxRetries)
	assert.Equal(t, 5*time.Second, config.ConnectionSettings.DialTimeout)
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	require.NotNil(t, config)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 5*time.Second, config.MaxDelay)
	assert.Equal(t, 0.1, config.JitterFactor)
	assert.True(t, config.LogRetryAttempt)
}

func TestServerInfo_DefaultValues(t *testing.T) {
	info := ServerInfo{}

	assert.Empty(t, info.RawInfo)
	assert.True(t, info.Timestamp.IsZero())
	assert.Empty(t, info.MemoryInfo)
	assert.Empty(t, info.StatsInfo)
	assert.Empty(t, info.MemoryInfoError)
	assert.Empty(t, info.StatsInfoError)
}

func TestServerInfo_WithValues(t *testing.T) {
	now := time.Now()
	info := ServerInfo{
		RawInfo:         "redis_version:7.0.0",
		Timestamp:       now,
		MemoryInfo:      "used_memory:123456",
		StatsInfo:       "total_commands_processed:1000",
		MemoryInfoError: "memory error",
		StatsInfoError:  "stats error",
	}

	assert.Equal(t, "redis_version:7.0.0", info.RawInfo)
	assert.Equal(t, now, info.Timestamp)
	assert.Equal(t, "used_memory:123456", info.MemoryInfo)
	assert.Equal(t, "total_commands_processed:1000", info.StatsInfo)
	assert.Equal(t, "memory error", info.MemoryInfoError)
	assert.Equal(t, "stats error", info.StatsInfoError)
}

func TestHealthStatus_DefaultValues(t *testing.T) {
	status := HealthStatus{}

	assert.False(t, status.Connected)
	assert.True(t, status.LastPing.IsZero())
	assert.Equal(t, time.Duration(0), status.PingLatency)
	assert.Nil(t, status.Errors)
	assert.Nil(t, status.ServerInfo)
}

func TestHealthStatus_WithValues(t *testing.T) {
	now := time.Now()
	serverInfo := &ServerInfo{
		RawInfo:   "redis_version:7.0.0",
		Timestamp: now,
	}
	status := HealthStatus{
		Connected:   true,
		LastPing:    now,
		PingLatency: 10 * time.Millisecond,
		Errors:      []string{"error1", "error2"},
		ServerInfo:  serverInfo,
	}

	assert.True(t, status.Connected)
	assert.Equal(t, now, status.LastPing)
	assert.Equal(t, 10*time.Millisecond, status.PingLatency)
	assert.Len(t, status.Errors, 2)
	assert.Equal(t, "error1", status.Errors[0])
	assert.Equal(t, "error2", status.Errors[1])
	assert.Equal(t, serverInfo, status.ServerInfo)
}

func TestConnectionRecoveryConfig_DefaultValues(t *testing.T) {
	config := ConnectionRecoveryConfig{}

	assert.False(t, config.Enabled)
	assert.Equal(t, time.Duration(0), config.CheckInterval)
	assert.Equal(t, 0, config.MaxRetries)
	assert.Equal(t, 0.0, config.BackoffFactor)
	assert.Equal(t, time.Duration(0), config.MaxBackoffDelay)
}

func TestConnectionRecoveryConfig_WithValues(t *testing.T) {
	config := ConnectionRecoveryConfig{
		Enabled:         true,
		CheckInterval:   30 * time.Second,
		MaxRetries:      5,
		BackoffFactor:   2.0,
		MaxBackoffDelay: 5 * time.Minute,
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, 30*time.Second, config.CheckInterval)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.Equal(t, 5*time.Minute, config.MaxBackoffDelay)
}

func TestDefaultConnectionRecoveryConfig(t *testing.T) {
	config := DefaultConnectionRecoveryConfig()

	require.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, 30*time.Second, config.CheckInterval)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.Equal(t, 5*time.Minute, config.MaxBackoffDelay)
}

func TestPoolHealthStats_DefaultValues(t *testing.T) {
	stats := PoolHealthStats{}

	assert.Equal(t, uint32(0), stats.Hits)
	assert.Equal(t, uint32(0), stats.Misses)
	assert.Equal(t, uint32(0), stats.Timeouts)
	assert.Equal(t, uint32(0), stats.TotalConns)
	assert.Equal(t, uint32(0), stats.IdleConns)
	assert.Equal(t, uint32(0), stats.StaleConns)
}

func TestPoolHealthStats_WithValues(t *testing.T) {
	stats := PoolHealthStats{
		Hits:       100,
		Misses:     10,
		Timeouts:   5,
		TotalConns: 20,
		IdleConns:  15,
		StaleConns: 2,
	}

	assert.Equal(t, uint32(100), stats.Hits)
	assert.Equal(t, uint32(10), stats.Misses)
	assert.Equal(t, uint32(5), stats.Timeouts)
	assert.Equal(t, uint32(20), stats.TotalConns)
	assert.Equal(t, uint32(15), stats.IdleConns)
	assert.Equal(t, uint32(2), stats.StaleConns)
}

func TestConnectionStatus_DefaultValues(t *testing.T) {
	status := ConnectionStatus{}

	assert.False(t, status.IsRecovering)
	assert.True(t, status.LastHealthCheck.IsZero())
	assert.False(t, status.RecoveryEnabled)
	assert.Equal(t, time.Duration(0), status.RecoveryInterval)
	assert.Equal(t, PoolHealthStats{}, status.PoolStats)
}

func TestConnectionStatus_WithValues(t *testing.T) {
	now := time.Now()
	poolStats := PoolHealthStats{
		Hits:       100,
		TotalConns: 20,
	}
	status := ConnectionStatus{
		IsRecovering:     true,
		LastHealthCheck:  now,
		RecoveryEnabled:  true,
		RecoveryInterval: 30 * time.Second,
		PoolStats:        poolStats,
	}

	assert.True(t, status.IsRecovering)
	assert.Equal(t, now, status.LastHealthCheck)
	assert.True(t, status.RecoveryEnabled)
	assert.Equal(t, 30*time.Second, status.RecoveryInterval)
	assert.Equal(t, poolStats, status.PoolStats)
}

func TestPingCheckResult_DefaultValues(t *testing.T) {
	result := PingCheckResult{}

	assert.False(t, result.Success)
	assert.Equal(t, time.Duration(0), result.Latency)
	assert.Empty(t, result.Error)
}

func TestPingCheckResult_WithValues(t *testing.T) {
	result := PingCheckResult{
		Success: true,
		Latency: 10 * time.Millisecond,
		Error:   "test error",
	}

	assert.True(t, result.Success)
	assert.Equal(t, 10*time.Millisecond, result.Latency)
	assert.Equal(t, "test error", result.Error)
}

func TestOperationCheckResult_DefaultValues(t *testing.T) {
	result := OperationCheckResult{}

	assert.False(t, result.Success)
	assert.Equal(t, time.Duration(0), result.Latency)
	assert.Empty(t, result.Error)
}

func TestOperationCheckResult_WithValues(t *testing.T) {
	result := OperationCheckResult{
		Success: true,
		Latency: 5 * time.Millisecond,
		Error:   "operation error",
	}

	assert.True(t, result.Success)
	assert.Equal(t, 5*time.Millisecond, result.Latency)
	assert.Equal(t, "operation error", result.Error)
}

func TestGetCheckResult_DefaultValues(t *testing.T) {
	result := GetCheckResult{}

	assert.False(t, result.Success)
	assert.Equal(t, time.Duration(0), result.Latency)
	assert.Empty(t, result.Error)
	assert.False(t, result.ValueMatch)
}

func TestGetCheckResult_WithValues(t *testing.T) {
	result := GetCheckResult{
		Success:    true,
		Latency:    3 * time.Millisecond,
		Error:      "get error",
		ValueMatch: true,
	}

	assert.True(t, result.Success)
	assert.Equal(t, 3*time.Millisecond, result.Latency)
	assert.Equal(t, "get error", result.Error)
	assert.True(t, result.ValueMatch)
}

func TestCleanupCheckResult_DefaultValues(t *testing.T) {
	result := CleanupCheckResult{}

	assert.Empty(t, result.Error)
}

func TestCleanupCheckResult_WithValues(t *testing.T) {
	result := CleanupCheckResult{
		Error: "cleanup error",
	}

	assert.Equal(t, "cleanup error", result.Error)
}

func TestOverallHealth_DefaultValues(t *testing.T) {
	health := OverallHealth{}

	assert.False(t, health.Healthy)
	assert.True(t, health.Timestamp.IsZero())
}

func TestOverallHealth_WithValues(t *testing.T) {
	now := time.Now()
	health := OverallHealth{
		Healthy:   true,
		Timestamp: now,
	}

	assert.True(t, health.Healthy)
	assert.Equal(t, now, health.Timestamp)
}

func TestHealthCheckResult_DefaultValues(t *testing.T) {
	result := HealthCheckResult{}

	assert.False(t, result.Ping.Success)
	assert.False(t, result.Set.Success)
	assert.False(t, result.Get.Success)
	assert.False(t, result.Get.ValueMatch)
	assert.Empty(t, result.Cleanup.Error)
	assert.Equal(t, PoolHealthStats{}, result.PoolStats)
	assert.False(t, result.Overall.Healthy)
	assert.True(t, result.Overall.Timestamp.IsZero())
}

func TestHealthCheckResult_WithValues(t *testing.T) {
	now := time.Now()
	poolStats := PoolHealthStats{
		Hits:       50,
		TotalConns: 10,
	}
	result := HealthCheckResult{
		Ping: PingCheckResult{
			Success: true,
			Latency: 10 * time.Millisecond,
		},
		Set: OperationCheckResult{
			Success: true,
			Latency: 5 * time.Millisecond,
		},
		Get: GetCheckResult{
			Success:    true,
			Latency:    3 * time.Millisecond,
			ValueMatch: true,
		},
		Cleanup: CleanupCheckResult{
			Error: "cleanup error",
		},
		PoolStats: poolStats,
		Overall: OverallHealth{
			Healthy:   true,
			Timestamp: now,
		},
	}

	assert.True(t, result.Ping.Success)
	assert.Equal(t, 10*time.Millisecond, result.Ping.Latency)
	assert.True(t, result.Set.Success)
	assert.Equal(t, 5*time.Millisecond, result.Set.Latency)
	assert.True(t, result.Get.Success)
	assert.Equal(t, 3*time.Millisecond, result.Get.Latency)
	assert.True(t, result.Get.ValueMatch)
	assert.Equal(t, "cleanup error", result.Cleanup.Error)
	assert.Equal(t, poolStats, result.PoolStats)
	assert.True(t, result.Overall.Healthy)
	assert.Equal(t, now, result.Overall.Timestamp)
}

func TestScanOptions_DefaultValues(t *testing.T) {
	options := ScanOptions{}

	assert.Empty(t, options.Pattern)
	assert.Equal(t, int64(0), options.Count)
	assert.Empty(t, options.Type)
}

func TestScanOptions_WithValues(t *testing.T) {
	options := ScanOptions{
		Pattern: "user:*",
		Count:   100,
		Type:    "string",
	}

	assert.Equal(t, "user:*", options.Pattern)
	assert.Equal(t, int64(100), options.Count)
	assert.Equal(t, "string", options.Type)
}

func TestScanResult_DefaultValues(t *testing.T) {
	result := ScanResult{}

	assert.Equal(t, uint64(0), result.Cursor)
	assert.Nil(t, result.Keys)
	assert.False(t, result.HasMore)
}

func TestScanResult_WithValues(t *testing.T) {
	keys := []string{"key1", "key2", "key3"}
	result := ScanResult{
		Cursor:  12345,
		Keys:    keys,
		HasMore: true,
	}

	assert.Equal(t, uint64(12345), result.Cursor)
	assert.Equal(t, keys, result.Keys)
	assert.Len(t, result.Keys, 3)
	assert.Equal(t, "key1", result.Keys[0])
	assert.Equal(t, "key2", result.Keys[1])
	assert.Equal(t, "key3", result.Keys[2])
	assert.True(t, result.HasMore)
}

func TestScanResult_EmptyKeys(t *testing.T) {
	result := ScanResult{
		Cursor:  0,
		Keys:    []string{},
		HasMore: false,
	}

	assert.Equal(t, uint64(0), result.Cursor)
	assert.Empty(t, result.Keys)
	assert.False(t, result.HasMore)
}

func TestPipelineFunc_TypeDefinition(t *testing.T) {
	// This test verifies that PipelineFunc is properly defined as a function type
	// We can't easily test the function type directly, but we can verify it exists
	// by checking that we can assign a function to it

	var pipelineFunc PipelineFunc = func(pipe redis.Pipeliner) error {
		return nil
	}

	// Verify it's not nil
	assert.NotNil(t, pipelineFunc)
}
