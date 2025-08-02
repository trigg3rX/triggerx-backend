package redis

import "time"

// MonitoringHooks defines callbacks for monitoring Redis operations
type MonitoringHooks struct {
	OnOperationStart   func(operation string, key string)
	OnOperationEnd     func(operation string, key string, duration time.Duration, err error)
	OnConnectionStatus func(connected bool, latency time.Duration)
	OnRecoveryStart    func(reason string)
	OnRecoveryEnd      func(success bool, attempts int, duration time.Duration)
	OnRetryAttempt     func(operation string, attempt int, err error)
}

// OperationMetrics tracks metrics for Redis operations
type OperationMetrics struct {
	TotalCalls     int64
	TotalDuration  time.Duration
	ErrorCount     int64
	LastError      error
	LastErrorTime  time.Time
	SuccessCount   int64
	RetryCount     int64
	AverageLatency time.Duration
}

// SetMonitoringHooks sets the monitoring hooks for the client
func (c *Client) SetMonitoringHooks(hooks *MonitoringHooks) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.monitoringHooks = hooks
}

// trackOperationStart tracks the start of an operation
func (c *Client) trackOperationStart(operation string, key string) {
	if c.monitoringHooks != nil && c.monitoringHooks.OnOperationStart != nil {
		c.monitoringHooks.OnOperationStart(operation, key)
	}
}

// trackOperationEnd tracks the end of an operation
func (c *Client) trackOperationEnd(operation string, key string, duration time.Duration, err error) {
	// Update internal metrics
	c.mu.Lock()
	if c.operationMetrics == nil {
		c.operationMetrics = make(map[string]*OperationMetrics)
	}

	metrics, exists := c.operationMetrics[operation]
	if !exists {
		metrics = &OperationMetrics{}
		c.operationMetrics[operation] = metrics
	}

	metrics.TotalCalls++
	metrics.TotalDuration += duration
	metrics.AverageLatency = time.Duration(int64(metrics.TotalDuration) / metrics.TotalCalls)

	if err != nil {
		metrics.ErrorCount++
		metrics.LastError = err
		metrics.LastErrorTime = time.Now()
	} else {
		metrics.SuccessCount++
	}
	c.mu.Unlock()

	// Call external monitoring hook
	if c.monitoringHooks != nil && c.monitoringHooks.OnOperationEnd != nil {
		c.monitoringHooks.OnOperationEnd(operation, key, duration, err)
	}
}

// trackConnectionStatus tracks connection status changes
func (c *Client) trackConnectionStatus(connected bool, latency time.Duration) {
	if c.monitoringHooks != nil && c.monitoringHooks.OnConnectionStatus != nil {
		c.monitoringHooks.OnConnectionStatus(connected, latency)
	}
}

// trackRecoveryStart tracks the start of connection recovery
func (c *Client) trackRecoveryStart(reason string) {
	if c.monitoringHooks != nil && c.monitoringHooks.OnRecoveryStart != nil {
		c.monitoringHooks.OnRecoveryStart(reason)
	}
}

// trackRecoveryEnd tracks the end of connection recovery
func (c *Client) trackRecoveryEnd(success bool, attempts int, duration time.Duration) {
	if c.monitoringHooks != nil && c.monitoringHooks.OnRecoveryEnd != nil {
		c.monitoringHooks.OnRecoveryEnd(success, attempts, duration)
	}
}

// trackRetryAttempt tracks retry attempts
func (c *Client) trackRetryAttempt(operation string, attempt int, err error) {
	// Update internal metrics
	c.mu.Lock()
	if c.operationMetrics == nil {
		c.operationMetrics = make(map[string]*OperationMetrics)
	}

	metrics, exists := c.operationMetrics[operation]
	if !exists {
		metrics = &OperationMetrics{}
		c.operationMetrics[operation] = metrics
	}

	metrics.RetryCount++
	c.mu.Unlock()

	// Call external monitoring hook
	if c.monitoringHooks != nil && c.monitoringHooks.OnRetryAttempt != nil {
		c.monitoringHooks.OnRetryAttempt(operation, attempt, err)
	}
}

// GetOperationMetrics returns current operation metrics
func (c *Client) GetOperationMetrics() map[string]*OperationMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*OperationMetrics)
	for op, metrics := range c.operationMetrics {
		result[op] = &OperationMetrics{
			TotalCalls:     metrics.TotalCalls,
			TotalDuration:  metrics.TotalDuration,
			ErrorCount:     metrics.ErrorCount,
			LastError:      metrics.LastError,
			LastErrorTime:  metrics.LastErrorTime,
			SuccessCount:   metrics.SuccessCount,
			RetryCount:     metrics.RetryCount,
			AverageLatency: metrics.AverageLatency,
		}
	}

	return result
}

// ResetOperationMetrics resets all operation metrics
func (c *Client) ResetOperationMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.operationMetrics = make(map[string]*OperationMetrics)
}
