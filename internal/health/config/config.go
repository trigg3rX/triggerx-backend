package config

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.HealthProcess)

	HealthRPCPort     string
	DatabaseIPAddress string
	ManagerIPAddress  string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	HealthRPCPort = os.Getenv("HEALTH_RPC_PORT")
	ManagerIPAddress = os.Getenv("MANAGER_IP_ADDRESS")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	if HealthRPCPort == "" || ManagerIPAddress == "" || DatabaseIPAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}
