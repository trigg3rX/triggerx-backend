package connection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// scyllaConnectionManager holds the database session and configuration.
type scyllaConnectionManager struct {
	session       interfaces.Sessioner
	gocqlxSession interfaces.GocqlxSessioner
	config        *Config
	logger        logging.Logger
	mu            sync.RWMutex
	// Connection state tracking
	lastHealthCheck time.Time
	healthStatus    bool
	reconnectCount  int
	// Metrics
	healthCheckCount    int64
	healthCheckFailures int64
	reconnectAttempts   int64
	// Circuit breaker
	circuitBreakerState CircuitBreakerState
	circuitBreakerCount int
	circuitBreakerTime  time.Time
}

var (
	once     sync.Once
	instance *scyllaConnectionManager
)

// NewConnection creates a new ScyllaDB connection.
// It uses a singleton pattern to ensure only one connection is created.
func NewConnection(config *Config, logger logging.Logger) (interfaces.Connection, error) {
	// Validate configuration before creating connection
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	var err error
	once.Do(func() {
		cluster := gocql.NewCluster(config.Hosts...)
		cluster.Keyspace = config.Keyspace
		cluster.Timeout = config.Timeout
		cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: config.Retries}
		cluster.ConnectTimeout = config.ConnectWait
		cluster.Consistency = config.Consistency
		cluster.ProtoVersion = config.ProtoVersion
		cluster.SocketKeepalive = config.SocketKeepalive
		cluster.MaxPreparedStmts = config.MaxPreparedStmts
		cluster.DefaultIdempotence = config.DefaultIdempotence

		session, sessionErr := cluster.CreateSession()
		if sessionErr != nil {
			err = sessionErr
			return
		}

		// Create gocqlx session wrapper
		gocqlxSession := &gocqlxSessionWrapper{session: session}

		instance = &scyllaConnectionManager{
			session:       session,
			gocqlxSession: gocqlxSession,
			config:        config,
			logger:        logger,
		}

		// Start a background goroutine for health checks and reconnection
		// Only start if HealthCheckInterval is configured
		if config.HealthCheckInterval > 0 {
			go instance.startHealthChecker()
		}
	})

	return instance, err
}

// GetSession returns the underlying gocql session.
func (m *scyllaConnectionManager) GetSession() interfaces.Sessioner {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.session
}

// GetGocqlxSession returns the gocqlx session wrapper.
func (m *scyllaConnectionManager) GetGocqlxSession() interfaces.GocqlxSessioner {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gocqlxSession
}

// Close closes the database connection.
func (m *scyllaConnectionManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.session != nil {
		m.session.Close()
	}
}

// HealthCheck performs a simple query to check the database connection.
func (m *scyllaConnectionManager) HealthCheck(ctx context.Context) error {
	// Check circuit breaker
	if m.isCircuitBreakerOpen() {
		return fmt.Errorf("circuit breaker is open, health check blocked")
	}

	m.mu.Lock()
	m.healthCheckCount++
	m.lastHealthCheck = time.Now()
	m.mu.Unlock()

	sess := m.GetSession()
	if sess == nil {
		m.mu.Lock()
		m.healthStatus = false
		m.healthCheckFailures++
		m.mu.Unlock()
		m.recordCircuitBreakerFailure()
		return fmt.Errorf("database session is nil")
	}

	err := sess.Query("SELECT release_version FROM system.local").WithContext(ctx).Exec()

	m.mu.Lock()
	if err != nil {
		m.healthStatus = false
		m.healthCheckFailures++
		m.mu.Unlock()
		m.recordCircuitBreakerFailure()
	} else {
		m.healthStatus = true
		m.mu.Unlock()
		m.recordCircuitBreakerSuccess()
	}

	return err
}

// startHealthChecker is the core of the auto-reconnection logic.
func (m *scyllaConnectionManager) startHealthChecker() {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if err := m.HealthCheck(ctx); err != nil {
			m.logger.Errorf("Database health check failed: %v. Attempting to reconnect...", err)
			m.reconnect()
		}
	}
}

