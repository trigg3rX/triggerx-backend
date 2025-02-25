package config

import (
	"os"
	// "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.DatabaseProcess)

	ManagerRPCAddress           string
	DatabaseRPCAddress          string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
	}

	ManagerRPCAddress = os.Getenv("MANAGER_RPC_ADDRESS")
	DatabaseRPCAddress = os.Getenv("DATABASE_RPC_ADDRESS")

	if ManagerRPCAddress == "" || DatabaseRPCAddress == "" {
		logger.Fatal(".env FILE NOT PRESENT AT EXPEXTED PATH")
	}

	// gin.SetMode(gin.ReleaseMode)
}