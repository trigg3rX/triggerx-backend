package redis

import (
	"errors"
	"sync"
	"time"

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
}

var (
	cacheStoreInstance CacheStore
	cacheInitOnce      sync.Once
	cacheStoreLogger   logging.Logger
)

// InitCache initializes the cache system with Redis-first approach
func InitCache() error {
	return InitCacheWithLogger(logging.GetServiceLogger())
}

// InitCacheWithLogger initializes the cache system with a specific logger
func InitCacheWithLogger(logger logging.Logger) error {
	var err error
	cacheInitOnce.Do(func() {
		cacheStoreLogger = logger

		// Use Redis availability check
		if config.IsRedisAvailable() {
			// Try to create enhanced Redis client
			redisClient, redisErr := NewRedisClient(logger)
			if redisErr == nil {
				// Test connection
				if pingErr := redisClient.Ping(); pingErr == nil {
					cacheStoreInstance = &RedisCache{client: redisClient}
					logger.Infof("Cache initialized with Redis (%s)", config.GetRedisType())
					return
				} else {
					logger.Warnf("Redis ping failed: %v", pingErr)
				}
			} else {
				logger.Warnf("Failed to create Redis client: %v", redisErr)
			}
		}

		// Fallback to FileCache with graceful degradation
		cacheStoreInstance = &FileCache{}
		logger.Warn("Cache initialized with file-based fallback (Redis unavailable)")
	})
	return err
}

// GetCacheStore returns the initialized cache instance
func GetCacheStore() (CacheStore, error) {
	if cacheStoreInstance == nil {
		return nil, errors.New("cache not initialized - call InitCache() first")
	}
	return cacheStoreInstance, nil
}

// IsCacheUsingRedis returns true if Redis cache is being used
func IsCacheUsingRedis() bool {
	if cacheStoreInstance == nil {
		return false
	}
	_, isRedis := cacheStoreInstance.(*RedisCache)
	return isRedis
}

// GetCacheStoreInfo returns information about the current cache configuration
func GetCacheStoreInfo() map[string]interface{} {
	info := map[string]interface{}{
		"initialized":     cacheStoreInstance != nil,
		"redis_available": config.IsRedisAvailable(),
	}

	if cacheStoreInstance != nil {
		if _, isRedis := cacheStoreInstance.(*RedisCache); isRedis {
			info["type"] = "redis"
			info["redis_type"] = config.GetRedisType()
		} else {
			info["type"] = "file"
		}
	} else {
		info["type"] = "none"
	}

	return info
}
