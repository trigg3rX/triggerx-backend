package cache

import (
	"context"
	"sync"
	"time"
)

type cache struct {
	mu    sync.RWMutex
	store map[int64]*JobState
}

// NewCache creates a new instance of the cache
func NewCache() Cache {
	return &cache{
		store: make(map[int64]*JobState),
	}
}

func (c *cache) Set(ctx context.Context, jobID int64, state *JobState) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[jobID] = state
	return nil
}

func (c *cache) Get(ctx context.Context, jobID int64) (*JobState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if state, exists := c.store[jobID]; exists {
		return state, nil
	}
	return nil, ErrNotFound
}

func (c *cache) Update(ctx context.Context, jobID int64, field string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, exists := c.store[jobID]
	if !exists {
		return ErrNotFound
	}

	switch field {
	case "last_executed":
		if t, ok := value.(time.Time); ok {
			state.LastExecuted = t
		}
	case "status":
		if s, ok := value.(string); ok {
			state.Status = s
		}
	default:
		if state.Metadata == nil {
			state.Metadata = make(map[string]interface{})
		}
		state.Metadata[field] = value
	}

	return nil
}

func (c *cache) Delete(ctx context.Context, jobID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, jobID)
	return nil
}

func (c *cache) GetAll(ctx context.Context) (map[int64]*JobState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[int64]*JobState)
	for id, state := range c.store {
		result[id] = state
	}
	return result, nil
}

func (c *cache) SaveState() error {
	// No-op for in-memory cache
	return nil
}

func (c *cache) LoadState() error {
	// No-op for in-memory cache
	return nil
}

func (c *cache) Close() error {
	// No-op for in-memory cache
	return nil
}
