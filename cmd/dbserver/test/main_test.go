package test

// import (
// 	"os"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
// 	"github.com/trigg3rX/triggerx-backend/pkg/datastore"
// 	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// )

// func TestMain(t *testing.T) {
// 	// Set environment variables for testing
// 	if err := os.Setenv("ENV", "development"); err != nil {
// 		t.Fatalf("Failed to set ENV: %v", err)
// 	}
// 	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
// 		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
// 	}
// 	if err := os.Setenv("DEV_MODE", "true"); err != nil {
// 		t.Fatalf("Failed to set DEV_MODE: %v", err)
// 	}
// 	if err := os.Setenv("DATABASE_HOST_ADDRESS", "localhost"); err != nil {
// 		t.Fatalf("Failed to set DATABASE_HOST_ADDRESS: %v", err)
// 	}
// 	if err := os.Setenv("DATABASE_HOST_PORT", "9042"); err != nil {
// 		t.Fatalf("Failed to set DATABASE_HOST_PORT: %v", err)
// 	}
// 	if err := os.Setenv("DATABASE_KEYSPACE", "triggerx"); err != nil {
// 		t.Fatalf("Failed to set DATABASE_KEYSPACE: %v", err)
// 	}
// 	if err := os.Setenv("DATABASE_USERNAME", ""); err != nil {
// 		t.Fatalf("Failed to set DATABASE_USERNAME: %v", err)
// 	}
// 	if err := os.Setenv("DATABASE_PASSWORD", ""); err != nil {
// 		t.Fatalf("Failed to set DATABASE_PASSWORD: %v", err)
// 	}

// 	// Test configuration initialization
// 	t.Run("Config Initialization", func(t *testing.T) {
// 		// Skip config initialization test if .env file is not present
// 		if _, err := os.Stat(".env"); os.IsNotExist(err) {
// 			t.Skip("Skipping config initialization test as .env file is not present")
// 			return
// 		}
// 		err := config.Init()
// 		assert.NoError(t, err, "Config initialization should not fail")
// 	})

// 	// Test logger initialization
// 	logConfig := logging.LoggerConfig{
// 		ProcessName:   logging.DatabaseProcess,
// 		IsDevelopment: true,
// 	}
// 	logger, err := logging.NewZapLogger(logConfig)
// 	if err != nil {
// 		t.Fatalf("Failed to initialize logger: %v", err)
// 	}
// 	t.Run("Logger Initialization", func(t *testing.T) {
// 		if logger == nil {
// 			t.Fatalf("Logger should not be nil")
// 		}
// 		assert.NoError(t, err, "Logger initialization should not fail")
// 	})

// 	// Test database connection
// 	t.Run("Database Connection", func(t *testing.T) {
// 		// Use environment variables directly for testing
// 		host := os.Getenv("DATABASE_HOST_ADDRESS")
// 		port := os.Getenv("DATABASE_HOST_PORT")
// 		if host == "" || port == "" {
// 			t.Skip("Skipping database connection test as host or port is not set")
// 			return
// 		}

// 		dbConfig := connection.NewConfig(host, port)
// 		dbConfig = dbConfig.WithKeyspace("triggerx")

// 		datastore, err := datastore.NewService(dbConfig, logger)
// 		if err != nil {
// 			t.Logf("Database connection error: %v", err)
// 			t.Skip("Skipping database connection test due to connection error")
// 			return
// 		}
// 		assert.NotNil(t, datastore, "Database connection should not be nil")
// 		logger.Info("Database connection successful")

// 		// Close connection
// 		if datastore != nil {
// 			datastore.Close()
// 		}
// 	})
// }
