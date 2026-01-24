package database

import (
	"crypto/tls"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// NewConnection creates a database connection
func NewConnection(logger observability.Logger) (*database.Connection, error) {
	dbConfig := database.NewConfig(
		config.GetDatabaseHostAddress(),
		config.GetDatabaseHostPort(),
	)
	dbConfig.Consistency = gocql.Quorum
	dbConfig.Timeout = 10 * time.Second
	dbConfig.Retries = 3
	dbConfig.ConnectWait = 5 * time.Second
	dbConfig.RetryConfig = retry.DefaultRetryConfig()

	// Configure authentication if provided
	if config.GetDatabaseUsername() != "" && config.GetDatabasePassword() != "" {
		dbConfig.WithAuthentication(config.GetDatabaseUsername(), config.GetDatabasePassword())
	}

	// Configure SSL/TLS if enabled
	if config.GetDatabaseSSLEnabled() {
		if config.GetDatabaseSSLCertPath() != "" && config.GetDatabaseSSLKeyPath() != "" {
			dbConfig.WithSSLCertificates(
				config.GetDatabaseSSLCertPath(),
				config.GetDatabaseSSLKeyPath(),
				config.GetDatabaseSSLCAPath(),
				config.GetDatabaseSSLInsecureSkipVerify(),
			)
		} else {
			dbConfig.WithSSL(&tls.Config{
				InsecureSkipVerify: config.GetDatabaseSSLInsecureSkipVerify(),
			})
		}
	}

	return database.NewConnection(dbConfig, logger)
}
