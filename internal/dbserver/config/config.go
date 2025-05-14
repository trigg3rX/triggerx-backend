package config

import (
	"os"
	"log"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/utils"
)

var (
	ManagerRPCAddress string
	DatabaseRPCPort     string

	DatabaseHost string
	DatabaseHostPort string

	EmailUser string
	EmailPassword string
	BotToken string

	AlchemyAPIKey string

	FaucetPrivateKey string
	FaucetFundAmount string

	UpstashRedisUrl string
	UpstashRedisRestToken string

	DevMode bool
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", "error", err)
	}
	DevMode = os.Getenv("DEV_MODE") == "true"

	ManagerRPCAddress = os.Getenv("MANAGER_RPC_ADDRESS")
	if !utils.IsValidRPCAddress(ManagerRPCAddress) {
		log.Fatal("Invalid Manager RPC Address")
	}
	DatabaseRPCPort = os.Getenv("DATABASE_RPC_PORT")
	if !utils.IsValidPort(DatabaseRPCPort) {
		log.Fatal("Invalid Database RPC Port")
	}

	DatabaseHost = os.Getenv("DATABASE_HOST")
	if !utils.IsValidIPAddress(DatabaseHost) {
		log.Fatal("Invalid Database Host")
	}
	DatabaseHostPort = os.Getenv("DATABASE_HOST_PORT")
	if !utils.IsValidPort(DatabaseHostPort) {
		log.Fatal("Invalid Database Host Port")
	}
	if !DevMode {
		EmailUser = os.Getenv("EMAIL_USER")
		if !utils.IsValidEmail(EmailUser) {
			log.Fatal("Invalid Email User")
		}
		EmailPassword = os.Getenv("EMAIL_PASS")
		if !utils.IsEmpty(EmailPassword) {
			log.Fatal("Invalid Email Password")
		}
		BotToken = os.Getenv("BOT_TOKEN")
		if !utils.IsEmpty(BotToken) {
			log.Fatal("Invalid Bot Token")
		}
	}

	AlchemyAPIKey = os.Getenv("ALCHEMY_API_KEY")
	if utils.IsEmpty(AlchemyAPIKey) {
		log.Fatal("Invalid Alchemy API Key")
	}
	FaucetPrivateKey = os.Getenv("FAUCET_PRIVATE_KEY")
	if !utils.IsValidPrivateKey(FaucetPrivateKey) {
		log.Fatal("Invalid Faucet Private Key")
	}
	FaucetFundAmount = os.Getenv("FAUCET_FUND_AMOUNT")
	if utils.IsEmpty(FaucetFundAmount) {
		log.Fatal("Invalid Faucet Fund Amount")
	}

	UpstashRedisUrl = os.Getenv("UPSTASH_REDIS_URL")
	if utils.IsEmpty(UpstashRedisUrl) {
		log.Fatal("Invalid Upstash Redis URL")
	}
	UpstashRedisRestToken = os.Getenv("UPSTASH_REDIS_REST_TOKEN")
	if utils.IsEmpty(UpstashRedisRestToken) {
		log.Fatal("Invalid Upstash Redis Rest Token")
	}
}
