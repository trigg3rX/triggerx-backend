package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	redisclient "github.com/trigg3rX/triggerx-backend/internal/redis"
)

type RedisCache struct {
	client *redisclient.Client
}

func (r *RedisCache) Get(key string) (string, error) {
	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		val, err := redisclient.GetClient().Get(ctx, key).Result()
		if err == redis.Nil {
			return "", nil
		}
		return val, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Get(ctx, key)
}

func (r *RedisCache) Set(key string, value string, ttl time.Duration) error {
	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return redisclient.GetClient().Set(ctx, key, value, ttl).Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Set(ctx, key, value, ttl)
}

func (r *RedisCache) Delete(key string) error {
	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return redisclient.GetClient().Del(ctx, key).Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Del(ctx, key)
}

func (r *RedisCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID

	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		res, err := redisclient.GetClient().SetNX(ctx, key, "1", ttl).Result()
		return res, err
	}

	// Use enhanced client with SetNX for atomic lock acquisition
	return r.client.SetNX(ctx, key, "1", ttl)
}

func (r *RedisCache) ReleasePerformerLock(performerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID

	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		return redisclient.GetClient().Del(ctx, key).Err()
	}

	return r.client.Del(ctx, key)
}
