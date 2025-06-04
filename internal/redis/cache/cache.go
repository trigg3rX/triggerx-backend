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
}

var (
	cacheStoreInstance CacheStore
	cacheInitOnce      sync.Once
	cacheStoreLogger   logging.Logger
)

// InitCacheWithLogger initializes the cache system with a specific logger
func InitCacheWithLogger(logger logging.Logger) error {
	var err error
	cacheInitOnce.Do(func() {
		cacheStoreLogger = logger

		// Try Upstash Redis first (cloud Redis - preferred)
		if config.IsUpstashEnabled() {
			redisClient, redisErr := redis.NewRedisClient(logger)
			if redisErr == nil {
				// Test connection
				if pingErr := redisClient.Ping(); pingErr == nil {
					cacheStoreInstance = &RedisCache{client: redisClient}
					cacheStoreLogger.Infof("Cache initialized with Upstash Redis")
					return
				} else {
					cacheStoreLogger.Warnf("Upstash Redis ping failed: %v", pingErr)
				}
			} else {
				cacheStoreLogger.Warnf("Failed to create Upstash Redis client: %v", redisErr)
			}
		}

		// Fallback to local Redis
		if config.IsLocalRedisEnabled() {
			redisClient, redisErr := redis.NewRedisClient(logger)
			if redisErr == nil {
				// Test connection
				if pingErr := redisClient.Ping(); pingErr == nil {
					cacheStoreInstance = &RedisCache{client: redisClient}
					cacheStoreLogger.Infof("Cache initialized with local Redis")
					return
				} else {
					cacheStoreLogger.Warnf("Local Redis ping failed: %v", pingErr)
				}
			} else {
				cacheStoreLogger.Warnf("Failed to create local Redis client: %v", redisErr)
			}
		}

		// No Redis available - fail initialization
		err = fmt.Errorf("no Redis configuration available - cache initialization failed")
		cacheStoreLogger.Errorf("Cache initialization failed: %v", err)
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

// GetCacheStoreInfo returns information about the current cache configuration
func GetCacheStoreInfo() map[string]interface{} {
	info := map[string]interface{}{
		"initialized":     cacheStoreInstance != nil,
		"upstash_enabled": config.IsUpstashEnabled(),
		"local_enabled":   config.IsLocalRedisEnabled(),
	}

	if cacheStoreInstance != nil {
		if _, isRedis := cacheStoreInstance.(*RedisCache); isRedis {
			info["type"] = "redis"
			if config.IsUpstashEnabled() {
				info["redis_type"] = "upstash"
			} else {
				info["redis_type"] = "local"
			}
		} else {
			info["type"] = "unknown"
		}
	} else {
		info["type"] = "none"
	}

	return info
}
