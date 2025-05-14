package config

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
	"log"

	"github.com/trigg3rX/triggerx-backend/pkg/utils"
)

var (
	HealthRPCPort     string
	DatabaseRPCAddress string

	DevMode bool
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	DevMode = os.Getenv("DEV_MODE") == "true"

	HealthRPCPort = os.Getenv("HEALTH_RPC_PORT")
	if !utils.IsValidPort(HealthRPCPort) {
		log.Fatal("Invalid Health RPC Port")
	}
	DatabaseRPCAddress = os.Getenv("DATABASE_RPC_ADDRESS")
	if !utils.IsValidRPCAddress(DatabaseRPCAddress) {
		log.Fatal("Invalid Database RPC Address")
	}

	gin.SetMode(gin.ReleaseMode)
}
