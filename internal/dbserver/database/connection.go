package database

import (
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// NewConnection creates a database connection
func NewConnection(logger observability.Logger) (*database.Connection, error) {
	dbConfig := &database.Config{
		Hosts:       []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:    "triggerx",
		Consistency: gocql.Quorum,
		Timeout:     config.GetDatabaseTimeout(),
		Retries:     config.GetDatabaseRetries(),
		ConnectWait: config.GetDatabaseConnectWait(),
		RetryConfig: retry.DefaultRetryConfig(),
	}

	// Configure authentication if provided
	if config.GetDatabaseUsername() != "" && config.GetDatabasePassword() != "" {
		dbConfig.WithAuthentication(config.GetDatabaseUsername(), config.GetDatabasePassword())
	}

	// Configure SSL/TLS if enabled
	if config.GetDatabaseSSLEnabled() {
		dbConfig.WithSSLCertificates(
			config.GetDatabaseSSLCertPath(),
			config.GetDatabaseSSLKeyPath(),
			config.GetDatabaseSSLCAPath(),
			config.GetDatabaseSSLInsecureSkipVerify(),
		)
	}

	return database.NewConnection(dbConfig, logger)
}
