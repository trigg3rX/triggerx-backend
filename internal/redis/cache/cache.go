package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// CacheStore interface defines the caching operations
type CacheStore interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Delete(key string) error
	AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error)
	ReleasePerformerLock(performerID string) error
	Close() error
}

var (
	cacheStoreInstance CacheStore
	cacheInitOnce      sync.Once
	cacheStoreLogger   logging.Logger
)

// InitCacheWithLogger initializes the cache system with a specific logger
func InitCacheWithLogger(logger logging.Logger) error {
	var initErr error
	cacheInitOnce.Do(func() {
		cacheStoreLogger = logger

		// Use centralized Redis client creation
		redisClient, err := redis.NewRedisClient(logger)
		if err != nil {
			initErr = fmt.Errorf("failed to create Redis client: %w", err)
			cacheStoreLogger.Errorf("Cache initialization failed: %v", initErr)
			return
		}

		// Test connection
		if err := redisClient.Ping(); err != nil {
			initErr = fmt.Errorf("redis connection test failed: %w", err)
			cacheStoreLogger.Errorf("Cache initialization failed: %v", initErr)
			return
		}

		cacheStoreInstance = &RedisCache{client: redisClient}
		cacheStoreLogger.Infof("Cache initialized with %s Redis", config.GetRedisType())
	})
	return initErr
}

// GetCacheStore returns the initialized cache instance
func GetCacheStore() (CacheStore, error) {
	if cacheStoreInstance == nil {
		return nil, errors.New("cache not initialized - call InitCacheWithLogger() first")
	}
	return cacheStoreInstance, nil
}

// GetCacheStoreInfo returns information about the current cache configuration
func GetCacheStoreInfo() map[string]interface{} {
	info := map[string]interface{}{
		"initialized": cacheStoreInstance != nil,
		"redis_type":  config.GetRedisType(),
	}

	if cacheStoreInstance != nil {
		info["type"] = "redis"
	}
	return info
}
