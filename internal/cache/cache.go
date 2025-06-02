package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Delete(key string) error
	AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error)
	ReleasePerformerLock(performerID string) error
}

var (
	cacheInstance Cache
	once          sync.Once
	logger        logging.Logger
)

// Init initializes the cache system with cloud-first Redis approach
func Init() error {
	return InitWithLogger(logging.GetServiceLogger())
}

// InitWithLogger initializes the cache system with a specific logger
func InitWithLogger(log logging.Logger) error {
	var err error
	once.Do(func() {
		logger = log

		// Use enhanced Redis availability check
		if redis.IsAvailable() {
			// Try to create enhanced Redis client
			redisClient, redisErr := redis.NewClient(logger)
			if redisErr == nil {
				// Test connection
				if pingErr := redisClient.Ping(); pingErr == nil {
					cacheInstance = &RedisCache{client: redisClient}
					logger.Infof("Cache initialized with Redis (%s)", getRedisType())
					return
				} else {
					logger.Warnf("Redis ping failed: %v", pingErr)
				}
			} else {
				logger.Warnf("Failed to create Redis client: %v", redisErr)
			}
		}

		// Fallback to FileCache with graceful degradation
		cacheInstance = &FileCache{}
		logger.Warn("Cache initialized with file-based fallback (Redis unavailable)")
	})
	return err
}

// GetCache returns the initialized cache instance
func GetCache() (Cache, error) {
	if cacheInstance == nil {
		return nil, errors.New("cache not initialized - call Init() first")
	}
	return cacheInstance, nil
}

// IsRedisAvailable returns true if Redis cache is being used
func IsRedisAvailable() bool {
	if cacheInstance == nil {
		return false
	}
	_, isRedis := cacheInstance.(*RedisCache)
	return isRedis
}

// GetCacheInfo returns information about the current cache configuration
func GetCacheInfo() map[string]interface{} {
	info := map[string]interface{}{
		"initialized":     cacheInstance != nil,
		"redis_available": redis.IsAvailable(),
	}

	if cacheInstance != nil {
		if _, isRedis := cacheInstance.(*RedisCache); isRedis {
			info["type"] = "redis"
			info["redis_type"] = getRedisType()
		} else {
			info["type"] = "file"
		}
	} else {
		info["type"] = "none"
	}

	return info
}

// getRedisType returns the type of Redis being used
func getRedisType() string {
	redisInfo := redis.GetRedisInfo()
	if redisType, ok := redisInfo["type"].(string); ok {
		return redisType
	}
	return "unknown"
}
