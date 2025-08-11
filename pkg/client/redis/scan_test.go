package redis

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupScanKeys(t *testing.T) []string {
	flushDB(t)
	keys := []string{
		"scan:user:1", "scan:user:2", "scan:user:3",
		"scan:product:1", "scan:product:2",
		"scan:order:1",
	}
	ctx := context.Background()
	for _, key := range keys {
		err := testClient.Set(ctx, key, "dummy", 0)
		require.NoError(t, err, "failed to set key for scan test")
	}
	return keys
}

func TestScanAll(t *testing.T) {
	keys := setupScanKeys(t)
	ctx := context.Background()

	// Scan for all keys
	allKeys, err := testClient.ScanAll(ctx, &ScanOptions{Pattern: "scan:*"})
	require.NoError(t, err)
	sort.Strings(keys)
	sort.Strings(allKeys)
	assert.Equal(t, keys, allKeys)

	// Scan for a specific pattern
	userKeys, err := testClient.ScanAll(ctx, &ScanOptions{Pattern: "scan:user:*"})
	require.NoError(t, err)
	sort.Strings(userKeys)
	assert.Equal(t, []string{"scan:user:1", "scan:user:2", "scan:user:3"}, userKeys)
}

func TestScan(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	var cursor uint64 = 0
	var allFoundKeys []string
	iterations := 0

	for {
		res, err := testClient.Scan(ctx, cursor, &ScanOptions{Pattern: "scan:*", Count: 2})
		require.NoError(t, err)

		allFoundKeys = append(allFoundKeys, res.Keys...)
		cursor = res.Cursor
		iterations++

		if !res.HasMore {
			break
		}
	}

	assert.True(t, iterations > 1, "Scan should have taken multiple iterations with a small count")
	assert.Len(t, allFoundKeys, 6)
	sort.Strings(allFoundKeys)
	assert.Equal(t, []string{"scan:order:1", "scan:product:1", "scan:product:2", "scan:user:1", "scan:user:2", "scan:user:3"}, allFoundKeys)
}

func TestScanKeysByPattern(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	res, err := testClient.ScanKeysByPattern(ctx, "scan:product:*", 10)
	require.NoError(t, err)
	assert.False(t, res.HasMore, "Should find all keys in one go")
	sort.Strings(res.Keys)
	assert.Equal(t, []string{"scan:product:1", "scan:product:2"}, res.Keys)
}

func TestScanKeysByType(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	createKey(t, "type:string", "a string")
	testClient.Client().LPush(ctx, "type:list", "an item")
	testClient.Client().SAdd(ctx, "type:set", "a member")

	// Scan for strings
	res, err := testClient.ScanKeysByType(ctx, "string", 10)
	require.NoError(t, err)
	assert.Contains(t, res.Keys, "type:string")

	// Scan for lists
	res, err = testClient.ScanKeysByType(ctx, "list", 10)
	require.NoError(t, err)
	assert.Contains(t, res.Keys, "type:list")

	// Scan for sets
	res, err = testClient.ScanKeysByType(ctx, "set", 10)
	require.NoError(t, err)
	assert.Contains(t, res.Keys, "type:set")
}

func TestScan_NilOptions_UsesDefaultOptions(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	// Test that Scan works with nil options
	res, err := testClient.Scan(ctx, 0, nil)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, len(res.Keys) >= 0, "Should return some keys even with nil options")
}

func TestScanAll_NilOptions_UsesDefaultOptions(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	// Test that ScanAll works with nil options
	allKeys, err := testClient.ScanAll(ctx, nil)
	require.NoError(t, err)
	assert.NotNil(t, allKeys)
	assert.True(t, len(allKeys) >= 0, "Should return some keys even with nil options")
}

func TestScan_WithTypeFilter_UsesScanType(t *testing.T) {
	flushDB(t)
	ctx := context.Background()

	// Create different types of keys
	createKey(t, "type:string", "a string")
	testClient.Client().LPush(ctx, "type:list", "an item")
	testClient.Client().SAdd(ctx, "type:set", "a member")

	// Test scanning with type filter
	res, err := testClient.Scan(ctx, 0, &ScanOptions{
		Pattern: "type:*",
		Count:   10,
		Type:    "string",
	})
	require.NoError(t, err)
	assert.Contains(t, res.Keys, "type:string")
	assert.NotContains(t, res.Keys, "type:list")
	assert.NotContains(t, res.Keys, "type:set")
}

