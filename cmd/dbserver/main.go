package main

import (
	"fmt"
	"time"

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

	// Initialize database config
	dbConfig := &database.Config{
		Hosts:       []string{"localhost"},
		Timeout:     10 * time.Second,
		Retries:     3,
		ConnectWait: 5 * time.Second,
	}

	// Initialize the existing database connection
	conn, err := database.NewConnection(dbConfig)
	if err != nil || conn == nil {
		logger.Fatalf("Failed to initialize main database connection: %v", err)
	}
	defer conn.Close()

	// Initialize a separate connection for registrar
	registrarConn, err := database.NewConnection(dbConfig)
	if err != nil || registrarConn == nil {
		logger.Fatalf("Failed to initialize registrar database connection: %v", err)
	}
	defer registrarConn.Close()

	// Ensure session is not nil before passing to registrar
	mainSession := conn.Session()
	registrarSession := registrarConn.Session()
	if mainSession == nil || registrarSession == nil {
		logger.Fatalf("Database sessions cannot be nil")
	}

	// Set both connections where needed
	server := dbserver.NewServer(conn, logging.DatabaseProcess)
	registrar.SetDatabaseConnection(mainSession, registrarSession)

	// Initialize and start HTTP server with database connection
	logger.Infof("Database Server initialized, starting on port %s...", config.DatabasePort)
	if err := server.Start(config.DatabasePort); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