// reconnect attempts to reconnect to the database.
func (m *scyllaConnectionManager) reconnect() {
	m.mu.Lock()
	m.reconnectAttempts++
	m.reconnectCount++
	attempts := m.reconnectAttempts
	m.mu.Unlock()

	// Check if logger and config are available
	if m.logger == nil || m.config == nil {
		return // Cannot reconnect without logger or config
	}

	m.logger.Infof("Attempting to reconnect to database (attempt #%d)", attempts)

	operation := func() error {
		cluster := gocql.NewCluster(m.config.Hosts...)
		cluster.Keyspace = m.config.Keyspace
		cluster.Timeout = m.config.Timeout
		cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: m.config.Retries}
		cluster.ConnectTimeout = m.config.ConnectWait
		cluster.Consistency = m.config.Consistency
		cluster.ProtoVersion = m.config.ProtoVersion
		cluster.SocketKeepalive = m.config.SocketKeepalive
		cluster.MaxPreparedStmts = m.config.MaxPreparedStmts
		cluster.DefaultIdempotence = m.config.DefaultIdempotence

		newSession, err := cluster.CreateSession()
		if err == nil {
			m.mu.Lock()
			if m.session != nil {
				m.session.Close() // Close the old, dead session
			}
			m.session = newSession
			// Update gocqlx session wrapper
			m.gocqlxSession = &gocqlxSessionWrapper{session: newSession}
			m.healthStatus = true // Reset health status on successful reconnect
			m.mu.Unlock()
			m.logger.Infof("Successfully reconnected to the database after %d attempts", attempts)
		}
		return err
	}

	// Use retry logic for reconnection attempts
	var cfg *retry.RetryConfig
	if m.config.RetryConfig != nil {
		cfg = m.config.RetryConfig
	} else {
		cfg = retry.DefaultRetryConfig()
	}

	err := retry.RetryFunc(context.Background(), operation, cfg, m.logger)
	if err != nil {
		m.logger.Errorf("Failed to reconnect to the database after %d attempts: %v", attempts, err)
		m.mu.Lock()
		m.healthStatus = false
		m.mu.Unlock()
	}
}

// SetSession sets the session for the connection (for testing).
func (m *scyllaConnectionManager) SetSession(session interfaces.Sessioner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.session = session
}

// SetConfig sets the config for the connection (for testing).
func (m *scyllaConnectionManager) SetConfig(config *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// SetLogger sets the logger for the connection (for testing).
func (m *scyllaConnectionManager) SetLogger(logger logging.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
}

// GetHealthStatus returns the current health status and metrics
func (m *scyllaConnectionManager) GetHealthStatus() (bool, time.Time, int64, int64, int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.healthStatus, m.lastHealthCheck, m.healthCheckCount, m.healthCheckFailures, m.reconnectAttempts
}

// IsHealthy returns whether the connection is currently healthy
func (m *scyllaConnectionManager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.healthStatus
}

// GetReconnectCount returns the number of reconnection attempts
func (m *scyllaConnectionManager) GetReconnectCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.reconnectCount
}

// isCircuitBreakerOpen checks if the circuit breaker is open
func (m *scyllaConnectionManager) isCircuitBreakerOpen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.circuitBreakerState == CircuitBreakerOpen {
		// Check if we should transition to half-open
		if time.Since(m.circuitBreakerTime) > 30*time.Second {
			m.mu.RUnlock()
			m.mu.Lock()
			if m.circuitBreakerState == CircuitBreakerOpen && time.Since(m.circuitBreakerTime) > 30*time.Second {
				m.circuitBreakerState = CircuitBreakerHalfOpen
				if m.logger != nil {
					m.logger.Info("Circuit breaker transitioning to half-open state")
				}
			}
			m.mu.Unlock()
			m.mu.RLock()
		}
	}

	return m.circuitBreakerState == CircuitBreakerOpen
}

// recordCircuitBreakerSuccess records a successful operation
func (m *scyllaConnectionManager) recordCircuitBreakerSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.circuitBreakerState == CircuitBreakerHalfOpen {
		m.circuitBreakerState = CircuitBreakerClosed
		m.circuitBreakerCount = 0
		if m.logger != nil {
			m.logger.Info("Circuit breaker closed due to successful operation")
		}
	}
}

