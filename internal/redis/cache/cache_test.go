package cache

import (
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func setupCacheTestLogger() logging.Logger {
	logConfig := logging.LoggerConfig{
		LogDir:          "data/logs",
		ProcessName:     "test-cache",
		Environment:     logging.Development,
		UseColors:       false,
		MinStdoutLevel:  logging.InfoLevel,
		MinFileLogLevel: logging.InfoLevel,
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		return &testCacheLogger{}
	}

	return logging.GetServiceLogger()
}

type testCacheLogger struct{}

func (l *testCacheLogger) Debug(msg string, fields ...interface{})                {}
func (l *testCacheLogger) Info(msg string, fields ...interface{})                 {}
func (l *testCacheLogger) Warn(msg string, fields ...interface{})                 {}
func (l *testCacheLogger) Error(msg string, fields ...interface{})                {}
func (l *testCacheLogger) Fatal(msg string, fields ...interface{})                {}
func (l *testCacheLogger) Debugf(format string, args ...interface{})              {}
func (l *testCacheLogger) Infof(format string, args ...interface{})               {}
func (l *testCacheLogger) Warnf(format string, args ...interface{})               {}
func (l *testCacheLogger) Errorf(format string, args ...interface{})              {}
func (l *testCacheLogger) Fatalf(format string, args ...interface{})              {}
func (l *testCacheLogger) With(fields ...interface{}) logging.Logger              { return l }
func (l *testCacheLogger) WithField(key string, value interface{}) logging.Logger { return l }

func TestCacheStoreInitialization(t *testing.T) {
	logger := setupCacheTestLogger()

	err := config.Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	err = InitCacheWithLogger(logger)
	if err != nil {
		t.Errorf("Cache initialization failed: %v", err)
	}

	cache, err := GetCacheStore()
	if err != nil {
		t.Fatalf("Failed to get cache instance: %v", err)
	}

	if cache == nil {
		t.Fatal("Cache instance is nil")
	}

	testKey := "test_key"
	testValue := "test_value"
	ttl := 1 * time.Minute

	err = cache.Set(testKey, testValue, ttl)
	if err != nil {
		t.Errorf("Failed to set cache value: %v", err)
	}

	retrievedValue, err := cache.Get(testKey)
	if err != nil {
		t.Errorf("Failed to get cache value: %v", err)
	}

	if retrievedValue != testValue {
		t.Errorf("Expected %s, got %s", testValue, retrievedValue)
	}

	err = cache.Delete(testKey)
	if err != nil {
		t.Errorf("Failed to delete cache value: %v", err)
	}

	_, err = cache.Get(testKey)
	if err == nil {
		t.Error("Expected error when getting deleted key, but got none")
	}
}

func TestPerformerLockSystem(t *testing.T) {
	logger := setupCacheTestLogger()

	err := config.Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	err = InitCacheWithLogger(logger)
	if err != nil {
		t.Errorf("Cache initialization failed: %v", err)
	}

	cache, err := GetCacheStore()
	if err != nil {
		t.Fatalf("Failed to get cache instance: %v", err)
	}

	performerID := "test_performer"
	lockTTL := 30 * time.Second

	acquired, err := cache.AcquirePerformerLock(performerID, lockTTL)
	if err != nil {
		t.Errorf("Failed to acquire performer lock: %v", err)
	}

	if !acquired {
		t.Error("Expected to acquire lock, but failed")
	}

	acquired2, err := cache.AcquirePerformerLock(performerID, lockTTL)
	if err != nil {
		t.Errorf("Unexpected error when trying to acquire existing lock: %v", err)
	}

	if acquired2 {
		t.Error("Expected to fail acquiring existing lock, but succeeded")
	}

	err = cache.ReleasePerformerLock(performerID)
	if err != nil {
		t.Errorf("Failed to release performer lock: %v", err)
	}

	acquired3, err := cache.AcquirePerformerLock(performerID, lockTTL)
	if err != nil {
		t.Errorf("Failed to acquire performer lock after release: %v", err)
	}

	if !acquired3 {
		t.Error("Expected to acquire lock after release, but failed")
	}

	err = cache.ReleasePerformerLock(performerID)
	if err != nil {
		t.Errorf("Failed to clean up performer lock: %v", err)
	}
}