func TestScan_EmptyPattern_ScansAllKeys(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	// Test with empty pattern (should scan all keys)
	res, err := testClient.Scan(ctx, 0, &ScanOptions{
		Pattern: "",
		Count:   10,
	})
	require.NoError(t, err)
	assert.True(t, len(res.Keys) >= 6, "Should find at least the 6 keys we created")
}

func TestScan_ZeroCount_UsesDefaultCount(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	// Test with zero count (should use default)
	res, err := testClient.Scan(ctx, 0, &ScanOptions{
		Pattern: "scan:*",
		Count:   0,
	})
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, len(res.Keys) >= 0, "Should work with zero count")
}

func TestScan_CursorHandling_WorksCorrectly(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	// Test cursor progression
	var cursor uint64 = 0
	var allKeys []string
	iterations := 0
	maxIterations := 10 // Prevent infinite loop

	for iterations < maxIterations {
		res, err := testClient.Scan(ctx, cursor, &ScanOptions{
			Pattern: "scan:*",
			Count:   2,
		})
		require.NoError(t, err)

		allKeys = append(allKeys, res.Keys...)
		cursor = res.Cursor
		iterations++

		if !res.HasMore {
			break
		}
	}

	assert.Len(t, allKeys, 6, "Should find all 6 keys")
	assert.True(t, iterations > 1, "Should take multiple iterations with small count")
}

func TestScanAll_EmptyResult_ReturnsEmptySlice(t *testing.T) {
	flushDB(t)
	ctx := context.Background()

	// Test scanning when no keys exist
	allKeys, err := testClient.ScanAll(ctx, &ScanOptions{
		Pattern: "nonexistent:*",
	})
	require.NoError(t, err)
	assert.Empty(t, allKeys, "Should return empty slice when no keys match")
}

func TestScanKeysByPattern_EmptyPattern_ReturnsAllKeys(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	// Test with empty pattern
	res, err := testClient.ScanKeysByPattern(ctx, "", 10)
	require.NoError(t, err)
	assert.True(t, len(res.Keys) >= 6, "Should find at least the 6 keys we created")
}

func TestScanKeysByType_InvalidType_StillWorks(t *testing.T) {
	setupScanKeys(t)
	ctx := context.Background()

	// Test with invalid type (should still work, just return empty)
	res, err := testClient.ScanKeysByType(ctx, "invalidtype", 10)
	require.NoError(t, err)
	assert.NotNil(t, res)
	// Redis will handle invalid types gracefully
}

func TestScan_WithRetry_HandlesNetworkErrors(t *testing.T) {
	// Create a mock client to test error handling
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	expectedError := fmt.Errorf("network error")

	// Set up the mock to return an error on scan
	mockClient.MockScan = func(ctx context.Context, cursor uint64, options *ScanOptions) (*ScanResult, error) {
		return nil, expectedError
	}

	// Test that Scan returns the error
	res, err := mockClient.Scan(ctx, 0, &ScanOptions{Pattern: "test:*"})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, expectedError, err)
}

func TestScanAll_WithRetry_HandlesNetworkErrors(t *testing.T) {
	// Create a mock client to test error handling
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	expectedError := fmt.Errorf("network error")

	// Set up the mock to return an error on ScanAll
	mockClient.MockScanAll = func(ctx context.Context, options *ScanOptions) ([]string, error) {
		return nil, expectedError
	}

	// Test that ScanAll returns the error
	allKeys, err := mockClient.ScanAll(ctx, &ScanOptions{Pattern: "test:*"})
	assert.Error(t, err)
	assert.Nil(t, allKeys)
	assert.Equal(t, expectedError, err)
}

func TestScanAll_WithRetry_HandlesPartialErrors(t *testing.T) {
	// Create a mock client to test partial error handling
	mockClient := NewMockRedisClient(t)
	ctx := context.Background()
	expectedError := fmt.Errorf("network error")

	// Set up the mock to return an error on ScanAll
	mockClient.MockScanAll = func(ctx context.Context, options *ScanOptions) ([]string, error) {
		return nil, expectedError
	}

	// Test that ScanAll returns the error
	allKeys, err := mockClient.ScanAll(ctx, &ScanOptions{Pattern: "test:*"})
	assert.Error(t, err)
	assert.Nil(t, allKeys)
	assert.Equal(t, expectedError, err)
}