// recordCircuitBreakerFailure records a failed operation
func (m *scyllaConnectionManager) recordCircuitBreakerFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.circuitBreakerCount++

	// Open circuit breaker after 5 consecutive failures
	if m.circuitBreakerCount >= 5 {
		m.circuitBreakerState = CircuitBreakerOpen
		m.circuitBreakerTime = time.Now()
		if m.logger != nil {
			m.logger.Warn("Circuit breaker opened due to repeated failures")
		}
	}
}

// gocqlxSessionWrapper wraps a gocql session to implement the GocqlxSessioner interface
type gocqlxSessionWrapper struct {
	session interfaces.Sessioner
}

// Query creates a new gocqlx query
func (w *gocqlxSessionWrapper) Query(stmt string, names []string) interfaces.GocqlxQueryer {
	// Use the deprecated but still functional gocqlx.Query for now
	// This is the correct way to create a gocqlx query from a gocql session
	query := gocqlx.Query(w.session.Query(stmt), names)
	return &gocqlxQueryWrapper{query: &realGocqlxQuery{query: query}}
}

// Close closes the underlying session
func (w *gocqlxSessionWrapper) Close() {
	w.session.Close()
}

// gocqlxQueryWrapper wraps a *gocqlx.Queryx to implement the GocqlxQueryer interface
type gocqlxQueryWrapper struct {
	query interfaces.GocqlxQueryer
}

// WithContext implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) WithContext(ctx context.Context) interfaces.GocqlxQueryer {
	return &gocqlxQueryWrapper{query: w.query.WithContext(ctx)}
}

// BindStruct implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) BindStruct(data interface{}) interfaces.GocqlxQueryer {
	return &gocqlxQueryWrapper{query: w.query.BindStruct(data)}
}

// BindMap implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) BindMap(data map[string]interface{}) interfaces.GocqlxQueryer {
	return &gocqlxQueryWrapper{query: w.query.BindMap(data)}
}

// ExecRelease implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) ExecRelease() error {
	return w.query.ExecRelease()
}

// GetRelease implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) GetRelease(dest interface{}) error {
	return w.query.GetRelease(dest)
}

// Select implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) Select(dest interface{}) error {
	return w.query.Select(dest)
}

// SelectRelease implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) SelectRelease(dest interface{}) error {
	return w.query.SelectRelease(dest)
}

// realGocqlxQuery wraps a real gocqlx.Queryx to implement the GocqlxQueryer interface
type realGocqlxQuery struct {
	query *gocqlx.Queryx
}

// WithContext implements interfaces.GocqlxQueryer
func (r *realGocqlxQuery) WithContext(ctx context.Context) interfaces.GocqlxQueryer {
	return &realGocqlxQuery{query: r.query.WithContext(ctx)}
}

// BindStruct implements interfaces.GocqlxQueryer
func (r *realGocqlxQuery) BindStruct(data interface{}) interfaces.GocqlxQueryer {
	return &realGocqlxQuery{query: r.query.BindStruct(data)}
}

// BindMap implements interfaces.GocqlxQueryer
func (r *realGocqlxQuery) BindMap(data map[string]interface{}) interfaces.GocqlxQueryer {
	return &realGocqlxQuery{query: r.query.BindMap(data)}
}

// ExecRelease implements interfaces.GocqlxQueryer
func (r *realGocqlxQuery) ExecRelease() error {
	return r.query.ExecRelease()
}

// GetRelease implements interfaces.GocqlxQueryer
func (r *realGocqlxQuery) GetRelease(dest interface{}) error {
	return r.query.GetRelease(dest)
}

// Select implements interfaces.GocqlxQueryer
func (r *realGocqlxQuery) Select(dest interface{}) error {
	return r.query.Select(dest)
}

// SelectRelease implements interfaces.GocqlxQueryer
func (r *realGocqlxQuery) SelectRelease(dest interface{}) error {
	return r.query.SelectRelease(dest)
}
