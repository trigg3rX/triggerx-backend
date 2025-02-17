package main

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/api"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	// Initialize development logger for database operations
	if err := logging.InitLogger(logging.Development, "database"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger(logging.Development, logging.DatabaseProcess)

	// Establish database connection with default configuration
	config := database.NewConfig()
	conn, err := database.NewConnection(config)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Initialize and start HTTP server with database connection
	server := api.NewServer(conn, logging.DatabaseProcess)
	logger.Info("Database Server initialized, starting on port 8080...")
	if err := server.Start("8080"); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
