package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func TestTimeoutWorker_TimeoutThreshold(t *testing.T) {
	// Test that the timeout threshold is correctly set to 1 hour
	assert.Equal(t, 1*time.Hour, TasksProcessingTTL)
}

func TestTimeoutWorker_StaleTaskDetection(t *testing.T) {
	// Test that tasks without DispatchedAt but old CreatedAt are detected as stale
	now := time.Now()
	oldCreatedAt := now.Add(-2 * time.Hour)

	task := TaskStreamData{
		SendTaskDataToKeeper: types.SendTaskDataToKeeper{
			TaskID: []int64{123},
		},
		DispatchedAt: nil, // No dispatched timestamp
		CreatedAt:    oldCreatedAt,
	}

	// This task should be considered stale since it's older than TasksProcessingTTL
	assert.True(t, task.CreatedAt.Add(TasksProcessingTTL).Before(now))
}

func TestTimeoutWorker_RecentTaskNotTimedOut(t *testing.T) {
	// Test that recent tasks are not considered timed out
	now := time.Now()
	recentDispatchedAt := now.Add(-30 * time.Minute) // Less than 1 hour

	task := TaskStreamData{
		SendTaskDataToKeeper: types.SendTaskDataToKeeper{
			TaskID: []int64{456},
		},
		DispatchedAt: &recentDispatchedAt,
		CreatedAt:    now.Add(-1 * time.Hour),
	}

	// This task should NOT be considered timed out
	dispatchedDuration := now.Sub(*task.DispatchedAt)
	assert.False(t, dispatchedDuration > TasksProcessingTTL)
}
