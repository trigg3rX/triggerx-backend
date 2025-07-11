package websocket

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// ReconnectManager handles WebSocket reconnection logic
type ReconnectManager struct {
	url            string
	logger         logging.Logger
	reconnectCount int
	maxRetries     int
	baseDelay      time.Duration
	maxDelay       time.Duration
	backoffFactor  float64
	jitter         bool
	mu             sync.RWMutex
	isRunning      bool
}

// ReconnectConfig holds reconnection configuration
type ReconnectConfig struct {
	MaxRetries    int           `default:"10"`
	BaseDelay     time.Duration `default:"5s"`
	MaxDelay      time.Duration `default:"300s"` // 5 minutes
	BackoffFactor float64       `default:"2.0"`
	Jitter        bool          `default:"true"`
}

// NewReconnectManager creates a new reconnection manager
func NewReconnectManager(url string, logger logging.Logger) *ReconnectManager {
	return &ReconnectManager{
		url:           url,
		logger:        logger,
		maxRetries:    10,
		baseDelay:     5 * time.Second,
		maxDelay:      5 * time.Minute,
		backoffFactor: 2.0,
		jitter:        true,
	}
}

// NewReconnectManagerWithConfig creates a new reconnection manager with custom config
func NewReconnectManagerWithConfig(url string, config ReconnectConfig, logger logging.Logger) *ReconnectManager {
	rm := NewReconnectManager(url, logger)

	if config.MaxRetries > 0 {
		rm.maxRetries = config.MaxRetries
	}
	if config.BaseDelay > 0 {
		rm.baseDelay = config.BaseDelay
	}
	if config.MaxDelay > 0 {
		rm.maxDelay = config.MaxDelay
	}
	if config.BackoffFactor > 0 {
		rm.backoffFactor = config.BackoffFactor
	}
	rm.jitter = config.Jitter

	return rm
}

// Start begins the reconnection process
func (rm *ReconnectManager) Start(ctx context.Context, connectFunc func() error) {
	rm.mu.Lock()
	if rm.isRunning {
		rm.mu.Unlock()
		return
	}
	rm.isRunning = true
	rm.mu.Unlock()

	defer func() {
		rm.mu.Lock()
		rm.isRunning = false
		rm.mu.Unlock()
	}()

	// Initial connection attempt
	if err := connectFunc(); err != nil {
		rm.logger.Errorf("Initial connection failed: %v", err)
		rm.startReconnectionLoop(ctx, connectFunc)
	} else {
		rm.logger.Info("Initial connection successful")
		rm.resetReconnectCount()
	}
}

// startReconnectionLoop handles the reconnection attempts with exponential backoff
func (rm *ReconnectManager) startReconnectionLoop(ctx context.Context, connectFunc func() error) {
	for {
		select {
		case <-ctx.Done():
			rm.logger.Info("Reconnection cancelled by context")
			return
		default:
		}

		if rm.maxRetries > 0 && rm.GetReconnectCount() >= rm.maxRetries {
			rm.logger.Errorf("Max reconnection attempts (%d) reached for %s", rm.maxRetries, rm.url)
			return
		}

		delay := rm.calculateDelay()
		rm.logger.Warnf("Reconnecting to %s in %v (attempt %d/%d)",
			rm.url, delay, rm.GetReconnectCount()+1, rm.maxRetries)

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		rm.incrementReconnectCount()

		if err := connectFunc(); err != nil {
			rm.logger.Errorf("Reconnection attempt %d failed: %v", rm.GetReconnectCount(), err)
			continue
		}

		rm.logger.Infof("Reconnection successful after %d attempts", rm.GetReconnectCount())
		rm.resetReconnectCount()
		return
	}
}

// calculateDelay calculates the delay for the next reconnection attempt
func (rm *ReconnectManager) calculateDelay() time.Duration {
	attempts := rm.GetReconnectCount()

	// Exponential backoff: baseDelay * (backoffFactor ^ attempts)
	delay := time.Duration(float64(rm.baseDelay) * math.Pow(rm.backoffFactor, float64(attempts)))

	// Cap at maxDelay
	if delay > rm.maxDelay {
		delay = rm.maxDelay
	}

	// Add jitter if enabled (±25% randomization)
	if rm.jitter {
		jitterRange := float64(delay) * 0.25
		jitterOffset := (2.0*time.Now().UnixNano()%int64(jitterRange) - int64(jitterRange)) / int64(time.Nanosecond)
		delay += time.Duration(jitterOffset)

		// Ensure delay is not negative
		if delay < 0 {
			delay = rm.baseDelay
		}
	}

	return delay
}

// GetReconnectCount returns the current reconnection count
func (rm *ReconnectManager) GetReconnectCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.reconnectCount
}

// incrementReconnectCount increments the reconnection counter
func (rm *ReconnectManager) incrementReconnectCount() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.reconnectCount++
}

// resetReconnectCount resets the reconnection counter
func (rm *ReconnectManager) resetReconnectCount() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.reconnectCount = 0
}

// TriggerReconnect manually triggers a reconnection
func (rm *ReconnectManager) TriggerReconnect(ctx context.Context, connectFunc func() error) {
	rm.logger.Info("Manual reconnection triggered")
	go rm.startReconnectionLoop(ctx, connectFunc)
}

// IsRunning returns whether the reconnection manager is currently running
func (rm *ReconnectManager) IsRunning() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.isRunning
}

// UpdateConfig updates the reconnection configuration
func (rm *ReconnectManager) UpdateConfig(config ReconnectConfig) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if config.MaxRetries > 0 {
		rm.maxRetries = config.MaxRetries
	}
	if config.BaseDelay > 0 {
		rm.baseDelay = config.BaseDelay
	}
	if config.MaxDelay > 0 {
		rm.maxDelay = config.MaxDelay
	}
	if config.BackoffFactor > 0 {
		rm.backoffFactor = config.BackoffFactor
	}
	rm.jitter = config.Jitter

	rm.logger.Info("Reconnection configuration updated")
}

// GetConfig returns the current reconnection configuration
func (rm *ReconnectManager) GetConfig() ReconnectConfig {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return ReconnectConfig{
		MaxRetries:    rm.maxRetries,
		BaseDelay:     rm.baseDelay,
		MaxDelay:      rm.maxDelay,
		BackoffFactor: rm.backoffFactor,
		Jitter:        rm.jitter,
	}
}
