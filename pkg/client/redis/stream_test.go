package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateStreamIfNotExists verifies that a stream can be created with a TTL.
func TestCreateStreamIfNotExists(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	streamKey := "test:stream:create"

	// Create the stream
	err := testClient.CreateStreamIfNotExists(ctx, streamKey, 1*time.Minute)
	require.NoError(t, err)

	// Verify the stream exists by checking its length (it will have one "init" message)
	length, err := testClient.XLen(ctx, streamKey)
	require.NoError(t, err)
	assert.Equal(t, int64(1), length)

	// Verify TTL is set
	ttl, err := testClient.TTL(ctx, streamKey)
	require.NoError(t, err)
	assert.Greater(t, ttl, 30*time.Second) // Should be close to 1 minute

	// Call it again, should be a no-op
	err = testClient.CreateStreamIfNotExists(ctx, streamKey, 1*time.Minute)
	require.NoError(t, err)
	length, err = testClient.XLen(ctx, streamKey)
	require.NoError(t, err)
	assert.Equal(t, int64(1), length, "Stream should not be modified on second call")
}

// TestCreateConsumerGroup verifies the standard consumer group creation.
func TestCreateConsumerGroup(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	streamKey := "test:stream:group"
	groupName := "test-group"

	// Create the group (and stream since it doesn't exist)
	err := testClient.CreateConsumerGroup(ctx, streamKey, groupName)
	require.NoError(t, err)

	// Verify group exists using the underlying client's XInfoGroups
	groups, err := testClient.Client().XInfoGroups(ctx, streamKey).Result()
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, groupName, groups[0].Name)

	// Call again, should not fail
	err = testClient.CreateConsumerGroup(ctx, streamKey, groupName)
	require.NoError(t, err, "Creating an existing group should not return an error")
}

// TestCreateConsumerGroupAtomic verifies the atomic version of group creation.
func TestCreateConsumerGroupAtomic(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	streamKey := "test:stream:group-atomic"
	groupName := "atomic-group"

	// First creation should return true
	created, err := testClient.CreateConsumerGroupAtomic(ctx, streamKey, groupName)
	require.NoError(t, err)
	assert.True(t, created, "Expected group to be created")

	// Second creation should return false
	created, err = testClient.CreateConsumerGroupAtomic(ctx, streamKey, groupName)
	require.NoError(t, err)
	assert.False(t, created, "Expected group to already exist")
}

// TestCreateStreamWithConsumerGroup verifies the combined creation method.
func TestCreateStreamWithConsumerGroup(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	streamKey := "test:stream:combo"
	groupName := "combo-group"

	err := testClient.CreateStreamWithConsumerGroup(ctx, streamKey, groupName, 2*time.Minute)
	require.NoError(t, err)

	// Verify stream and group exist
	groups, err := testClient.Client().XInfoGroups(ctx, streamKey).Result()
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, groupName, groups[0].Name)

	// Verify TTL
	ttl, err := testClient.TTL(ctx, streamKey)
	require.NoError(t, err)
	assert.Greater(t, ttl, 1*time.Minute, "TTL should be greater than 1 minute")
}

// TestStreamWorkflow covers the full lifecycle: Add, ReadGroup, Ack, Pending, Claim.
func TestStreamWorkflow(t *testing.T) {
	flushDB(t)
	ctx := context.Background()
	streamKey := "test:stream:workflow"
	groupName := "workflow-group"
	consumer1 := "consumer-1"
	consumer2 := "consumer-2"

	// 1. Create stream and group
	err := testClient.CreateConsumerGroup(ctx, streamKey, groupName)
	require.NoError(t, err)

	// 2. XAdd: Add some messages to the stream
	msg1ID, err := testClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{"event": "user_registered", "user_id": "123"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, msg1ID)

	msg2ID, err := testClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{"event": "order_placed", "order_id": "456"},
	})
	require.NoError(t, err)

	// 3. XLen: Verify stream length
	length, err := testClient.XLen(ctx, streamKey)
	require.NoError(t, err)
	assert.Equal(t, int64(2), length)

	// 4. XReadGroup: Consumer 1 reads one message
	readResult, err := testClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumer1,
		Streams:  []string{streamKey, ">"}, // Read new messages
		Count:    1,
		Block:    100 * time.Millisecond,
	})
	require.NoError(t, err)
	require.Len(t, readResult, 1, "Should have read one stream")
	require.Len(t, readResult[0].Messages, 1, "Should have one message in the stream result")
	assert.Equal(t, msg1ID, readResult[0].Messages[0].ID)
	assert.Equal(t, "user_registered", readResult[0].Messages[0].Values["event"])

	// 5. XPending: Check for pending messages (we haven't acknowledged yet)
	pending, err := testClient.XPending(ctx, streamKey, groupName)
	require.NoError(t, err)
	assert.Equal(t, int64(1), pending.Count, "Should be one message pending")
	assert.Equal(t, msg1ID, pending.Lower, "The pending message ID should match")

	// 6. XAck: Consumer 1 acknowledges the message
	err = testClient.XAck(ctx, streamKey, groupName, msg1ID)
	require.NoError(t, err)

	// 7. XPendingExt: Check pending messages again, should be empty now
	pendingExt, err := testClient.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamKey,
		Group:  groupName,
		Start:  "-",
		End:    "+",
		Count:  10,
	})
	require.NoError(t, err)
	assert.Empty(t, pendingExt, "Should be no more pending messages")

	// 8. XClaim: Test claiming a stale message
	// First, consumer 1 reads the next message but doesn't ack it
	readResult, err = testClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: groupName, Consumer: consumer1, Streams: []string{streamKey, ">"}, Count: 1,
	})
	require.NoError(t, err)
	require.Len(t, readResult[0].Messages, 1)
	staleMsgID := readResult[0].Messages[0].ID
	assert.Equal(t, msg2ID, staleMsgID)

	// Now, consumer 2 claims it
	claimResult := testClient.XClaim(ctx, &redis.XClaimArgs{
		Stream:   streamKey,
		Group:    groupName,
		Consumer: consumer2,
		MinIdle:  0, // Claim immediately for test purposes
		Messages: []string{staleMsgID},
	})
	require.NoError(t, claimResult.Err())

	claimedMsgs, err := claimResult.Result()
	require.NoError(t, err)
	require.Len(t, claimedMsgs, 1)
	assert.Equal(t, staleMsgID, claimedMsgs[0].ID)

	// Verify consumer 2 now owns the message by checking XPendingExt
	pendingExt, err = testClient.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamKey, Group: groupName, Start: "-", End: "+", Count: 10,
	})
	require.NoError(t, err)
	require.Len(t, pendingExt, 1)
	assert.Equal(t, consumer2, pendingExt[0].Consumer, "Consumer 2 should now own the message")
}
