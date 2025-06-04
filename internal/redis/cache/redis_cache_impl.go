package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
)

// RedisCache implements the Cache interface using Redis as the backend
type RedisCache struct {
	client *redisx.Client
}	

// Get retrieves a value from Redis cache
func (r *RedisCache) Get(key string) (string, error) {
	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		val, err := redisx.GetRedisClient().Get(ctx, key).Result()
		if err == redis.Nil {
			return "", nil
		}
		return val, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Get(ctx, key)
}

// Set stores a value in Redis cache with TTL
func (r *RedisCache) Set(key string, value string, ttl time.Duration) error {
	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return redisx.GetRedisClient().Set(ctx, key, value, ttl).Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Set(ctx, key, value, ttl)
}

// Delete removes a key from Redis cache
func (r *RedisCache) Delete(key string) error {
	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return redisx.GetRedisClient().Del(ctx, key).Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Del(ctx, key)
}

// AcquirePerformerLock acquires a distributed lock for a performer using Redis SetNX
func (r *RedisCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID

	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		res, err := redisx.GetRedisClient().SetNX(ctx, key, "1", ttl).Result()
		return res, err
	}

	// Use enhanced client with SetNX for atomic lock acquisition
	return r.client.SetNX(ctx, key, "1", ttl)
}

// ReleasePerformerLock releases a performer lock by deleting the key
func (r *RedisCache) ReleasePerformerLock(performerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID

	if r.client == nil {
		// Fallback to legacy client for backward compatibility
		return redisx.GetRedisClient().Del(ctx, key).Err()
	}

	return r.client.Del(ctx, key)
}
