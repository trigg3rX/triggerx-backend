package database

import (
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// NewConnection creates a database connection
func NewConnection(logger observability.Logger) (*database.Connection, error) {
	dbConfig := database.NewConfig(config.GetDatabaseHostAddress(), config.GetDatabaseHostPort())
	dbConfig.Consistency = gocql.Quorum
	dbConfig.Timeout = config.GetDatabaseTimeout()
	dbConfig.Retries = config.GetDatabaseRetries()
	dbConfig.ConnectWait = config.GetDatabaseConnectWait()
	dbConfig.RetryConfig = retry.DefaultRetryConfig()

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
