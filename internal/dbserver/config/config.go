package config

import (
	"github.com/joho/godotenv"
	"os"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.DatabaseProcess)

	ManagerIPAddress string
	DatabasePort     string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
	}

	ManagerIPAddress = os.Getenv("MANAGER_IP_ADDRESS")
	DatabasePort = os.Getenv("DATABASE_RPC_PORT")

	if ManagerIPAddress == "" || DatabasePort == "" {
		logger.Fatal(".env FILE NOT PRESENT AT EXPEXTED PATH")
	}
}
