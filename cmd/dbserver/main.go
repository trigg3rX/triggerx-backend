package main

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger logging.Logger

func main() {
	config.Init()

	if config.DevMode {
		if err := logging.InitLogger(logging.Development, "database"); err != nil {
			panic(fmt.Sprintf("Failed to initialize logger: %v", err))
		}
		logger = logging.GetLogger(logging.Development, logging.DatabaseProcess)
	} else {
		if err := logging.InitLogger(logging.Production, "database"); err != nil {
			panic(fmt.Sprintf("Failed to initialize logger: %v", err))
		}
		logger = logging.GetLogger(logging.Production, logging.DatabaseProcess)
	}
	logger.Info("Starting database server...")

	dbConfig := database.NewConfig(config.DatabaseHost, config.DatabaseHostPort)

	conn, err := database.NewConnection(dbConfig)
	if err != nil || conn == nil {
		logger.Fatalf("Failed to initialize main database connection: %v", err)
	}
	defer conn.Close()

	mainSession := conn.Session()
	if mainSession == nil {
		logger.Fatalf("Database session cannot be nil")
	}

	server := dbserver.NewServer(conn, logging.DatabaseProcess)

	logger.Infof("Database Server initialized, starting on port %s...", config.DatabaseRPCPort)
	if err := server.Start(config.DatabaseRPCPort); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
