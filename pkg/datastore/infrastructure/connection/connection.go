package connection

import (
	"context"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// scyllaConnectionManager holds the database session and configuration.
type scyllaConnectionManager struct {
	session       interfaces.Sessioner
	gocqlxSession interfaces.GocqlxSessioner
	config        *Config
	logger        logging.Logger
	mu            sync.RWMutex
}

var (
	once     sync.Once
	instance *scyllaConnectionManager
)

// NewConnection creates a new ScyllaDB connection.
// It uses a singleton pattern to ensure only one connection is created.
func NewConnection(config *Config, logger logging.Logger) (interfaces.Connection, error) {
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
	sess := m.GetSession()
	return sess.Query("SELECT release_version FROM system.local").WithContext(ctx).Exec()
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
	defer m.mu.Unlock()

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
			if m.session != nil {
				m.session.Close() // Close the old, dead session
			}
			m.session = newSession
			// Update gocqlx session wrapper
			m.gocqlxSession = &gocqlxSessionWrapper{session: newSession}
			m.logger.Infof("Successfully reconnected to the database.")
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

	retry.RetryFunc(context.Background(), operation, cfg, m.logger)
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

// gocqlxSessionWrapper wraps a gocql session to implement the GocqlxSessioner interface
type gocqlxSessionWrapper struct {
	session *gocql.Session
}

// Query creates a new gocqlx query
func (w *gocqlxSessionWrapper) Query(stmt string, names []string) interfaces.GocqlxQueryer {
	// Use the deprecated but still functional gocqlx.Query for now
	// This is the correct way to create a gocqlx query from a gocql session
	query := gocqlx.Query(w.session.Query(stmt), names)
	return &gocqlxQueryWrapper{query: query}
}

// Close closes the underlying session
func (w *gocqlxSessionWrapper) Close() {
	w.session.Close()
}

// gocqlxQueryWrapper wraps a *gocqlx.Queryx to implement the GocqlxQueryer interface
type gocqlxQueryWrapper struct {
	query *gocqlx.Queryx
}

// WithContext implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) WithContext(ctx context.Context) interfaces.GocqlxQueryer {
	return &gocqlxQueryWrapper{query: w.query.WithContext(ctx)}
}

// BindStruct implements interfaces.GocqlxQueryer
func (w *gocqlxQueryWrapper) BindStruct(data interface{}) interfaces.GocqlxQueryer {
	return &gocqlxQueryWrapper{query: w.query.BindStruct(data)}
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
