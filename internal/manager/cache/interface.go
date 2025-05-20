package cache

import (
	"context"
	"errors"
	"time"
)

// Common errors
var (
	ErrNotFound = errors.New("job not found in cache")
)

// Common job status values
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
)

// Common job types
const (
	TypeScheduled = "scheduled"
	TypeEvent     = "event"
	TypeCondition = "condition"
)

// JobState represents the state of a job in the cache
type JobState struct {
	// Created is the timestamp when the job was first created
	Created time.Time `json:"created"`

	// LastExecuted is the timestamp of the most recent execution attempt
	LastExecuted time.Time `json:"last_executed"`

	// Status represents the current job status (pending, running, completed, failed, canceled)
	Status string `json:"status"`

	// Type indicates the job trigger type (scheduled, event, condition)
	Type string `json:"type"`

	// Metadata stores additional job-specific data
	Metadata map[string]interface{} `json:"metadata"`
}

// Cache defines the interface for job state caching implementations.
// All implementations must be thread-safe and handle concurrent access.
type Cache interface {
	// Get retrieves a job's state from the cache.
	// Returns ErrNotFound if the job doesn't exist.
	// The returned JobState should be treated as immutable.
	Get(ctx context.Context, jobID int64) (*JobState, error)

	// Set stores a job's state in the cache.
	// If the job already exists, its state is completely replaced.
	// The provided state is copied to prevent external modifications.
	Set(ctx context.Context, jobID int64, state *JobState) error

	// Update modifies specific fields in a job's state.
	// Built-in fields (last_executed, status) are type-checked.
	// Custom fields are stored in Metadata.
	// Returns ErrNotFound if the job doesn't exist.
	Update(ctx context.Context, jobID int64, field string, value interface{}) error

	// Delete removes a job's state from the cache.
	// It's not an error if the job doesn't exist.
	Delete(ctx context.Context, jobID int64) error

	// GetAll retrieves all job states from the cache.
	// The returned states should be treated as immutable.
	// Returns an empty map if no jobs exist.
	GetAll(ctx context.Context) (map[int64]*JobState, error)

	// SaveState persists the current cache state to disk.
	// This is a no-op for distributed caches like Redis.
	SaveState() error

	// LoadState loads the cache state from disk.
	// This is a no-op for distributed caches like Redis.
	LoadState() error

	// Close cleans up any resources used by the cache.
	// After Close is called, the cache should not be used.
	Close() error
}
