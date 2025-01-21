package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/network"

	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/signerv2"

	// "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	// txservicemanager "github.com/trigg3rX/triggerx-contracts/bindings/contracts/TriggerXServiceManager"
	registryCoordinator "github.com/trigg3rX/triggerx-contracts/bindings/contracts/RegistryCoordinator"
)

type SignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    uint64
}

var (
	registryCoordinatorAddr = common.HexToAddress("0xB438C6Fc1652148BB758b939831f9A2cD59CE02b")
	// txservicemanagerAddress = "0x3d8C366becE100062d6BA471d03BbEc95fF5ac6A"
	// ownerAddr = common.HexToAddress("0xc073A5E091DC60021058346b10cD5A9b3F0619fE")
	ethClient                   *ethclient.Client
	logger                      logging.Logger
	txMgr                       *txmgr.SimpleTxManager
	registryCoordinatorContract *registryCoordinator.ContractRegistryCoordinator
)

func main() {
	// Change the logger initialization to assign to the package-level variable
	var err error
	logger, err = logging.NewZapLogger(logging.Development)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting quorum node...")

	ctx := context.Background()

	// Initialize registry
	registry, err := network.NewPeerRegistry()
	if err != nil {
		logger.Fatalf("Failed to initialize peer registry: %v", err)
	}

	// Setup P2P with registry
	config := network.P2PConfig{
		Name:    network.ServiceQuorum,
		Address: "/ip4/0.0.0.0/tcp/9002",
	}

	host, err := network.SetupP2PWithRegistry(ctx, config, registry)
	if err != nil {
		logger.Fatalf("Failed to setup P2P: %v", err)
	}

	// Initialize discovery service
	// discovery := network.NewDiscovery(ctx, host, config.Name)

	// Initialize messaging
	messaging := network.NewMessaging(host, config.Name)
	messaging.InitMessageHandling(func(msg network.Message) {
		logger.Infof("Received message from %s: %+v", msg.From, msg.Content)
	})

	// 	// Try to connect to manager
	// 	if _, err := discovery.ConnectToPeer(network.ServiceManager); err != nil {
	// 		logger.Warnf("Failed to connect to manager: %v", err)
	// 	}

	// 	logger.Infof("Quorum node is running. Node ID: %s", host.ID().String())
	// 	select {}

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	// Initialize Ethereum client
	ethClient, err = ethclient.Dial(os.Getenv("ETH_RPC_URL"))
	if err != nil {
		logger.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	// Get chain ID
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		logger.Fatalf("Failed to get chain ID: %v", err)
	}

	// Initialize private key and auth
	privateKey, err := crypto.HexToECDSA(os.Getenv("OWNER_PRIVATE_KEY"))
	if err != nil {
		logger.Fatalf("Failed to decode private key: %v", err)
	}

	signerV2, signerAddr, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: privateKey}, chainID)
	if err != nil {
		logger.Fatalf("Failed to create signer: %v", err)
	}

	txSender, err := wallet.NewPrivateKeyWallet(ethClient, signerV2, signerAddr, logger)
	if err != nil {
		logger.Fatalf("Failed to create transaction sender: %v", err)
	}

	txMgr = txmgr.NewSimpleTxManager(txSender, ethClient, logger, signerAddr)

	registryCoordinatorContract, err = registryCoordinator.NewContractRegistryCoordinator(registryCoordinatorAddr, ethClient)
	if err != nil {
		logger.Fatalf("Failed to create registry coordinator contract: %v", err)
	}


	logger.Info("Starting server on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
