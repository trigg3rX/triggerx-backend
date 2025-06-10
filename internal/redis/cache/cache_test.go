package cache

import (
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func setupCacheTestLogger() logging.Logger {
	logConfig := logging.LoggerConfig{
		ProcessName:   "test-cache",
		IsDevelopment: true,
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	return logger
}

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
