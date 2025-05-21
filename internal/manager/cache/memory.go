package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// MemoryCache implements the Cache interface using in-memory storage
type MemoryCache struct {
	states    map[int64]*JobState
	mu        sync.RWMutex
	logger    logging.Logger
	statePath string
}

// NewMemoryCache creates a new instance of MemoryCache
func NewMemoryCache(logger logging.Logger, statePath string) (*MemoryCache, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	cache := &MemoryCache{
		states:    make(map[int64]*JobState),
		logger:    logger,
		statePath: statePath,
	}

	if err := cache.LoadState(); err != nil {
		logger.Warnf("Failed to load cache state: %v", err)
	}

	return cache, nil
}

// deepCopyState creates a deep copy of a JobState to prevent data races
func (c *MemoryCache) deepCopyState(state *JobState) *JobState {
	if state == nil {
		return nil
	}

	newState := &JobState{
		Created:      state.Created,
		LastExecuted: state.LastExecuted,
		Status:       state.Status,
		Type:         state.Type,
	}

	if state.Metadata != nil {
		newState.Metadata = make(map[string]interface{}, len(state.Metadata))
		for k, v := range state.Metadata {
			newState.Metadata[k] = v
		}
	}

	return newState
}

// Get retrieves a job's state from the cache
func (c *MemoryCache) Get(ctx context.Context, jobID int64) (*JobState, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	state, exists := c.states[jobID]
	if !exists {
		return nil, fmt.Errorf("no state found for job %d", jobID)
	}

	return c.deepCopyState(state), nil
}

// Set stores a job's state in the cache
func (c *MemoryCache) Set(ctx context.Context, jobID int64, state *JobState) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if state == nil {
		return fmt.Errorf("cannot set nil state")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.states[jobID] = c.deepCopyState(state)
	return c.saveState()
}

// Update updates specific fields in a job's state
func (c *MemoryCache) Update(ctx context.Context, jobID int64, field string, value interface{}) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	state, exists := c.states[jobID]
	if !exists {
		return fmt.Errorf("no state found for job %d", jobID)
	}

	switch field {
	case "last_executed":
		if timestamp, ok := value.(time.Time); ok {
			state.LastExecuted = timestamp
		} else {
			return fmt.Errorf("invalid value type for last_executed: expected time.Time, got %T", value)
		}
	case "status":
		if status, ok := value.(string); ok {
			state.Status = status
		} else {
			return fmt.Errorf("invalid value type for status: expected string, got %T", value)
		}
	default:
		if state.Metadata == nil {
			state.Metadata = make(map[string]interface{})
		}
		state.Metadata[field] = value
	}

	return c.saveState()
}

// Delete removes a job's state from the cache
func (c *MemoryCache) Delete(ctx context.Context, jobID int64) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.states, jobID)
	return c.saveState()
}

// GetAll retrieves all job states from the cache
func (c *MemoryCache) GetAll(ctx context.Context) (map[int64]*JobState, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	states := make(map[int64]*JobState, len(c.states))
	for id, state := range c.states {
		states[id] = c.deepCopyState(state)
	}

	return states, nil
}

// saveState is an internal method for persisting state
func (c *MemoryCache) saveState() error {
	if c.statePath == "" {
		return nil // No persistence requested
	}

	// Ensure directory exists
	dir := filepath.Dir(c.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.Marshal(c.states)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	tempFile := c.statePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	if err := os.Rename(tempFile, c.statePath); err != nil {
		os.Remove(tempFile) // Clean up temp file if rename fails
		return fmt.Errorf("failed to atomically update state file: %w", err)
	}

	return nil
}

// SaveState persists the current cache state to disk
func (c *MemoryCache) SaveState() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.saveState()
}

// LoadState loads the cache state from disk
func (c *MemoryCache) LoadState() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.statePath == "" {
		return nil // No persistence requested
	}

	data, err := os.ReadFile(c.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	newStates := make(map[int64]*JobState)
	if err := json.Unmarshal(data, &newStates); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	c.states = newStates
	return nil
}

// Close cleans up any resources
func (c *MemoryCache) Close() error {
	return c.SaveState() // Save state one last time before closing
}
