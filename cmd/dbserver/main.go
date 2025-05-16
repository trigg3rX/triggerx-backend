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

	// Initialize database connections using our new configurations
	// Connection to ScyllaDB node 1
	scylla1Config := database.NewScylla1Config()
	conn1, err := database.NewConnection(scylla1Config)
	if err != nil || conn1 == nil {
		logger.Fatalf("Failed to initialize connection to ScyllaDB node 1: %v", err)
	}
	defer conn1.Close()

	// Connection to ScyllaDB node 2
	scylla2Config := database.NewScylla2Config()
	conn2, err := database.NewConnection(scylla2Config)
	if err != nil || conn2 == nil {
		logger.Fatalf("Failed to initialize connection to ScyllaDB node 2: %v", err)
	}
	defer conn2.Close()

	// Ensure sessions are not nil before proceeding
	session1 := conn1.Session()
	session2 := conn2.Session()
	if session1 == nil || session2 == nil {
		logger.Fatalf("Database sessions cannot be nil")
	}

	// Set up server with both connections
	server := dbserver.NewServer(conn1, logging.DatabaseProcess) // Using conn1 as primary
	// registrar.SetDatabaseConnection(session1, session2)

	// Initialize and start HTTP server
	logger.Infof("Database Server initialized, starting on port %s...", config.DatabasePort)
	if err := server.Start(config.DatabasePort); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}