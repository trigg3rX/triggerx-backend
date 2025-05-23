package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis"
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
			cacheInstance = &RedisCache{}
		} else {
			// Fallback to a FileCache only if it implements the Cache interface
			// Otherwise, set cacheInstance to nil
			var fileCache Cache = nil
			if fc, ok := interface{}(&FileCache{}).(Cache); ok {
				fileCache = fc
			}
			cacheInstance = fileCache
		}
	})
	return err

}

func GetCache() (Cache, error) {
	if cacheInstance == nil {
		return nil, errors.New("cache not initialized")
	}
	return cacheInstance, nil
}