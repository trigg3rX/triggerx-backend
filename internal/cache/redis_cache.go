package cache

import (
	"context"
	"time"

	redisclient "github.com/trigg3rX/triggerx-backend/internal/redis"
)

type RedisCache struct{}

func (r *RedisCache) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	val, err := redisclient.GetClient().Get(ctx, key).Result()
	return val, err
}

func (r *RedisCache) Set(key string, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return redisclient.GetClient().Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return redisclient.GetClient().Del(ctx, key).Err()
}

func (r *RedisCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    key := "performer:busy:" + performerID
    // NX = only set if not exists, EX = expire after ttl
    res, err := redisclient.GetClient().SetNX(ctx, key, "1", ttl).Result()
    return res, err
}

func (r *RedisCache) ReleasePerformerLock(performerID string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    key := "performer:busy:" + performerID
    return redisclient.GetClient().Del(ctx, key).Err()
}