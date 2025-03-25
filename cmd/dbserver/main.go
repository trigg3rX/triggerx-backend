package main

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	// "github.com/trigg3rX/triggerx-backend/internal/registrar"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	// Initialize development logger for database operations
	if err := logging.InitLogger(logging.Development, "database"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger(logging.Development, logging.DatabaseProcess)

	config.Init()

	// Establish database connection with default configuration
	cfg := database.NewConfig()
	conn, err := database.NewConnection(cfg)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Set the database connection for the registrar package
	// registrar.SetDatabaseConnection(conn)

	// Initialize and start HTTP server with database connection
	server := dbserver.NewServer(conn, logging.DatabaseProcess)
	logger.Infof("Database Server initialized, starting on port %s...", config.DatabasePort)
	if err := server.Start(config.DatabasePort); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
