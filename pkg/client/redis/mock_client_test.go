package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// MockRedisClient is a mock implementation of RedisClientInterface for testing.
type MockRedisClient struct {
	t *testing.T

	// Mock functions allow you to override behavior for each method.
	MockCheckConnection               func(ctx context.Context) error
	MockPing                          func(ctx context.Context) error
	MockClose                         func() error
	MockNewLock                       func(key string, ttl time.Duration, retryStrategy *RetryStrategy) (*Lock, error)
	MockExecutePipeline               func(ctx context.Context, fn PipelineFunc) ([]redis.Cmder, error)
	MockEval                          func(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	MockGet                           func(ctx context.Context, key string) (string, error)
	MockGetWithExists                 func(ctx context.Context, key string) (value string, exists bool, err error)
	MockSet                           func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	MockSetNX                         func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	MockDel                           func(ctx context.Context, keys ...string) error
	MockDelWithCount                  func(ctx context.Context, keys ...string) (deletedCount int64, err error)
	MockHSet                          func(ctx context.Context, key string, values ...interface{}) error
	MockHGet                          func(ctx context.Context, key, field string) (string, error)
	MockHGetWithExists                func(ctx context.Context, key, field string) (value string, exists bool, err error)
	MockHDel                          func(ctx context.Context, key string, fields ...string) error
	MockHDelWithCount                 func(ctx context.Context, key string, fields ...string) (deletedCount int64, err error)
	MockScan                          func(ctx context.Context, cursor uint64, options *ScanOptions) (*ScanResult, error)
	MockScanAll                       func(ctx context.Context, options *ScanOptions) ([]string, error)
	MockScanKeysByPattern             func(ctx context.Context, pattern string, count int64) (*ScanResult, error)
	MockScanKeysByType                func(ctx context.Context, keyType string, count int64) (*ScanResult, error)
	MockCreateStreamIfNotExists       func(ctx context.Context, stream string, ttl time.Duration) error
	MockCreateConsumerGroup           func(ctx context.Context, stream, group string) error
	MockCreateConsumerGroupAtomic     func(ctx context.Context, stream, group string) (created bool, err error)
	MockCreateStreamWithConsumerGroup func(ctx context.Context, stream, group string, ttl time.Duration) error
	MockXAdd                          func(ctx context.Context, args *redis.XAddArgs) (string, error)
	MockXLen                          func(ctx context.Context, stream string) (int64, error)
	MockXReadGroup                    func(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error)
	MockXAck                          func(ctx context.Context, stream, group, id string) error
	MockXPending                      func(ctx context.Context, stream, group string) (*redis.XPending, error)
	MockXPendingExt                   func(ctx context.Context, args *redis.XPendingExtArgs) ([]redis.XPendingExt, error)
	MockXClaim                        func(ctx context.Context, args *redis.XClaimArgs) *redis.XMessageSliceCmd
	MockZAdd                          func(ctx context.Context, key string, members ...redis.Z) (int64, error)
	MockZAddWithExists                func(ctx context.Context, key string, members ...redis.Z) (newElements int64, keyExisted bool, err error)
	MockZRevRange                     func(ctx context.Context, key string, start, stop int64) ([]string, error)
	MockZRemRangeByScore              func(ctx context.Context, key, min, max string) (int64, error)
	MockZCard                         func(ctx context.Context, key string) (int64, error)
	MockTTL                           func(ctx context.Context, key string) (time.Duration, error)
	MockRefreshTTL                    func(ctx context.Context, key string, ttl time.Duration) error
	MockRefreshStreamTTL              func(ctx context.Context, stream string, ttl time.Duration) error
	MockSetTTL                        func(ctx context.Context, key string, ttl time.Duration) error
	MockGetTTLStatus                  func(ctx context.Context, key string) (time.Duration, bool, error)
	MockGetHealthStatus               func(ctx context.Context) *HealthStatus
	MockIsHealthy                     func(ctx context.Context) bool
	MockPerformHealthCheck            func(ctx context.Context) (*HealthCheckResult, error)
	MockGetConnectionStatus           func() *ConnectionStatus
	MockGetOperationMetrics           func() map[string]*OperationMetrics
	MockResetOperationMetrics         func()
	MockSetMonitoringHooks            func(hooks *MonitoringHooks)
	MockSetRetryConfig                func(config *RetryConfig)
	MockSetConnectionRecoveryConfig   func(config *ConnectionRecoveryConfig)
	MockClient                        func() *redis.Client
}

// NewMockRedisClient creates a new mock client.
func NewMockRedisClient(t *testing.T) *MockRedisClient {
	return &MockRedisClient{t: t}
}

func (m *MockRedisClient) CheckConnection(ctx context.Context) error {
	if m.MockCheckConnection != nil {
		return m.MockCheckConnection(ctx)
	}
	m.t.Fatal("unexpected call to MockRedisClient.CheckConnection")
	return nil
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	if m.MockPing != nil {
		return m.MockPing(ctx)
	}
	m.t.Fatal("unexpected call to MockRedisClient.Ping")
	return nil
}

func (m *MockRedisClient) Close() error {
	if m.MockClose != nil {
		return m.MockClose()
	}
	m.t.Fatal("unexpected call to MockRedisClient.Close")
	return nil
}

func (m *MockRedisClient) NewLock(key string, ttl time.Duration, retryStrategy *RetryStrategy) (*Lock, error) {
	if m.MockNewLock != nil {
		return m.MockNewLock(key, ttl, retryStrategy)
	}
	m.t.Fatal("unexpected call to MockRedisClient.NewLock")
	return nil, nil
}

func (m *MockRedisClient) ExecutePipeline(ctx context.Context, fn PipelineFunc) ([]redis.Cmder, error) {
	if m.MockExecutePipeline != nil {
		return m.MockExecutePipeline(ctx, fn)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ExecutePipeline")
	return nil, nil
}

func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	if m.MockEval != nil {
		return m.MockEval(ctx, script, keys, args...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.Eval")
	return nil, nil
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.MockGet != nil {
		return m.MockGet(ctx, key)
	}
	m.t.Fatal("unexpected call to MockRedisClient.Get")
	return "", nil
}

func (m *MockRedisClient) GetWithExists(ctx context.Context, key string) (value string, exists bool, err error) {
	if m.MockGetWithExists != nil {
		return m.MockGetWithExists(ctx, key)
	}
	m.t.Fatal("unexpected call to MockRedisClient.GetWithExists")
	return "", false, nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.MockSet != nil {
		return m.MockSet(ctx, key, value, expiration)
	}
	m.t.Fatal("unexpected call to MockRedisClient.Set")
	return nil
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if m.MockSetNX != nil {
		return m.MockSetNX(ctx, key, value, expiration)
	}
	m.t.Fatal("unexpected call to MockRedisClient.SetNX")
	return false, nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	if m.MockDel != nil {
		return m.MockDel(ctx, keys...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.Del")
	return nil
}

func (m *MockRedisClient) DelWithCount(ctx context.Context, keys ...string) (deletedCount int64, err error) {
	if m.MockDelWithCount != nil {
		return m.MockDelWithCount(ctx, keys...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.DelWithCount")
	return 0, nil
}

// Hash operation mock implementations
func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	if m.MockHSet != nil {
		return m.MockHSet(ctx, key, values...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.HSet")
	return nil
}

func (m *MockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	if m.MockHGet != nil {
		return m.MockHGet(ctx, key, field)
	}
	m.t.Fatal("unexpected call to MockRedisClient.HGet")
	return "", nil
}

func (m *MockRedisClient) HGetWithExists(ctx context.Context, key, field string) (value string, exists bool, err error) {
	if m.MockHGetWithExists != nil {
		return m.MockHGetWithExists(ctx, key, field)
	}
	m.t.Fatal("unexpected call to MockRedisClient.HGetWithExists")
	return "", false, nil
}

func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	if m.MockHDel != nil {
		return m.MockHDel(ctx, key, fields...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.HDel")
	return nil
}

func (m *MockRedisClient) HDelWithCount(ctx context.Context, key string, fields ...string) (deletedCount int64, err error) {
	if m.MockHDelWithCount != nil {
		return m.MockHDelWithCount(ctx, key, fields...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.HDelWithCount")
	return 0, nil
}

func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, options *ScanOptions) (*ScanResult, error) {
	if m.MockScan != nil {
		return m.MockScan(ctx, cursor, options)
	}
	m.t.Fatal("unexpected call to MockRedisClient.Scan")
	return nil, nil
}

func (m *MockRedisClient) ScanAll(ctx context.Context, options *ScanOptions) ([]string, error) {
	if m.MockScanAll != nil {
		return m.MockScanAll(ctx, options)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ScanAll")
	return nil, nil
}

func (m *MockRedisClient) ScanKeysByPattern(ctx context.Context, pattern string, count int64) (*ScanResult, error) {
	if m.MockScanKeysByPattern != nil {
		return m.MockScanKeysByPattern(ctx, pattern, count)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ScanKeysByPattern")
	return nil, nil
}

func (m *MockRedisClient) ScanKeysByType(ctx context.Context, keyType string, count int64) (*ScanResult, error) {
	if m.MockScanKeysByType != nil {
		return m.MockScanKeysByType(ctx, keyType, count)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ScanKeysByType")
	return nil, nil
}

func (m *MockRedisClient) CreateStreamIfNotExists(ctx context.Context, stream string, ttl time.Duration) error {
	if m.MockCreateStreamIfNotExists != nil {
		return m.MockCreateStreamIfNotExists(ctx, stream, ttl)
	}
	m.t.Fatal("unexpected call to MockRedisClient.CreateStreamIfNotExists")
	return nil
}

func (m *MockRedisClient) CreateConsumerGroup(ctx context.Context, stream, group string) error {
	if m.MockCreateConsumerGroup != nil {
		return m.MockCreateConsumerGroup(ctx, stream, group)
	}
	m.t.Fatal("unexpected call to MockRedisClient.CreateConsumerGroup")
	return nil
}

func (m *MockRedisClient) CreateConsumerGroupAtomic(ctx context.Context, stream, group string) (created bool, err error) {
	if m.MockCreateConsumerGroupAtomic != nil {
		return m.MockCreateConsumerGroupAtomic(ctx, stream, group)
	}
	m.t.Fatal("unexpected call to MockRedisClient.CreateConsumerGroupAtomic")
	return false, nil
}

func (m *MockRedisClient) CreateStreamWithConsumerGroup(ctx context.Context, stream, group string, ttl time.Duration) error {
	if m.MockCreateStreamWithConsumerGroup != nil {
		return m.MockCreateStreamWithConsumerGroup(ctx, stream, group, ttl)
	}
	m.t.Fatal("unexpected call to MockRedisClient.CreateStreamWithConsumerGroup")
	return nil
}

func (m *MockRedisClient) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	if m.MockXAdd != nil {
		return m.MockXAdd(ctx, args)
	}
	m.t.Fatal("unexpected call to MockRedisClient.XAdd")
	return "", nil
}

func (m *MockRedisClient) XLen(ctx context.Context, stream string) (int64, error) {
	if m.MockXLen != nil {
		return m.MockXLen(ctx, stream)
	}
	m.t.Fatal("unexpected call to MockRedisClient.XLen")
	return 0, nil
}

func (m *MockRedisClient) XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error) {
	if m.MockXReadGroup != nil {
		return m.MockXReadGroup(ctx, args)
	}
	m.t.Fatal("unexpected call to MockRedisClient.XReadGroup")
	return nil, nil
}

func (m *MockRedisClient) XAck(ctx context.Context, stream, group, id string) error {
	if m.MockXAck != nil {
		return m.MockXAck(ctx, stream, group, id)
	}
	m.t.Fatal("unexpected call to MockRedisClient.XAck")
	return nil
}

func (m *MockRedisClient) XPending(ctx context.Context, stream, group string) (*redis.XPending, error) {
	if m.MockXPending != nil {
		return m.MockXPending(ctx, stream, group)
	}
	m.t.Fatal("unexpected call to MockRedisClient.XPending")
	return nil, nil
}

func (m *MockRedisClient) XPendingExt(ctx context.Context, args *redis.XPendingExtArgs) ([]redis.XPendingExt, error) {
	if m.MockXPendingExt != nil {
		return m.MockXPendingExt(ctx, args)
	}
	m.t.Fatal("unexpected call to MockRedisClient.XPendingExt")
	return nil, nil
}

func (m *MockRedisClient) XClaim(ctx context.Context, args *redis.XClaimArgs) *redis.XMessageSliceCmd {
	if m.MockXClaim != nil {
		return m.MockXClaim(ctx, args)
	}
	m.t.Fatal("unexpected call to MockRedisClient.XClaim")
	return nil
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	if m.MockZAdd != nil {
		return m.MockZAdd(ctx, key, members...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ZAdd")
	return 0, nil
}

func (m *MockRedisClient) ZAddWithExists(ctx context.Context, key string, members ...redis.Z) (newElements int64, keyExisted bool, err error) {
	if m.MockZAddWithExists != nil {
		return m.MockZAddWithExists(ctx, key, members...)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ZAddWithExists")
	return 0, false, nil
}

func (m *MockRedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if m.MockZRevRange != nil {
		return m.MockZRevRange(ctx, key, start, stop)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ZRevRange")
	return nil, nil
}

func (m *MockRedisClient) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	if m.MockZRemRangeByScore != nil {
		return m.MockZRemRangeByScore(ctx, key, min, max)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ZRemRangeByScore")
	return 0, nil
}

func (m *MockRedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	if m.MockZCard != nil {
		return m.MockZCard(ctx, key)
	}
	m.t.Fatal("unexpected call to MockRedisClient.ZCard")
	return 0, nil
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	if m.MockTTL != nil {
		return m.MockTTL(ctx, key)
	}
	m.t.Fatal("unexpected call to MockRedisClient.TTL")
	return 0, nil
}

func (m *MockRedisClient) RefreshTTL(ctx context.Context, key string, ttl time.Duration) error {
	if m.MockRefreshTTL != nil {
		return m.MockRefreshTTL(ctx, key, ttl)
	}
	m.t.Fatal("unexpected call to MockRedisClient.RefreshTTL")
	return nil
}

func (m *MockRedisClient) RefreshStreamTTL(ctx context.Context, stream string, ttl time.Duration) error {
	if m.MockRefreshStreamTTL != nil {
		return m.MockRefreshStreamTTL(ctx, stream, ttl)
	}
	m.t.Fatal("unexpected call to MockRedisClient.RefreshStreamTTL")
	return nil
}

func (m *MockRedisClient) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	if m.MockSetTTL != nil {
		return m.MockSetTTL(ctx, key, ttl)
	}
	m.t.Fatal("unexpected call to MockRedisClient.SetTTL")
	return nil
}

func (m *MockRedisClient) GetTTLStatus(ctx context.Context, key string) (time.Duration, bool, error) {
	if m.MockGetTTLStatus != nil {
		return m.MockGetTTLStatus(ctx, key)
	}
	m.t.Fatal("unexpected call to MockRedisClient.GetTTLStatus")
	return 0, false, nil
}

func (m *MockRedisClient) GetHealthStatus(ctx context.Context) *HealthStatus {
	if m.MockGetHealthStatus != nil {
		return m.MockGetHealthStatus(ctx)
	}
	m.t.Fatal("unexpected call to MockRedisClient.GetHealthStatus")
	return nil
}

func (m *MockRedisClient) IsHealthy(ctx context.Context) bool {
	if m.MockIsHealthy != nil {
		return m.MockIsHealthy(ctx)
	}
	m.t.Fatal("unexpected call to MockRedisClient.IsHealthy")
	return false
}

func (m *MockRedisClient) PerformHealthCheck(ctx context.Context) (*HealthCheckResult, error) {
	if m.MockPerformHealthCheck != nil {
		return m.MockPerformHealthCheck(ctx)
	}
	m.t.Fatal("unexpected call to MockRedisClient.PerformHealthCheck")
	return nil, nil
}

func (m *MockRedisClient) GetConnectionStatus() *ConnectionStatus {
	if m.MockGetConnectionStatus != nil {
		return m.MockGetConnectionStatus()
	}
	m.t.Fatal("unexpected call to MockRedisClient.GetConnectionStatus")
	return nil
}

func (m *MockRedisClient) GetOperationMetrics() map[string]*OperationMetrics {
	if m.MockGetOperationMetrics != nil {
		return m.MockGetOperationMetrics()
	}
	m.t.Fatal("unexpected call to MockRedisClient.GetOperationMetrics")
	return nil
}

func (m *MockRedisClient) ResetOperationMetrics() {
	if m.MockResetOperationMetrics != nil {
		m.MockResetOperationMetrics()
	}
	m.t.Fatal("unexpected call to MockRedisClient.ResetOperationMetrics")
}

func (m *MockRedisClient) SetMonitoringHooks(hooks *MonitoringHooks) {
	if m.MockSetMonitoringHooks != nil {
		m.MockSetMonitoringHooks(hooks)
	}
	m.t.Fatal("unexpected call to MockRedisClient.SetMonitoringHooks")
}

func (m *MockRedisClient) SetRetryConfig(config *RetryConfig) {
	if m.MockSetRetryConfig != nil {
		m.MockSetRetryConfig(config)
	}
	m.t.Fatal("unexpected call to MockRedisClient.SetRetryConfig")
}

func (m *MockRedisClient) SetConnectionRecoveryConfig(config *ConnectionRecoveryConfig) {
	if m.MockSetConnectionRecoveryConfig != nil {
		m.MockSetConnectionRecoveryConfig(config)
	}
	m.t.Fatal("unexpected call to MockRedisClient.SetConnectionRecoveryConfig")
}

func (m *MockRedisClient) Client() *redis.Client {
	if m.MockClient != nil {
		return m.MockClient()
	}
	m.t.Fatal("unexpected call to MockRedisClient.Client")
	return nil
}
