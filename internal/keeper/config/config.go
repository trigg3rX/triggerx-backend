package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/validator"
)

var (
	EthRPCUrl     string
	BaseRPCUrl    string
	AlchemyAPIKey string

	PrivateKeyConsensus  string
	PrivateKeyController string
	KeeperAddress        string

	PublicIPV4Address string
	PeerID            string

	OperatorRPCPort string

	DevMode bool
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	DevMode = os.Getenv("DEV_MODE") == "true"

	EthRPCUrl = os.Getenv("L1_RPC")
	if validator.IsEmpty(EthRPCUrl) {
		log.Fatal("Invalid L1 RPC Address")
	}

	AlchemyAPIKey = os.Getenv("ALCHEMY_API_KEY")
	if validator.IsEmpty(AlchemyAPIKey) {
		log.Fatal("Alchemy API Key is not set")
	}

	BaseRPCUrl = os.Getenv("L2_RPC")
	if validator.IsEmpty(BaseRPCUrl) {
		log.Fatal("Invalid L2 RPC Address")
	}

	PrivateKeyConsensus = os.Getenv("PRIVATE_KEY")
	if !validator.IsValidPrivateKey(PrivateKeyConsensus) {
		log.Fatal("Invalid Private Key")
	}

	PrivateKeyController = os.Getenv("OPERATOR_PRIVATE_KEY")
	if !validator.IsValidPrivateKey(PrivateKeyController) {
		log.Fatal("Invalid Operator Private Key")
	}

	KeeperAddress = os.Getenv("OPERATOR_ADDRESS")
	if !validator.IsValidAddress(KeeperAddress) {
		log.Fatal("Invalid Operator Address")
	}

	PublicIPV4Address = os.Getenv("PUBLIC_IPV4_ADDRESS")
	if !validator.IsValidIPAddress(PublicIPV4Address) {
		log.Fatal("Invalid Public IP Address")
	}

	PeerID = os.Getenv("PEER_ID")
	if !validator.IsValidPeerID(PeerID) {
		log.Fatal("Invalid Peer ID")
	}

	OperatorRPCPort = os.Getenv("OPERATOR_RPC_PORT")
	if validator.IsEmpty(OperatorRPCPort) {
		log.Fatal("Invalid Operator RPC Port")
	}

	checkDefaultValues()

	isRegistered := checkKeeperRegistration()
	if !isRegistered {
		log.Println("Keeper address is not yet registered on L2. Please register the address before continuing. If registered, please wait for the registration to be confirmed.")
		log.Fatal("Keeper address is not registered on L2")
	}

	gin.SetMode(gin.ReleaseMode)
}
