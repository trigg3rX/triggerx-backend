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
)

func Init() error {
	var err error
	once.Do(func() {
		err = redis.Ping()
		if err == nil {
			logging.GetServiceLogger().Infof("[Cache] Using RedisCache as backend.")
			cacheInstance = &RedisCache{}
		} else {
			logging.GetServiceLogger().Warnf("[Cache] Redis unavailable, falling back to FileCache: %v", err)
			var fileCache Cache = nil
			if fc, ok := interface{}(&FileCache{}).(Cache); ok {
				fileCache = fc
			}
			cacheInstance = fileCache
		}
	})
	if err != nil {
		logging.GetServiceLogger().Errorf("[Cache] Cache initialization error: %v", err)
	}
	return err
}

func GetCache() (Cache, error) {
	if cacheInstance == nil {
		logging.GetServiceLogger().Errorf("[Cache] Cache not initialized!")
		return nil, errors.New("cache not initialized")
	}
	return cacheInstance, nil
}
