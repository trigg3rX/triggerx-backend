package cache

import (
	"context"
	"time"

	redisclient "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type RedisCache struct{}

func (r *RedisCache) Get(key string) (string, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	val, err := redisclient.GetClient().Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			logging.GetServiceLogger().Warnf("[RedisCache] Cache miss for key: %s (%.2fms)", key, float64(time.Since(start).Microseconds())/1000)
		} else {
			logging.GetServiceLogger().Errorf("[RedisCache] Error getting key %s: %v (%.2fms)", key, err, float64(time.Since(start).Microseconds())/1000)
		}
		return "", err
	}
	logging.GetServiceLogger().Infof("[RedisCache] Cache hit for key: %s (%.2fms)", key, float64(time.Since(start).Microseconds())/1000)
	return val, nil
}

func (r *RedisCache) Set(key string, value string, ttl time.Duration) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := redisclient.GetClient().Set(ctx, key, value, ttl).Err()
	if err != nil {
		logging.GetServiceLogger().Errorf("[RedisCache] Error setting key %s: %v (%.2fms)", key, err, float64(time.Since(start).Microseconds())/1000)
	} else {
		logging.GetServiceLogger().Infof("[RedisCache] Set key: %s (ttl: %v, %.2fms)", key, ttl, float64(time.Since(start).Microseconds())/1000)
	}
	return err
}

func (r *RedisCache) Delete(key string) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := redisclient.GetClient().Del(ctx, key).Err()
	if err != nil {
		logging.GetServiceLogger().Errorf("[RedisCache] Error deleting key %s: %v (%.2fms)", key, err, float64(time.Since(start).Microseconds())/1000)
	} else {
		logging.GetServiceLogger().Infof("[RedisCache] Deleted key: %s (%.2fms)", key, float64(time.Since(start).Microseconds())/1000)
	}
	return err
}

func (r *RedisCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID
	res, err := redisclient.GetClient().SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		logging.GetServiceLogger().Errorf("[RedisCache] Error acquiring lock for performerID %s: %v (%.2fms)", performerID, err, float64(time.Since(start).Microseconds())/1000)
	} else {
		logging.GetServiceLogger().Infof("[RedisCache] Acquired lock for performerID: %s (result: %v, ttl: %v, %.2fms)", performerID, res, ttl, float64(time.Since(start).Microseconds())/1000)
	}
	return res, err
}

func (r *RedisCache) ReleasePerformerLock(performerID string) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID
	err := redisclient.GetClient().Del(ctx, key).Err()
	if err != nil {
		logging.GetServiceLogger().Errorf("[RedisCache] Error releasing lock for performerID %s: %v (%.2fms)", performerID, err, float64(time.Since(start).Microseconds())/1000)
	} else {
		logging.GetServiceLogger().Infof("[RedisCache] Released lock for performerID: %s (%.2fms)", performerID, float64(time.Since(start).Microseconds())/1000)
	}
	return err
}
