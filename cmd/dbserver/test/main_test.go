package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestMain(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("ENV", "development")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DEV_MODE", "true")
	os.Setenv("DATABASE_HOST_ADDRESS", "localhost")
	os.Setenv("DATABASE_HOST_PORT", "9042")
	os.Setenv("DATABASE_KEYSPACE", "triggerx")
	os.Setenv("DATABASE_USERNAME", "")
	os.Setenv("DATABASE_PASSWORD", "")

	// Test configuration initialization
	t.Run("Config Initialization", func(t *testing.T) {
		// Skip config initialization test if .env file is not present
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			t.Skip("Skipping config initialization test as .env file is not present")
			return
		}
		err := config.Init()
		assert.NoError(t, err, "Config initialization should not fail")
	})

	// Test logger initialization
	t.Run("Logger Initialization", func(t *testing.T) {
		logConfig := logging.LoggerConfig{
			LogDir:          logging.BaseDataDir,
			ProcessName:     "dbserver",
			Environment:     logging.Development,
			UseColors:       true,
			MinStdoutLevel:  logging.DebugLevel,
			MinFileLogLevel: logging.DebugLevel,
		}
		err := logging.InitServiceLogger(logConfig)
		assert.NoError(t, err, "Logger initialization should not fail")
	})

	// Test database connection
	t.Run("Database Connection", func(t *testing.T) {
		logger := logging.GetServiceLogger()

		// Use environment variables directly for testing
		host := os.Getenv("DATABASE_HOST_ADDRESS")
		port := os.Getenv("DATABASE_HOST_PORT")
		if host == "" || port == "" {
			t.Skip("Skipping database connection test as host or port is not set")
			return
		}

		dbConfig := database.NewConfig(host, port)
		dbConfig = dbConfig.WithKeyspace("triggerx")

		conn, err := database.NewConnection(dbConfig)
		if err != nil {
			t.Logf("Database connection error: %v", err)
			t.Skip("Skipping database connection test due to connection error")
			return
		}
		assert.NotNil(t, conn, "Database connection should not be nil")
		logger.Info("Database connection successful")

		// Test database operations
		session := conn.Session()
		assert.NotNil(t, session, "Database session should not be nil")

		// Test keyspace creation
		err = session.Query(fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}", dbConfig.Keyspace)).Exec()
		if err != nil {
			t.Logf("Keyspace creation error: %v", err)
		}

		// Close connection
		if conn != nil {
			conn.Close()
		}
	})
}
