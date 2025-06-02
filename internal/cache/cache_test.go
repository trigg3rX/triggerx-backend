package cache

import (
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func setupTestLogger() logging.Logger {
	// Initialize logger for testing
	logConfig := logging.LoggerConfig{
		LogDir:          "data/logs",
		ProcessName:     "test",
		Environment:     logging.Development,
		UseColors:       false,
		MinStdoutLevel:  logging.InfoLevel,
		MinFileLogLevel: logging.InfoLevel,
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		// Create a simple test logger that implements the Logger interface
		return &testLogger{}
	}

	return logging.GetServiceLogger()
}

// testLogger is a simple logger implementation for testing
type testLogger struct{}

func (l *testLogger) Debug(msg string, fields ...interface{})                {}
func (l *testLogger) Info(msg string, fields ...interface{})                 {}
func (l *testLogger) Warn(msg string, fields ...interface{})                 {}
func (l *testLogger) Error(msg string, fields ...interface{})                {}
func (l *testLogger) Fatal(msg string, fields ...interface{})                {}
func (l *testLogger) Debugf(format string, args ...interface{})              {}
func (l *testLogger) Infof(format string, args ...interface{})               {}
func (l *testLogger) Warnf(format string, args ...interface{})               {}
func (l *testLogger) Errorf(format string, args ...interface{})              {}
func (l *testLogger) Fatalf(format string, args ...interface{})              {}
func (l *testLogger) With(fields ...interface{}) logging.Logger              { return l }
func (l *testLogger) WithField(key string, value interface{}) logging.Logger { return l }

func TestCacheInitialization(t *testing.T) {
	// Initialize logger first
	logger := setupTestLogger()

	// Initialize config for testing (ignore errors in test environment)
	err := config.Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Test cache initialization
	err = InitWithLogger(logger)
	if err != nil {
		t.Errorf("Cache initialization failed: %v", err)
	}

	// Get cache instance
	cache, err := GetCache()
	if err != nil {
		t.Fatalf("Failed to get cache instance: %v", err)
	}

	if cache == nil {
		t.Fatal("Cache instance is nil")
	}

	// Test basic cache operations
	testKey := "test_key"
	testValue := "test_value"
	ttl := 1 * time.Minute

	// Test Set
	err = cache.Set(testKey, testValue, ttl)
	if err != nil {
		t.Errorf("Failed to set cache value: %v", err)
	}

	// Test Get
	retrievedValue, err := cache.Get(testKey)
	if err != nil {
		t.Errorf("Failed to get cache value: %v", err)
	}

	if retrievedValue != testValue {
		t.Errorf("Expected %s, got %s", testValue, retrievedValue)
	}

	// Test Delete
	err = cache.Delete(testKey)
	if err != nil {
		t.Errorf("Failed to delete cache value: %v", err)
	}

	// Verify deletion
	_, err = cache.Get(testKey)
	if err == nil {
		t.Error("Expected error when getting deleted key, but got none")
	}
}

func TestPerformerLock(t *testing.T) {
	// Initialize logger first
	logger := setupTestLogger()

	// Initialize config for testing (ignore errors in test environment)
	err := config.Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Test cache initialization
	err = InitWithLogger(logger)
	if err != nil {
		t.Errorf("Cache initialization failed: %v", err)
	}

	// Get cache instance
	cache, err := GetCache()
	if err != nil {
		t.Fatalf("Failed to get cache instance: %v", err)
	}

	performerID := "test_performer"
	lockTTL := 30 * time.Second

	// Test lock acquisition
	acquired, err := cache.AcquirePerformerLock(performerID, lockTTL)
	if err != nil {
		t.Errorf("Failed to acquire performer lock: %v", err)
	}

	if !acquired {
		t.Error("Expected to acquire lock, but failed")
	}

	// Test lock already exists (should fail to acquire)
	acquired2, err := cache.AcquirePerformerLock(performerID, lockTTL)
	if err != nil {
		t.Errorf("Unexpected error when trying to acquire existing lock: %v", err)
	}

	if acquired2 {
		t.Error("Expected to fail acquiring existing lock, but succeeded")
	}

	// Test lock release
	err = cache.ReleasePerformerLock(performerID)
	if err != nil {
		t.Errorf("Failed to release performer lock: %v", err)
	}

	// Test lock acquisition after release (should succeed)
	acquired3, err := cache.AcquirePerformerLock(performerID, lockTTL)
	if err != nil {
		t.Errorf("Failed to acquire performer lock after release: %v", err)
	}

	if !acquired3 {
		t.Error("Expected to acquire lock after release, but failed")
	}

	// Clean up
	err = cache.ReleasePerformerLock(performerID)
	if err != nil {
		t.Errorf("Failed to clean up performer lock: %v", err)
	}
}

func TestCacheInfo(t *testing.T) {
	// Initialize logger first
	logger := setupTestLogger()

	// Initialize config for testing (ignore errors in test environment)
	err := config.Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Test cache initialization
	err = InitWithLogger(logger)
	if err != nil {
		t.Errorf("Cache initialization failed: %v", err)
	}

	// Test cache info
	info := GetCacheInfo()
	if info == nil {
		t.Fatal("Cache info is nil")
	}

	// Check required fields
	if _, ok := info["initialized"]; !ok {
		t.Error("Cache info missing 'initialized' field")
	}

	if _, ok := info["redis_available"]; !ok {
		t.Error("Cache info missing 'redis_available' field")
	}

	if _, ok := info["type"]; !ok {
		t.Error("Cache info missing 'type' field")
	}

	t.Logf("Cache info: %+v", info)
}
