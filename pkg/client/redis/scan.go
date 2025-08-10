package redis

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

// Scan performs a cursor-based scan of keys matching a pattern.
// This is a safe alternative to the KEYS command that doesn't block the Redis server.
// It returns keys in batches, allowing for non-blocking iteration through large keyspaces.
func (c *Client) Scan(ctx context.Context, cursor uint64, options *ScanOptions) (*ScanResult, error) {
	var result *ScanResult
	err := c.executeWithRetry(ctx, func() error {
		var cmd *redis.ScanCmd

		if options == nil {
			options = &ScanOptions{}
		}

		// Use SCAN with pattern and count
		cmd = c.redisClient.Scan(ctx, cursor, options.Pattern, options.Count)

		// If type filter is specified, use SCAN with TYPE option
		if options.Type != "" {
			cmd = c.redisClient.ScanType(ctx, cursor, options.Pattern, options.Count, options.Type)
		}

		keys, nextCursor, err := cmd.Result()
		if err != nil {
			return err
		}

		result = &ScanResult{
			Cursor:  nextCursor,
			Keys:    keys,
			HasMore: nextCursor != 0,
		}

		return nil
	}, "Scan")

	return result, err
}

// ScanAll performs a complete scan of all keys matching a pattern.
// This method iterates through all keys using the cursor-based SCAN command
// and returns all matching keys. Use with caution for large keyspaces.
func (c *Client) ScanAll(ctx context.Context, options *ScanOptions) ([]string, error) {
	var allKeys []string
	var cursor uint64 = 0

	for {
		result, err := c.Scan(ctx, cursor, options)
		if err != nil {
			return nil, err
		}

		allKeys = append(allKeys, result.Keys...)

		if !result.HasMore {
			break
		}

		cursor = result.Cursor
	}

	return allKeys, nil
}

// ScanKeysByPattern scans keys matching a specific pattern with a default count.
// This is a convenience method for common pattern-based scanning.
func (c *Client) ScanKeysByPattern(ctx context.Context, pattern string, count int64) (*ScanResult, error) {
	options := &ScanOptions{
		Pattern: pattern,
		Count:   count,
	}
	return c.Scan(ctx, 0, options)
}

// ScanKeysByType scans keys of a specific type (string, list, set, zset, hash, stream).
// This is useful for finding keys of a particular data structure type.
func (c *Client) ScanKeysByType(ctx context.Context, keyType string, count int64) (*ScanResult, error) {
	options := &ScanOptions{
		Pattern: "*",
		Count:   count,
		Type:    keyType,
	}
	return c.Scan(ctx, 0, options)
}
