package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	// User Entered Information
	PrivateKeyPerformer         string
	PrivateKeyAttester          string
	OperatorAddress             string

	// Provided Information
	PinataApiKey                string
	PinataSecretApiKey          string
	IpfsHost                    string
	OTHENTIC_CLIENT_RPC_ADDRESS string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	PinataApiKey = os.Getenv("PINATA_API_KEY")
	PinataSecretApiKey = os.Getenv("PINATA_SECRET_API_KEY")
	IpfsHost = os.Getenv("IPFS_HOST")
	OTHENTIC_CLIENT_RPC_ADDRESS = os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS")
	PrivateKeyPerformer = os.Getenv("PRIVATE_KEY_PERFORMER")
	PrivateKeyAttester = os.Getenv("PRIVATE_KEY")
	OperatorAddress = os.Getenv("OPERATOR_ADDRESS")

	if PinataApiKey == "" || PinataSecretApiKey == "" || OTHENTIC_CLIENT_RPC_ADDRESS == "" || IpfsHost == "" {
		log.Fatal(".env FILE NOT PRESENT AT EXPEXTED PATH")
	}
		
	if PrivateKeyPerformer == "" || PrivateKeyAttester == "" || OperatorAddress == "" {
		log.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}